import test from 'node:test'
import assert from 'node:assert/strict'
import { getNavigationPresentation, getTableScrollWidth } from './navigation.mjs'

test('uses wide, standard, and drawer navigation presentations at shared breakpoints', () => {
  assert.equal(getNavigationPresentation(1280), 'wide')
  assert.equal(getNavigationPresentation(960), 'standard')
  assert.equal(getNavigationPresentation(800), 'drawer')
  assert.equal(getNavigationPresentation(480), 'drawer')
})

test('keeps wide tables reachable when high-DPI viewports narrow', () => {
  assert.equal(getTableScrollWidth('watchlist', 800), 980)
  assert.equal(getTableScrollWidth('screening', 1024), 860)
  assert.equal(getTableScrollWidth('radar', 1600), 1340)
})
