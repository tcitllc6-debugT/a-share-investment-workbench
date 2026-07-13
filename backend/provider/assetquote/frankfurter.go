package assetquote

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultFrankfurterBaseURL = "https://api.frankfurter.dev/v2/rate"

type FrankfurterProvider struct {
	client  *http.Client
	baseURL string
}

func NewFrankfurterProvider(client *http.Client) *FrankfurterProvider {
	return NewFrankfurterProviderWithBaseURL(client, defaultFrankfurterBaseURL)
}

func NewFrankfurterProviderWithBaseURL(client *http.Client, baseURL string) *FrankfurterProvider {
	if client == nil {
		client = defaultHTTPClient()
	}
	return &FrankfurterProvider{client: client, baseURL: strings.TrimRight(baseURL, "/")}
}

func (p *FrankfurterProvider) Name() string { return "frankfurter" }
func (p *FrankfurterProvider) Supports(request QuoteRequest) bool {
	_, _, ok := currencyPair(request)
	return request.AssetClass == AssetFX && ok
}

func (p *FrankfurterProvider) FetchQuotes(ctx context.Context, requests []QuoteRequest) ([]Quote, error) {
	return fetchQuoteBatch(ctx, requests, p.fetchQuote)
}

func (p *FrankfurterProvider) fetchQuote(ctx context.Context, request QuoteRequest) (Quote, error) {
	base, quoteCurrency, ok := currencyPair(request)
	if !ok {
		return Quote{}, ErrUnsupported
	}
	endpoint := p.baseURL + "/" + url.PathEscape(base) + "/" + url.PathEscape(quoteCurrency)
	var response map[string]any
	if err := doJSON(ctx, p.client, p.Name(), endpoint, &response); err != nil {
		return Quote{}, err
	}
	rate := numberValue(response["rate"])
	if rate <= 0 {
		return Quote{}, &HTTPError{Provider: p.Name(), Message: "rate is unavailable"}
	}
	asOf := time.Now()
	if date := stringValue(response["date"]); date != "" {
		if parsed, err := time.Parse("2006-01-02", date); err == nil {
			asOf = parsed
		}
	}
	return Quote{
		AssetClass: AssetFX, Symbol: base + "/" + quoteCurrency, Name: request.Name,
		Price: rate, Currency: quoteCurrency, AsOf: asOf, FetchedAt: time.Now(), Provider: p.Name(),
	}, nil
}

func currencyPair(request QuoteRequest) (string, string, bool) {
	symbol := strings.ToUpper(strings.TrimSpace(request.Symbol))
	for _, separator := range []string{"/", "-", "_"} {
		symbol = strings.ReplaceAll(symbol, separator, "")
	}
	base := strings.ToUpper(strings.TrimSpace(request.Meta("base_currency")))
	quote := strings.ToUpper(strings.TrimSpace(request.Meta("quote_currency")))
	if base == "" && len(symbol) == 6 {
		base = symbol[:3]
	}
	if quote == "" && len(symbol) == 6 {
		quote = symbol[3:]
	}
	if quote == "" {
		quote = strings.ToUpper(strings.TrimSpace(request.Currency))
	}
	if len(base) != 3 || len(quote) != 3 || base == quote {
		return "", "", false
	}
	return base, quote, true
}
