# AI 配置中心实施计划

**Goal:** 将 AI 配置迁移为 AI 控制台内独立、通俗、安全且可诊断的配置中心，并从普通设置页移除。

**Architecture:** 后端提供 AI 专用 DTO、事务保存、脱敏列表和连接诊断；前端用独立 AIConfiguration.vue 管理配置，AIConsole.vue 只在三个内部视图间切换。普通设置改用通用设置保存接口，不再读写 AI 领域。

**Tech Stack:** Go、GORM、SQLite、Vue 3、Naive UI、Wails、Node test runner。

## Global Constraints

- 只修改 C:\Users\lxd\Desktop\A股智投工作台\source。
- 新页面沿用 Prism Atlas 白蓝双主题。
- API 密钥读取时不得回传明文；空密钥保存已有配置表示保留。
- 普通设置保存不得覆盖 AI 配置。
- 所有行为测试先失败，再实现。

## Task 1: AI 配置领域模型与事务保存

- 修改 backend/data/settings_api.go。
- 新建 backend/data/ai_workspace_api_test.go。
- 先写失败测试覆盖密钥脱敏、空密钥保留、默认唯一性和普通设置不被改写。
- 实现 GetAIWorkspaceConfig、UpdateAIWorkspaceConfig、SanitizedAIConfigs。
- 运行 Go 聚焦测试。

## Task 2: Wails 接口与连接诊断

- 修改 app.go；新建 backend/data/ai_connection.go 与测试。
- 实现 GetAIWorkspaceConfig、SaveAIWorkspaceConfig、TestAIConnection。
- 用本地 HTTP 测试服务器覆盖成功、401、404 和超时。
- FetchAiModels 复用诊断逻辑，GetAiConfigs 返回脱敏数据。

## Task 3: 前端配置模型与路由

- 新建 ai-configuration.mjs 及测试。
- 修改 workspace.mjs、导航测试、路由测试和 stock.vue。
- AI 配置入口统一指向 aiConsole?view=configuration。
- 实现服务商预设、默认配置、校验和诊断结果语义。

## Task 4: 独立 AI 配置中心 UI

- 新建 AIConfiguration.vue 与源码契约测试。
- 修改 AIConsole.vue 增加第三个视图。
- 覆盖状态栏、模型服务、模型发现、连接测试、默认用途、高级设置、安全提示、空状态和错误状态。
- 适配日间/暗黑和紧凑窗口。

## Task 5: 从普通设置迁出

- 新建 settings-ai-separation.test.mjs。
- 增加 UpdateGeneralConfig，普通设置不再更新 AI 列和 ai_config 表。
- settings.vue 删除 AI 区域及相关逻辑；飞书模型下拉改读脱敏列表。
- 运行前后端聚焦测试。

## Task 6: 默认模型与发布验证

- 修改 openai_api.go 和 agent 默认配置解析。
- 测试默认研究/智能体模型和旧数据回退。
- 运行 npm test、npm run build、Go 聚焦测试与全包编译。
- Wails 构建、复制 release、校验 SHA-256。
- 桌面验证独立配置页、设置迁出、连接诊断、双主题与高 DPI。

## Self-Review

规格中的独立路由、专用保存、密钥脱敏、默认用途、连接诊断、设置迁出和发布验证均有对应任务；无占位项，接口命名一致。

## Completion Evidence

- 前端测试：`npm test`，51/51 通过。
- 后端聚焦测试：AI 工作区、密钥边界、连接诊断、默认用途与飞书动态默认模型均通过。
- 全包编译：`go test ./... -run '^$' -count=1` 通过。
- 桌面验证：AI 配置仅在 AI 控制台内展示；普通设置页无 AI 配置；模型发现和连接测试可用；紧凑窗口可纵向滚动。
- Windows 发布文件：`release/A股智投工作台.exe`。
- SHA-256：`1BF0D73CEB799BDD3DD033294AE2764B8B94E1077689E2BE0D95D2F5A5627679`。
