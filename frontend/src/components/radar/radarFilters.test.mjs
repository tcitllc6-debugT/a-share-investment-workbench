import test from 'node:test'
import assert from 'node:assert/strict'
import { buildRadarFilter } from './radarFilters.mjs'

test('builds a watchlist-only risk filter', () => {
  assert.deepEqual(buildRadarFilter({tab: 'risk', onlyFollowed: true}).categories, ['risk', 'breakdown'])
  assert.equal(buildRadarFilter({tab: 'risk', onlyFollowed: true}).onlyFollowed, true)
})
