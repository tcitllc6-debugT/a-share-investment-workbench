import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const read = (name) => readFileSync(new URL(name, import.meta.url), 'utf8')
const stock = read('stock.vue')
const trading = read('TradingRecordManager.vue')
const settings = read('settings.vue')
const app = read('../App.vue')

assert.match(stock, /watchlist-all-panel__scroll\s*\{\s*overflow:auto/)
assert.match(stock, /watchlist-all-panel__table\s*\{\s*width:100%;\s*min-width:1120px/)
assert.match(trading, /holding-summary-strip/)
assert.match(trading, /:cols="6"/)
assert.match(trading, /overflow-x:auto/)
assert.match(trading, /:scroll-x="1120"/)
assert.match(trading, /max-height:\s*88dvh/)
assert.doesNotMatch(settings, /useRoute/)
assert.doesNotMatch(settings, /aiSettingsRef/)
assert.match(settings, /settings-workspace :deep\(\.n-card__content\) \{ overflow-x:auto/)
assert.match(app, /workspace-shell__body[^}]*overflow:hidden/)
assert.match(app, /workspace-shell__main[^}]*overflow:hidden/)
assert.match(app, /workspace-shell__content[^}]*overflow-y:auto/)
assert.match(app, /workspace-shell__spin \.n-spin-content[^}]*height:100%/)
assert.match(app, /workspace-shell__spin \.n-spin-content[^}]*overflow:hidden/)

console.log('high-DPI source-contract checks passed')
