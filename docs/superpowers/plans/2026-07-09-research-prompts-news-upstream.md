# Research Prompts News Upstream Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Clarify the research-center prompt system, make the A-share replica easier to distribute, fix market-news original-link 404s, improve startup 24-hour news sync, and safely absorb the latest upstream go-stock changes.

**Architecture:** Keep the replica local-first and isolated from the upstream working copy. Implement prompt seeds in the existing SQLite prompt-template model, fix market-news URLs at the provider boundary, add startup backfill through existing public news providers, and document platform distribution constraints in replica docs.

**Tech Stack:** Go, Wails, Vue 3, Naive UI, SQLite/GORM, public market-news HTTP providers.

## Global Constraints

- Work only in `C:/Users/lxd/Desktop/A-Share-Desk-Replica-20260709/source`.
- Do not modify `C:/Users/lxd/Desktop/gupiao/github-a-share-projects/ArvinLovegood__go-stock`.
- Keep the app as a desktop Wails app, not a web-port service.
- Preserve local-first behavior: no default dependency on original-author VIP/community services.
- Use public provider URLs only when the generated original article link is known to be openable.
- Verify with tests/build checks before claiming completion.

---

### Task 1: Document Research Center And Prompt Semantics

**Files:**
- Modify: `docs/replica/research-center-and-distribution.md`

**Interfaces:**
- Consumes: existing pages `researchReport.vue`, `aiRecommendStocksList.vue`, `promptTemplateList.vue`, `promptPlaza.vue`, `promptQa.vue`.
- Produces: a concise product note explaining AI analysis reports, stock recommendation records, local prompt templates, prompt plaza, and Q&A plaza.

- [ ] **Step 1: Create the document section**

Add a section explaining:

```markdown
## Research Center

- AI analysis reports are persisted model outputs from stock or market-analysis flows.
- Stock recommendation records are structured records created when the AI calls the recommendation tool.
- Prompt templates are local reusable system/user prompts stored in SQLite.
- Prompt plaza is an optional remote community prompt marketplace, enabled only when `promptPlazaApiBase` is configured.
- Q&A plaza is an optional remote support/discussion area tied to the same prompt plaza account/API.
```

- [ ] **Step 2: Add distribution section**

Add a section explaining:

```markdown
## Distribution

End users should receive a packaged Windows executable or macOS build. Development builds require Go, Node.js/npm, Wails CLI, and platform WebView dependencies.
```

- [ ] **Step 3: Verify wording matches code**

Run:

```powershell
rg -n "promptPlazaApiBase|GetPromptTemplateList|GetAIResponseResultList|GetAiRecommendStocksList" frontend/src/components app.go
```

Expected: all documented concepts have a code reference.

### Task 2: Seed Local A-Share Prompt Templates

**Files:**
- Modify: `backend/data/prompt_template_api.go`
- Modify: `main.go`
- Test: `backend/data/prompt_template_api_test.go`

**Interfaces:**
- Produces: `DefaultPromptTemplates() []models.PromptTemplate`
- Produces: `SeedDefaultPromptTemplates() error`
- Consumes: `models.PromptTemplate` and global `db.Dao`.

- [ ] **Step 1: Add tests for built-in prompt metadata**

Add tests that assert default prompts are non-empty, names are unique by name/type, include both `模型系统Prompt` and `模型用户Prompt`, and contain A-share-focused wording.

- [ ] **Step 2: Add default prompt definitions**

Add at least five built-in templates:

```text
A股投研系统角色
A股盘前/盘后复盘
个股多因子分析
热点题材与资金流追踪
AI推荐记录结构化输出
风险控制与仓位纪律
```

- [ ] **Step 3: Add idempotent seed function**

Implement `SeedDefaultPromptTemplates()` with `FirstOrCreate` on `name` and `type`, so existing user templates are not overwritten.

- [ ] **Step 4: Call seed after AutoMigrate**

Call the seed function in `AutoMigrate()` after `PromptTemplate` migration.

- [ ] **Step 5: Run prompt tests**

Run:

```powershell
go test ./backend/data -run TestDefaultPromptTemplates -count=1
```

Expected: pass.

### Task 3: Fix Market News Original Links

**Files:**
- Modify: `backend/models/models.go`
- Modify: `backend/data/market_news_api.go`
- Modify: `frontend/src/components/newsList.vue`
- Test: `backend/data/market_news_api_test.go`

**Interfaces:**
- Produces: `TradingViewOriginalURL(item models.TVNews, detail *models.TVNewsDetail) string`
- Produces: `IsOpenableNewsURL(raw string) bool`

- [ ] **Step 1: Add TVNews link field**

Add:

```go
Link string `json:"link"`
```

to `models.TVNews`.

- [ ] **Step 2: Add URL builder tests**

Test these cases:

```go
link => link
storyPath "/news/gelonghui:abc:0/" => "https://cn.tradingview.com/news/gelonghui:abc:0/"
id-only item => ""
invalid javascript URL => ""
```

- [ ] **Step 3: Use URL builder in TradingViewNews**

Replace `fmt.Sprintf("https://cn.tradingview.com/news/%s", a.Id)` with the URL builder.

- [ ] **Step 4: Add front-end URL guard**

In `newsList.vue`, render "查看原文" only for `http://` or `https://` URLs.

- [ ] **Step 5: Run market-news URL tests**

Run:

```powershell
go test ./backend/data -run "TestTradingViewOriginalURL|TestIsOpenableNewsURL" -count=1
```

Expected: pass.

### Task 4: Improve Startup 24-Hour News Sync

**Files:**
- Modify: `backend/data/market_news_api.go`
- Modify: `app.go`
- Test: `backend/data/market_news_api_test.go`

**Interfaces:**
- Produces: `BackfillRecentMarketNews(crawlTimeOut int64) int`
- Consumes: existing `TelegraphList`, `GetSinaNews`, `TradingViewNews`, and `GetNews24HoursList`.

- [ ] **Step 1: Add public-provider backfill helper**

Create a helper that calls the three public providers once and returns the count of records in the last 24 hours after the fetch.

- [ ] **Step 2: Call helper at startup**

In `domReady`, call the helper before registering recurring news cron jobs.

- [ ] **Step 3: Keep disabled remote ntfy path explicit**

Do not restore the original `syncNews()` remote endpoint. Keep the local-first boundary documented.

- [ ] **Step 4: Verify startup path is present**

Run:

```powershell
rg -n "BackfillRecentMarketNews|TelegraphList\\(30\\)|TradingViewNews\\(\\)" app.go backend/data/market_news_api.go
```

Expected: helper and startup call are visible.

### Task 5: Check And Absorb Latest Upstream

**Files:**
- Modify: `docs/replica/upstream-sync-2026-07-09.md`
- Modify code only for safe, relevant upstream deltas that do not undo replica changes.

**Interfaces:**
- Consumes: GitHub `ArvinLovegood/go-stock` latest release/tag/commit information.
- Produces: a local sync note documenting latest upstream commit/tag and ported/skipped items.

- [ ] **Step 1: Clone upstream into temp**

Run:

```powershell
git clone --filter=blob:none https://github.com/ArvinLovegood/go-stock.git C:/tmp/go-stock-upstream-20260709
```

- [ ] **Step 2: Identify latest release tag**

Run:

```powershell
git -C C:/tmp/go-stock-upstream-20260709 tag --sort=-creatordate
git -C C:/tmp/go-stock-upstream-20260709 log --oneline --decorate -20
```

- [ ] **Step 3: Inspect today delta**

Run:

```powershell
git -C C:/tmp/go-stock-upstream-20260709 show --stat --name-status v2026.07.09.1-release
```

- [ ] **Step 4: Port safe deltas**

Only port code that is relevant and compatible with the local-first replica. Do not reintroduce original update/VIP/device-binding service dependencies.

- [ ] **Step 5: Record sync decision**

Document exact tag/commit inspected, items ported, and items intentionally skipped.

### Task 6: Build Verification And Release Refresh

**Files:**
- Modify: release artifact under `C:/Users/lxd/Desktop/A-Share-Desk-Replica-20260709/release` if build succeeds.

**Interfaces:**
- Consumes: Wails build pipeline.
- Produces: updated release executable for Windows if build succeeds.

- [ ] **Step 1: Run focused tests**

Run:

```powershell
go test ./backend/data -run "TestDefaultPromptTemplates|TestTradingViewOriginalURL|TestIsOpenableNewsURL" -count=1
```

- [ ] **Step 2: Run broader compile check**

Run:

```powershell
go test ./... -run "^$" -count=0
```

- [ ] **Step 3: Build Windows package**

Run:

```powershell
wails build
```

- [ ] **Step 4: Copy successful build artifact**

Copy `build/bin/a-share-desk.exe` to `C:/Users/lxd/Desktop/A-Share-Desk-Replica-20260709/release/a-share-desk.exe`.

- [ ] **Step 5: Commit scoped changes**

Run:

```powershell
git status --short
git add <changed files>
git commit -m "feat: improve research prompts and market news sync"
```

## Self-Review

- Spec coverage: all five user items map to Tasks 1 through 5, with verification in Task 6.
- Placeholder scan: no task is left without concrete files and commands.
- Type consistency: new helper names are defined before being consumed.
