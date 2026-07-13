# A股智投工作台稳定化与 GitHub-only 发布 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 消除自动弹网页、安全暴露和账本错误，完成启动优化、GitHub-only 更新及品牌图标替换。

**Architecture:** 保留 Wails + Go + Vue + SQLite 主体，把联网集成测试、后台任务、报价刷新和更新提供器拆成明确边界。业务计算向资产估值内核集中，桌面桥接层只负责 DTO、调度和系统能力。

**Tech Stack:** Go、GORM/SQLite、Wails v2、Vue 3、Vite、GitHub Actions、Ed25519、PowerShell/Bash。

## Global Constraints

- 默认测试不得打开真实浏览器或访问真实外网。
- 更新链路只能使用 HTTPS GitHub API、GitHub Releases 和已签名 Release 资产。
- 所有新增逻辑先写失败测试，再写最小实现。
- 不恢复赞助、VIP、设备绑定或上游社区依赖。
- 不破坏现有本地数据库和便携版数据目录。

---

### Task 1: 隔离真实浏览器和联网测试

**Files:** `backend/data/live_integration_test.go`、`backend/data/crawler_api_test.go`、`backend/data/openai_api_test.go`、`backend/data/stock_data_api_test.go`、`backend/data/web_search_api_test.go`

- [ ] 为 `liveIntegrationEnabled` 写环境变量测试。
- [ ] 实现 `requireLiveIntegration(t)`，默认调用 `t.Skip`。
- [ ] 给所有真实浏览器与真实网络用例添加显式门禁。
- [ ] 运行 `go test ./backend/data -count=1`，确认不启动浏览器。

### Task 2: 修复后端测试基线和不可达代码

**Files:** `backend/agent/agent_test.go`、`backend/agent/tools/data_tools_wrapper_test.go`、`app.go`

- [ ] 为 AI 工具测试创建独立内存数据库和必要种子。
- [ ] 删除 `CheckUpdate`、`syncNews`、`CheckStockBaseInfo` 中 return 后的旧实现。
- [ ] 运行 `go test ./...` 与 `go vet ./...`。

### Task 3: 收紧本地 AI Web 服务

**Files:** `ai-assistant-web/server.go`、`backend/data/ai_assistant_api.go`、`main.go` 及对应测试

- [ ] 测试默认地址只绑定 `127.0.0.1`。
- [ ] 测试非健康接口缺少令牌时返回 401。
- [ ] 加入随机令牌、严格 Origin、请求体上限、HTTP 超时与可关闭 Server。
- [ ] 将会话限制改为可控消息数和字节数。

### Task 4: 修正多资产账本

**Files:** `backend/data/asset_calculator.go`、`backend/data/portfolio_service.go`、对应测试

- [ ] 增加期货两次平仓现金流测试并修复累计值重复加入。
- [ ] 增加债券、贵金属、期货乘数与手续费最近交易金额测试。
- [ ] 统一交易本币金额函数，组合层复用该函数。
- [ ] 外币成本保留原始本币均价，汇率过期显式标记。

### Task 5: 并发刷新行情并优化启动

**Files:** `app_portfolio.go`、`backend/provider/assetquote/registry.go`、`app.go`、后台任务管理文件

- [ ] 用可注入报价器测试有限并发、顺序无关和单资产超时。
- [ ] 批量并发刷新持仓和主要汇率。
- [ ] 启动任务接入统一 context、singleflight 和状态记录。
- [ ] 股票基础数据按数据版本增量更新，禁止启动时先删表。

### Task 6: 前端首屏分包

**Files:** `frontend/src/router/router.js`、`frontend/vite.config.js`、前端测试

- [ ] 增加路由懒加载源码契约测试。
- [ ] 将非总览页面改为动态 import。
- [ ] 为 Vue、图表、Markdown/PDF 设定稳定 manual chunks。
- [ ] 构建并记录主包、首屏 CSS 和模块数量变化。

### Task 7: GitHub-only 更新

**Files:** `backend/provider/update`、`app_common.go`、`.github/workflows/a-share-update.yml`、发布脚本

- [ ] 用测试服务器模拟 GitHub latest release API 和 Release 资产。
- [ ] 新增 GitHub Release 更新提供器并保留签名/哈希验证。
- [ ] 删除服务器 Manifest URL、SCP/SSH 部署和服务器发布脚本入口。
- [ ] Release workflow 直接上传 Windows/macOS 资产与 manifest。

### Task 8: 新品牌图标与元数据

**Files:** `build/appicon.png`、`build/app.ico`、`build/windows/icon.ico`、macOS 资源、`wails.json`、`frontend/package.json`

- [ ] 生成并验证白蓝科技风应用图标。
- [ ] 生成多尺寸 PNG、ICO 和 ICNS 输入集。
- [ ] 统一应用名、作者、说明和版本来源。
- [ ] 检查 Windows 快捷方式与 macOS App 图标。

### Task 9: 全量验证与发布

**Files:** CI、README、SECURITY、NOTICE、发布文档

- [ ] 运行 Go 测试、vet、前端测试和生产构建。
- [ ] 运行关键页面及 300% 缩放截图验证。
- [ ] 构建 Windows 便携包并检查 GitHub 更新元数据。
- [ ] 更新安全策略、数据源说明、隐私与故障诊断文档。
- [ ] 提交并推送 `codex/project-hardening-github-only`。
