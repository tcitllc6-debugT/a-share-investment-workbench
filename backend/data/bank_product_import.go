package data

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"gorm.io/gorm"
)

type BankProductImportRow struct {
	RowNo         int            `json:"rowNo"`
	Institution   string         `json:"institution"`
	ProductCode   string         `json:"productCode"`
	ProductName   string         `json:"productName"`
	ValuationMode string         `json:"valuationMode"`
	Principal     float64        `json:"principal"`
	Quantity      float64        `json:"quantity"`
	NAV           float64        `json:"nav"`
	Currency      string         `json:"currency"`
	PurchaseDate  string         `json:"purchaseDate"`
	MaturityDate  string         `json:"maturityDate"`
	ExpectedYield float64        `json:"expectedYield"`
	Note          string         `json:"note"`
	Errors        []string       `json:"errors"`
	Metadata      map[string]any `json:"metadata"`
}

type BankProductImportPreview struct {
	BatchID string                 `json:"batchId"`
	Path    string                 `json:"path"`
	Rows    []BankProductImportRow `json:"rows"`
	Valid   int                    `json:"valid"`
	Invalid int                    `json:"invalid"`
}

func PreviewBankProductCSV(path string) (*BankProductImportPreview, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read bank product CSV: %w", err)
	}
	decoded := raw
	if !bytes.Equal(bytes.ToValidUTF8(raw, nil), raw) {
		decoded, err = io.ReadAll(transform.NewReader(bytes.NewReader(raw), simplifiedchinese.GB18030.NewDecoder()))
		if err != nil {
			return nil, fmt.Errorf("decode GB18030 CSV: %w", err)
		}
	}
	reader := csv.NewReader(bytes.NewReader(bytes.TrimPrefix(decoded, []byte{0xEF, 0xBB, 0xBF})))
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse bank product CSV: %w", err)
	}
	preview := &BankProductImportPreview{BatchID: uuid.NewString(), Path: path, Rows: []BankProductImportRow{}}
	if len(records) < 2 {
		return preview, fmt.Errorf("CSV must include a header and at least one data row")
	}
	header := map[string]int{}
	for index, value := range records[0] {
		header[strings.ToLower(strings.TrimSpace(value))] = index
	}
	read := func(record []string, keys ...string) string {
		for _, key := range keys {
			if index, ok := header[strings.ToLower(key)]; ok && index < len(record) {
				return strings.TrimSpace(record[index])
			}
		}
		return ""
	}
	for rowIndex, record := range records[1:] {
		row := BankProductImportRow{RowNo: rowIndex + 2, Errors: []string{}, Metadata: map[string]any{}}
		row.Institution = read(record, "发行机构", "institution")
		row.ProductCode = read(record, "产品代码", "product_code", "productcode")
		row.ProductName = read(record, "产品名称", "product_name", "productname")
		row.ValuationMode = read(record, "估值方式", "valuation_mode", "valuationmode")
		row.Currency = strings.ToUpper(read(record, "币种", "currency"))
		row.PurchaseDate = read(record, "购买日期", "purchase_date", "purchasedate")
		row.MaturityDate = read(record, "到期日期", "maturity_date", "maturitydate")
		row.Note = read(record, "备注", "note")
		row.Principal, _ = strconv.ParseFloat(read(record, "本金", "principal"), 64)
		row.Quantity, _ = strconv.ParseFloat(read(record, "份额", "quantity"), 64)
		row.NAV, _ = strconv.ParseFloat(read(record, "净值", "nav"), 64)
		row.ExpectedYield, _ = strconv.ParseFloat(strings.TrimSuffix(read(record, "业绩比较基准", "expected_yield", "expectedyield"), "%"), 64)
		if row.Institution == "" {
			row.Errors = append(row.Errors, "发行机构不能为空")
		}
		if row.ProductName == "" {
			row.Errors = append(row.Errors, "产品名称不能为空")
		}
		if row.ProductCode == "" {
			row.ProductCode = fmt.Sprintf("BANK-%d", row.RowNo)
		}
		if row.Currency == "" {
			row.Currency = "CNY"
		}
		if row.ValuationMode == "" {
			row.ValuationMode = "净值计价"
		}
		if row.Principal <= 0 && (row.Quantity <= 0 || row.NAV <= 0) {
			row.Errors = append(row.Errors, "本金或份额与净值必须大于零")
		}
		for _, value := range []struct{ name, date string }{{"购买日期", row.PurchaseDate}, {"到期日期", row.MaturityDate}} {
			if value.date != "" {
				if _, parseErr := time.Parse("2006-01-02", value.date); parseErr != nil {
					row.Errors = append(row.Errors, value.name+"格式应为 YYYY-MM-DD")
				}
			}
		}
		row.Metadata = map[string]any{"institution": row.Institution, "productCode": row.ProductCode, "valuationMode": row.ValuationMode, "principal": row.Principal, "nav": row.NAV, "expectedYield": row.ExpectedYield, "maturityDate": row.MaturityDate}
		if len(row.Errors) == 0 {
			preview.Valid++
		} else {
			preview.Invalid++
		}
		preview.Rows = append(preview.Rows, row)
	}
	return preview, nil
}

func ImportBankProductCSV(batchID, sourcePath string, rows []BankProductImportRow) error {
	if strings.TrimSpace(batchID) == "" {
		return fmt.Errorf("batch id is required")
	}
	for _, row := range rows {
		if len(row.Errors) > 0 {
			return fmt.Errorf("row %d contains validation errors", row.RowNo)
		}
	}
	database := NewPortfolioService().DB
	if database == nil {
		return fmt.Errorf("portfolio database is unavailable")
	}
	err := database.Transaction(func(tx *gorm.DB) error {
		batch := AssetImportBatch{BatchID: batchID, AssetType: AssetTypeBankWealth, SourceFile: sourcePath, RowCount: len(rows), Status: "importing"}
		if err := tx.Create(&batch).Error; err != nil {
			return err
		}
		service := NewPortfolioServiceWithDB(tx)
		for _, row := range rows {
			quantity, price := row.Quantity, 0.0
			if quantity > 0 && row.Principal > 0 {
				price = row.Principal / quantity
			} else if quantity > 0 {
				price = row.NAV
			}
			if quantity <= 0 || price <= 0 {
				quantity, price = 1, row.Principal
			}
			tradeTime := time.Now()
			if row.PurchaseDate != "" {
				tradeTime, _ = time.ParseInLocation("2006-01-02", row.PurchaseDate, time.Local)
			}
			transactionID, err := service.AddAssetTransaction(AssetTransactionInput{AssetType: AssetTypeBankWealth, Symbol: row.ProductCode, Name: row.ProductName, Direction: "buy", Quantity: quantity, Price: price, Currency: row.Currency, FXRateToCNY: 1, TradeTime: tradeTime, Note: row.Note, Metadata: row.Metadata, ExternalRef: fmt.Sprintf("bank-import:%s:%d", batchID, row.RowNo), ImportBatchID: batchID})
			if err != nil {
				return err
			}
			if row.NAV > 0 {
				var transaction AssetTransaction
				if err := tx.First(&transaction, transactionID).Error; err != nil {
					return err
				}
				now := time.Now()
				if err := tx.Create(&AssetQuote{InstrumentID: transaction.InstrumentID, Provider: "manual", Price: row.NAV, Currency: row.Currency, AsOf: now, FetchedAt: now, MetadataJSON: fmt.Sprintf(`{"importBatchId":%q}`, batchID)}).Error; err != nil {
					return err
				}
			}
		}
		return tx.Model(&batch).Updates(map[string]any{"status": "completed", "imported_count": len(rows)}).Error
	})
	if err != nil {
		return fmt.Errorf("import bank products: %w", err)
	}
	_, _ = CreatePortfolioSnapshot()
	return nil
}

func RollbackAssetImport(batchID string) error {
	database := NewPortfolioService().DB
	if database == nil {
		return fmt.Errorf("portfolio database is unavailable")
	}
	err := database.Transaction(func(tx *gorm.DB) error {
		var batch AssetImportBatch
		if err := tx.Where("batch_id = ?", batchID).First(&batch).Error; err != nil {
			return err
		}
		if batch.RolledBackAt != nil {
			return nil
		}
		if err := tx.Unscoped().Where("import_batch_id = ?", batchID).Delete(&AssetTransaction{}).Error; err != nil {
			return err
		}
		if err := tx.Where("metadata_json LIKE ?", "%\"importBatchId\":\""+batchID+"\"%").Delete(&AssetQuote{}).Error; err != nil {
			return err
		}
		now := time.Now()
		return tx.Model(&batch).Updates(map[string]any{"status": "rolled_back", "rolled_back_at": &now}).Error
	})
	if err != nil {
		return err
	}
	_, _ = CreatePortfolioSnapshot()
	return nil
}
