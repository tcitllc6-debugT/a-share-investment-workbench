# UI Trim, Current Hot Fix, And Windows Package Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove low-value community/about entry points, keep and explain AI stock recommendation records, fix the current-hot white popup behavior, and produce a Windows share package.

**Architecture:** Keep core local A-share research features and AI recommendation persistence. Remove only UI routes/tabs for prompt plaza, Q&A plaza, and About so backend-compatible APIs remain harmless. Fix current-hot external-topic navigation at the component boundary by opening the system browser instead of spawning an in-app WebView popup.

**Tech Stack:** Vue 3, Vue Router, Naive UI, Wails runtime, Go, PowerShell packaging.

## Global Constraints

- Work only in `C:/Users/lxd/Desktop/A-Share-Desk-Replica-20260709/source`.
- Do not modify the upstream source folder under `C:/Users/lxd/Desktop/gupiao`.
- Keep `股票推荐记录` because it is created by AI recommendation tools and used by price alert monitoring.
- Remove prompt plaza, Q&A plaza, and About from the visible app UI.
- Do not restore original-author hosted plaza/community dependencies.
- Package for Windows users as a zip that contains the exe and a short Chinese readme.

---

### Task 1: Remove Low-Value UI Entry Points

**Files:**
- Modify: `frontend/src/App.vue`
- Modify: `frontend/src/components/researchIndex.vue`
- Modify: `frontend/src/components/promptTemplateList.vue`
- Modify: `frontend/src/components/settings.vue`
- Modify: `frontend/src/router/router.js`

**Interfaces:**
- Consumes: menu option keys `promptPlaza`, `promptQa`, `about`.
- Produces: visible navigation without prompt plaza, Q&A plaza, or About.

- [ ] **Step 1: Remove App menu children**

Remove the two research children with keys `promptPlaza` and `promptQa`.

- [ ] **Step 2: Remove About menu item**

Remove the menu option with key `about`.

- [ ] **Step 3: Remove unused imports and tab panes**

In `researchIndex.vue`, remove `PromptPlaza` and `PromptQa` imports and their two `<n-tab-pane>` blocks.

- [ ] **Step 4: Remove About route import and route**

In `router.js`, remove `aboutView` import and the `/about` route.

- [ ] **Step 5: Remove remaining prompt-plaza visible controls**

In `promptTemplateList.vue`, remove remote share buttons, share modal, and share logic. In `settings.vue`, remove the prompt plaza API address form item.

- [ ] **Step 6: Verify strings disappear from visible UI files**

Run:

```powershell
rg -n "提示词广场|问答广场|promptPlaza|promptQa|name: 'about'|key: 'about'" frontend/src
```

Expected: no active UI entry points remain; component files may still exist if not imported.

### Task 2: Explain Stock Recommendation Records

**Files:**
- Modify: `docs/replica/research-center-and-distribution.md`

**Interfaces:**
- Consumes: existing `AiRecommendStocks` model/service/tool flow.
- Produces: documentation explaining why `股票推荐记录` remains.

- [ ] **Step 1: Add explanation**

Add a short paragraph saying recommendation records are AI-created structured stock picks, preserving model name, prompts, recommendation reason, risk level, and alert settings.

- [ ] **Step 2: Add deletion decision**

State that the feature is kept because it supports review, auditing, and price alert monitoring.

### Task 3: Fix Current Hot White Popup

**Files:**
- Modify: `frontend/src/components/HotTopics.vue`

**Interfaces:**
- Consumes: `OpenURL(url string)` Wails method.
- Produces: `showPage(htid)` opens Eastmoney topic pages in the system browser.

- [ ] **Step 1: Remove WebView popup path**

Delete `window.open()` usage and `Environment` import from `HotTopics.vue`.

- [ ] **Step 2: Validate topic URL**

Only call `OpenURL` when `htid` is non-empty.

- [ ] **Step 3: Verify no popup call remains**

Run:

```powershell
rg -n "window.open|Environment" frontend/src/components/HotTopics.vue
```

Expected: no matches.

### Task 4: Build And Package Windows Share Zip

**Files:**
- Create: `C:/Users/lxd/Desktop/A-Share-Desk-Replica-20260709/release/README-Windows.txt`
- Create: `C:/Users/lxd/Desktop/A-Share-Desk-Replica-20260709/release/a-share-desk-windows-20260709.zip`

**Interfaces:**
- Consumes: `build/bin/a-share-desk.exe`.
- Produces: shareable zip with exe and instructions.

- [ ] **Step 1: Run verification**

Run:

```powershell
go test ./... -run "^$" -count=0
npm --prefix frontend run build
wails build
```

- [ ] **Step 2: Copy executable with versioned name**

Copy `build/bin/a-share-desk.exe` to `release/a-share-desk-20260709-ui-trim.exe`.

- [ ] **Step 3: Create Windows readme**

Write a short readme explaining: unzip, run exe, Windows 10/11 recommended, install Microsoft Edge WebView2 Runtime if the app cannot open.

- [ ] **Step 4: Compress package**

Create `release/a-share-desk-windows-20260709.zip` containing the versioned exe and readme.

- [ ] **Step 5: Commit scoped changes**

Commit code/docs changes, not release zip/exe artifacts unless the repo already tracks release artifacts.
