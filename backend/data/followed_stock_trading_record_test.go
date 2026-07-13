package data

import (
	"go-stock/backend/db"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupFollowedStockTradingRecordTestDB(t *testing.T) {
	t.Helper()
	dao, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	db.Dao = dao
	if err := db.Dao.AutoMigrate(&FollowedStock{}, &TradingRecord{}); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
}

func TestSetCostPriceAndVolumeCreatesBuyTradingRecord(t *testing.T) {
	setupFollowedStockTradingRecordTestDB(t)

	if err := db.Dao.Create(&FollowedStock{
		StockCode: "sz300418",
		Name:      "昆仑万维",
		Time:      time.Now(),
	}).Error; err != nil {
		t.Fatalf("seed followed stock: %v", err)
	}

	got := NewStockDataApi().SetCostPriceAndVolume(47.5, 300, "sz300418")
	if got != "设置成功" {
		t.Fatalf("SetCostPriceAndVolume returned %q", got)
	}

	var records []TradingRecord
	if err := db.Dao.Find(&records).Error; err != nil {
		t.Fatalf("query records: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 trading record, got %d", len(records))
	}
	record := records[0]
	if record.StockCode != "sz300418" || record.StockName != "昆仑万维" {
		t.Fatalf("unexpected stock identity: %+v", record)
	}
	if record.Direction != "买入" {
		t.Fatalf("expected buy direction, got %q", record.Direction)
	}
	if record.Price != 47.5 || record.Volume != 300 {
		t.Fatalf("unexpected price/volume: %.2f/%d", record.Price, record.Volume)
	}
	if record.Reason != "自选股成本自动生成" {
		t.Fatalf("unexpected reason: %q", record.Reason)
	}
}

func TestSetCostPriceAndVolumeDoesNotDuplicateAutoTradingRecord(t *testing.T) {
	setupFollowedStockTradingRecordTestDB(t)

	if err := db.Dao.Create(&FollowedStock{
		StockCode: "sz300418",
		Name:      "昆仑万维",
		Time:      time.Now(),
	}).Error; err != nil {
		t.Fatalf("seed followed stock: %v", err)
	}

	api := NewStockDataApi()
	api.SetCostPriceAndVolume(47.5, 300, "sz300418")
	api.SetCostPriceAndVolume(47.5, 300, "sz300418")

	var count int64
	if err := db.Dao.Model(&TradingRecord{}).Count(&count).Error; err != nil {
		t.Fatalf("count records: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 auto trading record, got %d", count)
	}
}

func TestSetCostPriceAndVolumeUpdatesTheExistingAutoTradingRecord(t *testing.T) {
	setupFollowedStockTradingRecordTestDB(t)
	if err := db.Dao.Create(&FollowedStock{StockCode: "sz300418", Name: "昆仑万维", Time: time.Now()}).Error; err != nil {
		t.Fatalf("seed followed stock: %v", err)
	}

	api := NewStockDataApi()
	api.SetCostPriceAndVolume(47.5, 300, "sz300418")
	api.SetCostPriceAndVolume(50, 200, "sz300418")

	var records []TradingRecord
	if err := db.Dao.Order("id ASC").Find(&records).Error; err != nil {
		t.Fatalf("query records: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected a single updated auto record, got %+v", records)
	}
	if records[0].Price != 50 || records[0].Volume != 200 {
		t.Fatalf("expected updated auto record 50/200, got %.2f/%d", records[0].Price, records[0].Volume)
	}
}

func TestMigrateLegacyFollowedStockCostPositionsImportsOnlyMissingLedgers(t *testing.T) {
	setupFollowedStockTradingRecordTestDB(t)
	if err := db.Dao.Create(&[]FollowedStock{
		{StockCode: "sz300418", Name: "昆仑万维", CostPrice: 47.5, Volume: 300, Time: time.Now()},
		{StockCode: "sz300750", Name: "宁德时代", CostPrice: 200, Volume: 100, Time: time.Now()},
	}).Error; err != nil {
		t.Fatalf("seed followed stocks: %v", err)
	}
	if err := db.Dao.Create(&TradingRecord{StockCode: "sz300750", StockName: "宁德时代", Direction: "买入", Price: 180, Volume: 100, TradingTime: time.Now()}).Error; err != nil {
		t.Fatalf("seed existing ledger: %v", err)
	}

	if got := MigrateLegacyFollowedStockCostPositions(); got != 1 {
		t.Fatalf("expected one legacy import, got %d", got)
	}
	if got := MigrateLegacyFollowedStockCostPositions(); got != 0 {
		t.Fatalf("expected idempotent migration, got %d new records", got)
	}

	var records []TradingRecord
	if err := db.Dao.Order("stock_code ASC").Find(&records).Error; err != nil {
		t.Fatalf("query ledgers: %v", err)
	}
	if len(records) != 2 || records[0].StockCode != "sz300418" || records[0].Reason != "自选股成本自动生成" {
		t.Fatalf("unexpected ledger migration result: %+v", records)
	}
}
