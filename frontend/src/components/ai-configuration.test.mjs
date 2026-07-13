import assert from 'node:assert/strict'
import test from 'node:test'

import {
  AI_PROVIDER_PRESETS,
  connectionResultTone,
  createAIService,
  endpointChanged,
  preferredAIService,
  validateAIService,
} from './ai-configuration.mjs'

test('ships understandable provider presets including local Ollama', () => {
  const keys = AI_PROVIDER_PRESETS.map((item) => item.key)
  assert.ok(keys.includes('deepseek'))
  assert.ok(keys.includes('dashscope'))
  assert.ok(keys.includes('volcengine'))
  assert.ok(keys.includes('ollama'))
  assert.ok(keys.includes('custom'))
})

test('creates safe defaults for hosted and local services', () => {
  const deepseek = createAIService('deepseek')
  assert.equal(deepseek.baseUrl, 'https://api.deepseek.com')
  assert.equal(deepseek.modelName, 'deepseek-chat')
  assert.equal(deepseek.timeOut, 300)
  assert.equal(deepseek.apiKey, '')

  const ollama = createAIService('ollama')
  assert.equal(ollama.baseUrl, 'http://127.0.0.1:11434/v1')
  assert.equal(ollama.apiKey, '')
})

test('validates the fields people can act on', () => {
  assert.deepEqual(validateAIService(createAIService('deepseek')), {
    apiKey: '请输入服务商提供的 API 密钥',
  })
  assert.deepEqual(validateAIService({
    ...createAIService('ollama'),
    modelName: '',
  }), {
    modelName: '请选择或输入模型名称',
  })

  assert.deepEqual(validateAIService({
    ...createAIService('ollama'),
    modelName: '',
  }, { requireModel: false }), {})
})

test('maps connection results to clear visual tones', () => {
  assert.equal(connectionResultTone({ success: true }), 'success')
  assert.equal(connectionResultTone({ reachable: true, manualModelAllowed: true }), 'warning')
  assert.equal(connectionResultTone({ code: 'invalid_key' }), 'error')
})

test('selects the configured default for each AI purpose', () => {
  const services = [
    { ID: 1, isDefaultResearch: false, isDefaultAgent: false },
    { ID: 2, isDefaultResearch: true, isDefaultAgent: false },
    { ID: 3, isDefaultResearch: false, isDefaultAgent: true },
  ]
  assert.equal(preferredAIService(services, 'research').ID, 2)
  assert.equal(preferredAIService(services, 'agent').ID, 3)
  assert.equal(preferredAIService([{ ID: 4 }], 'research').ID, 4)
  assert.equal(preferredAIService([], 'research'), null)
})

test('requires a fresh key when a saved remote endpoint changes', () => {
  const service = {
    ...createAIService('deepseek'),
    ID: 8,
    hasApiKey: true,
    _originalBaseUrl: 'https://api.deepseek.com',
    baseUrl: 'https://other-provider.example/v1',
  }
  assert.equal(endpointChanged(service), true)
  assert.deepEqual(validateAIService(service), {
    apiKey: '接口地址已更改，请重新输入 API 密钥',
  })

  service.baseUrl = 'https://api.deepseek.com/'
  assert.equal(endpointChanged(service), false)
  assert.deepEqual(validateAIService(service), {})
})
