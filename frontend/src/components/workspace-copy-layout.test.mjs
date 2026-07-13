import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'

const read = (name) => readFileSync(new URL(name, import.meta.url), 'utf8')
const overview = read('./DecisionWorkbench.vue')
const screening = read('./ScreeningHub.vue')
const ai = read('./AIConsole.vue')
const stock = read('./stock.vue')
const app = read('../App.vue')

test('uses short understandable one-line workspace copy', () => {
  assert.match(overview, /一屏掌握自选、持仓与市场动态/)
  assert.match(overview, /快捷进入常用工具/)
  assert.match(screening, /按形态、趋势和策略筛选 A 股/)
  assert.match(ai, /复盘研究结论，把证据和判断留在可追溯的记录里/)
  assert.match(ai, /管理本地提示词模板，让每次分析按你熟悉的框架展开/)
  assert.doesNotMatch(screening, /把形态、趋势和条件策略放进同一套观察体系/)
  assert.doesNotMatch(ai, /把研究过程留在可复核的记录里/)
})

test('secondary navigation is the only switcher in screening and AI workspaces', () => {
  assert.doesNotMatch(screening, /<n-tabs|<n-tab-pane/)
  assert.doesNotMatch(ai, /class="ai-console__tabs"/)
  assert.doesNotMatch(ai, /role="tablist"/)
})

test('top status only reports the A-share market', () => {
  const statusFunction = app.slice(app.indexOf('function updateMarketStatus'), app.indexOf('async function handleKlineAnalysisClick'))
  assert.match(statusFunction, /IsTradingTime/)
  assert.doesNotMatch(statusFunction, /IsHKTradingTime|IsUSTradingTime|港股|美股/)
})

test('watchlist toolbar keeps search group and count together', () => {
  assert.match(stock, /watchlist-all-panel__toolbar[^}]*justify-content:flex-start/s)
  assert.match(stock, /watchlist-all-panel__count/)
})
