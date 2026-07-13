export function canonicalStockSymbol(value) {
  const raw = String(value || '').trim().toLowerCase().replace(/\s+/g, '')
  const suffixMatch = raw.match(/^(\d{5,6})\.(sh|sz|bj|hk|us)$/)
  if (suffixMatch) {
    const [, digits, market] = suffixMatch
    if (market === 'us') return `gb_${digits}`
    return `${market}${digits}`
  }
  if (/^(sh|sz|bj)\d{6}$/.test(raw) || /^hk\d{5}$/.test(raw) || /^gb_.+/.test(raw)) return raw
  if (/^us.+/.test(raw)) return `gb_${raw.slice(2)}`
  return raw
}

export function watchlistSymbols(followedStocks = []) {
  return [...new Set(followedStocks
    .map((stock) => canonicalStockSymbol(stock?.StockCode ?? stock))
    .filter(Boolean))]
}

function stockDigits(value) {
  const symbol = canonicalStockSymbol(value)
  const mainlandMatch = symbol.match(/^(?:sh|sz|bj)(\d{6})$/)
  const hkMatch = symbol.match(/^hk(\d{5})$/)
  if (mainlandMatch) return mainlandMatch[1]
  if (hkMatch) return hkMatch[1]
  return /^\d{5,6}$/.test(symbol) ? symbol : ''
}

export function isWatchlistMember(followedSymbols, quoteSymbol) {
  const symbol = canonicalStockSymbol(quoteSymbol)
  if (followedSymbols.includes(symbol)) return true

  const digits = stockDigits(symbol)
  return Boolean(digits) && followedSymbols.some((followedSymbol) => stockDigits(followedSymbol) === digits)
}

export function findFollowedStock(followedStocks = [], quoteSymbol) {
  return followedStocks.find((stock) => isWatchlistMember(
    [canonicalStockSymbol(stock?.StockCode ?? stock)],
    quoteSymbol,
  ))
}

export function buildWatchlistFallbackRow(stock = {}) {
  const code = canonicalStockSymbol(stock.StockCode ?? stock.stockCode)
  const price = Number(stock.Price) || 0
  return {
    '股票代码': code,
    '股票名称': stock.Name || stock.name || code,
    '当前价格': price,
    '卖一报价': price,
    '昨日收盘价': price,
    changePercent: 0,
    sort: Number(stock.Sort) || 0,
    quotePending: true,
  }
}
