package data

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go-stock/backend/db"
	"go-stock/backend/logger"

	"github.com/PuerkitoBio/goquery"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/strutil"
)

func checkIsIndexBasic(stock string) bool {
	count := int64(0)
	db.Dao.Model(&IndexBasic{}).Where("name =  ?", stock).Count(&count)
	return count > 0
}

func GetFinancialReportsByXUEQIU(stockCode string, crawlTimeOut int64) *[]string {
	if strutil.HasPrefixAny(stockCode, []string{"HK", "hk"}) {
		stockCode = strings.ReplaceAll(stockCode, "hk", "")
		stockCode = strings.ReplaceAll(stockCode, "HK", "")
	}
	if strutil.HasPrefixAny(stockCode, []string{"us", "gb_"}) {
		stockCode = strings.ReplaceAll(stockCode, "us", "")
		stockCode = strings.ReplaceAll(stockCode, "gb_", "")
	}
	url := fmt.Sprintf("https://xueqiu.com/snowman/S/%s/detail#/ZYCWZB", stockCode)
	waitVisible := "div.tab-table-responsive table"
	crawlerAPI := CrawlerApi{}
	crawlerBaseInfo := CrawlerBaseInfo{
		Name:        "TestCrawler",
		Description: "Test Crawler Description",
		BaseUrl:     "https://xueqiu.com",
		Headers:     map[string]string{"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36 Edg/133.0.0.0"},
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(crawlTimeOut)*time.Second)
	defer cancel()
	crawlerAPI = crawlerAPI.NewCrawler(ctx, crawlerBaseInfo)

	var markdown strings.Builder
	markdown.WriteString("\n## 财务数据：\n")
	html, ok := crawlerAPI.GetHtml(url, waitVisible, true)
	if !ok {
		return &[]string{""}
	}
	document, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		logger.SugaredLogger.Error(err.Error())
	}
	GetTableMarkdown(document, waitVisible, &markdown)
	return &[]string{markdown.String()}
}

func GetFinancialReports(stockCode string, crawlTimeOut int64) *[]string {
	url := "https://emweb.securities.eastmoney.com/pc_hsf10/pages/index.html?type=web&code=" + stockCode + "#/cwfx"
	waitVisible := "div.report_table table"
	if strutil.HasPrefixAny(stockCode, []string{"HK", "hk"}) {
		stockCode = strings.ReplaceAll(stockCode, "hk", "")
		stockCode = strings.ReplaceAll(stockCode, "HK", "")
		url = "https://emweb.securities.eastmoney.com/PC_HKF10/pages/home/index.html?code=" + stockCode + "&type=web&color=w#/NewFinancialAnalysis"
		waitVisible = "div table.commonTable"
	}
	if strutil.HasPrefixAny(stockCode, []string{"us", "gb_"}) {
		stockCode = strings.ReplaceAll(stockCode, "us", "")
		stockCode = strings.ReplaceAll(stockCode, "gb_", "")
		url = "https://emweb.securities.eastmoney.com/pc_usf10/pages/index.html?type=web&code=" + stockCode + "#/cwfx"
		waitVisible = "div.zyzb_table_detail table"

	}

	//logger.SugaredLogger.Infof("GetFinancialReports搜索股票-%s: %s", stockCode, url)

	crawlerAPI := CrawlerApi{}
	crawlerBaseInfo := CrawlerBaseInfo{
		Name:        "TestCrawler",
		Description: "Test Crawler Description",
		BaseUrl:     "https://emweb.securities.eastmoney.com",
		Headers:     map[string]string{"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36 Edg/133.0.0.0"},
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(crawlTimeOut)*time.Second)
	defer cancel()
	crawlerAPI = crawlerAPI.NewCrawler(ctx, crawlerBaseInfo)

	var markdown strings.Builder
	markdown.WriteString("\n## 财务数据：\n")
	html, ok := crawlerAPI.GetHtml(url, waitVisible, true)
	if !ok {
		return &[]string{""}
	}
	document, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		logger.SugaredLogger.Error(err.Error())
	}
	GetTableMarkdown(document, waitVisible, &markdown)
	return &[]string{markdown.String()}
}

func GetTelegraphList(crawlTimeOut int64) *[]string {
	clsURL := "https://www.cls.cn/api/cache?app=CailianpressWeb&name=telegraph&os=web&sv=8.7.9"
	response, err := SharedHTTPClient.SetTimeout(time.Duration(crawlTimeOut)*time.Second).R().
		SetHeader("Referer", "https://www.cls.cn/").
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0").
		Get(clsURL)
	if err != nil {
		return &[]string{}
	}
	res := map[string]any{}
	if err := json.Unmarshal(response.Body(), &res); err != nil {
		return &[]string{}
	}
	var telegraph []string
	if v, _ := convertor.ToInt(res["errno"]); v == 0 {
		if res["data"] == nil {
			return &[]string{}
		}
		dataMap, ok := res["data"].(map[string]any)
		if !ok {
			return &[]string{}
		}
		rollData, ok := dataMap["roll_data"].([]any)
		if !ok {
			return &[]string{}
		}
		for _, v := range rollData {
			news, ok := v.(map[string]any)
			if !ok {
				continue
			}
			content, _ := news["content"].(string)
			if content != "" {
				telegraph = append(telegraph, ReplaceSensitiveWords(content))
			}
		}
	}
	return &telegraph
}

func GetTopNewsList(crawlTimeOut int64) *[]string {
	url := "https://www.cls.cn"
	response, err := SharedHTTPClient.SetTimeout(time.Duration(crawlTimeOut)*time.Second).R().
		SetHeader("Referer", "https://www.cls.cn/").
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36 Edg/117.0.2045.60").
		Get(url)
	if err != nil {
		return &[]string{}
	}
	//logger.SugaredLogger.Info(string(response.Body()))
	document, err := goquery.NewDocumentFromReader(strings.NewReader(string(response.Body())))
	if err != nil {
		return &[]string{}
	}
	var telegraph []string
	document.Find("div.home-article-title a,div.home-article-rec a").Each(func(i int, selection *goquery.Selection) {
		//logger.SugaredLogger.Info(selection.Text())
		telegraph = append(telegraph, ReplaceSensitiveWords(selection.Text()))
	})
	return &telegraph
}
