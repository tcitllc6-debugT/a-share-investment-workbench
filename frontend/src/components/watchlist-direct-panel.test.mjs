import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'

const source = readFileSync(new URL('./stock.vue', import.meta.url), 'utf8')

test('renders the all-watchlist data outside the legacy tab pane', () => {
  const tabsEnd = source.indexOf('</n-tabs>')
  const directPanel = source.indexOf('class="watchlist-all-panel"')

  assert.ok(directPanel > tabsEnd, 'the all-watchlist panel must not depend on n-tab-pane rendering')
  assert.match(source, /v-for="result in allTableData"/)
  assert.match(source, /showLightweightKline\(result\['股票代码'\],result\['股票名称'\]\)/)
})

test('restores the complete per-stock action strip in the direct watchlist panel', () => {
  const directPanel = source.slice(source.indexOf('class="watchlist-all-panel"'))

  for (const label of ['多周期', '日K', '分时', '资金', '详情', '公告', '研报', '成本', 'AI研判', '设置分组', '取消关注']) {
    assert.match(directPanel, new RegExp(`>${label}</n-button>|>${label}</n-dropdown>`))
  }

  for (const handler of [
    'showLightweightKline', 'showK', 'showFenshi', 'showMoney', 'search',
    'searchNotice', 'searchStockReport', 'setStock', 'openAiAnalysis',
    'handleSetGroupSelect', 'removeMonitor',
  ]) {
    assert.match(directPanel, new RegExp(`${handler}\\(`))
  }
})

test('keeps AI research visible in the direct watchlist regardless of configuration', () => {
  const directPanel = source.slice(source.indexOf('class="watchlist-all-panel"'))

  assert.match(directPanel, /@click="openAiAnalysis\(result\['股票名称'\],result\['股票代码'\]\)"/)
  assert.doesNotMatch(directPanel, /v-if="data\.openAiEnable"[^>]*>AI研判/)
  assert.match(source, /function openAiAnalysis\(stock, stockCode\)/)
  assert.match(source, /请先在 AI 控制台配置模型服务/)
})
