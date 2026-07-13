import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'

const source = readFileSync(new URL('./settings.vue', import.meta.url), 'utf8')

test('general settings no longer owns AI model configuration', () => {
  assert.doesNotMatch(source, /AI模型服务配置/)
  assert.doesNotMatch(source, /AI设置/)
  assert.doesNotMatch(source, /formValue\.openAI/)
  assert.doesNotMatch(source, /addAiConfig/)
  assert.doesNotMatch(source, /FetchAiModels/)
  assert.doesNotMatch(source, /FetchAiModelInfo/)
  assert.doesNotMatch(source, /aiSettingsRef/)
})

test('general settings uses a non-AI save boundary and sanitized model choices', () => {
  assert.match(source, /UpdateGeneralConfig/)
  assert.match(source, /GetAiConfigs/)
  assert.doesNotMatch(source, /aiConfigs:\s*formValue/)
  assert.doesNotMatch(source, /openAiEnable:/)
})
