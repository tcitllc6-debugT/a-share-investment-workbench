export const ASSET_TYPES = [
  { value: 'stock', label: '股票', shortLabel: '股票', description: 'A股与其他已录入股票持仓' },
  { value: 'fund', label: '基金 / ETF', shortLabel: '基金ETF', description: '场内 ETF、场外基金与公募产品' },
  { value: 'bond', label: '债券', shortLabel: '债券', description: '国债、可转债与信用债' },
  { value: 'precious_metal', label: '贵金属', shortLabel: '贵金属', description: '黄金、白银等贵金属资产' },
  { value: 'forex', label: '外汇', shortLabel: '外汇', description: '主要货币与人民币折算敞口' },
  { value: 'future', label: '期货', shortLabel: '期货', description: '商品与金融期货合约' },
  { value: 'crypto', label: '加密货币', shortLabel: '加密货币', description: '按公开行情折算的数字资产' },
  { value: 'bank_product', label: '银行理财', shortLabel: '银行理财', description: '净值型与本金计价理财产品' },
]

export const ASSET_TYPE_MAP = Object.fromEntries(ASSET_TYPES.map((item) => [item.value, item]))

export const ASSET_DISPLAY_LABELS = Object.freeze({
  stock: '股票', fund: '基金 / ETF', bond: '债券', precious_metal: '贵金属',
  forex: '外汇', future: '期货', futures: '期货', crypto: '加密货币',
  bank_product: '银行理财', bank_wealth: '银行理财', other: '其他资产',
})

export const ACTION_DISPLAY_LABELS = Object.freeze({
  buy: '买入', sell: '卖出', close: '清仓', subscribe: '申购', redeem: '赎回',
  mature: '到期结算', open_long: '开多', open_short: '开空', close_long: '平多', close_short: '平空',
})

export const CURRENCY_DISPLAY_LABELS = Object.freeze({
  CNY: '人民币', USD: '美元', HKD: '港币', EUR: '欧元', JPY: '日元',
  GBP: '英镑', AUD: '澳元', CAD: '加元', CHF: '瑞士法郎', SGD: '新加坡元',
  NZD: '新西兰元', KRW: '韩元', THB: '泰铢', MYR: '马来西亚林吉特',
  TRY: '土耳其里拉', BRL: '巴西雷亚尔', INR: '印度卢比', ZAR: '南非兰特',
  USDT: '美元稳定币', USDC: '美元稳定币', FDUSD: '美元稳定币',
})

export function getAssetDisplayLabel(value) {
  return ASSET_DISPLAY_LABELS[String(value || '').toLowerCase()] || '其他资产'
}

export function getActionDisplayLabel(value) {
  return ACTION_DISPLAY_LABELS[String(value || '').toLowerCase()] || '其他操作'
}

export function getCurrencyDisplayLabel(value) {
  return CURRENCY_DISPLAY_LABELS[String(value || '').toUpperCase()] || '其他币种'
}

const COMMON_FIELDS = [
  { key: 'instrumentCode', label: '产品代码', type: 'text', required: true, placeholder: '例如 600519.SH' },
  { key: 'instrumentName', label: '产品名称', type: 'text', required: true, placeholder: '输入便于识别的名称' },
  { key: 'quoteSymbol', label: '行情代码', type: 'text', placeholder: '可选，例如 nf_AU0、BTCUSDT' },
  { key: 'action', label: '交易动作', type: 'select', required: true },
  { key: 'quantity', label: '数量', type: 'number', required: true, min: 0, step: 'any' },
  { key: 'price', label: '成交价 / 净值', type: 'number', required: true, min: 0, step: 'any' },
  { key: 'currency', label: '计价货币', type: 'select', required: true },
  { key: 'fee', label: '费用', type: 'number', min: 0, step: 'any' },
  { key: 'tradedAt', label: '交易时间', type: 'datetime-local', required: true },
  { key: 'note', label: '备注', type: 'textarea', wide: true, placeholder: '记录交易依据、渠道或补充信息' },
]

const SPECIFIC_FIELDS = {
  stock: [
    { key: 'exchange', label: '交易市场', type: 'select', options: ['上海', '深圳', '北京', '香港', '美国'] },
  ],
  fund: [
    { key: 'fundKind', label: '基金类型', type: 'select', options: ['ETF', 'LOF', '开放式基金', '封闭式基金', '其他'] },
  ],
  bond: [
    { key: 'faceValue', label: '每张面值', type: 'number', min: 0, step: 'any' },
    { key: 'couponRate', label: '票面利率 (%)', type: 'number', min: 0, step: 'any' },
    { key: 'accruedInterest', label: '应计利息', type: 'number', min: 0, step: 'any' },
    { key: 'couponFrequency', label: '付息频率', type: 'select', options: ['到期一次', '每年', '每半年', '每季度'] },
    { key: 'maturityDate', label: '到期日', type: 'date' },
  ],
  precious_metal: [
    { key: 'unit', label: '计量单位', type: 'select', required: true, options: ['克', '千克', '盎司'] },
    { key: 'purity', label: '纯度', type: 'text', placeholder: '例如 Au99.99' },
  ],
  forex: [
    { key: 'baseCurrency', label: '基础货币', type: 'text', required: true, placeholder: '例如 USD' },
    { key: 'quoteCurrency', label: '报价货币', type: 'text', required: true, placeholder: '例如 CNY' },
  ],
  future: [
    { key: 'side', label: '持仓方向', type: 'select', required: true, options: ['多头', '空头'] },
    { key: 'multiplier', label: '合约乘数', type: 'number', required: true, min: 0, step: 'any' },
    { key: 'marginRatio', label: '保证金比例 (%)', type: 'number', min: 0, step: 'any' },
    { key: 'settlementPrice', label: '结算价', type: 'number', min: 0, step: 'any' },
    { key: 'expiryDate', label: '到期日', type: 'date' },
  ],
  crypto: [
    { key: 'network', label: '网络 / 托管平台', type: 'text', placeholder: '例如 Bitcoin、Ethereum 或交易所名称' },
  ],
  bank_product: [
    { key: 'institution', label: '发行机构', type: 'text', required: true },
    { key: 'valuationMode', label: '估值方式', type: 'select', required: true, options: ['本金计价', '净值计价'] },
    { key: 'principal', label: '投入本金', type: 'number', min: 0, step: 'any' },
    { key: 'nav', label: '最新净值', type: 'number', min: 0, step: 'any' },
    { key: 'expectedYield', label: '业绩比较基准 (%)', type: 'number', min: 0, step: 'any' },
    { key: 'maturityDate', label: '到期日', type: 'date' },
  ],
}

const ACTIONS = {
  stock: [['buy', '买入 / 加仓'], ['sell', '卖出 / 减仓'], ['close', '清仓']],
  fund: [['buy', '申购 / 加仓'], ['sell', '赎回 / 减仓'], ['close', '全部赎回']],
  bond: [['buy', '买入 / 加仓'], ['sell', '卖出 / 减仓'], ['redeem', '到期兑付']],
  precious_metal: [['buy', '买入 / 加仓'], ['sell', '卖出 / 减仓'], ['close', '全部卖出']],
  forex: [['buy', '买入 / 增持'], ['sell', '卖出 / 减持'], ['close', '平掉敞口']],
  future: [['open_long', '开多'], ['open_short', '开空'], ['close_long', '平多'], ['close_short', '平空']],
  crypto: [['buy', '买入 / 增持'], ['sell', '卖出 / 减持'], ['close', '全部卖出']],
  bank_product: [['subscribe', '申购 / 追加'], ['redeem', '赎回 / 减少'], ['mature', '到期结算']],
}

export const CURRENCY_OPTIONS = [
  'CNY', 'USD', 'HKD', 'EUR', 'JPY', 'GBP', 'AUD', 'CAD', 'CHF', 'SGD',
  'NZD', 'KRW', 'THB', 'MYR', 'TRY', 'BRL', 'INR', 'ZAR',
].map((value) => ({ value, label: CURRENCY_DISPLAY_LABELS[value] }))

export function getActionOptions(assetType) {
  return (ACTIONS[assetType] || ACTIONS.stock).map(([value, label]) => ({ value, label }))
}

export function getAssetFormFields(assetType) {
  return [...COMMON_FIELDS, ...(SPECIFIC_FIELDS[assetType] || [])].map((field) => {
    if (field.key === 'action') return { ...field, options: getActionOptions(assetType) }
    if (field.key === 'currency') return { ...field, options: CURRENCY_OPTIONS }
    return field
  })
}

export function createAssetTransactionDraft(assetType, position = null, requestedAction = '') {
  const now = new Date(Date.now() - new Date().getTimezoneOffset() * 60000).toISOString().slice(0, 16)
  const defaultAction = requestedAction || getActionOptions(assetType)[0]?.value || 'buy'
  const draft = {
    id: 0,
    assetType,
    instrumentId: '',
    instrumentCode: '',
    instrumentName: '',
    quoteSymbol: '',
    action: defaultAction,
    quantity: null,
    price: null,
    currency: assetType === 'crypto' ? 'USD' : 'CNY',
    fee: 0,
    tradedAt: now,
    note: '',
  }
  if (!position) return draft

  const read = (...keys) => keys.map((key) => position?.[key]).find((value) => value !== undefined && value !== null)
  return {
    ...draft,
    instrumentId: read('instrumentId', 'InstrumentID', 'instrumentID', 'ID') || '',
    instrumentCode: read('instrumentCode', 'InstrumentCode', 'code', 'Code') || '',
    instrumentName: read('instrumentName', 'InstrumentName', 'name', 'Name') || '',
    currency: read('currency', 'Currency') || 'CNY',
    price: Number(read('latestPrice', 'LatestPrice', 'price', 'Price') || 0) || null,
    quantity: requestedAction === 'close' ? Number(read('quantity', 'Quantity') || 0) || null : null,
    ...Object.fromEntries(Object.keys(position).filter((key) => key in draft).map((key) => [key, position[key]])),
  }
}

export function validateAssetTransaction(fields, draft) {
  const errors = {}
  fields.forEach((field) => {
    const value = draft[field.key]
    if (field.required && (value === '' || value === null || value === undefined)) {
      errors[field.key] = `请填写${field.label}`
      return
    }
    if (field.type === 'number' && value !== '' && value !== null && value !== undefined) {
      const numericValue = Number(value)
      if (!Number.isFinite(numericValue)) errors[field.key] = `${field.label}必须是有效数字`
      else if (field.min !== undefined && numericValue < field.min) errors[field.key] = `${field.label}不能小于 ${field.min}`
    }
  })
  return errors
}
