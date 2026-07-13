# B 版 A 股工作台 UI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将现有 Vue/Wails 桌面应用完整迁移到顶部一级导航、左侧二级导航的 B 版信息架构，并保持现有业务能力、300% 缩放和发布链路可用。

**Architecture:** 以纯 JS 导航模型作为唯一配置源，应用壳只负责布局、路由映射和响应式抽屉，业务页面继续使用现有 Wails API。持仓中心与 AI 控制台使用薄包装组件，避免复制交易和设置保存逻辑。

**Tech Stack:** Vue 3、Vue Router 4、Naive UI、Vite 7、Node test、Wails v2、Go 1.26、SQLite/GORM。

## Global Constraints

- 正式修改目录固定为 `C:\Users\lxd\Desktop\A-Share-Desk-Replica-20260709\source`。
- 保留当前工作树中所有既有修改，不执行 reset、checkout、clean 或整文件回退。
- 保留现有路由名称和 Wails 事件兼容层。
- 不手工编辑 `frontend/wailsjs/**`，不修改数据库、行情、更新签名和安装包协议。
- 导航断点统一为 720、960、1280；正文不小于 12px；弹窗最大高度 88dvh。
- 每个行为变更先写失败测试并观察 RED，再实现 GREEN。
- 不创建提交，除非用户另行要求。

---

### Task 1: 单一导航模型与路由解析

**Files:**
- Create: `frontend/src/navigation/workspace.mjs`
- Create: `frontend/src/navigation/workspace.test.mjs`
- Modify: `frontend/src/utils/navigation.mjs`
- Modify: `frontend/src/utils/navigation.test.mjs`

**Interfaces:**
- Produces: `PRIMARY_SECTIONS`、`getSectionForRoute(route)`、`getSecondaryItems(sectionKey)`、`buildSecondaryLocation(sectionKey, itemKey)`、`getWorkspacePresentation(width)`。

- [ ] **Step 1: Write the failing tests**，断言七个一级板块顺序、市场 13 项、选股 3 项、AI 3 项和旧路由映射。
- [ ] **Step 2: Run RED**：`node --test frontend/src/navigation/workspace.test.mjs frontend/src/utils/navigation.test.mjs`，预期因模块或导出不存在失败。
- [ ] **Step 3: Implement the pure navigation model**，路由位置使用现有 name/query，不依赖 Vue 或 Wails。
- [ ] **Step 4: Run GREEN**，预期全部通过。

### Task 2: B 版设计令牌与应用壳组件

**Files:**
- Create: `frontend/src/styles/tokens.css`
- Create: `frontend/src/styles/base.css`
- Create: `frontend/src/components/shell/WorkspaceTopNavigation.vue`
- Rewrite: `frontend/src/components/shell/WorkspaceNavigation.vue`
- Rewrite: `frontend/src/components/shell/WorkspaceHeader.vue`
- Modify: `frontend/src/main.js`

**Interfaces:**
- Consumes: Task 1 navigation model.
- Produces: 顶部一级导航、当前板块二级导航、搜索工具栏和二级导航开关事件。

- [ ] **Step 1: Add failing source-contract tests**，检查七个标签、ARIA 文本、抽屉按钮和统一 CSS 变量引用。
- [ ] **Step 2: Run RED**，预期缺少新组件和令牌失败。
- [ ] **Step 3: Add tokens/base styles**：高度链、box-sizing、focus-visible、reduced-motion、滚动契约和 720/960/1280 断点。
- [ ] **Step 4: Implement shell components**，顶部导航使用 RouterLink，二级导航从当前 route 计算，不复制配置。
- [ ] **Step 5: Run GREEN and `npm run build`**。

### Task 3: 路由与 App 壳集成

**Files:**
- Modify: `frontend/src/router/router.js`
- Modify: `frontend/src/App.vue`
- Create: `frontend/src/components/HoldingCenter.vue`
- Create: `frontend/src/components/AIConsole.vue`

**Interfaces:**
- Adds route names: `holdingCenter`、`aiConsole`。
- Keeps all existing route names and paths available.

- [ ] **Step 1: Add failing route-source tests**，断言默认总览、独立持仓/AI 路由、现有兼容路由和懒加载边界。
- [ ] **Step 2: Run RED**。
- [ ] **Step 3: Add thin wrapper pages**：持仓承载 `TradingRecordManager`；AI 页按 query 承载分析记录、推荐记录和 AI 配置入口。
- [ ] **Step 4: Replace App template with B shell**，保留现有 Wails 生命周期、更新提示、搜索、浮动助手和事件监听。
- [ ] **Step 5: Remove only dead visible-menu rendering**，保留自选分组刷新所需逻辑直到 Task 5 完成。
- [ ] **Step 6: Run GREEN and production build**。

### Task 4: 市场、自选与选股路由闭环

**Files:**
- Modify: `frontend/src/components/market.vue`
- Modify: `frontend/src/components/stock.vue`
- Modify: `frontend/src/components/ScreeningHub.vue`
- Modify: `frontend/src/components/DecisionWorkbench.vue`
- Create: `frontend/src/navigation/tab-sync.mjs`
- Create: `frontend/src/navigation/tab-sync.test.mjs`

**Interfaces:**
- Produces: `normaliseMarketTab(query)`、`normaliseScreeningTab(query)`、`normaliseAIConsoleTab(query)`。

- [ ] **Step 1: Write failing tab normalisation tests**，覆盖空 query、非法值和合法深链。
- [ ] **Step 2: Run RED**。
- [ ] **Step 3: Implement normalisers and route watchers**，刷新或连续点击二级导航时同步页面内部 tab。
- [ ] **Step 4: Map screen tabs exactly**：`pattern -> allStockList`、`duo-kong -> DuoKongLineSelector`、`strategy -> SelectStock`。
- [ ] **Step 5: Update overview links** 到独立持仓中心、市场资讯和三个选股入口。
- [ ] **Step 6: Run GREEN and build**。

### Task 5: 持仓中心、AI 控制台与设置归属

**Files:**
- Modify: `frontend/src/components/HoldingCenter.vue`
- Modify: `frontend/src/components/AIConsole.vue`
- Modify: `frontend/src/components/settings.vue`
- Modify: `frontend/src/components/researchIndex.vue`

**Interfaces:**
- AI 配置入口使用 `settings?section=ai`，不复制整份 SettingConfig 保存逻辑。
- 旧 `research` 路由继续可用，但不显示提示词广场、问答广场、交易日志或选股入口。

- [ ] **Step 1: Write failing source-contract tests**，断言 AI 三项、持仓交易记录和设置 section 深链。
- [ ] **Step 2: Run RED**。
- [ ] **Step 3: Implement wrappers and section scroll/focus**，设置页接收 route query 并定位稳定 section ID。
- [ ] **Step 4: Trim old research aggregation**，只保留兼容所需内容，业务组件本身不删除。
- [ ] **Step 5: Run GREEN and build**。

### Task 6: 300% 缩放与高风险页面适配

**Files:**
- Modify: `frontend/src/components/stock.vue`
- Modify: `frontend/src/components/TradingRecordManager.vue`
- Modify: `frontend/src/components/settings.vue`
- Modify: `frontend/src/components/kline-analysis.vue`
- Modify: `frontend/src/components/KLineChart.vue`
- Modify: `window_size.go`
- Modify: `window_size_test.go`
- Modify: `app_windows.go`

**Interfaces:**
- Tables expose reachable horizontal scrolling.
- Modals fit `96vw x 88dvh`.
- Windows initial size uses work area, not full monitor bounds.

- [ ] **Step 1: Add failing boundary tests** for width presentation and Windows work-area clamping.
- [ ] **Step 2: Run RED**。
- [ ] **Step 3: Add `scroll-x`/responsive grids/modal bounds** to the named high-risk pages without rewriting business logic.
- [ ] **Step 4: Add ResizeObserver/dispose lifecycle** to legacy ECharts KLineChart.
- [ ] **Step 5: Fix Windows work-area sizing** with existing platform helpers.
- [ ] **Step 6: Run targeted JS/Go tests and build**。

### Task 7: Verification, visual review and release safety

**Files:**
- Modify: `frontend/package.json`
- Create: `docs/verification/b-workspace-ui.md`

**Interfaces:**
- Adds `npm test` executing all `.test.mjs` files.

- [ ] **Step 1: Run all JS tests** and record pass count.
- [ ] **Step 2: Run `npm run build`** and record bundle warning status.
- [ ] **Step 3: Run `go test ./... -run '^$' -count=0`** and targeted deterministic Go tests.
- [ ] **Step 4: Run `wails build -clean`** on Windows.
- [ ] **Step 5: Start the built app and inspect 1440x900、1024x768、800x600 and 300% equivalent viewports**，检查七个板块、全部二级入口、K 线非空和弹窗边界。
- [ ] **Step 6: Verify updater/portable/installer files remain untouched** and document macOS build as requiring native macOS runner.
- [ ] **Step 7: Request final code review and fix all Critical/Important findings**。

