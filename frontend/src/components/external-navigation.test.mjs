import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'

const source = readFileSync(new URL('stockhotmap.vue', import.meta.url), 'utf8')

test('curated finance sites open only after an explicit user action', () => {
  assert.doesNotMatch(source, /<iframe|EmbeddedUrl|gushitong\.baidu\.com/)
  assert.match(source, /OpenURL\(site\.url\)/)
})
