# A-Share Hardening Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add A-share holding advice prompts, DuoKongLine full-market screening, market bulletin fallback data, and automatic trading log creation from self-selected stock costs.

**Architecture:** Backend owns reusable calculation and persistence behavior so Wails bindings and agent tools can consume the same source of truth. Frontend adds a research-center tab that mirrors the existing shape/indicator stock selection layout and calls one Wails API. Market bulletin charts gain a data bootstrap path instead of silently showing blank panels.

**Tech Stack:** Go 1.26, GORM/SQLite, Wails v2.11, Vue 3, Naive UI, Vite, existing EastMoney/Sina/Tdx K-line providers.

## Global Constraints

- Keep all work inside `C:\Users\lxd\Desktop\A-Share-Desk-Replica-20260709\source`.
- Do not modify the upstream source clone under `C:\Users\lxd\Desktop\gupiao\github-a-share-projects\ArvinLovegood__go-stock`.
- Use existing Wails, Vue, Naive UI, GORM, and K-line helper patterns.
- New behavior must be A-share focused and must not invent real-time data.
- New production logic must have targeted tests where practical.

---

### Task 1: Built-In Holding Advice Prompt

**Files:**
- Modify: `backend/data/prompt_template_api.go`
- Modify: `backend/data/prompt_template_api_test.go`

**Interfaces:**
- Produces default prompt template named `自选持仓成本复盘与投资建议`.

- [ ] Add a failing assertion in `TestDefaultPromptTemplates` that the template exists and mentions `costPrice`, `costVolume`, and `交易日志`.
- [ ] Run `go test ./backend/data -run TestDefaultPromptTemplates -count=1` and confirm the new assertion fails.
- [ ] Add the prompt to `DefaultPromptTemplates()`.
- [ ] Run the same test and confirm it passes.

### Task 2: Self-Selected Cost Writes Trading Log

**Files:**
- Modify: `backend/data/stock_data_api.go`
- Create: `backend/data/followed_stock_trading_record_test.go`

**Interfaces:**
- `StockDataApi.SetCostPriceAndVolume(price float64, volume int64, stockCode string) string` keeps its current return type.
- When cost and volume are valid, it creates one buy `TradingRecord` with reason `自选股成本自动生成`.

- [ ] Write a failing test that seeds one `FollowedStock`, calls `SetCostPriceAndVolume(47.5, 300, "sz300418")`, and expects one buy trading record with price `47.5`, volume `300`, and matching stock name.
- [ ] Write a failing test that calls the same setter twice and expects only one auto-created trading record for the same stock/price/volume/day.
- [ ] Implement helper logic to normalize stock code, update followed cost, and insert an idempotent buy log.
- [ ] Run `go test ./backend/data -run "TestSetCostPriceAndVolume" -count=1`.

### Task 3: DuoKongLine Screening Backend

**Files:**
- Create: `backend/data/duo_kong_line_api.go`
- Create: `backend/data/duo_kong_line_api_test.go`
- Modify: `app_common.go`

**Interfaces:**
- `DuoKongLineCondition` includes `Signal`, `Days`, `Limit`, `Keyword`, and optional minimum deviation.
- `DuoKongLineResult` includes code, name, industry, concept, latest close, fine line, thick line, deviation, signal, and signal date.
- Wails method: `GetDuoKongLineSelections(condition data.DuoKongLineCondition) []data.DuoKongLineResult`.

- [ ] Write tests for EMA values, fine/thick line construction, `fine_cross_up`, `fine_cross_down`, `fine_above_thick`, `fine_below_thick`, `both_rising`, and `both_falling`.
- [ ] Implement calculation using the existing screenshot-derived formula.
- [ ] Implement scanner that reads `AllStockInfo`, fetches day K-lines, evaluates condition, and limits results.
- [ ] Add the Wails App wrapper.
- [ ] Run `go test ./backend/data -run "TestDuoKongLine" -count=1`.

### Task 4: DuoKongLine Screening Frontend

**Files:**
- Create: `frontend/src/components/DuoKongLineSelector.vue`
- Modify: `frontend/src/components/researchIndex.vue`
- Modify: `frontend/src/App.vue`

**Interfaces:**
- Vue component calls generated Wails function `GetDuoKongLineSelections`.
- Research-center tab name: `多空线选股`.

- [ ] Add the new component with a compact left condition panel and right data table matching `SelectStock.vue` density.
- [ ] Add condition controls for cross-up, cross-down, above, below, both rising, both falling, days, limit, and keyword.
- [ ] Add operation buttons for K-line viewing and following stock where existing helpers make sense.
- [ ] Register the tab in `researchIndex.vue` and the left menu in `App.vue`.

### Task 5: Market Bulletin Empty-State and Data Bootstrap

**Files:**
- Modify: `backend/data/stock_changes_api.go`
- Modify: `backend/data/stock_change_history_api.go`
- Modify: `frontend/src/components/AnalyzeMartket.vue`
- Test: `backend/data/stock_changes_api_test.go`
- Test: `backend/data/stock_change_history_api_test.go`

**Interfaces:**
- `GetDailyChangeStats`, `GetChangeTypeDailyStats`, and `GetChangeRank` return useful data after bootstrap when live EastMoney data is reachable.
- Empty frontend panels display a clear message rather than a blank white area.

- [ ] Add test proving stock info enrichment queries by `securitycode`, not `sec_uri_tycode`.
- [ ] Add test proving `GetHistoryList` defaults page size when caller passes zero.
- [ ] Fix the enrichment query and page-size default.
- [ ] Add app/frontend bootstrap path: if history charts are empty, call `SaveStockChangesToHistory` for all known change types once, then reload charts.
- [ ] Add chart empty-state text for daily, type, rank, and bull/bear panels.

### Task 6: Bindings, Verification, and Packaging

**Files:**
- Generated: `frontend/wailsjs/go/main/App.*`
- Generated: `frontend/wailsjs/go/models.ts`

- [ ] Run `wails generate` after adding Wails methods.
- [ ] Run focused Go tests for prompt, cost log, DuoKongLine, and stock change fixes.
- [ ] Run `go test ./... -run "^$" -count=0`.
- [ ] Run `npm --prefix frontend run build`.
- [ ] Run `wails build`.
- [ ] Repackage a new Windows zip under `C:\Users\lxd\Desktop\A-Share-Desk-Replica-20260709\release`.

## Self-Review

- Spec coverage: all five user requirements map to Tasks 1-6.
- Placeholder scan: no TODO/TBD placeholders are present.
- Type consistency: backend condition/result names are defined in Task 3 and consumed in Task 4.
