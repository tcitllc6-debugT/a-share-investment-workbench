package assetquote

import (
	"context"
	"errors"
	"testing"
	"time"
)

type testProvider struct {
	name  string
	price float64
	err   error
	calls int
}

type delayedProvider struct {
	delay time.Duration
}

func (p *delayedProvider) Name() string               { return "delayed" }
func (p *delayedProvider) Supports(QuoteRequest) bool { return true }
func (p *delayedProvider) FetchQuotes(ctx context.Context, requests []QuoteRequest) ([]Quote, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(p.delay):
	}
	return []Quote{{Price: 1, Currency: "CNY", AsOf: time.Now()}}, nil
}

func (p *testProvider) Name() string               { return p.name }
func (p *testProvider) Supports(QuoteRequest) bool { return true }
func (p *testProvider) FetchQuotes(_ context.Context, requests []QuoteRequest) ([]Quote, error) {
	p.calls++
	if p.err != nil {
		return nil, p.err
	}
	return []Quote{{Price: p.price, Currency: "CNY", AsOf: time.Now()}}, nil
}

func TestRegistryFallsBackAndCaches(t *testing.T) {
	bad := &testProvider{name: "bad", err: errors.New("offline")}
	good := &testProvider{name: "good", price: 12.5}
	config := DefaultRegistryConfig()
	config.MaxRetries = 0
	config.MinRequestGap = 0
	registry := NewRegistry(config, bad, good)
	request := QuoteRequest{AssetClass: AssetStock, Symbol: "600000", Currency: "CNY"}
	first, err := registry.Quote(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if first.Quote.Price != 12.5 || first.Quote.Provider != "good" {
		t.Fatalf("unexpected quote: %+v", first)
	}
	second, err := registry.Quote(context.Background(), request)
	if err != nil || !second.Cached {
		t.Fatalf("expected cached quote: %+v %v", second, err)
	}
	if good.calls != 1 {
		t.Fatalf("provider calls=%d", good.calls)
	}
}

func TestRegistryReturnsStaleValueAfterProviderFailure(t *testing.T) {
	provider := &testProvider{name: "only", price: 8}
	config := DefaultRegistryConfig()
	config.MaxRetries = 0
	config.MinRequestGap = 0
	config.CacheTTL = time.Millisecond
	config.StaleTTL = time.Hour
	registry := NewRegistry(config, provider)
	request := QuoteRequest{AssetClass: AssetStock, Symbol: "000001", Currency: "CNY"}
	if _, err := registry.Quote(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	time.Sleep(2 * time.Millisecond)
	provider.err = errors.New("offline")
	result, err := registry.Quote(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Stale || !result.Cached {
		t.Fatalf("expected stale cache: %+v", result)
	}
}

func TestEastmoneySecIDAcceptsCommonDesktopSymbols(t *testing.T) {
	tests := map[string]string{"600519.SH": "1.600519", "300418.SZ": "0.300418", "430047.BJ": "0.430047", "00700.HK": "116.700"}
	for symbol, want := range tests {
		if got := eastmoneySecID(QuoteRequest{AssetClass: AssetStock, Symbol: symbol}); got != want {
			t.Fatalf("%s => %s, want %s", symbol, got, want)
		}
	}
}

func TestDefaultRegistryAllowsSlowPreciousMetalFeed(t *testing.T) {
	config := DefaultRegistryConfig()
	if config.RequestTimeout < 10*time.Second {
		t.Fatalf("precious metal quote timeout=%s, want at least 10s", config.RequestTimeout)
	}
}

func TestRegistryQuotesRefreshesIndependentSymbolsConcurrently(t *testing.T) {
	config := DefaultRegistryConfig()
	config.MaxRetries = 0
	config.MinRequestGap = 0
	config.MaxConcurrency = 4
	registry := NewRegistry(config, &delayedProvider{delay: 100 * time.Millisecond})
	requests := []QuoteRequest{
		{AssetClass: AssetStock, Symbol: "600000", Currency: "CNY"},
		{AssetClass: AssetStock, Symbol: "600001", Currency: "CNY"},
		{AssetClass: AssetStock, Symbol: "600002", Currency: "CNY"},
		{AssetClass: AssetStock, Symbol: "600003", Currency: "CNY"},
	}
	started := time.Now()
	results, err := registry.Quotes(context.Background(), requests)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != len(requests) {
		t.Fatalf("results = %d", len(results))
	}
	if elapsed := time.Since(started); elapsed >= 250*time.Millisecond {
		t.Fatalf("batch quote took %s; requests appear sequential", elapsed)
	}
}
