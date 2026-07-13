package assetquote

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFundEstimateProviderReadsIntradayEstimate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/002169.js" {
			t.Fatalf("unexpected estimate path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`jsonpgz({"fundcode":"002169","name":"永赢稳益债券","jzrq":"2026-07-09","dwjz":"1.1393","gsz":"1.1395","gszzl":"0.02","gztime":"2026-07-10 15:00"});`))
	}))
	defer server.Close()

	provider := NewFundEstimateProviderWithBaseURL(server.Client(), server.URL)
	quotes, err := provider.FetchQuotes(context.Background(), []QuoteRequest{{AssetClass: AssetFund, Symbol: "150.002169", Currency: "CNY", Metadata: map[string]string{"secid": "150.002169", "classify": "OTCFUND"}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(quotes) != 1 || quotes[0].Price != 1.1395 || quotes[0].Provider != "fund-estimate" {
		t.Fatalf("unexpected fund estimate: %+v", quotes)
	}
}

func TestFundNAVProviderReadsLatestPublishedNAV(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("fundCode") != "002169" {
			t.Fatalf("unexpected fund NAV query: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"Data":{"LSJZList":[{"FSRQ":"2026-07-10","DWJZ":"1.1393","JZZZL":"0.00"}]},"ErrCode":0}`))
	}))
	defer server.Close()

	provider := NewFundNAVProviderWithBaseURL(server.Client(), server.URL)
	quotes, err := provider.FetchQuotes(context.Background(), []QuoteRequest{{AssetClass: AssetFund, Symbol: "150.002169", Currency: "CNY", Metadata: map[string]string{"secid": "150.002169"}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(quotes) != 1 || quotes[0].Price != 1.1393 || quotes[0].Provider != "fund-nav" {
		t.Fatalf("unexpected fund NAV: %+v", quotes)
	}
}

func TestSinaFundProviderReadsLatestFundNAV(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("list") != "f_002169" {
			t.Fatalf("unexpected Sina fund query: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`var hq_str_f_002169="永赢稳益债券,1.1393,1.3612,1.1393,2026-07-10,22.4455";`))
	}))
	defer server.Close()

	provider := NewSinaFundProviderWithBaseURL(server.Client(), server.URL)
	quotes, err := provider.FetchQuotes(context.Background(), []QuoteRequest{{AssetClass: AssetFund, Symbol: "150.002169", Currency: "CNY", Metadata: map[string]string{"secid": "150.002169"}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(quotes) != 1 || quotes[0].Price != 1.1393 || quotes[0].Provider != "sina-fund" {
		t.Fatalf("unexpected Sina fund NAV: %+v", quotes)
	}
}

func TestDefaultRegistryHasAutomaticFallbacksForAllMarketAssets(t *testing.T) {
	registry := NewDefaultRegistry(DefaultRegistryConfig())
	requests := []QuoteRequest{
		{AssetClass: AssetStock, Symbol: "1.600703", Currency: "CNY", Metadata: map[string]string{"secid": "1.600703"}},
		{AssetClass: AssetFund, Symbol: "150.002169", Currency: "CNY", Metadata: map[string]string{"secid": "150.002169", "classify": "OTCFUND"}},
		{AssetClass: AssetBond, Symbol: "1.113001", Currency: "CNY", Metadata: map[string]string{"secid": "1.113001"}},
		{AssetClass: AssetPreciousMetal, Symbol: "118.AU9999", Currency: "CNY", Metadata: map[string]string{"secid": "118.AU9999"}},
		{AssetClass: AssetFX, Symbol: "USD/CNY", Currency: "CNY", Metadata: map[string]string{"base_currency": "USD", "quote_currency": "CNY"}},
		{AssetClass: AssetFuture, Symbol: "115.FG608", Currency: "CNY", Metadata: map[string]string{"secid": "115.FG608"}},
		{AssetClass: AssetCrypto, Symbol: "BTC/USDT", Currency: "USD", Metadata: map[string]string{"binance_symbol": "BTCUSDT"}},
	}
	for _, request := range requests {
		if got := len(registry.matchingProviders(request)); got < 2 {
			t.Fatalf("%s has %d providers, want at least 2", request.AssetClass, got)
		}
	}
}
