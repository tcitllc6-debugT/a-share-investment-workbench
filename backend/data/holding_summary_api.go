package data

import (
	"go-stock/backend/db"
	"sort"
	"strings"
)

type HoldingSummary struct {
	TotalHoldingCount      int                  `json:"totalHoldingCount"`
	TotalPrincipal         float64              `json:"totalPrincipal"`
	TotalMarketValue       float64              `json:"totalMarketValue"`
	TotalFloatingProfit    float64              `json:"totalFloatingProfit"`
	TotalFloatingProfitPct float64              `json:"totalFloatingProfitPct"`
	Holdings               []HoldingSummaryItem `json:"holdings"`
}

type HoldingSummaryItem struct {
	StockCode          string          `json:"stockCode"`
	StockName          string          `json:"stockName"`
	CostPrice          float64         `json:"costPrice"`
	Volume             int64           `json:"volume"`
	LatestPrice        float64         `json:"latestPrice"`
	Principal          float64         `json:"principal"`
	MarketValue        float64         `json:"marketValue"`
	FloatingProfit     float64         `json:"floatingProfit"`
	FloatingProfitPct  float64         `json:"floatingProfitPct"`
	ChangePercent      float64         `json:"changePercent"`
	EntryPrice         float64         `json:"entryPrice"`
	TakeProfitPrice    float64         `json:"takeProfitPrice"`
	StopLossPrice      float64         `json:"stopLossPrice"`
	TradingRecordCount int             `json:"tradingRecordCount"`
	TradingRecords     []TradingRecord `json:"tradingRecords"`
}

type HoldingSummaryApi struct{}

type fifoLot struct {
	volume int64
	cost   float64
}

type fifoPosition struct {
	lots    []fifoLot
	records []TradingRecord
	name    string
}

func NewHoldingSummaryApi() *HoldingSummaryApi {
	return &HoldingSummaryApi{}
}

func (a *HoldingSummaryApi) GetHoldingSummary() *HoldingSummary {
	var followedStocks []FollowedStock
	db.Dao.Model(&FollowedStock{}).
		Where("is_del = ?", 0).
		Order("sort ASC, time DESC").
		Find(&followedStocks)
	positions := a.getFIFOPositions()

	summary := &HoldingSummary{Holdings: make([]HoldingSummaryItem, 0, len(positions))}
	seen := make(map[string]bool, len(followedStocks))
	for _, stock := range followedStocks {
		code := strings.ToLower(strings.TrimSpace(stock.StockCode))
		seen[code] = true
		position := positions[code]
		if position == nil || position.volume() <= 0 {
			continue
		}
		summary.addHolding(position.toHoldingSummaryItem(stock))
	}

	// The trading ledger is authoritative, so keep open positions visible even
	// after a stock was removed from the watchlist.
	extraCodes := make([]string, 0)
	for code, position := range positions {
		if !seen[code] && position.volume() > 0 {
			extraCodes = append(extraCodes, code)
		}
	}
	sort.Strings(extraCodes)
	for _, code := range extraCodes {
		position := positions[code]
		summary.addHolding(position.toHoldingSummaryItem(FollowedStock{
			StockCode: code,
			Name:      position.name,
		}))
	}

	summary.TotalHoldingCount = len(summary.Holdings)
	if summary.TotalPrincipal > 0 {
		summary.TotalFloatingProfitPct = summary.TotalFloatingProfit / summary.TotalPrincipal * 100
	}
	return summary
}

func (summary *HoldingSummary) addHolding(item HoldingSummaryItem) {
	summary.Holdings = append(summary.Holdings, item)
	summary.TotalPrincipal += item.Principal
	summary.TotalMarketValue += item.MarketValue
	summary.TotalFloatingProfit += item.FloatingProfit
}

func (a *HoldingSummaryApi) getFIFOPositions() map[string]*fifoPosition {
	var records []TradingRecord
	db.Dao.Model(&TradingRecord{}).Order("trading_time ASC, id ASC").Find(&records)
	positions := make(map[string]*fifoPosition)
	for _, record := range records {
		code := strings.ToLower(strings.TrimSpace(record.StockCode))
		if code == "" || record.Volume <= 0 {
			continue
		}
		position := positions[code]
		if position == nil {
			position = &fifoPosition{}
			positions[code] = position
		}
		position.records = append(position.records, record)
		if strings.TrimSpace(record.StockName) != "" {
			position.name = record.StockName
		}

		switch record.Direction {
		case "买入":
			feePerShare := record.Fee / float64(record.Volume)
			position.lots = append(position.lots, fifoLot{volume: record.Volume, cost: record.Price + feePerShare})
		case "卖出":
			position.remove(record.Volume)
		}
	}
	return positions
}

func (position *fifoPosition) remove(volume int64) {
	for volume > 0 && len(position.lots) > 0 {
		lot := &position.lots[0]
		if lot.volume > volume {
			lot.volume -= volume
			return
		}
		volume -= lot.volume
		position.lots = position.lots[1:]
	}
}

func (position *fifoPosition) volume() int64 {
	var volume int64
	for _, lot := range position.lots {
		volume += lot.volume
	}
	return volume
}

func (position *fifoPosition) principal() float64 {
	principal := 0.0
	for _, lot := range position.lots {
		principal += lot.cost * float64(lot.volume)
	}
	return principal
}

func (position *fifoPosition) toHoldingSummaryItem(stock FollowedStock) HoldingSummaryItem {
	volume := position.volume()
	principal := position.principal()
	costPrice := 0.0
	if volume > 0 {
		costPrice = principal / float64(volume)
	}
	latestPrice := stock.Price
	if latestPrice <= 0 {
		latestPrice = costPrice
	}
	marketValue := latestPrice * float64(volume)
	floatingProfit := marketValue - principal
	floatingProfitPct := 0.0
	if principal > 0 {
		floatingProfitPct = floatingProfit / principal * 100
	}
	records := append([]TradingRecord(nil), position.records...)
	for i := range records {
		records[i].Amount = records[i].Price * float64(records[i].Volume)
	}
	sort.SliceStable(records, func(i, j int) bool {
		if records[i].TradingTime.Equal(records[j].TradingTime) {
			return records[i].ID > records[j].ID
		}
		return records[i].TradingTime.After(records[j].TradingTime)
	})
	stockName := stock.Name
	if strings.TrimSpace(stockName) == "" {
		stockName = position.name
	}
	return HoldingSummaryItem{
		StockCode:          stock.StockCode,
		StockName:          stockName,
		CostPrice:          costPrice,
		Volume:             volume,
		LatestPrice:        latestPrice,
		Principal:          principal,
		MarketValue:        marketValue,
		FloatingProfit:     floatingProfit,
		FloatingProfitPct:  floatingProfitPct,
		ChangePercent:      stock.ChangePercent,
		EntryPrice:         stock.EntryPrice,
		TakeProfitPrice:    stock.TakeProfitPrice,
		StopLossPrice:      stock.StopLossPrice,
		TradingRecordCount: len(records),
		TradingRecords:     records,
	}
}
