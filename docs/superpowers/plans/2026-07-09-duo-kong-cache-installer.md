# Duo Kong Cache And Installer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make Duo Kong Line screening scan all A-share stocks for signals appearing in the most recent 3 trading days, add market-board filters and multi-period viewing, cache K-line data with background batch scanning, expose structured holding summaries to AI, and produce a Windows NSIS installer.

**Architecture:** Keep the existing Duo Kong calculation as the source of truth. Add a small backend scan task layer with in-memory K-line cache and progress state, then poll it from the Vue selector. Add a focused holding-summary service/tool so prompts can consume structured holdings and trading logs.

**Tech Stack:** Go/Wails v2 backend, GORM models, Vue 3 + Naive UI frontend, existing Wails-generated bindings, NSIS through `wails build -nsis`.

## Global Constraints

- Work only in `C:\Users\lxd\Desktop\A-Share-Desk-Replica-20260709\source`.
- Do not touch the upstream project clone.
- Use TDD for behavior changes.
- Default Duo Kong signal window is the most recent 3 trading days.
- Scan all matching stocks; result count is not a scan limit.
- Use existing UI patterns from `SelectStock.vue` and `DuoKongLineSelector.vue`.
- Windows distribution target is a user-friendly NSIS installer.

---

### Task 1: Duo Kong Backend Semantics

**Files:**
- Modify: `backend/data/duo_kong_line_api.go`
- Modify: `backend/data/duo_kong_line_api_test.go`

**Interfaces:**
- Consumes: `DuoKongLineCondition`
- Produces: `SignalDays int`, `MarketBoard string`, full-market selection with no hard result limit unless explicitly set

- [ ] **Step 1: Write failing tests**

Add tests covering:

```go
func TestDuoKongSignalWindowMatchesRecentThreeTradingDays(t *testing.T) {
	points := []DuoKongLinePoint{
		{Day: "1", Fine: 8, Thick: 10},
		{Day: "2", Fine: 11, Thick: 10},
		{Day: "3", Fine: 12, Thick: 11},
		{Day: "4", Fine: 12.5, Thick: 11.5},
	}
	if !duoKongMatchesSignal(points, "fine_cross_up", 3, 0) {
		t.Fatal("expected cross from recent 3 trading days to match")
	}
	if duoKongMatchesSignal(points, "fine_cross_up", 1, 0) {
		t.Fatal("did not expect cross outside most recent 1 trading day to match")
	}
}
```

Also test `duoKongMatchesMarketBoard` for `STAR`, `CHINEXT`, `MAIN`, and `ALL`.

- [ ] **Step 2: Run tests red**

Run:

```powershell
go test ./backend/data -run "TestDuoKongSignalWindow|TestDuoKongMarketBoard" -count=1
```

Expected: fail because market-board matcher and `SignalDays` semantics are missing.

- [ ] **Step 3: Implement minimal backend semantics**

Add fields:

```go
SignalDays int `json:"signalDays"`
MarketBoard string `json:"marketBoard"`
```

Use `SignalDays` for `duoKongMatchesSignal`, keep K-line lookback as `Days` with default `90`, and make `Limit <= 0` mean no limit.

- [ ] **Step 4: Run tests green**

Run the same focused Go test command and expect pass.

### Task 2: K-line Cache And Background Scan

**Files:**
- Modify: `backend/data/duo_kong_line_api.go`
- Test: `backend/data/duo_kong_line_cache_test.go`
- Modify: `app_common.go`

**Interfaces:**
- Produces: `StartDuoKongLineScan(condition data.DuoKongLineCondition) string`
- Produces: `GetDuoKongLineScanTask(taskID string) *data.DuoKongLineScanTask`
- Produces: cached K-line provider used by `GetSelections`

- [ ] **Step 1: Write failing tests**

Create a fake K-line fetcher that counts calls and prove a second scan for the same stock/day count reuses cache.

- [ ] **Step 2: Run tests red**

Run:

```powershell
go test ./backend/data -run "TestDuoKongKLineCache|TestDuoKongBackgroundScan" -count=1
```

- [ ] **Step 3: Implement cache and scan tasks**

Use package-level maps protected by `sync.RWMutex`. Cache key is `kLineCode|period|days`. Scan task contains `id/status/progress/total/results/error/startedAt/finishedAt`.

- [ ] **Step 4: Expose Wails wrappers**

Add methods in `app_common.go`.

- [ ] **Step 5: Run tests green**

Run focused Go tests.

### Task 3: Duo Kong Frontend

**Files:**
- Modify: `frontend/src/components/DuoKongLineSelector.vue`
- Modify: `frontend/src/App.vue`
- Modify: `frontend/src/components/researchIndex.vue`

**Interfaces:**
- Consumes: `StartDuoKongLineScan`, `GetDuoKongLineScanTask`

- [ ] **Step 1: Update UI copy**

Replace “周期天数/结果数量” with “信号窗口：近 N 日” and “K线计算长度”.

- [ ] **Step 2: Add board filter**

Options: 全部、主板、创业板、科创板、北交所.

- [ ] **Step 3: Add progress polling**

Start background scan and poll every 700ms until status is `done` or `failed`.

- [ ] **Step 4: Add multi-period button**

Use existing `OpenURL` to open EastMoney full-screen chart for the selected stock.

- [ ] **Step 5: Build frontend**

Run:

```powershell
npm --prefix frontend run build
```

### Task 4: AI Holding Summary Tool

**Files:**
- Create: `backend/data/holding_summary_api.go`
- Create: `backend/data/holding_summary_api_test.go`
- Modify: `app_common.go`
- Modify: `backend/data/tool_agent_extra.go`
- Modify: `backend/data/tools.go`

**Interfaces:**
- Produces: `GetHoldingSummary() *data.HoldingSummary`
- Produces AI tool name: `GetHoldingSummary`

- [ ] **Step 1: Write failing tests**

Seed followed stocks and trading records in sqlite. Assert summary includes cost price, volume, principal, latest snapshot price fallback, floating P/L, and related trading log count.

- [ ] **Step 2: Run tests red**

Run:

```powershell
go test ./backend/data -run "TestHoldingSummary" -count=1
```

- [ ] **Step 3: Implement summary service and Wails method**

Read `FollowedStock` where `cost_price > 0 AND volume > 0`, join relevant `TradingRecord`, compute structured totals.

- [ ] **Step 4: Register AI tool**

Expose as a data tool so the agent can call it when asked for持仓复盘.

- [ ] **Step 5: Run tests green**

Run focused tests.

### Task 5: Installer And Verification

**Files:**
- Generated: `build/bin/*.exe`
- Generated: `build/bin/*.exe` NSIS installer or `build/bin/*installer*.exe`

- [ ] **Step 1: Regenerate Wails bindings**

Run:

```powershell
wails build -s -nopackage
```

- [ ] **Step 2: Run verification**

Run:

```powershell
go test ./backend/data -run "TestDuoKong|TestHoldingSummary|TestSetCostPriceAndVolume" -count=1
go test ./... -run "^$" -count=0
npm --prefix frontend run build
```

- [ ] **Step 3: Build NSIS installer**

Run:

```powershell
wails build -clean -nsis -webview2 download
```

- [ ] **Step 4: Copy installer to desktop release folder**

Create `C:\Users\lxd\Desktop\A-Share-Desk-Release-20260709-NSIS` and copy the installer plus README.

- [ ] **Step 5: Commit**

Commit code changes after verification.

## Self-Review

- Spec coverage: all five user requirements are represented.
- Placeholder scan: no TBD/TODO placeholders remain.
- Type consistency: Wails wrappers and frontend calls use explicit names.
