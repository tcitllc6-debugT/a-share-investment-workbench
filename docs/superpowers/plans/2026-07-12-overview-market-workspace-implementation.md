# 总览与市场工作区优化实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 修复总览和市场快讯首屏数据问题，重构工作区信息层级与导航，并输出外部项目功能吸收清单。

**Architecture:** 使用独立异步资源状态替代整页 `Promise.all`；通过纯函数规范化 Wails 自选数据；市场页按当前路由懒加载功能数据。页面切换完全由工作区侧栏和路由 query 驱动。

**Tech Stack:** Vue 3、Naive UI、Wails、Go、GORM、SQLite、Node test runner。

## Global Constraints

- 仅修改 `C:\Users\lxd\Desktop\A股智投工作台\source`。
- 白色为主、蓝色为辅，同时保持暗黑主题。
- 页面标题和副标题必须通俗、单行、简短。
- 不复制 `github-a-share-projects` 中的代码。
- 先写失败测试，再实现，再运行完整验证。

### Task 1: 总览数据模型与独立加载

**Files:**
- Create: `frontend/src/components/decision-workbench.mjs`
- Create: `frontend/src/components/decision-workbench.test.mjs`
- Modify: `frontend/src/components/DecisionWorkbench.vue`

- [ ] 为 Wails 大写字段、自选保底展示、资讯字段和市场温度编写失败测试。
- [ ] 实现规范化纯函数与独立资源加载，不允许单项失败清空整页。
- [ ] 增加市场温度和明确的空状态、局部刷新状态。
- [ ] 运行聚焦测试。

### Task 2: 市场快讯首屏性能

**Files:**
- Create: `frontend/src/components/market-first-paint.test.mjs`
- Modify: `frontend/src/components/market.vue`
- Modify: `backend/data/market_news_api.go`
- Modify: `app.go`

- [ ] 用源码契约测试锁定快讯首屏不得加载全球指数和行业排名。
- [ ] 提供轻量本地快讯读取接口，消除标签 N+1 查询。
- [ ] 按路由标签懒加载指数、行业、提示词和 AI 配置。
- [ ] 延迟热词分析，后台同步事件到达后增量更新。

### Task 3: 文案、导航与布局

**Files:**
- Create: `frontend/src/components/workspace-copy-layout.test.mjs`
- Modify: `frontend/src/components/DecisionWorkbench.vue`
- Modify: `frontend/src/components/ScreeningHub.vue`
- Modify: `frontend/src/components/AIConsole.vue`
- Modify: `frontend/src/components/stock.vue`
- Modify: `frontend/src/App.vue`

- [ ] 锁定三处新副标题、A 股状态和无页内重复标签。
- [ ] 让选股和 AI 控制台由路由直接渲染当前页面。
- [ ] 对齐自选搜索、分组和数量工具栏。
- [ ] 完成紧凑窗口与 300% 缩放适配。

### Task 4: 外部项目功能审计

**Files:**
- Create: `docs/research/github-a-share-projects-feature-audit.md`

- [ ] 逐个识别 16 个项目的核心能力、依赖和许可证。
- [ ] 按近期可做、中期可做、暂不建议分级。
- [ ] 为每个建议标注落入本软件的页面、数据依赖和实现边界。

### Task 5: 验证与发布

- [ ] 运行 `npm test`。
- [ ] 运行 Go 聚焦测试和 `go test ./... -run '^$' -count=1`。
- [ ] 运行 `wails build`，复制到 `release` 并校验哈希。
- [ ] 桌面验证总览、自选、市场快讯、侧栏切换、A 股状态和高 DPI。

## Self-Review

计划覆盖用户 14 项要求：性能、自选、总览补充、文案、工具栏对齐、移除重复标签、A 股状态、外部项目审计、测试和发布均有明确任务与证据。
