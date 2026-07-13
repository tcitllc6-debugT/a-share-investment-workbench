package assetquote

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultExchangeRateBaseURL = "https://open.er-api.com/v6/latest"

type ExchangeRateProvider struct {
	client  *http.Client
	baseURL string
}

func NewExchangeRateProvider(client *http.Client) *ExchangeRateProvider {
	return NewExchangeRateProviderWithBaseURL(client, defaultExchangeRateBaseURL)
}

func NewExchangeRateProviderWithBaseURL(client *http.Client, baseURL string) *ExchangeRateProvider {
	if client == nil {
		client = defaultHTTPClient()
	}
	return &ExchangeRateProvider{client: client, baseURL: strings.TrimRight(baseURL, "/")}
}

func (p *ExchangeRateProvider) Name() string { return "exchangerate-api" }

func (p *ExchangeRateProvider) Supports(request QuoteRequest) bool {
	_, _, ok := currencyPair(request)
	return request.AssetClass == AssetFX && ok
}

func (p *ExchangeRateProvider) FetchQuotes(ctx context.Context, requests []QuoteRequest) ([]Quote, error) {
	return fetchQuoteBatch(ctx, requests, p.fetchQuote)
}

func (p *ExchangeRateProvider) fetchQuote(ctx context.Context, request QuoteRequest) (Quote, error) {
	base, quoteCurrency, ok := currencyPair(request)
	if !ok {
		return Quote{}, ErrUnsupported
	}
	var response struct {
		UpdatedAt int64              `json:"time_last_update_unix"`
		Rates     map[string]float64 `json:"rates"`
	}
	if err := doJSON(ctx, p.client, p.Name(), p.baseURL+"/"+url.PathEscape(base), &response); err != nil {
		return Quote{}, err
	}
	rate := response.Rates[quoteCurrency]
	if rate <= 0 {
		return Quote{}, &HTTPError{Provider: p.Name(), Message: "rate is unavailable"}
	}
	asOf := time.Now()
	if response.UpdatedAt > 0 {
		asOf = time.Unix(response.UpdatedAt, 0)
	}
	return Quote{AssetClass: AssetFX, Symbol: base + "/" + quoteCurrency, Name: request.Name, Price: rate, Currency: quoteCurrency, AsOf: asOf, FetchedAt: time.Now(), Provider: p.Name()}, nil
}
