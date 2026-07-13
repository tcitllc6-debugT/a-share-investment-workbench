function firstDefined(source, keys) {
  for (const key of keys) {
    if (source?.[key] !== undefined && source?.[key] !== null) return source[key]
  }
  return undefined
}

function finiteOrNull(value) {
  if (value === '' || value === undefined || value === null) return null
  const number = Number(value)
  return Number.isFinite(number) ? number : null
}

export function normaliseFollowedStock(source = {}) {
  return {
    stockCode: String(firstDefined(source, ['StockCode', 'stockCode', 'code']) || '').trim(),
    name: String(firstDefined(source, ['Name', 'name', 'stockName']) || '').trim(),
    price: finiteOrNull(firstDefined(source, ['Price', 'price', 'latestPrice'])),
    changePercent: finiteOrNull(firstDefined(source, ['ChangePercent', 'changePercent', 'changeRate'])),
  }
}

export function mergeWatchlistQuotes(followedStocks = [], realTimeQuotes = []) {
  const quotesByCode = new Map(realTimeQuotes.map((quote) => {
    const row = normaliseFollowedStock({
      StockCode: firstDefined(quote, ['Code', 'code', 'StockCode', 'stockCode']),
      Name: firstDefined(quote, ['Name', 'name', 'StockName', 'stockName']),
      Price: firstDefined(quote, ['Price', 'price', 'latestPrice']),
      ChangePercent: firstDefined(quote, ['ChangePercent', 'changePercent', 'changeRate']),
    })
    return [row.stockCode.toLowerCase(), row]
  }))

  return followedStocks.map((stock) => {
    const row = normaliseFollowedStock(stock)
    const quote = quotesByCode.get(row.stockCode.toLowerCase())
    if (!quote) return row
    return {
      ...row,
      name: quote.name || row.name,
      price: quote.price ?? row.price,
      changePercent: quote.changePercent ?? row.changePercent,
    }
  })
}

export function normaliseNewsItem(source = {}) {
  return {
    id: firstDefined(source, ['ID', 'id']) ?? `${firstDefined(source, ['Time', 'time']) || ''}-${firstDefined(source, ['Content', 'content']) || ''}`,
    time: String(firstDefined(source, ['Time', 'time']) || '--:--'),
    content: String(firstDefined(source, ['Content', 'content', 'Title', 'title']) || '').trim(),
    source: String(firstDefined(source, ['Source', 'source']) || '').trim(),
  }
}

function newsContentKey(item) {
  return String(item?.content || '')
    .toLowerCase()
    .replace(/\s+/g, '')
    .replace(/[，。！？、；：,.!?;:'"“”‘’（）()【】\[\]]/g, '')
}

export function interleaveNewsSources(sourceLists = [], limit = 6) {
  const queues = sourceLists
    .map(list => (Array.isArray(list) ? list : []).map(normaliseNewsItem).filter(item => item.content))
    .filter(list => list.length)
  const result = []
  const seen = new Set()
  let cursor = 0

  while (result.length < limit && queues.some(queue => cursor < queue.length)) {
    for (const queue of queues) {
      const item = queue[cursor]
      if (!item) continue
      const key = newsContentKey(item)
      if (!key || seen.has(key)) continue
      seen.add(key)
      result.push(item)
      if (result.length >= limit) break
    }
    cursor += 1
  }
  return result
}

export function latestMarketPulse(rows = []) {
  const source = Array.isArray(rows) && rows.length ? rows[rows.length - 1] : {}
  return {
    upCount: Number(source.upCount || 0),
    downCount: Number(source.downCount || 0),
    limitUp: Number(source.limitUp || 0),
    limitDown: Number(source.limitDown || 0),
    upRatio: Number(source.upRatio || 0),
    sentimentDesc: String(source.sentimentDesc || '等待统计'),
    dataTime: String(source.dataTime || ''),
  }
}

export async function loadDashboardResources(resources, onValue, onError = () => {}) {
  const tasks = Object.entries(resources).map(async ([key, load]) => {
    try {
      const value = await load()
      onValue(key, value)
      return { key, status: 'fulfilled' }
    } catch (error) {
      onError(key, error)
      return { key, status: 'rejected', error }
    }
  })
  return Promise.all(tasks)
}
