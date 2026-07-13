package data

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestAssetMigrationIsIdempotentAndReconcilesLegacyStockLedger(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(&TradingRecord{}); err != nil {
		t.Fatal(err)
	}
	legacy := TradingRecord{StockCode: "600519.SH", StockName: "贵州茅台", Direction: "买入", Price: 1500, Volume: 10, TradingTime: time.Now()}
	if err := database.Create(&legacy).Error; err != nil {
		t.Fatal(err)
	}
	if err := RunAssetSchemaMigration(database, ":memory:"); err != nil {
		t.Fatal(err)
	}
	if err := RunAssetSchemaMigration(database, ":memory:"); err != nil {
		t.Fatal(err)
	}
	var count int64
	if err := database.Model(&AssetTransaction{}).Where("external_ref = ?", "legacy:trading_records:1").Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("migrated count=%d", count)
	}
	var version AssetSchemaVersion
	if err := database.Order("version DESC").First(&version).Error; err != nil {
		t.Fatal(err)
	}
	if version.Version != CurrentAssetSchemaVersion {
		t.Fatalf("version=%d", version.Version)
	}
}
