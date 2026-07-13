import assert from 'node:assert/strict'
import test from 'node:test'

import {
  latestMarketPulse,
  mergeWatchlistQuotes,
  loadDashboardResources,
  normaliseFollowedStock,
  normaliseNewsItem,
} from './decision-workbench.mjs'

test('normalises Wails FollowedStock fields for immediate watchlist rendering', () => {
  assert.deepEqual(normaliseFollowedStock({
    StockCode: 'sz300418',
    Name: '昆仑万维',
    Price: 47.5,
    ChangePercent: 2.36,
  }), {
    stockCode: 'sz300418',
    name: '昆仑万维',
    price: 47.5,
    changePercent: 2.36,
  })
})

test('keeps a locally followed stock visible when quote fields are empty', () => {
  const row = normaliseFollowedStock({ StockCode: 'sh600519', Name: '贵州茅台' })
  assert.equal(row.name, '贵州茅台')
  assert.equal(row.price, null)
  assert.equal(row.changePercent, null)
})

test('overlays real-time quote fields onto the local overview watchlist', () => {
  const rows = mergeWatchlistQuotes(
    [{ StockCode: 'sz300418', Name: '昆仑万维', Price: 48.25, ChangePercent: 0 }],
    [{ Code: 'sz300418', Price: '48.29', ChangePercent: -2.29 }],
  )
  assert.deepEqual(rows, [{
    stockCode: 'sz300418',
    name: '昆仑万维',
    price: 48.29,
    changePercent: -2.29,
  }])
})

test('normalises news fields from generated and plain Wails payloads', () => {
  assert.deepEqual(normaliseNewsItem({ ID: 8, Time: '09:31', Content: 'A股开盘' }), {
    id: 8,
    time: '09:31',
    content: 'A股开盘',
    source: '',
  })
})

test('publishes successful dashboard resources even when another resource fails', async () => {
  const values = {}
  const errors = []
  await loadDashboardResources({
    watchlist: async () => [{ StockCode: 'sz300418', Name: '昆仑万维' }],
    news: async () => { throw new Error('timeout') },
  }, (key, value) => { values[key] = value }, (key) => errors.push(key))
  assert.equal(values.watchlist.length, 1)
  assert.deepEqual(errors, ['news'])
})

test('uses the latest local market statistic as the overview pulse', () => {
  const pulse = latestMarketPulse([
    { dataTime: '09:35', upCount: 1000, downCount: 3000 },
    { dataTime: '14:55', upCount: 3210, downCount: 1840, limitUp: 68, limitDown: 7, upRatio: 63.6, sentimentDesc: '偏强' },
  ])
  assert.equal(pulse.upCount, 3210)
  assert.equal(pulse.limitUp, 68)
  assert.equal(pulse.sentimentDesc, '偏强')
})
