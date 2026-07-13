import test from 'node:test'
import assert from 'node:assert/strict'

import {
  buildWatchlistFallbackRow,
  canonicalStockSymbol,
  findFollowedStock,
  isWatchlistMember,
  watchlistSymbols,
} from './stock-symbol.mjs'

test('normalises equivalent A-share provider symbols to one identity', () => {
  assert.equal(canonicalStockSymbol('sz300418'), 'sz300418')
  assert.equal(canonicalStockSymbol('300418.SZ'), 'sz300418')
  assert.equal(canonicalStockSymbol('SZ300418'), 'sz300418')
})

test('keeps followed symbols visible when quote payloads use another provider format', () => {
  const symbols = watchlistSymbols([{ StockCode: 'sz300418' }, { StockCode: 'sh600519' }])
  assert.deepEqual(symbols, ['sz300418', 'sh600519'])
  assert.equal(isWatchlistMember(symbols, '300418.SZ'), true)
  assert.equal(isWatchlistMember(symbols, '300418'), true)
  assert.equal(isWatchlistMember(symbols, '600519.SH'), true)
  assert.equal(isWatchlistMember(symbols, '000001.SZ'), false)
})

test('finds followed stock details using a provider-formatted quote code', () => {
  const followed = [{ StockCode: 'sz300418', CostPrice: 47.5 }]
  assert.equal(findFollowedStock(followed, '300418.SZ')?.CostPrice, 47.5)
})

test('builds an immediately visible row from locally saved watchlist data', () => {
  const row = buildWatchlistFallbackRow({
    StockCode: 'sz300418',
    Name: '昆仑万维',
    Price: 49.42,
    Sort: 1,
  })

  assert.deepEqual(row, {
    '股票代码': 'sz300418',
    '股票名称': '昆仑万维',
    '当前价格': 49.42,
    '卖一报价': 49.42,
    '昨日收盘价': 49.42,
    changePercent: 0,
    sort: 1,
    quotePending: true,
  })
})
