import assert from 'node:assert/strict'
import test from 'node:test'
import { applyThemeMode, persistThemeMode, resolveThemeMode } from './theme-mode.mjs'

test('stored mode wins over the legacy config preference', () => {
  assert.equal(resolveThemeMode('dark', 'light'), 'light')
  assert.equal(resolveThemeMode('light', 'dark'), 'dark')
  assert.equal(resolveThemeMode('dark', null), 'dark')
})

test('persist and apply write the Prism Atlas theme marker', () => {
  const writes = []
  const storage = { setItem: (...args) => writes.push(args) }
  const root = {
    dataset: {},
    setAttribute: (key, value) => { root[key] = value },
  }

  persistThemeMode('dark', storage)
  applyThemeMode('dark', { documentElement: root })

  assert.deepEqual(writes, [['a-share-theme-mode', 'dark']])
  assert.equal(root.dataset.themeMode, 'dark')
  assert.equal(root['data-theme-mode'], 'dark')
})
