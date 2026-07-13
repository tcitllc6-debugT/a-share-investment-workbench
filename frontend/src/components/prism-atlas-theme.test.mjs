import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'

const root = new URL('../', import.meta.url)
const read = (name) => readFileSync(new URL(name, root), 'utf8')

test('Prism Atlas tokens define paired white-blue and midnight themes', () => {
  const tokens = read('styles/tokens.css')
  const base = read('styles/base.css')

  assert.match(tokens, /:root\[data-theme-mode="light"\]/)
  assert.match(tokens, /:root\[data-theme-mode="dark"\]/)
  assert.match(tokens, /--workspace-accent:\s*#3478f6/i)
  assert.match(tokens, /--workspace-glass/)
  assert.match(base, /prefers-reduced-motion/)
  assert.match(base, /scrollbar-color/)
})

test('the app applies persisted theme state before mounting Vue', () => {
  const main = read('main.js')
  assert.match(main, /THEME_MODE_STORAGE_KEY/)
  assert.match(main, /applyThemeMode/)
  assert.ok(main.indexOf('applyThemeMode') < main.indexOf("app.mount('#app')"))
})

test('settings keeps update checks manual and exposes both theme choices', () => {
  const settings = read('components/settings.vue')

  assert.match(settings, /检查更新/)
  assert.match(settings, /CheckManagedUpdate/)
  assert.match(settings, /ApplyManagedUpdate/)
  assert.match(settings, /日间主题/)
  assert.match(settings, /暗黑主题/)
  assert.match(settings, /updateSettings/)
  const app = read('App.vue')
  assert.match(app, /EventsOn\("updateSettings"/)
})
