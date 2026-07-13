# 持仓产品搜索与概览中文化实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 建立按资产类别搜索标准产品的新增持仓流程，并完成持仓概览中文化和字体优化。

**Architecture:** 新增独立 `assetsearch` 提供器层，Wails 暴露统一搜索方法；前端交易弹窗使用独立产品选择器消费标准结果。概览只消费中文映射后的显示模型，不直接显示内部枚举。

**Tech Stack:** Go 1.23、Wails v2、Vue 3、Naive UI、Node test、Go httptest。

## Global Constraints

- 银行理财保留手动输入和 CSV 导入。
- 其他资产必须先选择数据源结果，产品身份字段只读。
- 所有用户可见文案使用简体中文。
- 延续当前浅色/深色主题和 300% DPI 适配。

---

### Task 1: 统一资产搜索服务

**Files:**
- Create: `backend/provider/assetsearch/types.go`
- Create: `backend/provider/assetsearch/service.go`
- Create: `backend/provider/assetsearch/eastmoney.go`
- Create: `backend/provider/assetsearch/binance.go`
- Test: `backend/provider/assetsearch/service_test.go`
- Modify: `app_portfolio.go`

- [ ] 先写失败测试，覆盖东方财富分类过滤、币安交易对解析、空关键词和数量限制。
- [ ] 运行 `go test ./backend/provider/assetsearch -count=1` 确认测试因实现缺失失败。
- [ ] 实现 `Search(ctx, assetType, keyword, limit)` 和 Wails 方法 `SearchAssetInstruments`。
- [ ] 再次运行测试并确认通过。

### Task 2: 搜索式新增持仓弹窗

**Files:**
- Create: `frontend/src/components/portfolio/AssetProductSearch.vue`
- Modify: `frontend/src/components/portfolio/AssetTransactionEditor.vue`
- Modify: `frontend/src/components/portfolio/asset-form-schema.mjs`
- Modify: `frontend/src/components/HoldingCenter.vue`
- Modify: `frontend/src/components/portfolio/AssetPositionTable.vue`
- Modify: `frontend/src/components/TradingRecordManager.vue`
- Test: `frontend/src/components/portfolio-product-search.test.mjs`

- [ ] 先写失败合同测试，验证“增加一笔持仓”、产品搜索接口和银行理财手动输入分支。
- [ ] 运行单测确认失败。
- [ ] 实现搜索、选择、只读身份字段及已有持仓直达交易表单。
- [ ] 统一股票交易弹窗与按钮文案。
- [ ] 运行测试确认通过。

### Task 3: 概览中文化和字体清晰度

**Files:**
- Modify: `frontend/src/components/portfolio/PortfolioOverview.vue`
- Modify: `frontend/src/components/portfolio/asset-form-schema.mjs`
- Test: `frontend/src/components/portfolio-overview-localization.test.mjs`

- [ ] 先写失败测试，禁止 `bank_wealth`、`futures`、`buy/sell` 直接显示。
- [ ] 增加资产、动作、币种中文映射和零值过滤。
- [ ] 将币种卡片改为“币种分布（人民币折算）”并提升字号、字重、对比度。
- [ ] 运行测试确认通过。

### Task 4: 集中验证与发布

**Files:**
- Regenerate: `frontend/wailsjs/go/main/App.*`

- [ ] 运行前端全量测试和生产构建。
- [ ] 运行新增 Go 测试及相关数据层测试。
- [ ] 运行 `wails generate module` 与 `wails build -clean`。
- [ ] 检查 `git diff --check`、工作树和用户可见英文残留。
- [ ] 提交并推送 `codex/multi-asset-portfolio`，同步 `main`。
