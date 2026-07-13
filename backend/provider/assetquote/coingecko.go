package assetquote

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultCoinGeckoBaseURL = "https://api.coingecko.com/api/v3/simple/price"

var commonCoinGeckoIDs = map[string]string{
	"BTC": "bitcoin", "ETH": "ethereum", "BNB": "binancecoin", "SOL": "solana",
	"XRP": "ripple", "ADA": "cardano", "DOGE": "dogecoin", "TRX": "tron",
	"AVAX": "avalanche-2", "DOT": "polkadot", "LINK": "chainlink", "LTC": "litecoin",
}

type CoinGeckoProvider struct {
	client  *http.Client
	baseURL string
}

func NewCoinGeckoProvider(client *http.Client) *CoinGeckoProvider {
	return NewCoinGeckoProviderWithBaseURL(client, defaultCoinGeckoBaseURL)
}

func NewCoinGeckoProviderWithBaseURL(client *http.Client, baseURL string) *CoinGeckoProvider {
	if client == nil {
		client = defaultHTTPClient()
	}
	return &CoinGeckoProvider{client: client, baseURL: strings.TrimRight(baseURL, "/")}
}

func (p *CoinGeckoProvider) Name() string { return "coingecko" }
func (p *CoinGeckoProvider) Supports(request QuoteRequest) bool {
	return request.AssetClass == AssetCrypto && coinGeckoID(request) != ""
}

func (p *CoinGeckoProvider) FetchQuotes(ctx context.Context, requests []QuoteRequest) ([]Quote, error) {
	return fetchQuoteBatch(ctx, requests, p.fetchQuote)
}

func (p *CoinGeckoProvider) fetchQuote(ctx context.Context, request QuoteRequest) (Quote, error) {
	id := coinGeckoID(request)
	if id == "" {
		return Quote{}, ErrUnsupported
	}
	quoteCurrency := strings.ToLower(currencyOr(request.Currency, request.Meta("quote_currency")))
	if quoteCurrency == "" || quoteCurrency == "usdt" || quoteCurrency == "usdc" {
		quoteCurrency = "usd"
	}
	query := url.Values{}
	query.Set("ids", id)
	query.Set("vs_currencies", quoteCurrency)
	query.Set("include_24hr_change", "true")
	query.Set("include_last_updated_at", "true")
	var response map[string]map[string]any
	if err := doJSON(ctx, p.client, p.Name(), p.baseURL+"?"+query.Encode(), &response); err != nil {
		return Quote{}, err
	}
	data := response[id]
	price := numberValue(data[quoteCurrency])
	if price <= 0 {
		return Quote{}, &HTTPError{Provider: p.Name(), Message: "quote price is unavailable"}
	}
	changePercent := numberValue(data[quoteCurrency+"_24h_change"])
	asOf := time.Now()
	if unix := int64(numberValue(data["last_updated_at"])); unix > 0 {
		asOf = time.Unix(unix, 0)
	}
	return Quote{
		AssetClass: AssetCrypto, Symbol: request.Symbol, Name: request.Name, Price: price,
		Currency: strings.ToUpper(quoteCurrency), ChangePercent: changePercent,
		AsOf: asOf, FetchedAt: time.Now(), Provider: p.Name(),
	}, nil
}

func coinGeckoID(request QuoteRequest) string {
	if request.CoinGeckoID != "" {
		return strings.ToLower(request.CoinGeckoID)
	}
	if explicit := request.Meta("coingecko_id"); explicit != "" {
		return strings.ToLower(explicit)
	}
	symbol := strings.ToUpper(strings.TrimSpace(request.Symbol))
	for _, separator := range []string{"/", "-", "_"} {
		if index := strings.Index(symbol, separator); index > 0 {
			symbol = symbol[:index]
			break
		}
	}
	for _, quote := range []string{"USDT", "FDUSD", "USDC", "USD", "BTC", "ETH", "BNB", "EUR", "TRY"} {
		if strings.HasSuffix(symbol, quote) && len(symbol) > len(quote) {
			symbol = strings.TrimSuffix(symbol, quote)
			break
		}
	}
	return commonCoinGeckoIDs[symbol]
}
