# Upstream Sync 2026-07-09

## Source Checked

- Repository: `ArvinLovegood/go-stock`
- Latest release tag checked: `v2026.07.09.1-release`
- Commit checked: `92e174b15778fd8dea5e7208d976a702cb9edad0`
- Commit time: `2026-07-09 09:49:10 +0800`
- Commit title: `feat(ai):为AI推荐记录关联模型与提示词上下文并优化前端展示`

## Upstream Change Summary

The upstream update mainly does five things:

- Stores the actual model name, system prompt, and user prompt with AI stock recommendation records.
- Shows that generation context in the AI recommendation detail modal.
- Moves the K-line follow button into the chart toolbar area.
- Fixes pagination views that lacked total item counts.
- Updates generated TypeScript model definitions for the new backend fields.

## Ported Into This Replica

- Added `systemPrompt` and `userPrompt` fields to `AiRecommendStocks`.
- Captured system/user prompt context in OpenAI tool execution and agent tool execution.
- Injected model and prompt context when creating one or multiple AI recommendation records.
- Added tests for recommendation metadata injection in agent tools.
- Added model/prompt context display to `aiRecommendStocksList.vue`.
- Updated `frontend/wailsjs/go/models.ts` for the new recommendation fields.
- Added prompt-template pagination page synchronization and total item counts.
- Added cron-task pagination `itemCount`.
- Added menu dropdown height limiting in `App.vue`.
- Moved the K-line follow/unfollow control from the bottom hint row into the main toolbar.

## Intentionally Not Ported

- Upstream README sponsor/comment text was not copied into the replica documentation.
- The upstream prompt plaza default remote API setting was not restored. This replica keeps `promptPlazaApiBase` blank by default so prompt templates remain local-first.
- Original-author hosted VIP/update/device-binding style dependencies remain outside the replica boundary.

## Replica-Specific Additions In The Same Pass

- Seeded local A-share prompt templates for research workflows.
- Documented research-center semantics and distribution requirements.
- Fixed TradingView original-article URLs by preferring `link` and `storyPath` instead of constructing links from internal ids.
- Added a public-provider startup backfill for recent market news while keeping the original remote push sync disabled.
