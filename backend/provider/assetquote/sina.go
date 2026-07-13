package assetquote

import (
	"bytes"
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

const defaultSinaBaseURL = "https://hq.sinajs.cn/list="

type SinaProvider struct {
	client  *http.Client
	baseURL string
}

func NewSinaProvider(client *http.Client) *SinaProvider {
	return NewSinaProviderWithBaseURL(client, defaultSinaBaseURL)
}

func NewSinaProviderWithBaseURL(client *http.Client, baseURL string) *SinaProvider {
	if client == nil {
		client = defaultHTTPClient()
	}
	return &SinaProvider{client: client, baseURL: baseURL}
}

func (p *SinaProvider) Name() string { return "sina" }

func (p *SinaProvider) Supports(request QuoteRequest) bool {
	switch request.AssetClass {
	case AssetStock, AssetFund, AssetBond, AssetPreciousMetal, AssetFuture:
		return sinaSymbol(request) != ""
	default:
		return false
	}
}

func (p *SinaProvider) FetchQuotes(ctx context.Context, requests []QuoteRequest) ([]Quote, error) {
	return fetchQuoteBatch(ctx, requests, p.fetchQuote)
}

func (p *SinaProvider) fetchQuote(ctx context.Context, request QuoteRequest) (Quote, error) {
	symbol := sinaSymbol(request)
	if symbol == "" {
		return Quote{}, ErrUnsupported
	}
	return fetchSinaQuote(ctx, p.client, p.baseURL, p.Name(), request, symbol)
}

type SinaReferenceProvider struct {
	client  *http.Client
	baseURL string
}

func NewSinaReferenceProvider(client *http.Client) *SinaReferenceProvider {
	return NewSinaReferenceProviderWithBaseURL(client, defaultSinaBaseURL)
}

func NewSinaReferenceProviderWithBaseURL(client *http.Client, baseURL string) *SinaReferenceProvider {
	if client == nil {
		client = defaultHTTPClient()
	}
	return &SinaReferenceProvider{client: client, baseURL: baseURL}
}

func (p *SinaReferenceProvider) Name() string { return "sina-reference" }

func (p *SinaReferenceProvider) Supports(request QuoteRequest) bool {
	return sinaReferenceSymbol(request) != ""
}

func (p *SinaReferenceProvider) FetchQuotes(ctx context.Context, requests []QuoteRequest) ([]Quote, error) {
	return fetchQuoteBatch(ctx, requests, p.fetchQuote)
}

func (p *SinaReferenceProvider) fetchQuote(ctx context.Context, request QuoteRequest) (Quote, error) {
	symbol := sinaReferenceSymbol(request)
	if symbol == "" {
		return Quote{}, ErrUnsupported
	}
	parseRequest := request
	parseRequest.AssetClass = AssetFuture
	quote, err := fetchSinaQuote(ctx, p.client, p.baseURL, p.Name(), parseRequest, symbol)
	if err != nil {
		return Quote{}, err
	}
	if multiplier := sinaReferencePriceMultiplier(request, symbol); multiplier != 1 {
		quote.Price *= multiplier
		quote.PreviousClose *= multiplier
		quote.Change *= multiplier
	}
	quote.AssetClass, quote.Symbol, quote.Name = request.AssetClass, request.Symbol, request.Name
	return quote, nil
}

func fetchSinaQuote(ctx context.Context, client *http.Client, baseURL, provider string, request QuoteRequest, symbol string) (Quote, error) {
	body, err := doRequest(ctx, client, provider, sinaQuoteURL(baseURL, symbol), map[string]string{"Referer": "https://finance.sina.com.cn/"})
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
		return Quote{}, &HTTPError{Provider: provider, Message: "malformed quote response"}
	}
	fields := strings.Split(line[start+1:end], ",")
	if len(fields) < 10 || strings.TrimSpace(fields[0]) == "" {
		return Quote{}, &HTTPError{Provider: provider, Message: "quote is unavailable"}
	}
	defaultPriceIndex, defaultPreviousCloseIndex := 3, 2
	if request.AssetClass == AssetFuture {
		defaultPriceIndex, defaultPreviousCloseIndex = 8, 5
	}
	priceIndex := metaIndex(request, "price_index", defaultPriceIndex)
	previousCloseIndex := metaIndex(request, "previous_close_index", defaultPreviousCloseIndex)
	price := parseFloat(fields, priceIndex)
	previousClose := parseFloat(fields, previousCloseIndex)
	if price <= 0 {
		price = parseFloat(fields, metaIndex(request, "open_index", 1))
	}
	if price <= 0 {
		return Quote{}, &HTTPError{Provider: provider, Message: "quote price is unavailable"}
	}
	change := price - previousClose
	changePercent := 0.0
	if previousClose > 0 {
		changePercent = change / previousClose * 100
	}
	asOf := time.Now()
	if len(fields) > 31 {
		if parsed, parseErr := time.ParseInLocation("2006-01-02 15:04:05", fields[30]+" "+fields[31], time.Local); parseErr == nil {
			asOf = parsed
		}
	}
	return Quote{
		AssetClass: request.AssetClass, Symbol: request.Symbol, Exchange: request.Exchange,
		Name: strings.TrimSpace(fields[0]), Price: price, Currency: currencyOr(request.Currency, "CNY"),
		PreviousClose: previousClose, Change: change, ChangePercent: changePercent,
		Volume: parseFloat(fields, metaIndex(request, "volume_index", 8)), Turnover: parseFloat(fields, metaIndex(request, "turnover_index", 9)), AsOf: asOf,
		FetchedAt: time.Now(), Provider: provider,
	}, nil
}

func sinaQuoteURL(baseURL, symbol string) string {
	escaped := url.QueryEscape(symbol)
	if strings.Contains(baseURL, "?") || strings.HasSuffix(baseURL, "=") {
		return baseURL + escaped
	}
	return strings.TrimRight(baseURL, "/") + "?list=" + escaped
}

func sinaSymbol(request QuoteRequest) string {
	if explicit := request.Meta("sina_symbol"); explicit != "" {
		return explicit
	}
	if symbol := mainlandQuoteSymbol(request); symbol != "" {
		return symbol
	}
	symbol := strings.ToLower(strings.TrimSpace(request.Symbol))
	if strings.HasPrefix(symbol, "sh") || strings.HasPrefix(symbol, "sz") || strings.HasPrefix(symbol, "bj") || strings.HasPrefix(symbol, "hk") || strings.HasPrefix(symbol, "gb_") || strings.HasPrefix(symbol, "nf_") || strings.HasPrefix(symbol, "hf_") {
		return symbol
	}
	exchange := strings.ToUpper(strings.TrimSpace(request.Exchange))
	if exchange == "SH" || exchange == "SSE" {
		return "sh" + symbol
	}
	if exchange == "SZ" || exchange == "SZSE" {
		return "sz" + symbol
	}
	if exchange == "BJ" || exchange == "BSE" {
		return "bj" + symbol
	}
	if len(symbol) == 6 {
		switch symbol[0] {
		case '5', '6', '9':
			return "sh" + symbol
		case '0', '1', '2', '3':
			return "sz" + symbol
		case '4', '8':
			return "bj" + symbol
		}
	}
	return ""
}

func sinaReferenceSymbol(request QuoteRequest) string {
	if explicit := request.Meta("sina_reference_symbol"); explicit != "" {
		return strings.ToLower(explicit)
	}
	symbol := strings.ToUpper(strings.TrimSpace(request.Symbol))
	if parts := strings.SplitN(symbol, ".", 2); len(parts) == 2 {
		symbol = parts[1]
	}
	switch request.AssetClass {
	case AssetPreciousMetal:
		switch symbol {
		case "AU9999", "AU99.99", "AU9995":
			return "nf_AU0"
		case "AGTD", "AG999", "AG9999", "AG99.9":
			return "nf_AG0"
		}
	case AssetFuture:
		letters := strings.Builder{}
		for _, character := range symbol {
			if character >= 'A' && character <= 'Z' {
				letters.WriteRune(character)
			} else {
				break
			}
		}
		if letters.Len() > 0 {
			return "nf_" + letters.String() + "0"
		}
	}
	return ""
}

func sinaReferencePriceMultiplier(request QuoteRequest, referenceSymbol string) float64 {
	unit := strings.ToLower(strings.TrimSpace(request.Meta("unit")))
	if strings.EqualFold(referenceSymbol, "nf_AG0") && (unit == "克" || unit == "g" || unit == "gram") {
		return 0.001
	}
	return 1
}

func mainlandQuoteSymbol(request QuoteRequest) string {
	value := strings.ToUpper(strings.TrimSpace(request.Meta("sina_symbol")))
	if value != "" {
		return strings.ToLower(value)
	}
	for _, candidate := range []string{request.Symbol, request.Meta("secid"), request.Meta("quoteId")} {
		value = strings.ToUpper(strings.TrimSpace(candidate))
		if value == "" {
			continue
		}
		if strings.HasPrefix(value, "SH") || strings.HasPrefix(value, "SZ") || strings.HasPrefix(value, "BJ") || strings.HasPrefix(value, "HK") {
			return strings.ToLower(value)
		}
		if parts := strings.SplitN(value, ".", 2); len(parts) == 2 {
			market, code := parts[0], strings.TrimSpace(parts[1])
			switch market {
			case "1":
				return "sh" + strings.ToLower(code)
			case "0":
				if strings.HasPrefix(code, "4") || strings.HasPrefix(code, "8") {
					return "bj" + strings.ToLower(code)
				}
				return "sz" + strings.ToLower(code)
			case "116":
				return "hk" + strings.TrimLeft(strings.ToLower(code), "0")
			case "SH", "SSE":
				return "sh" + strings.ToLower(code)
			case "SZ", "SZSE":
				return "sz" + strings.ToLower(code)
			case "BJ", "BSE":
				return "bj" + strings.ToLower(code)
			}
		}
	}
	return ""
}

func parseFloat(fields []string, index int) float64 {
	if index < 0 || index >= len(fields) {
		return 0
	}
	value, err := strconv.ParseFloat(strings.TrimSpace(fields[index]), 64)
	if err != nil {
		return 0
	}
	return value
}

func metaIndex(request QuoteRequest, key string, fallback int) int {
	value := request.Meta(key)
	if value == "" {
		return fallback
	}
	index, err := strconv.Atoi(value)
	if err != nil || index < 0 {
		return fallback
	}
	return index
}
