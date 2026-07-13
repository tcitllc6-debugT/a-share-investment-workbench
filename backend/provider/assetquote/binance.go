package assetquote

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const defaultBinanceBaseURL = "https://data-api.binance.vision/api/v3/ticker/24hr"

type BinanceProvider struct {
	client  *http.Client
	baseURL string
}

func NewBinanceProvider(client *http.Client) *BinanceProvider {
	return NewBinanceProviderWithBaseURL(client, defaultBinanceBaseURL)
}

func NewBinanceProviderWithBaseURL(client *http.Client, baseURL string) *BinanceProvider {
	if client == nil {
		client = defaultHTTPClient()
	}
	return &BinanceProvider{client: client, baseURL: strings.TrimRight(baseURL, "/")}
}

func (p *BinanceProvider) Name() string { return "binance" }
func (p *BinanceProvider) Supports(request QuoteRequest) bool {
	return request.AssetClass == AssetCrypto && binanceSymbol(request) != ""
}

func (p *BinanceProvider) FetchQuotes(ctx context.Context, requests []QuoteRequest) ([]Quote, error) {
	return fetchQuoteBatch(ctx, requests, p.fetchQuote)
}

func (p *BinanceProvider) fetchQuote(ctx context.Context, request QuoteRequest) (Quote, error) {
	symbol := binanceSymbol(request)
	if symbol == "" {
		return Quote{}, ErrUnsupported
	}
	query := url.Values{"symbol": []string{symbol}}
	var response struct {
		Symbol             string `json:"symbol"`
		LastPrice          string `json:"lastPrice"`
		PrevClosePrice     string `json:"prevClosePrice"`
		PriceChange        string `json:"priceChange"`
		PriceChangePercent string `json:"priceChangePercent"`
		Volume             string `json:"volume"`
		QuoteVolume        string `json:"quoteVolume"`
		CloseTime          int64  `json:"closeTime"`
	}
	if err := doJSON(ctx, p.client, p.Name(), p.baseURL+"?"+query.Encode(), &response); err != nil {
		return Quote{}, err
	}
	price, _ := strconv.ParseFloat(response.LastPrice, 64)
	if price <= 0 {
		return Quote{}, &HTTPError{Provider: p.Name(), Message: "quote price is unavailable"}
	}
	previousClose, _ := strconv.ParseFloat(response.PrevClosePrice, 64)
	change, _ := strconv.ParseFloat(response.PriceChange, 64)
	changePercent, _ := strconv.ParseFloat(response.PriceChangePercent, 64)
	volume, _ := strconv.ParseFloat(response.Volume, 64)
	turnover, _ := strconv.ParseFloat(response.QuoteVolume, 64)
	asOf := time.Now()
	if response.CloseTime > 0 {
		asOf = time.UnixMilli(response.CloseTime)
	}
	return Quote{
		AssetClass: AssetCrypto, Symbol: request.Symbol, Name: request.Name, Price: price,
		Currency: cryptoQuoteCurrency(request, symbol), PreviousClose: previousClose,
		Change: change, ChangePercent: changePercent, Volume: volume, Turnover: turnover,
		AsOf: asOf, FetchedAt: time.Now(), Provider: p.Name(),
	}, nil
}

func binanceSymbol(request QuoteRequest) string {
	if explicit := request.Meta("binance_symbol"); explicit != "" {
		return strings.ToUpper(strings.TrimSpace(explicit))
	}
	symbol := strings.ToUpper(strings.TrimSpace(request.Symbol))
	for _, separator := range []string{"/", "-", "_"} {
		symbol = strings.ReplaceAll(symbol, separator, "")
	}
	if symbol == "" {
		return ""
	}
	for _, quote := range []string{"USDT", "FDUSD", "USDC", "BTC", "ETH", "BNB", "EUR", "TRY"} {
		if strings.HasSuffix(symbol, quote) && len(symbol) > len(quote) {
			return symbol
		}
	}
	quote := strings.ToUpper(strings.TrimSpace(request.Meta("quote_currency")))
	if quote == "" {
		quote = "USDT"
	}
	return symbol + quote
}

func cryptoQuoteCurrency(request QuoteRequest, symbol string) string {
	if quote := strings.ToUpper(strings.TrimSpace(request.Meta("quote_currency"))); quote != "" {
		return quote
	}
	for _, quote := range []string{"USDT", "FDUSD", "USDC", "BTC", "ETH", "BNB", "EUR", "TRY"} {
		if strings.HasSuffix(symbol, quote) {
			return quote
		}
	}
	return currencyOr(request.Currency, "USD")
}
