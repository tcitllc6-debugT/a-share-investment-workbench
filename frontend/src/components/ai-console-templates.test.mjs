import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const source = readFileSync(new URL('./AIConsole.vue', import.meta.url), 'utf8')

test('AI console exposes prompt templates as a dedicated local workspace', () => {
  assert.match(source, /PromptTemplateList/)
  assert.match(source, /view === 'templates'/)
  assert.match(source, /本地提示词模板/)
  assert.doesNotMatch(source, /提示词广场/)
})
