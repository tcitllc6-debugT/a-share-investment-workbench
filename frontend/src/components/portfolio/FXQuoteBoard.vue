<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { callPortfolioApi, formatNumber, formatPercent, formatTime, readValue } from './portfolio-api.mjs'

const quotes = ref([])
const loading = ref(true)
const refreshing = ref(false)
const error = ref('')
let refreshTimer = null

const currencyNames = {
  USD: '美元', HKD: '港元', EUR: '欧元', GBP: '英镑',
  JPY: '日元', AUD: '澳元', CAD: '加元', SGD: '新加坡元',
}

const rows = computed(() => quotes.value.map((quote) => {
  const base = String(readValue(quote, 'baseCurrency', 'BaseCurrency') || '').toUpperCase()
  const rate = Number(readValue(quote, 'rate', 'Rate') || 0)
  return {
    key: base,
    label: currencyNames[base] || base,
    pair: `${base}/CNY`,
    rate,
    changePercent: Number(readValue(quote, 'changePercent', 'ChangePercent') || 0),
    provider: String(readValue(quote, 'provider', 'Provider') || ''),
    fetchedAt: readValue(quote, 'fetchedAt', 'FetchedAt', 'asOf', 'AsOf'),
    points: sparklinePoints(readValue(quote, 'history', 'History') || [], rate),
  }
}))

function sparklinePoints(history, fallbackRate) {
  const values = history.map((point) => Number(readValue(point, 'rate', 'Rate'))).filter((value) => Number.isFinite(value) && value > 0)
  if (!values.length && fallbackRate > 0) values.push(fallbackRate)
  if (values.length === 1) values.push(values[0])
  if (!values.length) return ''
  const min = Math.min(...values)
  const max = Math.max(...values)
  const range = max - min || Math.max(max * 0.001, 0.0001)
  return values.map((value, index) => {
    const x = values.length === 1 ? 50 : index / (values.length - 1) * 100
    const y = 86 - (value - min) / range * 72
    return `${x.toFixed(1)},${y.toFixed(1)}`
  }).join(' ')
}

async function loadCached() {
  try {
    const data = await callPortfolioApi('GetMajorFXQuotes')
    quotes.value = Array.isArray(data) ? data : []
  } catch (cause) {
    error.value = cause?.message || '暂时无法读取外汇报价'
  } finally {
    loading.value = false
  }
}

async function refresh() {
  refreshing.value = true
  error.value = ''
  try {
    const data = await callPortfolioApi('RefreshMajorFXQuotes')
    quotes.value = Array.isArray(data) ? data : []
    if (!quotes.value.length) error.value = '报价源暂不可用，稍后会自动重试。'
  } catch (cause) {
    error.value = cause?.message || '外汇报价更新失败，稍后会自动重试。'
  } finally {
    loading.value = false
    refreshing.value = false
  }
}

onMounted(async () => {
  await loadCached()
  await refresh()
  refreshTimer = window.setInterval(refresh, 60_000)
})

onBeforeUnmount(() => {
  if (refreshTimer) window.clearInterval(refreshTimer)
})
</script>

<template>
  <section class="fx-board" aria-live="polite">
    <header class="fx-board__header">
      <div><h2>实时外汇报价</h2><p>主流货币兑人民币，每分钟自动更新</p></div>
      <button type="button" :disabled="refreshing" @click="refresh">{{ refreshing ? '更新中' : '刷新' }}</button>
    </header>
    <div v-if="loading && !rows.length" class="fx-board__loading"><i /><i /><i /><i /></div>
    <div v-else-if="rows.length" class="fx-board__grid">
      <article v-for="item in rows" :key="item.key" class="fx-quote" :class="{ 'is-up': item.changePercent > 0, 'is-down': item.changePercent < 0 }">
        <div class="fx-quote__top"><span>{{ item.label }}</span><small>{{ item.pair }}</small></div>
        <strong>1 {{ item.key }} = {{ formatNumber(item.rate, item.key === 'JPY' ? 5 : 4) }} 元</strong>
        <div class="fx-quote__footer">
          <span>{{ formatPercent(item.changePercent) }}</span>
          <svg viewBox="0 0 100 100" role="img" :aria-label="`${item.label}报价走势`" preserveAspectRatio="none"><polyline :points="item.points" /></svg>
        </div>
        <small class="fx-quote__meta">{{ formatTime(item.fetchedAt) }} · {{ item.provider || '公开报价' }}</small>
      </article>
    </div>
    <div v-else class="fx-board__empty">{{ error || '正在连接外汇报价源…' }}</div>
    <p v-if="error && rows.length" class="fx-board__notice">{{ error }}</p>
  </section>
</template>

<style scoped>
.fx-board{min-width:0;overflow:hidden;border:1px solid var(--workspace-border);border-radius:var(--workspace-radius-md);background:var(--workspace-surface-solid);box-shadow:var(--workspace-shadow-sm)}
.fx-board__header{min-height:56px;display:flex;align-items:center;justify-content:space-between;gap:12px;padding:10px 14px;border-bottom:1px solid var(--workspace-border);background:color-mix(in srgb,var(--workspace-surface-muted) 44%,var(--workspace-surface-solid))}.fx-board h2{margin:0;color:var(--workspace-text);font-size:15px;font-weight:700}.fx-board p{margin:3px 0 0;color:var(--workspace-text-secondary);font-size:12px;line-height:1.5}.fx-board__header button{min-height:30px;padding:0 10px;border:1px solid var(--workspace-border);border-radius:var(--workspace-radius-xs);color:var(--workspace-accent);background:var(--workspace-surface-solid);font-size:12px;font-weight:650;cursor:pointer}.fx-board__header button:disabled{opacity:.6;cursor:wait}.fx-board__header button:not(:disabled):hover{border-color:var(--workspace-accent);background:var(--workspace-accent-soft)}
.fx-board__grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));}.fx-quote{min-width:0;padding:11px 13px;border-right:1px solid var(--workspace-border);border-bottom:1px solid var(--workspace-border)}.fx-quote:nth-child(2n){border-right:0}.fx-quote:nth-last-child(-n+2){border-bottom:0}.fx-quote__top,.fx-quote__footer{display:flex;align-items:center;justify-content:space-between;gap:8px}.fx-quote__top>span{color:var(--workspace-text);font-size:12px;font-weight:700}.fx-quote__top small,.fx-quote__meta{color:var(--workspace-text-secondary);font-family:var(--workspace-number-font);font-size:10px}.fx-quote strong{display:block;margin:6px 0 7px;color:var(--workspace-text);font-family:var(--workspace-number-font);font-size:14px;font-variant-numeric:tabular-nums}.fx-quote__footer>span{color:var(--workspace-text-secondary);font-family:var(--workspace-number-font);font-size:11px;font-weight:650}.fx-quote.is-up .fx-quote__footer>span{color:var(--workspace-positive)}.fx-quote.is-down .fx-quote__footer>span{color:var(--workspace-negative)}.fx-quote svg{width:62px;height:17px;overflow:visible}.fx-quote polyline{fill:none;stroke:var(--workspace-accent);stroke-linecap:round;stroke-linejoin:round;stroke-width:6;vector-effect:non-scaling-stroke;transition:all .35s ease}.fx-quote.is-up polyline{stroke:var(--workspace-positive)}.fx-quote.is-down polyline{stroke:var(--workspace-negative)}.fx-quote__meta{display:block;margin-top:5px;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}.fx-board__loading{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:12px;padding:13px}.fx-board__loading i{height:76px;border-radius:6px;background:linear-gradient(90deg,var(--workspace-surface-muted),var(--workspace-surface-solid),var(--workspace-surface-muted));background-size:200% 100%;animation:fx-skeleton 1.2s linear infinite}.fx-board__empty,.fx-board__notice{margin:0;padding:25px 14px;color:var(--workspace-text-secondary);font-size:12px;text-align:center}.fx-board__notice{padding:7px 14px;border-top:1px solid var(--workspace-border);color:var(--workspace-warning);text-align:left}@keyframes fx-skeleton{to{background-position:-200% 0}}
@media(max-width:520px){.fx-board__grid{grid-template-columns:minmax(0,1fr)}.fx-quote,.fx-quote:nth-child(2n){border-right:0;border-bottom:1px solid var(--workspace-border)}.fx-quote:last-child{border-bottom:0}}
</style>
