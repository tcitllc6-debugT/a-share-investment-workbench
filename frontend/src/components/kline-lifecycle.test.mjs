import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const analysis = readFileSync(new URL('./kline-analysis.vue', import.meta.url), 'utf8')
const chart = readFileSync(new URL('./KLineChart.vue', import.meta.url), 'utf8')

assert.match(analysis, /EventsOff\('klineSelectStock'\)/)
assert.match(chart, /onBeforeUnmount/)
assert.match(chart, /ResizeObserver/)
assert.match(chart, /\.dispose\(\)/)

console.log('K line lifecycle source-contract checks passed')
