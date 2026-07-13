package data

import (
	"math"
	"testing"
	"time"
)

func TestCalculateAssetPositionAcrossAssetClasses(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name       string
		instrument AssetInstrument
		trades     []AssetTransaction
		quote      float64
		wantQty    float64
		wantValue  float64
	}{
		{"stock fifo", AssetInstrument{ID: 1, AssetType: AssetTypeStock, Symbol: "600000.SH", Name: "浦发银行", Currency: "CNY"}, []AssetTransaction{{ID: 1, Direction: "buy", Quantity: 100, Price: 10, FXRateToCNY: 1, TradeTime: now}, {ID: 2, Direction: "sell", Quantity: 40, Price: 12, FXRateToCNY: 1, TradeTime: now.Add(time.Minute)}}, 11, 60, 660},
		{"fund", AssetInstrument{ID: 2, AssetType: AssetTypeFund, Symbol: "510300", Name: "沪深300ETF", Currency: "CNY"}, []AssetTransaction{{Direction: "buy", Quantity: 1000, Price: 4, FXRateToCNY: 1}}, 4.2, 1000, 4200},
		{"bond clean plus accrued", AssetInstrument{ID: 3, AssetType: AssetTypeBond, Symbol: "BOND", Name: "测试债", Currency: "CNY", MetadataJSON: `{"faceValue":100,"accruedInterest":1}`}, []AssetTransaction{{Direction: "buy", Quantity: 10, Price: 99, FXRateToCNY: 1}}, 100, 10, 1010},
		{"metal unit conversion", AssetInstrument{ID: 4, AssetType: AssetTypePreciousMetal, Symbol: "AU", Name: "黄金", Currency: "CNY", MetadataJSON: `{"unitToQuote":2}`}, []AssetTransaction{{Direction: "buy", Quantity: 3, Price: 500, FXRateToCNY: 1}}, 520, 3, 3120},
		{"forex balance", AssetInstrument{ID: 5, AssetType: AssetTypeForex, Symbol: "USD", Name: "美元", Currency: "USD"}, []AssetTransaction{{Direction: "buy", Quantity: 100, Price: 1, FXRateToCNY: 7.2}}, 1, 100, 720},
		{"crypto fractional", AssetInstrument{ID: 6, AssetType: AssetTypeCrypto, Symbol: "BTC", Name: "Bitcoin", Currency: "USD"}, []AssetTransaction{{Direction: "buy", Quantity: .01, Price: 50000, FXRateToCNY: 7.2}}, 60000, .01, 4320},
		{"bank nav", AssetInstrument{ID: 7, AssetType: AssetTypeBankWealth, Symbol: "BANK", Name: "稳健理财", Currency: "CNY", MetadataJSON: `{"valuationMode":"nav"}`}, []AssetTransaction{{Direction: "buy", Quantity: 10000, Price: 1, FXRateToCNY: 1}}, 1.02, 10000, 10200},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			quote := &AssetQuote{Price: test.quote, Currency: test.instrument.Currency, Provider: "test", AsOf: now}
			position := calculateAssetPosition(test.instrument, test.trades, quote, map[bool]float64{true: 7.2, false: 1}[test.instrument.Currency != "CNY"])
			if math.Abs(position.Quantity-test.wantQty) > 1e-6 {
				t.Fatalf("quantity=%v want %v", position.Quantity, test.wantQty)
			}
			if math.Abs(position.MarketValueCNY-test.wantValue) > 1e-6 {
				t.Fatalf("value=%v want %v", position.MarketValueCNY, test.wantValue)
			}
		})
	}
}

func TestCalculateFuturesLongAndShort(t *testing.T) {
	instrument := AssetInstrument{ID: 8, AssetType: AssetTypeFutures, Symbol: "IF", Name: "股指期货", Currency: "CNY", MetadataJSON: `{"multiplier":300,"marginRatio":0.1}`}
	position := calculateAssetPosition(instrument, []AssetTransaction{{Direction: "open_long", Quantity: 2, Price: 4000, FXRateToCNY: 1}}, &AssetQuote{Price: 4100, Provider: "test", AsOf: time.Now()}, 1)
	if position.FloatingPnLCNY != 60000 {
		t.Fatalf("floating pnl=%v", position.FloatingPnLCNY)
	}
	if position.NotionalCNY != 2460000 {
		t.Fatalf("notional=%v", position.NotionalCNY)
	}
}

func TestCalculateFuturesMultipleClosesDoNotDoubleCountCashFlow(t *testing.T) {
	instrument := AssetInstrument{ID: 9, AssetType: AssetTypeFutures, Symbol: "IF", Currency: "CNY", MetadataJSON: `{"multiplier":10,"marginRatio":0.1}`}
	trades := []AssetTransaction{
		{ID: 1, Direction: "open_long", Quantity: 2, Price: 100, FXRateToCNY: 1},
		{ID: 2, Direction: "close_long", Quantity: 1, Price: 110, FXRateToCNY: 1},
		{ID: 3, Direction: "close_long", Quantity: 1, Price: 120, FXRateToCNY: 1},
	}
	position := calculateAssetPosition(instrument, trades, &AssetQuote{Price: 120, Provider: "test", AsOf: time.Now()}, 1)
	if position.RealizedPnLCNY != 300 {
		t.Fatalf("realized pnl = %v, want 300", position.RealizedPnLCNY)
	}
	if position.NetCashFlowCNY != 300 {
		t.Fatalf("net cash flow = %v, want 300", position.NetCashFlowCNY)
	}
}

func TestQuoteFreshnessUsesFetchedTimeAndAssetSpecificTTL(t *testing.T) {
	now := time.Now()
	instrument := AssetInstrument{ID: 10, AssetType: AssetTypeStock, Symbol: "600000.SH", Currency: "CNY"}
	trade := []AssetTransaction{{Direction: "buy", Quantity: 1, Price: 10, FXRateToCNY: 1}}

	fresh := calculateAssetPosition(instrument, trade, &AssetQuote{Price: 11, Provider: "test", AsOf: now.Add(-72 * time.Hour), FetchedAt: now}, 1)
	if fresh.QuoteStale {
		t.Fatal("a quote fetched now should be fresh even when the market timestamp is older")
	}

	stale := calculateAssetPosition(instrument, trade, &AssetQuote{Price: 11, Provider: "test", AsOf: now, FetchedAt: now.Add(-20 * time.Minute)}, 1)
	if !stale.QuoteStale {
		t.Fatal("a stock quote fetched 20 minutes ago should be stale")
	}

	manual := calculateAssetPosition(instrument, trade, &AssetQuote{Price: 11, Provider: "manual", AsOf: now.Add(-72 * time.Hour), FetchedAt: now.Add(-72 * time.Hour)}, 1)
	if manual.QuoteStale {
		t.Fatal("manual quote should remain classified as manual, not stale")
	}
}
