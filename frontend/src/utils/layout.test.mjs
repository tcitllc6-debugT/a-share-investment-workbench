import assert from 'node:assert/strict'
import { getLayoutMode } from './layout.mjs'

assert.equal(getLayoutMode(640), 'compact')
assert.equal(getLayoutMode(1000), 'standard')
assert.equal(getLayoutMode(1500), 'wide')

console.log('layout mode tests passed')
