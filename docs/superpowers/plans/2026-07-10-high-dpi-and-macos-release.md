# High DPI Accessibility And macOS Release Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make A-Share Desk usable on Windows displays scaled to 300% and add a reproducible macOS universal-DMG release pipeline.

**Architecture:** The Go host calculates a large, bounded first-run window instead of using half of the logical screen. The Vue app owns the responsive shell: it reserves space for a compact bottom navigation, switches large-screen data surfaces to a deliberate narrow layout, and exposes an in-app readability scale. GitHub Actions builds the existing Wails app on a macOS runner and uploads a universal DMG artifact.

**Tech Stack:** Go 1.24, Wails v2, Vue 3, Naive UI, Vite, GitHub Actions, macOS/Xcode.

## Global Constraints

- Preserve the user's existing saved window size; only improve first-run sizing and clamp invalid restored dimensions.
- Treat 300% Windows scaling as a narrow logical viewport; do not reduce text to force a desktop layout.
- Keep primary actions visible; secondary navigation and table columns may move to overflow/details.
- The macOS build must run only on macOS and produce `A-Share Desk.app` inside a universal DMG.
- Do not publish unsigned production artifacts as a stable release; publish them as CI artifacts until Developer ID signing and notarization secrets exist.

---

### Task 1: Record Findings And Add Host-Sizing Tests

**Files:**
- Create: `window_size.go`
- Create: `window_size_test.go`
- Modify: `main.go`

**Interfaces:**
- Produces `resolveInitialWindowSize(screenWidth, screenHeight, savedWidth, savedHeight int) (int, int)`.
- `main.go` uses the helper before constructing `options.App`.

- [ ] **Step 1: Write failing Go tests**

```go
func TestResolveInitialWindowSizeUsesLargeFirstRunWindow(t *testing.T) {
    width, height := resolveInitialWindowSize(640, 360, 0, 0)
    if width != 640 || height != 360 {
        t.Fatalf("got %dx%d, want 640x360", width, height)
    }
}

func TestResolveInitialWindowSizeClampsSavedWindowToScreen(t *testing.T) {
    width, height := resolveInitialWindowSize(1280, 720, 2200, 1500)
    if width != 1280 || height != 720 {
        t.Fatalf("got %dx%d, want 1280x720", width, height)
    }
}
```

- [ ] **Step 2: Run the focused tests and observe the missing-helper failure**

Run: `go test . -run TestResolveInitialWindowSize -count=1`

- [ ] **Step 3: Implement the small pure sizing helper**

```go
func resolveInitialWindowSize(screenWidth, screenHeight, savedWidth, savedHeight int) (int, int) {
    if savedWidth > 0 && savedHeight > 0 {
        return min(savedWidth, screenWidth), min(savedHeight, screenHeight)
    }
    return screenWidth, screenHeight
}
```

- [ ] **Step 4: Replace the 50% first-run calculation in `main.go` and add Wails min-size bounds only where the current logical screen can support them.**

- [ ] **Step 5: Run `go test . -run TestResolveInitialWindowSize -count=1` and `go test ./... -run '^$' -count=0`.**

### Task 2: Implement The Responsive Application Shell

**Files:**
- Modify: `frontend/src/App.vue`
- Modify: `frontend/package.json`
- Create: `frontend/src/utils/layout.ts`
- Create: `frontend/src/utils/layout.test.mjs`

**Interfaces:**
- `getLayoutMode(width)` returns `compact`, `standard`, or `wide`.
- The root application adds `app-shell--compact` based on the active layout mode.
- User-selected readability scale is persisted in local storage under `a-share-desk:ui-scale`.

- [ ] **Step 1: Write a failing Node test for compact mode at a 640px logical viewport.**

```js
import assert from 'node:assert/strict'
import { getLayoutMode } from './layout.ts'

assert.equal(getLayoutMode(640), 'compact')
assert.equal(getLayoutMode(1000), 'standard')
assert.equal(getLayoutMode(1500), 'wide')
```

- [ ] **Step 2: Run `node frontend/src/utils/layout.test.mjs` and observe the import failure.**

- [ ] **Step 3: Implement the layout helper and update `App.vue`.**

The compact shell must reserve the actual navigation height, make navigation horizontally scrollable, apply a readable but bounded text scale, and replace `max-height: calc(92vh)` with a `calc(100dvh - var(--app-nav-height))` content region.

- [ ] **Step 4: Run the layout test, then `npm run build` in `frontend`.**

### Task 3: Make High-Value Data Surfaces Degrade Gracefully

**Files:**
- Modify: `frontend/src/components/DuoKongLineSelector.vue`
- Modify: `frontend/src/components/StockLightweightKlineChart.vue`
- Modify: relevant list components with fixed-height tables, beginning with `frontend/src/components/allStockInfoList.vue`

**Interfaces:**
- Compact data tables retain code, name, signal/status, latest price, and action columns.
- Secondary columns remain available in the existing detail/multi-period view rather than being silently lost.
- K-line dialogs use `min(100vw, ...)` / `min(100dvh, ...)` constraints and scroll their controls independently from the chart.

- [ ] **Step 1: Add a failing static layout regression check covering the compact selector/table class and K-line dialog bounds.**

- [ ] **Step 2: Apply focused responsive classes and CSS media queries; do not alter screening semantics or K-line calculations.**

- [ ] **Step 3: Build the frontend and inspect both 640px and 1440px layouts through the Wails dev server or browser screenshot tooling.**

### Task 4: Add macOS Universal-DMG CI Build

**Files:**
- Create: `.github/workflows/release-macos.yml`
- Modify: `scripts/build-macos.sh`
- Modify: `README.md`

**Interfaces:**
- Manual workflow dispatch accepts `version` and an optional `publish` flag.
- The `macos-14` runner executes `scripts/build-macos.sh darwin/universal yes`.
- The resulting DMG is uploaded as `A-Share-Desk-macos-universal`.

- [ ] **Step 1: Add a workflow syntax check fixture or validate YAML through Python before adding the workflow.**

- [ ] **Step 2: Implement the workflow with pinned macOS runner, Go/Node setup, `npm ci`, Wails installation, universal build, and artifact upload.**

- [ ] **Step 3: Update the macOS script to read the current product version and accept Developer ID signing only from `MACOS_SIGN_IDENTITY`; leave notarization disabled until Apple credentials are configured.**

- [ ] **Step 4: Validate YAML parsing locally and run the script's non-macOS guard to confirm it fails with a clear message on Windows.**

### Task 5: Audit Report And Release Verification

**Files:**
- Create: `docs/audit-2026-07-10.md`

- [ ] **Step 1: Record findings ranked P0/P1/P2 with evidence and a practical next action.**

- [ ] **Step 2: Run `go test ./... -count=1`, `npm run build`, `git diff --check`, and the focused layout/window tests.**

- [ ] **Step 3: Build the Windows portable package to ensure Wails packaging continues to work.**

- [ ] **Step 4: Report the produced Windows artifact and explain that the macOS DMG must be downloaded from the completed GitHub Actions artifact because this Windows machine cannot create a signed DMG.**
