# Decision Workbench And Intraday Radar Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the bottom-menu desktop surface with the approved decision workbench, consolidate stock screening, and add a safe intraday radar that finds, alerts, and opens analysis without placing orders.

**Architecture:** Keep the Wails/Go/SQLite application boundary. Vue owns layout, navigation, filter state, tables, and route transitions. Go owns raw Eastmoney event retrieval, radar classification, priority/deduplication, follow-list matching, persisted history, and Windows notifications. The radar produces read-only events and allows only K-line, AI-analysis, and follow-stock actions.

**Tech Stack:** Go, GORM/SQLite, Wails v2 bindings/runtime events, Vue 3 Composition API, Naive UI, existing lightweight-charts and Windows notification provider.

## Global Constraints

- Do not add broker APIs, order placement, account credentials, or automatic trading.
- Keep existing uncommitted A-Share feature work intact; do not reset or revert files.
- Do not create commits in this shared dirty worktree unless the user explicitly requests a commit.
- Use the existing red-up/green-down A-share convention and local-first SQLite storage.
- Keep all fixed UI controls accessible at CSS viewport widths of 1440px, 960px, and 480px.
- Reuse existing Wails bindings and regenerate frontend bindings after exported Go APIs change.
- Add a failing test before new Go behavior and run it before the implementation step.

---

### Task 1: Add radar domain types, classification, and alert deduplication

**Files:**
- Create: `backend/data/intraday_radar.go`
- Create: `backend/data/intraday_radar_test.go`
- Modify: `app_common.go`
- Modify: `frontend/wailsjs/go/main/App.d.ts`
- Modify: `frontend/wailsjs/go/main/App.js`
- Modify: `frontend/wailsjs/go/models.ts`

**Interfaces:**
- Consumes: `StockChangesApi.GetAllStockChangesWithPaging(int)`, `StockDataApi.GetFollowList(int)`, `NewAlertWindowsApi`.
- Produces: `RadarFilter`, `RadarEvent`, `RadarSnapshot`, `RadarSettings`, `NewIntradayRadarService`, `GetIntradayRadarSnapshot`, `RefreshIntradayRadar`, `GetIntradayRadarSettings`, `UpdateIntradayRadarSettings`.

- [ ] **Step 1: Write the failing classification and cooldown tests**

```go
func TestClassifyRadarEvent(t *testing.T) {
    event := classifyRadarEvent(StockChangeItem{Code: "300418", ChangeType: 8201, TypeName: "火箭发射"}, true)
    if event.Category != RadarCategoryMomentum || event.Priority != RadarPriorityP1 || !event.IsFollowed {
        t.Fatalf("unexpected radar event: %#v", event)
    }
}

func TestRadarNotificationCooldown(t *testing.T) {
    service := NewIntradayRadarServiceWithClock(func() time.Time { return time.Date(2026, 7, 10, 10, 0, 0, 0, time.Local) })
    event := RadarEvent{Code: "300418", Category: RadarCategoryMomentum, Priority: RadarPriorityP1}
    if !service.shouldNotify(event) || service.shouldNotify(event) {
        t.Fatal("same event should notify once during cooldown")
    }
}
```

- [ ] **Step 2: Run the targeted tests and verify they fail**

Run: `go test ./backend/data -run 'TestClassifyRadarEvent|TestRadarNotificationCooldown' -count=1`

Expected: FAIL because `classifyRadarEvent` and `NewIntradayRadarServiceWithClock` do not exist.

- [ ] **Step 3: Implement read-only radar mapping and snapshot service**

```go
type RadarEvent struct {
    StockChangeItem
    Category   string `json:"category"`
    Priority   string `json:"priority"`
    IsFollowed bool   `json:"isFollowed"`
    IsHolding  bool   `json:"isHolding"`
    Action     string `json:"action"`
}

func (s *IntradayRadarService) Refresh(filter RadarFilter) *RadarSnapshot {
    raw := NewStockChangesApi().GetAllStockChangesWithPaging(500)
    if raw == nil { return s.lastSnapshot("数据源暂不可用") }
    events := s.classifyAndFilter(raw.Data, filter)
    s.persist(events)
    s.notify(events)
    return s.store(events)
}
```

Implement the six design categories, P1/P2/P3 priority policy, follow-list/holding matching, 15-minute in-memory notification cooldown, last-success metadata, and only `open_kline`, `open_ai_analysis`, `follow_stock` action values.

- [ ] **Step 4: Expose the Wails methods and regenerate bindings**

```go
func (a *App) RefreshIntradayRadar(filter data.RadarFilter) *data.RadarSnapshot {
    return data.NewIntradayRadarService().Refresh(filter)
}
```

Generate bindings with `wails generate module` (or the installed Wails equivalent), then confirm the TypeScript declaration exposes the four radar methods.

- [ ] **Step 5: Run the Go tests and commit the coherent backend unit**

Run: `go test ./backend/data -run 'TestClassifyRadarEvent|TestRadarNotificationCooldown|Test.*StockChange' -count=1`

Expected: PASS.

Record the passing command output in the task update; leave the shared worktree uncommitted.

### Task 2: Build the desktop shell and approved decision workbench

**Files:**
- Create: `frontend/src/components/DecisionWorkbench.vue`
- Create: `frontend/src/components/shell/WorkspaceNavigation.vue`
- Create: `frontend/src/components/shell/WorkspaceHeader.vue`
- Create: `frontend/src/utils/navigation.mjs`
- Create: `frontend/src/utils/navigation.test.mjs`
- Modify: `frontend/src/App.vue`
- Modify: `frontend/src/router/router.js`
- Modify: `frontend/src/style.css`

**Interfaces:**
- Consumes: `GetHoldingSummary`, `GetFollowList`, `GetTelegraphList`, `GlobalStockIndexes`, `GetDuoKongLineSelections` and router names.
- Produces: `dashboard` route, stable `primaryNavigation` config, responsive `compact/standard/wide` shell, global `openKline` and `openResearch` route/event transitions.

- [ ] **Step 1: Write the failing navigation breakpoint tests**

```js
import { getNavigationPresentation } from './navigation.mjs'

test('uses icon rail below 960px and drawer below 640px', () => {
  expect(getNavigationPresentation(1200)).toBe('full')
  expect(getNavigationPresentation(800)).toBe('rail')
  expect(getNavigationPresentation(480)).toBe('drawer')
})
```

- [ ] **Step 2: Run the frontend test and verify it fails**

Run: `node --test frontend/src/utils/navigation.test.mjs`

Expected: FAIL because `navigation.mjs` does not export `getNavigationPresentation`.

- [ ] **Step 3: Implement the shell without deleting existing functional pages**

```vue
<WorkspaceNavigation :items="primaryNavigation" :presentation="navigationPresentation" />
<section class="workspace-shell__main">
  <WorkspaceHeader :market-status="marketStatus" @search="openSymbol" />
  <main class="workspace-shell__content"><RouterView /></main>
</section>
```

Make `dashboard` the default route and retain `stock`, `market`, `klineAnalysis`, `research`, and `settings` routes. The shell must preserve existing update notification logic and Wails `--wails-draggable: no-drag` control regions.

- [ ] **Step 4: Implement DecisionWorkbench with live data and truthful empty states**

```vue
<section class="decision-workbench__grid">
  <HoldingOverview :summary="holdingSummary" :rows="followedStocks" />
  <ActionQueue :items="actionItems" @open="openAction" />
  <RadarSummary :snapshot="radarSnapshot" @open="goToRadar" />
  <NewsDigest :items="news" />
</section>
```

Use only successful backend data. Render `等待首次同步` if an API returns no data and preserve the existing K-line modal component for stock actions.

- [ ] **Step 5: Run layout tests, frontend build, and commit the UI unit**

Run: `node --test frontend/src/utils/navigation.test.mjs frontend/src/components/duo-kong-layout.test.mjs`

Run: `npm --prefix frontend run build`

Expected: all tests pass and Vite reports a successful production build.

Record the passing command output in the task update; leave the shared worktree uncommitted.

### Task 3: Consolidate existing stock screens into the unified selector

**Files:**
- Create: `frontend/src/components/ScreeningHub.vue`
- Create: `frontend/src/components/screening/ScreeningResultActions.vue`
- Create: `frontend/src/components/screening/ScreeningHub.test.mjs`
- Modify: `frontend/src/components/researchIndex.vue`
- Modify: `frontend/src/components/DuoKongLineSelector.vue`
- Modify: `frontend/src/components/allStockList.vue`
- Modify: `frontend/src/components/SelectStock.vue`
- Modify: `frontend/src/router/router.js`

**Interfaces:**
- Consumes: existing `DuoKongLineSelector`, `AllStockList`, `SelectStock`, custom strategy APIs and `StockLightweightKlineChart`.
- Produces: `screening` route and shared result action contract `{ openKline(row), openAiAnalysis(row), followStock(row) }`.

- [ ] **Step 1: Write the failing selector configuration test**

```js
import { screeningTabs } from '../ScreeningHub.vue'

test('screening hub exposes the approved six tabs in order', () => {
  expect(screeningTabs.map(item => item.key)).toEqual(['quick', 'duo-kong', 'pattern', 'indicator', 'query', 'saved'])
})
```

- [ ] **Step 2: Run the test and verify it fails**

Run: `node --test frontend/src/components/screening/ScreeningHub.test.mjs`

Expected: FAIL because `ScreeningHub.vue` does not exist.

- [ ] **Step 3: Implement the selector hub and shared result actions**

```js
export const screeningTabs = [
  { key: 'quick', label: '快捷策略' },
  { key: 'duo-kong', label: '多空线' },
  { key: 'pattern', label: '形态' },
  { key: 'indicator', label: '指标' },
  { key: 'query', label: '条件式' },
  { key: 'saved', label: '我的策略' },
]
```

Move the three scan components out of the research-center visual flow by composing them in `ScreeningHub`; do not duplicate their data fetching. Add K-line, multi-period, AI analysis, and follow actions to result tables where the source row has a code.

- [ ] **Step 4: Reduce research center to research-only tabs**

Keep AI reports, recommendation records, trading log, prompt templates, scheduled tasks, and MCP services. Remove shape, indicator, and DuoKong panels from its tab list but keep their components available through the selector.

- [ ] **Step 5: Run the selector test/build and commit**

Run: `node --test frontend/src/components/screening/ScreeningHub.test.mjs`

Run: `npm --prefix frontend run build`

Expected: PASS and successful build.

Record the passing command output in the task update; leave the shared worktree uncommitted.

### Task 4: Deliver the intraday radar interface and safe analysis actions

**Files:**
- Create: `frontend/src/components/IntradayRadar.vue`
- Create: `frontend/src/components/radar/radarFilters.mjs`
- Create: `frontend/src/components/radar/radarFilters.test.mjs`
- Modify: `frontend/src/components/stockChangesMonitor.vue`
- Modify: `frontend/src/components/settings.vue`
- Modify: `frontend/src/router/router.js`

**Interfaces:**
- Consumes: `RefreshIntradayRadar`, `GetIntradayRadarSnapshot`, `GetIntradayRadarSettings`, `UpdateIntradayRadarSettings`, `EventsEmit('klineSelectStock')`.
- Produces: `radar` route; filter state matching `RadarFilter`; settings controls for P1/P2 and watchlist-only notifications.

- [ ] **Step 1: Write the failing UI filter test**

```js
import { buildRadarFilter } from './radarFilters.mjs'

test('builds a watchlist-only risk filter', () => {
  expect(buildRadarFilter({ tab: 'risk', onlyFollowed: true })).toMatchObject({ categories: ['risk', 'breakdown'], onlyFollowed: true })
})
```

- [ ] **Step 2: Run the test and verify it fails**

Run: `node --test frontend/src/components/radar/radarFilters.test.mjs`

Expected: FAIL because `buildRadarFilter` does not exist.

- [ ] **Step 3: Implement the radar screen and refresh loop**

```vue
<RadarFilters v-model="filters" @refresh="refresh" />
<RadarEventFeed :events="snapshot.events" @open-kline="openKline" @open-ai="openAi" @follow="follow" />
<RadarPulse :summary="snapshot.summary" :last-success-at="snapshot.lastSuccessAt" :error="snapshot.error" />
```

Use 15-second refresh only while the app reports A-share trading time; show manual-refresh and last-success metadata at all times. Send existing Wails desktop notifications only from Go. Route K-line through `EventsEmit('klineSelectStock', ...)` and analysis through `/research` query parameters.

- [ ] **Step 4: Move the legacy monitor behind a history link and add settings**

Retain `stockChangesMonitor.vue` as the detailed history page, linked from radar as `查看历史`. In settings add explicit controls: radar enabled, P1 notifications, P2 notifications, only follow/holding notifications, and cooldown minutes. Do not show any buy/sell command.

- [ ] **Step 5: Run tests/build and commit the feature**

Run: `go test ./backend/data -run 'TestClassifyRadarEvent|TestRadarNotificationCooldown' -count=1`

Run: `node --test frontend/src/components/radar/radarFilters.test.mjs`

Run: `npm --prefix frontend run build`

Expected: all commands pass.

Record the passing command output in the task update; leave the shared worktree uncommitted.

### Task 5: High-DPI verification, production build, and review

**Files:**
- Modify: `frontend/src/components/shell/WorkspaceNavigation.vue`
- Modify: `frontend/src/components/DecisionWorkbench.vue`
- Modify: `frontend/src/components/IntradayRadar.vue`
- Modify: `docs/a-share-desk-user-manual.md`

**Interfaces:**
- Consumes: all completed routes and data APIs.
- Produces: documented navigation and radar usage, verified 1440px/960px/480px screenshots.

- [ ] **Step 1: Add failing layout assertions for the compact rail and result overflow**

```js
test('compact mode keeps the primary action reachable', () => {
  expect(getNavigationPresentation(480)).toBe('drawer')
  expect(getTableScrollWidth('radar', 480)).toBeGreaterThan(700)
})
```

- [ ] **Step 2: Run the assertions and verify they fail before the final CSS adjustment**

Run: `node --test frontend/src/utils/navigation.test.mjs`

Expected: FAIL until `getTableScrollWidth` is introduced.

- [ ] **Step 3: Implement responsive constraints and document user-facing workflow**

Set fixed header/side navigation dimensions with CSS grid; use horizontal result-table scroll at narrow widths; move noncritical columns behind responsive visibility. Document where to find the selector, radar categories, notification behavior, and the fact that radar cannot place trades.

- [ ] **Step 4: Run full verification**

Run: `go test ./... -count=1`

Run: `npm --prefix frontend run build`

Run: `wails build -clean`

Expected: Go tests pass, Vite build succeeds, and Wails produces `build/bin/a-share-desk.exe`.

- [ ] **Step 5: Capture responsive screenshots and commit**

Use Playwright with the local development surface at 1440px, 960px, and 480px. Confirm no overlapping controls, no clipped navigation, and no blank radar state after a handled data-source failure.

Record the passing command output in the task update; leave the shared worktree uncommitted.
