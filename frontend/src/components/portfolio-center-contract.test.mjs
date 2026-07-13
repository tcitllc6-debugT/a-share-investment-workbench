import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'

const read = (path) => fs.readFileSync(new URL(path, import.meta.url), 'utf8')

test('portfolio center exposes all requested asset workspaces', () => {
  const navigation = read('../navigation/workspace.mjs')
  for (const value of ['overview','stock','fund','bond','precious_metal','forex','future','crypto','bank_product']) {
    assert.match(navigation, new RegExp(`queryValue: '${value}'`))
  }
})

test('portfolio center remains usable at high DPI with both scroll axes', () => {
  const center = read('./HoldingCenter.vue')
  const table = read('./portfolio/AssetPositionTable.vue')
  assert.match(center, /min-resolution:2\.5dppx/)
  assert.match(center, /overflow:auto/)
  assert.match(table, /overflow:auto/)
  assert.match(table, /min-width:1180px/)
})

test('bank CSV and structured portfolio APIs are wired', () => {
  const center = read('./HoldingCenter.vue')
  const bank = read('./portfolio/BankProductImport.vue')
  assert.match(center, /GetPortfolioOverview/)
  assert.match(center, /RefreshAssetQuotes/)
  assert.match(bank, /PreviewBankProductCSV/)
  assert.match(bank, /ImportBankProductCSV/)
  assert.match(bank, /RollbackAssetImport/)
})

test('new holdings select a standard product before entering transaction data', () => {
  const center = read('./HoldingCenter.vue')
  const editor = read('./portfolio/AssetTransactionEditor.vue')
  const search = read('./portfolio/AssetProductSearch.vue')
  assert.match(center, /增加一笔持仓/)
  assert.doesNotMatch(center, /记一笔交易/)
  assert.match(editor, /AssetProductSearch/)
  assert.match(search, /SearchAssetInstruments/)
  assert.match(editor, /selectedProduct/)
  assert.match(editor, /metadata\?\.base_currency/)
  assert.match(editor, /metadata\?\.quote_currency/)
  assert.match(search, /产品代码由数据源自动匹配/)
})

test('bank wealth keeps manual product fields while other assets use search', () => {
  const editor = read('./portfolio/AssetTransactionEditor.vue')
  const search = read('./portfolio/AssetProductSearch.vue')
  assert.match(editor, /assetType === 'bank_product'/)
  assert.match(editor, /银行理财保留手动录入/)
  assert.match(editor, /<AssetProductSearch\s+v-else/s)
  assert.match(editor, /:busy="saving"/)
  assert.match(editor, /:locked="Boolean\(position\)"/)
  assert.doesNotMatch(editor, /:disabled="saving \|\| Boolean\(position\)"/)
  assert.doesNotMatch(search, /<input[^>]+:disabled=/s)
  assert.match(search, /:disabled="busy \|\| locked"[^>]+@click="clearSelection"/s)
})

test('asset transaction modal uses a div root compatible with the focus trap', () => {
  const editor = read('./portfolio/AssetTransactionEditor.vue')
  assert.match(editor, /<div class="asset-transaction-editor" role="dialog"/)
  assert.doesNotMatch(editor, /<section class="asset-transaction-editor"/)
})

test('asset workspace refresh button fetches quotes instead of only reloading local positions', () => {
  const center = read('./HoldingCenter.vue')
  const table = read('./portfolio/AssetPositionTable.vue')
  assert.match(table, /emit\('refresh-quotes'\)/)
  assert.match(center, /@refresh-quotes="refreshQuotes"/)
  assert.doesNotMatch(center, /<AssetPositionTable[^>]+@refresh="load"/s)
})

test('portfolio overview uses stable recent-transaction columns and live FX quotations', () => {
  const overview = read('./portfolio/PortfolioOverview.vue')
  const fxBoard = read('./portfolio/FXQuoteBoard.vue')
  assert.match(overview, /<colgroup>/)
  assert.match(overview, /table-layout:fixed/)
  assert.match(overview, /FXQuoteBoard/)
  assert.doesNotMatch(overview, /overview-panel--quotes/)
  assert.match(fxBoard, /GetMajorFXQuotes/)
  assert.match(fxBoard, /RefreshMajorFXQuotes/)
  assert.match(fxBoard, /USD: '美元'/)
  assert.match(fxBoard, /HKD: '港元'/)
  assert.match(fxBoard, /\$\{base\}\/CNY/)
})

test('stock holding entry uses the same add holding wording and autocomplete flow', () => {
  const records = read('./TradingRecordManager.vue')
  assert.match(records, /增加一笔持仓/)
  assert.match(records, /n-auto-complete/)
  assert.match(records, /GetAllStockInfoList/)
})

test('portfolio tables share readable typography, column alignment, and semantic action colors', () => {
  const positions = read('./portfolio/AssetPositionTable.vue')
  const overview = read('./portfolio/PortfolioOverview.vue')

  assert.match(positions, /--portfolio-table-font-size:13px/)
  assert.match(positions, /th\s*\{[^}]*text-align:center/s)
  assert.match(positions, /td\s*\{[^}]*text-align:center/s)
  assert.match(positions, /button\.is-add/)
  assert.match(positions, /button\.is-reduce/)
  assert.match(positions, /button\.is-danger/)

  assert.match(overview, /--portfolio-table-font-size:13px/)
  assert.match(overview, /recent-table th:nth-child\(2\)[^}]*text-align:center/s)
  assert.match(overview, /recent-table td:nth-child\(2\)[^}]*text-align:center/s)
})
