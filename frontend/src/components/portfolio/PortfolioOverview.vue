<script setup>
import { computed } from 'vue'
import FXQuoteBoard from './FXQuoteBoard.vue'
import { getActionDisplayLabel, getAssetDisplayLabel } from './asset-form-schema.mjs'
import { formatMoney, formatNumber, formatPercent, formatTime, readValue } from './portfolio-api.mjs'

const props = defineProps({
  overview: { type: Object, default: () => ({}) },
  loading: { type: Boolean, default: false },
})
const emit = defineEmits(['select-asset'])

const metrics = computed(() => [
  { label: '总市值', value: formatMoney(readValue(props.overview, 'marketValueCny', 'MarketValueCNY', 'totalMarketValueCny', 'TotalMarketValueCNY')), hint: '全部资产折算人民币' },
  { label: '持仓成本', value: formatMoney(readValue(props.overview, 'totalCostCny', 'TotalCostCNY', 'totalCost', 'TotalCost')), hint: '交易账本累计成本' },
  { label: '浮动盈亏', value: formatMoney(readValue(props.overview, 'floatingPnlCny', 'FloatingPnLCNY')), hint: formatPercent(readValue(props.overview, 'floatingPnlPct', 'FloatingPnLPct')), signed: true },
  { label: '已实现盈亏', value: formatMoney(readValue(props.overview, 'realizedPnlCny', 'RealizedPnLCNY')), hint: `净现金流 ${formatMoney(readValue(props.overview, 'netCashFlowCny', 'NetCashFlowCNY'))}`, signed: true },
])

function normaliseSeries(value, labelKeys, valueKeys) {
  if (Array.isArray(value)) {
    return value.map((item) => ({
      key: String(readValue(item, ...labelKeys) || 'other'),
      value: Number(readValue(item, ...valueKeys) || 0),
      percent: Number(readValue(item, 'percent', 'Percent', 'ratio', 'Ratio') || 0),
    }))
  }
  if (value && typeof value === 'object') return Object.entries(value).map(([key, amount]) => ({ key, value: Number(readValue(amount, ...valueKeys) ?? amount ?? 0), percent: Number(readValue(amount, 'allocationPct', 'AllocationPct') || 0) }))
  return []
}

const allocations = computed(() => {
  const source = readValue(props.overview, 'assetAllocation', 'AssetAllocation', 'allocations', 'Allocations', 'byAssetType', 'ByAssetType')
  const rows = normaliseSeries(source, ['assetType', 'AssetType', 'key', 'Key'], ['marketValueCny', 'MarketValueCNY', 'value', 'Value']).filter((row) => row.value > 0)
  const total = rows.reduce((sum, row) => sum + Math.max(0, row.value), 0)
  return rows.map((row) => ({ ...row, percent: row.percent || (total ? row.value / total * 100 : 0) })).sort((a, b) => b.value - a.value)
})

const recentTransactions = computed(() => readValue(props.overview, 'recentTransactions', 'RecentTransactions', 'transactions', 'Transactions') || [])

function metricClass(metric) {
  if (!metric.signed) return ''
  const raw = metric.label === '浮动盈亏'
    ? readValue(props.overview, 'floatingPnlCny', 'FloatingPnLCNY')
    : readValue(props.overview, 'realizedPnlCny', 'RealizedPnLCNY')
  return Number(raw) > 0 ? 'is-profit' : Number(raw) < 0 ? 'is-loss' : ''
}
</script>

<template>
  <div class="portfolio-overview" :aria-busy="loading">
    <section class="portfolio-metrics" aria-label="持仓关键指标">
      <div v-for="metric in metrics" :key="metric.label">
        <span>{{ metric.label }}</span>
        <strong :class="metricClass(metric)">{{ loading ? '—' : metric.value }}</strong>
        <small>{{ metric.hint }}</small>
      </div>
    </section>

    <div class="portfolio-overview__grid">
      <section class="overview-panel overview-panel--allocation">
        <header><div><h2>资产配置</h2><p>各类金融产品的人民币市值占比</p></div><span>{{ allocations.length }} 类资产</span></header>
        <div v-if="loading" class="panel-loading"><i /><i /><i /></div>
        <div v-else-if="allocations.length" class="allocation-list">
          <button v-for="item in allocations" :key="item.key" type="button" @click="emit('select-asset', item.key)">
            <span><b>{{ getAssetDisplayLabel(item.key) }}</b><small>{{ formatMoney(item.value) }}</small></span>
            <i><em :style="{ width: `${Math.min(100, item.percent)}%` }" /></i>
            <strong>{{ item.percent.toFixed(1) }}%</strong>
          </button>
        </div>
        <div v-else class="panel-empty">录入交易后，这里会显示资产配置。</div>
      </section>

      <FXQuoteBoard />

      <section class="overview-panel overview-panel--transactions">
        <header><div><h2>最近交易</h2><p>跨资产交易账本的最新变动</p></div></header>
        <div class="recent-table" tabindex="0" aria-label="最近交易，可横向滚动">
          <table>
            <colgroup>
              <col class="recent-table__time"><col class="recent-table__asset"><col class="recent-table__product"><col class="recent-table__action"><col class="recent-table__quantity"><col class="recent-table__amount">
            </colgroup>
            <thead><tr><th>时间</th><th>资产</th><th>产品</th><th>动作</th><th>数量</th><th>金额</th></tr></thead>
            <tbody>
              <tr v-if="recentTransactions.length === 0"><td colspan="6" class="panel-empty">暂无交易记录。</td></tr>
              <tr v-for="(row, index) in recentTransactions" :key="readValue(row, 'id', 'ID') || index">
                <td>{{ formatTime(readValue(row, 'tradedAt', 'TradedAt')) }}</td>
                <td>{{ getAssetDisplayLabel(readValue(row, 'assetType', 'AssetType')) }}</td>
                <td><b>{{ readValue(row, 'instrumentName', 'InstrumentName') || '--' }}</b><small>{{ readValue(row, 'instrumentCode', 'InstrumentCode') || '' }}</small></td>
                <td>{{ getActionDisplayLabel(readValue(row, 'action', 'Action')) }}</td>
                <td>{{ formatNumber(readValue(row, 'quantity', 'Quantity')) }}</td>
                <td>{{ formatMoney(readValue(row, 'amountCny', 'AmountCNY', 'amount', 'Amount')) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped>
.portfolio-overview { --overview-label-size:12px;--overview-body-size:13px;--portfolio-table-font-size:13px;min-width:0;display:grid;gap:12px;color:var(--workspace-text);font-family:var(--workspace-font),"Microsoft YaHei UI","PingFang SC",sans-serif;letter-spacing:0;-webkit-font-smoothing:antialiased;text-rendering:optimizeLegibility; }
.portfolio-metrics { min-width:0;display:grid;grid-template-columns:repeat(4,minmax(0,1fr));overflow:hidden;border:1px solid var(--workspace-border);border-radius:var(--workspace-radius-md);background:var(--workspace-surface-solid);box-shadow:var(--workspace-shadow-sm); }
.portfolio-metrics>div { min-width:0;padding:14px 16px;border-right:1px solid var(--workspace-border); }.portfolio-metrics>div:last-child { border-right:0; }
.portfolio-metrics span,.portfolio-metrics small,.portfolio-metrics strong { display:block;overflow:hidden;text-overflow:ellipsis;white-space:nowrap; }.portfolio-metrics span { color:var(--workspace-text-muted);font-size:10px; }.portfolio-metrics strong { margin:6px 0 3px;color:var(--workspace-text);font-family:var(--workspace-number-font);font-size:20px;font-variant-numeric:tabular-nums; }.portfolio-metrics small { color:var(--workspace-text-muted);font-size:9px; }.portfolio-metrics strong.is-profit { color:var(--workspace-positive); }.portfolio-metrics strong.is-loss { color:var(--workspace-negative); }
.portfolio-overview__grid { min-width:0;display:grid;grid-template-columns:minmax(0,1.15fr) minmax(260px,.85fr);gap:12px; }
.overview-panel { min-width:0;overflow:hidden;border:1px solid var(--workspace-border);border-radius:var(--workspace-radius-md);background:var(--workspace-surface-solid);box-shadow:var(--workspace-shadow-sm); }.overview-panel>header { min-height:52px;display:flex;align-items:center;justify-content:space-between;gap:12px;padding:9px 13px;border-bottom:1px solid var(--workspace-border);background:color-mix(in srgb,var(--workspace-surface-muted) 44%,var(--workspace-surface-solid)); }.overview-panel h2 { margin:0;color:var(--workspace-text);font-size:13px; }.overview-panel header p { margin:3px 0 0;color:var(--workspace-text-muted);font-size:9px; }.overview-panel header>span { color:var(--workspace-text-muted);font-size:9px; }.overview-panel header button { min-height:30px;padding:0 9px;border:1px solid var(--workspace-border);border-radius:var(--workspace-radius-xs);color:var(--workspace-accent);background:var(--workspace-surface-solid);font-size:10px;cursor:pointer; }.overview-panel header button:hover { border-color:var(--workspace-accent);background:var(--workspace-accent-soft); }
.allocation-list { padding:5px 12px 9px; }.allocation-list button { width:100%;display:grid;grid-template-columns:minmax(120px,1fr) minmax(100px,1.5fr) 50px;align-items:center;gap:12px;padding:9px 2px;border:0;border-bottom:1px solid var(--workspace-border);color:var(--workspace-text);background:transparent;text-align:left;cursor:pointer; }.allocation-list button:last-child { border-bottom:0; }.allocation-list button:hover { color:var(--workspace-accent); }.allocation-list span b,.allocation-list span small { display:block; }.allocation-list span b { font-size:11px; }.allocation-list span small { margin-top:3px;color:var(--workspace-text-muted);font-family:var(--workspace-number-font);font-size:9px; }.allocation-list button>i { height:5px;overflow:hidden;border-radius:999px;background:var(--workspace-surface-muted); }.allocation-list button>i em { display:block;height:100%;border-radius:inherit;background:var(--workspace-accent); }.allocation-list button>strong { text-align:right;font-family:var(--workspace-number-font);font-size:10px; }
.overview-panel--transactions { grid-column:1/-1; }
.recent-table { max-height:260px;overflow:auto;outline:0; }.recent-table:focus-visible { box-shadow:inset var(--workspace-focus-ring); }.recent-table table { width:100%;min-width:720px;table-layout:fixed;border-collapse:collapse;font-size:var(--portfolio-table-font-size); }.recent-table__time { width:16%; }.recent-table__asset { width:12%; }.recent-table__product { width:30%; }.recent-table__action { width:12%; }.recent-table__quantity { width:12%; }.recent-table__amount { width:18%; }.recent-table th { position:sticky;top:0;padding:9px 10px;color:var(--workspace-text-secondary);background:var(--workspace-surface-muted);font-size:12px;font-weight:700;text-align:center; }.recent-table td { padding:10px;border-top:1px solid var(--workspace-border);color:var(--workspace-text-secondary);font-family:var(--workspace-number-font);font-size:var(--portfolio-table-font-size);font-weight:500;text-align:center;white-space:nowrap;overflow:hidden;text-overflow:ellipsis; }.recent-table th:first-child,.recent-table td:first-child { text-align:center; }.recent-table th:nth-child(2),.recent-table td:nth-child(2),.recent-table th:nth-child(4),.recent-table td:nth-child(4) { text-align:center; }.recent-table th:nth-child(3),.recent-table td:nth-child(3) { text-align:left; }.recent-table th:nth-child(5),.recent-table th:nth-child(6),.recent-table td:nth-child(5),.recent-table td:nth-child(6) { text-align:right; }.recent-table td b,.recent-table td small { display:block;overflow:hidden;text-overflow:ellipsis; }.recent-table td b { color:var(--workspace-text);font-family:var(--workspace-font);font-size:13px;font-weight:700; }.recent-table td small { margin-top:3px;color:var(--workspace-text-secondary);font-size:11px; }
.panel-empty { padding:28px 14px;color:var(--workspace-text-muted);font-size:10px;text-align:center; }.panel-loading { padding:12px; }.panel-loading i { display:block;height:10px;margin:9px 0;border-radius:4px;background:var(--workspace-surface-muted); }
.portfolio-metrics span { color:var(--workspace-text-secondary);font-size:var(--overview-label-size);font-weight:650; }
.portfolio-metrics strong { font-size:22px;font-weight:700; }
.portfolio-metrics small { color:var(--workspace-text-secondary);font-size:var(--overview-body-size); }
.overview-panel>header { min-height:56px;padding:10px 14px; }
.overview-panel h2 { font-size:15px;font-weight:700; }
.overview-panel header p { color:var(--workspace-text-secondary);font-size:var(--overview-body-size);line-height:1.5; }
.overview-panel header>span,.overview-panel header button { font-size:var(--overview-label-size);font-weight:650; }
.allocation-list span b,.allocation-list button>strong { font-size:var(--overview-body-size);font-weight:700; }
.allocation-list span b,.recent-table td:nth-child(2),.recent-table td:nth-child(4) { font-family:var(--workspace-font),"Microsoft YaHei UI","PingFang SC",sans-serif; }
.allocation-list span small { color:var(--workspace-text-secondary);font-size:var(--overview-body-size); }
.recent-table table { font-size:var(--overview-body-size); }
.recent-table th { color:var(--workspace-text-secondary);font-weight:700; }
.recent-table td b { font-weight:700; }
.recent-table td small,.panel-empty { color:var(--workspace-text-secondary);font-size:var(--overview-label-size); }
@media (max-width:980px), (min-resolution:2.5dppx) { .portfolio-metrics { grid-template-columns:repeat(2,minmax(0,1fr)); }.portfolio-metrics>div:nth-child(2) { border-right:0; }.portfolio-metrics>div:nth-child(-n+2) { border-bottom:1px solid var(--workspace-border); }.portfolio-overview__grid { grid-template-columns:minmax(0,1fr); }.overview-panel--transactions { grid-column:auto; } }
@media (max-width:520px) { .portfolio-metrics { grid-template-columns:minmax(0,1fr); }.portfolio-metrics>div { border-right:0;border-bottom:1px solid var(--workspace-border); } }
@media (min-resolution:2.5dppx) { .portfolio-overview { --overview-label-size:13px;--overview-body-size:13px; }.portfolio-metrics strong { font-size:23px; }.overview-panel h2 { font-size:16px; } }
</style>
