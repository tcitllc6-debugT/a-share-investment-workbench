# A-Share Desk Reliability And Update Design

## Goal

Make the A-share replica dependable for full-market Duo Kong screening, make the transaction ledger the single holding truth, make startup data initialization observable and safe, and ship a signed self-update channel owned by A-Share Desk.

## Scope And Decisions

### Duo Kong screening

- The signal window is always the most recent three trading days. The backend fetches 120 daily bars; this is an implementation detail and is not editable in the UI.
- Signals have explicit categories: events (`fine_cross_up`, `fine_cross_down`), states (`fine_above_thick`, `fine_below_thick`), and trends (`both_rising`, `both_falling`, `fine_turn_up`, `fine_turn_down`, `thick_turn_up`, `thick_turn_down`).
- A match records the actual trigger bar, current deviation, trigger deviation, consecutive state days, and whether an up-cross has stayed above the thick line.
- The scan uses a bounded worker pool and supports cancellation. It returns only batched result snapshots and separate progress counters rather than cloning the complete result set for every stock.
- A SQLite `duo_kong_kline_cache` table stores normalized daily bars by code and date. During trading hours cache entries are refreshed after five minutes; after close the entry is valid until the next local trading day. Empty or failed fetches are never treated as valid cache entries.
- Default universe excludes ST names, suspended quotes, stocks listed less than sixty calendar days, and the BSE. Board, industry and concept filters are applied before K-line work. Price, free market cap, turnover, amount and volume ratio are based on the current quote snapshot and are shown with their update time.

### Holding ledger

- `trading_records` is the source of truth for position quantity and FIFO cost. `followed_stock.cost_price` and `volume` remain a convenience import for users who enter an existing position manually.
- A cost import creates or updates exactly one synthetic `AUTO_COST_IMPORT` buy record per followed stock. It never creates a second purchase merely because the user edits the cost on a later day.
- The holding summary and its AI tool use the FIFO ledger and include explicit flags when an imported cost record is being used.

### Startup and market news

- Schema migration runs before Wails starts. A version table records the last completed migration. Before a schema change the SQLite database is copied to `data/backups` with a retention limit.
- Market news stores both `DataTime` (published time) and `FetchedAt` (local ingestion time). Recent-news queries use `DataTime`; fallback source timestamps use `FetchedAt`.
- Startup news fetches run concurrently with per-source deadlines and persist a source health record containing last start, success, failure text, item count and finish time. The frontend can show this status.

### Distribution and updates

- The update server serves static release assets and a signed `manifest.json`. The manifest contains channel, semantic version, release notes, minimum supported version, published time, URL, byte size and SHA-256 for each platform asset.
- The application embeds only the Ed25519 public key. The private signing key remains off-server and is supplied to the local release command through an environment variable. HTTP transport is acceptable only because every manifest and payload digest is signature-verified; a custom domain with HTTPS can be enabled later without client design changes.
- Windows updates download a zip release, verify its signature and SHA-256, then launch a small external updater after the main process exits. The updater atomically backs up, replaces, restarts and rolls back when the new process does not acknowledge startup.
- macOS release packaging is corrected to the A-Share Desk artifact name. The application can check and display signed update metadata, while automatic replacement remains disabled until a signed and notarized macOS build can be verified on macOS hardware.

## Non-goals

- No broker integration or order placement.
- No attempt to silently bypass Windows SmartScreen or macOS Gatekeeper. Code signing and notarization remain separate credentials and release steps.
- No remote storage of the user database, API keys or update signing key.

## Verification

- Deterministic Go tests cover signal semantics, persistent cache freshness, scan cancellation and result ordering, FIFO summary and synthetic-cost idempotency, migrations, news timestamp/health handling, manifest signature/hash validation and semantic version comparisons.
- Node tests cover the indicator output and selector sorting helpers. Production frontend builds and a Windows NSIS package are produced after the full test suite.
