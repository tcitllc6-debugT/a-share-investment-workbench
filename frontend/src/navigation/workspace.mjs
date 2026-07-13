const MARKET_ITEMS = [
  ['telegraph', '市场快讯'],
  ['global-index', '全球股指'],
  ['major-index', '重大指数'],
  ['industry-rank', '行业排名'],
  ['stock-flow', '个股资金流向'],
  ['sector-flow', '板块资金流向'],
  ['concept-flow', '概念资金流向'],
  ['dragon-tiger', '龙虎榜'],
  ['stock-research', '个股研报'],
  ['company-notice', '公司公告'],
  ['industry-research', '行业研究'],
  ['hot-now', '当前热门'],
  ['featured-sites', '名站优选'],
].map(([key, label]) => ({ key, label, routeName: 'market', queryKey: 'tab' }))

const SECTION_ITEMS = {
  overview: [{ key: 'home', label: '总览', routeName: 'dashboard' }],
  watchlist: [
    { key: 'all', label: '全部自选', routeName: 'stock' },
    { key: 'kline', label: 'K 线分析', routeName: 'klineAnalysis' },
  ],
  market: MARKET_ITEMS,
  screening: [
    { key: 'pattern', label: '形态选股', routeName: 'screening', queryKey: 'mode' },
    { key: 'duo-kong', label: '多空线选股', routeName: 'screening', queryKey: 'mode' },
    { key: 'strategy', label: '策略选股', routeName: 'screening', queryKey: 'mode' },
    { key: 'radar', label: '盘中异动雷达', routeName: 'radar' },
  ],
  holding: [
    { key: 'portfolio-overview', label: '持仓概览', routeName: 'holdingCenter', queryKey: 'assetType', queryValue: 'overview' },
    { key: 'stock-holdings', label: '股票持仓', routeName: 'holdingCenter', queryKey: 'assetType', queryValue: 'stock' },
    { key: 'fund-holdings', label: '基金 / ETF', routeName: 'holdingCenter', queryKey: 'assetType', queryValue: 'fund' },
    { key: 'bond-holdings', label: '债券持仓', routeName: 'holdingCenter', queryKey: 'assetType', queryValue: 'bond' },
    { key: 'metal-holdings', label: '贵金属持仓', routeName: 'holdingCenter', queryKey: 'assetType', queryValue: 'precious_metal' },
    { key: 'forex-holdings', label: '外汇持仓', routeName: 'holdingCenter', queryKey: 'assetType', queryValue: 'forex' },
    { key: 'future-holdings', label: '期货持仓', routeName: 'holdingCenter', queryKey: 'assetType', queryValue: 'future' },
    { key: 'crypto-holdings', label: '加密货币持仓', routeName: 'holdingCenter', queryKey: 'assetType', queryValue: 'crypto' },
    { key: 'bank-holdings', label: '银行理财', routeName: 'holdingCenter', queryKey: 'assetType', queryValue: 'bank_product' },
  ],
  ai: [
    { key: 'analysis', label: 'AI 分析记录', routeName: 'aiConsole', queryKey: 'view' },
    { key: 'recommendations', label: '股票推荐记录', routeName: 'aiConsole', queryKey: 'view' },
    { key: 'templates', label: '提示词模板', routeName: 'aiConsole', queryKey: 'view' },
    { key: 'configuration', label: 'AI 配置', routeName: 'aiConsole', queryKey: 'view' },
  ],
  settings: [
    { key: 'general', label: '常规设置', routeName: 'settings' },
  ],
  about: [],
}

export const PRIMARY_SECTIONS = [
  { key: 'overview', label: '总览', routeName: 'dashboard', icon: 'LayoutDashboard' },
  { key: 'watchlist', label: '股票自选', routeName: 'stock', icon: 'Star' },
  { key: 'market', label: '市场资讯', routeName: 'market', icon: 'ChartNoAxesCombined' },
  { key: 'screening', label: '多维度选股', routeName: 'screening', icon: 'ScanSearch' },
  { key: 'holding', label: '持仓中心', routeName: 'holdingCenter', icon: 'WalletCards' },
  { key: 'ai', label: 'AI 控制台', routeName: 'aiConsole', icon: 'Sparkles' },
  { key: 'settings', label: '设置', routeName: 'settings', icon: 'Settings2' },
  { key: 'about', label: '关于', routeName: 'about', icon: 'CircleHelp' },
]

const ROUTE_SECTIONS = {
  dashboard: 'overview',
  stock: 'watchlist',
  klineAnalysis: 'watchlist',
  market: 'market',
  screening: 'screening',
  radar: 'screening',
  holdingCenter: 'holding',
  fund: 'holding',
  aiConsole: 'ai',
  research: 'ai',
  agent: 'ai',
  settings: 'settings',
  cronTasks: 'settings',
  mcpServers: 'settings',
  about: 'about',
}

export function getSectionForRoute(route) {
  return ROUTE_SECTIONS[route?.name] || 'overview'
}

export function getSecondaryItems(sectionKey) {
  return SECTION_ITEMS[sectionKey] || []
}

export function buildSecondaryLocation(sectionKey, itemKey, existingQuery = {}) {
  const item = getSecondaryItems(sectionKey).find((entry) => entry.key === itemKey)
  if (!item) return { name: PRIMARY_SECTIONS.find((section) => section.key === sectionKey)?.routeName || 'dashboard', query: { ...existingQuery } }

  const query = { ...existingQuery }
  if (item.queryKey) query[item.queryKey] = item.queryValue || itemKey
  return { name: item.routeName, query }
}

export function getWorkspacePresentation(viewportWidth) {
  if (viewportWidth < 960) return { sidebarMode: 'drawer', density: 'compact', secondaryWidth: 296 }
  if (viewportWidth < 1280) return { sidebarMode: 'standard', density: 'compact', secondaryWidth: 184 }
  return { sidebarMode: 'wide', density: 'comfortable', secondaryWidth: 216 }
}
