package assetsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	defaultEastmoneySearchURL = "https://searchapi.eastmoney.com/api/suggest/get"
)

type eastmoneyRecord struct {
	Code             string `json:"Code"`
	Name             string `json:"Name"`
	PinYin           string `json:"PinYin"`
	ExchangeCode     string `json:"JYS"`
	Classify         string `json:"Classify"`
	MarketType       string `json:"MarketType"`
	SecurityType     string `json:"SecurityType"`
	SecurityTypeName string `json:"SecurityTypeName"`
	QuoteID          string `json:"QuoteID"`
	UnifiedCode      string `json:"UnifiedCode"`
}

func (s *Service) searchEastmoney(ctx context.Context, assetType, keyword string, limit int) ([]Instrument, error) {
	query := url.Values{}
	query.Set("input", keyword)
	query.Set("type", "14")
	query.Set("count", fmt.Sprint(maxSearchResults))
	requestURL := s.eastmoneyURL + "?" + query.Encode()
	var payload struct {
		QuotationCodeTable struct {
			Data []eastmoneyRecord `json:"Data"`
		} `json:"QuotationCodeTable"`
	}
	if err := getJSON(ctx, s.client, requestURL, &payload); err != nil {
		return nil, fmt.Errorf("东方财富产品搜索失败：%w", err)
	}
	results := make([]Instrument, 0, limit)
	for _, record := range payload.QuotationCodeTable.Data {
		if !eastmoneyMatchesAssetType(record, assetType) {
			continue
		}
		results = append(results, normalizeEastmoneyRecord(record, assetType))
		if len(results) == limit {
			break
		}
	}
	return results, nil
}

func eastmoneyMatchesAssetType(record eastmoneyRecord, assetType string) bool {
	classification := strings.ToLower(strings.Join([]string{record.Classify, record.SecurityType, record.SecurityTypeName, record.Name}, " "))
	code := strings.ToUpper(strings.TrimSpace(record.Code))
	isFuture := containsAny(classification, "future", "futures", "期货", "主连", "次主连")
	isPreciousName := containsAny(classification, "贵金属", "黄金", "白银", "铂", "钯", "伦敦金", "伦敦银") || hasAnyPrefix(code, "XAU", "XAG", "HLAU", "HLAG", "AGTD", "AUTD")
	isPreciousSpot := isPreciousName && !isFuture && (containsAny(classification, "现货", "sge", "forpm", "hkpm") || hasAnyPrefix(code, "XAU", "XAG", "HLAU", "HLAG", "AGTD", "AUTD"))
	switch assetType {
	case AssetStock:
		return (strings.EqualFold(strings.TrimSpace(record.Classify), "AStock") || containsAny(classification, "沪a", "深a", "北a", "a股")) &&
			!containsAny(classification, "fund", "基金", "bond", "债")
	case AssetFund:
		return containsAny(classification, "fund", "基金", "etf", "lof")
	case AssetBond:
		return containsAny(classification, "bond", "债")
	case AssetPreciousMetal:
		return isPreciousSpot
	case AssetForex:
		return containsAny(classification, "forex", "外汇", "汇率") && normalizeForexPair(code) != ""
	case AssetFutures:
		return isFuture
	default:
		return false
	}
}

func normalizeEastmoneyRecord(record eastmoneyRecord, assetType string) Instrument {
	code := strings.TrimSpace(record.Code)
	if code == "" {
		code = strings.TrimSpace(record.UnifiedCode)
	}
	exchange, suffix := eastmoneyExchange(record)
	symbol := code
	if suffix != "" && !strings.Contains(symbol, ".") {
		symbol += "." + suffix
	}
	quoteSymbol := strings.TrimSpace(record.QuoteID)
	if quoteSymbol == "" {
		quoteSymbol = symbol
	}
	instrument := Instrument{
		Symbol: symbol, Name: strings.TrimSpace(record.Name), Exchange: exchange,
		QuoteSymbol: quoteSymbol, Currency: "CNY", Source: "eastmoney",
		Category: assetCategoryName(assetType),
		Metadata: map[string]string{
			"classify": record.Classify, "securityType": record.SecurityType,
			"securityTypeName": record.SecurityTypeName, "marketType": record.MarketType,
			"exchangeCode": record.ExchangeCode, "pinyin": record.PinYin,
			"quoteId": quoteSymbol, "secid": quoteSymbol,
		},
	}
	if assetType == AssetPreciousMetal {
		instrument.Unit = "克"
	}
	if assetType == AssetForex {
		pair := normalizeForexPair(code)
		if pair != "" {
			base, quoteCurrency := pair[:3], pair[4:]
			instrument.Symbol = pair
			instrument.QuoteSymbol = pair
			instrument.Currency = quoteCurrency
			instrument.Metadata["base_currency"] = base
			instrument.Metadata["quote_currency"] = quoteCurrency
			instrument.Metadata["upstream_quote_id"] = quoteSymbol
		}
	}
	return instrument
}

var searchableCurrencies = map[string]bool{
	"AUD": true, "BRL": true, "CAD": true, "CHF": true, "CNY": true,
	"CZK": true, "DKK": true, "EUR": true, "GBP": true, "HKD": true,
	"HUF": true, "IDR": true, "ILS": true, "INR": true, "ISK": true,
	"JPY": true, "KRW": true, "MXN": true, "MYR": true, "NOK": true,
	"NZD": true, "PHP": true, "PLN": true, "RON": true, "SEK": true,
	"SGD": true, "THB": true, "TRY": true, "USD": true, "ZAR": true,
}

func normalizeForexPair(code string) string {
	value := strings.ToUpper(strings.TrimSpace(code))
	for _, separator := range []string{"/", "-", "_"} {
		value = strings.ReplaceAll(value, separator, "")
	}
	// 东方财富的人民币中间价代码通常带 C 后缀，例如 USDCNYC。
	if len(value) == 7 && strings.HasSuffix(value, "C") {
		value = strings.TrimSuffix(value, "C")
	}
	if len(value) != 6 {
		return ""
	}
	base, quote := value[:3], value[3:]
	if base == quote || !searchableCurrencies[base] || !searchableCurrencies[quote] {
		return ""
	}
	return base + "/" + quote
}

func eastmoneyExchange(record eastmoneyRecord) (string, string) {
	typeName := strings.ToLower(record.SecurityTypeName)
	quoteMarket := ""
	if parts := strings.SplitN(strings.TrimSpace(record.QuoteID), ".", 2); len(parts) == 2 {
		quoteMarket = parts[0]
	}
	if containsAny(typeName, "沪a", "上海") || quoteMarket == "1" || record.ExchangeCode == "2" {
		return "上海证券交易所", "SH"
	}
	if containsAny(typeName, "北a", "北京") || (quoteMarket == "0" && hasAnyPrefix(strings.TrimSpace(record.Code), "4", "8")) {
		return "北京证券交易所", "BJ"
	}
	if containsAny(typeName, "深a", "深圳") || quoteMarket == "0" || record.ExchangeCode == "6" {
		return "深圳证券交易所", "SZ"
	}
	if containsAny(typeName, "外汇", "汇率") {
		return "国际外汇市场", ""
	}
	if quoteMarket == "118" || strings.EqualFold(record.Classify, "SGE") {
		return "上海黄金交易所", ""
	}
	if quoteMarket == "123" || strings.EqualFold(record.Classify, "HKPM") {
		return "香港贵金属市场", ""
	}
	if containsAny(typeName, "期货", "主连", "贵金属", "黄金", "白银") {
		return "国内期货市场", ""
	}
	return "东方财富行情", ""
}

func assetCategoryName(assetType string) string {
	switch assetType {
	case AssetStock:
		return "股票"
	case AssetFund:
		return "基金"
	case AssetBond:
		return "债券"
	case AssetPreciousMetal:
		return "贵金属"
	case AssetForex:
		return "外汇"
	case AssetFutures:
		return "期货"
	case AssetCrypto:
		return "加密货币"
	default:
		return "金融产品"
	}
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, strings.ToLower(needle)) {
			return true
		}
	}
	return false
}

func hasAnyPrefix(value string, prefixes ...string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}

func getJSON(ctx context.Context, client *http.Client, requestURL string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "A股智投工作台/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("上游返回 HTTP %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("响应数据无法解析：%w", err)
	}
	return nil
}
