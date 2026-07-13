package assetquote

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultTencentBaseURL = "https://qt.gtimg.cn/q"

type TencentProvider struct {
	client  *http.Client
	baseURL string
}

func NewTencentProvider(client *http.Client) *TencentProvider {
	return NewTencentProviderWithBaseURL(client, defaultTencentBaseURL)
}

func NewTencentProviderWithBaseURL(client *http.Client, baseURL string) *TencentProvider {
	if client == nil {
		client = defaultHTTPClient()
	}
	return &TencentProvider{client: client, baseURL: strings.TrimRight(baseURL, "/")}
}

func (p *TencentProvider) Name() string { return "tencent" }

func (p *TencentProvider) Supports(request QuoteRequest) bool {
	switch request.AssetClass {
	case AssetStock, AssetFund, AssetBond:
		return mainlandQuoteSymbol(request) != ""
	default:
		return false
	}
}

func (p *TencentProvider) FetchQuotes(ctx context.Context, requests []QuoteRequest) ([]Quote, error) {
	return fetchQuoteBatch(ctx, requests, p.fetchQuote)
}

func (p *TencentProvider) fetchQuote(ctx context.Context, request QuoteRequest) (Quote, error) {
	symbol := mainlandQuoteSymbol(request)
	if symbol == "" {
		return Quote{}, ErrUnsupported
	}
	body, err := doRequest(ctx, p.client, p.Name(), p.baseURL+"?"+url.Values{"q": []string{symbol}}.Encode(), nil)
	if err != nil {
		return Quote{}, err
	}
	line := strings.TrimSpace(string(body))
	start, end := strings.Index(line, "\""), strings.LastIndex(line, "\"")
	if start < 0 || end <= start {
		return Quote{}, &HTTPError{Provider: p.Name(), Message: "malformed quote response"}
	}
	fields := strings.Split(line[start+1:end], "~")
	if len(fields) < 5 || strings.TrimSpace(fields[1]) == "" {
		return Quote{}, &HTTPError{Provider: p.Name(), Message: "quote is unavailable"}
	}
	price := parseFloat(fields, 3)
	if price <= 0 {
		return Quote{}, &HTTPError{Provider: p.Name(), Message: "quote price is unavailable"}
	}
	previousClose := parseFloat(fields, 4)
	change, changePercent := parseFloat(fields, 31), parseFloat(fields, 32)
	if change == 0 && previousClose > 0 {
		change = price - previousClose
	}
	if changePercent == 0 && previousClose > 0 {
		changePercent = change / previousClose * 100
	}
	return Quote{
		AssetClass: request.AssetClass, Symbol: request.Symbol, Name: strings.TrimSpace(fields[1]),
		Exchange: request.Exchange, Price: price, Currency: currencyOr(request.Currency, "CNY"),
		PreviousClose: previousClose, Change: change, ChangePercent: changePercent,
		Volume: parseFloat(fields, 6), Turnover: parseFloat(fields, 37),
		AsOf: time.Now(), FetchedAt: time.Now(), Provider: p.Name(),
	}, nil
}
