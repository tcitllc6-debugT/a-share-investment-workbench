export function emaValues(values, period) {
  const result = []
  const alpha = 2 / (period + 1)
  let previous = null

  for (const raw of values || []) {
    const value = Number(raw)
    if (!Number.isFinite(value)) {
      result.push(null)
      continue
    }

    previous = previous == null ? value : alpha * value + (1 - alpha) * previous
    result.push(previous)
  }

  return result
}

const RISE_COLOR = '#ef4444'
const FALL_COLOR = '#22c55e'

function toColoredLineData(times, values) {
  const line = []
  for (let i = 0; i < times.length; i++) {
    const time = times[i]
    const value = values[i]
    if (!Number.isFinite(value)) {
      line.push({ time })
      continue
    }

    const previous = values[i - 1]
    const isRising = i === 0 || !Number.isFinite(previous) || value >= previous
    line.push({
      time,
      value,
      color: isRising ? RISE_COLOR : FALL_COLOR,
    })
  }

  return line
}

export function buildDuoKongLineData(candles) {
  const rows = Array.isArray(candles) ? candles : []
  const times = []
  const closes = []
  const weightedPrices = []

  for (const candle of rows) {
    const time = candle?.time
    const open = Number(candle?.open)
    const high = Number(candle?.high)
    const low = Number(candle?.low)
    const close = Number(candle?.close)
    if (time == null || ![open, high, low, close].every(Number.isFinite)) continue

    times.push(time)
    closes.push(close)
    weightedPrices.push((open + high + low + 2 * close) / 5)
  }

  const fine = toColoredLineData(times, emaValues(closes, 10))
  const thick = toColoredLineData(times, emaValues(weightedPrices, 30))

  return {
    fine,
    thick,
  }
}
