package data

import (
	"fmt"
	"go-stock/backend/db"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const CurrentAssetSchemaVersion = 3

// RunAssetSchemaMigration performs a synchronous, idempotent schema upgrade.
// It must finish before the Wails window is created.
func RunAssetSchemaMigration(database *gorm.DB, sqlitePath string) error {
	if database == nil {
		return fmt.Errorf("asset migration: database is nil")
	}
	current, err := readAssetSchemaVersion(database)
	if err != nil {
		return err
	}
	if current >= CurrentAssetSchemaVersion {
		return nil
	}
	if _, err := db.BackupSQLite(database, sqlitePath, fmt.Sprintf("asset-v%d-to-v%d", current, CurrentAssetSchemaVersion)); err != nil {
		return fmt.Errorf("asset migration backup: %w", err)
	}

	err = database.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(
			&AssetInstrument{},
			&AssetTransaction{},
			&AssetQuote{},
			&FXRate{},
			&PortfolioSnapshot{},
			&AssetSchemaVersion{},
			&AssetImportBatch{},
		); err != nil {
			return fmt.Errorf("create unified asset tables: %w", err)
		}
		if err := MigrateLegacyStockLedger(tx); err != nil {
			return err
		}
		version := AssetSchemaVersion{
			Version:   CurrentAssetSchemaVersion,
			AppliedAt: time.Now(),
			Checksum:  "unified-ledger-v3",
		}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&version).Error; err != nil {
			return fmt.Errorf("record asset schema version: %w", err)
		}
		return reconcileLegacyStockLedger(tx)
	})
	if err != nil {
		return fmt.Errorf("asset schema migration v%d: %w", CurrentAssetSchemaVersion, err)
	}
	return nil
}

func readAssetSchemaVersion(database *gorm.DB) (int, error) {
	if !database.Migrator().HasTable(&AssetSchemaVersion{}) {
		return 0, nil
	}
	var version AssetSchemaVersion
	err := database.Order("version DESC").First(&version).Error
	if err == gorm.ErrRecordNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("read asset schema version: %w", err)
	}
	return version.Version, nil
}

// MigrateLegacyStockLedger copies legacy trading records into the new ledger.
// The source rows remain untouched and stable external references make retries
// safe after a crash or interrupted startup.
func MigrateLegacyStockLedger(tx *gorm.DB) error {
	if tx == nil || !tx.Migrator().HasTable(&TradingRecord{}) {
		return nil
	}
	var records []TradingRecord
	if err := tx.Order("trading_time ASC, id ASC").Find(&records).Error; err != nil {
		return fmt.Errorf("read legacy stock ledger: %w", err)
	}
	for _, record := range records {
		symbol := strings.ToLower(strings.TrimSpace(record.StockCode))
		if symbol == "" {
			symbol = "legacy-stock-" + strconv.FormatUint(uint64(record.ID), 10)
		}
		instrument, err := ensureLegacyStockInstrument(tx, symbol, record.StockName)
		if err != nil {
			return err
		}
		externalRef := fmt.Sprintf("legacy:trading_records:%d", record.ID)
		var count int64
		if err := tx.Model(&AssetTransaction{}).Where("external_ref = ?", externalRef).Count(&count).Error; err != nil {
			return fmt.Errorf("check legacy transaction %d: %w", record.ID, err)
		}
		if count > 0 {
			continue
		}
		direction := normalizeAssetDirection(record.Direction)
		tradeTime := record.TradingTime
		if tradeTime.IsZero() {
			tradeTime = record.CreatedAt
		}
		if tradeTime.IsZero() {
			tradeTime = time.Now()
		}
		transaction := AssetTransaction{
			InstrumentID: instrument.ID,
			AssetType:    AssetTypeStock,
			Direction:    direction,
			Quantity:     float64(record.Volume),
			Price:        record.Price,
			Currency:     "CNY",
			Fee:          record.Fee,
			FXRateToCNY:  1,
			TradeTime:    tradeTime,
			Note:         record.Reason,
			ExternalRef:  externalRef,
			MetadataJSON: fmt.Sprintf("{\"legacyMarketValue\":%g,\"legacyClosePrice\":%g}", record.MarketValue, record.RecordedClosePrice),
		}
		if err := tx.Create(&transaction).Error; err != nil {
			return fmt.Errorf("migrate legacy transaction %d: %w", record.ID, err)
		}
	}
	return nil
}

func ensureLegacyStockInstrument(tx *gorm.DB, symbol, name string) (*AssetInstrument, error) {
	var instrument AssetInstrument
	err := tx.Where("asset_type = ? AND symbol = ? AND exchange = ?", AssetTypeStock, symbol, "").First(&instrument).Error
	if err == nil {
		if instrument.Name == "" && strings.TrimSpace(name) != "" {
			instrument.Name = strings.TrimSpace(name)
			if err := tx.Save(&instrument).Error; err != nil {
				return nil, fmt.Errorf("update legacy stock instrument %s: %w", symbol, err)
			}
		}
		return &instrument, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("find legacy stock instrument %s: %w", symbol, err)
	}
	instrument = AssetInstrument{
		AssetType:   AssetTypeStock,
		Symbol:      symbol,
		Name:        strings.TrimSpace(name),
		Currency:    "CNY",
		ExternalRef: "legacy:stock:" + symbol,
		IsActive:    true,
	}
	if instrument.Name == "" {
		instrument.Name = symbol
	}
	if err := tx.Create(&instrument).Error; err != nil {
		return nil, fmt.Errorf("create legacy stock instrument %s: %w", symbol, err)
	}
	return &instrument, nil
}

func reconcileLegacyStockLedger(tx *gorm.DB) error {
	if !tx.Migrator().HasTable(&TradingRecord{}) {
		return nil
	}
	var legacyCount int64
	if err := tx.Model(&TradingRecord{}).Count(&legacyCount).Error; err != nil {
		return fmt.Errorf("count legacy stock transactions: %w", err)
	}
	var migratedCount int64
	if err := tx.Model(&AssetTransaction{}).Where("external_ref LIKE ?", "legacy:trading_records:%").Count(&migratedCount).Error; err != nil {
		return fmt.Errorf("count migrated stock transactions: %w", err)
	}
	if migratedCount != legacyCount {
		return fmt.Errorf("legacy stock reconciliation failed: source=%d migrated=%d", legacyCount, migratedCount)
	}
	return nil
}

func normalizeAssetDirection(direction string) string {
	switch strings.ToLower(strings.TrimSpace(direction)) {
	case "买入", "buy", "purchase", "subscribe", "申购", "存入":
		return "buy"
	case "卖出", "sell", "redeem", "mature", "close", "赎回", "到期结算", "取出":
		return "sell"
	case "开多", "open_long":
		return "open_long"
	case "开空", "open_short":
		return "open_short"
	case "平多", "close_long":
		return "close_long"
	case "平空", "close_short":
		return "close_short"
	case "利息", "分红", "派息", "coupon", "income":
		return "income"
	case "费用", "expense":
		return "expense"
	default:
		return strings.ToLower(strings.TrimSpace(direction))
	}
}
