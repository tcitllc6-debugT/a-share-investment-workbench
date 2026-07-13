# A-Share Desk Reliability And Update Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver reliable full-market Duo Kong screening, FIFO-ledger holdings, safe startup/news synchronization, and a server-hosted signed Windows self-update path.

**Architecture:** Preserve the existing Wails/Vue structure while moving scan state and K-line cache into focused Go services. Keep updates static and server-simple: a signed manifest drives a client-side verifier and an external Windows swap helper, with no private key on the server.

**Tech Stack:** Go 1.26, Wails v2, GORM/SQLite, Vue 3 + Naive UI, NSIS, Ed25519, SHA-256, nginx static hosting.

## Global Constraints

- Work only in `C:\Users\lxd\Desktop\A-Share-Desk-Replica-20260709\source`.
- Never modify the upstream clone.
- Daily K-line lookback is fixed at 120 and the signal window is fixed at the most recent three trading days.
- The release signing private key must be ignored by git and never uploaded to the server.
- Every behavior change starts with a focused failing test and ends with its passing test.

### Task 1: Persistent Duo Kong domain model

**Files:** `backend/data/duo_kong_line_api.go`, `backend/data/duo_kong_line_api_test.go`, `backend/data/duo_kong_line_cache_test.go`, `backend/models/models.go`, `main.go`.

- [ ] Add failing tests for actual trigger dates, event/state/trend semantics, three-day windows, consecutive days, retained-above status and deterministic ranking.
- [ ] Add `DuoKongKLineCache` GORM model and migrate it with the application schema.
- [ ] Replace process-memory-only rows with SQLite entries keyed by code/date, with five-minute intraday freshness and next-trading-day closed-market freshness.
- [ ] Implement bounded workers, cancellation, task expiry and batched result snapshots.
- [ ] Run `go test ./backend/data -run 'DuoKong' -count=1`.

### Task 2: Duo Kong filters, local charts and selector UI

**Files:** `backend/data/duo_kong_line_api.go`, `app_common.go`, `frontend/src/components/DuoKongLineSelector.vue`, `frontend/src/components/kline/indicators/duoKongLine.test.mjs`.

- [ ] Add failing backend tests for board/ST/new-listing and quote-metric filters.
- [ ] Add current-quote enrichment and filters for price, free market cap, turnover, amount, volume ratio, industry, concept and multi-select boards.
- [ ] Add selector sections for event/state/trend conditions, practical filters and scan health counters.
- [ ] Make both `日K` and `多周期` use the local `StockLightweightKlineChart` modal.
- [ ] Make default ranking signal freshness, absolute deviation and volume confirmation; add an interactive deviation-column sorter.
- [ ] Run focused Go tests, `node --test src/components/kline/indicators/duoKongLine.test.mjs`, and `npm run build` from `frontend`.

### Task 3: Ledger-first holding summary

**Files:** `backend/data/stock_data_api.go`, `backend/data/holding_summary_api.go`, `backend/data/holding_summary_api_test.go`, `backend/data/followed_stock_trading_record_test.go`.

- [ ] Add failing tests that editing a cost import updates one synthetic record and that FIFO holdings reflect buys, sells and fees.
- [ ] Mark synthetic imports with a stable machine-readable reason and make cost edits idempotent.
- [ ] Calculate holding summary quantities, cost, market value and floating profit from FIFO lots, exposing import provenance to the AI tool.
- [ ] Run `go test ./backend/data -run 'HoldingSummary|SetCostPrice|TradingRecord' -count=1`.

### Task 4: Safe migrations and observable news synchronization

**Files:** `backend/db/db.go`, `main.go`, `backend/models/models.go`, `backend/data/market_news_api.go`, `backend/data/market_news_api_test.go`, `app.go`, `app_common.go`, `frontend/src/components/market.vue`.

- [ ] Add migration and news-health tests using a temporary SQLite database and fake fetchers.
- [ ] Run migrations synchronously before the UI starts, create a pre-migration backup and retain the latest five backups.
- [ ] Add `FetchedAt` to telegraphs and source-health persistence.
- [ ] Fetch startup sources concurrently under source-specific deadlines; return status and show the last synchronization state in the market UI.
- [ ] Run focused tests and `go test ./... -run '^$' -count=0`.

### Task 5: Signed update channel and Windows updater

**Files:** `backend/provider/update/*`, `app.go`, `app_common.go`, `frontend/src/components/about.vue`, `frontend/src/App.vue`, `update_helper_windows.go`, `scripts/build-windows.sh`, `scripts/build-macos.sh`, `scripts/release-update.ps1`, `deploy/update-server/*`, `.gitignore`.

- [ ] Add failing tests for semantic-version ordering, Ed25519 manifest verification, SHA-256 verification and platform-asset selection.
- [ ] Implement server-manifest update provider; remove the disabled-provider short circuit.
- [ ] Add a persistent update endpoint setting and automatic startup check with a user-controlled update modal.
- [ ] Implement Windows staged download, external replacement, backup, restart and startup acknowledgement rollback.
- [ ] Correct A-Share Desk macOS artifact names and document signing/notarization inputs.
- [ ] Add release script that packages assets, creates signed manifest JSON and produces server deployment files without storing secrets.
- [ ] Run update package tests and build a Windows NSIS installer with WebView2 download support.

### Task 6: Deploy and validate

**Files:** generated release assets and server configuration only.

- [ ] Inspect the user-provided server before changing it; preserve unrelated services.
- [ ] Install a least-privilege static update directory and service configuration.
- [ ] Upload a signed test manifest and test asset, then validate the public endpoint and client metadata check.
- [ ] Build the release installer, copy it to the desktop release location and document the endpoint and release command.
- [ ] Run `go test ./... -count=1`, `npm --prefix frontend run build`, `wails build -clean -webview2 download`, and NSIS packaging.
