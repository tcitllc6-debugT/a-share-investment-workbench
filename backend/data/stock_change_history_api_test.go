package data

import (
	"go-stock/backend/db"
	"go-stock/backend/models"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestGetHistoryListDefaultsPageSize(t *testing.T) {
	dao, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	db.Dao = dao
	if err := db.Dao.AutoMigrate(&models.StockChangeHistory{}); err != nil {
		t.Fatalf("migrate stock change history: %v", err)
	}
	if err := db.Dao.Create(&models.StockChangeHistory{
		ChangeDate: "2026-07-09",
		ChangeTime: "10:00:00",
		StockCode:  "300418",
		StockName:  "昆仑万维",
		ChangeType: 8201,
		TypeName:   "火箭发射",
	}).Error; err != nil {
		t.Fatalf("seed history: %v", err)
	}

	got, err := NewStockChangeHistoryService().GetHistoryList(models.StockChangeHistoryQuery{Page: 1})
	if err != nil {
		t.Fatalf("GetHistoryList returned error: %v", err)
	}
	if got.PageSize <= 0 {
		t.Fatalf("expected default page size, got %d", got.PageSize)
	}
	if got.TotalPages != 1 {
		t.Fatalf("expected 1 total page, got %d", got.TotalPages)
	}
	if len(got.List) != 1 {
		t.Fatalf("expected 1 row, got %d", len(got.List))
	}
}
