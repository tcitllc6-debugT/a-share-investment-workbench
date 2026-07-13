package data

import (
	"encoding/json"
	"math"
	"sort"
	"strings"
	"time"
)

const quantityEpsilon = 1e-9

type assetLot struct {
	quantity     float64
	unitCostCNY  float64
	entryPrice   float64
	entryFXToCNY float64
}

func calculateAssetPosition(instrument AssetInstrument, transactions []AssetTransaction, quote *AssetQuote, currentFX float64) AssetPosition {
	metadata := decodeMetadata(instrument.MetadataJSON)
	position := AssetPosition{
		InstrumentID: instrument.ID,
		AssetType:    instrument.AssetType,
		Symbol:       instrument.Symbol,
		Name:         instrument.Name,
		Exchange:     instrument.Exchange,
		Unit:         instrument.Unit,
		Currency:     normalizeCurrency(instrument.Currency),
		FXRateToCNY:  currentFX,
		Metadata:     metadata,
		QuoteStale:   true,
	}
	if position.FXRateToCNY <= 0 {
		position.FXRateToCNY = 1
	}
	if quote != nil && quote.Price > 0 {
		position.LatestPrice = quote.Price
		position.QuoteAsOf = &quote.AsOf
		position.QuoteProvider = quote.Provider
		position.QuoteStale = isAssetQuoteStale(instrument.AssetType, quote, time.Now())
	}

	sort.SliceStable(transactions, func(i, j int) bool {
		if transactions[i].TradeTime.Equal(transactions[j].TradeTime) {
			return transactions[i].ID < transactions[j].ID
		}
		return transactions[i].TradeTime.Before(transactions[j].TradeTime)
	})

	if instrument.AssetType == AssetTypeFutures {
		calculateFuturesPosition(&position, transactions, metadata)
	} else {
		calculateSpotPosition(&position, transactions, metadata)
	}
	if position.LatestPrice <= 0 {
		position.LatestPrice = position.AverageCost
	}
	applyCurrentValuation(&position, metadata)
	return position
}

func isAssetQuoteStale(assetType string, quote *AssetQuote, now time.Time) bool {
	if quote == nil || quote.Price <= 0 {
		return true
	}
	if strings.EqualFold(strings.TrimSpace(quote.Provider), "manual") {
		return false
	}
	if quote.IsStale {
		return true
	}
	observedAt := quote.FetchedAt
	if observedAt.IsZero() {
		observedAt = quote.AsOf
	}
	if observedAt.IsZero() {
		return true
	}
	ttl := 15 * time.Minute
	switch assetType {
	case AssetTypeCrypto, AssetTypeForex:
		ttl = 5 * time.Minute
	case AssetTypeBond:
		ttl = 30 * time.Minute
	case AssetTypeBankWealth:
		ttl = 24 * time.Hour
	}
	return now.Sub(observedAt) > ttl
}

func calculateSpotPosition(position *AssetPosition, transactions []AssetTransaction, instrumentMeta map[string]any) {
	lots := make([]assetLot, 0)
	for _, transaction := range transactions {
		quantity := math.Abs(transaction.Quantity)
		if quantity <= quantityEpsilon {
			continue
		}
		direction := normalizeAssetDirection(transaction.Direction)
		fx := transaction.FXRateToCNY
		if fx <= 0 {
			fx = position.FXRateToCNY
		}
		transactionMeta := mergeMetadata(instrumentMeta, decodeMetadata(transaction.MetadataJSON))
		totalLocal := transactionLocalValue(position.AssetType, quantity, transaction.Price, transactionMeta)
		switch direction {
		case "buy", "deposit", "open_long":
			costCNY := totalLocal*fx + transaction.Fee*fx
			lots = append(lots, assetLot{quantity: quantity, unitCostCNY: costCNY / quantity, entryPrice: transaction.Price, entryFXToCNY: fx})
			position.NetCashFlowCNY -= costCNY
		case "sell", "withdraw", "close_long":
			proceedsCNY := totalLocal*fx - transaction.Fee*fx
			costCNY, removed := consumeAssetLots(&lots, quantity)
			if removed > quantityEpsilon {
				position.RealizedPnLCNY += proceedsCNY*(removed/quantity) - costCNY
			}
			position.NetCashFlowCNY += proceedsCNY
		case "income", "coupon", "dividend":
			income := totalLocal*fx - transaction.Fee*fx
			position.RealizedPnLCNY += income
			position.NetCashFlowCNY += income
		case "expense":
			expense := totalLocal*fx + transaction.Fee*fx
			position.RealizedPnLCNY -= expense
			position.NetCashFlowCNY -= expense
		}
	}
	for _, lot := range lots {
		position.Quantity += lot.quantity
		position.CostCNY += lot.quantity * lot.unitCostCNY
	}
	if position.Quantity > quantityEpsilon {
		position.AverageCost = position.CostCNY / position.Quantity / position.FXRateToCNY
		if position.AssetType == AssetTypeBond {
			faceValue := metadataNumber(instrumentMeta, "faceValue", 100)
			if faceValue > 0 {
				position.AverageCost = position.AverageCost / faceValue * 100
			}
		}
	}
}

func calculateFuturesPosition(position *AssetPosition, transactions []AssetTransaction, metadata map[string]any) {
	multiplier := metadataNumber(metadata, "multiplier", 1)
	marginRatio := metadataNumber(metadata, "marginRatio", 0.1)
	if multiplier <= 0 {
		multiplier = 1
	}
	if marginRatio < 0 {
		marginRatio = 0
	}
	longLots := make([]assetLot, 0)
	shortLots := make([]assetLot, 0)
	for _, transaction := range transactions {
		quantity := math.Abs(transaction.Quantity)
		if quantity <= quantityEpsilon {
			continue
		}
		fx := transaction.FXRateToCNY
		if fx <= 0 {
			fx = position.FXRateToCNY
		}
		direction := normalizeAssetDirection(transaction.Direction)
		switch direction {
		case "open_long", "buy":
			longLots = append(longLots, assetLot{quantity: quantity, entryPrice: transaction.Price, entryFXToCNY: fx})
			position.NetCashFlowCNY -= transaction.Fee * fx
		case "open_short":
			shortLots = append(shortLots, assetLot{quantity: quantity, entryPrice: transaction.Price, entryFXToCNY: fx})
			position.NetCashFlowCNY -= transaction.Fee * fx
		case "close_long", "sell":
			realized := closeFuturesLots(&longLots, quantity, transaction.Price, multiplier, false) - transaction.Fee*fx
			position.RealizedPnLCNY += realized
			position.NetCashFlowCNY += realized
		case "close_short":
			realized := closeFuturesLots(&shortLots, quantity, transaction.Price, multiplier, true) - transaction.Fee*fx
			position.RealizedPnLCNY += realized
			position.NetCashFlowCNY += realized
		case "income":
			income := transaction.Price*quantity*fx - transaction.Fee*fx
			position.RealizedPnLCNY += income
			position.NetCashFlowCNY += income
		case "expense":
			expense := transaction.Price*quantity*fx + transaction.Fee*fx
			position.RealizedPnLCNY -= expense
			position.NetCashFlowCNY -= expense
		}
	}
	longQuantity, shortQuantity := 0.0, 0.0
	entryValue := 0.0
	for _, lot := range longLots {
		longQuantity += lot.quantity
		entryValue += lot.quantity * lot.entryPrice
	}
	for _, lot := range shortLots {
		shortQuantity += lot.quantity
		entryValue += lot.quantity * lot.entryPrice
	}
	position.Quantity = longQuantity - shortQuantity
	totalQuantity := longQuantity + shortQuantity
	if totalQuantity > quantityEpsilon {
		position.AverageCost = entryValue / totalQuantity
	}
	position.NotionalCNY = totalQuantity * position.LatestPrice * multiplier * position.FXRateToCNY
	position.MarginCNY = position.NotionalCNY * marginRatio
	position.CostCNY = position.MarginCNY
	for _, lot := range longLots {
		position.FloatingPnLCNY += (position.LatestPrice - lot.entryPrice) * lot.quantity * multiplier * position.FXRateToCNY
	}
	for _, lot := range shortLots {
		position.FloatingPnLCNY += (lot.entryPrice - position.LatestPrice) * lot.quantity * multiplier * position.FXRateToCNY
	}
	position.MarketValueCNY = position.MarginCNY + position.FloatingPnLCNY
}

func applyCurrentValuation(position *AssetPosition, metadata map[string]any) {
	if position.AssetType == AssetTypeFutures {
		if position.CostCNY > quantityEpsilon {
			position.FloatingPnLPct = position.FloatingPnLCNY / position.CostCNY * 100
		}
		return
	}
	if position.AssetType == AssetTypeBond {
		accrued := metadataNumber(metadata, "accruedInterest", 0)
		faceValue := metadataNumber(metadata, "faceValue", 100)
		position.MarketValueCNY = position.Quantity * position.LatestPrice / 100 * faceValue * position.FXRateToCNY
		position.AccruedInterestCNY = position.Quantity * accrued / 100 * faceValue * position.FXRateToCNY
		position.MarketValueCNY += position.AccruedInterestCNY
	} else {
		localValue := transactionLocalValue(position.AssetType, position.Quantity, position.LatestPrice, metadata)
		position.MarketValueCNY = localValue * position.FXRateToCNY
	}
	position.FloatingPnLCNY = position.MarketValueCNY - position.CostCNY
	if position.CostCNY > quantityEpsilon {
		position.FloatingPnLPct = position.FloatingPnLCNY / position.CostCNY * 100
	}
}

func transactionLocalValue(assetType string, quantity, price float64, metadata map[string]any) float64 {
	switch assetType {
	case AssetTypeBond:
		faceValue := metadataNumber(metadata, "faceValue", 100)
		accrued := metadataNumber(metadata, "accruedInterest", 0)
		return quantity * (price + accrued) / 100 * faceValue
	case AssetTypePreciousMetal:
		unitToQuote := metadataNumber(metadata, "unitToQuote", 1)
		return quantity * unitToQuote * price
	case AssetTypeFutures:
		multiplier := metadataNumber(metadata, "multiplier", 1)
		return quantity * price * multiplier
	case AssetTypeBankWealth:
		valuationMode := strings.ToLower(metadataString(metadata, "valuationMode", "nav"))
		if valuationMode == "principal" {
			return quantity * price
		}
		return quantity * price
	default:
		return quantity * price
	}
}

func consumeAssetLots(lots *[]assetLot, quantity float64) (float64, float64) {
	remaining := quantity
	cost := 0.0
	removed := 0.0
	for remaining > quantityEpsilon && len(*lots) > 0 {
		lot := &(*lots)[0]
		take := math.Min(remaining, lot.quantity)
		cost += take * lot.unitCostCNY
		removed += take
		lot.quantity -= take
		remaining -= take
		if lot.quantity <= quantityEpsilon {
			*lots = (*lots)[1:]
		}
	}
	return cost, removed
}

func closeFuturesLots(lots *[]assetLot, quantity, closePrice, multiplier float64, short bool) float64 {
	remaining := quantity
	realized := 0.0
	for remaining > quantityEpsilon && len(*lots) > 0 {
		lot := &(*lots)[0]
		take := math.Min(remaining, lot.quantity)
		difference := closePrice - lot.entryPrice
		if short {
			difference = -difference
		}
		realized += difference * take * multiplier * lot.entryFXToCNY
		lot.quantity -= take
		remaining -= take
		if lot.quantity <= quantityEpsilon {
			*lots = (*lots)[1:]
		}
	}
	return realized
}

func decodeMetadata(raw string) map[string]any {
	result := map[string]any{}
	if strings.TrimSpace(raw) != "" {
		_ = json.Unmarshal([]byte(raw), &result)
	}
	return result
}

func mergeMetadata(base, override map[string]any) map[string]any {
	result := make(map[string]any, len(base)+len(override))
	for key, value := range base {
		result[key] = value
	}
	for key, value := range override {
		result[key] = value
	}
	return result
}

func metadataNumber(metadata map[string]any, key string, fallback float64) float64 {
	value, exists := metadata[key]
	if !exists {
		return fallback
	}
	switch typed := value.(type) {
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	case json.Number:
		if parsed, err := typed.Float64(); err == nil {
			return parsed
		}
	}
	return fallback
}

func metadataString(metadata map[string]any, key, fallback string) string {
	if value, ok := metadata[key].(string); ok && strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}

func normalizeCurrency(currency string) string {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" || currency == "RMB" || currency == "CNH" {
		return "CNY"
	}
	return currency
}
