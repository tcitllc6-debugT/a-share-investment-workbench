# Go Stock Replica Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn the current go-stock repository into an independently buildable and maintainable desktop replica that does not depend on the original author's hosted services for normal use.

**Architecture:** Keep the Wails + Go + Vue desktop architecture. Add independence through dependency inventory, provider boundaries, configuration defaults, rebranding, and targeted removal or replacement of original-author service calls.

**Tech Stack:** Go 1.26.x, Wails 2.11.x, Vue 3, Vite 7, Naive UI, SQLite/Gorm, Resty, Eino/OpenAI-compatible AI integrations.

## Global Constraints

- Keep the main product as a desktop Wails app, not a browser-first SaaS.
- Do not make destructive git operations.
- Preserve user/local data under `data/`.
- Do not require access to `go-stock.sparkmemory.top` for normal use.
- Do not require original-author VIP, sponsor code, device binding, or update mirrors.
- Keep GPLv3 compliance in mind for any distributed derivative.
- Use additive schema changes during the first phase.
- External data-source failures must not panic the app.

---

## File Structure

Expected new files:

- `docs/replica/external-dependencies.md`: human-readable inventory of external endpoints.
- `backend/data/external_dependency_inventory.go`: optional static inventory helpers if code-level inventory is useful.
- `backend/provider/share/provider.go`: share/export abstraction.
- `backend/provider/update/provider.go`: update-check abstraction.
- `backend/provider/prompt/provider.go`: prompt library abstraction.
- `backend/provider/README.md`: provider boundary notes.

Expected modified files:

- `app.go`: remove or route original-author share/update/VIP/device-binding flows through providers.
- `main.go`: keep startup stable and remove any mandatory original-author service assumptions.
- `backend/data/settings_api.go`: change unsafe defaults for prompt plaza/update/original services.
- `ai-assistant-web/server.go`: remove or make configurable the VIP2 gate and share upload dependency.
- `frontend/src/components/about.vue`: rebrand visible app identity and remove original donation/community content.
- `frontend/src/components/settings.vue`: update defaults and labels for prompt plaza, update, and provider configuration.
- `frontend/src/components/promptPlaza.vue`: support local-only or configurable prompt library behavior.
- `frontend/src/components/promptQa.vue`: support local-only or configurable prompt library behavior.
- `wails.json`: update visible product metadata.
- `README.md`: replace with replica build/run documentation.

---

### Task 1: Baseline Build And Environment Record

**Files:**
- Create: `docs/replica/build-baseline.md`

**Interfaces:**
- Consumes: existing repository.
- Produces: a reproducible baseline document with exact build commands and observed outputs.

- [ ] **Step 1: Write baseline document**

Create `docs/replica/build-baseline.md` with:

````markdown
# Build Baseline

## Environment

- Go: `go version`
- Node: `node --version`
- npm: `npm --version`
- Wails: `wails version`

## Commands

```powershell
npm --prefix frontend run build
go test ./... -run '^$' -count=0
wails build
```

## Expected Result

- Frontend production build succeeds.
- Go packages compile without running integration tests.
- Wails creates `build/bin/go-stock.exe`.

## Known Warnings

- Vite may warn about large chunks.
- Wails binding generation may print unresolved helper types such as `time.Time` or `gorm.DeletedAt`; these warnings do not currently block the build.
```
````

- [ ] **Step 2: Run build commands**

Run:

```powershell
npm --prefix frontend run build
go test ./... -run '^$' -count=0
wails build
```

Expected: all commands exit with code 0.

- [ ] **Step 3: Commit**

```powershell
git add docs/replica/build-baseline.md
git commit -m "docs: record replica build baseline"
```

---

### Task 2: External Dependency Inventory

**Files:**
- Create: `docs/replica/external-dependencies.md`

**Interfaces:**
- Consumes: repository search results for URLs and service names.
- Produces: a dependency table used by later replacement tasks.

- [ ] **Step 1: Collect endpoint references**

Run:

```powershell
rg "https?://|go-stock\\.sparkmemory\\.top|8\\.134\\.249\\.145|api\\.github\\.com/repos/ArvinLovegood" -n app.go main.go backend ai-assistant-web frontend/src
```

Expected: references to original services, public market APIs, AI providers, and static upstream URLs.

- [ ] **Step 2: Create dependency inventory**

Create `docs/replica/external-dependencies.md` with this table structure:

```markdown
# External Dependency Inventory

## Original-Author Hosted Services

| Feature | File | Endpoint | Current Behavior | Replica Decision |
| --- | --- | --- | --- | --- |
| Share upload | `app.go`, `ai-assistant-web/server.go` | `http://go-stock.sparkmemory.top:16688/upload` | Uploads analysis text | Replace with local export provider |
| News sync | `app.go` | `http://go-stock.sparkmemory.top:16666/FinancialNews/json` | Pulls synced news | Disable or replace with public news provider |
| Prompt plaza | `frontend/src/components/*.vue`, `settings.vue` | `http://go-stock.sparkmemory.top:1918/api` | Remote prompt plaza | Local prompt library by default |
| Embedded website | `frontend/src/components/stockhotmap.vue` | `https://go-stock.sparkmemory.top:16667/go-stock` | Embedded external page | Remove or make optional |
| Stock basic mirror | `app.go` | `http://8.134.249.145:18080/go-stock/*.json` | Downloads stock basics | Replace with bundled JSON or user-configured mirror |

## Public Data Sources

| Feature | File | Endpoint | Key Required | Failure Fallback |
| --- | --- | --- | --- | --- |
| Sina quote | `backend/data/stock_data_api.go` | `http://hq.sinajs.cn` | No | Empty quote with error log |
| Tencent quote | `backend/data/stock_data_api.go` | `http://qt.gtimg.cn` | No | Empty quote with error log |
| Tushare | `backend/data/stock_data_api.go` | `http://api.tushare.pro` | Yes | Use bundled stock basics |
| Eastmoney | multiple `backend/data` files | `eastmoney.com` APIs | Sometimes | Fallback K-line provider or empty state |
| Cailianpress | `app.go`, `backend/data/market_news_api.go` | `cls.cn` APIs | No | Empty news list |

## User Configured Services

| Feature | Config Field | Required For |
| --- | --- | --- |
| AI models | `ai_config` rows | AI analysis and agent chat |
| Tushare | `settings.tushare_token` | Stock basic refresh |
| DingTalk | `settings.ding_robot` | DingTalk notifications |
| Feishu | `settings.feishu_*` | Feishu notifications and bot |
| HTTP proxy | `settings.http_proxy` | Optional crawler/model proxy |
```

- [ ] **Step 3: Verify inventory has no empty decisions**

Run:

```powershell
Select-String -Path docs/replica/external-dependencies.md -Pattern "TBD|TODO"
```

Expected: no matches.

- [ ] **Step 4: Commit**

```powershell
git add docs/replica/external-dependencies.md
git commit -m "docs: inventory replica external dependencies"
```

---

### Task 3: Disable Original-Author Defaults

**Files:**
- Modify: `backend/data/settings_api.go`
- Modify: `frontend/src/components/settings.vue`
- Modify: `frontend/src/components/promptTemplateList.vue`
- Modify: `frontend/src/components/promptPlaza.vue`
- Modify: `frontend/src/components/promptQa.vue`

**Interfaces:**
- Consumes: dependency inventory from Task 2.
- Produces: settings where original-author endpoints are blank or explicitly user-configured.

- [ ] **Step 1: Write failing settings test**

Create or update a Go test that initializes an empty SQLite database and calls `data.GetSettingConfig()`.

Test behavior:

```go
func TestReplicaSettingsDoNotRequireOriginalAuthorServices(t *testing.T) {
	cfg := GetSettingConfig()
	require.NotNil(t, cfg)
	require.NotContains(t, cfg.PromptPlazaApiBase, "go-stock.sparkmemory.top")
}
```

Run:

```powershell
go test ./backend/data -run TestReplicaSettingsDoNotRequireOriginalAuthorServices -count=1
```

Expected: fail before implementation if the default still contains the original endpoint.

- [ ] **Step 2: Update settings defaults**

Change default prompt plaza behavior so an empty value means local-only mode.

Implementation rule:

```go
if settings.PromptPlazaApiBase == "" {
    settings.PromptPlazaApiBase = ""
}
```

Do not inject `http://go-stock.sparkmemory.top:1918/api` as a default anywhere.

- [ ] **Step 3: Update frontend copy**

Replace placeholder text that suggests the original prompt plaza endpoint with:

```text
留空则使用本地提示词模板；如需远程广场，请填写你自己的 API 地址
```

- [ ] **Step 4: Run tests and builds**

Run:

```powershell
go test ./backend/data -run TestReplicaSettingsDoNotRequireOriginalAuthorServices -count=1
npm --prefix frontend run build
go test ./... -run '^$' -count=0
```

Expected: all pass.

- [ ] **Step 5: Commit**

```powershell
git add backend/data/settings_api.go frontend/src/components/settings.vue frontend/src/components/promptTemplateList.vue frontend/src/components/promptPlaza.vue frontend/src/components/promptQa.vue backend/data/*_test.go
git commit -m "refactor: remove original prompt plaza defaults"
```

---

### Task 4: Local Share Provider

**Files:**
- Create: `backend/provider/share/provider.go`
- Modify: `app.go`
- Modify: `ai-assistant-web/server.go`

**Interfaces:**
- Consumes: analysis text, title, stock code, stock name.
- Produces: local Markdown export path or a typed error.

- [ ] **Step 1: Write provider test**

Create `backend/provider/share/provider_test.go`:

```go
package share

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLocalProviderWritesMarkdown(t *testing.T) {
	dir := t.TempDir()
	p := NewLocalProvider(dir)

	path, err := p.ShareText("hello", "AI助手")

	require.NoError(t, err)
	require.FileExists(t, path)
	require.Equal(t, filepath.Join(dir, "AI助手.md"), path)
	b, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(b), "hello")
}
```

Run:

```powershell
go test ./backend/provider/share -run TestLocalProviderWritesMarkdown -count=1
```

Expected: fail because provider does not exist.

- [ ] **Step 2: Implement provider**

Create `backend/provider/share/provider.go`:

```go
package share

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Provider interface {
	ShareText(text string, title string) (string, error)
}

type LocalProvider struct {
	dir string
}

func NewLocalProvider(dir string) *LocalProvider {
	return &LocalProvider{dir: dir}
}

func (p *LocalProvider) ShareText(text string, title string) (string, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", errors.New("share text is empty")
	}
	title = sanitizeTitle(title)
	if title == "" {
		title = "analysis"
	}
	if err := os.MkdirAll(p.dir, 0755); err != nil {
		return "", err
	}
	path := filepath.Join(p.dir, title+".md")
	return path, os.WriteFile(path, []byte(text), 0644)
}

func sanitizeTitle(title string) string {
	title = strings.TrimSpace(title)
	re := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]`)
	title = re.ReplaceAllString(title, "_")
	title = strings.Trim(title, ". ")
	return title
}
```

- [ ] **Step 3: Route app share calls through provider**

In `app.go`, replace direct POST calls to `http://go-stock.sparkmemory.top:16688/upload` with local provider usage.

Behavior:

- `ShareText` returns the saved Markdown path.
- `ShareAnalysis` writes analysis text locally.
- No original upload endpoint is called.

- [ ] **Step 4: Route web assistant share through provider**

In `ai-assistant-web/server.go`, replace the direct upload POST with the local provider.

Return JSON:

```json
{"message":"saved to <path>"}
```

- [ ] **Step 5: Run tests and build**

```powershell
go test ./backend/provider/share -count=1
go test ./... -run '^$' -count=0
wails build
```

Expected: all pass.

- [ ] **Step 6: Commit**

```powershell
git add backend/provider/share app.go ai-assistant-web/server.go
git commit -m "feat: replace remote share upload with local export"
```

---

### Task 5: Update Provider

**Files:**
- Create: `backend/provider/update/provider.go`
- Modify: `app.go`
- Modify: `wails.json`

**Interfaces:**
- Consumes: current version and optional update URL.
- Produces: disabled update result or release metadata.

- [ ] **Step 1: Write disabled-provider test**

Create `backend/provider/update/provider_test.go`:

```go
package update

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDisabledProviderReturnsNoUpdate(t *testing.T) {
	p := DisabledProvider{}
	result, err := p.Check(context.Background(), "1.0.0")
	require.NoError(t, err)
	require.False(t, result.Available)
	require.Equal(t, "disabled", result.Source)
}
```

Expected: fail before provider exists.

- [ ] **Step 2: Implement update provider**

Create:

```go
package update

import "context"

type Result struct {
	Available bool
	Version   string
	URL       string
	Source    string
}

type Provider interface {
	Check(ctx context.Context, currentVersion string) (Result, error)
}

type DisabledProvider struct{}

func (DisabledProvider) Check(ctx context.Context, currentVersion string) (Result, error) {
	return Result{Available: false, Source: "disabled"}, nil
}
```

- [ ] **Step 3: Disable original update channel by default**

In `app.go`, make `CheckUpdate` return a disabled/no-update message unless a replica-owned update URL is configured.

Do not call:

```text
https://api.github.com/repos/ArvinLovegood/go-stock
https://gh.927223.xyz/
```

for replica builds.

- [ ] **Step 4: Update product metadata**

In `wails.json`, replace original author/product comments with replica-neutral values.

- [ ] **Step 5: Run verification**

```powershell
go test ./backend/provider/update -count=1
go test ./... -run '^$' -count=0
wails build
```

- [ ] **Step 6: Commit**

```powershell
git add backend/provider/update app.go wails.json
git commit -m "feat: disable original update channel by default"
```

---

### Task 6: Remove VIP And Device Binding From Core Flows

**Files:**
- Modify: `app.go`
- Modify: `backend/data/sponsor_vip.go`
- Modify: `ai-assistant-web/server.go`
- Modify: `frontend/src/App.vue`
- Modify: `frontend/src/components/about.vue`
- Modify: `frontend/src/components/kline-analysis.vue`

**Interfaces:**
- Consumes: feature availability checks.
- Produces: local feature flags or all-local access for non-cloud features.

- [ ] **Step 1: Write feature-gate test**

Add a test that verifies core local features do not require sponsor code.

```go
func TestLocalFeaturesAreAvailableWithoutSponsorCode(t *testing.T) {
	require.True(t, IsLocalFeatureEnabled("kline_analysis"))
	require.True(t, IsLocalFeatureEnabled("ai_assistant_web"))
}
```

Expected: fail until local feature flag helper exists.

- [ ] **Step 2: Implement local feature flag helper**

Create helper near existing sponsor/VIP logic:

```go
func IsLocalFeatureEnabled(feature string) bool {
	switch feature {
	case "kline_analysis", "ai_assistant_web":
		return true
	default:
		return true
	}
}
```

- [ ] **Step 3: Replace VIP2 checks for local features**

Replace checks that block K-line analysis or web assistant solely because VIP2 is missing.

Keep optional sponsor-code display only as legacy data if needed, but do not block local features.

- [ ] **Step 4: Remove device-binding requirement from normal UI**

Ensure `CheckDeviceBinding` is not required for core app usage.

- [ ] **Step 5: Run verification**

```powershell
go test ./backend/data -run TestLocalFeaturesAreAvailableWithoutSponsorCode -count=1
npm --prefix frontend run build
go test ./... -run '^$' -count=0
```

- [ ] **Step 6: Commit**

```powershell
git add app.go backend/data/sponsor_vip.go ai-assistant-web/server.go frontend/src/App.vue frontend/src/components/about.vue frontend/src/components/kline-analysis.vue
git commit -m "refactor: remove sponsor gates from local features"
```

---

### Task 7: Rebrand Visible UI

**Files:**
- Modify: `frontend/src/components/about.vue`
- Modify: `frontend/src/App.vue`
- Modify: `wails.json`
- Modify: `README.md`
- Optional Modify: `build/appicon.png`, `build/app.ico`

**Interfaces:**
- Consumes: chosen product name.
- Produces: visible app identity without original author donation/community content.

- [ ] **Step 1: Choose placeholder brand**

Use the temporary internal name:

```text
A-Share Desk
```

This can be changed later.

- [ ] **Step 2: Replace About page content**

Remove:

- Original author donation QR images.
- Original QQ group.
- Original sponsor copy.
- Original GitHub issue/release links as primary product links.

Add:

- GPLv3 attribution.
- Investment risk disclaimer.
- Local build information.
- User-configured provider note.

- [ ] **Step 3: Replace title and metadata**

Update:

- `wails.json` product name and comments.
- Window title in `main.go` or `App.vue` if needed.

- [ ] **Step 4: Build frontend**

```powershell
npm --prefix frontend run build
```

Expected: pass.

- [ ] **Step 5: Commit**

```powershell
git add frontend/src/components/about.vue frontend/src/App.vue wails.json README.md build/appicon.png build/app.ico
git commit -m "chore: rebrand visible application identity"
```

---

### Task 8: Provider Boundary Documentation

**Files:**
- Create: `backend/provider/README.md`

**Interfaces:**
- Consumes: provider decisions from previous tasks.
- Produces: developer guidance for future data-source refactors.

- [ ] **Step 1: Create provider README**

Create:

```markdown
# Provider Boundaries

Provider packages isolate unstable external services from the Wails API and Vue UI.

## Rules

- New original-author service calls are not allowed.
- User-configured services must be optional.
- Public data endpoints must timeout and return typed empty/error states.
- UI code should not know concrete endpoint URLs.
- Provider tests should cover missing config and failure fallback.

## Current Providers

- `share`: local Markdown export.
- `update`: disabled-by-default update checks.
- `prompt`: local prompt template library or configurable remote endpoint.

## Planned Providers

- `market`: quote, K-line, minute line, market statistics.
- `news`: market news, telegraph, hot topics, calendar.
- `research`: announcements and research reports.
- `fund`: fund ranking, holdings, net value, fund K-line.
```

- [ ] **Step 2: Commit**

```powershell
git add backend/provider/README.md
git commit -m "docs: describe replica provider boundaries"
```

---

### Task 9: Final Smoke Verification

**Files:**
- Create: `docs/replica/smoke-test.md`

**Interfaces:**
- Consumes: completed replica changes.
- Produces: manual verification checklist.

- [ ] **Step 1: Create smoke checklist**

Create:

```markdown
# Replica Smoke Test

## Build

- [ ] `npm --prefix frontend run build`
- [ ] `go test ./... -run '^$' -count=0`
- [ ] `wails build`

## Launch

- [ ] Start `build/bin/go-stock.exe`
- [ ] App opens without requiring original-author service access
- [ ] SQLite database initializes under `data/`

## Core Workflows

- [ ] Add a stock to watchlist
- [ ] Create a group
- [ ] Refresh price
- [ ] Open K-line view
- [ ] Open market/news tab
- [ ] Configure AI provider
- [ ] Run AI analysis
- [ ] Export/share analysis locally
- [ ] Open settings
- [ ] Open About page

## Negative Cases

- [ ] App handles missing AI key
- [ ] App handles network failure
- [ ] App handles empty market data
- [ ] App handles disabled update provider
```

- [ ] **Step 2: Run smoke test**

Run the commands and manually check the app.

- [ ] **Step 3: Commit**

```powershell
git add docs/replica/smoke-test.md
git commit -m "docs: add replica smoke test checklist"
```

---

## Execution Order

Recommended order:

1. Task 1
2. Task 2
3. Task 3
4. Task 4
5. Task 5
6. Task 6
7. Task 7
8. Task 8
9. Task 9

Tasks 3, 4, 5, and 6 are the highest-value independence tasks. Tasks 7 through 9 make the result distributable and easier to maintain.

## Self-Review Notes

- The plan keeps the app as a Wails desktop product.
- Original-author hosted services are explicitly inventoried and replaced or disabled.
- Each implementation task has a test or verification command.
- No task requires SaaS migration, billing, or a new remote account system.
- The plan uses provider boundaries without forcing a large rewrite of all market-data code in the first phase.
