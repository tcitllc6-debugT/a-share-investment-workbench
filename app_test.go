package main

import (
	"context"
	"encoding/json"
	"go-stock/backend/data"
	"go-stock/backend/db"
	"go-stock/backend/logger"
	"go-stock/backend/models"
	"go-stock/backend/util"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
)

// @Author spark
// @Date 2025/2/24 9:35
// @Desc
// -----------------------------------------------------------------------------------
func TestIsHKTradingTime(t *testing.T) {
	f := IsHKTradingTime(time.Now())
	t.Log(f)
}

func TestIsUSTradingTime(t *testing.T) {

	date := time.Now()
	hour, minute, _ := date.Clock()
	logger.SugaredLogger.Infof("当前时间: %d:%d", hour, minute)

	t.Log(IsUSTradingTime(time.Now()))
}

func TestCheckStockBaseInfo(t *testing.T) {
	if testing.Short() || os.Getenv("GO_STOCK_INTEGRATION") != "1" {
		t.Skip("requires a live Wails lifecycle context and rewrites the local stock database")
	}
	db.Init("./data/stock.db")
	NewApp().CheckStockBaseInfo(context.Background())
}

func TestJson(t *testing.T) {
	db.Init("./data/stock.db")

	jsonStr := "{\n\t\t\"id\" : 3334,\n\t\t\"created_at\" : \"2025-02-28 16:49:31.8342514+08:00\",\n\t\t\"updated_at\" : \"2025-02-28 16:49:31.8342514+08:00\",\n\t\t\"deleted_at\" : null,\n\t\t\"code\" : \"PUK.US\",\n\t\t\"name\" : \"英国保诚集团\",\n\t\t\"full_name\" : \"\",\n\t\t\"e_name\" : \"\",\n\t\t\"exchange\" : \"NASDAQ\",\n\t\t\"type\" : \"stock\",\n\t\t\"is_del\" : 0,\n\t\t\"bk_name\" : null,\n\t\t\"bk_code\" : null\n\t}"

	v := &models.StockInfoUS{}
	json.Unmarshal([]byte(jsonStr), v)
	logger.SugaredLogger.Infof("v:%+v", v)

	db.Dao.Model(v).Updates(v)

}

func TestUpdateCheck(t *testing.T) {
	releaseVersion := &models.GitHubReleaseVersion{}
	_, err := resty.New().R().
		SetResult(releaseVersion).
		SetHeader("Accept", "application/vnd.github+json").
		SetHeader("X-GitHub-Api-Version", "2022-11-28").
		Get("https://api.github.com/repos/ArvinLovegood/go-stock/releases/latest")
	//  https://api.github.com/repos/OWNER/REPO/releases/latest
	if err != nil {
		logger.SugaredLogger.Errorf("get github release version error:%s", err.Error())
		return
	}
	logger.SugaredLogger.Infof("releaseVersion:%+v", releaseVersion)
}

func TestGetScreenResolution(t *testing.T) {
	x, y, w, h, err := getScreenResolution()
	if err != nil {
		logger.SugaredLogger.Errorf("get screen resolution error:%s", err.Error())
		return
	}
	logger.SugaredLogger.Infof("x:%d,y:%d,w:%d,h:%d", x, y, w, h)

}

func TestCheckUpdate(t *testing.T) {
	db.Init("./data/stock.db")
	NewApp().CheckUpdate(1)
}

func TestCheckDeviceBindingAllowsLocalMode(t *testing.T) {
	result := (&App{}).CheckDeviceBinding("", "")
	if result["bound"] != true {
		t.Fatalf("expected local mode to be treated as bound, got %#v", result)
	}
}

func TestCheckDeviceBindingNeverCallsRemoteServices(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"data":{"bound":false,"deviceCount":99,"maxDevices":1}}`))
	}))
	defer server.Close()

	result := (&App{}).CheckDeviceBinding("legacy-token", server.URL)
	if calls.Load() != 0 {
		t.Fatalf("device binding compatibility method made %d remote request(s)", calls.Load())
	}
	if result["bound"] != true {
		t.Fatalf("local build must always be treated as bound, got %#v", result)
	}
}

func TestLocalAccessCompatibilityMethods(t *testing.T) {
	app := &App{}
	if got := app.GetMachineId(); got != "local" {
		t.Fatalf("GetMachineId() = %q, want local", got)
	}
	result := app.CheckSponsorCode("")
	if result["code"] != 1 {
		t.Fatalf("local access should not require a sponsor code, got %#v", result)
	}
}

func TestGetAiRecommendStocksList(t *testing.T) {
	db.Init("./data/stock.db")

	str := "{\"startDate\": \"2026-03-20 00:00:00\", \"endDate\": \"2026-03-27 23:59:59\", \"page\": 1, \"pageSize\": 5000}"
	query := &models.AiRecommendStocksQuery{}
	json.Unmarshal([]byte(str), query)

	pageData, err := data.NewAiRecommendStocksService().GetAiRecommendStocksList(query)
	logger.SugaredLogger.Infof("pageData:%+v", pageData.List)
	if err != nil {
		pageData = &models.AiRecommendStocksPageData{}
	}
	var dataExport []models.AiRecommendStocksMdExport
	for _, v := range pageData.List {
		dataExport = append(dataExport, v.ToMdExportStruct())
	}
	content := util.MarkdownTableWithTitle("近期AI分析/推荐股票明细列表", dataExport)
	logger.SugaredLogger.Infof("content:%s", content)
}

func TestSummaryStockNews(t *testing.T) {
	db.Init("./data/stock.db")
	question := "分析今日的市场行情走势是否和券商的观点一致"
	app := NewApp()
	msgs := data.NewDeepSeekOpenAi(app.ctx, 0).NewSummaryStockNewsStreamWithTools(question, nil, app.AiTools, true, nil)

	content := &strings.Builder{}
	for msg := range msgs {
		logger.SugaredLogger.Infof("msg:%+v", msg)
		content.WriteString(msg["content"].(string))
	}
	logger.SugaredLogger.Infof("content:%s", content.String())
}

func TestCalculateNextRunTime(t *testing.T) {
	db.Init("./data/stock.db")
	t.Log(NewApp().CalculateNextRunTime("0 0 0 * * ?"))
}

func TestFetchAiModels(t *testing.T) {
	app := NewApp()
	models := app.FetchAiModels("https://ark.cn-beijing.volces.com/api/v3", "")
	t.Log(models)

}
func TestGetLatestTradingDay(t *testing.T) {
	app := NewApp()
	date := app.GetLatestTradingDay()
	t.Log(date)
}
