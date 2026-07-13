import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const read = (path) => readFileSync(new URL(path, import.meta.url), 'utf8')
const holdingCenter = read('./HoldingCenter.vue')
const records = read('./TradingRecordManager.vue')

assert.match(holdingCenter, /金融产品持仓/)
assert.match(holdingCenter, /currentView === 'stock'/)
assert.match(records, /holding-summary-strip/)
assert.match(records, /:cols="6"/)
assert.match(records, /overflow-x:auto/)

console.log('holding summary layout source-contract checks passed')
