package assetsearch

import "strings"

type builtinInstrument struct {
	assetType string
	keywords  []string
	item      Instrument
}

var builtinInstruments = []builtinInstrument{
	{AssetPreciousMetal, []string{"黄金", "au", "au9999", "金"}, Instrument{Symbol: "AU9999", Name: "黄金9999", Exchange: "上海黄金交易所", QuoteSymbol: "118.AU9999", Currency: "CNY", Unit: "克", Source: "builtin", Category: "贵金属", Metadata: map[string]string{"secid": "118.AU9999", "quote_provider": "eastmoney", "sina_reference_symbol": "nf_AU0", "purity": "Au99.99", "unitToQuote": "1"}}},
	{AssetPreciousMetal, []string{"白银", "ag", "agtd", "银"}, Instrument{Symbol: "AGTD", Name: "白银T+D", Exchange: "上海黄金交易所", QuoteSymbol: "118.AGTD", Currency: "CNY", Unit: "千克", Source: "builtin", Category: "贵金属", Metadata: map[string]string{"secid": "118.AGTD", "quote_provider": "eastmoney", "sina_reference_symbol": "nf_AG0", "purity": "Ag99.9", "unitToQuote": "1"}}},
	{AssetPreciousMetal, []string{"伦敦金", "现货黄金", "xau"}, Instrument{Symbol: "XAU", Name: "伦敦金（现货黄金）", Exchange: "国际现货", QuoteSymbol: "XAU", Currency: "USD", Unit: "盎司", Source: "builtin", Category: "贵金属", Metadata: map[string]string{"sina_symbol": "hf_XAU", "quote_provider": "sina", "price_index": "0", "previous_close_index": "7", "unitToQuote": "1"}}},
	{AssetPreciousMetal, []string{"伦敦银", "现货白银", "xag"}, Instrument{Symbol: "XAG", Name: "伦敦银（现货白银）", Exchange: "国际现货", QuoteSymbol: "XAG", Currency: "USD", Unit: "盎司", Source: "builtin", Category: "贵金属", Metadata: map[string]string{"sina_symbol": "hf_XAG", "quote_provider": "sina", "price_index": "0", "previous_close_index": "7", "unitToQuote": "1"}}},
	{AssetForex, []string{"美元", "usd", "美元人民币", "usdcny"}, Instrument{Symbol: "USD/CNY", Name: "美元/人民币", Exchange: "外汇市场", QuoteSymbol: "USD/CNY", Currency: "CNY", Source: "builtin", Category: "外汇", Metadata: map[string]string{"base_currency": "USD", "quote_currency": "CNY", "quote_provider": "frankfurter"}}},
	{AssetForex, []string{"欧元", "eur", "欧元人民币", "eurcny"}, Instrument{Symbol: "EUR/CNY", Name: "欧元/人民币", Exchange: "外汇市场", QuoteSymbol: "EUR/CNY", Currency: "CNY", Source: "builtin", Category: "外汇", Metadata: map[string]string{"base_currency": "EUR", "quote_currency": "CNY", "quote_provider": "frankfurter"}}},
	{AssetForex, []string{"港币", "hkd", "港币人民币", "hkdcny"}, Instrument{Symbol: "HKD/CNY", Name: "港币/人民币", Exchange: "外汇市场", QuoteSymbol: "HKD/CNY", Currency: "CNY", Source: "builtin", Category: "外汇", Metadata: map[string]string{"base_currency": "HKD", "quote_currency": "CNY", "quote_provider": "frankfurter"}}},
	{AssetForex, []string{"日元", "jpy", "日元人民币", "jpycny"}, Instrument{Symbol: "JPY/CNY", Name: "日元/人民币", Exchange: "外汇市场", QuoteSymbol: "JPY/CNY", Currency: "CNY", Source: "builtin", Category: "外汇", Metadata: map[string]string{"base_currency": "JPY", "quote_currency": "CNY", "quote_provider": "frankfurter"}}},
	{AssetFutures, []string{"黄金", "沪金", "au0", "黄金主连"}, Instrument{Symbol: "AU0", Name: "黄金主连", Exchange: "上海期货交易所", QuoteSymbol: "AU0", Currency: "CNY", Unit: "手", Source: "builtin", Category: "期货", Metadata: map[string]string{"sina_symbol": "nf_AU0", "quote_provider": "sina", "price_index": "8", "previous_close_index": "10"}}},
	{AssetFutures, []string{"白银", "沪银", "ag0", "白银主连"}, Instrument{Symbol: "AG0", Name: "白银主连", Exchange: "上海期货交易所", QuoteSymbol: "AG0", Currency: "CNY", Unit: "手", Source: "builtin", Category: "期货", Metadata: map[string]string{"sina_symbol": "nf_AG0", "quote_provider": "sina", "price_index": "8", "previous_close_index": "10"}}},
	{AssetFutures, []string{"玻璃", "fg0", "玻璃主连"}, Instrument{Symbol: "FG0", Name: "玻璃主连", Exchange: "郑州商品交易所", QuoteSymbol: "FG0", Currency: "CNY", Unit: "手", Source: "builtin", Category: "期货", Metadata: map[string]string{"sina_symbol": "nf_FG0", "quote_provider": "sina", "price_index": "8", "previous_close_index": "10"}}},
}

func searchBuiltins(assetType, keyword string, limit int) []Instrument {
	query := normalizeBuiltinKeyword(keyword)
	results := make([]Instrument, 0, limit)
	for _, preset := range builtinInstruments {
		if preset.assetType != assetType || !matchesBuiltin(preset, query) {
			continue
		}
		item := preset.item
		item.Metadata = cloneMetadata(item.Metadata)
		results = append(results, item)
		if len(results) == limit {
			break
		}
	}
	return results
}

func mergeInstruments(primary, secondary []Instrument, limit int) []Instrument {
	results := make([]Instrument, 0, limit)
	seen := map[string]bool{}
	for _, item := range append(primary, secondary...) {
		key := strings.ToUpper(strings.TrimSpace(item.Symbol)) + "|" + strings.ToUpper(strings.TrimSpace(item.QuoteSymbol))
		if key == "|" || seen[key] {
			continue
		}
		seen[key] = true
		results = append(results, item)
		if len(results) == limit {
			break
		}
	}
	return results
}

func matchesBuiltin(item builtinInstrument, query string) bool {
	if query == "" {
		return false
	}
	for _, keyword := range item.keywords {
		candidate := normalizeBuiltinKeyword(keyword)
		if strings.Contains(candidate, query) || strings.Contains(query, candidate) {
			return true
		}
	}
	return false
}

func normalizeBuiltinKeyword(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	return strings.NewReplacer("/", "", "-", "", "_", "", " ", "").Replace(value)
}

func cloneMetadata(source map[string]string) map[string]string {
	result := make(map[string]string, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}
