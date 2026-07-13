import test from 'node:test'
import assert from 'node:assert/strict'

import {
  normaliseAIConsoleTab,
  normaliseMarketTab,
  normaliseScreeningTab,
  toMarketTabKey,
} from './tab-sync.mjs'

test('normalises market deep links and legacy labels to supported market tabs', () => {
  assert.equal(normaliseMarketTab({ tab: 'telegraph' }), '市场快讯')
  assert.equal(normaliseMarketTab({ tab: 'featured-sites' }), '名站优选')
  assert.equal(normaliseMarketTab({ name: '龙虎榜' }), '龙虎榜')
  assert.equal(normaliseMarketTab({ tab: 'missing' }), '市场快讯')
  assert.equal(toMarketTabKey('当前热门'), 'hot-now')
})

test('normalises the three approved screening entry points with a stable default', () => {
  assert.equal(normaliseScreeningTab({ mode: 'pattern' }), 'pattern')
  assert.equal(normaliseScreeningTab({ mode: 'duo-kong' }), 'duo-kong')
  assert.equal(normaliseScreeningTab({ tab: 'strategy' }), 'strategy')
  assert.equal(normaliseScreeningTab({ mode: 'quick' }), 'pattern')
})

test('normalises AI console deep links and falls back to analysis records', () => {
  assert.equal(normaliseAIConsoleTab({ view: 'recommendations' }), 'recommendations')
  assert.equal(normaliseAIConsoleTab({ view: 'anything-else' }), 'analysis')
})
