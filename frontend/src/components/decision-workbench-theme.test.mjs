import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'

const source = readFileSync(new URL('./DecisionWorkbench.vue', import.meta.url), 'utf8')
const styles = source.slice(source.indexOf('<style scoped>'))

test('decision workbench uses Prism Atlas surfaces in both themes', () => {
  assert.match(styles, /background:\s*var\(--workspace-canvas\)/)
  assert.match(styles, /background:\s*var\(--workspace-surface-solid\)/)
  assert.match(styles, /color:\s*var\(--workspace-text\)/)
  assert.match(styles, /border[^;}]*var\(--workspace-border\)/)
  assert.doesNotMatch(styles, /background:#fff(?:fff)?(?:[;}])/i)
  assert.doesNotMatch(styles, /background:#f4f6f7/i)
})
