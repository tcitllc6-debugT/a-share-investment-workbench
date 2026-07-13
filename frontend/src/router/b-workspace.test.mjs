import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const read = (path) => readFileSync(new URL(path, import.meta.url), 'utf8')
const router = read('./router.js')
const app = read('../App.vue')
const holding = read('../components/HoldingCenter.vue')
const ai = read('../components/AIConsole.vue')
const about = read('../components/AboutWorkspace.vue')

assert.match(router, /HoldingCenter/)
assert.match(router, /AIConsole/)
assert.match(router, /name:\s*'holdingCenter'/)
assert.match(router, /name:\s*'aiConsole'/)
assert.match(router, /name:\s*'klineAnalysis'/)
assert.match(router, /name:\s*'radar'/)
assert.match(router, /name:\s*'about'/)

assert.match(app, /WorkspaceTopNavigation/)
assert.match(app, /@toggle-secondary/)
assert.match(app, /secondaryOpen/)
assert.match(app, /WorkspaceNavigation/)
assert.doesNotMatch(app, /pendingUpdate/)
assert.doesNotMatch(app, /updateNotificationHandler/)
assert.match(app, /showManagedUpdateNotification/)
const mounted = app.slice(app.indexOf('onMounted(() => {'))
assert.ok(
  mounted.indexOf('EventsOn("updateAvailable"') < mounted.indexOf('GetConfig().then'),
  'the update listener must be ready before asynchronous configuration finishes',
)
assert.match(mounted, /EventsOn\("updateAvailable", showManagedUpdateNotification\)/)

assert.match(holding, /TradingRecordManager/)
assert.match(ai, /ResearchReport/)
assert.match(ai, /AiRecommendStocksList/)
assert.match(ai, /AIConfiguration/)
assert.match(ai, /view === 'configuration'/)
assert.doesNotMatch(ai, /name: 'settings'/)
assert.match(about, /多个开源项目进行二次开发/)
assert.match(about, /GitHub 发布，应用内提醒/)
assert.match(about, /开源说明/)
assert.match(about, /ArvinLovegood\/go-stock/)
assert.match(about, /本项目与上游作者无隶属、授权或官方合作关系/)

console.log('B workspace route and shell source-contract checks passed')
