package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go-stock/backend/data"
	"go-stock/backend/db"
	"go-stock/backend/models"
	assetquote "go-stock/backend/provider/assetquote"
	assetsearch "go-stock/backend/provider/assetsearch"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gorm.io/gorm/clause"
)

var portfolioQuoteRegistry = assetquote.NewDefaultRegistry(assetquote.DefaultRegistryConfig())
var portfolioAssetSearchService = assetsearch.NewService(assetsearch.Config{})

var majorFXCurrencies = []string{"USD", "HKD", "EUR", "GBP", "JPY", "AUD", "CAD", "SGD"}

var majorFXQuoteRegistry = func() *assetquote.Registry {
	config := assetquote.DefaultRegistryConfig()
	// Foreign-exchange rates are a dedicated live board, not a portfolio conversion cache.
	config.FXCacheTTL = time.Minute
	return assetquote.NewDefaultRegistry(config)
}()

type AssetQuoteRefreshResult struct {
	Requested int                         `json:"requested"`
	Updated   int                         `json:"updated"`
	Stale     int                         `json:"stale"`
	Failed    int                         `json:"failed"`
	Errors    []string                    `json:"errors"`
	Health    []assetquote.ProviderHealth `json:"health"`
	UpdatedAt time.Time                   `json:"updatedAt"`
}

func (a *App) GetPortfolioOverview() (*data.PortfolioOverview, error) {
	return data.GetPortfolioOverview()
}

// GetMajorFXQuotes returns the most recently persisted public FX rates, including a
// short history that the desktop client uses for a compact movement chart.
func (a *App) GetMajorFXQuotes() []data.MajorFXQuote {
	return loadMajorFXQuotes()
}

// RefreshMajorFXQuotes refreshes common currencies independently of the user's
// holdings so the FX board is useful even before a foreign-currency trade is recorded.
func (a *App) RefreshMajorFXQuotes() []data.MajorFXQuote {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	requests := make([]assetquote.QuoteRequest, 0, len(majorFXCurrencies))
	for _, currency := range majorFXCurrencies {
		requests = append(requests, assetquote.QuoteRequest{
			AssetClass: assetquote.AssetFX,
			Symbol:     currency + "/CNY",
			Currency:   "CNY",
			Provider:   "sina-fx",
			Metadata:   map[string]string{"base_currency": currency, "quote_currency": "CNY"},
		})
	}
	results, _ := majorFXQuoteRegistry.Quotes(ctx, requests)
	for index, quoteResult := range results {
		if quoteResult.Quote.Price <= 0 {
			continue
		}
		currency := majorFXCurrencies[index]
		quote := quoteResult.Quote
		stored := data.FXRate{
			BaseCurrency: currency, QuoteCurrency: "CNY", Rate: quote.Price,
			Provider: quote.Provider, AsOf: quote.AsOf, FetchedAt: quote.FetchedAt,
		}
		_ = db.Dao.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "base_currency"}, {Name: "quote_currency"}, {Name: "provider"}, {Name: "as_of"}},
			DoUpdates: clause.AssignmentColumns([]string{"rate", "fetched_at", "updated_at"}),
		}).Create(&stored).Error
	}
	return loadMajorFXQuotes()
}

func loadMajorFXQuotes() []data.MajorFXQuote {
	quotes := make([]data.MajorFXQuote, 0, len(majorFXCurrencies))
	for _, currency := range majorFXCurrencies {
		var rows []data.FXRate
		if err := db.Dao.Where("base_currency = ? AND quote_currency = ?", currency, "CNY").Order("as_of DESC").Limit(12).Find(&rows).Error; err != nil || len(rows) == 0 {
			continue
		}
		latest := rows[0]
		history := make([]data.FXQuotePoint, 0, len(rows))
		for index := len(rows) - 1; index >= 0; index-- {
			history = append(history, data.FXQuotePoint{At: rows[index].AsOf, Rate: rows[index].Rate})
		}
		changePercent := 0.0
		if len(rows) > 1 && rows[1].Rate > 0 {
			changePercent = (latest.Rate - rows[1].Rate) / rows[1].Rate * 100
		}
		quotes = append(quotes, data.MajorFXQuote{
			BaseCurrency: currency, QuoteCurrency: "CNY", Rate: latest.Rate, ChangePercent: changePercent,
			Provider: latest.Provider, AsOf: latest.AsOf, FetchedAt: latest.FetchedAt, History: history,
		})
	}
	return quotes
}

func (a *App) GetAssetPositions(assetType string) ([]data.AssetPosition, error) {
	return data.GetAssetPositions(normalizeUIAssetType(assetType))
}

func (a *App) SearchAssetInstruments(assetType, keyword string, limit int) ([]assetsearch.Instrument, error) {
	ctx := context.Background()
	if a.ctx != nil {
		ctx = a.ctx
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	return portfolioAssetSearchService.Search(ctx, normalizeUIAssetType(assetType), keyword, limit)
}

func (a *App) AddAssetTransaction(input data.AssetTransactionInput) (uint, error) {
	input.AssetType = normalizeUIAssetType(input.AssetType)
	return data.AddAssetTransaction(input)
}

func (a *App) UpdateAssetTransaction(input data.AssetTransactionInput) error {
	input.AssetType = normalizeUIAssetType(input.AssetType)
	return data.UpdateAssetTransaction(input)
}

func (a *App) DeleteAssetTransaction(id uint) error { return data.DeleteAssetTransaction(id) }

func (a *App) GetPortfolioSnapshots(page, pageSize int) (*data.PortfolioSnapshotPage, error) {
	return data.GetPortfolioSnapshots(page, pageSize)
}

func (a *App) CreatePortfolioSnapshot() (*data.PortfolioSnapshot, error) {
	return data.CreatePortfolioSnapshot()
}

func (a *App) PreviewBankProductCSV(path string) (*data.BankProductImportPreview, error) {
	return data.PreviewBankProductCSV(path)
}

func (a *App) SelectBankProductCSV() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   "选择银行理财 CSV",
		Filters: []runtime.FileFilter{{DisplayName: "CSV 文件 (*.csv)", Pattern: "*.csv"}},
	})
}

func (a *App) ImportBankProductCSV(batchID, sourcePath string, rows []data.BankProductImportRow) error {
	return data.ImportBankProductCSV(batchID, sourcePath, rows)
}

func (a *App) RollbackAssetImport(batchID string) error { return data.RollbackAssetImport(batchID) }

func (a *App) GetTelegraphBriefListBySources(sources []string, limit int) *[]*models.Telegraph {
	return data.NewMarketNewsApi().GetTelegraphBriefListBySources(sources, limit)
}

func (a *App) GetStockAvailableVolume(stockCode string) int64 {
	return data.NewStockDataApi().GetStockAvailableVolume(stockCode)
}

func (a *App) GetAssetQuoteProviderHealth() []assetquote.ProviderHealth {
	return portfolioQuoteRegistry.Health()
}

func (a *App) SaveManualAssetQuote(instrumentID uint, price float64, currency string) error {
	if instrumentID == 0 || price <= 0 {
		return fmt.Errorf("产品和报价必须有效")
	}
	now := time.Now()
	quote := data.AssetQuote{InstrumentID: instrumentID, Provider: "manual", Price: price, Currency: strings.ToUpper(strings.TrimSpace(currency)), AsOf: now, FetchedAt: now}
	if quote.Currency == "" {
		quote.Currency = "CNY"
	}
	if err := db.Dao.Create(&quote).Error; err != nil {
		return err
	}
	_, _ = data.CreatePortfolioSnapshot()
	return nil
}

func (a *App) RefreshAssetQuotes(assetTypes []string) AssetQuoteRefreshResult {
	result := AssetQuoteRefreshResult{Errors: []string{}, UpdatedAt: time.Now()}
	requestedTypes := map[string]bool{}
	for _, assetType := range assetTypes {
		requestedTypes[normalizeUIAssetType(assetType)] = true
	}
	var instruments []data.AssetInstrument
	query := db.Dao.Where("is_active = ?", true)
	if len(requestedTypes) > 0 {
		keys := make([]string, 0, len(requestedTypes))
		for key := range requestedTypes {
			if key != "" {
				keys = append(keys, key)
			}
		}
		if len(keys) > 0 {
			query = query.Where("asset_type IN ?", keys)
		}
	}
	if err := query.Find(&instruments).Error; err != nil {
		result.Errors = append(result.Errors, err.Error())
		result.Failed++
		result.Health = portfolioQuoteRegistry.Health()
		return result
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	seenCurrencies := map[string]bool{"CNY": true}
	quoteInstruments := make([]data.AssetInstrument, 0, len(instruments))
	quoteRequests := make([]assetquote.QuoteRequest, 0, len(instruments))
	for _, instrument := range instruments {
		currency := strings.ToUpper(strings.TrimSpace(instrument.Currency))
		if currency != "" {
			seenCurrencies[currency] = true
		}
		if instrument.AssetType == data.AssetTypeBankWealth {
			continue
		}
		quoteInstruments = append(quoteInstruments, instrument)
		quoteRequests = append(quoteRequests, instrumentQuoteRequest(instrument))
		result.Requested++
	}
	quoteResults, quoteErr := portfolioQuoteRegistry.Quotes(ctx, quoteRequests)
	if quoteErr != nil {
		result.Errors = append(result.Errors, quoteErr.Error())
	}
	for index, quoteResult := range quoteResults {
		instrument := quoteInstruments[index]
		if quoteResult.Quote.Price <= 0 {
			result.Failed++
			continue
		}
		quote := quoteResult.Quote
		stored := data.AssetQuote{InstrumentID: instrument.ID, Provider: quote.Provider, Price: quote.Price, Currency: quote.Currency, AsOf: quote.AsOf, FetchedAt: quote.FetchedAt, IsStale: quoteResult.Stale || quote.Stale}
		metadata, _ := json.Marshal(quote)
		stored.MetadataJSON = string(metadata)
		if err := db.Dao.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "instrument_id"}, {Name: "provider"}, {Name: "as_of"}}, DoUpdates: clause.AssignmentColumns([]string{"price", "currency", "fetched_at", "is_stale", "metadata_json", "updated_at"})}).Create(&stored).Error; err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", instrument.Symbol, err))
			continue
		}
		result.Updated++
		if stored.IsStale {
			result.Stale++
		}
	}
	fxCurrencies := make([]string, 0, len(seenCurrencies))
	fxRequests := make([]assetquote.QuoteRequest, 0, len(seenCurrencies))
	for currency := range seenCurrencies {
		if currency == "CNY" {
			continue
		}
		fxCurrencies = append(fxCurrencies, currency)
		fxRequests = append(fxRequests, assetquote.QuoteRequest{AssetClass: assetquote.AssetFX, Symbol: currency + "/CNY", Currency: "CNY", Provider: "sina-fx", Metadata: map[string]string{"base_currency": currency, "quote_currency": "CNY"}})
		result.Requested++
	}
	fxResults, fxErr := portfolioQuoteRegistry.Quotes(ctx, fxRequests)
	if fxErr != nil {
		result.Errors = append(result.Errors, fxErr.Error())
	}
	for index, fxResult := range fxResults {
		currency := fxCurrencies[index]
		if fxResult.Quote.Price <= 0 {
			result.Failed++
			continue
		}
		fx := data.FXRate{BaseCurrency: currency, QuoteCurrency: "CNY", Rate: fxResult.Quote.Price, Provider: fxResult.Quote.Provider, AsOf: fxResult.Quote.AsOf, FetchedAt: fxResult.Quote.FetchedAt}
		if err := db.Dao.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "base_currency"}, {Name: "quote_currency"}, {Name: "provider"}, {Name: "as_of"}}, DoUpdates: clause.AssignmentColumns([]string{"rate", "fetched_at", "updated_at"})}).Create(&fx).Error; err != nil {
			result.Failed++
			result.Errors = append(result.Errors, err.Error())
			continue
		}
		result.Updated++
	}
	if _, err := data.CreatePortfolioSnapshot(); err != nil {
		result.Errors = append(result.Errors, err.Error())
	}
	result.Health = portfolioQuoteRegistry.Health()
	return result
}

func instrumentQuoteRequest(instrument data.AssetInstrument) assetquote.QuoteRequest {
	class := assetquote.AssetClass(instrument.AssetType)
	switch instrument.AssetType {
	case data.AssetTypeForex:
		class = assetquote.AssetFX
	case data.AssetTypeFutures:
		class = assetquote.AssetFuture
	}
	metadataAny := map[string]any{}
	_ = json.Unmarshal([]byte(instrument.MetadataJSON), &metadataAny)
	metadata := map[string]string{}
	for key, value := range metadataAny {
		metadata[key] = fmt.Sprint(value)
	}
	symbol := strings.TrimSpace(instrument.QuoteSymbol)
	if symbol == "" {
		symbol = instrument.Symbol
	}
	return assetquote.QuoteRequest{AssetClass: class, Symbol: symbol, Name: instrument.Name, Exchange: instrument.Exchange, Currency: instrument.Currency, Provider: strings.ToLower(strings.TrimSpace(metadata["quote_provider"])), Metadata: metadata}
}

func normalizeUIAssetType(assetType string) string {
	switch strings.ToLower(strings.TrimSpace(assetType)) {
	case "future":
		return data.AssetTypeFutures
	case "bank_product":
		return data.AssetTypeBankWealth
	default:
		return strings.ToLower(strings.TrimSpace(assetType))
	}
}
