package data

import "time"

type AssetTransactionInput struct {
	ID            uint           `json:"id"`
	InstrumentID  uint           `json:"instrumentId"`
	AssetType     string         `json:"assetType"`
	Symbol        string         `json:"symbol"`
	Name          string         `json:"name"`
	Exchange      string         `json:"exchange"`
	QuoteSymbol   string         `json:"quoteSymbol"`
	Unit          string         `json:"unit"`
	Direction     string         `json:"direction"`
	Quantity      float64        `json:"quantity"`
	Price         float64        `json:"price"`
	Currency      string         `json:"currency"`
	Fee           float64        `json:"fee"`
	FXRateToCNY   float64        `json:"fxRateToCny"`
	TradeTime     time.Time      `json:"tradeTime"`
	Note          string         `json:"note"`
	Metadata      map[string]any `json:"metadata"`
	ExternalRef   string         `json:"externalRef"`
	ImportBatchID string         `json:"importBatchId"`
}

type AssetPosition struct {
	InstrumentID       uint           `json:"instrumentId"`
	AssetType          string         `json:"assetType"`
	Symbol             string         `json:"symbol"`
	Name               string         `json:"name"`
	Exchange           string         `json:"exchange"`
	Unit               string         `json:"unit"`
	Currency           string         `json:"currency"`
	Quantity           float64        `json:"quantity"`
	AverageCost        float64        `json:"averageCost"`
	LatestPrice        float64        `json:"latestPrice"`
	FXRateToCNY        float64        `json:"fxRateToCny"`
	CostCNY            float64        `json:"costCny"`
	MarketValueCNY     float64        `json:"marketValueCny"`
	FloatingPnLCNY     float64        `json:"floatingPnlCny"`
	FloatingPnLPct     float64        `json:"floatingPnlPct"`
	RealizedPnLCNY     float64        `json:"realizedPnlCny"`
	NetCashFlowCNY     float64        `json:"netCashFlowCny"`
	NotionalCNY        float64        `json:"notionalCny"`
	MarginCNY          float64        `json:"marginCny"`
	AccruedInterestCNY float64        `json:"accruedInterestCny"`
	QuoteAsOf          *time.Time     `json:"quoteAsOf"`
	QuoteProvider      string         `json:"quoteProvider"`
	QuoteStale         bool           `json:"quoteStale"`
	Metadata           map[string]any `json:"metadata"`
}

type AssetClassSummary struct {
	AssetType      string  `json:"assetType"`
	PositionCount  int     `json:"positionCount"`
	CostCNY        float64 `json:"costCny"`
	MarketValueCNY float64 `json:"marketValueCny"`
	FloatingPnLCNY float64 `json:"floatingPnlCny"`
	RealizedPnLCNY float64 `json:"realizedPnlCny"`
	NetCashFlowCNY float64 `json:"netCashFlowCny"`
	AllocationPct  float64 `json:"allocationPct"`
}

type CurrencyExposure struct {
	Currency       string  `json:"currency"`
	MarketValueCNY float64 `json:"marketValueCny"`
	AllocationPct  float64 `json:"allocationPct"`
}

// FXQuotePoint is a compact historical point used by the portfolio FX board.
// It deliberately keeps the source timestamp so the UI can show real quote movement.
type FXQuotePoint struct {
	At   time.Time `json:"at"`
	Rate float64   `json:"rate"`
}

// MajorFXQuote represents one unit of BaseCurrency converted into CNY.
type MajorFXQuote struct {
	BaseCurrency  string         `json:"baseCurrency"`
	QuoteCurrency string         `json:"quoteCurrency"`
	Rate          float64        `json:"rate"`
	ChangePercent float64        `json:"changePercent"`
	Provider      string         `json:"provider"`
	AsOf          time.Time      `json:"asOf"`
	FetchedAt     time.Time      `json:"fetchedAt"`
	History       []FXQuotePoint `json:"history"`
}

type PortfolioQuoteSummary struct {
	Fresh  int `json:"fresh"`
	Stale  int `json:"stale"`
	Manual int `json:"manual"`
	Error  int `json:"error"`
}

type PortfolioRecentTransaction struct {
	ID             uint      `json:"id"`
	AssetType      string    `json:"assetType"`
	InstrumentName string    `json:"instrumentName"`
	InstrumentCode string    `json:"instrumentCode"`
	Action         string    `json:"action"`
	Quantity       float64   `json:"quantity"`
	AmountCNY      float64   `json:"amountCny"`
	TradedAt       time.Time `json:"tradedAt"`
}

type PortfolioOverview struct {
	AsOf               time.Time                    `json:"asOf"`
	TotalCostCNY       float64                      `json:"totalCostCny"`
	MarketValueCNY     float64                      `json:"marketValueCny"`
	FloatingPnLCNY     float64                      `json:"floatingPnlCny"`
	FloatingPnLPct     float64                      `json:"floatingPnlPct"`
	RealizedPnLCNY     float64                      `json:"realizedPnlCny"`
	NetCashFlowCNY     float64                      `json:"netCashFlowCny"`
	PositionCount      int                          `json:"positionCount"`
	StaleQuoteCount    int                          `json:"staleQuoteCount"`
	ByAssetType        map[string]AssetClassSummary `json:"byAssetType"`
	CurrencyExposure   []CurrencyExposure           `json:"currencyExposure"`
	Positions          []AssetPosition              `json:"positions"`
	QuoteSummary       PortfolioQuoteSummary        `json:"quoteSummary"`
	RecentTransactions []PortfolioRecentTransaction `json:"recentTransactions"`
}

type PortfolioSnapshotPage struct {
	List     []PortfolioSnapshot `json:"list"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"pageSize"`
}
