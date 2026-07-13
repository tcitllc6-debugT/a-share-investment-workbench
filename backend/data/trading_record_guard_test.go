package data

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestTradingRecordSellGuardUsesFIFOAvailableVolume(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(&TradingRecord{}); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	buy := TradingRecord{StockCode: "300418.SZ", StockName: "昆仑万维", Direction: "买入", Volume: 300, Price: 47.5, TradingTime: now}
	if err := database.Create(&buy).Error; err != nil {
		t.Fatal(err)
	}
	available, err := getStockAvailableVolume(database, "300418.sz", 0)
	if err != nil || available != 300 {
		t.Fatalf("available=%d err=%v", available, err)
	}
	if err := validateTradingRecordSell(database, TradingRecord{StockCode: "300418.SZ", Direction: "卖出", Volume: 301, TradingTime: now.Add(time.Minute)}, 0); err == nil {
		t.Fatal("expected oversell rejection")
	}
	if err := validateTradingRecordSell(database, TradingRecord{StockCode: "300418.SZ", Direction: "卖出", Volume: 100, TradingTime: now.Add(time.Minute)}, 0); err != nil {
		t.Fatal(err)
	}
}
