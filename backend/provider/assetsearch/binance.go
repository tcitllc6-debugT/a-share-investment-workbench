package assetsearch

import (
	"context"
	"fmt"
	"strings"
	"time"
)

const defaultBinanceExchangeInfoURL = "https://data-api.binance.vision/api/v3/exchangeInfo"

type binanceSymbol struct {
	Symbol               string `json:"symbol"`
	Status               string `json:"status"`
	BaseAsset            string `json:"baseAsset"`
	QuoteAsset           string `json:"quoteAsset"`
	IsSpotTradingAllowed bool   `json:"isSpotTradingAllowed"`
}

func (s *Service) searchBinance(ctx context.Context, keyword string, limit int) ([]Instrument, error) {
	symbols, err := s.binanceDirectory(ctx)
	if err != nil {
		return nil, err
	}
	keyword = strings.ToUpper(strings.TrimSpace(keyword))
	results := make([]Instrument, 0, limit)
	for _, record := range symbols {
		if strings.ToUpper(record.Status) != "TRADING" || !record.IsSpotTradingAllowed {
			continue
		}
		symbol := strings.ToUpper(strings.TrimSpace(record.Symbol))
		base := strings.ToUpper(strings.TrimSpace(record.BaseAsset))
		quote := strings.ToUpper(strings.TrimSpace(record.QuoteAsset))
		valuationCurrency := cryptoValuationCurrency(quote)
		if valuationCurrency == "" {
			continue
		}
		if !strings.Contains(symbol, keyword) && !strings.Contains(base, keyword) && !strings.Contains(quote, keyword) {
			continue
		}
		results = append(results, Instrument{
			Symbol: base + "/" + quote, Name: base + "/" + quote,
			Exchange: "币安现货", QuoteSymbol: symbol, Currency: valuationCurrency, Unit: base,
			Source: "binance", Category: assetCategoryName(AssetCrypto),
			Metadata: map[string]string{
				"baseAsset": base, "quoteAsset": quote, "status": record.Status,
				"binance_symbol": symbol, "quote_currency": quote,
			},
		})
		if len(results) == limit {
			break
		}
	}
	return results, nil
}

func (s *Service) binanceDirectory(ctx context.Context) ([]binanceSymbol, error) {
	s.binanceMu.RLock()
	if len(s.binanceItems) > 0 && time.Now().Before(s.binanceUntil) {
		items := s.binanceItems
		s.binanceMu.RUnlock()
		return items, nil
	}
	s.binanceMu.RUnlock()

	var payload struct {
		Symbols []binanceSymbol `json:"symbols"`
	}
	if err := getJSON(ctx, s.client, s.binanceURL, &payload); err != nil {
		return nil, fmt.Errorf("币安产品搜索失败：%w", err)
	}
	s.binanceMu.Lock()
	s.binanceItems = payload.Symbols
	s.binanceUntil = time.Now().Add(10 * time.Minute)
	s.binanceMu.Unlock()
	return payload.Symbols, nil
}

func cryptoValuationCurrency(quote string) string {
	switch strings.ToUpper(strings.TrimSpace(quote)) {
	case "USDT", "USDC", "FDUSD", "TUSD", "DAI", "USD":
		return "USD"
	case "EUR", "TRY", "GBP", "AUD", "BRL", "JPY":
		return strings.ToUpper(strings.TrimSpace(quote))
	default:
		return ""
	}
}
