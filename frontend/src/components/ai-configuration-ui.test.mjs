import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'

const read = (name) => readFileSync(new URL(name, import.meta.url), 'utf8')

test('AI configuration is a complete independent control surface', () => {
  const source = read('./AIConfiguration.vue')
  assert.match(source, /GetAIWorkspaceConfig/)
  assert.match(source, /SaveAIWorkspaceConfig/)
  assert.match(source, /TestAIConnection/)
  assert.match(source, /选择服务商/)
  assert.match(source, /API 密钥/)
  assert.match(source, /获取模型/)
  assert.match(source, /测试连接/)
  assert.match(source, /默认研究模型/)
  assert.match(source, /默认智能体模型/)
  assert.match(source, /高级设置/)
  assert.match(source, /研究偏好/)
  assert.match(source, /密钥只保存在本机/)
  assert.match(source, /dialog\.warning/)
  assert.match(source, /workspace\.value\.enabled = true/)
  assert.doesNotMatch(source, /Crawler Timeout/)
})

test('AI configuration supports compact and high-DPI layouts', () => {
  const source = read('./AIConfiguration.vue')
  assert.match(source, /grid-template-columns/)
  assert.match(source, /overflow-y:\s*auto/)
  assert.match(source, /@media\s*\(max-width:\s*1100px\)/)
  assert.match(source, /@media\s*\(max-width:\s*720px\)/)
  assert.match(source, /var\(--workspace-/)
})

test('AI console owns configuration as its third view', () => {
  const source = read('./AIConsole.vue')
  assert.match(source, /import AIConfiguration/)
  assert.match(source, /switchView\('configuration'\)/)
  assert.match(source, /view === 'configuration'/)
  assert.match(source, /<AIConfiguration/)
  assert.doesNotMatch(source, /name: 'settings'/)
})
