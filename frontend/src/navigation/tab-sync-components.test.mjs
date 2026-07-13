import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const read = (path) => readFileSync(new URL(path, import.meta.url), 'utf8')
const market = read('../components/market.vue')
const screening = read('../components/ScreeningHub.vue')
const overview = read('../components/DecisionWorkbench.vue')

assert.match(market, /normaliseMarketTab/)
assert.match(market, /toMarketTabKey/)
assert.match(market, /watch\(/)
assert.match(market, /EventsOff\("tradingViewNews"\)/)
assert.match(screening, /normaliseScreeningTab/)
assert.match(screening, /mode/)
assert.match(screening, /DuoKongLineSelector/)
assert.match(screening, /AllStockList/)
assert.match(screening, /SelectStock/)
assert.match(overview, /name: 'radar'/)
assert.match(overview, /name: 'holdingCenter'/)
assert.match(overview, /mode: 'duo-kong'/)

console.log('query-driven market and screening source-contract checks passed')
