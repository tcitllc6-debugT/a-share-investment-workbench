package assetsearch

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

const maxSearchResults = 20

type Service struct {
	client       *http.Client
	eastmoneyURL string
	binanceURL   string
	binanceMu    sync.RWMutex
	binanceItems []binanceSymbol
	binanceUntil time.Time
}

func NewService(config Config) *Service {
	client := config.Client
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	eastmoneyURL := strings.TrimSpace(config.EastmoneyURL)
	if eastmoneyURL == "" {
		eastmoneyURL = defaultEastmoneySearchURL
	}
	binanceURL := strings.TrimSpace(config.BinanceURL)
	if binanceURL == "" {
		binanceURL = defaultBinanceExchangeInfoURL
	}
	return &Service{client: client, eastmoneyURL: eastmoneyURL, binanceURL: binanceURL}
}

func (s *Service) Search(ctx context.Context, assetType, keyword string, limit int) ([]Instrument, error) {
	assetType = normalizeAssetType(assetType)
	keyword = strings.TrimSpace(keyword)
	if keyword == "" || assetType == AssetBankWealth {
		return []Instrument{}, nil
	}
	limit = normalizeLimit(limit)
	builtins := searchBuiltins(assetType, keyword, limit)
	var remote []Instrument
	var err error
	switch assetType {
	case AssetCrypto:
		remote, err = s.searchBinance(ctx, keyword, limit)
	case AssetStock, AssetFund, AssetBond, AssetPreciousMetal, AssetForex, AssetFutures:
		remote, err = s.searchEastmoney(ctx, assetType, keyword, limit)
	default:
		return nil, fmt.Errorf("不支持的资产类型：%s", assetType)
	}
	if err != nil {
		if len(builtins) > 0 {
			return builtins, nil
		}
		return nil, err
	}
	return mergeInstruments(builtins, remote, limit), nil
}

func normalizeAssetType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "future":
		return AssetFutures
	case "bank_product":
		return AssetBankWealth
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}

func normalizeLimit(limit int) int {
	if limit <= 0 || limit > maxSearchResults {
		return maxSearchResults
	}
	return limit
}
