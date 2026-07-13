package assetquote

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegistryFallsBackAfterPreferredProviderFails(t *testing.T) {
	primary := &testProvider{name: "primary", err: ErrUnsupported}
	backup := &testProvider{name: "backup", price: 88.6}
	config := DefaultRegistryConfig()
	config.MaxRetries = 0
	config.MinRequestGap = 0
	registry := NewRegistry(config, primary, backup)

	result, err := registry.Quote(context.Background(), QuoteRequest{AssetClass: AssetPreciousMetal, Symbol: "AU9999", Currency: "CNY", Provider: "primary"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Quote.Provider != "backup" || result.Quote.Price != 88.6 {
		t.Fatalf("expected backup quote, got %+v", result.Quote)
	}
}

func TestTencentProviderQuotesEastmoneyStyleSecurityID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("q") != "sh600703" {
			t.Fatalf("unexpected Tencent symbol: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`v_sh600703="1~三安光电~600703~16.94~17.80~18.10~0~0~0~0~0~0~0~0~0~0~0~0~0~0~0~0~0~0~0~0~0~0~0~~20260710161453~-0.86~-4.83~18.69~16.93~";`))
	}))
	defer server.Close()

	provider := NewTencentProviderWithBaseURL(server.Client(), server.URL)
	quotes, err := provider.FetchQuotes(context.Background(), []QuoteRequest{{AssetClass: AssetStock, Symbol: "1.600703", Currency: "CNY", Metadata: map[string]string{"secid": "1.600703"}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(quotes) != 1 || quotes[0].Price != 16.94 || quotes[0].PreviousClose != 17.80 {
		t.Fatalf("unexpected Tencent quote: %+v", quotes)
	}
}

func TestSinaFXProviderUsesCurrentPairPrice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("list") != "fx_susdcny" {
			t.Fatalf("unexpected Sina FX symbol: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`var hq_str_fx_susdcny="02:56:42,6.7667,6.7952,6.7809,316,6.7953,6.7953,6.7637,6.7809,在岸人民币,0,0,0.0316,来源,0,0,,2026-07-11";`))
	}))
	defer server.Close()

	provider := NewSinaFXProviderWithBaseURL(server.Client(), server.URL)
	quotes, err := provider.FetchQuotes(context.Background(), []QuoteRequest{{AssetClass: AssetFX, Symbol: "USD/CNY", Currency: "CNY"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(quotes) != 1 || quotes[0].Price != 6.7809 {
		t.Fatalf("unexpected Sina FX quote: %+v", quotes)
	}
}

func TestSinaSymbolsDeriveFromStoredEastmoneyIdentifiers(t *testing.T) {
	if got := sinaSymbol(QuoteRequest{AssetClass: AssetStock, Symbol: "1.600703", Metadata: map[string]string{"secid": "1.600703"}}); got != "sh600703" {
		t.Fatalf("stock symbol=%q", got)
	}
	if got := sinaReferenceSymbol(QuoteRequest{AssetClass: AssetPreciousMetal, Symbol: "118.AU9999"}); got != "nf_AU0" {
		t.Fatalf("gold reference symbol=%q", got)
	}
	if got := sinaReferenceSymbol(QuoteRequest{AssetClass: AssetPreciousMetal, Symbol: "118.AG999"}); got != "nf_AG0" {
		t.Fatalf("silver reference symbol=%q", got)
	}
}

func TestSinaReferenceProviderFallsBackToGoldMainContract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("list") != "nf_AU0" {
			t.Fatalf("unexpected Sina reference symbol: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`var hq_str_nf_AU0="黄金连续,023000,895.000,898.480,889.800,0.000,896.020,896.380,896.380,0.000,901.420,1,1,121960.000,98735,沪,黄金,2026-07-11,1";`))
	}))
	defer server.Close()

	provider := NewSinaReferenceProviderWithBaseURL(server.Client(), server.URL)
	quotes, err := provider.FetchQuotes(context.Background(), []QuoteRequest{{AssetClass: AssetPreciousMetal, Symbol: "AU9999", Currency: "CNY"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(quotes) != 1 || quotes[0].Price != 896.38 {
		t.Fatalf("unexpected gold reference quote: %+v", quotes)
	}
}

func TestSinaReferenceProviderConvertsSilverMainContractToGram(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("list") != "nf_AG0" {
			t.Fatalf("unexpected Sina reference symbol: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`var hq_str_nf_AG0="白银连续,023000,14537.000,14660.000,14387.000,0.000,14573.000,14589.000,14589.000,0.000,14674.000,1,5,195687.000,261773,沪,白银,2026-07-11,1";`))
	}))
	defer server.Close()

	provider := NewSinaReferenceProviderWithBaseURL(server.Client(), server.URL)
	quotes, err := provider.FetchQuotes(context.Background(), []QuoteRequest{{AssetClass: AssetPreciousMetal, Symbol: "AG999", Currency: "CNY", Metadata: map[string]string{"unit": "克"}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(quotes) != 1 || quotes[0].Price != 14.589 {
		t.Fatalf("unexpected silver reference quote: %+v", quotes)
	}
}

func TestExchangeRateProviderUsesPublicBackupRate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/USD" {
			t.Fatalf("unexpected exchange-rate path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"time_last_update_unix":1783814551,"rates":{"CNY":6.78999}}`))
	}))
	defer server.Close()

	provider := NewExchangeRateProviderWithBaseURL(server.Client(), server.URL)
	quotes, err := provider.FetchQuotes(context.Background(), []QuoteRequest{{AssetClass: AssetFX, Symbol: "USD/CNY", Currency: "CNY"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(quotes) != 1 || quotes[0].Price != 6.78999 {
		t.Fatalf("unexpected exchange-rate quote: %+v", quotes)
	}
}
