package assetquote

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

type AssetClass string

const (
	AssetStock         AssetClass = "stock"
	AssetFund          AssetClass = "fund"
	AssetBond          AssetClass = "bond"
	AssetPreciousMetal AssetClass = "precious_metal"
	AssetFX            AssetClass = "fx"
	AssetFuture        AssetClass = "future"
	AssetCrypto        AssetClass = "crypto"
)

var (
	ErrUnsupported    = errors.New("quote provider does not support request")
	ErrInvalidRequest = errors.New("invalid quote request")
)

type QuoteRequest struct {
	AssetClass  AssetClass        `json:"assetClass"`
	Symbol      string            `json:"symbol"`
	Name        string            `json:"name,omitempty"`
	Exchange    string            `json:"exchange,omitempty"`
	Currency    string            `json:"currency,omitempty"`
	Provider    string            `json:"provider,omitempty"`
	CoinGeckoID string            `json:"coinGeckoId,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

func (r QuoteRequest) Normalized() QuoteRequest {
	r.AssetClass = AssetClass(strings.ToLower(strings.TrimSpace(string(r.AssetClass))))
	r.Symbol = strings.ToUpper(strings.TrimSpace(r.Symbol))
	r.Name = strings.TrimSpace(r.Name)
	r.Exchange = strings.ToUpper(strings.TrimSpace(r.Exchange))
	r.Currency = strings.ToUpper(strings.TrimSpace(r.Currency))
	r.Provider = strings.ToLower(strings.TrimSpace(r.Provider))
	r.CoinGeckoID = strings.ToLower(strings.TrimSpace(r.CoinGeckoID))
	return r
}

func (r QuoteRequest) Validate() error {
	if strings.TrimSpace(r.Symbol) == "" {
		return fmt.Errorf("%w: symbol is empty", ErrInvalidRequest)
	}
	if strings.TrimSpace(string(r.AssetClass)) == "" {
		return fmt.Errorf("%w: asset class is empty", ErrInvalidRequest)
	}
	return nil
}

func (r QuoteRequest) CacheKey() string {
	r = r.Normalized()
	return strings.Join([]string{
		string(r.AssetClass), r.Exchange, r.Symbol, r.Currency, r.Provider,
		strings.ToUpper(r.Meta("base_currency")), strings.ToUpper(r.Meta("quote_currency")),
		r.Meta("secid"), strings.ToLower(r.Meta("sina_symbol")),
		strings.ToUpper(r.Meta("binance_symbol")), strings.ToLower(r.Meta("coingecko_id")),
	}, "|")
}

func (r QuoteRequest) Meta(key string) string {
	if r.Metadata == nil {
		return ""
	}
	return strings.TrimSpace(r.Metadata[key])
}

type Quote struct {
	AssetClass    AssetClass `json:"assetClass"`
	Symbol        string     `json:"symbol"`
	Name          string     `json:"name,omitempty"`
	Exchange      string     `json:"exchange,omitempty"`
	Price         float64    `json:"price"`
	Currency      string     `json:"currency"`
	PreviousClose float64    `json:"previousClose,omitempty"`
	Change        float64    `json:"change,omitempty"`
	ChangePercent float64    `json:"changePercent,omitempty"`
	Volume        float64    `json:"volume,omitempty"`
	Turnover      float64    `json:"turnover,omitempty"`
	AsOf          time.Time  `json:"asOf"`
	FetchedAt     time.Time  `json:"fetchedAt"`
	Provider      string     `json:"provider"`
	Stale         bool       `json:"stale"`
}

type HealthState string

const (
	HealthHealthy  HealthState = "healthy"
	HealthDegraded HealthState = "degraded"
	HealthOpen     HealthState = "open"
	HealthUnknown  HealthState = "unknown"
)

type ProviderHealth struct {
	Name                string      `json:"name"`
	State               HealthState `json:"state"`
	ConsecutiveFailures int         `json:"consecutiveFailures"`
	CircuitOpenUntil    time.Time   `json:"circuitOpenUntil,omitempty"`
	LastSuccessAt       time.Time   `json:"lastSuccessAt,omitempty"`
	LastFailureAt       time.Time   `json:"lastFailureAt,omitempty"`
	LastError           string      `json:"lastError,omitempty"`
	LastLatencyMs       int64       `json:"lastLatencyMs"`
	StaleHits           uint64      `json:"staleHits"`
}

type QuoteResult struct {
	Quote    Quote          `json:"quote"`
	Health   ProviderHealth `json:"health"`
	Cached   bool           `json:"cached"`
	Stale    bool           `json:"stale"`
	Attempts int            `json:"attempts"`
}

type QuoteProvider interface {
	Name() string
	Supports(QuoteRequest) bool
	FetchQuotes(context.Context, []QuoteRequest) ([]Quote, error)
}

type HTTPError struct {
	Provider   string
	StatusCode int
	Retryable  bool
	Message    string
}

func (e *HTTPError) Error() string {
	if e.StatusCode == 0 {
		return fmt.Sprintf("%s: %s", e.Provider, e.Message)
	}
	return fmt.Sprintf("%s: HTTP %d: %s", e.Provider, e.StatusCode, e.Message)
}

func isRetryable(err error) bool {
	if err == nil || errors.Is(err, ErrUnsupported) || errors.Is(err, ErrInvalidRequest) || errors.Is(err, context.Canceled) {
		return false
	}
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.Retryable
	}
	return true
}
