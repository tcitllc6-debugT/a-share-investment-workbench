//go:build integration

package data

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"
	"time"

	"go-stock/backend/db"
	"go-stock/backend/logger"
	"go-stock/backend/models"
	"go-stock/backend/util"

	"github.com/coocood/freecache"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/random"
	"github.com/glebarez/sqlite"
	"github.com/tidwall/gjson"
	"gorm.io/gorm"
)

// @Author spark
// @Date 2025/4/23 17:58
// @Desc
//-----------------------------------------------------------------------------------

func TestIsOpenableNewsURL(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want bool
	}{
		{name: "https", raw: "https://cn.tradingview.com/news/example/", want: true},
		{name: "http", raw: "http://example.com/news", want: true},
		{name: "relative", raw: "/news/example/", want: false},
		{name: "javascript", raw: "javascript:alert(1)", want: false},
		{name: "empty", raw: "", want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsOpenableNewsURL(tc.raw); got != tc.want {
				t.Fatalf("IsOpenableNewsURL(%q) = %v, want %v", tc.raw, got, tc.want)
			}
		})
	}
}

func TestTradingViewOriginalURL(t *testing.T) {
	withLink := models.TVNews{
		Id:        "tag:reuters.com,2026:newsml_TEST:0",
		Link:      "https://www.gelonghui.com/news/5265434",
		StoryPath: "/news/gelonghui:58a4f6bb9e3f9:0/",
	}
	if got := TradingViewOriginalURL(withLink, nil); got != withLink.Link {
		t.Fatalf("expected direct link %q, got %q", withLink.Link, got)
	}

	withStoryPath := models.TVNews{
		Id:        "gelonghui:58a4f6bb9e3f9:0",
		StoryPath: "/news/gelonghui:58a4f6bb9e3f9:0/",
	}
	wantStoryURL := "https://cn.tradingview.com/news/gelonghui:58a4f6bb9e3f9:0/"
	if got := TradingViewOriginalURL(withStoryPath, nil); got != wantStoryURL {
		t.Fatalf("expected story URL %q, got %q", wantStoryURL, got)
	}

	idOnly := models.TVNews{Id: "tag:reuters.com,2026:newsml_TEST:0"}
	if got := TradingViewOriginalURL(idOnly, nil); got != "" {
		t.Fatalf("expected empty URL for id-only news, got %q", got)
	}

	withBadLink := models.TVNews{
		Link:      "javascript:alert(1)",
		StoryPath: "news/reuters.com,2026:newsml_TEST:0/",
	}
	wantFallback := "https://cn.tradingview.com/news/reuters.com,2026:newsml_TEST:0/"
	if got := TradingViewOriginalURL(withBadLink, nil); got != wantFallback {
		t.Fatalf("expected story fallback %q, got %q", wantFallback, got)
	}
}

func TestGetNews24HoursListUsesPublishedTimeInsteadOfFetchTime(t *testing.T) {
	dao, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	db.Dao = dao
	if err := db.Dao.AutoMigrate(&models.Telegraph{}, &models.TelegraphTags{}, &models.Tags{}); err != nil {
		t.Fatalf("migrate news tables: %v", err)
	}
	now := time.Now()
	oldPublished := now.Add(-25 * time.Hour)
	newPublished := now.Add(-2 * time.Hour)
	if err := db.Dao.Create(&[]models.Telegraph{
		{Title: "old-but-just-fetched", Content: "old", Source: "test", DataTime: &oldPublished},
		{Title: "recent", Content: "recent", Source: "test", DataTime: &newPublished},
	}).Error; err != nil {
		t.Fatalf("seed news: %v", err)
	}

	news := NewMarketNewsApi().GetNews24HoursList("test", 10)
	if news == nil || len(*news) != 1 || (*news)[0].Title != "recent" {
		t.Fatalf("expected only recently published news, got %+v", news)
	}
}

func TestGetTelegraphBriefListReturnsNewestLocalRowsWithoutTagHydration(t *testing.T) {
	dao, err := gorm.Open(sqlite.Open("file:market-news-brief?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	db.Dao = dao
	if err := db.Dao.AutoMigrate(&models.Telegraph{}, &models.TelegraphTags{}, &models.Tags{}); err != nil {
		t.Fatalf("migrate news tables: %v", err)
	}
	now := time.Now()
	for i := 0; i < 12; i++ {
		published := now.Add(time.Duration(i) * time.Minute)
		row := models.Telegraph{Title: "brief-" + strconv.Itoa(i), Content: "content", Source: "财联社电报", DataTime: &published}
		if err := db.Dao.Create(&row).Error; err != nil {
			t.Fatalf("seed news: %v", err)
		}
	}

	news := NewMarketNewsApi().GetTelegraphBriefList("财联社电报", 5)
	if news == nil || len(*news) != 5 {
		t.Fatalf("expected five brief rows, got %+v", news)
	}
	if (*news)[0].Title != "brief-11" {
		t.Fatalf("expected newest row first, got %q", (*news)[0].Title)
	}
	if len((*news)[0].TelegraphTags) != 0 || len((*news)[0].SubjectTags) != 0 {
		t.Fatalf("brief query must not hydrate tags: %+v", (*news)[0])
	}
}

func TestBackfillRecentMarketNewsRunsSourcesInParallelAndRecordsStatus(t *testing.T) {
	dao, err := gorm.Open(sqlite.Open("file:market-news-sync?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	db.Dao = dao
	if err := db.Dao.AutoMigrate(&models.Telegraph{}, &models.TelegraphTags{}, &models.Tags{}, &models.NewsSyncStatus{}); err != nil {
		t.Fatalf("migrate news tables: %v", err)
	}
	originalSources := marketNewsBackfillSources
	defer func() { marketNewsBackfillSources = originalSources }()
	started := make(chan struct{}, 3)
	release := make(chan struct{})
	makeFetcher := func() marketNewsFetcher {
		return func(_ MarketNewsApi, _ int64) ([]models.Telegraph, error) {
			started <- struct{}{}
			<-release
			return []models.Telegraph{}, nil
		}
	}
	marketNewsBackfillSources = []marketNewsBackfillSource{
		{name: "source-a", fetch: makeFetcher()},
		{name: "source-b", fetch: makeFetcher()},
		{name: "source-c", fetch: makeFetcher()},
	}
	go func() {
		<-started
		<-started
		<-started
		close(release)
	}()

	done := make(chan struct{})
	go func() {
		NewMarketNewsApi().BackfillRecentMarketNews(1)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("market news sources did not run concurrently")
	}

	var statuses []models.NewsSyncStatus
	if err := db.Dao.Order("source ASC").Find(&statuses).Error; err != nil {
		t.Fatalf("query source statuses: %v", err)
	}
	if len(statuses) != 3 {
		t.Fatalf("expected status for each source, got %+v", statuses)
	}
	for _, status := range statuses {
		if status.Status != "success" || status.LastSuccessAt == nil || status.LastFinishedAt == nil {
			t.Fatalf("expected successful source status, got %+v", status)
		}
	}
}

func TestGetSinaNews(t *testing.T) {
	db.Init("../../data/stock.db")
	InitAnalyzeSentiment()
	news := NewMarketNewsApi().GetSinaNews(30)
	for i, telegraph := range *news {
		logger.SugaredLogger.Debugf("key: %+v, value: %+v", i, telegraph)

	}
	//NewMarketNewsApi().GetNewTelegraph(30)

}

func TestGlobalStockIndexes(t *testing.T) {
	resp := NewMarketNewsApi().GlobalStockIndexes(30)

	bs, _ := json.Marshal(resp)
	logger.SugaredLogger.Debugf("JSON:\n%s", string(bs))

	md := NewMarketNewsApi().GlobalStockIndexesToReadable(resp)

	logger.SugaredLogger.Debugf("resp: \n%s", md)
}

func TestGetIndustryRank(t *testing.T) {
	res := NewMarketNewsApi().GetIndustryRank("0", 10)
	for s, a := range res["data"].([]any) {
		logger.SugaredLogger.Debugf("key: %+v, value: %+v", s, a)
	}
}
func TestGetIndustryMoneyRankSina(t *testing.T) {
	res := NewMarketNewsApi().GetIndustryMoneyRankSina("0", "netamount")
	for i, re := range res {
		logger.SugaredLogger.Debugf("key: %+v, value: %+v", i, re)

	}
}
func TestGetMoneyRankSina(t *testing.T) {
	res := NewMarketNewsApi().GetMoneyRankSina("r3_net")
	for i, re := range res {
		logger.SugaredLogger.Debugf("key: %+v, value: %+v", i, re)
	}
}

func TestGetStockMoneyTrendByDay(t *testing.T) {
	res := NewMarketNewsApi().GetStockMoneyTrendByDay("sh600438", 360)
	for i, re := range res {
		logger.SugaredLogger.Debugf("key: %+v, value: %+v", i, re)
	}
}

func TestLongTiger(t *testing.T) {
	db.Init("../../data/stock.db")

	NewMarketNewsApi().LongTiger("2025-06-08")
}

func TestStockResearchReport(t *testing.T) {
	db.Init("../../data/stock.db")
	resp := NewMarketNewsApi().StockResearchReport("002046", 7)
	for _, a := range resp {
		logger.SugaredLogger.Debugf("value: %+v", a)
		data := a.(map[string]any)
		logger.SugaredLogger.Debugf("value: %s  infoCode:%s", data["title"], data["infoCode"])
		NewMarketNewsApi().GetIndustryReportInfo(data["infoCode"].(string))
	}
}

func TestIndustryResearchReport(t *testing.T) {
	db.Init("../../data/stock.db")
	resp := NewMarketNewsApi().IndustryResearchReport("", 7)
	for _, a := range resp {
		logger.SugaredLogger.Debugf("value: %+v", a)
		data := a.(map[string]any)
		logger.SugaredLogger.Debugf("value: %s  infoCode:%s", data["title"], data["infoCode"])
		logger.SugaredLogger.Debugf("url: https://pdf.dfcfw.com/pdf/H3_%s_1.pdf", data["infoCode"])
		//NewMarketNewsApi().GetIndustryReportInfo(data["infoCode"].(string))
	}
}

func TestStockNotice(t *testing.T) {
	db.Init("../../data/stock.db")
	resp := NewMarketNewsApi().StockNotice("600584,600900")
	// 转为可表格化的结构：东方财富接口返回的 list 中每项为 map，字段名多为驼峰
	type row struct {
		Title      string `md:"公告标题"`
		NoticeDate string `md:"公告日期"`
		ColumnName string `md:"公告类型"`
	}
	var rows []row
	for _, a := range resp {
		m, ok := a.(map[string]any)
		if !ok {
			continue
		}
		if m["columns"].([]any) != nil && len(m["columns"].([]any)) > 0 {
			columns := m["columns"].([]any)[0].(map[string]any)
			rows = append(rows, row{
				Title:      convertor.ToString(m["title"]),
				NoticeDate: convertor.ToString(m["notice_date"]),
				ColumnName: convertor.ToString(columns["column_name"]),
			})
		} else {
			rows = append(rows, row{
				Title:      convertor.ToString(m["title"]),
				NoticeDate: convertor.ToString(m["notice_date"]),
			})
		}

	}
	md := util.MarkdownTableWithTitle("上市公司公告", rows)
	logger.SugaredLogger.Infof(md)
}

func TestEMDictCode(t *testing.T) {
	db.Init("../../data/stock.db")
	resp := NewMarketNewsApi().EMDictCode("016", freecache.NewCache(100))
	for _, a := range resp {
		logger.SugaredLogger.Debugf("value: %+v", a)
	}
	bytes, err := json.Marshal(resp)
	if err != nil {
		return
	}
	dict := &[]models.BKDict{}
	json.Unmarshal(bytes, dict)
	logger.SugaredLogger.Debugf("value: %s", string(bytes))
	md := util.MarkdownTableWithTitle("行业/板块代码", dict)
	logger.SugaredLogger.Debugf(md)

}

func TestTradingViewNews(t *testing.T) {
	db.Init("../../data/stock.db")
	InitAnalyzeSentiment()
	NewMarketNewsApi().TradingViewNews()
}

func TestXUEQIUHotStock(t *testing.T) {
	db.Init("../../data/stock.db")
	res := NewMarketNewsApi().XUEQIUHotStock(50, "10")
	for _, a := range *res {
		logger.SugaredLogger.Debugf("value: %+v", a)
	}

	md := util.MarkdownTableWithTitle("当前热门股票排名", res)
	logger.SugaredLogger.Debugf(md)
}

func TestHotEvent(t *testing.T) {
	db.Init("../../data/stock.db")
	res := NewMarketNewsApi().HotEvent(50)
	for _, a := range *res {
		logger.SugaredLogger.Debugf("value: %+v", a)
	}

}

func TestHotTopic(t *testing.T) {
	db.Init("../../data/stock.db")
	res := NewMarketNewsApi().HotTopic(10)
	for _, a := range res {
		logger.SugaredLogger.Debugf("value: %+v", a)
	}

}

func TestInvestCalendar(t *testing.T) {
	db.Init("../../data/stock.db")
	res := NewMarketNewsApi().InvestCalendar("2025-06")
	for _, a := range res {
		bytes, err := json.Marshal(a)
		if err != nil {
			continue
		}
		date := gjson.Get(string(bytes), "date")
		list := gjson.Get(string(bytes), "list")

		logger.SugaredLogger.Debugf("value: %+v,list: %+v", date.String(), list)
	}
}

func TestClsCalendar(t *testing.T) {
	db.Init("../../data/stock.db")
	res := NewMarketNewsApi().ClsCalendar()
	md := strings.Builder{}
	for _, a := range res {
		bytes, err := json.Marshal(a)
		if err != nil {
			continue
		}
		//logger.SugaredLogger.Debugf("value: %+v", string(bytes))
		date := gjson.Get(string(bytes), "calendar_day")
		md.WriteString("\n### 事件/会议日期：" + date.String())
		list := gjson.Get(string(bytes), "items")
		//logger.SugaredLogger.Debugf("value: %+v,list: %+v", date.String(), list)
		list.ForEach(func(key, value gjson.Result) bool {
			logger.SugaredLogger.Debugf("key: %+v,value: %+v", key.String(), gjson.Get(value.String(), "title"))
			md.WriteString("\n- " + gjson.Get(value.String(), "title").String())
			return true
		})
	}
	logger.SugaredLogger.Debugf("md:\n %s", md.String())
}

func TestGetGDP(t *testing.T) {
	res := NewMarketNewsApi().GetGDP()
	md := util.MarkdownTableWithTitle("国内生产总值(GDP)", res.GDPResult.Data)
	logger.SugaredLogger.Debugf(md)
}
func TestGetCPI(t *testing.T) {
	res := NewMarketNewsApi().GetCPI()
	md := util.MarkdownTableWithTitle("居民消费价格指数(CPI)", res.CPIResult.Data)
	logger.SugaredLogger.Debugf(md)
}

// PPI
func TestGetPPI(t *testing.T) {
	res := NewMarketNewsApi().GetPPI()
	md := util.MarkdownTableWithTitle("工业品出厂价格指数(PPI)", res.PPIResult.Data)
	logger.SugaredLogger.Debugf(md)
}

// PMI
func TestGetPMI(t *testing.T) {
	res := NewMarketNewsApi().GetPMI()
	md := util.MarkdownTableWithTitle("采购经理人指数(PMI)", res.PMIResult.Data)
	logger.SugaredLogger.Debugf(md)
}
func TestGetIndustryReportInfo(t *testing.T) {
	NewMarketNewsApi().GetIndustryReportInfo("AP202507151709216483")
}

func TestReutersNew(t *testing.T) {
	db.Init("../../data/stock.db")
	NewMarketNewsApi().ReutersNew()
}

func TestInteractiveAnswer(t *testing.T) {
	db.Init("../../data/stock.db")
	datas := NewMarketNewsApi().InteractiveAnswer(1, 100, "立讯精密")
	logger.SugaredLogger.Debugf("PageSize:%d", datas.PageSize)
	md := util.MarkdownTableWithTitle("投资互动", datas.Results)
	logger.SugaredLogger.Debugf(md)

}
func TestGetNewsList2(t *testing.T) {
	db.Init("../../data/stock.db")
	news := NewMarketNewsApi().GetNewsList2("财联社电报", random.RandInt(100, 500))
	messageText := strings.Builder{}
	for _, telegraph := range *news {
		messageText.WriteString("## " + telegraph.Time + ":" + "\n")
		messageText.WriteString("### " + telegraph.Content + "\n")
	}
	logger.SugaredLogger.Debugf("value: %s", messageText.String())
}

func TestTelegraphList(t *testing.T) {
	db.Init("../../data/stock.db")
	InitAnalyzeSentiment()
	NewMarketNewsApi().TelegraphList(30)
}

func TestGetSecuritiesCompanyOpinion(t *testing.T) {
	res := NewMarketNewsApi().GetSecuritiesCompanyOpinion("2026-03-01", "2026-03-03")
	md := strings.Builder{}
	for _, d := range res.Data {
		md.WriteString(d.OpinionData + "\n")
	}
	logger.SugaredLogger.Debugf("%s", md.String())

}
func TestGetNewsListData(t *testing.T) {
	db.Init("../../data/stock.db")
	list, total := NewMarketNewsApi().GetNewsListData("", time.Now().Add(-time.Hour*24*2), 1, 2)
	md := strings.Builder{}
	md.WriteString("### " + "最近新闻资讯" + "（共 " + strconv.FormatInt(total, 10) + " 条）\r\n")
	for _, d := range *list {
		md.WriteString(d.DataTime.Format(time.DateTime) + " " + d.Content + "\r\n")
	}
	//logger.SugaredLogger.Infof("%s", md.String())
}
