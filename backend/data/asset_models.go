package data

import (
	"time"

	"gorm.io/gorm"
)

const (
	AssetTypeStock         = "stock"
	AssetTypeFund          = "fund"
	AssetTypeBond          = "bond"
	AssetTypePreciousMetal = "precious_metal"
	AssetTypeForex         = "forex"
	AssetTypeFutures       = "futures"
	AssetTypeCrypto        = "crypto"
	AssetTypeBankWealth    = "bank_wealth"
)

var SupportedAssetTypes = []string{
	AssetTypeStock,
	AssetTypeFund,
	AssetTypeBond,
	AssetTypePreciousMetal,
	AssetTypeForex,
	AssetTypeFutures,
	AssetTypeCrypto,
	AssetTypeBankWealth,
}

// AssetInstrument is the stable identity shared by transactions and quotes.
// Asset-specific fields live in MetadataJSON so the core schema can evolve
// without widening the table for every new product type.
type AssetInstrument struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	AssetType    string         `gorm:"size:32;not null;uniqueIndex:idx_asset_instrument_identity,priority:1" json:"assetType"`
	Symbol       string         `gorm:"size:128;not null;uniqueIndex:idx_asset_instrument_identity,priority:2" json:"symbol"`
	Exchange     string         `gorm:"size:64;uniqueIndex:idx_asset_instrument_identity,priority:3" json:"exchange"`
	Name         string         `gorm:"size:256;not null" json:"name"`
	Currency     string         `gorm:"size:12;not null;default:CNY" json:"currency"`
	QuoteSymbol  string         `gorm:"size:128;index" json:"quoteSymbol"`
	Unit         string         `gorm:"size:32" json:"unit"`
	MetadataJSON string         `gorm:"type:text" json:"metadataJson"`
	ExternalRef  string         `gorm:"size:256;index" json:"externalRef"`
	IsActive     bool           `gorm:"not null;default:true;index" json:"isActive"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (AssetInstrument) TableName() string { return "asset_instruments" }

// AssetTransaction is the authoritative ledger. Positions are always derived
// from ordered transactions; callers must not persist calculated balances.
type AssetTransaction struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	InstrumentID  uint           `gorm:"not null;index" json:"instrumentId"`
	AssetType     string         `gorm:"size:32;not null;index" json:"assetType"`
	Direction     string         `gorm:"size:32;not null;index" json:"direction"`
	Quantity      float64        `gorm:"not null" json:"quantity"`
	Price         float64        `gorm:"not null" json:"price"`
	Currency      string         `gorm:"size:12;not null;default:CNY" json:"currency"`
	Fee           float64        `gorm:"not null;default:0" json:"fee"`
	FXRateToCNY   float64        `gorm:"not null;default:0" json:"fxRateToCny"`
	TradeTime     time.Time      `gorm:"not null;index" json:"tradeTime"`
	Note          string         `gorm:"type:text" json:"note"`
	MetadataJSON  string         `gorm:"type:text" json:"metadataJson"`
	ExternalRef   string         `gorm:"size:256;index" json:"externalRef"`
	ImportBatchID string         `gorm:"size:64;index" json:"importBatchId"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (AssetTransaction) TableName() string { return "asset_transactions" }

type AssetQuote struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	InstrumentID uint      `gorm:"not null;uniqueIndex:idx_asset_quote_point,priority:1" json:"instrumentId"`
	Provider     string    `gorm:"size:64;not null;uniqueIndex:idx_asset_quote_point,priority:2" json:"provider"`
	Price        float64   `gorm:"not null" json:"price"`
	Currency     string    `gorm:"size:12;not null;default:CNY" json:"currency"`
	AsOf         time.Time `gorm:"not null;index;uniqueIndex:idx_asset_quote_point,priority:3" json:"asOf"`
	FetchedAt    time.Time `gorm:"not null;index" json:"fetchedAt"`
	IsStale      bool      `gorm:"not null;default:false;index" json:"isStale"`
	MetadataJSON string    `gorm:"type:text" json:"metadataJson"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func (AssetQuote) TableName() string { return "asset_quotes" }

// FXRate means one unit of BaseCurrency equals Rate units of QuoteCurrency.
type FXRate struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	BaseCurrency  string    `gorm:"size:12;not null;uniqueIndex:idx_fx_rate_point,priority:1" json:"baseCurrency"`
	QuoteCurrency string    `gorm:"size:12;not null;uniqueIndex:idx_fx_rate_point,priority:2" json:"quoteCurrency"`
	Rate          float64   `gorm:"not null" json:"rate"`
	Provider      string    `gorm:"size:64;not null;uniqueIndex:idx_fx_rate_point,priority:3" json:"provider"`
	AsOf          time.Time `gorm:"not null;index;uniqueIndex:idx_fx_rate_point,priority:4" json:"asOf"`
	FetchedAt     time.Time `gorm:"not null;index" json:"fetchedAt"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func (FXRate) TableName() string { return "fx_rates" }

type PortfolioSnapshot struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	SnapshotDate    string    `gorm:"size:10;not null;uniqueIndex" json:"snapshotDate"`
	TotalCostCNY    float64   `json:"totalCostCny"`
	MarketValueCNY  float64   `json:"marketValueCny"`
	FloatingPnLCNY  float64   `gorm:"column:floating_pnl_cny" json:"floatingPnlCny"`
	RealizedPnLCNY  float64   `gorm:"column:realized_pnl_cny" json:"realizedPnlCny"`
	NetCashFlowCNY  float64   `json:"netCashFlowCny"`
	BreakdownJSON   string    `gorm:"type:text" json:"breakdownJson"`
	CurrencyJSON    string    `gorm:"type:text" json:"currencyJson"`
	StaleQuoteCount int       `json:"staleQuoteCount"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func (PortfolioSnapshot) TableName() string { return "portfolio_snapshots" }

type AssetSchemaVersion struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Version   int       `gorm:"not null;uniqueIndex" json:"version"`
	AppliedAt time.Time `gorm:"not null" json:"appliedAt"`
	Checksum  string    `gorm:"size:128" json:"checksum"`
}

func (AssetSchemaVersion) TableName() string { return "asset_schema_versions" }

type AssetImportBatch struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	BatchID       string     `gorm:"size:64;not null;uniqueIndex" json:"batchId"`
	AssetType     string     `gorm:"size:32;not null;index" json:"assetType"`
	SourceFile    string     `gorm:"size:512" json:"sourceFile"`
	RowCount      int        `json:"rowCount"`
	ImportedCount int        `json:"importedCount"`
	Status        string     `gorm:"size:24;not null;index" json:"status"`
	CreatedAt     time.Time  `json:"createdAt"`
	RolledBackAt  *time.Time `json:"rolledBackAt"`
}

func (AssetImportBatch) TableName() string { return "asset_import_batches" }
