# Shared Stock K-line Modal Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Open the overview K-line action in the same reusable in-app multi-period K-line modal as pattern screening.

**Architecture:** Introduce one presentation-only modal wrapper around the existing multi-period chart. Keep selection and symbol normalization in each page while sharing modal structure and chart lifecycle.

**Tech Stack:** Vue 3 SFC, Naive UI modal, Node test runner, Vite.

## Global Constraints

- Do not change `StockLightweightKlineChart.vue` behavior.
- Do not navigate away from the overview when K-line is clicked.
- Keep the modal responsive at normal and high-DPI viewport sizes.

---

### Task 1: Add a regression contract

**Files:**
- Create: `frontend/src/components/overview-kline-modal.test.mjs`

**Interfaces:**
- Consumes: Vue component source files.
- Produces: A regression contract for shared modal usage and route removal.

- [ ] Write a failing source-contract test requiring `StockKlineModal` in the overview and pattern screener.
- [ ] Run `node --test src/components/overview-kline-modal.test.mjs` and confirm it fails because the shared component does not exist.

### Task 2: Implement the shared modal

**Files:**
- Create: `frontend/src/components/StockKlineModal.vue`
- Modify: `frontend/src/components/DecisionWorkbench.vue`
- Modify: `frontend/src/components/SelectStock.vue`

**Interfaces:**
- Consumes: `show`, `code`, `stockName`, `darkTheme`, and `chartHeight` props.
- Produces: `update:show` and the shared multi-period chart dialog.

- [ ] Wrap `StockLightweightKlineChart` in a responsive Naive UI card modal.
- [ ] Replace overview route navigation with modal state.
- [ ] Replace the pattern screener's local modal markup with the shared component.
- [ ] Run the focused test and confirm it passes.

### Task 3: Verify the release build

**Files:**
- Verify: `frontend/src/components/*.test.mjs`
- Verify: `frontend/dist`

**Interfaces:**
- Consumes: Completed Vue sources.
- Produces: Passing tests and a compiled frontend bundle.

- [ ] Run `npm test` and require zero failures.
- [ ] Run `npm run build` and require a successful Vite build.
