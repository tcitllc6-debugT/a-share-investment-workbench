import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import * as schema from './portfolio/asset-form-schema.mjs'

const overview = fs.readFileSync(new URL('./portfolio/PortfolioOverview.vue', import.meta.url), 'utf8')
const fxBoard = fs.readFileSync(new URL('./portfolio/FXQuoteBoard.vue', import.meta.url), 'utf8')
const workspaceHeader = fs.readFileSync(new URL('./shell/WorkspaceHeader.vue', import.meta.url), 'utf8')

test('portfolio overview maps internal asset and action enums to Chinese labels', () => {
  assert.equal(typeof schema.getAssetDisplayLabel, 'function')
  assert.equal(schema.getAssetDisplayLabel('bank_wealth'), '银行理财')
  assert.equal(schema.getAssetDisplayLabel('bank_product'), '银行理财')
  assert.equal(schema.getAssetDisplayLabel('futures'), '期货')
  assert.equal(schema.getAssetDisplayLabel('future'), '期货')

  assert.equal(typeof schema.getActionDisplayLabel, 'function')
  assert.equal(schema.getActionDisplayLabel('buy'), '买入')
  assert.equal(schema.getActionDisplayLabel('sell'), '卖出')
  assert.equal(schema.getActionDisplayLabel('subscribe'), '申购')
  assert.equal(schema.getActionDisplayLabel('redeem'), '赎回')
})

test('portfolio overview displays Chinese currency names', () => {
  assert.equal(typeof schema.getCurrencyDisplayLabel, 'function')
  assert.equal(schema.getCurrencyDisplayLabel('CNY'), '人民币')
  assert.equal(schema.getCurrencyDisplayLabel('USD'), '美元')
  assert.equal(schema.getCurrencyDisplayLabel('HKD'), '港币')
  assert.equal(schema.getCurrencyDisplayLabel('EUR'), '欧元')
  assert.equal(schema.getCurrencyDisplayLabel('USDT'), '美元稳定币')
  assert.ok(schema.CURRENCY_OPTIONS.every((item) => typeof item === 'object' && item.label && item.value))
})

test('portfolio overview filters zero-value asset and currency rows', () => {
  assert.match(overview, /filter\(\(row\) => row\.value > 0\)/)
  assert.doesNotMatch(overview, /ASSET_TYPE_MAP\[item\.key\]\?\.label \|\| item\.key/)
  assert.doesNotMatch(overview, /<b>\{\{ item\.key \}\}<\/b>/)
})

test('portfolio overview uses a live FX board instead of currency exposure and quote health', () => {
  assert.match(overview, /FXQuoteBoard/)
  assert.match(fxBoard, /实时外汇报价/)
  assert.doesNotMatch(overview, /币种分布（人民币折算）/)
  assert.doesNotMatch(overview, /报价状态/)
  assert.doesNotMatch(overview, /QuoteStatus/)
})

test('portfolio overview uses clear Chinese headings and high-DPI typography', () => {
  assert.doesNotMatch(overview, /币种敞口/)
  assert.match(overview, /--overview-label-size:\s*12px/)
  assert.match(overview, /--overview-body-size:\s*13px/)
  assert.match(overview, /@media \(min-resolution:2\.5dppx\)/)
  assert.match(overview, /--overview-label-size:\s*13px/)
  assert.match(overview, /font-weight:700/)
})

test('workspace page badge uses Chinese section label instead of an English key', () => {
  assert.match(workspaceHeader, /sectionBadges/)
  assert.match(workspaceHeader, /holding:\s*'持仓'/)
  assert.match(workspaceHeader, /\{\{ sectionBadge \}\}/)
  assert.doesNotMatch(workspaceHeader, /key\?\.toUpperCase/)
  assert.doesNotMatch(workspaceHeader, /'OVERVIEW'/)
})
