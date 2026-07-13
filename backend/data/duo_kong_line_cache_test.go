package data

import (
	"context"
	"go-stock/backend/db"
	"go-stock/backend/models"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type countingDuoKongKLineFetcher struct {
	calls         int
	requestedDays []int64
	rows          []KLineData
}

type blockingDuoKongKLineFetcher struct {
	rows    []KLineData
	started chan struct{}
	release <-chan struct{}
}

func (f *blockingDuoKongKLineFetcher) GetCommonKLineData(stockCode string, kLineType string, days int64) *[]KLineData {
	select {
	case f.started <- struct{}{}:
	default:
	}
	<-f.release
	rows := append([]KLineData(nil), f.rows...)
	return &rows
}

func (f *countingDuoKongKLineFetcher) GetCommonKLineData(stockCode string, kLineType string, days int64) *[]KLineData {
	f.calls++
	f.requestedDays = append(f.requestedDays, days)
	rows := append([]KLineData(nil), f.rows...)
	return &rows
}

func TestDuoKongKLineCacheRefreshesOnlyTheRecentTail(t *testing.T) {
	resetDuoKongLineTestState()
	dao, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	db.Dao = dao
	if err := db.Dao.AutoMigrate(&models.DuoKongKLineCache{}); err != nil {
		t.Fatalf("migrate persistent cache table: %v", err)
	}
	staleAt := time.Now().Add(-7 * 24 * time.Hour)
	if err := db.Dao.Create(&[]models.DuoKongKLineCache{
		{StockCode: "sz300418", KLineType: "day", Day: "2026-07-08", Close: "10", FetchedAt: staleAt},
		{StockCode: "sz300418", KLineType: "day", Day: "2026-07-09", Close: "11", FetchedAt: staleAt},
	}).Error; err != nil {
		t.Fatalf("seed stale cache: %v", err)
	}
	fetcher := &countingDuoKongKLineFetcher{rows: []KLineData{
		{Day: "2026-07-09", Close: "11.5"},
		{Day: "2026-07-10", Close: "12"},
	}}
	api := &DuoKongLineApi{kLineFetcher: fetcher}
	rows := api.getCachedKLineData("sz300418", "day", 120)

	if len(fetcher.requestedDays) != 1 || fetcher.requestedDays[0] != 5 {
		t.Fatalf("expected a five-day tail refresh, got requests %#v", fetcher.requestedDays)
	}
	if len(rows) != 3 || rows[1].Close != "11.5" || rows[2].Day != "2026-07-10" {
		t.Fatalf("expected stale cache and refreshed tail to be merged, got %+v", rows)
	}
}

func resetDuoKongLineTestState() {
	db.Dao = nil
	duoKongKLineCacheMu.Lock()
	duoKongKLineCache = map[string]duoKongKLineCacheEntry{}
	duoKongKLineCacheMu.Unlock()

	duoKongScanTaskMu.Lock()
	duoKongScanTasks = map[string]*DuoKongLineScanTask{}
	duoKongScanCancels = map[string]context.CancelFunc{}
	duoKongScanTaskMu.Unlock()
}

func TestDuoKongKLineCacheReusesFetchedData(t *testing.T) {
	resetDuoKongLineTestState()

	fetcher := &countingDuoKongKLineFetcher{
		rows: []KLineData{
			{Day: "1", Open: "10", High: "11", Low: "9", Close: "10"},
			{Day: "2", Open: "10", High: "12", Low: "9", Close: "11"},
		},
	}
	api := &DuoKongLineApi{kLineFetcher: fetcher}

	first := api.getCachedKLineData("sz300418", "day", 90)
	second := api.getCachedKLineData("sz300418", "day", 90)

	if len(first) != 2 || len(second) != 2 {
		t.Fatalf("expected cached kline rows, got %d and %d", len(first), len(second))
	}
	if fetcher.calls != 1 {
		t.Fatalf("expected one upstream kline fetch, got %d", fetcher.calls)
	}
}

func TestDuoKongKLineCacheSurvivesMemoryReset(t *testing.T) {
	resetDuoKongLineTestState()
	dao, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	db.Dao = dao
	if err := db.Dao.AutoMigrate(&models.DuoKongKLineCache{}); err != nil {
		t.Fatalf("migrate persistent cache table: %v", err)
	}

	fetcher := &countingDuoKongKLineFetcher{rows: []KLineData{
		{Day: "1", Open: "10", High: "11", Low: "9", Close: "10"},
		{Day: "2", Open: "10", High: "12", Low: "9", Close: "11"},
	}}
	api := &DuoKongLineApi{kLineFetcher: fetcher}
	api.getCachedKLineData("sz300418", "day", 120)

	duoKongKLineCacheMu.Lock()
	duoKongKLineCache = map[string]duoKongKLineCacheEntry{}
	duoKongKLineCacheMu.Unlock()

	rows := api.getCachedKLineData("sz300418", "day", 120)
	if len(rows) != 2 {
		t.Fatalf("expected persistent cache rows, got %d", len(rows))
	}
	if fetcher.calls != 1 {
		t.Fatalf("expected persistent cache to avoid a second fetch, got %d calls", fetcher.calls)
	}
}

func TestDuoKongBackgroundScanRecordsProgressAndResults(t *testing.T) {
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
		SECURITYNAMEABBR: "昆仑万维",
		MARKET:           "SZ",
		INDUSTRY:         "互联网",
		CONCEPT:          "AI应用",
	}).Error; err != nil {
		t.Fatalf("seed stock info: %v", err)
	}

	fetcher := &countingDuoKongKLineFetcher{
		rows: []KLineData{
			{Day: "1", Open: "8", High: "10", Low: "7", Close: "8"},
			{Day: "2", Open: "10", High: "12", Low: "9", Close: "11"},
			{Day: "3", Open: "11", High: "13", Low: "10", Close: "12"},
			{Day: "4", Open: "12", High: "14", Low: "11", Close: "13"},
		},
	}
	api := &DuoKongLineApi{kLineFetcher: fetcher}
	taskID := api.StartScan(DuoKongLineCondition{
		Signal:      "fine_cross_up",
		Days:        90,
		SignalDays:  3,
		MarketBoard: "CHINEXT",
	})

	var task *DuoKongLineScanTask
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		task = api.GetScanTask(taskID)
		if task != nil && task.Status == "done" {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	if task == nil {
		t.Fatal("expected scan task")
	}
	if task.Status != "done" {
		t.Fatalf("expected done task, got %+v", task)
	}
	if task.Total != 1 || task.Progress != 1 {
		t.Fatalf("unexpected progress: %+v", task)
	}
	if len(task.Results) != 1 {
		t.Fatalf("expected one result, got %d", len(task.Results))
	}
	if task.Results[0].Code != "300418.SZ" {
		t.Fatalf("unexpected scan result: %+v", task.Results[0])
	}
}

func TestDuoKongBackgroundScanCanBeCancelled(t *testing.T) {
	resetDuoKongLineTestState()
	dao, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	db.Dao = dao
	if err := db.Dao.AutoMigrate(&models.AllStockInfo{}); err != nil {
		t.Fatalf("migrate all stock info: %v", err)
	}
	if err := db.Dao.Create(&[]models.AllStockInfo{
		{SECUCODE: "300418.SZ", SECURITYCODE: "300418", SECURITYNAMEABBR: "Alpha", MARKET: "SZ"},
		{SECUCODE: "300419.SZ", SECURITYCODE: "300419", SECURITYNAMEABBR: "Beta", MARKET: "SZ"},
	}).Error; err != nil {
		t.Fatalf("seed stock info: %v", err)
	}

	release := make(chan struct{})
	api := &DuoKongLineApi{kLineFetcher: &blockingDuoKongKLineFetcher{
		started: make(chan struct{}, 2),
		release: release,
		rows: []KLineData{
			{Day: "1", Open: "10", High: "10", Low: "10", Close: "10"},
			{Day: "2", Open: "20", High: "20", Low: "20", Close: "20"},
		},
	}}
	taskID := api.StartScan(DuoKongLineCondition{Signal: "fine_cross_up", MarketBoard: "CHINEXT"})
	select {
	case <-api.kLineFetcher.(*blockingDuoKongKLineFetcher).started:
	case <-time.After(time.Second):
		t.Fatal("scan did not begin fetching")
	}
	if !api.CancelScan(taskID) {
		t.Fatal("expected running scan to be cancellable")
	}
	close(release)

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		task := api.GetScanTask(taskID)
		if task != nil && task.Status == "cancelled" {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("expected cancelled task, got %+v", api.GetScanTask(taskID))
}
