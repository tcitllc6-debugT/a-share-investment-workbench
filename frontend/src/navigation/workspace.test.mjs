import test from 'node:test'
import assert from 'node:assert/strict'

import {
  PRIMARY_SECTIONS,
  buildSecondaryLocation,
  getSecondaryItems,
  getSectionForRoute,
  getWorkspacePresentation,
} from './workspace.mjs'

test('exposes the approved workspaces with About at the far end', () => {
  assert.deepEqual(
    PRIMARY_SECTIONS.map((section) => section.key),
    ['overview', 'watchlist', 'market', 'screening', 'holding', 'ai', 'settings', 'about'],
  )
})

test('keeps legacy routes in the correct visible workspace', () => {
  assert.equal(getSectionForRoute({ name: 'dashboard' }), 'overview')
  assert.equal(getSectionForRoute({ name: 'stock' }), 'watchlist')
  assert.equal(getSectionForRoute({ name: 'klineAnalysis' }), 'watchlist')
  assert.equal(getSectionForRoute({ name: 'market' }), 'market')
  assert.equal(getSectionForRoute({ name: 'screening' }), 'screening')
  assert.equal(getSectionForRoute({ name: 'radar' }), 'screening')
  assert.equal(getSectionForRoute({ name: 'research' }), 'ai')
  assert.equal(getSectionForRoute({ name: 'agent' }), 'ai')
  assert.equal(getSectionForRoute({ name: 'fund' }), 'holding')
  assert.equal(getSectionForRoute({ name: 'cronTasks' }), 'settings')
  assert.equal(getSectionForRoute({ name: 'mcpServers' }), 'settings')
  assert.equal(getSectionForRoute({ name: 'about' }), 'about')
})

test('makes all market, screening, and AI destinations available as contextual items', () => {
  assert.equal(getSecondaryItems('market').length, 13)
  assert.deepEqual(
    getSecondaryItems('screening').map((item) => item.key),
    ['pattern', 'duo-kong', 'strategy', 'radar'],
  )
  assert.deepEqual(
    getSecondaryItems('ai').map((item) => item.key),
    ['analysis', 'recommendations', 'templates', 'configuration'],
  )
})

test('only lists settings destinations that are currently implemented', () => {
  assert.deepEqual(getSecondaryItems('settings').map((item) => item.key), ['general'])
})

test('builds query-driven secondary locations without discarding existing query values', () => {
  assert.deepEqual(
    buildSecondaryLocation('market', 'telegraph', { source: 'overview' }),
    { name: 'market', query: { source: 'overview', tab: 'telegraph' } },
  )
  assert.deepEqual(
    buildSecondaryLocation('screening', 'duo-kong', { board: 'star' }),
    { name: 'screening', query: { board: 'star', mode: 'duo-kong' } },
  )
  assert.deepEqual(
    buildSecondaryLocation('ai', 'configuration', { from: 'nav' }),
    { name: 'aiConsole', query: { from: 'nav', view: 'configuration' } },
  )
  assert.deepEqual(
    buildSecondaryLocation('ai', 'templates', { from: 'nav' }),
    { name: 'aiConsole', query: { from: 'nav', view: 'templates' } },
  )
})

test('uses stable responsive presentations for normal and high-DPI viewports', () => {
  assert.deepEqual(getWorkspacePresentation(1600), {
    sidebarMode: 'wide', density: 'comfortable', secondaryWidth: 216,
  })
  assert.deepEqual(getWorkspacePresentation(1024), {
    sidebarMode: 'standard', density: 'compact', secondaryWidth: 184,
  })
  assert.deepEqual(getWorkspacePresentation(800), {
    sidebarMode: 'drawer', density: 'compact', secondaryWidth: 296,
  })
  assert.deepEqual(getWorkspacePresentation(390), {
    sidebarMode: 'drawer', density: 'compact', secondaryWidth: 296,
  })
})
