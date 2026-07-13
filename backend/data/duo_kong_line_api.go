package data

import (
	"context"
	"fmt"
	"go-stock/backend/db"
	"go-stock/backend/logger"
	"go-stock/backend/models"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

type DuoKongLineCondition struct {
	Signal               string   `json:"signal"`
	Days                 int      `json:"days"`
	SignalDays           int      `json:"signalDays"`
	Limit                int      `json:"limit"`
	Keyword              string   `json:"keyword"`
	MinDeviation         float64  `json:"minDeviation"`
	MarketBoard          string   `json:"marketBoard"`
	MarketBoards         []string `json:"marketBoards"`
	Industries           []string `json:"industries"`
	Concepts             []string `json:"concepts"`
	MinPrice             float64  `json:"minPrice"`
	MaxPrice             float64  `json:"maxPrice"`
	MinTurnoverRate      float64  `json:"minTurnoverRate"`
	MaxTurnoverRate      float64  `json:"maxTurnoverRate"`
	MinDealAmount        float64  `json:"minDealAmount"`
	MaxDealAmount        float64  `json:"maxDealAmount"`
	MinVolumeRatio       float64  `json:"minVolumeRatio"`
	MaxVolumeRatio       float64  `json:"maxVolumeRatio"`
	MinFreeMarketCap     float64  `json:"minFreeMarketCap"`
	MaxFreeMarketCap     float64  `json:"maxFreeMarketCap"`
	OnlyFirstSignalToday bool     `json:"onlyFirstSignalToday"`
	MinAboveDays         int      `json:"minAboveDays"`
	KeepAboveAfterCross  bool     `json:"keepAboveAfterCross"`
}

type DuoKongLinePoint struct {
	Day   string  `json:"day"`
	Close float64 `json:"close"`
	Fine  float64 `json:"fine"`
	Thick float64 `json:"thick"`
}

type DuoKongLineResult struct {
	Code                 string  `json:"code"`
	Name                 string  `json:"name"`
	Industry             string  `json:"industry"`
	Concept              string  `json:"concept"`
	Close                float64 `json:"close"`
	Fine                 float64 `json:"fine"`
	Thick                float64 `json:"thick"`
	Deviation            float64 `json:"deviation"`
	Signal               string  `json:"signal"`
	SignalDate           string  `json:"signalDate"`
	SignalAge            int     `json:"signalAge"`
	TriggerDeviation     float64 `json:"triggerDeviation"`
	ConsecutiveAboveDays int     `json:"consecutiveAboveDays"`
	KeepsAboveAfterCross bool    `json:"keepsAboveAfterCross"`
	VolumeRatio          float64 `json:"volumeRatio"`
	TurnoverRate         float64 `json:"turnoverRate"`
	DealAmount           float64 `json:"dealAmount"`
	FreeMarketCap        float64 `json:"freeMarketCap"`
	KLineCode            string  `json:"kLineCode"`
	Market               string  `json:"market"`
	MatchedDays          int     `json:"matchedDays"`
}

type DuoKongLineScanTask struct {
	ID         string              `json:"id"`
	Status     string              `json:"status"`
	Progress   int                 `json:"progress"`
	Total      int                 `json:"total"`
	Results    []DuoKongLineResult `json:"results"`
	Error      string              `json:"error"`
	StartedAt  string              `json:"startedAt"`
	FinishedAt string              `json:"finishedAt"`
}

type duoKongKLineFetcher interface {
	GetCommonKLineData(stockCode string, kLineType string, days int64) *[]KLineData
}

type duoKongKLineCacheEntry struct {
	rows      []KLineData
	expiresAt time.Time
}

type DuoKongLineApi struct {
	stockApi     *StockDataApi
	kLineFetcher duoKongKLineFetcher
}

var (
	duoKongKLineCacheMu  sync.RWMutex
	duoKongKLineCache    = map[string]duoKongKLineCacheEntry{}
	duoKongKLineCacheTTL = 6 * time.Hour

	duoKongScanTaskMu  sync.RWMutex
	duoKongScanTasks   = map[string]*DuoKongLineScanTask{}
	duoKongScanCancels = map[string]context.CancelFunc{}
)

func NewDuoKongLineApi() *DuoKongLineApi {
	stockApi := NewStockDataApi()
	return &DuoKongLineApi{
		stockApi:     stockApi,
		kLineFetcher: stockApi,
	}
}

func duoKongEMAValues(values []float64, period int) []float64 {
	result := make([]float64, 0, len(values))
	if period <= 0 {
		return result
	}
	alpha := 2.0 / float64(period+1)
	var previous float64
	hasPrevious := false
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			result = append(result, math.NaN())
			continue
		}
		if !hasPrevious {
			previous = value
			hasPrevious = true
		} else {
			previous = alpha*value + (1-alpha)*previous
		}
		result = append(result, previous)
	}
	return result
}

func buildDuoKongLineSeries(klines []KLineData) []DuoKongLinePoint {
	days := make([]string, 0, len(klines))
	closes := make([]float64, 0, len(klines))
	weighted := make([]float64, 0, len(klines))
	for _, k := range klines {
		open, ok1 := parseDuoKongFloat(k.Open)
		high, ok2 := parseDuoKongFloat(k.High)
		low, ok3 := parseDuoKongFloat(k.Low)
		closePrice, ok4 := parseDuoKongFloat(k.Close)
		if k.Day == "" || !ok1 || !ok2 || !ok3 || !ok4 {
			continue
		}
		days = append(days, k.Day)
		closes = append(closes, closePrice)
		weighted = append(weighted, (open+high+low+2*closePrice)/5)
	}
	fine := duoKongEMAValues(closes, 10)
	thick := duoKongEMAValues(weighted, 30)
	points := make([]DuoKongLinePoint, 0, len(days))
	for i := range days {
		points = append(points, DuoKongLinePoint{
			Day:   days[i],
			Close: closes[i],
			Fine:  fine[i],
			Thick: thick[i],
		})
	}
	return points
}

func parseDuoKongFloat(raw string) (float64, bool) {
	v, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, false
	}
	return v, true
}

func duoKongMatchesSignal(points []DuoKongLinePoint, signal string, withinDays int, minDeviation float64) bool {
	_, matched := duoKongSignalIndex(points, signal, withinDays, minDeviation)
	return matched
}

// duoKongSignalIndex returns the actual bar that satisfies a signal so result rows
// can describe the trigger date instead of always reporting the newest K-line date.
func duoKongSignalIndex(points []DuoKongLinePoint, signal string, withinDays int, minDeviation float64) (int, bool) {
	if len(points) < 2 {
		return -1, false
	}
	if withinDays <= 0 {
		withinDays = 1
	}
	latest := points[len(points)-1]
	if minDeviation > 0 && math.Abs(duoKongDeviation(latest.Fine, latest.Thick)) < minDeviation {
		return -1, false
	}
	if signal == "" {
		signal = "fine_cross_up"
	}

	start := len(points) - withinDays - 1
	if start < 1 {
		start = 1
	}
	for i := len(points) - 1; i >= start; i-- {
		prev := points[i-1]
		curr := points[i]
		switch signal {
		case "fine_cross_up":
			if prev.Fine <= prev.Thick && curr.Fine > curr.Thick {
				return i, true
			}
		case "fine_cross_down":
			if prev.Fine >= prev.Thick && curr.Fine < curr.Thick {
				return i, true
			}
		case "fine_above_thick":
			return len(points) - 1, latest.Fine > latest.Thick
		case "fine_below_thick":
			return len(points) - 1, latest.Fine < latest.Thick
		case "both_rising":
			return len(points) - 1, latest.Fine >= prev.Fine && latest.Thick >= prev.Thick
		case "both_falling":
			return len(points) - 1, latest.Fine < prev.Fine && latest.Thick < prev.Thick
		case "fine_turn_up":
			if i >= 2 {
				before := points[i-2]
				if prev.Fine <= before.Fine && curr.Fine > prev.Fine {
					return i, true
				}
			}
		case "fine_turn_down":
			if i >= 2 {
				before := points[i-2]
				if prev.Fine >= before.Fine && curr.Fine < prev.Fine {
					return i, true
				}
			}
		case "thick_turn_up":
			if i >= 2 {
				before := points[i-2]
				if prev.Thick <= before.Thick && curr.Thick > prev.Thick {
					return i, true
				}
			}
		case "thick_turn_down":
			if i >= 2 {
				before := points[i-2]
				if prev.Thick >= before.Thick && curr.Thick < prev.Thick {
					return i, true
				}
			}
		default:
			return -1, false
		}
	}
	return -1, false
}

func duoKongDeviation(fine, thick float64) float64 {
	if thick == 0 {
		return 0
	}
	return (fine - thick) / thick * 100
}

func duoKongMatchesMarketBoard(code string, board string) bool {
	code = strings.TrimSpace(code)
	board = strings.ToUpper(strings.TrimSpace(board))
	if board == "" || board == "ALL" {
		return true
	}
	switch board {
	case "STAR":
		return strings.HasPrefix(code, "688") || strings.HasPrefix(code, "689")
	case "CHINEXT":
		return strings.HasPrefix(code, "300") || strings.HasPrefix(code, "301")
	case "BSE":
		return strings.HasPrefix(code, "8") || strings.HasPrefix(code, "4")
	case "MAIN":
		return strings.HasPrefix(code, "000") ||
			strings.HasPrefix(code, "001") ||
			strings.HasPrefix(code, "002") ||
			strings.HasPrefix(code, "003") ||
			strings.HasPrefix(code, "600") ||
			strings.HasPrefix(code, "601") ||
			strings.HasPrefix(code, "603") ||
			strings.HasPrefix(code, "605")
	default:
		return true
	}
}

func (a *DuoKongLineApi) GetSelections(condition DuoKongLineCondition) []DuoKongLineResult {
	condition = normalizeDuoKongCondition(condition)
	stocks := a.getSelectionStocks(condition)
	results := make([]DuoKongLineResult, 0)
	for _, stock := range stocks {
		result, ok := a.evaluateStock(stock, condition)
		if !ok {
			continue
		}
		results = append(results, result)
		if condition.Limit > 0 && len(results) >= condition.Limit {
			break
		}
	}
	sortDuoKongResults(results)
	return results
}

func (a *DuoKongLineApi) StartScan(condition DuoKongLineCondition) string {
	condition = normalizeDuoKongCondition(condition)
	taskID := fmt.Sprintf("duokong-%d", time.Now().UnixNano())
	ctx, cancel := context.WithCancel(context.Background())
	task := &DuoKongLineScanTask{
		ID:        taskID,
		Status:    "running",
		Results:   []DuoKongLineResult{},
		StartedAt: time.Now().Format(time.RFC3339),
	}
	duoKongScanTaskMu.Lock()
	duoKongScanTasks[taskID] = task
	duoKongScanCancels[taskID] = cancel
	duoKongScanTaskMu.Unlock()

	go a.runScanTask(ctx, taskID, condition)
	return taskID
}

func (a *DuoKongLineApi) CancelScan(taskID string) bool {
	duoKongScanTaskMu.Lock()
	defer duoKongScanTaskMu.Unlock()
	task := duoKongScanTasks[taskID]
	cancel := duoKongScanCancels[taskID]
	if task == nil || cancel == nil || task.Status != "running" {
		return false
	}
	task.Status = "cancelling"
	cancel()
	return true
}

func (a *DuoKongLineApi) GetScanTask(taskID string) *DuoKongLineScanTask {
	duoKongScanTaskMu.RLock()
	task := duoKongScanTasks[taskID]
	if task == nil {
		duoKongScanTaskMu.RUnlock()
		return nil
	}
	clone := *task
	clone.Results = append([]DuoKongLineResult(nil), task.Results...)
	duoKongScanTaskMu.RUnlock()
	return &clone
}

func (a *DuoKongLineApi) runScanTask(ctx context.Context, taskID string, condition DuoKongLineCondition) {
	stocks := a.getSelectionStocks(condition)
	duoKongScanTaskMu.Lock()
	if task := duoKongScanTasks[taskID]; task != nil {
		task.Total = len(stocks)
	}
	duoKongScanTaskMu.Unlock()

	defer func() {
		duoKongScanTaskMu.Lock()
		delete(duoKongScanCancels, taskID)
		duoKongScanTaskMu.Unlock()
		if r := recover(); r != nil {
			duoKongScanTaskMu.Lock()
			if task := duoKongScanTasks[taskID]; task != nil {
				task.Status = "error"
				task.Error = fmt.Sprint(r)
				task.FinishedAt = time.Now().Format(time.RFC3339)
			}
			duoKongScanTaskMu.Unlock()
		}
	}()

	type scanOutcome struct {
		result  DuoKongLineResult
		matched bool
	}
	workerCount := len(stocks)
	if workerCount > 8 {
		workerCount = 8
	}
	if workerCount == 0 {
		workerCount = 1
	}
	jobs := make(chan models.AllStockInfo)
	outcomes := make(chan scanOutcome, workerCount)
	var workers sync.WaitGroup
	workers.Add(workerCount)
	for worker := 0; worker < workerCount; worker++ {
		go func() {
			defer workers.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case stock, ok := <-jobs:
					if !ok {
						return
					}
					result, matched := a.evaluateStock(stock, condition)
					select {
					case <-ctx.Done():
						return
					case outcomes <- scanOutcome{result: result, matched: matched}:
					}
				}
			}
		}()
	}
	go func() {
		defer close(jobs)
		for _, stock := range stocks {
			select {
			case <-ctx.Done():
				return
			case jobs <- stock:
			}
		}
	}()
	go func() {
		workers.Wait()
		close(outcomes)
	}()

	results := make([]DuoKongLineResult, 0)
	processed := 0
	lastPublished := time.Now()
	for outcome := range outcomes {
		processed++
		if outcome.matched {
			results = append(results, outcome.result)
		}
		if processed%32 == 0 || time.Since(lastPublished) >= 500*time.Millisecond {
			sortDuoKongResults(results)
			duoKongScanTaskMu.Lock()
			if task := duoKongScanTasks[taskID]; task != nil {
				task.Progress = processed
				task.Results = append([]DuoKongLineResult(nil), results...)
			}
			duoKongScanTaskMu.Unlock()
			lastPublished = time.Now()
		}
	}
	sortDuoKongResults(results)

	duoKongScanTaskMu.Lock()
	if task := duoKongScanTasks[taskID]; task != nil {
		if ctx.Err() != nil {
			task.Status = "cancelled"
		} else {
			task.Status = "done"
			task.Progress = task.Total
		}
		task.Results = results
		task.FinishedAt = time.Now().Format(time.RFC3339)
	}
	duoKongScanTaskMu.Unlock()
}

func normalizeDuoKongCondition(condition DuoKongLineCondition) DuoKongLineCondition {
	condition.Days = 120
	condition.SignalDays = 3
	if condition.Signal == "" {
		condition.Signal = "fine_cross_up"
	}
	if strings.TrimSpace(condition.MarketBoard) == "" {
		condition.MarketBoard = "ALL"
	}
	if len(condition.MarketBoards) == 0 {
		condition.MarketBoards = []string{condition.MarketBoard}
	}
	if condition.MinAboveDays < 0 {
		condition.MinAboveDays = 0
	}
	return condition
}

func (a *DuoKongLineApi) getSelectionStocks(condition DuoKongLineCondition) []models.AllStockInfo {
	condition = normalizeDuoKongCondition(condition)
	var stocks []models.AllStockInfo
	query := db.Dao.Model(&models.AllStockInfo{}).
		Where("sec_uri_tycode IS NOT NULL AND sec_uri_tycode != ''")
	keyword := strings.TrimSpace(condition.Keyword)
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("sec_uri_tycode LIKE ? OR secucode LIKE ? OR sec_uri_tynameabbr LIKE ? OR industry LIKE ? OR concept LIKE ?", like, like, like, like, like)
	}
	if err := query.Order("sec_uri_tycode ASC").Find(&stocks).Error; err != nil {
		logger.SugaredLogger.Errorf("duo kong line failed to read stock list: %s", err.Error())
		return nil
	}

	listingDates := duoKongListingDates()
	filtered := make([]models.AllStockInfo, 0, len(stocks))
	for _, stock := range stocks {
		if duoKongMatchesUniverse(stock, condition, listingDates) {
			filtered = append(filtered, stock)
		}
	}
	return filtered
}

func duoKongListingDates() map[string]time.Time {
	if db.Dao == nil || !db.Dao.Migrator().HasTable(&StockBasic{}) {
		return map[string]time.Time{}
	}
	var basics []StockBasic
	if err := db.Dao.Select("symbol", "list_date").Find(&basics).Error; err != nil {
		return map[string]time.Time{}
	}
	dates := make(map[string]time.Time, len(basics))
	for _, basic := range basics {
		if date, ok := duoKongParseListingDate(basic.ListDate); ok {
			dates[strings.TrimSpace(basic.Symbol)] = date
		}
	}
	return dates
}

func duoKongParseListingDate(raw string) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	for _, layout := range []string{"20060102", "2006-01-02", time.RFC3339} {
		if value, err := time.ParseInLocation(layout, raw, time.Local); err == nil {
			return value, true
		}
	}
	return time.Time{}, false
}

func duoKongMatchesUniverse(stock models.AllStockInfo, condition DuoKongLineCondition, listingDates map[string]time.Time) bool {
	code := strings.TrimSpace(stock.SECURITYCODE)
	if !duoKongMatchesAnyMarketBoard(code, condition.MarketBoards) {
		return false
	}
	if duoKongMatchesMarketBoard(code, "BSE") {
		return false
	}
	name := strings.ToUpper(strings.TrimSpace(stock.SECURITYNAMEABBR))
	if strings.HasPrefix(name, "ST") || strings.HasPrefix(name, "*ST") {
		return false
	}
	priceRaw := strings.TrimSpace(stock.NEWPRICE)
	price, priceOK := parseDuoKongFloat(priceRaw)
	if priceRaw != "" && (!priceOK || price <= 0) {
		return false
	}
	if condition.MinPrice > 0 || condition.MaxPrice > 0 {
		if !priceOK || !duoKongInFloatRange(price, condition.MinPrice, condition.MaxPrice) {
			return false
		}
	}
	if listedAt, ok := listingDates[code]; ok && time.Since(listedAt) < 60*24*time.Hour {
		return false
	}
	turnover, _ := parseDuoKongFloat(stock.TURNOVERRATE)
	amount, _ := parseDuoKongFloat(stock.DEALAMOUNT)
	volumeRatio, _ := parseDuoKongFloat(stock.VOLUMERATIO)
	freeMarketCap, freeMarketCapOK := parseDuoKongFloat(stock.FREEMARKETCAP)
	if !duoKongInFloatRange(turnover, condition.MinTurnoverRate, condition.MaxTurnoverRate) ||
		!duoKongInFloatRange(amount, condition.MinDealAmount, condition.MaxDealAmount) ||
		!duoKongInFloatRange(volumeRatio, condition.MinVolumeRatio, condition.MaxVolumeRatio) {
		return false
	}
	if condition.MinFreeMarketCap > 0 || condition.MaxFreeMarketCap > 0 {
		if !freeMarketCapOK || !duoKongInFloatRange(freeMarketCap, condition.MinFreeMarketCap, condition.MaxFreeMarketCap) {
			return false
		}
	}
	if !duoKongMatchesTextFilter(stock.INDUSTRY, condition.Industries) || !duoKongMatchesTextFilter(stock.CONCEPT, condition.Concepts) {
		return false
	}
	return true
}

func duoKongMatchesAnyMarketBoard(code string, boards []string) bool {
	if len(boards) == 0 {
		return true
	}
	for _, board := range boards {
		if duoKongMatchesMarketBoard(code, board) {
			return true
		}
	}
	return false
}

func duoKongInFloatRange(value, min, max float64) bool {
	if min > 0 && value < min {
		return false
	}
	return max <= 0 || value <= max
}

func duoKongMatchesTextFilter(value string, filters []string) bool {
	if len(filters) == 0 {
		return true
	}
	value = strings.ToLower(value)
	for _, filter := range filters {
		filter = strings.ToLower(strings.TrimSpace(filter))
		if filter != "" && strings.Contains(value, filter) {
			return true
		}
	}
	return false
}

func (a *DuoKongLineApi) evaluateStock(stock models.AllStockInfo, condition DuoKongLineCondition) (DuoKongLineResult, bool) {
	kLineCode := duoKongKLineCode(stock)
	if kLineCode == "" {
		return DuoKongLineResult{}, false
	}
	klines := a.getCachedKLineData(kLineCode, "day", condition.Days)
	if len(klines) < 2 {
		return DuoKongLineResult{}, false
	}
	points := buildDuoKongLineSeries(klines)
	signalIndex, matched := duoKongSignalIndex(points, condition.Signal, condition.SignalDays, condition.MinDeviation)
	if !matched || signalIndex < 0 {
		return DuoKongLineResult{}, false
	}
	latest := points[len(points)-1]
	signalPoint := points[signalIndex]
	consecutiveAboveDays := duoKongConsecutiveAboveDays(points)
	keepsAboveAfterCross := duoKongKeepsAboveAfterCross(points, signalIndex, condition.Signal)
	if condition.OnlyFirstSignalToday && signalIndex != len(points)-1 {
		return DuoKongLineResult{}, false
	}
	if condition.MinAboveDays > 0 && consecutiveAboveDays < condition.MinAboveDays {
		return DuoKongLineResult{}, false
	}
	if condition.KeepAboveAfterCross && !keepsAboveAfterCross {
		return DuoKongLineResult{}, false
	}
	volumeRatio, _ := parseDuoKongFloat(stock.VOLUMERATIO)
	turnoverRate, _ := parseDuoKongFloat(stock.TURNOVERRATE)
	dealAmount, _ := parseDuoKongFloat(stock.DEALAMOUNT)
	freeMarketCap, _ := parseDuoKongFloat(stock.FREEMARKETCAP)
	return DuoKongLineResult{
		Code:                 stock.SECUCODE,
		Name:                 stock.SECURITYNAMEABBR,
		Industry:             stock.INDUSTRY,
		Concept:              stock.CONCEPT,
		Close:                latest.Close,
		Fine:                 latest.Fine,
		Thick:                latest.Thick,
		Deviation:            duoKongDeviation(latest.Fine, latest.Thick),
		Signal:               condition.Signal,
		SignalDate:           signalPoint.Day,
		SignalAge:            len(points) - 1 - signalIndex,
		TriggerDeviation:     duoKongDeviation(signalPoint.Fine, signalPoint.Thick),
		ConsecutiveAboveDays: consecutiveAboveDays,
		KeepsAboveAfterCross: keepsAboveAfterCross,
		VolumeRatio:          volumeRatio,
		TurnoverRate:         turnoverRate,
		DealAmount:           dealAmount,
		FreeMarketCap:        freeMarketCap,
		KLineCode:            kLineCode,
		Market:               stock.MARKET,
		MatchedDays:          condition.SignalDays,
	}, true
}

func duoKongConsecutiveAboveDays(points []DuoKongLinePoint) int {
	count := 0
	for i := len(points) - 1; i >= 0; i-- {
		if points[i].Fine <= points[i].Thick {
			break
		}
		count++
	}
	return count
}

func duoKongKeepsAboveAfterCross(points []DuoKongLinePoint, signalIndex int, signal string) bool {
	if signal != "fine_cross_up" || signalIndex < 0 || signalIndex >= len(points) {
		return true
	}
	for i := signalIndex; i < len(points); i++ {
		if points[i].Fine <= points[i].Thick {
			return false
		}
	}
	return true
}

func sortDuoKongResults(results []DuoKongLineResult) {
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].SignalAge != results[j].SignalAge {
			return results[i].SignalAge < results[j].SignalAge
		}
		leftDeviation := math.Abs(results[i].Deviation)
		rightDeviation := math.Abs(results[j].Deviation)
		if leftDeviation != rightDeviation {
			return leftDeviation > rightDeviation
		}
		if results[i].VolumeRatio != results[j].VolumeRatio {
			return results[i].VolumeRatio > results[j].VolumeRatio
		}
		return results[i].Code < results[j].Code
	})
}

func (a *DuoKongLineApi) getCachedKLineData(kLineCode string, kLineType string, days int) []KLineData {
	key := fmt.Sprintf("%s|%s", strings.ToLower(strings.TrimSpace(kLineCode)), kLineType)
	now := time.Now()
	duoKongKLineCacheMu.RLock()
	if entry, ok := duoKongKLineCache[key]; ok && now.Before(entry.expiresAt) {
		rows := duoKongTailRows(entry.rows, days)
		duoKongKLineCacheMu.RUnlock()
		return rows
	}
	duoKongKLineCacheMu.RUnlock()

	persistentRows, fetchedAt, hasPersistentRows := loadPersistentDuoKongKLines(kLineCode, kLineType, 0)
	if hasPersistentRows && duoKongCacheFresh(fetchedAt, now) {
		cacheDuoKongKLines(key, persistentRows, duoKongCacheExpiry(now))
		return duoKongTailRows(persistentRows, days)
	}

	fetcher := a.kLineFetcher
	if fetcher == nil {
		if a.stockApi == nil {
			a.stockApi = NewStockDataApi()
		}
		fetcher = a.stockApi
	}
	requestDays := days
	if requestDays <= 0 {
		requestDays = 120
	}
	if hasPersistentRows {
		requestDays = 5
	}
	raw := fetcher.GetCommonKLineData(kLineCode, kLineType, int64(requestDays))
	freshRows := []KLineData{}
	if raw != nil {
		freshRows = append(freshRows, (*raw)...)
	}
	if len(freshRows) > 0 {
		storePersistentDuoKongKLines(kLineCode, kLineType, freshRows, now)
		if rows, _, ok := loadPersistentDuoKongKLines(kLineCode, kLineType, 0); ok {
			persistentRows = rows
		} else if hasPersistentRows {
			persistentRows = duoKongMergeKLineRows(persistentRows, freshRows)
		} else {
			persistentRows = freshRows
		}
	} else if !hasPersistentRows {
		persistentRows = freshRows
	}

	cacheDuoKongKLines(key, persistentRows, duoKongCacheExpiry(now))
	return duoKongTailRows(persistentRows, days)
}

func duoKongTailRows(rows []KLineData, days int) []KLineData {
	rows = append([]KLineData(nil), rows...)
	if days > 0 && len(rows) > days {
		rows = rows[len(rows)-days:]
	}
	return rows
}

func duoKongMergeKLineRows(existing, fresh []KLineData) []KLineData {
	byDay := make(map[string]KLineData, len(existing)+len(fresh))
	for _, row := range append(append([]KLineData(nil), existing...), fresh...) {
		if strings.TrimSpace(row.Day) != "" {
			byDay[row.Day] = row
		}
	}
	rows := make([]KLineData, 0, len(byDay))
	for _, row := range byDay {
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Day < rows[j].Day })
	return rows
}

func cacheDuoKongKLines(key string, rows []KLineData, expiresAt time.Time) {
	duoKongKLineCacheMu.Lock()
	duoKongKLineCache[key] = duoKongKLineCacheEntry{
		rows:      append([]KLineData(nil), rows...),
		expiresAt: expiresAt,
	}
	duoKongKLineCacheMu.Unlock()
}

func loadPersistentDuoKongKLines(kLineCode, kLineType string, days int) ([]KLineData, time.Time, bool) {
	if db.Dao == nil || !db.Dao.Migrator().HasTable(&models.DuoKongKLineCache{}) {
		return nil, time.Time{}, false
	}
	var cached []models.DuoKongKLineCache
	err := db.Dao.Where("stock_code = ? AND k_line_type = ?", strings.ToLower(strings.TrimSpace(kLineCode)), kLineType).
		Order("day ASC").Find(&cached).Error
	if err != nil || len(cached) == 0 {
		return nil, time.Time{}, false
	}
	if days > 0 && len(cached) > days {
		cached = cached[len(cached)-days:]
	}
	rows := make([]KLineData, 0, len(cached))
	var fetchedAt time.Time
	for _, item := range cached {
		if item.FetchedAt.After(fetchedAt) {
			fetchedAt = item.FetchedAt
		}
		rows = append(rows, KLineData{
			Day:          item.Day,
			Open:         item.Open,
			High:         item.High,
			Low:          item.Low,
			Close:        item.Close,
			Volume:       item.Volume,
			Amount:       item.Amount,
			TurnoverRate: item.TurnoverRate,
			VolumeRatio:  item.VolumeRatio,
		})
	}
	return rows, fetchedAt, true
}

func storePersistentDuoKongKLines(kLineCode, kLineType string, rows []KLineData, fetchedAt time.Time) {
	if db.Dao == nil || len(rows) == 0 || !db.Dao.Migrator().HasTable(&models.DuoKongKLineCache{}) {
		return
	}
	stockCode := strings.ToLower(strings.TrimSpace(kLineCode))
	entries := make([]models.DuoKongKLineCache, 0, len(rows))
	for _, row := range rows {
		if strings.TrimSpace(row.Day) == "" {
			continue
		}
		entries = append(entries, models.DuoKongKLineCache{
			StockCode: stockCode, KLineType: kLineType, Day: row.Day,
			Open: row.Open, High: row.High, Low: row.Low, Close: row.Close,
			Volume: row.Volume, Amount: row.Amount, TurnoverRate: row.TurnoverRate,
			VolumeRatio: row.VolumeRatio, FetchedAt: fetchedAt,
		})
	}
	if len(entries) == 0 {
		return
	}
	days := make([]string, 0, len(entries))
	for _, entry := range entries {
		days = append(days, entry.Day)
	}
	if err := db.Dao.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("stock_code = ? AND k_line_type = ? AND day IN ?", stockCode, kLineType, days).Delete(&models.DuoKongKLineCache{}).Error; err != nil {
			return err
		}
		return tx.CreateInBatches(entries, 200).Error
	}); err != nil {
		logger.SugaredLogger.Warnf("store persistent duo kong kline cache failed for %s: %v", stockCode, err)
	}
}

func duoKongCacheExpiry(now time.Time) time.Time {
	if duoKongMarketIsOpen(now) {
		return now.Add(5 * time.Minute)
	}
	return duoKongNextTradingDayStart(now)
}

func duoKongCacheFresh(fetchedAt, now time.Time) bool {
	if fetchedAt.IsZero() {
		return false
	}
	if duoKongMarketIsOpen(now) {
		return now.Sub(fetchedAt) <= 5*time.Minute
	}
	return fetchedAt.After(duoKongPreviousTradingDayStart(now).Add(-time.Nanosecond))
}

func duoKongMarketIsOpen(now time.Time) bool {
	if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
		return false
	}
	minutes := now.Hour()*60 + now.Minute()
	return minutes >= 9*60+30 && minutes < 15*60
}

func duoKongPreviousTradingDayStart(now time.Time) time.Time {
	day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if duoKongMarketIsOpen(now) {
		day = day.AddDate(0, 0, -1)
	}
	for day.Weekday() == time.Saturday || day.Weekday() == time.Sunday {
		day = day.AddDate(0, 0, -1)
	}
	return day
}

func duoKongNextTradingDayStart(now time.Time) time.Time {
	day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, 1)
	for day.Weekday() == time.Saturday || day.Weekday() == time.Sunday {
		day = day.AddDate(0, 0, 1)
	}
	return day
}

func duoKongKLineCode(stock models.AllStockInfo) string {
	secucode := strings.TrimSpace(stock.SECUCODE)
	if strings.Contains(secucode, ".") {
		parts := strings.Split(secucode, ".")
		if len(parts) == 2 {
			return strings.ToLower(parts[1]) + parts[0]
		}
	}
	code := strings.TrimSpace(stock.SECURITYCODE)
	if code == "" {
		return ""
	}
	market := strings.ToLower(strings.TrimSpace(stock.MARKET))
	switch market {
	case "sh", "sse":
		return "sh" + code
	case "sz", "szse":
		return "sz" + code
	case "bj", "bse":
		return "bj" + code
	}
	if strings.HasPrefix(code, "6") || strings.HasPrefix(code, "5") {
		return "sh" + code
	}
	if strings.HasPrefix(code, "8") || strings.HasPrefix(code, "4") {
		return "bj" + code
	}
	return "sz" + code
}
