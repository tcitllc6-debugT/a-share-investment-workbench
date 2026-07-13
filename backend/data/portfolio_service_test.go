package data

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestPortfolioServiceRejectsReductionBeyondPosition(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(&AssetInstrument{}, &AssetTransaction{}, &AssetQuote{}, &FXRate{}, &PortfolioSnapshot{}); err != nil {
		t.Fatal(err)
	}
	service := NewPortfolioServiceWithDB(database)
	base := AssetTransactionInput{AssetType: AssetTypeCrypto, Symbol: "BTC", Name: "Bitcoin", Currency: "USD", FXRateToCNY: 7.2, Quantity: 1, Price: 50000, TradeTime: time.Now()}
	base.Direction = "buy"
	if _, err := service.AddAssetTransaction(base); err != nil {
		t.Fatal(err)
	}
	base.Direction, base.Quantity = "sell", 2
	if _, err := service.AddAssetTransaction(base); err == nil {
		t.Fatal("expected reduction guard")
	}
}

func TestPortfolioRecentTransactionUsesAssetSpecificContractValue(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(&AssetInstrument{}, &AssetTransaction{}, &AssetQuote{}, &FXRate{}, &PortfolioSnapshot{}); err != nil {
		t.Fatal(err)
	}
	instrument := AssetInstrument{AssetType: AssetTypeFutures, Symbol: "IF", Name: "股指期货", Currency: "CNY", MetadataJSON: `{"multiplier":300,"marginRatio":0.1}`, IsActive: true}
	if err := database.Create(&instrument).Error; err != nil {
		t.Fatal(err)
	}
	transaction := AssetTransaction{InstrumentID: instrument.ID, AssetType: AssetTypeFutures, Direction: "open_long", Quantity: 2, Price: 4000, Currency: "CNY", FXRateToCNY: 1, TradeTime: time.Now()}
	if err := database.Create(&transaction).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.Create(&AssetQuote{InstrumentID: instrument.ID, Provider: "test", Price: 4100, Currency: "CNY", AsOf: time.Now(), FetchedAt: time.Now()}).Error; err != nil {
		t.Fatal(err)
	}

	overview, err := NewPortfolioServiceWithDB(database).GetPortfolioOverview()
	if err != nil {
		t.Fatal(err)
	}
	if len(overview.RecentTransactions) != 1 {
		t.Fatalf("recent transactions = %d", len(overview.RecentTransactions))
	}
	if got := overview.RecentTransactions[0].AmountCNY; got != 2_400_000 {
		t.Fatalf("recent amount = %v, want 2400000", got)
	}
}
