import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'

const source = readFileSync(new URL('settings.vue', import.meta.url), 'utf8')

test('settings no longer persists or validates sponsor codes', () => {
  assert.doesNotMatch(source, /CheckSponsorCode|sponsorCode|赞助码|VIP功能/)
})
