<script setup>
import { computed } from 'vue'
import QuoteStatus from './QuoteStatus.vue'
import { ASSET_TYPE_MAP } from './asset-form-schema.mjs'
import { formatMoney, formatNumber, formatPercent, readValue } from './portfolio-api.mjs'

const props = defineProps({
  assetType: { type: String, required: true },
  positions: { type: Array, default: () => [] },
  loading: { type: Boolean, default: false },
})
const emit = defineEmits(['transact', 'refresh-quotes', 'manual-quote'])

const extraColumns = {
  stock: [{ key: 'exchange', label: '市场', width: 82 }],
  fund: [{ key: 'fundKind', label: '基金类型', width: 98 }],
  bond: [
    { key: 'couponRate', label: '票息', width: 76, formatter: (value) => value === undefined ? '--' : `${Number(value).toFixed(2)}%` },
    { key: 'maturityDate', label: '到期日', width: 100 },
  ],
  precious_metal: [{ key: 'unit', label: '单位', width: 72 }, { key: 'purity', label: '成色', width: 84 }],
  forex: [{ key: 'pair', label: '货币对', width: 92 }],
  future: [{ key: 'side', label: '方向', width: 72 }, { key: 'expiryDate', label: '到期日', width: 100 }],
  crypto: [{ key: 'network', label: '网络 / 平台', width: 104 }],
  bank_product: [{ key: 'institution', label: '发行机构', width: 112 }, { key: 'maturityDate', label: '到期日', width: 100 }],
}

const assetMeta = computed(() => ASSET_TYPE_MAP[props.assetType] || ASSET_TYPE_MAP.stock)
const specificColumns = computed(() => extraColumns[props.assetType] || [])

function read(row, key) {
  const pascalKey = key.charAt(0).toUpperCase() + key.slice(1)
  if (key === 'pair') {
    const direct = readValue(row, 'pair', 'Pair')
    if (direct) return direct
    const base = readValue(row, 'baseCurrency', 'BaseCurrency')
    const quote = readValue(row, 'quoteCurrency', 'QuoteCurrency')
    return [base, quote].filter(Boolean).join('/') || undefined
  }
  if (key === 'instrumentName') return readValue(row, key, pascalKey, 'name', 'Name')
  if (key === 'instrumentCode') return readValue(row, key, pascalKey, 'symbol', 'Symbol')
  if (key === 'unrealizedPnlCny') return readValue(row, key, pascalKey, 'floatingPnlCny', 'FloatingPnLCNY')
  if (key === 'returnRate') return readValue(row, key, pascalKey, 'floatingPnlPct', 'FloatingPnLPct')
  const direct = readValue(row, key, pascalKey)
  if (direct !== undefined) return direct
  return readValue(row?.metadata, key, pascalKey)
}

function rowIdentity(row, index) {
  return readValue(row, 'positionId', 'PositionID', 'instrumentId', 'InstrumentID', 'instrumentCode', 'InstrumentCode') || index
}

function profitClass(value) {
  const number = Number(value)
  if (!Number.isFinite(number) || number === 0) return ''
  return number > 0 ? 'is-profit' : 'is-loss'
}

function requestTransaction(action, row) {
  emit('transact', { action, position: row })
}
</script>

<template>
  <section class="asset-position-table" :aria-busy="loading">
    <header class="asset-position-table__header">
      <div>
        <h2>{{ assetMeta.label }}持仓</h2>
        <p>{{ assetMeta.description }}</p>
      </div>
      <button type="button" class="text-action" :disabled="loading" @click="emit('refresh-quotes')">更新报价</button>
    </header>

    <div class="asset-position-table__scroller" tabindex="0" :aria-label="`${assetMeta.label}持仓表，可横向滚动`">
      <table>
        <colgroup>
          <col class="instrument-column">
          <col v-for="index in 6" :key="`numeric-${index}`" class="numeric-column">
          <col v-for="column in specificColumns" :key="`specific-${column.key}`" :style="{ width: `${column.width}px` }">
          <col class="quote-column">
          <col class="action-column">
        </colgroup>
        <thead>
          <tr>
            <th class="instrument-column">产品</th>
            <th class="numeric-column">数量</th>
            <th class="numeric-column">平均成本</th>
            <th class="numeric-column">最新价</th>
            <th class="numeric-column">人民币市值</th>
            <th class="numeric-column">浮动盈亏</th>
            <th class="numeric-column">收益率</th>
            <th v-for="column in specificColumns" :key="column.key" class="center-column" :style="{ minWidth: `${column.width}px` }">{{ column.label }}</th>
            <th class="quote-column">报价</th>
            <th class="action-column">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading" v-for="index in 4" :key="`skeleton-${index}`" class="skeleton-row">
            <td colspan="12"><i /></td>
          </tr>
          <tr v-else-if="positions.length === 0">
            <td colspan="12" class="empty-state">
              <strong>还没有{{ assetMeta.label }}持仓</strong>
              <span>点击右上角“增加一笔持仓”，先搜索产品再录入成交信息。</span>
            </td>
          </tr>
          <tr v-for="(row, index) in positions" v-else :key="rowIdentity(row, index)">
            <td class="instrument-cell">
              <b>{{ read(row, 'instrumentName') || '--' }}</b>
              <span>{{ read(row, 'instrumentCode') || read(row, 'productCode') || '--' }} · {{ read(row, 'currency') || 'CNY' }}</span>
            </td>
            <td class="numeric-cell">{{ formatNumber(read(row, 'quantity')) }}</td>
            <td class="numeric-cell">{{ formatMoney(read(row, 'averageCost'), read(row, 'currency') || 'CNY') }}</td>
            <td class="numeric-cell">{{ formatNumber(read(row, 'latestPrice')) }}</td>
            <td class="numeric-cell numeric-cell--strong">{{ formatMoney(read(row, 'marketValueCny')) }}</td>
            <td class="numeric-cell" :class="profitClass(read(row, 'unrealizedPnlCny'))">{{ formatMoney(read(row, 'unrealizedPnlCny')) }}</td>
            <td class="numeric-cell" :class="profitClass(read(row, 'returnRate'))">{{ formatPercent(read(row, 'returnRate')) }}</td>
            <td v-for="column in specificColumns" :key="column.key" class="center-cell">
              {{ column.formatter ? column.formatter(read(row, column.key)) : (read(row, column.key) ?? '--') }}
            </td>
            <td class="quote-cell"><QuoteStatus :quote="row" compact /></td>
            <td class="action-cell"><div class="action-buttons">
              <button v-if="assetType === 'bank_product'" type="button" class="is-manual" @click="emit('manual-quote', row)">更新净值</button>
              <button type="button" class="is-add" @click="requestTransaction('add', row)">加仓</button>
              <button type="button" class="is-reduce" @click="requestTransaction('reduce', row)">减仓</button>
              <button type="button" class="is-danger" @click="requestTransaction('close', row)">平仓</button>
            </div></td>
          </tr>
        </tbody>
      </table>
    </div>
    <footer>
      <span>金额统一折算为人民币；原币价格与汇率由报价服务保留。</span>
      <span>共 {{ positions.length }} 项持仓</span>
    </footer>
  </section>
</template>

<style scoped>
.asset-position-table { --portfolio-table-font-size:13px;min-width:0;border:1px solid var(--workspace-border);border-radius:var(--workspace-radius-md);background:var(--workspace-surface-solid);box-shadow:var(--workspace-shadow-sm);overflow:hidden;color:var(--workspace-text);font-family:var(--workspace-font),"Microsoft YaHei UI","PingFang SC",sans-serif;letter-spacing:0;-webkit-font-smoothing:antialiased;text-rendering:optimizeLegibility; }
.asset-position-table__header { min-height:56px;display:flex;align-items:center;justify-content:space-between;gap:16px;padding:10px 14px;border-bottom:1px solid var(--workspace-border);background:color-mix(in srgb,var(--workspace-surface-muted) 44%,var(--workspace-surface-solid)); }
.asset-position-table__header h2 { margin:0;color:var(--workspace-text);font-size:16px;font-weight:700; }
.asset-position-table__header p { margin:3px 0 0;color:var(--workspace-text-secondary);font-size:12px;line-height:1.5; }
.text-action { min-height:32px;padding:0 10px;border:1px solid var(--workspace-border);border-radius:var(--workspace-radius-sm);color:var(--workspace-accent);background:var(--workspace-surface-solid);font-size:12px;font-weight:650;cursor:pointer; }
.text-action:hover { border-color:var(--workspace-accent);background:var(--workspace-accent-soft); }
.text-action:disabled { opacity:.5;cursor:wait; }
.asset-position-table__scroller { width:100%;max-height:min(58dvh,680px);overflow:auto;overscroll-behavior:contain;scrollbar-gutter:stable;outline:none; }
.asset-position-table__scroller:focus-visible { box-shadow:inset var(--workspace-focus-ring); }
table { width:100%;min-width:1180px;table-layout:fixed;border-collapse:separate;border-spacing:0;color:var(--workspace-text-secondary);font-size:var(--portfolio-table-font-size); }
th { position:sticky;top:0;z-index:2;height:40px;padding:8px 10px;border-bottom:1px solid var(--workspace-border);color:var(--workspace-text-secondary);background:var(--workspace-surface-muted);font-size:12px;font-weight:700;text-align:center;white-space:nowrap; }
td { height:52px;padding:8px 10px;border-bottom:1px solid var(--workspace-border);background:var(--workspace-surface-solid);font-size:var(--portfolio-table-font-size);font-weight:500;text-align:center;white-space:nowrap; }
tbody tr:hover td { background:var(--workspace-surface-hover); }
tbody tr:last-child td { border-bottom:0; }
.instrument-column { width:220px;text-align:left; }.numeric-column { width:112px;text-align:right; }.quote-column { width:92px;text-align:center; }.center-column,.center-cell,.quote-cell { text-align:center; }
.instrument-cell b,.instrument-cell span { display:block;max-width:220px;overflow:hidden;text-overflow:ellipsis; }
.instrument-cell b { color:var(--workspace-text);font-size:13px;font-weight:700; }.instrument-cell span { margin-top:3px;color:var(--workspace-text-secondary);font-family:var(--workspace-number-font);font-size:11px; }
.numeric-cell { text-align:right;font-family:var(--workspace-number-font);font-variant-numeric:tabular-nums; }.numeric-cell--strong { color:var(--workspace-text);font-weight:700; }
.is-profit { color:var(--workspace-positive); }.is-loss { color:var(--workspace-negative); }
.action-column { position:sticky;right:0;z-index:3;width:202px;background:var(--workspace-surface-muted); }.action-cell { position:sticky;right:0;z-index:1;background:var(--workspace-surface-solid);box-shadow:-10px 0 18px -16px rgba(23,48,81,.48); }.action-buttons { display:flex;align-items:center;justify-content:center;gap:6px; }
tbody tr:hover .action-cell { background:var(--workspace-surface-hover); }
.action-cell button { min-height:28px;padding:0 8px;border:1px solid transparent;border-radius:var(--workspace-radius-xs);font-size:11px;font-weight:650;cursor:pointer;transition:background .16s ease,border-color .16s ease,transform .16s ease; }.action-cell button:hover { transform:translateY(-1px); }.action-cell button.is-manual { color:var(--workspace-accent);border-color:color-mix(in srgb,var(--workspace-accent) 25%,var(--workspace-border));background:var(--workspace-accent-soft); }.action-cell button.is-add { color:var(--workspace-negative);border-color:color-mix(in srgb,var(--workspace-negative) 28%,var(--workspace-border));background:color-mix(in srgb,var(--workspace-negative) 10%,var(--workspace-surface-solid)); }.action-cell button.is-reduce { color:var(--workspace-warning);border-color:color-mix(in srgb,var(--workspace-warning) 32%,var(--workspace-border));background:color-mix(in srgb,var(--workspace-warning) 11%,var(--workspace-surface-solid)); }.action-cell button.is-danger { color:var(--workspace-positive);border-color:color-mix(in srgb,var(--workspace-positive) 28%,var(--workspace-border));background:color-mix(in srgb,var(--workspace-positive) 9%,var(--workspace-surface-solid)); }
.empty-state { height:220px;text-align:center;white-space:normal; }.empty-state strong,.empty-state span { display:block; }.empty-state strong { color:var(--workspace-text);font-size:13px; }.empty-state span { margin-top:6px;color:var(--workspace-text-muted);font-size:10px; }
.skeleton-row td { height:48px;padding:12px; }.skeleton-row i { display:block;width:100%;height:10px;border-radius:4px;background:linear-gradient(90deg,var(--workspace-surface-muted),var(--workspace-surface-hover),var(--workspace-surface-muted));background-size:200% 100%;animation:skeleton 1.2s ease-in-out infinite; }
footer { min-height:38px;display:flex;align-items:center;justify-content:space-between;gap:16px;padding:7px 12px;border-top:1px solid var(--workspace-border);color:var(--workspace-text-secondary);background:var(--workspace-surface-muted);font-size:11px; }
@keyframes skeleton { to { background-position:-200% 0; } }
@media (max-width:719px), (min-resolution:2.5dppx) { .asset-position-table__header { align-items:flex-start; }.asset-position-table__header p { max-width:52ch;white-space:normal; }.asset-position-table__scroller { max-height:52dvh; } footer { align-items:flex-start;flex-direction:column;gap:3px; } }
@media (prefers-reduced-motion:reduce) { .skeleton-row i { animation:none; } }
</style>
