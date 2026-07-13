package assetquote

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

const defaultSinaFXBaseURL = "https://hq.sinajs.cn/list="

type SinaFXProvider struct {
	client  *http.Client
	baseURL string
}

func NewSinaFXProvider(client *http.Client) *SinaFXProvider {
	return NewSinaFXProviderWithBaseURL(client, defaultSinaFXBaseURL)
}

func NewSinaFXProviderWithBaseURL(client *http.Client, baseURL string) *SinaFXProvider {
	if client == nil {
		client = defaultHTTPClient()
	}
	return &SinaFXProvider{client: client, baseURL: baseURL}
}

func (p *SinaFXProvider) Name() string { return "sina-fx" }

func (p *SinaFXProvider) Supports(request QuoteRequest) bool {
	_, _, ok := currencyPair(request)
	return request.AssetClass == AssetFX && ok
}

func (p *SinaFXProvider) FetchQuotes(ctx context.Context, requests []QuoteRequest) ([]Quote, error) {
	return fetchQuoteBatch(ctx, requests, p.fetchQuote)
}

func (p *SinaFXProvider) fetchQuote(ctx context.Context, request QuoteRequest) (Quote, error) {
	base, quoteCurrency, ok := currencyPair(request)
	if !ok {
		return Quote{}, ErrUnsupported
	}
	symbol := "fx_s" + strings.ToLower(base+quoteCurrency)
	body, err := doRequest(ctx, p.client, p.Name(), sinaQuoteURL(p.baseURL, symbol), map[string]string{"Referer": "https://finance.sina.com.cn/"})
	if err != nil {
		return Quote{}, err
	}
	decoded, _, decodeErr := transform.Bytes(simplifiedchinese.GBK.NewDecoder(), body)
	if decodeErr == nil && bytes.Contains(decoded, []byte("=")) {
		body = decoded
	}
	line := strings.TrimSpace(string(body))
	start, end := strings.Index(line, "\""), strings.LastIndex(line, "\"")
	if start < 0 || end <= start {
		return Quote{}, &HTTPError{Provider: p.Name(), Message: "malformed quote response"}
	}
	fields := strings.Split(line[start+1:end], ",")
	if len(fields) < 9 {
		return Quote{}, &HTTPError{Provider: p.Name(), Message: "quote is unavailable"}
	}
	price := parseFloat(fields, 8)
	if price <= 0 {
		price = parseFloat(fields, 1)
	}
	if price <= 0 {
		return Quote{}, &HTTPError{Provider: p.Name(), Message: "quote price is unavailable"}
	}
	previousClose := parseFloat(fields, 3)
	change := price - previousClose
	changePercent := 0.0
	if previousClose > 0 {
		changePercent = change / previousClose * 100
	}
	return Quote{
		AssetClass: AssetFX, Symbol: base + "/" + quoteCurrency, Name: request.Name,
		Price: price, Currency: quoteCurrency, PreviousClose: previousClose,
		Change: change, ChangePercent: changePercent, AsOf: time.Now(), FetchedAt: time.Now(), Provider: p.Name(),
	}, nil
}
