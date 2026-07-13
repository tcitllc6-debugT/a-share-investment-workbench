# Prism Atlas 白蓝双主题 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将桌面端改造成白色主界面、蓝色强调色的高级消费级 A 股工作台，提供稳定的日间/暗黑主题，恢复自选列表 AI 研判，并取消启动时的更新弹窗。

**Architecture:** 通过语义化 CSS token 让现有 Vue + Naive UI 页面共享白蓝主题，而不重写后端数据流。主题模式由一个小型前端工具模块负责解析、持久化及应用到 `<html>`；`App.vue` 将该状态映射给 Naive UI。自选页保留原有 AI 分析能力，但将“是否已配置”从可见性条件改为按钮状态与明确引导。

**Tech Stack:** Vue 3、Vite 7、Naive UI、Wails、Node 内置 test runner、Go/Wails Windows build。

## Global Constraints

- 只修改 `C:\Users\lxd\Desktop\A股智投工作台\source`，不得修改上游 `A-Share-Desk-Replica-20260709`。
- 视觉方向固定为日间 `#F5F8FC/#FFFFFF`、品牌蓝 `#3478F6`，暗黑为深蓝灰搭配柔和蓝，禁止沿用暗绿色作为主色。
- 日间与暗黑模式都必须以语义 token 实现，组件中不新增散落的主题色 hex 值。
- 启动不弹版本更新通知；“设置”中的手动检查更新能力保留。
- AI 研判操作必须始终可见。未配置 AI 时明确提示并跳转至 AI 配置，而不是静默隐藏。
- 支持高 DPI：工作区保留可滚动主内容与横向表格滚动，不用固定缩放比例掩盖内容。
- 每项行为先完成失败测试，再写最小实现；完成前运行全量前端测试、Vite 构建、Go 测试和 Wails Windows 构建。

---

## File Structure

- Create: `frontend/src/utils/theme-mode.mjs` - 主题模式的解析、持久化与 DOM 应用，保持纯函数可测试。
- Create: `frontend/src/utils/theme-mode.test.mjs` - 日间/暗黑持久化与配置回退的行为测试。
- Modify: `frontend/src/styles/tokens.css` - Prism Atlas 的双主题语义变量、排版、玻璃层与无障碍焦点 token。
- Modify: `frontend/src/styles/base.css` - 全局背景、滚动、交互反馈及减少动态效果的基础规则。
- Modify: `frontend/src/main.js` - 在应用启动前加载持久主题和 Prism Atlas 样式。
- Modify: `frontend/src/App.vue` - 将主题工具与 Naive UI 对接、移除启动更新通知、传递主题切换事件，并改造应用壳层。
- Modify: `frontend/src/components/shell/WorkspaceTopNavigation.vue` - 品牌、图标导航、日夜切换与高 DPI 自适应顶部栏。
- Modify: `frontend/src/components/shell/WorkspaceNavigation.vue` - 带图标的上下文导航与轻玻璃层级。
- Modify: `frontend/src/components/shell/WorkspaceHeader.vue` - 更清楚的页面上下文和操作区。
- Modify: `frontend/src/components/shell/workspace-shell.test.mjs` - 壳层主题切换、图标与可访问性契约。
- Modify: `frontend/src/components/stock.vue` - 常显的 AI 研判按钮、未配置引导、白蓝化自选表格操作条。
- Modify: `frontend/src/components/watchlist-direct-panel.test.mjs` - AI 操作常显与配置引导回归测试。
- Modify: `frontend/src/components/settings.vue` - 主题模式选项与显式“检查更新”入口，保持手动更新逻辑。
- Create: `frontend/src/components/prism-atlas-theme.test.mjs` - 对主要工作区使用 token、暗黑覆盖及减少动态效果的静态契约测试。
- Modify: `README.md` - 记录独立版运行入口、主题切换与 AI 未配置时的操作方式。

## Task 1: 建立可测试的主题模式工具

**Files:**
- Create: `frontend/src/utils/theme-mode.test.mjs`
- Create: `frontend/src/utils/theme-mode.mjs`

**Interfaces:**
- Produces: `THEME_MODE_STORAGE_KEY`, `resolveThemeMode(preferredMode, storedMode)`, `persistThemeMode(mode, storage)`, `applyThemeMode(mode, document)`。
- Consumes: `preferredMode` 为 `'light' | 'dark' | undefined`；返回值始终为 `'light' | 'dark'`。

- [ ] **Step 1: Write the failing test**

```js
import assert from 'node:assert/strict'
import test from 'node:test'
import { applyThemeMode, persistThemeMode, resolveThemeMode } from './theme-mode.mjs'

test('stored mode wins over the legacy config preference', () => {
  assert.equal(resolveThemeMode('dark', 'light'), 'light')
  assert.equal(resolveThemeMode('light', 'dark'), 'dark')
  assert.equal(resolveThemeMode('dark', null), 'dark')
})

test('persist and apply write the Prism Atlas theme marker', () => {
  const writes = []
  const storage = { setItem: (...args) => writes.push(args) }
  const root = { dataset: {}, setAttribute: (key, value) => { root[key] = value } }
  persistThemeMode('dark', storage)
  applyThemeMode('dark', { documentElement: root })
  assert.deepEqual(writes, [['a-share-theme-mode', 'dark']])
  assert.equal(root.dataset.themeMode, 'dark')
  assert.equal(root['data-theme-mode'], 'dark')
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `npm test -- src/utils/theme-mode.test.mjs`

Expected: FAIL because `theme-mode.mjs` does not exist.

- [ ] **Step 3: Write minimal implementation**

```js
export const THEME_MODE_STORAGE_KEY = 'a-share-theme-mode'
const MODES = new Set(['light', 'dark'])

export function resolveThemeMode(preferredMode, storedMode) {
  return MODES.has(storedMode) ? storedMode : preferredMode === 'dark' ? 'dark' : 'light'
}

export function persistThemeMode(mode, storage = globalThis.localStorage) {
  storage?.setItem?.(THEME_MODE_STORAGE_KEY, resolveThemeMode(mode))
}

export function applyThemeMode(mode, documentRef = globalThis.document) {
  const resolved = resolveThemeMode(mode)
  const root = documentRef?.documentElement
  if (root) {
    root.dataset.themeMode = resolved
    root.setAttribute('data-theme-mode', resolved)
  }
  return resolved
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `npm test -- src/utils/theme-mode.test.mjs`

Expected: PASS with 2 passing subtests.

- [ ] **Step 5: Commit**

The standalone copy has no Git metadata. Record the completed task in this plan instead of committing.

## Task 2: 将 Prism Atlas token 和基础交互接入全局样式

**Files:**
- Create: `frontend/src/components/prism-atlas-theme.test.mjs`
- Modify: `frontend/src/styles/tokens.css`
- Modify: `frontend/src/styles/base.css`
- Modify: `frontend/src/main.js`

**Interfaces:**
- Consumes: `<html data-theme-mode="light|dark">` produced by Task 1.
- Produces: `--workspace-*` token 在两种主题中均可用；`main.js` 在 mount 前应用持久模式。

- [ ] **Step 1: Write the failing test**

```js
import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'

const root = new URL('../', import.meta.url)
const read = (name) => readFileSync(new URL(name, root), 'utf8')

test('Prism Atlas tokens define paired white-blue and midnight themes', () => {
  const tokens = read('styles/tokens.css')
  const base = read('styles/base.css')
  assert.match(tokens, /:root\[data-theme-mode="light"\]/)
  assert.match(tokens, /:root\[data-theme-mode="dark"\]/)
  assert.match(tokens, /--workspace-accent:\s*#3478f6/i)
  assert.match(tokens, /--workspace-glass/)
  assert.match(base, /prefers-reduced-motion/)
  assert.match(base, /scrollbar-color/)
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `npm test -- src/components/prism-atlas-theme.test.mjs`

Expected: FAIL because paired selectors and glass tokens are absent.

- [ ] **Step 3: Write minimal implementation**

Use these mandatory token values and never hardcode them in pages:

```css
:root[data-theme-mode="light"] {
  --workspace-canvas: #f5f8fc;
  --workspace-surface: rgba(255, 255, 255, .86);
  --workspace-surface-solid: #ffffff;
  --workspace-text: #162033;
  --workspace-text-muted: #64748b;
  --workspace-border: rgba(148, 163, 184, .24);
  --workspace-accent: #3478f6;
  --workspace-accent-soft: #e9f1ff;
  --workspace-glass: blur(20px) saturate(150%);
}
:root[data-theme-mode="dark"] {
  --workspace-canvas: #0b1020;
  --workspace-surface: rgba(19, 28, 49, .82);
  --workspace-surface-solid: #131c31;
  --workspace-text: #edf4ff;
  --workspace-text-muted: #a6b6cf;
  --workspace-border: rgba(166, 187, 222, .17);
  --workspace-accent: #77aaff;
  --workspace-accent-soft: rgba(72, 130, 255, .18);
  --workspace-glass: blur(22px) saturate(135%);
}
```

In `main.js`, import `applyThemeMode`, read `localStorage.getItem('a-share-theme-mode')`, and call `applyThemeMode(storedMode)` before `app.mount`.

- [ ] **Step 4: Run test to verify it passes**

Run: `npm test -- src/components/prism-atlas-theme.test.mjs`

Expected: PASS.

- [ ] **Step 5: Run focused build sanity check**

Run: `npm run build`

Expected: Vite completes with exit code 0.

## Task 3: 应用壳层与导航改造并去除启动更新推送

**Files:**
- Modify: `frontend/src/App.vue`
- Modify: `frontend/src/components/shell/WorkspaceTopNavigation.vue`
- Modify: `frontend/src/components/shell/WorkspaceNavigation.vue`
- Modify: `frontend/src/components/shell/WorkspaceHeader.vue`
- Modify: `frontend/src/components/shell/workspace-shell.test.mjs`

**Interfaces:**
- Consumes: Task 1 的 `resolveThemeMode`, `persistThemeMode`, `applyThemeMode`。
- Produces: `WorkspaceTopNavigation` emits `theme-change(mode)` and `App.vue` maps it to Naive UI `darkTheme | null`.
- Produces: `updateAvailable` event is drained without creating `notification.create`; manual update bindings remain available to settings.

- [ ] **Step 1: Write the failing test**

Append these assertions to `workspace-shell.test.mjs`:

```js
const app = read('../../App.vue')
assert.match(topNavigation, /theme-change/)
assert.match(topNavigation, /aria-label="切换日间或暗黑主题"/)
assert.match(topNavigation, /SunnyOutline|MoonOutline/)
assert.match(secondaryNavigation, /NIcon|n-icon/)
assert.match(app, /@theme-change="setThemeMode"/)
assert.match(app, /EventsOn\("updateAvailable", \(\) => \{\}\)/)
assert.doesNotMatch(app, /发现新版本/)
assert.match(app, /workspace-shell--prism/)
```

- [ ] **Step 2: Run test to verify it fails**

Run: `npm test -- src/components/shell/workspace-shell.test.mjs`

Expected: FAIL because no theme event and startup update notification remains.

- [ ] **Step 3: Write minimal implementation**

In `App.vue`:

```js
const themeMode = ref('light')
function setThemeMode(mode) {
  themeMode.value = applyThemeMode(mode)
  persistThemeMode(themeMode.value)
  enableDarkTheme.value = themeMode.value === 'dark' ? darkTheme : null
}

GetConfig().then((res) => {
  setThemeMode(resolveThemeMode(res.darkTheme ? 'dark' : 'light', localStorage.getItem(THEME_MODE_STORAGE_KEY)))
})

EventsOn('updateAvailable', () => {})
```

Delete `pendingUpdate`, `updateNotificationHandler`, the `createDiscreteApi` update-notification block and the `ApplyManagedUpdate` import from `App.vue`; do not delete settings-page manual update code. Add `workspace-shell--prism` to the shell class. Top navigation must show one accessible icon button that emits the next mode. Navigation links need a consistent `NIcon` glyph passed by the workspace navigation mapping. Use `--workspace-glass` for shell/header/nav backdrop effects and retain a clear active state with text plus blue indicator.

- [ ] **Step 4: Run test to verify it passes**

Run: `npm test -- src/components/shell/workspace-shell.test.mjs`

Expected: PASS.

- [ ] **Step 5: Run focused integration build**

Run: `npm run build`

Expected: Vite completes with exit code 0 and no unresolved import.

## Task 4: 恢复始终可见的自选 AI 研判并提供未配置引导

**Files:**
- Modify: `frontend/src/components/watchlist-direct-panel.test.mjs`
- Modify: `frontend/src/components/stock.vue`

**Interfaces:**
- Produces: `openAiAnalysis(stockName, stockCode)`; it invokes `aiCheckStock` when `data.openAiEnable` is true, otherwise emits a clear message and routes to `name: 'settings', query: { section: 'ai' }`.
- Consumes: existing `aiCheckStock`, `message`, and Vue router already used by `stock.vue`.

- [ ] **Step 1: Write the failing test**

Append these tests:

```js
test('keeps AI research visible in the direct watchlist regardless of configuration', () => {
  const directPanel = source.slice(source.indexOf('class="watchlist-all-panel"'))
  assert.match(directPanel, /@click="openAiAnalysis\(result\['股票名称'\],result\['股票代码'\]\)"/)
  assert.doesNotMatch(directPanel, /v-if="data\.openAiEnable"[^>]*>AI分析/)
  assert.match(source, /function openAiAnalysis\(stock, stockCode\)/)
  assert.match(source, /请先在 AI 控制台配置模型服务/)
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `npm test -- src/components/watchlist-direct-panel.test.mjs`

Expected: FAIL because direct-panel AI action is conditionally hidden.

- [ ] **Step 3: Write minimal implementation**

```js
function openAiAnalysis(stock, stockCode) {
  if (data.openAiEnable) {
    aiCheckStock(stock, stockCode)
    return
  }
  message.info('请先在 AI 控制台配置模型服务，再开始 AI 研判')
  router.push({ name: 'settings', query: { section: 'ai' } })
}
```

Replace every self-watchlist AI button `v-if="data.openAiEnable"` with a persistent `AI研判` button bound to `openAiAnalysis`. Its `type` and tooltip must express either “开始 AI 研判” or “需先配置 AI 服务”. When configuration is loaded, assign `data.openAiEnable = Boolean(result.openAiEnable)` so stale truthy state cannot persist.

- [ ] **Step 4: Run test to verify it passes**

Run: `npm test -- src/components/watchlist-direct-panel.test.mjs`

Expected: PASS with all action-strip tests green.

- [ ] **Step 5: Run page regression suite**

Run: `npm test -- src/components/watchlist-direct-panel.test.mjs src/components/workspace-high-dpi.test.mjs`

Expected: PASS.

## Task 5: 设置页的主题选择和手动更新入口

**Files:**
- Modify: `frontend/src/components/settings.vue`
- Modify: `frontend/src/components/prism-atlas-theme.test.mjs`

**Interfaces:**
- Consumes: `CheckManagedUpdate` / `ApplyManagedUpdate` only from settings; startup path does not surface them.
- Produces: a named manual action `检查更新` and a theme UI that explains day/night selection but has no automatic update prompt.

- [ ] **Step 1: Write the failing test**

```js
test('settings keeps update checks manual and exposes both theme choices', () => {
  const settings = read('components/settings.vue')
  assert.match(settings, /检查更新/)
  assert.match(settings, /CheckManagedUpdate/)
  assert.match(settings, /日间主题|暗黑主题/)
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `npm test -- src/components/prism-atlas-theme.test.mjs`

Expected: FAIL until the explicit button and user-facing theme language exist.

- [ ] **Step 3: Write minimal implementation**

Keep current saved `darkTheme` compatibility, add an explicit card-level action:

```vue
<n-button secondary type="primary" :loading="checkingUpdate" @click="checkForUpdate">
  检查更新
</n-button>
```

`checkForUpdate` calls `CheckManagedUpdate()` and only then presents the result in the settings page. Present the theme control as segmented choices or a labelled switch with “日间主题 / 暗黑主题”; save to the existing config and invoke the same DOM theme event used by the application shell.

- [ ] **Step 4: Run test to verify it passes**

Run: `npm test -- src/components/prism-atlas-theme.test.mjs`

Expected: PASS.

- [ ] **Step 5: Run Vite build**

Run: `npm run build`

Expected: exit code 0.

## Task 6: 端到端验证、桌面构建和独立发布物更新

**Files:**
- Modify: `README.md`
- Build output: `C:\Users\lxd\Desktop\A股智投工作台\release\A股智投工作台.exe`

**Interfaces:**
- Consumes: all preceding Vue changes and existing Wails configuration.
- Produces: independent Windows executable and documented launch path.

- [ ] **Step 1: Write the release-readiness source assertion**

Add the README section before building:

```markdown
## 本地启动

Windows 独立版入口：`release/A股智投工作台.exe`。首次运行后可在“设置”中切换日间/暗黑主题、手动检查更新；自选列表的“AI研判”在未配置模型时会直接带你前往 AI 配置。
```

- [ ] **Step 2: Run full frontend tests**

Run: `npm test`

Expected: all Node test subtests pass.

- [ ] **Step 3: Run frontend production build**

Run: `npm run build`

Expected: Vite production bundle succeeds.

- [ ] **Step 4: Run Go tests**

Run: `go test ./...`

Expected: all packages pass.

- [ ] **Step 5: Build Wails Windows release and copy it**

Run: `wails build -platform windows/amd64 -o A股智投工作台.exe`

Then copy the verified binary from `build/bin/A股智投工作台.exe` to `C:\Users\lxd\Desktop\A股智投工作台\release\A股智投工作台.exe`.

Expected: binary exists in both locations and has a nonzero length.

- [ ] **Step 6: Desktop validation**

Launch only the target release executable. Verify these exact user-visible behaviors:

1. 日间启动为白底蓝色强调，暗黑切换后导航、表格、文本与边界仍清晰。
2. 不出现“发现新版本”启动通知。
3. 自选列表每条都包含 `AI研判`；没有 AI 配置时点击给出配置引导而不消失。
4. 缩小窗口后主内容可纵向滚动，宽表格可横向滚动。
5. “设置”中可找到 `检查更新`，但不会在启动时自动弹更新卡片。

- [ ] **Step 7: Record verification evidence**

Append exact test/build commands, timestamps, and observed desktop results to this plan under `## Verification Evidence` before claiming completion.

## Self-Review

1. **Spec coverage:** Theme system (Tasks 1-3), navigation/shell redesign (Task 3), AI visibility (Task 4), update popup removal plus manual update (Tasks 3 and 5), high DPI behavior (Tasks 2, 4 and 6), independent packaging (Task 6) are covered.
2. **Placeholder scan:** The plan contains no TBD/TODO/“implement later” placeholders. Concrete paths, APIs, assertions and commands are provided.
3. **Type consistency:** `themeMode`, `setThemeMode`, `resolveThemeMode`, `persistThemeMode`, `applyThemeMode`, and `openAiAnalysis` are named consistently through their producing and consuming tasks.

## Execution Approach

The user authorized execution in this session. Execute inline with `superpowers:executing-plans`, keeping a review checkpoint after every task.

## Verification Evidence

- 2026-07-11：`npm test` 通过，40 项测试、0 失败。覆盖主题解析与持久化、Prism Atlas 双主题、启动更新提示移除、自选 AI 研判常显、多空线、高 DPI 与导航契约。
- 2026-07-11：`npm run build` 通过，Vite 生产构建成功。构建器仅报告既有大包体积警告，不影响产物生成。
- 2026-07-11：`go test ./backend/data -run TestInvestCalendar -count=1` 通过；`InvestCalendar` 对异常外部响应增加空值保护。
- 2026-07-11：完整 `go test ./backend/data -count=1` 未作为发布门禁通过，因为上游遗留套件包含真实浏览器、占位机器人 URL 和外部 AI/行情接口，10 分钟内超时；该限制已如实记录，未将其误报为通过。
- 2026-07-11：`wails build -clean -platform windows/amd64 -o "A股智投工作台.exe"` 通过，生成 `build/bin/A股智投工作台.exe`。
- 桌面实测：暗黑主题所有总览面板正确跟随；切换日间主题后恢复白色画布与蓝色强调；启动时未出现更新卡片。
- 桌面实测：搜索并关注 `300418` 后立即出现昆仑万维；操作条包含多周期、日K、分时、资金、详情、公告、研报、成本、AI研判、设置分组、取消关注。未配置 AI 时按钮仍显示并提示“需先配置 AI 服务”。
- 桌面实测：设置页同时显示日间主题、暗黑主题与手动检查更新，并明确说明不再启动弹出更新提示。
- 高 DPI：`workspace-high-dpi.test.mjs` 与布局测试通过；工作区保留纵向滚动，宽表格保留横向滚动。
- 最终发布冒烟：直接启动 `release/A股智投工作台.exe` 成功，窗口可响应；紧凑窗口下出现纵向滚动条，工作台内容没有被裁掉，也没有启动更新卡片。
- 最终产物：源码构建与发布目录 exe 均为 173,137,920 字节，SHA-256 均为 `9F3FC84632F3D50676D3BAF74E841D37D8A515A5FE94BC4143C3755670A62214`。
