import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const source = readFileSync(new URL('./DuoKongLineSelector.vue', import.meta.url), 'utf8')

assert.match(source, /const visibleColumns = computed\(/)
assert.match(source, /:columns="visibleColumns"/)
assert.match(source, /:scroll-x="isCompact \? 760 : 1800"/)
assert.match(source, /class="duo-kong-results"/)

console.log('duo kong compact layout regression checks passed')
