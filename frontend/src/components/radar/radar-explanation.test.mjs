import test from 'node:test'
import assert from 'node:assert/strict'

import { describeRadarEvent, radarHowItWorks } from './radar-explanation.mjs'

test('explains an opportunity event in reviewable language', () => {
  const detail = describeRadarEvent({ category: 'momentum', changeType: 8201, typeName: '火箭发射' })
  assert.match(detail.trigger, /火箭发射/)
  assert.match(detail.nextAction, /K线/)
})

test('distinguishes concrete momentum triggers inside the same category', () => {
  const rocket = describeRadarEvent({ category: 'momentum', changeType: 8201 })
  const rebound = describeRadarEvent({ category: 'momentum', changeType: 8202 })
  assert.notEqual(rocket.trigger, rebound.trigger)
  assert.match(rebound.trigger, /快速反弹/)
})

test('explains risk and capital events without treating them as trade orders', () => {
  assert.match(describeRadarEvent({ category: 'risk' }).trigger, /风险/)
  assert.match(describeRadarEvent({ category: 'capital' }).trigger, /资金/)
  assert.match(radarHowItWorks, /不构成买卖指令/)
})
