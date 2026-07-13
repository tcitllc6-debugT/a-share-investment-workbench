export const MARKET_TABS = {
  telegraph: '市场快讯',
  'global-index': '全球股指',
  'major-index': '重大指数',
  'industry-rank': '行业排名',
  'stock-flow': '个股资金流向',
  'sector-flow': '板块资金流向',
  'concept-flow': '概念资金流向',
  'dragon-tiger': '龙虎榜',
  'stock-research': '个股研报',
  'company-notice': '公司公告',
  'industry-research': '行业研究',
  'hot-now': '当前热门',
  'featured-sites': '名站优选',
}

const SCREENING_TABS = new Set(['pattern', 'duo-kong', 'strategy'])
const AI_TABS = new Set(['analysis', 'recommendations'])

export function normaliseMarketTab(query = {}) {
  const candidate = query.tab || query.name
  if (MARKET_TABS[candidate]) return MARKET_TABS[candidate]
  if (Object.values(MARKET_TABS).includes(candidate)) return candidate
  return MARKET_TABS.telegraph
}

export function toMarketTabKey(label) {
  return Object.entries(MARKET_TABS).find(([, value]) => value === label)?.[0] || 'telegraph'
}

export function normaliseScreeningTab(query = {}) {
  const candidate = query.mode || query.tab
  return SCREENING_TABS.has(candidate) ? candidate : 'pattern'
}

export function normaliseAIConsoleTab(query = {}) {
  return AI_TABS.has(query.view) ? query.view : 'analysis'
}
