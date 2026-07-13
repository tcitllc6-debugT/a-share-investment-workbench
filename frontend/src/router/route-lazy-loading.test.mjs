import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'

const source = readFileSync(new URL('./router.js', import.meta.url), 'utf8')
const mainSource = readFileSync(new URL('../main.js', import.meta.url), 'utf8')
const viteSource = readFileSync(new URL('../../vite.config.js', import.meta.url), 'utf8')
const dashboardSource = readFileSync(new URL('../components/DecisionWorkbench.vue', import.meta.url), 'utf8')

test('only the startup dashboard is imported eagerly', () => {
  assert.match(source, /import DecisionWorkbench from/)
  assert.doesNotMatch(source, /import stockView from/)
  assert.doesNotMatch(source, /import marketView from/)
  assert.doesNotMatch(source, /import HoldingCenter from/)
  assert.doesNotMatch(source, /import AIConsole from/)
})

test('feature routes are loaded on demand', () => {
  for (const component of [
    'stock.vue',
    'market.vue',
    'ScreeningHub.vue',
    'HoldingCenter.vue',
    'AIConsole.vue',
    'settings.vue',
  ]) {
    assert.match(source, new RegExp(`\\(\\) => import\\([^\\n]*${component.replace('.', '\\.')}[^\\n]*\\)`))
	}
})

test('Naive UI components are resolved per view instead of globally registered', () => {
  assert.doesNotMatch(mainSource, /import naive from ['"]naive-ui['"]/)
  assert.doesNotMatch(mainSource, /app\.use\(naive\)/)
  assert.match(viteSource, /NaiveUiResolver/)
  assert.match(viteSource, /Components\([\s\S]*NaiveUiResolver\(\)/)
})

test('the full K-line workbench loads only after the dashboard modal is opened', () => {
  assert.doesNotMatch(dashboardSource, /import StockKlineModal from/)
  assert.match(dashboardSource, /defineAsyncComponent\(\(\) => import\(['"]\.\/StockKlineModal\.vue['"]\)\)/)
  assert.match(dashboardSource, /<StockKlineModal[\s\S]*v-if="klineModalShow"/)
})
