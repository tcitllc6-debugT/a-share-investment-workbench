import test from 'node:test'
import assert from 'node:assert/strict'
import {
  buildDuoKongLineData,
  emaValues,
} from './duoKongLine.mjs'

function assertClose(actual, expected) {
  assert.ok(Math.abs(actual - expected) < 1e-12, `${actual} should be close to ${expected}`)
}

test('emaValues starts from the first finite value and applies standard EMA smoothing', () => {
  const values = emaValues([10, 12, 14, 13], 3)

  assert.deepEqual(values, [10, 11, 12.5, 12.75])
})

test('buildDuoKongLineData calculates only two colored lines from screenshot-derived formula', () => {
  const candles = [
    { time: 1, open: 10, high: 12, low: 9, close: 11 },
    { time: 2, open: 11, high: 13, low: 10, close: 12 },
    { time: 3, open: 8, high: 9, low: 7, close: 9 },
    { time: 4, open: 9, high: 10, low: 8, close: 10 },
  ]

  const data = buildDuoKongLineData(candles)

  assert.deepEqual(Object.keys(data).sort(), ['fine', 'thick'])

  assertClose(data.fine[0].value, 11)
  assert.equal(data.fine[0].color, '#ef4444')
  assertClose(data.fine[1].value, 11.181818181818182)
  assert.equal(data.fine[1].color, '#ef4444')
  assertClose(data.fine[2].value, 10.78512396694215)
  assert.equal(data.fine[2].color, '#22c55e')

  assertClose(data.thick[0].value, 10.6)
  assert.equal(data.thick[0].color, '#ef4444')
  assertClose(data.thick[1].value, 10.664516129032258)
  assert.equal(data.thick[1].color, '#ef4444')
  assertClose(data.thick[2].value, 10.518418314255984)
  assert.equal(data.thick[2].color, '#22c55e')
})
