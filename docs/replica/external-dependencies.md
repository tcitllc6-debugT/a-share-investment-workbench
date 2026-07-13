# External Dependency Inventory

This inventory separates original-author services from public data sources and user-configured services. The replica must not require original-author services for normal use.

## Original-Author Hosted Services

| Feature | File | Endpoint | Current Behavior | Replica Decision |
| --- | --- | --- | --- | --- |
| Release check | `app.go` | `https://api.github.com/repos/ArvinLovegood/go-stock/releases/latest` | Checks original upstream release | Replace with disabled-by-default update provider or replica-owned GitHub repo |
| Release list | `app.go` | `https://api.github.com/repos/ArvinLovegood/go-stock/releases` | Lists original upstream releases | Replace with replica-owned update provider |
| Release tag lookup | `app.go` | `https://api.github.com/repos/ArvinLovegood/go-stock/git/ref/tags/*` | Looks up original release tag commit | Remove from default replica path |
| Release download | `app.go` | `https://github.com/ArvinLovegood/go-stock/releases/download/*` | Downloads original upstream binary | Replace with replica-owned release URL |
| Release mirror | `app.go` | `https://gh.927223.xyz/https://github.com/ArvinLovegood/go-stock/*` | Uses third-party mirror for original binary | Remove from default replica path |
| News sync | `app.go` | `http://go-stock.sparkmemory.top:16666/FinancialNews/json` | Pulls synced financial news | Disable or replace with public news provider |
| Share upload | `app.go`, `ai-assistant-web/server.go` | `http://go-stock.sparkmemory.top:16688/upload` | Uploads analysis text | Replace with local export provider |
| Prompt plaza | `app.go`, `frontend/src/components/prompt*.vue`, `frontend/src/components/settings.vue` | `http://go-stock.sparkmemory.top:1918/api` | Remote prompt plaza and account APIs | Local prompt template library by default; optional user-configured remote |
| Embedded community page | `frontend/src/components/stockhotmap.vue` | `https://go-stock.sparkmemory.top:16667/go-stock` | Embeds original hosted page | Remove or make optional user-configured URL |
| Stock basic mirror | `app.go` | `http://8.134.249.145:18080/go-stock/stock_basic.json` | Downloads A-share stock basics | Prefer bundled JSON; later allow user-configured mirror |
| HK stock basic mirror | `app.go` | `http://8.134.249.145:18080/go-stock/stock_base_info_hk.json` | Downloads HK stock basics | Prefer bundled JSON; later allow user-configured mirror |
| US stock basic mirror | `app.go` | `http://8.134.249.145:18080/go-stock/stock_base_info_us.json` | Downloads US stock basics | Prefer bundled JSON; later allow user-configured mirror |
| Static original assets | `frontend/src/components/about.vue`, `frontend/src/components/market.vue`, `frontend/src/components/stock.vue`, `frontend/src/components/agent-chat.vue` | `https://raw.githubusercontent.com/ArvinLovegood/go-stock/*`, `https://github.com/ArvinLovegood/go-stock/*` | Loads original icons, screenshots, payment assets, links | Replace with local replica assets and attribution links |

## Public Data Sources

| Feature | File | Endpoint | Key Required | Failure Fallback |
| --- | --- | --- | --- | --- |
| Sina quote | `backend/data/stock_data_api.go`, `backend/data/fund_data_api.go`, `backend/data/market_news_api.go` | `http://hq.sinajs.cn`, `https://finance.sina.com.cn`, `https://vip.stock.finance.sina.com.cn` | No | Return empty quote/news result and log source error |
| Tencent quote/rank | `backend/data/stock_data_api.go`, `backend/data/market_news_api.go` | `http://qt.gtimg.cn`, `https://proxy.finance.qq.com` | No | Return empty quote/rank result and log source error |
| Tushare | `backend/data/stock_data_api.go`, `backend/data/tushare_data_api.go` | `http://api.tushare.pro` | Yes | Use bundled stock basics; show configuration hint for refresh |
| Eastmoney quote/K-line/F10 | `app.go`, `backend/data/eastmoney_*`, `backend/data/market_news_api.go`, `backend/data/fund_*` | `eastmoney.com`, `dfcfw.com`, `1234567.com.cn` APIs | Sometimes | Fallback to alternate K-line provider or empty state |
| Cailianpress | `app.go`, `backend/data/market_news_api.go` | `https://www.cls.cn`, `https://x-quote.cls.cn` | No | Empty news/market list with log |
| Xueqiu hot data | `backend/data/market_news_api.go`, `backend/data/xueqiu_chromedp.go` | `https://xueqiu.com`, `https://stock.xueqiu.com` | Cookie may be required | Cookie fetch fallback; empty hot list if blocked |
| Timor holiday | `app.go`, `backend/agent/tools/holiday_tool.go` | `https://timor.tech/api/holiday/*` | No | Weekend-only trading-day fallback where possible |
| US market day status | `app.go` | `https://fincalapi.com/v1/day_status` | No | Conservative non-trading or unknown state |
| TradingView news | `backend/data/market_news_api.go` | `https://news-mediator.tradingview.com`, `https://news-headlines.tradingview.com` | No | Empty global news result |
| Wallstreetcn | `backend/data/wallstreetcn_api.go` | Wallstreetcn market/news APIs | No | Empty market/news result |
| Iwencai | `backend/data/iwencai_api.go` | `https://openapi.iwencai.com`, `https://www.iwencai.com` | API key for open API; gateway may not | Configuration hint or empty result |
| Fund data | `backend/data/fund_data_api.go`, `backend/data/fund_kline_api.go` | Eastmoney fund APIs, Sina fund quotes | No | Empty fund ranking/K-line/net-value result |
| Browser crawler pages | `backend/data/crawler_api.go`, `backend/data/web_search_api.go` | Finance/news pages opened through browser | Browser path may be required | Return crawler error and configuration hint |

## User-Configured Services

| Feature | Config Field | Required For | Replica Policy |
| --- | --- | --- | --- |
| AI models | `ai_config.base_url`, `ai_config.api_key`, `ai_config.model_name` | AI analysis, agent chat, AI assistant web | Keep; missing config must show a setup hint |
| Tushare | `settings.tushare_token` | Stock/index basic refresh | Keep optional; bundled data remains default |
| DingTalk | `settings.ding_robot`, `settings.ding_push_enable` | DingTalk notifications | Keep optional |
| Feishu webhook | `settings.feishu_robot`, `settings.feishu_secret` | Feishu notifications | Keep optional |
| Feishu app bot | `settings.feishu_app_id`, `settings.feishu_app_secret`, bot settings | Feishu AI bot | Keep optional |
| HTTP proxy | `settings.http_proxy`, `settings.http_proxy_enabled` | Crawlers and some model calls | Keep optional |
| Browser path | `settings.browser_path` | Chromedp crawler and web search | Keep optional with auto-detect |
| Iwencai API | `settings.iwencai_api_key` | Open Iwencai query APIs | Keep optional |
| Eastmoney AI/API key | `settings.em_api_key` | Eastmoney AI/F10 flows where needed | Keep optional |
| Prompt plaza API base | `settings.prompt_plaza_api_base` | Remote prompt plaza if user owns one | Blank means local-only mode |
| MCP servers | `mcp_servers` table | Agent external tools | Keep optional and user-controlled |

## Static External Links To Rebrand

| Area | File | Current Link Type | Replica Decision |
| --- | --- | --- | --- |
| About page | `frontend/src/components/about.vue` | Original GitHub, release, issue, donation, avatar, sponsor links | Replace with replica project links, GPL attribution, and risk disclaimer |
| Agent/market/stock icons | `frontend/src/components/agent-chat.vue`, `frontend/src/components/market.vue`, `frontend/src/components/stock.vue` | Original raw GitHub app icon | Replace with local bundled asset |
| Embedded market pages | `frontend/src/components/stockhotmap.vue` | Third-party market websites | Keep only where they add value and are clearly external |
| Research report PDFs | `frontend/src/components/IndustryResearchReportList.vue` | Eastmoney PDF link | Keep as external-open action with failure tolerance |

## Replacement Priority

1. Share upload to local export provider.
2. Original update channel to disabled-by-default update provider.
3. Prompt plaza default to local-only mode.
4. VIP/device gates for local-only features.
5. Original branding/assets/links.
6. Stock basic mirror to bundled-data-first behavior.
7. Public data-source provider boundaries.
