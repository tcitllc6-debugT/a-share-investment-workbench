package data

import (
	"go-stock/backend/db"
	"go-stock/backend/models"
	"math"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func assertFloatClose(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-10 {
		t.Fatalf("got %.12f, want %.12f", got, want)
	}
}

func TestDuoKongEMAValues(t *testing.T) {
	got := duoKongEMAValues([]float64{10, 12, 14, 13}, 3)

	want := []float64{10, 11, 12.5, 12.75}
	for i := range want {
		assertFloatClose(t, got[i], want[i])
	}
}

func TestBuildDuoKongLineSeriesUsesTwoScreenshotDerivedLines(t *testing.T) {
	series := buildDuoKongLineSeries([]KLineData{
		{Day: "1", Open: "10", High: "12", Low: "9", Close: "11"},
		{Day: "2", Open: "11", High: "13", Low: "10", Close: "12"},
		{Day: "3", Open: "8", High: "9", Low: "7", Close: "9"},
		{Day: "4", Open: "9", High: "10", Low: "8", Close: "10"},
	})

	if len(series) != 4 {
		t.Fatalf("expected 4 rows, got %d", len(series))
	}
	assertFloatClose(t, series[0].Fine, 11)
	assertFloatClose(t, series[0].Thick, 10.6)
	assertFloatClose(t, series[1].Fine, 11.181818181818182)
	assertFloatClose(t, series[1].Thick, 10.664516129032258)
	assertFloatClose(t, series[2].Fine, 10.78512396694215)
	assertFloatClose(t, series[2].Thick, 10.518418314255984)
}

func TestDuoKongSignalMatching(t *testing.T) {
	rows := []DuoKongLinePoint{
		{Day: "1", Fine: 9, Thick: 10},
		{Day: "2", Fine: 11, Thick: 10},
		{Day: "3", Fine: 12, Thick: 11},
	}

	if !duoKongMatchesSignal(rows, "fine_cross_up", 1, 0) {
		t.Fatal("expected fine_cross_up to match")
	}
	if duoKongMatchesSignal(rows, "fine_cross_down", 1, 0) {
		t.Fatal("did not expect fine_cross_down to match")
	}
	if !duoKongMatchesSignal(rows, "fine_above_thick", 1, 0) {
		t.Fatal("expected fine_above_thick to match")
	}
	if !duoKongMatchesSignal(rows, "both_rising", 1, 0) {
		t.Fatal("expected both_rising to match")
	}
	if !duoKongMatchesSignal(rows, "fine_cross_up", 1, 5) {
		t.Fatal("expected deviation threshold to use latest point")
	}
	if duoKongMatchesSignal(rows, "fine_cross_up", 1, 20) {
		t.Fatal("did not expect oversized deviation threshold to match")
	}
}

func TestDuoKongSignalWindowMatchesRecentThreeTradingDays(t *testing.T) {
	rows := []DuoKongLinePoint{
		{Day: "1", Fine: 8, Thick: 10},
		{Day: "2", Fine: 11, Thick: 10},
		{Day: "3", Fine: 12, Thick: 11},
		{Day: "4", Fine: 12.5, Thick: 11.5},
	}

	if !duoKongMatchesSignal(rows, "fine_cross_up", 3, 0) {
		t.Fatal("expected fine_cross_up from the recent 3 trading days to match")
	}
	if duoKongMatchesSignal(rows, "fine_cross_up", 1, 0) {
		t.Fatal("did not expect fine_cross_up outside the most recent 1 trading day to match")
	}
}

func TestDuoKongMarketBoardMatching(t *testing.T) {
	tests := []struct {
		name  string
		code  string
		board string
		want  bool
	}{
		{name: "all includes every stock", code: "688376", board: "ALL", want: true},
		{name: "star board", code: "688376", board: "STAR", want: true},
		{name: "chinext board", code: "300418", board: "CHINEXT", want: true},
		{name: "beijing board", code: "830799", board: "BSE", want: true},
		{name: "main board sh", code: "600519", board: "MAIN", want: true},
		{name: "main board sz", code: "000001", board: "MAIN", want: true},
		{name: "main board excludes star", code: "688376", board: "MAIN", want: false},
		{name: "star excludes chinext", code: "300418", board: "STAR", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := duoKongMatchesMarketBoard(tt.code, tt.board); got != tt.want {
				t.Fatalf("duoKongMatchesMarketBoard(%q, %q)=%v, want %v", tt.code, tt.board, got, tt.want)
			}
		})
	}
}

func TestDuoKongFineTurnUpRequiresAReversal(t *testing.T) {
	points := []DuoKongLinePoint{
		{Day: "1", Fine: 12, Thick: 10},
		{Day: "2", Fine: 10, Thick: 10},
		{Day: "3", Fine: 9, Thick: 10},
		{Day: "4", Fine: 10, Thick: 10},
	}

	if !duoKongMatchesSignal(points, "fine_turn_up", 3, 0) {
		t.Fatal("expected a fine-line down-to-up reversal to match")
	}
}

func TestDuoKongSelectionReportsTheActualCrossDate(t *testing.T) {
	resetDuoKongLineTestState()
	dao, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	db.Dao = dao
	if err := db.Dao.AutoMigrate(&models.AllStockInfo{}); err != nil {
		t.Fatalf("migrate all stock info: %v", err)
	}
	if err := db.Dao.Create(&models.AllStockInfo{
		SECUCODE:         "300418.SZ",
		SECURITYCODE:     "300418",
		SECURITYNAMEABBR: "Kunlun",
		MARKET:           "SZ",
	}).Error; err != nil {
		t.Fatalf("seed stock info: %v", err)
	}

	api := &DuoKongLineApi{kLineFetcher: &countingDuoKongKLineFetcher{rows: []KLineData{
		{Day: "2026-01-01", Open: "10", High: "10", Low: "10", Close: "10"},
		{Day: "2026-01-02", Open: "20", High: "20", Low: "20", Close: "20"},
		{Day: "2026-01-03", Open: "20", High: "20", Low: "20", Close: "20"},
		{Day: "2026-01-04", Open: "20", High: "20", Low: "20", Close: "20"},
	}}}

	results := api.GetSelections(DuoKongLineCondition{
		Signal:      "fine_cross_up",
		SignalDays:  3,
		MarketBoard: "CHINEXT",
	})
	if len(results) != 1 {
		t.Fatalf("expected one result, got %d", len(results))
	}
	if results[0].SignalDate != "2026-01-02" {
		t.Fatalf("expected actual cross date 2026-01-02, got %s", results[0].SignalDate)
	}
}

func TestDuoKongDefaultUniverseExcludesSTBSEAndRecentListings(t *testing.T) {
	dao, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	db.Dao = dao
	if err := db.Dao.AutoMigrate(&models.AllStockInfo{}, &StockBasic{}); err != nil {
		t.Fatalf("migrate stock tables: %v", err)
	}
	stocks := []models.AllStockInfo{
		{SECUCODE: "600001.SH", SECURITYCODE: "600001", SECURITYNAMEABBR: "Alpha", MARKET: "SH", NEWPRICE: "12", DEALAMOUNT: "10000000"},
		{SECUCODE: "600002.SH", SECURITYCODE: "600002", SECURITYNAMEABBR: "*ST Beta", MARKET: "SH", NEWPRICE: "12", DEALAMOUNT: "10000000"},
		{SECUCODE: "830001.BJ", SECURITYCODE: "830001", SECURITYNAMEABBR: "Beijing", MARKET: "BJ", NEWPRICE: "12", DEALAMOUNT: "10000000"},
		{SECUCODE: "300001.SZ", SECURITYCODE: "300001", SECURITYNAMEABBR: "Recent", MARKET: "SZ", NEWPRICE: "12", DEALAMOUNT: "10000000"},
	}
	if err := db.Dao.Create(&stocks).Error; err != nil {
		t.Fatalf("seed stock universe: %v", err)
	}
	old := time.Now().AddDate(0, 0, -90).Format("20060102")
	recent := time.Now().AddDate(0, 0, -20).Format("20060102")
	if err := db.Dao.Create(&[]StockBasic{
		{Symbol: "600001", ListDate: old},
		{Symbol: "600002", ListDate: old},
		{Symbol: "830001", ListDate: old},
		{Symbol: "300001", ListDate: recent},
	}).Error; err != nil {
		t.Fatalf("seed listing dates: %v", err)
	}

	selected := NewDuoKongLineApi().getSelectionStocks(DuoKongLineCondition{MarketBoard: "ALL"})
	if len(selected) != 1 || selected[0].SECURITYCODE != "600001" {
		t.Fatalf("expected only the eligible main-board stock, got %+v", selected)
	}
}

func TestDuoKongUniverseAppliesFreeMarketCapRange(t *testing.T) {
	condition := DuoKongLineCondition{MinFreeMarketCap: 50_000_000_000, MaxFreeMarketCap: 200_000_000_000}
	listingDates := map[string]time.Time{}
	withinRange := models.AllStockInfo{SECURITYCODE: "300001", SECURITYNAMEABBR: "测试科技", FREEMARKETCAP: "100000000000"}
	tooSmall := models.AllStockInfo{SECURITYCODE: "300002", SECURITYNAMEABBR: "测试制造", FREEMARKETCAP: "49000000000"}
	unknownCap := models.AllStockInfo{SECURITYCODE: "300003", SECURITYNAMEABBR: "测试能源"}

	if !duoKongMatchesUniverse(withinRange, condition, listingDates) {
		t.Fatal("expected stock inside the free-market-cap range to be included")
	}
	if duoKongMatchesUniverse(tooSmall, condition, listingDates) {
		t.Fatal("expected stock below the free-market-cap range to be excluded")
	}
	if duoKongMatchesUniverse(unknownCap, condition, listingDates) {
		t.Fatal("expected stock without a free-market-cap quote to be excluded when the filter is active")
	}
}
