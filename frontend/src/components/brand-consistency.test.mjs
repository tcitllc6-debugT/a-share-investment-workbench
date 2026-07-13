import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'

const read = (path) => readFileSync(new URL(path, import.meta.url), 'utf8')

test('desktop, browser and shell branding use A股智投工作台', () => {
  const app = read('../App.vue')
  const stock = read('stock.vue')
  const settings = read('settings.vue')
  const index = read('../../index.html')
  const wails = read('../../../wails.json')

  assert.match(app, /WindowSetTitle\("A股智投工作台 /)
  assert.match(index, /<title>A股智投工作台<\/title>/)
  assert.match(wails, /"name": "A股智投工作台"/)
  assert.match(wails, /"productName": "A股智投工作台"/)
  assert.doesNotMatch(stock, /A-Share Desk/)
  assert.doesNotMatch(settings, /A-Share Desk/)
  assert.doesNotMatch(stock, /title: '发现新版本/)
})
