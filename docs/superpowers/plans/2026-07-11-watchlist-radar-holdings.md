# Watchlist, Radar, and Holdings Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make watchlist additions visible, make the radar understandable, keep holdings and small windows usable, and publish the updated replica into a clean desktop folder.

**Architecture:** Extract pure symbol identity and radar-copy helpers so view components do not compare raw provider symbols. Keep the current Wails APIs and trading-record ledger intact; update only presentation, view synchronization, and local project delivery.

**Tech Stack:** Vue 3, Naive UI, Node test, Wails v2, Go, PowerShell.

## Global Constraints

- Work only in `C:\Users\lxd\Desktop\A-Share-Desk-Replica-20260709\source` until release copying.
- Do not edit generated `frontend/wailsjs/**` files manually.
- Preserve existing user data and do not copy data, logs, or databases into the clean release folder.
- Radar remains analysis-only and cannot place orders.
- Use failing tests before production changes.

---

### Task 1: Canonical watchlist symbol identity

**Files:**
- Create: `frontend/src/utils/stock-symbol.mjs`
- Create: `frontend/src/utils/stock-symbol.test.mjs`
- Modify: `frontend/src/components/stock.vue`

**Interfaces:**
- `canonicalStockSymbol(value)` returns a lowercase exchange-prefixed identity.
- `synchronizeWatchlistQuotes(followedStocks, quotePayloads)` retains every followed symbol until a quote is received.

- [ ] Write failing tests for `sz300418`, `300418.SZ`, and `SZ300418` sharing one identity and for an unquoted followed stock being retained.
- [ ] Run `node --test src/utils/stock-symbol.test.mjs` and confirm RED.
- [ ] Use the helpers in follow, group-switch, and quote update paths.
- [ ] Run the focused test and `npm test`.

### Task 2: Radar explanation and holdings strip

**Files:**
- Create: `frontend/src/components/radar/radar-explanation.mjs`
- Create: `frontend/src/components/radar/radar-explanation.test.mjs`
- Modify: `frontend/src/components/IntradayRadar.vue`
- Modify: `frontend/src/components/HoldingCenter.vue`
- Modify: `frontend/src/components/TradingRecordManager.vue`

**Interfaces:**
- `describeRadarEvent(event)` returns trigger copy and a review action for an event category.
- The holdings summary has a single horizontal metric strip with overflow at narrow width.

- [ ] Write RED tests for opportunity/risk/capital explanations and source-contract checks for the one-line holdings strip.
- [ ] Add the radar explanation disclosure and per-event explanation.
- [ ] Rename the holding surface and replace the wrapping statistics grid with the horizontal strip.
- [ ] Run focused tests and `npm test`.

### Task 3: Small-window scrolling

**Files:**
- Modify: `frontend/src/App.vue`
- Modify: `frontend/src/components/shell/WorkspaceNavigation.vue`
- Modify: `frontend/src/components/workspace-high-dpi.test.mjs`

**Interfaces:**
- Main content has a reachable vertical scroll container.
- Secondary navigation can scroll independently.

- [ ] Write RED source-contract checks for explicit main/secondary overflow behavior.
- [ ] Add bounded flex heights and vertical overflow without enabling page-wide clipping.
- [ ] Run focused tests and `npm test`.

### Task 4: Independent desktop project directory

**Files:**
- Create: `C:\Users\lxd\Desktop\A股智投工作台\README.md`
- Create: `C:\Users\lxd\Desktop\A股智投工作台\source\...`
- Create: `C:\Users\lxd\Desktop\A股智投工作台\release\...`

**Interfaces:**
- The target directory contains source and a clean Windows release, but no user data, logs, or source `node_modules`.

- [ ] Build the desktop executable with `wails build -clean`.
- [ ] Copy source and release using explicit exclusions for generated local state.
- [ ] Verify the expected executable and README exist and excluded directories do not.

### Task 5: Verification and report design acceptance

**Files:**
- Modify: `docs/superpowers/specs/2026-07-11-watchlist-radar-holdings-design.md`

- [ ] Run `npm test`, `npm run build`, `go test ./... -run '^$' -count=0`, and `wails build -clean`.
- [ ] Check the built desktop shell at a compact viewport and confirm the fixed watchlist, radar guidance, and reachable scrolling surfaces.
- [ ] Record that the research committee report is a manual single-stock design pending a dedicated data-source and model-cost implementation phase.

