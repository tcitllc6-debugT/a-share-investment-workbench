package assetsearch

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
)

func TestSearchEastmoneyFiltersEverySupportedAssetType(t *testing.T) {
	items := []map[string]any{
		{"Code": "600519", "Name": "贵州茅台", "JYS": "2", "Classify": "AStock", "SecurityTypeName": "沪A", "QuoteID": "1.600519"},
		{"Code": "AAPL", "Name": "苹果", "JYS": "105", "Classify": "UsStock", "SecurityTypeName": "美股", "QuoteID": "105.AAPL"},
		{"Code": "510300", "Name": "沪深300ETF", "JYS": "2", "Classify": "Fund", "SecurityTypeName": "ETF", "QuoteID": "1.510300"},
		{"Code": "019547", "Name": "24国债16", "JYS": "2", "Classify": "Bond", "SecurityTypeName": "国债", "QuoteID": "1.019547"},
		{"Code": "AGTD", "Name": "白银T+D", "JYS": "118", "Classify": "SGE", "SecurityTypeName": "现货", "QuoteID": "118.AGTD"},
		{"Code": "USDCNYC", "Name": "美元人民币中间价", "JYS": "120", "Classify": "Forex", "SecurityTypeName": "外汇", "QuoteID": "120.USDCNYC"},
		{"Code": "AU0", "Name": "沪金主连", "JYS": "113", "Classify": "Future", "SecurityTypeName": "贵金属期货", "QuoteID": "113.AU0"},
		{"Code": "IF0", "Name": "沪深主连", "JYS": "8", "Classify": "Future", "SecurityTypeName": "股指期货", "QuoteID": "8.IF0"},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("input"); got != "沪" {
			t.Fatalf("unexpected keyword %q", got)
		}
		if got := r.URL.Query().Get("token"); got != "" {
			t.Fatalf("public suggestion request must not embed a fixed token, got %q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"QuotationCodeTable": map[string]any{"Data": items}})
	}))
	defer server.Close()

	service := NewService(Config{Client: server.Client(), EastmoneyURL: server.URL, BinanceURL: server.URL})
	tests := []struct {
		assetType    string
		wantSymbol   string
		wantCategory string
	}{
		{"stock", "600519.SH", "股票"},
		{"fund", "510300.SH", "基金"},
		{"bond", "019547.SH", "债券"},
		{"precious_metal", "AGTD", "贵金属"},
		{"forex", "USD/CNY", "外汇"},
		{"futures", "IF0", "期货"},
	}
	for _, tc := range tests {
		t.Run(tc.assetType, func(t *testing.T) {
			got, err := service.Search(context.Background(), tc.assetType, "沪", 20)
			if err != nil {
				t.Fatal(err)
			}
			var item Instrument
			for _, candidate := range got {
				if candidate.Symbol == tc.wantSymbol {
					item = candidate
					break
				}
			}
			if item.Symbol != tc.wantSymbol || item.Category != tc.wantCategory {
				t.Fatalf("unexpected normalized item: %#v", item)
			}
			if item.Source != "eastmoney" || item.QuoteSymbol == "" || item.Currency == "" || item.Exchange == "" || item.Metadata == nil {
				t.Fatalf("standard fields must be populated: %#v", item)
			}
			if tc.assetType != AssetForex && item.Metadata["secid"] != item.QuoteSymbol {
				t.Fatalf("Eastmoney quote metadata is incomplete: %#v", item.Metadata)
			}
		})
	}
}

func TestPreciousMetalFuturesStayInFuturesAndForexUsesStandardPair(t *testing.T) {
	goldFuture := eastmoneyRecord{Code: "au2612", Name: "沪金2612", Classify: "Futures", SecurityTypeName: "期货", QuoteID: "113.au2612"}
	if eastmoneyMatchesAssetType(goldFuture, AssetPreciousMetal) {
		t.Fatal("gold futures must not be classified as spot precious metal")
	}
	if !eastmoneyMatchesAssetType(goldFuture, AssetFutures) {
		t.Fatal("gold futures must remain searchable in futures")
	}
	forex := normalizeEastmoneyRecord(eastmoneyRecord{Code: "USDCNYC", Name: "美元人民币中间价", Classify: "FOREX", SecurityTypeName: "外汇", QuoteID: "120.USDCNYC"}, AssetForex)
	if forex.Symbol != "USD/CNY" || forex.QuoteSymbol != "USD/CNY" {
		t.Fatalf("forex result must use provider-compatible pair, got %#v", forex)
	}
	if forex.Metadata["base_currency"] != "USD" || forex.Metadata["quote_currency"] != "CNY" {
		t.Fatalf("forex metadata must identify both currencies, got %#v", forex.Metadata)
	}
}

func TestEastmoneyUsesQuoteIDInsteadOfJYSForSecurityExchange(t *testing.T) {
	for _, tc := range []struct {
		record     eastmoneyRecord
		wantSymbol string
		wantMarket string
	}{
		{eastmoneyRecord{Code: "600519", QuoteID: "1.600519", ExchangeCode: "2", SecurityTypeName: "A股"}, "600519.SH", "上海证券交易所"},
		{eastmoneyRecord{Code: "000001", QuoteID: "0.000001", ExchangeCode: "6", SecurityTypeName: "A股"}, "000001.SZ", "深圳证券交易所"},
	} {
		got := normalizeEastmoneyRecord(tc.record, AssetStock)
		if got.Symbol != tc.wantSymbol || got.Exchange != tc.wantMarket {
			t.Fatalf("unexpected market mapping for %#v: %#v", tc.record, got)
		}
	}
}

func TestDefaultBinanceSearchUsesPublicMarketDataMirror(t *testing.T) {
	service := NewService(Config{})
	if service.binanceURL != "https://data-api.binance.vision/api/v3/exchangeInfo" {
		t.Fatalf("unexpected Binance directory endpoint %q", service.binanceURL)
	}
}

func TestSearchBinanceReturnsTradingSpotPairsOnly(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		_ = json.NewEncoder(w).Encode(map[string]any{"symbols": []map[string]any{
			{"symbol": "BTCUSDT", "status": "TRADING", "baseAsset": "BTC", "quoteAsset": "USDT", "isSpotTradingAllowed": true},
			{"symbol": "BTCDOWNUSDT", "status": "BREAK", "baseAsset": "BTCDOWN", "quoteAsset": "USDT", "isSpotTradingAllowed": true},
			{"symbol": "BTCUSD_PERP", "status": "TRADING", "baseAsset": "BTC", "quoteAsset": "USD", "isSpotTradingAllowed": false},
			{"symbol": "ETHUSDT", "status": "TRADING", "baseAsset": "ETH", "quoteAsset": "USDT", "isSpotTradingAllowed": true},
			{"symbol": "ETHBTC", "status": "TRADING", "baseAsset": "ETH", "quoteAsset": "BTC", "isSpotTradingAllowed": true},
		}})
	}))
	defer server.Close()

	service := NewService(Config{Client: server.Client(), EastmoneyURL: server.URL, BinanceURL: server.URL})
	got, err := service.Search(context.Background(), "crypto", "btc", 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("expected one tradable BTC spot pair, got %#v", got)
	}
	item := got[0]
	if item.Symbol != "BTC/USDT" || item.QuoteSymbol != "BTCUSDT" || item.Currency != "USD" || item.Source != "binance" || item.Category != "加密货币" {
		t.Fatalf("unexpected Binance normalization: %#v", item)
	}
	if item.Metadata["binance_symbol"] != "BTCUSDT" || item.Metadata["quote_currency"] != "USDT" {
		t.Fatalf("Binance quote metadata is incomplete: %#v", item.Metadata)
	}
	if _, err := service.Search(context.Background(), "crypto", "eth", 20); err != nil {
		t.Fatal(err)
	}
	if requests.Load() != 1 {
		t.Fatalf("Binance product directory should be cached, got %d requests", requests.Load())
	}
}

func TestSearchSkipsNetworkForEmptyKeywordAndBankWealth(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		http.Error(w, "should not be called", http.StatusInternalServerError)
	}))
	defer server.Close()

	service := NewService(Config{Client: server.Client(), EastmoneyURL: server.URL, BinanceURL: server.URL})
	for _, tc := range []struct{ assetType, keyword string }{{"stock", " "}, {"bank_wealth", "稳健"}} {
		got, err := service.Search(context.Background(), tc.assetType, tc.keyword, 20)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 0 {
			t.Fatalf("expected no results, got %#v", got)
		}
	}
	if requests.Load() != 0 {
		t.Fatalf("expected zero upstream calls, got %d", requests.Load())
	}
}

func TestSearchProvidesBuiltInGoldAndForexWhenRemoteSuggestionsDoNotContainThem(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"QuotationCodeTable": map[string]any{"Data": []any{}}})
	}))
	defer server.Close()

	service := NewService(Config{Client: server.Client(), EastmoneyURL: server.URL, BinanceURL: server.URL})
	for _, tc := range []struct {
		assetType string
		keyword   string
		want      string
		quoteID   string
	}{
		{AssetPreciousMetal, "黄金", "AU9999", "118.AU9999"},
		{AssetForex, "美元", "USD/CNY", "USD/CNY"},
		{AssetFutures, "黄金", "AU0", "AU0"},
	} {
		got, err := service.Search(context.Background(), tc.assetType, tc.keyword, 20)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) == 0 || got[0].Symbol != tc.want || got[0].QuoteSymbol != tc.quoteID {
			t.Fatalf("%s %q should expose %s, got %#v", tc.assetType, tc.keyword, tc.want, got)
		}
	}
}

func TestSearchCapsResultsAtTwenty(t *testing.T) {
	items := make([]map[string]any, 25)
	for i := range items {
		code := "60" + strconv.Itoa(1000+i)
		items[i] = map[string]any{"Code": code, "Name": "测试股票" + strconv.Itoa(i), "JYS": "1", "Classify": "AStock", "SecurityTypeName": "沪A", "QuoteID": "1." + code}
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"QuotationCodeTable": map[string]any{"Data": items}})
	}))
	defer server.Close()

	service := NewService(Config{Client: server.Client(), EastmoneyURL: server.URL, BinanceURL: server.URL})
	got, err := service.Search(context.Background(), "stock", "测试", 999)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 20 {
		t.Fatalf("expected hard limit 20, got %d", len(got))
	}
}
