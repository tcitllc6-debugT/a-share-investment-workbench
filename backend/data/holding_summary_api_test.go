package data

import (
	"go-stock/backend/db"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupHoldingSummaryTestDB(t *testing.T) {
	t.Helper()
	dao, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	db.Dao = dao
	if err := db.Dao.AutoMigrate(&FollowedStock{}, &TradingRecord{}); err != nil {
		t.Fatalf("migrate holding summary tables: %v", err)
	}
}

func TestHoldingSummaryUsesStructuredPositionsAndTradingRecords(t *testing.T) {
	setupHoldingSummaryTestDB(t)
	if err := db.Dao.Create(&FollowedStock{
		StockCode: "sz300418",
		Name:      "昆仑万维",
		CostPrice: 47.5,
		Volume:    300,
		Price:     52.5,
	}).Error; err != nil {
		t.Fatalf("seed followed stock: %v", err)
	}
	if err := db.Dao.Create(&TradingRecord{
		StockCode:   "sz300418",
		StockName:   "昆仑万维",
		Direction:   "买入",
		Price:       47.5,
		Volume:      300,
		TradingTime: time.Date(2026, 7, 8, 9, 30, 0, 0, time.Local),
		Reason:      "自选成本自动记录",
	}).Error; err != nil {
		t.Fatalf("seed trading record: %v", err)
	}

	summary := NewHoldingSummaryApi().GetHoldingSummary()
	if summary.TotalHoldingCount != 1 {
		t.Fatalf("expected one holding, got %+v", summary)
	}
	item := summary.Holdings[0]
	if item.Principal != 14250 {
		t.Fatalf("unexpected principal: %.2f", item.Principal)
	}
	if item.MarketValue != 15750 {
		t.Fatalf("unexpected market value: %.2f", item.MarketValue)
	}
	if item.FloatingProfit != 1500 {
		t.Fatalf("unexpected floating profit: %.2f", item.FloatingProfit)
	}
	if item.TradingRecordCount != 1 || len(item.TradingRecords) != 1 {
		t.Fatalf("expected trading record attached, got %+v", item)
	}
}

func TestHoldingSummaryUsesFIFORecordsInsteadOfFollowedStockCost(t *testing.T) {
	setupHoldingSummaryTestDB(t)
	if err := db.Dao.Create(&FollowedStock{
		StockCode: "sz300418",
		Name:      "昆仑万维",
		CostPrice: 99,
		Volume:    999,
		Price:     15,
	}).Error; err != nil {
		t.Fatalf("seed followed stock: %v", err)
	}
	base := time.Date(2026, 7, 8, 9, 30, 0, 0, time.Local)
	if err := db.Dao.Create(&[]TradingRecord{
		{StockCode: "sz300418", StockName: "昆仑万维", Direction: "买入", Price: 10, Volume: 100, TradingTime: base},
		{StockCode: "sz300418", StockName: "昆仑万维", Direction: "买入", Price: 12, Volume: 100, TradingTime: base.Add(time.Minute)},
		{StockCode: "sz300418", StockName: "昆仑万维", Direction: "卖出", Price: 20, Volume: 100, TradingTime: base.Add(2 * time.Minute)},
	}).Error; err != nil {
		t.Fatalf("seed FIFO records: %v", err)
	}

	summary := NewHoldingSummaryApi().GetHoldingSummary()
	if summary.TotalHoldingCount != 1 {
		t.Fatalf("expected one FIFO holding, got %+v", summary)
	}
	item := summary.Holdings[0]
	if item.Volume != 100 || item.CostPrice != 12 {
		t.Fatalf("expected remaining FIFO lot 12/100, got %.2f/%d", item.CostPrice, item.Volume)
	}
	if item.Principal != 1200 || item.MarketValue != 1500 || item.FloatingProfit != 300 {
		t.Fatalf("unexpected FIFO summary: %+v", item)
	}
}
