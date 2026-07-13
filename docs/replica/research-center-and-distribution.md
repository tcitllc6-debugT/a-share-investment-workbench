# Research Center, Prompts, Distribution, And News Sync

## Research Center

The research center is a record and reuse area for AI-assisted research workflows.

- **AI analysis reports** are persisted model outputs from stock analysis, market-news analysis, scheduled analysis, and assistant flows. They are stored as `AIResponseResult` records and shown by `researchReport.vue`.
- **Stock recommendation records** are structured recommendation rows created when an AI workflow calls the recommendation tool. They are stored as `AiRecommendStocks` records and shown by `aiRecommendStocksList.vue`.
- **Prompt templates** are local reusable prompts stored in SQLite. The app separates `模型系统Prompt` and `模型用户Prompt`; these templates are consumed by market analysis, stock analysis, agent chat, and the AI assistant web view.
- **Prompt plaza** was an optional remote prompt community in the original project. It is removed from the visible replica UI because this build is local-first and should not depend on the original community service.
- **Q&A plaza** was an optional remote question-and-answer area tied to the same prompt plaza API/account. It is also removed from the visible replica UI.

## Built-In Prompt Strategy

The replica now seeds a small A-share prompt library on database initialization. Seeding is idempotent by prompt name and type, so user-created or edited templates are not overwritten.

Built-in prompt groups:

- A-share research assistant system role.
- A-share pre-market/post-market review.
- Individual-stock multi-factor analysis.
- Hot theme and capital-flow tracking.
- Structured AI recommendation output.
- Risk-control and position-discipline review.

## Stock Recommendation Records Decision

`股票推荐记录` is kept in this replica. It is not just a passive history page: AI workflows can create structured stock-pick rows through the recommendation tools, and each row can preserve the stock code, stock name, recommendation reason, risk level, target/stop-loss fields, the model name, and the system/user prompts used at generation time.

Keeping it helps A-share review and auditing: you can see which prompt/model produced a recommendation, compare later price movement against the original reasoning, and use the existing alert-monitoring path for enabled recommendations. The prompt plaza and Q&A plaza are community-service surfaces and have been removed from the visible local-first UI, but the local recommendation record flow remains useful.

## Distribution

This is a Wails desktop application, not a web-port project. End users should receive packaged binaries.

Recommended sharing model:

- **Windows users:** distribute a signed or zipped `a-share-desk.exe`. A normal user should not need Go, Node.js, npm, or Wails.
- **macOS users:** build and distribute a macOS package/app from macOS. macOS builds should be notarized for the smoothest install experience.
- **Developers or self-build users:** require Go, Node.js/npm, Wails CLI, and platform WebView dependencies.

Current local build commands:

```powershell
npm --prefix frontend install
npm --prefix frontend run build
go test ./... -run '^$' -count=0
wails build
```

## Market News Original Links

The market-news page shows "查看原文" when a news record has an openable URL. The 404 issue came from provider URL construction:

- TradingView returns an internal `id`, a `storyPath`, and sometimes a direct `link`.
- The previous code built `https://cn.tradingview.com/news/{id}` from the internal id.
- Some ids are not valid public article paths, so the generated link opened a 404 page.

The replica now prefers provider `link`, falls back to TradingView `storyPath`, and hides the button for missing or unsafe URLs.

## Startup News Sync

There are two distinct news paths:

- **Public-provider startup backfill:** the replica fetches Cailianpress, Sina Finance, and TradingView once at startup, then counts recent 24-hour records.
- **Original remote push sync:** the original `syncNews()` ntfy-style remote stream is intentionally disabled in this replica because it depends on the original author's hosted service.

This means local startup sync exists for public providers, including TradingView foreign news, but it does not depend on the original remote 24-hour push archive.
