import assert from 'node:assert/strict'
import fs from 'node:fs'
import test from 'node:test'

const overview = fs.readFileSync(new URL('./DecisionWorkbench.vue', import.meta.url), 'utf8')
const patternScreener = fs.readFileSync(new URL('./SelectStock.vue', import.meta.url), 'utf8')
const sharedModalPath = new URL('./StockKlineModal.vue', import.meta.url)

test('overview and pattern screening share the in-app multi-period K-line modal', () => {
  assert.equal(fs.existsSync(sharedModalPath), true, 'shared StockKlineModal component should exist')
  assert.match(overview, /defineAsyncComponent\(\(\) => import\(['"]\.\/StockKlineModal\.vue['"]\)\)/)
  assert.match(overview, /<StockKlineModal/)
  assert.doesNotMatch(overview, /EventsEmit\(['"]klineSelectStock['"]/)
  assert.doesNotMatch(overview, /router\.push\(\{name:\s*['"]klineAnalysis['"]/)
  assert.match(patternScreener, /import StockKlineModal from ['"]\.\/StockKlineModal\.vue['"]/)
  assert.match(patternScreener, /<StockKlineModal/)
})
