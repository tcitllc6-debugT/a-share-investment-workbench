function getAppApi() {
  return globalThis?.window?.go?.main?.App || null
}

export async function callPortfolioApi(method, ...args) {
  const api = getAppApi()
  if (!api || typeof api[method] !== 'function') {
    const error = new Error(`Wails API ${method} 尚未接入`)
    error.code = 'API_NOT_READY'
    throw error
  }
  return api[method](...args)
}

export function unwrapList(result, keys = []) {
  if (Array.isArray(result)) return result
  for (const key of keys) {
    if (Array.isArray(result?.[key])) return result[key]
  }
  return []
}

export function readValue(source, ...keys) {
  for (const key of keys) {
    const value = source?.[key]
    if (value !== undefined && value !== null) return value
  }
  return undefined
}

export function formatMoney(value, currency = 'CNY') {
  const numeric = Number(value)
  if (!Number.isFinite(numeric)) return '--'
  return new Intl.NumberFormat('zh-CN', {
    style: 'currency', currency, currencyDisplay: 'narrowSymbol', maximumFractionDigits: 2,
  }).format(numeric)
}

export function formatNumber(value, maximumFractionDigits = 4) {
  const numeric = Number(value)
  if (!Number.isFinite(numeric)) return '--'
  return new Intl.NumberFormat('zh-CN', { maximumFractionDigits }).format(numeric)
}

export function formatPercent(value) {
  const numeric = Number(value)
  if (!Number.isFinite(numeric)) return '--'
  return `${numeric >= 0 ? '+' : ''}${numeric.toFixed(2)}%`
}

export function formatTime(value) {
  if (!value) return '--'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return String(value)
  return new Intl.DateTimeFormat('zh-CN', {
    month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', hour12: false,
  }).format(date)
}
