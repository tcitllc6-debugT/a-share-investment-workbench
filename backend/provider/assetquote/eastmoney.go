package assetquote

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const defaultEastmoneyBaseURL = "https://push2.eastmoney.com/api/qt/stock/get"

type EastmoneyProvider struct {
	client  *http.Client
	baseURL string
}

func NewEastmoneyProvider(client *http.Client) *EastmoneyProvider {
	return NewEastmoneyProviderWithBaseURL(client, defaultEastmoneyBaseURL)
}

func NewEastmoneyProviderWithBaseURL(client *http.Client, baseURL string) *EastmoneyProvider {
	if client == nil {
		client = defaultHTTPClient()
	}
	return &EastmoneyProvider{client: client, baseURL: strings.TrimRight(baseURL, "/")}
}

func (p *EastmoneyProvider) Name() string { return "eastmoney" }

func (p *EastmoneyProvider) Supports(request QuoteRequest) bool {
	switch request.AssetClass {
	case AssetStock, AssetBond, AssetPreciousMetal, AssetFuture:
		return eastmoneySecID(request) != ""
	case AssetFund:
		return !isOTCFund(request) && eastmoneySecID(request) != ""
	default:
		return false
	}
}

func (p *EastmoneyProvider) FetchQuotes(ctx context.Context, requests []QuoteRequest) ([]Quote, error) {
	return fetchQuoteBatch(ctx, requests, p.fetchQuote)
}

func (p *EastmoneyProvider) fetchQuote(ctx context.Context, request QuoteRequest) (Quote, error) {
	secID := eastmoneySecID(request)
	if secID == "" {
		return Quote{}, ErrUnsupported
	}
	query := url.Values{}
	query.Set("secid", secID)
	query.Set("fields", "f57,f58,f43,f47,f48,f59,f60,f170")
	var response struct {
		Data map[string]any `json:"data"`
	}
	if err := doJSON(ctx, p.client, p.Name(), p.baseURL+"?"+query.Encode(), &response); err != nil {
		return Quote{}, err
	}
	if len(response.Data) == 0 {
		return Quote{}, &HTTPError{Provider: p.Name(), Message: "empty quote data"}
	}
	decimalValue, hasDecimals := response.Data["f59"]
	decimals := int(numberValue(decimalValue))
	if !hasDecimals || decimals < 0 || decimals > 8 {
		decimals = 2
	}
	scale := math.Pow10(decimals)
	price := numberValue(response.Data["f43"]) / scale
	previousClose := numberValue(response.Data["f60"]) / scale
	if price <= 0 {
		return Quote{}, &HTTPError{Provider: p.Name(), Message: "quote price is unavailable"}
	}
	change := price - previousClose
	changePercent := numberValue(response.Data["f170"]) / 100
	if changePercent == 0 && previousClose > 0 {
		changePercent = change / previousClose * 100
	}
	now := time.Now()
	return Quote{
		AssetClass: request.AssetClass, Symbol: request.Symbol, Exchange: request.Exchange,
		Name: stringValue(response.Data["f58"]), Price: price, Currency: currencyOr(request.Currency, "CNY"),
		PreviousClose: previousClose, Change: change, ChangePercent: changePercent,
		Volume: numberValue(response.Data["f47"]), Turnover: numberValue(response.Data["f48"]),
		AsOf: now, FetchedAt: now, Provider: p.Name(),
	}, nil
}

func eastmoneySecID(request QuoteRequest) string {
	if explicit := request.Meta("secid"); explicit != "" {
		return explicit
	}
	symbol := strings.TrimSpace(request.Symbol)
	if strings.Contains(symbol, ".") {
		parts := strings.SplitN(symbol, ".", 2)
		if len(parts) == 2 && (parts[0] == "0" || parts[0] == "1" || parts[0] == "116") {
			return symbol
		}
		if len(parts) == 2 {
			code, suffix := strings.TrimSpace(parts[0]), strings.ToUpper(strings.TrimSpace(parts[1]))
			switch suffix {
			case "SH", "SSE":
				return "1." + code
			case "SZ", "SZSE", "BJ", "BSE":
				return "0." + code
			case "HK", "HKEX":
				return "116." + strings.TrimLeft(code, "0")
			}
		}
	}
	symbol = strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(strings.ToUpper(symbol), "SH"), "SZ"), "BJ")
	exchange := strings.ToUpper(strings.TrimSpace(request.Exchange))
	if exchange == "SH" || exchange == "SSE" || exchange == "上海" {
		return "1." + symbol
	}
	if exchange == "SZ" || exchange == "SZSE" || exchange == "BJ" || exchange == "BSE" || exchange == "深圳" || exchange == "北京" {
		return "0." + symbol
	}
	if exchange == "HK" || exchange == "HKEX" || exchange == "香港" {
		return "116." + strings.TrimLeft(symbol, "0")
	}
	if len(symbol) != 6 {
		return ""
	}
	switch symbol[0] {
	case '5', '6', '9':
		return "1." + symbol
	case '0', '1', '2', '3', '4', '8':
		return "0." + symbol
	default:
		return ""
	}
}

func numberValue(value any) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	case string:
		parsed, _ := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		return parsed
	default:
		return 0
	}
}

func stringValue(value any) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func currencyOr(currency, fallback string) string {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		return fallback
	}
	return currency
}
