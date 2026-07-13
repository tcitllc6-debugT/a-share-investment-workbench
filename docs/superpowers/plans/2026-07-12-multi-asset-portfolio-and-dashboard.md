# Multi-Asset Portfolio and Dashboard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Optimize the dashboard and watchlist, add stock quick trades, and build a unified CNY-denominated portfolio center for stocks, funds, bonds, precious metals, forex, futures, crypto, and bank wealth products.

**Architecture:** Preserve the current stock APIs as compatibility surfaces while introducing a generic instrument/transaction/quote/FX ledger. Positions are derived from immutable transactions, with asset-specific calculators behind one portfolio service and replaceable quote providers behind one batch interface.

**Tech Stack:** Go 1.24, GORM/SQLite, Wails v2, Vue 3, Naive UI, Node test runner, Vite.

## Global Constraints

- The application overview continues to show stock holdings only.
- The portfolio center overview aggregates every asset class in CNY.
- Bank wealth product valuation is manual or CSV-based.
- No broker, bank, exchange account, or automatic order integration is introduced.
- No Python runtime is bundled into the desktop application.
- Existing stock transaction and FIFO results must survive migration unchanged.
- Per user instruction, implementation tasks complete before the consolidated test phase.

---

### Task 1: Dashboard local-first radar and mixed news

**Files:**
- Modify: `frontend/src/components/DecisionWorkbench.vue`
- Modify: `frontend/src/components/decision-workbench.mjs`
- Modify: `backend/data/market_news_api.go`
- Modify: `app.go`
- Modify generated bindings: `frontend/wailsjs/go/main/App.js`, `frontend/wailsjs/go/main/App.d.ts`

**Interfaces:**
- Consumes: `GetIntradayRadarSnapshot`, `RefreshIntradayRadar`.
- Produces: `GetTelegraphBriefListBySources(sources []string, limit int)` and mixed source rows.

- [ ] Replace the dashboard network radar load with immediate snapshot publication and detached refresh.
- [ ] Add a batch lightweight news query for `财联社电报` and `新浪财经`.
- [ ] Normalize, deduplicate, interleave, and source-label dashboard news.
- [ ] Keep stale successful data visible when background refresh fails.

### Task 2: Watchlist toolbar simplification

**Files:**
- Modify: `frontend/src/components/stock.vue`

**Interfaces:**
- Consumes: existing left sidebar group navigation.
- Produces: a toolbar containing only search and result count.

- [ ] Remove the `全部分组` select from the direct all-watchlist panel.
- [ ] Remove its computed options and filter state only when no other view consumes them.
- [ ] Keep per-stock `设置分组` and sidebar grouping intact.

### Task 3: Stock quick add and sell transactions

**Files:**
- Modify: `frontend/src/components/TradingRecordManager.vue`
- Modify: `backend/data/stock_data_api.go`
- Modify: `app.go`

**Interfaces:**
- Consumes: existing `AddTradingRecord`, latest stock price, FIFO holdings.
- Produces: `GetStockAvailableVolume(stockCode string) int64` and prefilled quick transaction dialogs.

- [ ] Add `加仓` and `卖出` buttons beside edit/delete/K-line actions.
- [ ] Open the add dialog with identity, direction, current price, and available volume context.
- [ ] Reject a sell quantity above the FIFO remaining position in both frontend and backend.
- [ ] Refresh records, statistics, stock holdings, and dashboard events after save.

### Task 4: Unified asset schema and synchronous migration

**Files:**
- Create: `backend/data/asset_models.go`
- Create: `backend/data/asset_migration.go`
- Modify: `backend/db/schema_migration.go`
- Modify: `main.go`

**Interfaces:**
- Produces: `AssetInstrument`, `AssetTransaction`, `AssetQuote`, `FXRate`, `PortfolioSnapshot`, `SchemaVersion`.
- Produces: `MigrateLegacyStockLedger(tx *gorm.DB) error`.

- [ ] Add the five generic ledger tables and schema version metadata.
- [ ] Back up the SQLite database before the new schema version is applied.
- [ ] Migrate legacy stock records idempotently with stable `external_ref` values.
- [ ] Keep the legacy stock tables readable for one compatibility cycle.
- [ ] Abort startup with a clear error if migration or reconciliation fails.

### Task 5: Asset calculators and portfolio service

**Files:**
- Create: `backend/data/asset_calculator.go`
- Create: `backend/data/portfolio_service.go`
- Create: `backend/data/portfolio_types.go`
- Modify: `app.go`

**Interfaces:**
- Produces: `GetPortfolioOverview() PortfolioOverview`.
- Produces: `GetAssetPositions(assetType string) []AssetPosition`.
- Produces: `AddAssetTransaction(input AssetTransactionInput)`, `UpdateAssetTransaction`, `DeleteAssetTransaction`.

- [ ] Implement FIFO stock/fund/spot calculators.
- [ ] Implement bond clean price, accrued interest, and coupon cash flow calculations.
- [ ] Implement unit conversion and FX conversion for precious metals.
- [ ] Implement currency balance valuation for forex.
- [ ] Implement long/short futures P&L, multiplier, margin, and notional calculations.
- [ ] Implement fractional crypto balances and quote-currency conversion.
- [ ] Implement principal and NAV bank product valuation.
- [ ] Aggregate total cost, market value, floating P&L, realized P&L, cash flow, currency exposure, and stale quote counts.

### Task 6: Quote and FX provider framework

**Files:**
- Create: `backend/provider/assetquote/provider.go`
- Create: `backend/provider/assetquote/registry.go`
- Create: `backend/provider/assetquote/cache.go`
- Create: `backend/provider/assetquote/eastmoney.go`
- Create: `backend/provider/assetquote/sina.go`
- Create: `backend/provider/assetquote/frankfurter.go`
- Create: `backend/provider/assetquote/binance.go`
- Create: `backend/provider/assetquote/coingecko.go`
- Modify: `app.go`

**Interfaces:**
- Produces: `QuoteProvider.Supports` and `FetchQuotes`.
- Produces: `RefreshAssetQuotes(assetTypes []string) QuoteRefreshResult`.
- Produces: provider health information and manual quote fallback.

- [ ] Implement provider registry, timeout, retry, rate limit, and per-provider circuit breaker.
- [ ] Reuse current Eastmoney/Sina stock and fund APIs through adapters.
- [ ] Map bond, metal, forex, and futures public endpoints through Go adapters.
- [ ] Add Frankfurter/ECB daily FX fallback.
- [ ] Add Binance public crypto quotes with CoinGecko keyless fallback.
- [ ] Cache last successful quotes and retain them on provider failure.

### Task 7: Multi-asset portfolio center shell

**Files:**
- Modify: `frontend/src/components/HoldingCenter.vue`
- Create: `frontend/src/components/portfolio/PortfolioOverview.vue`
- Create: `frontend/src/components/portfolio/AssetPositionTable.vue`
- Create: `frontend/src/components/portfolio/AssetTransactionEditor.vue`
- Create: `frontend/src/components/portfolio/QuoteStatus.vue`
- Modify: `frontend/src/navigation/workspace-navigation.mjs`
- Modify: `frontend/src/components/shell/WorkspaceSidebar.vue`

**Interfaces:**
- Consumes: portfolio APIs and query `assetType`.
- Produces: nine sidebar destinations and category-specific views.

- [ ] Rename the section to `金融产品持仓` with the one-line subtitle `查看各类资产的市值、成本与收益`.
- [ ] Add overview, stock, fund/ETF, bond, precious metal, forex, futures, crypto, and bank product destinations.
- [ ] Render all-asset totals, allocation, currency exposure, quote freshness, and recent transactions in the overview.
- [ ] Render category-specific columns without nesting cards.
- [ ] Keep horizontal and vertical scrolling reachable at 300% DPI.

### Task 8: Asset transaction forms

**Files:**
- Create: `frontend/src/components/portfolio/asset-form-schema.mjs`
- Modify: `frontend/src/components/portfolio/AssetTransactionEditor.vue`
- Modify: `frontend/src/components/portfolio/AssetPositionTable.vue`

**Interfaces:**
- Consumes: asset type and selected instrument.
- Produces: validated transaction payloads for all eight asset classes.

- [ ] Add shared identity, quantity, price, currency, fee, time, and notes fields.
- [ ] Add bond face value, coupon, accrued interest, frequency, and maturity fields.
- [ ] Add metal unit fields.
- [ ] Add futures side, multiplier, margin, settlement, and expiry fields.
- [ ] Add bank product principal/NAV mode, institution, product code, and maturity fields.
- [ ] Add add/reduce/close actions appropriate to each asset class.

### Task 9: Bank product CSV import

**Files:**
- Create: `backend/data/bank_product_import.go`
- Modify: `app.go`
- Create: `frontend/src/components/portfolio/BankProductImport.vue`

**Interfaces:**
- Produces: `PreviewBankProductCSV(path string)` and `ImportBankProductCSV(batchID string, rows []BankProductImportRow)`.
- Produces: `RollbackAssetImport(batchID string)`.

- [ ] Parse UTF-8 and GB18030 CSV files with a documented column template.
- [ ] Validate products, currencies, dates, quantities, and NAV values before writing.
- [ ] Show a preview with row-level errors.
- [ ] Import atomically and support complete batch rollback.

### Task 10: Portfolio snapshots and AI summary

**Files:**
- Create: `backend/data/portfolio_snapshot.go`
- Modify: `backend/data/tool_agent_extra.go`
- Modify: `backend/data/tools.go`
- Modify: `frontend/src/components/AIConsole.vue`

**Interfaces:**
- Produces: `CreatePortfolioSnapshot` and `GetPortfolioSnapshots`.
- Produces: AI tool `GetMultiAssetPortfolioSummary`.

- [ ] Save one daily snapshot after successful quote refresh or transaction mutation.
- [ ] Expose asset allocation, currency exposure, stale data, concentration, and P&L to AI as structured JSON.
- [ ] Keep stock-only AI tools backward compatible.

### Task 11: Consolidated test phase

**Files:**
- Create/modify: backend `*_test.go` and frontend `*.test.mjs` files corresponding to Tasks 1-10.

**Interfaces:**
- Consumes: all completed implementation surfaces.
- Produces: regression evidence for calculations, migration, UI contracts, and builds.

- [ ] Add dashboard snapshot and mixed-news tests.
- [ ] Add quick-trade and oversell tests.
- [ ] Add table-driven calculator tests for all eight asset classes.
- [ ] Add migration idempotency and reconciliation tests.
- [ ] Add provider fallback, stale cache, timeout, and FX tests.
- [ ] Add bank CSV preview/import/rollback tests.
- [ ] Add sidebar, copy, responsive, and 300% DPI source-contract tests.
- [ ] Run `npm test` in `frontend` and require zero failures.
- [ ] Run `go test ./... -count=1` and require zero failures.
- [ ] Run `npm run build` and `wails build -clean`.

### Task 12: Git history and GitHub delivery

**Files:**
- Track all source, documentation, migrations, and tests.
- Exclude generated runtime data, logs, caches, secrets, and local database backups.

**Interfaces:**
- Produces: branch `codex/multi-asset-portfolio`, local commits, remote push, and pull request when GitHub authorization is available.

- [ ] Initialize Git with `main`, create `codex/multi-asset-portfolio`, and verify `.gitignore` coverage.
- [ ] Commit the imported clean baseline separately from the feature implementation.
- [ ] Commit implementation and verification evidence.
- [ ] Connect the user-authorized GitHub remote, push the branch, and create a pull request.
