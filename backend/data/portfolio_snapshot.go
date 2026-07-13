package data

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func CreatePortfolioSnapshot() (*PortfolioSnapshot, error) {
	return CreatePortfolioSnapshotWithDB(NewPortfolioService().DB, time.Now())
}

func CreatePortfolioSnapshotWithDB(database *gorm.DB, at time.Time) (*PortfolioSnapshot, error) {
	if database == nil {
		return nil, fmt.Errorf("portfolio database is unavailable")
	}
	overview, err := NewPortfolioServiceWithDB(database).GetPortfolioOverview()
	if err != nil {
		return nil, err
	}
	breakdown, _ := json.Marshal(overview.ByAssetType)
	currencies, _ := json.Marshal(overview.CurrencyExposure)
	snapshot := PortfolioSnapshot{
		SnapshotDate:    at.In(time.Local).Format("2006-01-02"),
		TotalCostCNY:    overview.TotalCostCNY,
		MarketValueCNY:  overview.MarketValueCNY,
		FloatingPnLCNY:  overview.FloatingPnLCNY,
		RealizedPnLCNY:  overview.RealizedPnLCNY,
		NetCashFlowCNY:  overview.NetCashFlowCNY,
		BreakdownJSON:   string(breakdown),
		CurrencyJSON:    string(currencies),
		StaleQuoteCount: overview.StaleQuoteCount,
	}
	err = database.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "snapshot_date"}},
		DoUpdates: clause.AssignmentColumns([]string{"total_cost_cny", "market_value_cny", "floating_pnl_cny", "realized_pnl_cny", "net_cash_flow_cny", "breakdown_json", "currency_json", "stale_quote_count", "updated_at"}),
	}).Create(&snapshot).Error
	if err != nil {
		return nil, fmt.Errorf("save portfolio snapshot: %w", err)
	}
	return &snapshot, nil
}

func GetPortfolioSnapshots(page, pageSize int) (*PortfolioSnapshotPage, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 365 {
		pageSize = 30
	}
	result := &PortfolioSnapshotPage{Page: page, PageSize: pageSize, List: []PortfolioSnapshot{}}
	database := NewPortfolioService().DB
	if database == nil {
		return result, fmt.Errorf("portfolio database is unavailable")
	}
	if err := database.Model(&PortfolioSnapshot{}).Count(&result.Total).Error; err != nil {
		return result, err
	}
	if err := database.Order("snapshot_date DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&result.List).Error; err != nil {
		return result, err
	}
	return result, nil
}
