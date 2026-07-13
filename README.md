# A股智投工作台

面向个人投资研究的本地优先 A 股桌面工作台。把自选看盘、市场资讯、多维选股、盘中异动、统一持仓和 AI 投研放进一套连续的研究流程，让每天的观察、筛选、记录与复盘留在同一个桌面里。

## 核心能力

- **股票自选与多周期 K 线**：支持自选分组、实时行情、日 K/分时/资金、公告、研报和 AI 研判；总览与自选使用同一套实时行情链路。
- **市场资讯与盘中异动**：聚合市场快讯、指数、行业与板块资金、热点内容和盘中异动雷达，发现信号后可直接跳转至个股分析。
- **多维度选股**：提供形态选股、多空线选股和策略选股。多空线支持近 3 个交易日信号、趋势状态、板块/行业/概念筛选、结果排序与多周期 K 线复核。
- **统一资产配置**：在一个持仓中心记录股票、基金/ETF、债券、贵金属、外汇、期货、加密资产和银行理财，统一折算人民币并展示市值、成本、盈亏和交易流水。
- **AI 投研工作流**：内置 AI 配置、提示词模板、分析记录、股票推荐记录与持仓数据摘要；模型密钥和个人研究数据优先保存在本机。
- **桌面体验与发布**：支持高 DPI、日间/暗黑主题、Windows 便携版、签名校验的在线更新，以及 macOS 构建工作流。

## 本项目新增与改造

- 重新规划为“总览、股票自选、市场资讯、多维度选股、持仓中心、AI 控制台、设置、关于”八个工作区，统一导航、交互与视觉语言。
- 增加多空线指标及全市场扫描能力，覆盖细线/粗线交叉、站上站下、双线趋势、拐头、信号新鲜度与多周期复核。
- 增加跨资产持仓中心与本地账本，支持多资产报价、汇率折算、资产配置和交易记录管理。
- 增加盘中异动雷达，第一版专注发现、提醒与跳转分析，为后续合规扩展预留接口。
- 将 AI 配置从通用设置中拆分为独立控制台，补齐服务商预设、连接测试、用途选择和本地提示词模板。
- 改造市场资讯首屏加载和总览自选行情刷新逻辑，优先展示本地内容，再异步叠加实时数据。
- 建立自己的签名更新清单、Windows 便携版和 GitHub 发布链路，更新不再依赖上游发布通道。
- 从本地功能中移除赞助商、VIP 和设备绑定限制。
- 避免默认依赖原作者社区服务。

## Build

Requirements:

- Go
- Node.js and npm
- Wails CLI

Commands:

```powershell
npm --prefix frontend install
npm --prefix frontend run build
go test ./... -run '^$' -count=0
wails build
```

Windows build output is written to:

```text
build/bin/A股智投工作台.exe
```

## 本地启动

Windows 独立版入口：`release/A股智投工作台.exe`。首次运行后可在“设置”中切换日间/暗黑主题并手动检查更新。自选列表的“AI研判”始终显示；未配置模型服务时，点击后会引导前往 AI 配置。

日间主题以白色为主、蓝色为强调色；暗黑主题使用深蓝灰工作区。主题会保存在本机，下一次启动继续使用上次选择。

## AI 配置

AI 模型服务统一在“AI 控制台 > AI 配置”中管理，普通“设置”页不再包含 AI 选项。每个模型服务可以独立配置服务商、接口地址、API 密钥和模型，并可直接获取模型列表、测试连接与查看失败原因。

- “默认研究模型”用于个股研判、研究报告和市场分析。
- “默认智能体模型”用于 AI 助手与自动化任务。
- API 密钥仅保存在本机数据库中；页面读取和设置导出都不会返回密钥明文。
- 修改接口地址后必须重新填写 API 密钥，避免旧密钥被发送到新的服务地址。
- 连接测试只发送最小诊断请求，不包含股票、自选或持仓数据。

## macOS DMG

每次推送到 `main` 后，发布工作流会在 macOS runner 上构建同时支持 Intel 与 Apple Silicon 的通用 DMG，并把它加入同一版本的 GitHub Release。`Build macOS DMG` 手动工作流仍可用于单独验证 macOS 构建，产物会保存为 Actions Artifact。

当前公开 CI 产物使用 ad-hoc 签名，未经过 Apple 公证；首次打开时可能需要在 Finder 中右键选择“打开”。取得 Developer ID 后，可在 GitHub Actions 中配置证书并增加 notarization 步骤，消除 Gatekeeper 提示。

## GitHub 自动更新与发布

- 客户端只从本仓库的 HTTPS GitHub Releases 下载 `manifest.json` 和版本压缩包，不再使用自建更新服务器。
- 推送到 `main` 后，`Publish A股智投工作台 Update` 工作流会发布 Windows 热更新包、Windows 便携版、macOS 通用 DMG 和签名清单。
- 仓库需要配置 Actions Secret：`UPDATE_MANIFEST_PRIVATE_KEY`。私钥只放在 GitHub Secrets；客户端仅内置对应公钥。
- 应用启动后在后台检查正式版本，发现新版本时提示；Windows 可由更新助手原地替换并保留旧版回滚状态，macOS 会打开对应 DMG 下载页；“设置”和“关于”仍提供手动检查入口。
- 客户端在应用更新前校验 Ed25519 签名和 SHA-256 摘要，更新异常时保留原版本并记录回滚状态。

## 本地数据与外部数据源

The app still uses public market data APIs and user-configured AI model providers. Replica-specific boundaries are being moved under `backend/provider`.

Current local-first defaults:

- Prompt plaza API: blank means local prompt-template mode.
- Share export: writes Markdown under `data/shares`.
- Update check: uses this repository's signed GitHub Releases channel.

## A股智投工作台

面向 A 股研究的本地优先桌面工作台，覆盖自选与行情、K 线和多维选股、盘中异动、统一持仓、AI 投研与本地提示词模板。项目持续以自己的 GitHub 仓库、签名更新包和发布流程迭代。

## 开源说明

- 本项目包含基于 ArvinLovegood/go-stock 修改的代码。
- 上游项目：https://github.com/ArvinLovegood/go-stock
- 本项目按 GNU GPL v3.0 发布。
- 修改内容与对应源码：https://github.com/tcitllc6-debugT/a-share-investment-workbench
- 本项目与上游作者无隶属、授权或官方合作关系。

## Risk Notice

Market data and AI output are for learning and research only. They are not investment advice, trading advice, or a guarantee of correctness. Investment involves risk.
