import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const directory = new URL('.', import.meta.url)
const read = (name) => readFileSync(new URL(name, directory), 'utf8')

const topNavigation = read('WorkspaceTopNavigation.vue')
const secondaryNavigation = read('WorkspaceNavigation.vue')
const header = read('WorkspaceHeader.vue')
const app = read('../../App.vue')
const tokens = read('../../styles/tokens.css')
const base = read('../../styles/base.css')

assert.match(topNavigation, /PRIMARY_SECTIONS/)
assert.match(topNavigation, /@click="emit\('toggle-secondary'\)"/)
assert.match(topNavigation, /aria-label="打开当前工作区导航"/)
assert.match(topNavigation, /RouterLink/)
assert.match(topNavigation, /emit\('search'/)
assert.match(topNavigation, /theme-change/)
assert.match(topNavigation, /aria-label="切换日间或暗黑主题"/)
assert.match(topNavigation, /SunnyOutline|MoonOutline/)

assert.match(secondaryNavigation, /getSectionForRoute/)
assert.match(secondaryNavigation, /getSecondaryItems/)
assert.match(secondaryNavigation, /buildSecondaryLocation/)
assert.match(secondaryNavigation, /normaliseMarketTab/)
assert.match(secondaryNavigation, /normaliseScreeningTab/)
assert.match(secondaryNavigation, /secondary-open/)
assert.match(secondaryNavigation, /aria-current/)
assert.match(secondaryNavigation, /NIcon|n-icon/)

assert.match(header, /workspace-header__context/)

assert.match(tokens, /--workspace-topbar-height: 60px/)
assert.match(tokens, /--workspace-secondary-wide: 220px/)
assert.match(tokens, /--workspace-secondary-standard: 188px/)
assert.match(tokens, /--workspace-focus-ring/)
assert.match(base, /box-sizing: border-box/)
assert.match(base, /:focus-visible/)
assert.match(base, /prefers-reduced-motion/)
assert.match(app, /@theme-change="setThemeMode"/)
assert.match(app, /EventsOn\("updateAvailable", showManagedUpdateNotification\)/)
assert.match(app, /发现新版本 v/)
assert.match(app, /ApplyManagedUpdate/)
assert.match(app, /workspace-shell--prism/)

console.log('workspace shell source-contract checks passed')
