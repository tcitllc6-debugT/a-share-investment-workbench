import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'

const source = readFileSync(new URL('./market.vue', import.meta.url), 'utf8')
const app = readFileSync(new URL('../App.vue', import.meta.url), 'utf8')

test('market quick-news first paint uses the lightweight local query', () => {
  assert.match(source, /GetTelegraphBriefList/)
  assert.match(source, /loadMarketNews/)
  assert.match(source, /showMarketAnalysis/)
})

test('market setup does not eagerly fetch indexes and industry ranks', () => {
  const setup = source.slice(source.indexOf('onBeforeMount(()'), source.indexOf('onBeforeUnmount(()'))
  assert.doesNotMatch(setup, /getIndex\(\)/)
  assert.doesNotMatch(setup, /industryRank\(\)/)
  assert.doesNotMatch(setup, /GetPromptTemplates/)
  assert.doesNotMatch(setup, /GetAiConfigs/)
})

test('market analysis waits until quick-news content can paint', () => {
  assert.match(source, /v-if="showMarketAnalysis"[^>]*AnalyzeMartket|<AnalyzeMartket[^>]*v-if="showMarketAnalysis"/)
})

test('startup does not force a full-screen skeleton before local content paints', () => {
  assert.match(app, /const loading = ref\(false\)/)
})
