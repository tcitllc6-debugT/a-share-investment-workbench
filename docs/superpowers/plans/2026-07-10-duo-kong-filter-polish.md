# Duo Kong Filter Polish Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver a compact Duo Kong selector, clean concept options, and a settings-page signed update check.

**Architecture:** Normalize stored concept strings in the Go data API so every caller receives clean names. Keep the existing selector scan contract, removing unused UI fields from its submitted condition. Reuse the managed-update API from settings rather than legacy GitHub update code.

**Tech Stack:** Go, GORM/SQLite, Vue 3, Naive UI, Wails.

## Global Constraints

- Signal calculation remains fixed to 120 daily bars and a 3-trading-day signal window.
- The managed manifest is the only update source for the new settings action.
- Do not expose update signing material or server credentials.

---

### Task 1: Normalize concept options

**Files:**
- Modify: `backend/data/stock_data_api.go`
- Test: `backend/data/stock_data_api_test.go`

- [ ] Add a failing test for JSON-array concept values, plain values, blanks and duplicates.
- [ ] Implement parsing, trimming, deduplication and stable sorting in `GetAllConcepts`.
- [ ] Run `go test ./backend/data -run TestNormalizeConceptValues -count=1`.

### Task 2: Compact selector layout

**Files:**
- Modify: `frontend/src/components/DuoKongLineSelector.vue`

- [ ] Remove numeric market metric controls and their serialized condition fields.
- [ ] Render all eight signal buttons in one visual row without group headings.
- [ ] Render board, industry, concept and minimum deviation in one responsive row.
- [ ] Remove the result table free-market-capitalization column.

### Task 3: Add settings update action

**Files:**
- Modify: `frontend/src/components/settings.vue`

- [ ] Import `CheckManagedUpdate` and `ApplyManagedUpdate`.
- [ ] Add a settings action that reports latest state, version and notes, then starts the existing verified update flow on confirmation.
- [ ] Verify the front-end build succeeds.
