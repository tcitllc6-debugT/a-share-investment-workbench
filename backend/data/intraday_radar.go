package data

import (
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	RadarCategoryMomentum  = "momentum"
	RadarCategoryStrength  = "strength"
	RadarCategoryCapital   = "capital"
	RadarCategoryLimit     = "limit"
	RadarCategoryRisk      = "risk"
	RadarCategoryBreakdown = "breakdown"
	RadarPriorityP1        = "P1"
	RadarPriorityP2        = "P2"
	RadarPriorityP3        = "P3"
	defaultRadarCooldown   = 15 * time.Minute
)

type RadarFilter struct {
	Categories   []string `json:"categories"`
	OnlyFollowed bool     `json:"onlyFollowed"`
	OnlyHoldings bool     `json:"onlyHoldings"`
	MinAmount    float64  `json:"minAmount"`
	Industry     string   `json:"industry"`
	Concept      string   `json:"concept"`
	MarketBoards []string `json:"marketBoards"`
}

type RadarEvent struct {
	StockChangeItem
	Category   string `json:"category"`
	Priority   string `json:"priority"`
	IsFollowed bool   `json:"isFollowed"`
	IsHolding  bool   `json:"isHolding"`
	Action     string `json:"action"`
}

type RadarSummary struct {
	Total      int            `json:"total"`
	P1         int            `json:"p1"`
	P2         int            `json:"p2"`
	P3         int            `json:"p3"`
	Categories map[string]int `json:"categories"`
}

type RadarSnapshot struct {
	Events        []RadarEvent `json:"events"`
	Summary       RadarSummary `json:"summary"`
	LastSuccessAt time.Time    `json:"lastSuccessAt"`
	LastError     string       `json:"lastError"`
	Source        string       `json:"source"`
}

type RadarSettings struct {
	Enabled            bool `json:"enabled"`
	NotifyP1           bool `json:"notifyP1"`
	NotifyP2           bool `json:"notifyP2"`
	OnlyFollowedNotify bool `json:"onlyFollowedNotify"`
	CooldownMinutes    int  `json:"cooldownMinutes"`
}

type IntradayRadarService struct {
	mu           sync.Mutex
	now          func() time.Time
	lastNotified map[string]time.Time
	snapshot     RadarSnapshot
	settings     RadarSettings
}

var defaultIntradayRadarService = NewIntradayRadarService()

func NewIntradayRadarService() *IntradayRadarService {
	return NewIntradayRadarServiceWithClock(time.Now)
}

func NewIntradayRadarServiceWithClock(now func() time.Time) *IntradayRadarService {
	return &IntradayRadarService{
		now:          now,
		lastNotified: make(map[string]time.Time),
		settings: RadarSettings{
			Enabled:            true,
			NotifyP1:           true,
			NotifyP2:           false,
			OnlyFollowedNotify: false,
			CooldownMinutes:    int(defaultRadarCooldown / time.Minute),
		},
	}
}

func DefaultIntradayRadarService() *IntradayRadarService {
	return defaultIntradayRadarService
}

func (s *IntradayRadarService) GetSnapshot() RadarSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return cloneRadarSnapshot(s.snapshot)
}

func (s *IntradayRadarService) GetSettings() RadarSettings {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.settings
}

func (s *IntradayRadarService) UpdateSettings(settings RadarSettings) RadarSettings {
	s.mu.Lock()
	defer s.mu.Unlock()
	if settings.CooldownMinutes < 1 {
		settings.CooldownMinutes = int(defaultRadarCooldown / time.Minute)
	}
	if settings.CooldownMinutes > 120 {
		settings.CooldownMinutes = 120
	}
	s.settings = settings
	return s.settings
}

func (s *IntradayRadarService) Refresh(filter RadarFilter) RadarSnapshot {
	response := NewStockChangesApi().GetAllStockChangesWithPaging(500)
	if response == nil {
		s.mu.Lock()
		s.snapshot.LastError = "异动数据源暂不可用"
		s.mu.Unlock()
		return s.GetSnapshot()
	}

	followed, holdings := radarStockSets()
	events := make([]RadarEvent, 0, len(response.Data))
	for _, item := range response.Data {
		code := radarStockCode(item.Code)
		event := classifyRadarEvent(item, followed[code], holdings[code])
		if !filter.matches(event) {
			continue
		}
		events = append(events, event)
	}

	sort.SliceStable(events, func(i, j int) bool {
		if radarPriorityRank(events[i].Priority) != radarPriorityRank(events[j].Priority) {
			return radarPriorityRank(events[i].Priority) < radarPriorityRank(events[j].Priority)
		}
		if events[i].IsHolding != events[j].IsHolding {
			return events[i].IsHolding
		}
		if events[i].IsFollowed != events[j].IsFollowed {
			return events[i].IsFollowed
		}
		return events[i].Time > events[j].Time
	})

	_, _ = NewStockChangeHistoryService().SaveStockChangesWithDedup(response.Data)
	snapshot := RadarSnapshot{
		Events:        events,
		Summary:       summarizeRadarEvents(events),
		LastSuccessAt: s.now(),
		Source:        "东方财富异动",
	}
	s.mu.Lock()
	s.snapshot = snapshot
	s.mu.Unlock()
	s.notify(events)
	return s.GetSnapshot()
}

func (s *IntradayRadarService) shouldNotify(event RadarEvent) bool {
	if event.Priority == RadarPriorityP3 {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	cooldown := time.Duration(s.settings.CooldownMinutes) * time.Minute
	key := event.Code + ":" + event.Category
	if last, ok := s.lastNotified[key]; ok && s.now().Sub(last) < cooldown {
		return false
	}
	s.lastNotified[key] = s.now()
	return true
}

func (s *IntradayRadarService) notify(events []RadarEvent) {
	settings := s.GetSettings()
	if !settings.Enabled {
		return
	}
	for _, event := range events {
		if event.Priority == RadarPriorityP1 && !settings.NotifyP1 {
			continue
		}
		if event.Priority == RadarPriorityP2 && !settings.NotifyP2 {
			continue
		}
		if settings.OnlyFollowedNotify && !event.IsFollowed && !event.IsHolding {
			continue
		}
		if !s.shouldNotify(event) {
			continue
		}
		go NewAlertWindowsApi("A股智投工作台 · 异动雷达", event.Name+" · "+event.TypeName, "点击应用中的异动雷达查看K线与分析入口", "").SendNotification()
	}
}

func classifyRadarEvent(item StockChangeItem, isFollowed, isHolding bool) RadarEvent {
	category := radarCategoryForChangeType(item.ChangeType)
	priority := RadarPriorityP3
	if isFollowed || isHolding {
		priority = RadarPriorityP1
	} else if isP2ChangeType(item.ChangeType) {
		priority = RadarPriorityP2
	}
	return RadarEvent{
		StockChangeItem: item,
		Category:        category,
		Priority:        priority,
		IsFollowed:      isFollowed,
		IsHolding:       isHolding,
		Action:          "open_kline",
	}
}

func radarCategoryForChangeType(changeType int) string {
	switch changeType {
	case 8201, 8202, 8207, 8209, 8211:
		return RadarCategoryMomentum
	case 8213, 8215:
		return RadarCategoryStrength
	case 8193, 64, 8194, 128:
		return RadarCategoryCapital
	case 4, 8, 16, 32:
		return RadarCategoryLimit
	case 8203, 8204, 8208, 8210, 8212:
		return RadarCategoryRisk
	case 8214, 8216:
		return RadarCategoryBreakdown
	default:
		return RadarCategoryMomentum
	}
}

func isP2ChangeType(changeType int) bool {
	switch changeType {
	case 8201, 8202, 8193, 8194, 4, 8, 16, 32, 8203, 8204, 8211, 8212, 8213, 8214:
		return true
	default:
		return false
	}
}

func (filter RadarFilter) matches(event RadarEvent) bool {
	if filter.OnlyFollowed && !event.IsFollowed && !event.IsHolding {
		return false
	}
	if filter.OnlyHoldings && !event.IsHolding {
		return false
	}
	if filter.MinAmount > 0 && event.Amount < filter.MinAmount {
		return false
	}
	if len(filter.Categories) > 0 && !containsRadarValue(filter.Categories, event.Category) {
		return false
	}
	if keyword := strings.TrimSpace(filter.Industry); keyword != "" && !strings.Contains(strings.ToLower(event.Industry), strings.ToLower(keyword)) {
		return false
	}
	if keyword := strings.TrimSpace(filter.Concept); keyword != "" && !strings.Contains(strings.ToLower(event.Concept), strings.ToLower(keyword)) {
		return false
	}
	if len(filter.MarketBoards) > 0 && !containsRadarValue(filter.MarketBoards, radarBoardForCode(event.Code)) {
		return false
	}
	return true
}

func radarStockSets() (map[string]bool, map[string]bool) {
	followed := make(map[string]bool)
	for _, stock := range dereferenceFollowedStocks(NewStockDataApi().GetFollowList(0)) {
		followed[radarStockCode(stock.StockCode)] = true
	}
	holdings := make(map[string]bool)
	for _, holding := range NewHoldingSummaryApi().GetHoldingSummary().Holdings {
		holdings[radarStockCode(holding.StockCode)] = true
	}
	return followed, holdings
}

func dereferenceFollowedStocks(stocks *[]FollowedStock) []FollowedStock {
	if stocks == nil {
		return nil
	}
	return *stocks
}

func radarStockCode(code string) string {
	code = strings.ToLower(strings.TrimSpace(code))
	for _, prefix := range []string{"sh", "sz", "bj"} {
		if strings.HasPrefix(code, prefix) {
			code = strings.TrimPrefix(code, prefix)
		}
	}
	return strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(code, ".sh"), ".sz"), ".bj")
}

func radarBoardForCode(code string) string {
	code = radarStockCode(code)
	switch {
	case strings.HasPrefix(code, "688"):
		return "STAR"
	case strings.HasPrefix(code, "300") || strings.HasPrefix(code, "301"):
		return "CHINEXT"
	default:
		return "MAIN"
	}
}

func containsRadarValue(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(value, target) || strings.EqualFold(value, "ALL") {
			return true
		}
	}
	return false
}

func radarPriorityRank(priority string) int {
	switch priority {
	case RadarPriorityP1:
		return 1
	case RadarPriorityP2:
		return 2
	default:
		return 3
	}
}

func summarizeRadarEvents(events []RadarEvent) RadarSummary {
	summary := RadarSummary{Total: len(events), Categories: make(map[string]int)}
	for _, event := range events {
		summary.Categories[event.Category]++
		switch event.Priority {
		case RadarPriorityP1:
			summary.P1++
		case RadarPriorityP2:
			summary.P2++
		default:
			summary.P3++
		}
	}
	return summary
}

func cloneRadarSnapshot(snapshot RadarSnapshot) RadarSnapshot {
	clone := snapshot
	clone.Events = append([]RadarEvent(nil), snapshot.Events...)
	clone.Summary.Categories = make(map[string]int, len(snapshot.Summary.Categories))
	for key, value := range snapshot.Summary.Categories {
		clone.Summary.Categories[key] = value
	}
	return clone
}
