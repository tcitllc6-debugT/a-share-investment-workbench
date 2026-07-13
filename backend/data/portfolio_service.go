package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-stock/backend/db"
	"math"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

type PortfolioService struct {
	DB *gorm.DB
}

func NewPortfolioService() *PortfolioService { return &PortfolioService{DB: db.Dao} }

func NewPortfolioServiceWithDB(database *gorm.DB) *PortfolioService {
	return &PortfolioService{DB: database}
}

func GetPortfolioOverview() (*PortfolioOverview, error) {
	return NewPortfolioService().GetPortfolioOverview()
}

func GetAssetPositions(assetType string) ([]AssetPosition, error) {
	return NewPortfolioService().GetAssetPositions(assetType)
}

func AddAssetTransaction(input AssetTransactionInput) (uint, error) {
	return NewPortfolioService().AddAssetTransaction(input)
}

func UpdateAssetTransaction(input AssetTransactionInput) error {
	return NewPortfolioService().UpdateAssetTransaction(input)
}

func DeleteAssetTransaction(id uint) error {
	return NewPortfolioService().DeleteAssetTransaction(id)
}

func (service *PortfolioService) GetPortfolioOverview() (*PortfolioOverview, error) {
	positions, err := service.GetAssetPositions("")
	if err != nil {
		return nil, err
	}
	overview := &PortfolioOverview{
		AsOf:               time.Now(),
		ByAssetType:        make(map[string]AssetClassSummary, len(SupportedAssetTypes)),
		CurrencyExposure:   make([]CurrencyExposure, 0),
		Positions:          positions,
		RecentTransactions: []PortfolioRecentTransaction{},
	}
	currencies := make(map[string]float64)
	for _, assetType := range SupportedAssetTypes {
		overview.ByAssetType[assetType] = AssetClassSummary{AssetType: assetType}
	}
	for _, position := range positions {
		overview.TotalCostCNY += position.CostCNY
		overview.MarketValueCNY += position.MarketValueCNY
		overview.FloatingPnLCNY += position.FloatingPnLCNY
		overview.RealizedPnLCNY += position.RealizedPnLCNY
		overview.NetCashFlowCNY += position.NetCashFlowCNY
		if position.QuoteStale {
			overview.StaleQuoteCount++
			overview.QuoteSummary.Stale++
		} else if position.QuoteProvider == "manual" {
			overview.QuoteSummary.Manual++
		} else if position.QuoteProvider != "" {
			overview.QuoteSummary.Fresh++
		} else {
			overview.QuoteSummary.Error++
		}
		summary := overview.ByAssetType[position.AssetType]
		summary.AssetType = position.AssetType
		summary.PositionCount++
		summary.CostCNY += position.CostCNY
		summary.MarketValueCNY += position.MarketValueCNY
		summary.FloatingPnLCNY += position.FloatingPnLCNY
		summary.RealizedPnLCNY += position.RealizedPnLCNY
		summary.NetCashFlowCNY += position.NetCashFlowCNY
		overview.ByAssetType[position.AssetType] = summary
		currencies[position.Currency] += position.MarketValueCNY
	}
	overview.PositionCount = len(positions)
	if overview.TotalCostCNY > quantityEpsilon {
		overview.FloatingPnLPct = overview.FloatingPnLCNY / overview.TotalCostCNY * 100
	}
	for assetType, summary := range overview.ByAssetType {
		if overview.MarketValueCNY > quantityEpsilon {
			summary.AllocationPct = summary.MarketValueCNY / overview.MarketValueCNY * 100
		}
		overview.ByAssetType[assetType] = summary
	}
	for currency, value := range currencies {
		exposure := CurrencyExposure{Currency: currency, MarketValueCNY: value}
		if overview.MarketValueCNY > quantityEpsilon {
			exposure.AllocationPct = value / overview.MarketValueCNY * 100
		}
		overview.CurrencyExposure = append(overview.CurrencyExposure, exposure)
	}
	sort.Slice(overview.CurrencyExposure, func(i, j int) bool {
		return overview.CurrencyExposure[i].MarketValueCNY > overview.CurrencyExposure[j].MarketValueCNY
	})
	var recent []AssetTransaction
	if err := service.DB.Order("trade_time DESC, id DESC").Limit(12).Find(&recent).Error; err == nil {
		instrumentIDs := make([]uint, 0, len(recent))
		seenInstrumentIDs := make(map[uint]struct{}, len(recent))
		for _, transaction := range recent {
			if _, exists := seenInstrumentIDs[transaction.InstrumentID]; !exists {
				seenInstrumentIDs[transaction.InstrumentID] = struct{}{}
				instrumentIDs = append(instrumentIDs, transaction.InstrumentID)
			}
		}
		var recentInstruments []AssetInstrument
		if len(instrumentIDs) > 0 {
			_ = service.DB.Where("id IN ?", instrumentIDs).Find(&recentInstruments).Error
		}
		instrumentMap := make(map[uint]AssetInstrument, len(recentInstruments))
		for _, instrument := range recentInstruments {
			instrumentMap[instrument.ID] = instrument
		}
		for _, transaction := range recent {
			instrument := instrumentMap[transaction.InstrumentID]
			fx := transaction.FXRateToCNY
			if normalizeCurrency(transaction.Currency) == "CNY" {
				fx = 1
			}
			if fx <= 0 {
				fx = 1
			}
			amountCNY := transactionLocalValue(transaction.AssetType, math.Abs(transaction.Quantity), transaction.Price, decodeMetadata(instrument.MetadataJSON)) * fx
			overview.RecentTransactions = append(overview.RecentTransactions, PortfolioRecentTransaction{ID: transaction.ID, AssetType: transaction.AssetType, InstrumentName: instrument.Name, InstrumentCode: instrument.Symbol, Action: transaction.Direction, Quantity: transaction.Quantity, AmountCNY: amountCNY, TradedAt: transaction.TradeTime})
		}
	}
	return overview, nil
}

func (service *PortfolioService) GetAssetPositions(assetType string) ([]AssetPosition, error) {
	if service == nil || service.DB == nil {
		return nil, fmt.Errorf("portfolio database is unavailable")
	}
	assetType = strings.TrimSpace(assetType)
	if assetType != "" && !isSupportedAssetType(assetType) {
		return nil, fmt.Errorf("unsupported asset type %q", assetType)
	}
	query := service.DB.Where("is_active = ?", true)
	if assetType != "" {
		query = query.Where("asset_type = ?", assetType)
	}
	var instruments []AssetInstrument
	if err := query.Order("asset_type ASC, symbol ASC").Find(&instruments).Error; err != nil {
		return nil, fmt.Errorf("list asset instruments: %w", err)
	}
	if len(instruments) == 0 {
		return []AssetPosition{}, nil
	}
	ids := make([]uint, 0, len(instruments))
	for _, instrument := range instruments {
		ids = append(ids, instrument.ID)
	}
	var transactions []AssetTransaction
	if err := service.DB.Where("instrument_id IN ?", ids).Order("trade_time ASC, id ASC").Find(&transactions).Error; err != nil {
		return nil, fmt.Errorf("list asset transactions: %w", err)
	}
	transactionsByInstrument := make(map[uint][]AssetTransaction)
	for _, transaction := range transactions {
		transactionsByInstrument[transaction.InstrumentID] = append(transactionsByInstrument[transaction.InstrumentID], transaction)
	}
	quotes, err := service.latestQuotes(ids)
	if err != nil {
		return nil, err
	}
	fxRates, err := service.latestCNYRates()
	if err != nil {
		return nil, err
	}
	positions := make([]AssetPosition, 0, len(instruments))
	for _, instrument := range instruments {
		instrumentTransactions := transactionsByInstrument[instrument.ID]
		if len(instrumentTransactions) == 0 {
			continue
		}
		currency := normalizeCurrency(instrument.Currency)
		fx := fxRates[currency]
		if currency == "CNY" {
			fx = 1
		}
		if fx <= 0 {
			fx = latestTransactionFX(instrumentTransactions)
		}
		position := calculateAssetPosition(instrument, instrumentTransactions, quotes[instrument.ID], fx)
		if math.Abs(position.Quantity) <= quantityEpsilon && math.Abs(position.RealizedPnLCNY) <= quantityEpsilon {
			continue
		}
		positions = append(positions, position)
	}
	sort.SliceStable(positions, func(i, j int) bool {
		if positions[i].AssetType == positions[j].AssetType {
			return positions[i].MarketValueCNY > positions[j].MarketValueCNY
		}
		return positions[i].AssetType < positions[j].AssetType
	})
	return positions, nil
}

func (service *PortfolioService) AddAssetTransaction(input AssetTransactionInput) (uint, error) {
	if service == nil || service.DB == nil {
		return 0, fmt.Errorf("portfolio database is unavailable")
	}
	if err := validateAssetTransactionInput(input); err != nil {
		return 0, err
	}
	var created AssetTransaction
	err := service.DB.Transaction(func(tx *gorm.DB) error {
		instrument, err := ensureAssetInstrument(tx, input)
		if err != nil {
			return err
		}
		if input.ExternalRef != "" {
			var count int64
			if err := tx.Model(&AssetTransaction{}).Where("external_ref = ?", input.ExternalRef).Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				return fmt.Errorf("transaction external reference already exists")
			}
		}
		created, err = inputToAssetTransaction(input, instrument)
		if err != nil {
			return err
		}
		if err := validateAssetReduction(tx, created, 0); err != nil {
			return err
		}
		return tx.Create(&created).Error
	})
	if err != nil {
		return 0, fmt.Errorf("add asset transaction: %w", err)
	}
	_, _ = CreatePortfolioSnapshotWithDB(service.DB, time.Now())
	return created.ID, nil
}

func (service *PortfolioService) UpdateAssetTransaction(input AssetTransactionInput) error {
	if input.ID == 0 {
		return fmt.Errorf("transaction id is required")
	}
	if err := validateAssetTransactionInput(input); err != nil {
		return err
	}
	err := service.DB.Transaction(func(tx *gorm.DB) error {
		var existing AssetTransaction
		if err := tx.First(&existing, input.ID).Error; err != nil {
			return err
		}
		instrument, err := ensureAssetInstrument(tx, input)
		if err != nil {
			return err
		}
		updated, err := inputToAssetTransaction(input, instrument)
		if err != nil {
			return err
		}
		updated.ID = existing.ID
		updated.CreatedAt = existing.CreatedAt
		if err := validateAssetReduction(tx, updated, existing.ID); err != nil {
			return err
		}
		return tx.Model(&existing).Select(
			"instrument_id", "asset_type", "direction", "quantity", "price", "currency", "fee",
			"fx_rate_to_cny", "trade_time", "note", "metadata_json", "external_ref", "import_batch_id",
		).Updates(&updated).Error
	})
	if err != nil {
		return fmt.Errorf("update asset transaction: %w", err)
	}
	_, _ = CreatePortfolioSnapshotWithDB(service.DB, time.Now())
	return nil
}

func (service *PortfolioService) DeleteAssetTransaction(id uint) error {
	if id == 0 {
		return fmt.Errorf("transaction id is required")
	}
	result := service.DB.Delete(&AssetTransaction{}, id)
	if result.Error != nil {
		return fmt.Errorf("delete asset transaction: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	_, _ = CreatePortfolioSnapshotWithDB(service.DB, time.Now())
	return nil
}

func (service *PortfolioService) latestQuotes(instrumentIDs []uint) (map[uint]*AssetQuote, error) {
	var quotes []AssetQuote
	if err := service.DB.Where("instrument_id IN ?", instrumentIDs).Order("as_of DESC, id DESC").Find(&quotes).Error; err != nil {
		return nil, fmt.Errorf("load asset quotes: %w", err)
	}
	result := make(map[uint]*AssetQuote, len(instrumentIDs))
	for index := range quotes {
		quote := &quotes[index]
		if result[quote.InstrumentID] == nil {
			result[quote.InstrumentID] = quote
		}
	}
	return result, nil
}

func (service *PortfolioService) latestCNYRates() (map[string]float64, error) {
	var rates []FXRate
	if err := service.DB.Where("quote_currency = ? OR base_currency = ?", "CNY", "CNY").Order("as_of DESC, id DESC").Find(&rates).Error; err != nil {
		return nil, fmt.Errorf("load CNY FX rates: %w", err)
	}
	result := map[string]float64{"CNY": 1}
	for _, rate := range rates {
		base := normalizeCurrency(rate.BaseCurrency)
		quote := normalizeCurrency(rate.QuoteCurrency)
		if rate.Rate <= 0 {
			continue
		}
		if quote == "CNY" {
			if result[base] == 0 {
				result[base] = rate.Rate
			}
		} else if base == "CNY" && result[quote] == 0 {
			result[quote] = 1 / rate.Rate
		}
	}
	return result, nil
}

func ensureAssetInstrument(tx *gorm.DB, input AssetTransactionInput) (*AssetInstrument, error) {
	if input.InstrumentID > 0 {
		var instrument AssetInstrument
		if err := tx.First(&instrument, input.InstrumentID).Error; err != nil {
			return nil, fmt.Errorf("find instrument: %w", err)
		}
		if instrument.AssetType != input.AssetType {
			return nil, fmt.Errorf("instrument asset type does not match transaction")
		}
		return &instrument, nil
	}
	symbol := strings.TrimSpace(input.Symbol)
	exchange := strings.TrimSpace(input.Exchange)
	var instrument AssetInstrument
	err := tx.Where("asset_type = ? AND symbol = ? AND exchange = ?", input.AssetType, symbol, exchange).First(&instrument).Error
	if err == nil {
		updates := map[string]any{}
		if name := strings.TrimSpace(input.Name); name != "" {
			updates["name"] = name
		}
		if currency := normalizeCurrency(input.Currency); currency != "" {
			updates["currency"] = currency
		}
		if quoteSymbol := strings.TrimSpace(input.QuoteSymbol); quoteSymbol != "" {
			updates["quote_symbol"] = quoteSymbol
		}
		if unit := strings.TrimSpace(input.Unit); unit != "" {
			updates["unit"] = unit
		}
		if len(input.Metadata) > 0 {
			metadata, marshalErr := json.Marshal(input.Metadata)
			if marshalErr != nil {
				return nil, fmt.Errorf("encode instrument metadata: %w", marshalErr)
			}
			updates["metadata_json"] = string(metadata)
		}
		if len(updates) > 0 {
			if updateErr := tx.Model(&instrument).Updates(updates).Error; updateErr != nil {
				return nil, fmt.Errorf("update instrument: %w", updateErr)
			}
		}
		return &instrument, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("find instrument: %w", err)
	}
	metadata, err := json.Marshal(input.Metadata)
	if err != nil {
		return nil, fmt.Errorf("encode instrument metadata: %w", err)
	}
	instrument = AssetInstrument{
		AssetType:    input.AssetType,
		Symbol:       symbol,
		Exchange:     exchange,
		Name:         strings.TrimSpace(input.Name),
		Currency:     normalizeCurrency(input.Currency),
		QuoteSymbol:  strings.TrimSpace(input.QuoteSymbol),
		Unit:         strings.TrimSpace(input.Unit),
		MetadataJSON: string(metadata),
		IsActive:     true,
	}
	if instrument.Name == "" {
		instrument.Name = symbol
	}
	if err := tx.Create(&instrument).Error; err != nil {
		return nil, fmt.Errorf("create instrument: %w", err)
	}
	return &instrument, nil
}

func inputToAssetTransaction(input AssetTransactionInput, instrument *AssetInstrument) (AssetTransaction, error) {
	metadata, err := json.Marshal(input.Metadata)
	if err != nil {
		return AssetTransaction{}, fmt.Errorf("encode transaction metadata: %w", err)
	}
	tradeTime := input.TradeTime
	if tradeTime.IsZero() {
		tradeTime = time.Now()
	}
	fx := input.FXRateToCNY
	if normalizeCurrency(input.Currency) == "CNY" {
		fx = 1
	}
	return AssetTransaction{
		InstrumentID:  instrument.ID,
		AssetType:     input.AssetType,
		Direction:     normalizeAssetDirection(input.Direction),
		Quantity:      math.Abs(input.Quantity),
		Price:         input.Price,
		Currency:      normalizeCurrency(input.Currency),
		Fee:           input.Fee,
		FXRateToCNY:   fx,
		TradeTime:     tradeTime,
		Note:          strings.TrimSpace(input.Note),
		MetadataJSON:  string(metadata),
		ExternalRef:   strings.TrimSpace(input.ExternalRef),
		ImportBatchID: strings.TrimSpace(input.ImportBatchID),
	}, nil
}

func validateAssetTransactionInput(input AssetTransactionInput) error {
	if !isSupportedAssetType(input.AssetType) {
		return fmt.Errorf("unsupported asset type %q", input.AssetType)
	}
	if input.InstrumentID == 0 && strings.TrimSpace(input.Symbol) == "" {
		return fmt.Errorf("instrument symbol is required")
	}
	direction := normalizeAssetDirection(input.Direction)
	if direction == "" {
		return fmt.Errorf("transaction direction is required")
	}
	if input.Quantity <= 0 {
		return fmt.Errorf("transaction quantity must be greater than zero")
	}
	if input.Price < 0 || input.Fee < 0 {
		return fmt.Errorf("price and fee cannot be negative")
	}
	if normalizeCurrency(input.Currency) != "CNY" && input.FXRateToCNY < 0 {
		return fmt.Errorf("CNY exchange rate cannot be negative")
	}
	return nil
}

func isSupportedAssetType(assetType string) bool {
	for _, supported := range SupportedAssetTypes {
		if assetType == supported {
			return true
		}
	}
	return false
}

func latestTransactionFX(transactions []AssetTransaction) float64 {
	for index := len(transactions) - 1; index >= 0; index-- {
		if transactions[index].FXRateToCNY > 0 {
			return transactions[index].FXRateToCNY
		}
	}
	return 1
}

func validateAssetReduction(tx *gorm.DB, candidate AssetTransaction, excludeID uint) error {
	direction := normalizeAssetDirection(candidate.Direction)
	if direction != "sell" && direction != "close_long" && direction != "close_short" {
		return nil
	}
	var transactions []AssetTransaction
	query := tx.Where("instrument_id = ?", candidate.InstrumentID)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	if err := query.Order("trade_time ASC, id ASC").Find(&transactions).Error; err != nil {
		return err
	}
	available := 0.0
	for _, transaction := range transactions {
		itemDirection := normalizeAssetDirection(transaction.Direction)
		if candidate.AssetType == AssetTypeFutures {
			if direction == "close_short" {
				if itemDirection == "open_short" {
					available += transaction.Quantity
				}
				if itemDirection == "close_short" {
					available -= transaction.Quantity
				}
			} else {
				if itemDirection == "open_long" || itemDirection == "buy" {
					available += transaction.Quantity
				}
				if itemDirection == "close_long" || itemDirection == "sell" {
					available -= transaction.Quantity
				}
			}
		} else {
			if itemDirection == "buy" || itemDirection == "deposit" {
				available += transaction.Quantity
			}
			if itemDirection == "sell" || itemDirection == "withdraw" {
				available -= transaction.Quantity
			}
		}
	}
	if candidate.Quantity > available+quantityEpsilon {
		return fmt.Errorf("可减少数量不足：当前可用 %.8g，本次 %.8g", math.Max(0, available), candidate.Quantity)
	}
	return nil
}
