import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const source = readFileSync(new URL('./researchIndex.vue', import.meta.url), 'utf8')

assert.match(source, /AIConsole/)
assert.doesNotMatch(source, /PromptTemplateList/)
assert.doesNotMatch(source, /TradingRecordManager/)
assert.doesNotMatch(source, /MCPServiceManager/)

console.log('legacy research route compatibility checks passed')
