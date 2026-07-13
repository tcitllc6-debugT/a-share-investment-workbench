<script setup>
import {computed, defineAsyncComponent, reactive, ref} from 'vue'
import {useRouter} from 'vue-router'
import {
  GetFollowList,
  GetFollowedStockQuotes,
  GetHoldingSummary,
  GetIntradayRadarSnapshot,
  GetTelegraphBriefListBySources,
  GetTodayMarketStatistic,
  RefreshIntradayRadar,
} from '../../wailsjs/go/main/App'
import {
  interleaveNewsSources,
  latestMarketPulse,
  loadDashboardResources,
  mergeWatchlistQuotes,
  normaliseFollowedStock,
} from './decision-workbench.mjs'

const StockKlineModal = defineAsyncComponent(() => import('./StockKlineModal.vue'))

const router = useRouter()
const holdings = ref({totalMarketValue: 0, totalFloatingProfit: 0, totalFloatingProfitPct: 0, holdings: []})
const watchlist = ref([])
const radarEvents = ref([])
const news = ref([])
const marketPulse = ref(latestMarketPulse())
const resourceLoading = reactive({holdings: true, watchlist: true, radar: true, news: true, pulse: true})
const resourceErrors = reactive({})
const klineModalShow = ref(false)
const klineStockCode = ref('')
const klineStockName = ref('')
const klineDarkTheme = ref(false)
let loadGeneration = 0

const dashboardRadarFilter = {
  categories: [],
  onlyFollowed: false,
  onlyHoldings: false,
  minAmount: 0,
  industry: '',
  concept: '',
  marketBoards: [],
}

const actionItems = computed(() => radarEvents.value
  .filter(item => [8203, 8204, 4, 8201].includes(item.changeType))
  .slice(0, 4))
const radarSummary = computed(() => ({
  total: radarEvents.value.length,
  up: radarEvents.value.filter(item => [8201, 8202, 8193, 4, 64].includes(item.changeType)).length,
  down: radarEvents.value.filter(item => [8203, 8204, 8194, 8, 128].includes(item.changeType)).length,
}))
const hasErrors = computed(() => Object.keys(resourceErrors).length > 0)

function money(value) {
  return new Intl.NumberFormat('zh-CN', {style: 'currency', currency: 'CNY', maximumFractionDigits: 0}).format(Number(value || 0))
}
function percent(value, fallback = '--') {
  if (value === null || value === undefined || !Number.isFinite(Number(value))) return fallback
  const number = Number(value)
  return `${number >= 0 ? '+' : ''}${number.toFixed(2)}%`
}
function quote(value) {
  return value === null || value === undefined ? '--' : Number(value).toFixed(2)
}
function up(value) { return Number(value || 0) >= 0 ? 'up' : 'down' }
function toKlineCode(code) {
  const value = String(code || '').trim()
  const suffixCode = value.match(/^(\d{6})\.(SH|SZ|BJ)$/i)
  if (suffixCode) return `${suffixCode[1]}.${suffixCode[2].toUpperCase()}`
  const prefixCode = value.match(/^(sh|sz|bj)(\d{6})$/i)
  if (prefixCode) return `${prefixCode[2]}.${prefixCode[1].toUpperCase()}`
  if (!/^\d{6}$/.test(value)) return ''
  if (/^[65]/.test(value)) return `${value}.SH`
  if (/^[894]/.test(value)) return `${value}.BJ`
  return `${value}.SZ`
}
function openKline(item) {
  const code = toKlineCode(item.stockCode || item.code)
  if (!code) return
  klineStockCode.value = code
  klineStockName.value = item.stockName || item.name || ''
  klineDarkTheme.value = document.documentElement.dataset.themeMode === 'dark'
  klineModalShow.value = true
}
function openRadar() { router.push({name: 'radar'}) }

function applyRadarSnapshot(snapshot) {
  radarEvents.value = Array.isArray(snapshot?.events) ? snapshot.events : []
  if (snapshot?.lastError) resourceErrors.radar = true
  else delete resourceErrors.radar
}

async function loadMixedNews() {
  const rows = await GetTelegraphBriefListBySources(['财联社电报', '新浪财经'], 12)
  return interleaveNewsSources([
    (rows || []).filter(item => item.source === '财联社电报'),
    (rows || []).filter(item => item.source === '新浪财经'),
  ], 6)
}

async function loadRadarSnapshot(generation) {
  try {
    const snapshot = await GetIntradayRadarSnapshot()
    if (generation !== loadGeneration) return
    applyRadarSnapshot(snapshot)
  } catch (error) {
    if (generation === loadGeneration) resourceErrors.radar = true
  } finally {
    if (generation === loadGeneration) resourceLoading.radar = false
  }

  void RefreshIntradayRadar(dashboardRadarFilter).then(snapshot => {
    if (generation !== loadGeneration) return
    applyRadarSnapshot(snapshot)
  }).catch(() => {
    if (generation === loadGeneration) resourceErrors.radar = true
  })
}

async function load() {
  const generation = ++loadGeneration
  Object.keys(resourceLoading).forEach(key => { resourceLoading[key] = true })
  Object.keys(resourceErrors).forEach(key => { delete resourceErrors[key] })
  void loadRadarSnapshot(generation)
  await loadDashboardResources({
    holdings: GetHoldingSummary,
    watchlist: () => GetFollowList(0),
    news: loadMixedNews,
    pulse: GetTodayMarketStatistic,
  }, (key, value) => {
    if (generation !== loadGeneration) return
    if (key === 'holdings') holdings.value = value || holdings.value
    if (key === 'watchlist') {
      const localRows = (value || []).map(normaliseFollowedStock).filter(item => item.stockCode).slice(0, 5)
      watchlist.value = localRows
      void GetFollowedStockQuotes().then((quotes) => {
        if (generation !== loadGeneration) return
        watchlist.value = mergeWatchlistQuotes(value || [], quotes || []).filter(item => item.stockCode).slice(0, 5)
      }).catch(() => {})
    }
    if (key === 'news') news.value = value
    if (key === 'pulse') marketPulse.value = latestMarketPulse(value)
    resourceLoading[key] = false
  }, (key) => {
    if (generation !== loadGeneration) return
    resourceErrors[key] = true
    resourceLoading[key] = false
  })
}

load()
</script>

<template>
  <div class="decision-workbench">
    <div class="decision-workbench__heading">
      <h1>一屏掌握自选、持仓与市场动态</h1>
      <button type="button" @click="load">刷新数据</button>
    </div>
    <p v-if="hasErrors" class="decision-workbench__notice">部分数据暂时不可用，其他内容已正常显示。</p>

    <section class="decision-workbench__grid">
      <article class="panel panel--holdings">
        <header><div><h2>持仓与自选</h2><span>持仓来自交易记录，自选来自关注列表</span></div><RouterLink :to="{name: 'stock'}">管理自选</RouterLink></header>
        <div class="metrics">
          <div><label>持仓市值</label><strong>{{ money(holdings.totalMarketValue) }}</strong><small>{{ holdings.holdings?.length || 0 }} 只持仓</small></div>
          <div><label>浮动盈亏</label><strong :class="up(holdings.totalFloatingProfit)">{{ money(holdings.totalFloatingProfit) }}</strong><small :class="up(holdings.totalFloatingProfitPct)">{{ percent(holdings.totalFloatingProfitPct) }}</small></div>
          <div><label>自选数量</label><strong>{{ watchlist.length }}</strong><small>显示前 5 只</small></div>
        </div>
        <div class="watch-rows">
          <button v-for="row in watchlist" :key="row.stockCode" type="button" @click="openKline(row)">
            <span><b>{{ row.name || '未命名股票' }}</b><small>{{ row.stockCode }}</small></span>
            <span class="number">{{ quote(row.price) }}</span>
            <em :class="up(row.changePercent)">{{ percent(row.changePercent) }}</em>
            <i>K线</i>
          </button>
          <p v-if="resourceLoading.watchlist" class="panel-state">正在读取本地自选…</p>
          <p v-else-if="!watchlist.length" class="panel-state">暂无自选股票，可在股票自选中添加。</p>
        </div>
      </article>

      <article class="panel panel--attention">
        <header><div><h2>今日关注</h2><span>优先查看强信号</span></div><button class="text-button" type="button" @click="openRadar">查看全部</button></header>
        <div class="action-list">
          <button v-for="item in actionItems" :key="item.code + item.time + item.changeType" type="button" @click="openKline(item)">
            <i :class="item.changeRate >= 0 ? 'red-dot' : 'green-dot'"></i>
            <span><b>{{ item.name }} · {{ item.typeName }}</b><small>{{ item.time }} · {{ percent(item.changeRate) }}</small></span><em>分析</em>
          </button>
          <p v-if="!resourceLoading.radar && !actionItems.length" class="panel-state">暂无需要优先查看的异动。</p>
        </div>
      </article>

      <article class="panel panel--pulse">
        <header><div><h2>A股市场温度</h2><span>{{ marketPulse.dataTime ? `更新于 ${marketPulse.dataTime}` : '最近一次本地统计' }}</span></div></header>
        <div class="pulse-score"><strong>{{ marketPulse.sentimentDesc }}</strong><span>红盘率 {{ marketPulse.upRatio.toFixed(1) }}%</span></div>
        <div class="pulse-grid"><span><b class="up">{{ marketPulse.upCount }}</b>上涨</span><span><b class="down">{{ marketPulse.downCount }}</b>下跌</span><span><b class="up">{{ marketPulse.limitUp }}</b>涨停</span><span><b class="down">{{ marketPulse.limitDown }}</b>跌停</span></div>
      </article>

      <article class="panel panel--radar">
        <header><div><h2>盘中异动</h2><span>捕捉涨跌与成交异动</span></div><button class="text-button" type="button" @click="openRadar">打开雷达</button></header>
        <div class="radar-summary"><strong>{{ radarSummary.total }}</strong><span>条当日异动</span><div><b class="up">{{ radarSummary.up }} 上攻</b><b class="down">{{ radarSummary.down }} 风险</b></div></div>
      </article>

      <article class="panel panel--guide">
        <header><div><h2>快捷进入常用工具</h2></div></header>
        <div class="guide-actions"><RouterLink :to="{name: 'screening', query: {mode: 'duo-kong'}}">多空线选股 <span>近 3 日信号</span></RouterLink><RouterLink :to="{name: 'screening', query: {mode: 'pattern'}}">形态选股 <span>扫描 A 股</span></RouterLink><RouterLink :to="{name: 'holdingCenter'}">股票持仓 <span>核对持仓成本</span></RouterLink></div>
      </article>

      <article class="panel panel--news">
        <header><div><h2>市场快讯</h2><span>最新 A 股财经动态</span></div><RouterLink :to="{name: 'market'}">进入市场资讯</RouterLink></header>
        <div class="news-list">
          <button v-for="item in news" :key="`${item.source}-${item.id}`" type="button"><time>{{ item.time }}</time><span><b>{{ item.source === '财联社电报' ? '财联社' : item.source }}</b>{{ item.content }}</span></button>
          <p v-if="resourceLoading.news" class="panel-state">正在读取本地快讯…</p>
          <p v-else-if="!news.length" class="panel-state">本地暂无快讯，后台同步完成后会自动出现。</p>
        </div>
      </article>
    </section>
    <StockKlineModal
      v-if="klineModalShow"
      v-model:show="klineModalShow"
      :code="klineStockCode"
      :stock-name="klineStockName"
      :dark-theme="klineDarkTheme"
      :chart-height="500"
    />
  </div>
</template>

<style scoped>
.decision-workbench{min-height:100%;padding:20px;color:var(--workspace-text);background:var(--workspace-canvas)}
.decision-workbench__heading{display:flex;align-items:center;justify-content:space-between;gap:16px;margin-bottom:16px}.decision-workbench__heading h1{margin:0;font-size:21px;letter-spacing:0}.decision-workbench__heading button{min-height:34px;padding:0 12px;border:1px solid var(--workspace-border);border-radius:var(--workspace-radius-sm);color:var(--workspace-text-secondary);background:var(--workspace-surface-solid);cursor:pointer}.decision-workbench__heading button:hover{color:var(--workspace-accent);border-color:var(--workspace-accent);background:var(--workspace-accent-soft)}
.decision-workbench__notice{margin:0 0 12px;padding:8px 10px;color:var(--workspace-warning);background:color-mix(in srgb,var(--workspace-warning) 10%,var(--workspace-surface-solid));border:1px solid color-mix(in srgb,var(--workspace-warning) 30%,var(--workspace-border));font-size:12px}
.decision-workbench__grid{display:grid;grid-template-columns:repeat(12,minmax(0,1fr));gap:14px}.panel{min-width:0;overflow:hidden;color:var(--workspace-text);background:var(--workspace-surface-solid);border:1px solid var(--workspace-border);border-radius:var(--workspace-radius-md);box-shadow:var(--workspace-shadow-sm)}.panel--holdings{grid-column:span 8}.panel--attention{grid-column:span 4}.panel--pulse,.panel--radar,.panel--guide{grid-column:span 4}.panel--news{grid-column:1/-1}
.panel header{min-height:48px;padding:8px 15px;display:flex;align-items:center;border-bottom:1px solid var(--workspace-border);background:color-mix(in srgb,var(--workspace-surface-muted) 42%,var(--workspace-surface-solid))}.panel h2{margin:0;font-size:14px}.panel header span{display:block;margin-top:2px;color:var(--workspace-text-muted);font-size:10px}.panel header a,.text-button{margin-left:auto;border:0;color:var(--workspace-link);background:transparent;text-decoration:none;font-size:12px;cursor:pointer}
.metrics{display:grid;grid-template-columns:repeat(3,1fr);padding:14px 15px;border-bottom:1px solid var(--workspace-border)}.metrics>div{padding:0 14px;border-right:1px solid var(--workspace-border)}.metrics>div:first-child{padding-left:0}.metrics>div:last-child{border:0}.metrics label,.metrics small{display:block;color:var(--workspace-text-muted);font-size:10px}.metrics strong{display:block;margin:5px 0 2px;font-family:var(--workspace-number-font);font-size:19px}.up{color:var(--workspace-positive)}.down{color:var(--workspace-negative)}
.watch-rows button{display:grid;grid-template-columns:minmax(120px,1.4fr) 80px 80px 42px;align-items:center;width:100%;padding:9px 15px;border:0;border-bottom:1px solid var(--workspace-border);color:var(--workspace-text-secondary);background:var(--workspace-surface-solid);text-align:right;cursor:pointer}.watch-rows button:hover,.action-list button:hover,.news-list button:hover{background:var(--workspace-surface-hover)}.watch-rows span:first-child{text-align:left}.watch-rows b,.watch-rows small{display:block}.watch-rows small{margin-top:2px;color:var(--workspace-text-muted);font-size:10px}.watch-rows .number,.watch-rows em{font-family:var(--workspace-number-font);font-style:normal}.watch-rows i{justify-self:end;padding:3px 5px;border:1px solid var(--workspace-border);border-radius:3px;color:var(--workspace-text-secondary);font-size:10px;font-style:normal}
.action-list{padding:0 14px}.action-list button{display:grid;grid-template-columns:7px 1fr auto;align-items:center;gap:9px;width:100%;padding:10px 0;border:0;border-bottom:1px solid var(--workspace-border);color:var(--workspace-text);background:var(--workspace-surface-solid);text-align:left;cursor:pointer}.action-list button:last-child{border:0}.action-list i{width:7px;height:7px;border-radius:50%}.red-dot{background:var(--workspace-positive)}.green-dot{background:var(--workspace-negative)}.action-list b,.action-list small{display:block}.action-list b{font-size:11px}.action-list small{margin-top:2px;color:var(--workspace-text-muted);font-size:10px}.action-list em{color:var(--workspace-link);font-size:11px;font-style:normal}
.pulse-score{display:flex;align-items:baseline;justify-content:space-between;padding:14px 15px 9px}.pulse-score strong{font-size:20px}.pulse-score span{color:var(--workspace-text-muted);font-size:11px}.pulse-grid{display:grid;grid-template-columns:repeat(4,1fr);padding:8px 15px 15px;gap:8px}.pulse-grid span{color:var(--workspace-text-muted);font-size:10px}.pulse-grid b{display:block;margin-bottom:3px;font-family:var(--workspace-number-font);font-size:16px}.radar-summary{padding:17px 15px}.radar-summary strong{font-family:var(--workspace-number-font);font-size:26px}.radar-summary>span{margin-left:6px;color:var(--workspace-text-muted);font-size:11px}.radar-summary div{display:flex;gap:14px;margin-top:11px}.radar-summary b{font-size:11px}
.guide-actions{display:grid;padding:6px 8px}.guide-actions a{display:flex;justify-content:space-between;padding:10px;color:var(--workspace-text-secondary);border-bottom:1px solid var(--workspace-border);text-decoration:none;font-size:12px}.guide-actions a:hover{color:var(--workspace-accent);background:var(--workspace-accent-soft)}.guide-actions a:last-child{border:0}.guide-actions span{color:var(--workspace-text-muted);font-size:10px}
.news-list{padding:0 15px}.news-list button{display:grid;grid-template-columns:54px 1fr;gap:10px;align-items:start;width:100%;padding:10px 0;border:0;border-bottom:1px solid var(--workspace-border);color:var(--workspace-text);background:var(--workspace-surface-solid);text-align:left;cursor:pointer}.news-list time{color:var(--workspace-text-muted);font-family:var(--workspace-number-font);font-size:11px}.news-list span{font-size:12px;line-height:1.45}.news-list span b{display:inline-block;margin-right:7px;padding:1px 5px;border-radius:3px;color:var(--workspace-accent);background:var(--workspace-accent-soft);font-size:10px;font-weight:600}.panel-state{margin:0;padding:18px 15px;color:var(--workspace-text-muted);font-size:11px}
@media(max-width:1100px){.panel--holdings,.panel--attention{grid-column:span 12}.panel--pulse,.panel--radar,.panel--guide{grid-column:span 4}}
@media(max-width:760px){.decision-workbench{padding:14px}.decision-workbench__heading h1{font-size:18px}.panel--pulse,.panel--radar,.panel--guide{grid-column:span 12}.metrics{grid-template-columns:1fr;gap:11px}.metrics>div{padding:0;border:0}.watch-rows{overflow:auto}.watch-rows button{min-width:470px}.pulse-grid{grid-template-columns:repeat(2,1fr)}}
</style>
