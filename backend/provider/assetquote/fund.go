package assetquote

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultFundEstimateBaseURL = "https://fundgz.1234567.com.cn/js"
	defaultFundNAVBaseURL      = "https://api.fund.eastmoney.com/f10/lsjz"
	defaultSinaFundBaseURL     = "https://hq.sinajs.cn/list="
)

type FundEstimateProvider struct {
	client  *http.Client
	baseURL string
}

func NewFundEstimateProvider(client *http.Client) *FundEstimateProvider {
	return NewFundEstimateProviderWithBaseURL(client, defaultFundEstimateBaseURL)
}

func NewFundEstimateProviderWithBaseURL(client *http.Client, baseURL string) *FundEstimateProvider {
	if client == nil {
		client = defaultHTTPClient()
	}
	return &FundEstimateProvider{client: client, baseURL: strings.TrimRight(baseURL, "/")}
}

func (p *FundEstimateProvider) Name() string { return "fund-estimate" }

func (p *FundEstimateProvider) Supports(request QuoteRequest) bool {
	return request.AssetClass == AssetFund && isOTCFund(request)
}

func (p *FundEstimateProvider) FetchQuotes(ctx context.Context, requests []QuoteRequest) ([]Quote, error) {
	return fetchQuoteBatch(ctx, requests, p.fetchQuote)
}

func (p *FundEstimateProvider) fetchQuote(ctx context.Context, request QuoteRequest) (Quote, error) {
	code := fundCode(request)
	if code == "" {
		return Quote{}, ErrUnsupported
	}
	body, err := doRequest(ctx, p.client, p.Name(), p.baseURL+"/"+url.PathEscape(code)+".js", map[string]string{"Referer": "https://fund.eastmoney.com/"})
	if err != nil {
		return Quote{}, err
	}
	line := strings.TrimSpace(string(body))
	start, end := strings.Index(line, "("), strings.LastIndex(line, ")")
	if start < 0 || end <= start {
		return Quote{}, &HTTPError{Provider: p.Name(), Message: "malformed fund estimate response"}
	}
	var response struct {
		Name       string `json:"name"`
		NAVDate    string `json:"jzrq"`
		NAV        string `json:"dwjz"`
		Estimate   string `json:"gsz"`
		EstimatePC string `json:"gszzl"`
		EstimateAt string `json:"gztime"`
	}
	if err := json.Unmarshal([]byte(line[start+1:end]), &response); err != nil {
		return Quote{}, &HTTPError{Provider: p.Name(), Message: "decode fund estimate: " + err.Error()}
	}
	price := parseStringFloat(response.Estimate)
	if price <= 0 {
		return Quote{}, &HTTPError{Provider: p.Name(), Message: "fund estimate is unavailable"}
	}
	previous := parseStringFloat(response.NAV)
	asOf := parseFundTime(response.EstimateAt, response.NAVDate)
	return Quote{
		AssetClass: AssetFund, Symbol: request.Symbol, Name: strings.TrimSpace(response.Name), Price: price,
		Currency: currencyOr(request.Currency, "CNY"), PreviousClose: previous,
		Change: price - previous, ChangePercent: parseStringFloat(response.EstimatePC),
		AsOf: asOf, FetchedAt: time.Now(), Provider: p.Name(),
	}, nil
}

type FundNAVProvider struct {
	client  *http.Client
	baseURL string
}

func NewFundNAVProvider(client *http.Client) *FundNAVProvider {
	return NewFundNAVProviderWithBaseURL(client, defaultFundNAVBaseURL)
}

func NewFundNAVProviderWithBaseURL(client *http.Client, baseURL string) *FundNAVProvider {
	if client == nil {
		client = defaultHTTPClient()
	}
	return &FundNAVProvider{client: client, baseURL: strings.TrimRight(baseURL, "/")}
}

func (p *FundNAVProvider) Name() string { return "fund-nav" }

func (p *FundNAVProvider) Supports(request QuoteRequest) bool {
	return request.AssetClass == AssetFund && isOTCFund(request)
}

func (p *FundNAVProvider) FetchQuotes(ctx context.Context, requests []QuoteRequest) ([]Quote, error) {
	return fetchQuoteBatch(ctx, requests, p.fetchQuote)
}

func (p *FundNAVProvider) fetchQuote(ctx context.Context, request QuoteRequest) (Quote, error) {
	code := fundCode(request)
	if code == "" {
		return Quote{}, ErrUnsupported
	}
	query := url.Values{"fundCode": []string{code}, "pageIndex": []string{"1"}, "pageSize": []string{"1"}}
	var response struct {
		Data struct {
			Items []struct {
				Date     string `json:"FSRQ"`
				NAV      string `json:"DWJZ"`
				ChangePC string `json:"JZZZL"`
			} `json:"LSJZList"`
		} `json:"Data"`
	}
	if err := doJSONWithHeaders(ctx, p.client, p.Name(), p.baseURL+"?"+query.Encode(), map[string]string{"Referer": "http://fundf10.eastmoney.com/"}, &response); err != nil {
		return Quote{}, err
	}
	if len(response.Data.Items) == 0 {
		return Quote{}, &HTTPError{Provider: p.Name(), Message: "fund NAV is unavailable"}
	}
	item := response.Data.Items[0]
	price := parseStringFloat(item.NAV)
	if price <= 0 {
		return Quote{}, &HTTPError{Provider: p.Name(), Message: "fund NAV is unavailable"}
	}
	return Quote{AssetClass: AssetFund, Symbol: request.Symbol, Name: request.Name, Price: price, Currency: currencyOr(request.Currency, "CNY"), ChangePercent: parseStringFloat(item.ChangePC), AsOf: parseFundTime("", item.Date), FetchedAt: time.Now(), Provider: p.Name()}, nil
}

type SinaFundProvider struct {
	client  *http.Client
	baseURL string
}

func NewSinaFundProvider(client *http.Client) *SinaFundProvider {
	return NewSinaFundProviderWithBaseURL(client, defaultSinaFundBaseURL)
}

func NewSinaFundProviderWithBaseURL(client *http.Client, baseURL string) *SinaFundProvider {
	if client == nil {
		client = defaultHTTPClient()
	}
	return &SinaFundProvider{client: client, baseURL: baseURL}
}

func (p *SinaFundProvider) Name() string { return "sina-fund" }

func (p *SinaFundProvider) Supports(request QuoteRequest) bool {
	return request.AssetClass == AssetFund && isOTCFund(request)
}

func (p *SinaFundProvider) FetchQuotes(ctx context.Context, requests []QuoteRequest) ([]Quote, error) {
	return fetchQuoteBatch(ctx, requests, p.fetchQuote)
}

func (p *SinaFundProvider) fetchQuote(ctx context.Context, request QuoteRequest) (Quote, error) {
	code := fundCode(request)
	if code == "" {
		return Quote{}, ErrUnsupported
	}
	body, err := doRequest(ctx, p.client, p.Name(), sinaQuoteURL(p.baseURL, "f_"+code), map[string]string{"Referer": "https://finance.sina.com.cn/"})
	if err != nil {
		return Quote{}, err
	}
	line := strings.TrimSpace(string(body))
	start, end := strings.Index(line, "\""), strings.LastIndex(line, "\"")
	if start < 0 || end <= start {
		return Quote{}, &HTTPError{Provider: p.Name(), Message: "malformed fund NAV response"}
	}
	fields := strings.Split(line[start+1:end], ",")
	if len(fields) < 5 || strings.TrimSpace(fields[0]) == "" {
		return Quote{}, &HTTPError{Provider: p.Name(), Message: "fund NAV is unavailable"}
	}
	price := parseFloat(fields, 1)
	if price <= 0 {
		return Quote{}, &HTTPError{Provider: p.Name(), Message: "fund NAV is unavailable"}
	}
	return Quote{AssetClass: AssetFund, Symbol: request.Symbol, Name: strings.TrimSpace(fields[0]), Price: price, Currency: currencyOr(request.Currency, "CNY"), PreviousClose: parseFloat(fields, 3), AsOf: parseFundTime("", strings.TrimSpace(fields[4])), FetchedAt: time.Now(), Provider: p.Name()}, nil
}

func isOTCFund(request QuoteRequest) bool {
	if strings.EqualFold(strings.TrimSpace(request.Meta("classify")), "OTCFUND") {
		return true
	}
	return strings.HasPrefix(strings.TrimSpace(request.Meta("secid")), "150.") || strings.HasPrefix(strings.TrimSpace(request.Symbol), "150.")
}

func fundCode(request QuoteRequest) string {
	for _, candidate := range []string{request.Meta("fund_code"), request.Symbol, request.Meta("secid"), request.Meta("quoteId")} {
		value := strings.TrimSpace(candidate)
		if parts := strings.SplitN(value, ".", 2); len(parts) == 2 {
			value = parts[1]
		}
		value = strings.TrimPrefix(strings.ToLower(value), "f_")
		if len(value) != 6 {
			continue
		}
		valid := true
		for _, character := range value {
			if character < '0' || character > '9' {
				valid = false
				break
			}
		}
		if valid {
			return value
		}
	}
	return ""
}

func parseStringFloat(value string) float64 {
	parsed, _ := strconv.ParseFloat(strings.TrimSpace(value), 64)
	return parsed
}

func parseFundTime(primary, fallback string) time.Time {
	for _, value := range []string{primary, fallback} {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if parsed, err := time.ParseInLocation("2006-01-02 15:04", value, time.Local); err == nil {
			return parsed
		}
		if parsed, err := time.ParseInLocation("2006-01-02", value, time.Local); err == nil {
			return parsed
		}
	}
	return time.Now()
}
