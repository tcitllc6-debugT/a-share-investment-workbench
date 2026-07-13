package data

import (
	"go-stock/backend/db"
	"os"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"golang.org/x/text/encoding/simplifiedchinese"
	"gorm.io/gorm"
)

func TestPreviewBankProductCSVSupportsUTF8AndGB18030(t *testing.T) {
	content := "发行机构,产品代码,产品名称,估值方式,本金,份额,净值,币种,购买日期,到期日期\n测试银行,P001,稳健一号,净值计价,10000,10000,1.02,CNY,2026-01-02,2027-01-02\n"
	for _, tc := range []struct {
		name  string
		bytes []byte
	}{
		{"utf8", []byte(content)},
		{"gb18030", func() []byte {
			value, err := simplifiedchinese.GB18030.NewEncoder().Bytes([]byte(content))
			if err != nil {
				t.Fatal(err)
			}
			return value
		}()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "bank.csv")
			if err := os.WriteFile(path, tc.bytes, 0600); err != nil {
				t.Fatal(err)
			}
			preview, err := PreviewBankProductCSV(path)
			if err != nil {
				t.Fatal(err)
			}
			if preview.Valid != 1 || preview.Invalid != 0 {
				t.Fatalf("preview=%+v", preview)
			}
			if preview.Rows[0].ProductCode != "P001" {
				t.Fatalf("unexpected row: %+v", preview.Rows[0])
			}
		})
	}
}

func TestImportAndRollbackBankProductBatch(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(&AssetInstrument{}, &AssetTransaction{}, &AssetQuote{}, &FXRate{}, &PortfolioSnapshot{}, &AssetImportBatch{}); err != nil {
		t.Fatal(err)
	}
	previous := db.Dao
	db.Dao = database
	t.Cleanup(func() { db.Dao = previous })
	rows := []BankProductImportRow{{RowNo: 2, Institution: "测试银行", ProductCode: "P001", ProductName: "稳健一号", ValuationMode: "净值计价", Principal: 10000, Quantity: 10000, NAV: 1.02, Currency: "CNY", Metadata: map[string]any{"institution": "测试银行", "valuationMode": "nav"}}}
	if err := ImportBankProductCSV("batch-1", "test.csv", rows); err != nil {
		t.Fatal(err)
	}
	var count int64
	if err := database.Model(&AssetTransaction{}).Where("import_batch_id = ?", "batch-1").Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("count=%d err=%v", count, err)
	}
	var snapshotCount int64
	if err := database.Model(&PortfolioSnapshot{}).Count(&snapshotCount).Error; err != nil || snapshotCount != 1 {
		t.Fatalf("snapshot count=%d err=%v", snapshotCount, err)
	}
	positions, err := NewPortfolioServiceWithDB(database).GetAssetPositions(AssetTypeBankWealth)
	if err != nil || len(positions) != 1 {
		t.Fatalf("positions=%+v err=%v", positions, err)
	}
	if positions[0].CostCNY != 10000 || positions[0].MarketValueCNY != 10200 {
		t.Fatalf("unexpected valuation: %+v", positions[0])
	}
	if err := RollbackAssetImport("batch-1"); err != nil {
		t.Fatal(err)
	}
	if err := database.Model(&AssetTransaction{}).Where("import_batch_id = ?", "batch-1").Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("after rollback count=%d err=%v", count, err)
	}
	if err := database.Model(&AssetQuote{}).Where("metadata_json LIKE ?", "%batch-1%").Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("quotes after rollback=%d err=%v", count, err)
	}
}
