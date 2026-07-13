<script setup>
import {computed, onBeforeUnmount, onMounted, ref} from 'vue'
import {useRouter} from 'vue-router'
import {EventsEmit} from '../../wailsjs/runtime'
import {Follow, GetIntradayRadarSnapshot, RefreshIntradayRadar} from '../../wailsjs/go/main/App'
import {describeRadarEvent, radarHowItWorks} from './radar/radar-explanation.mjs'
import {buildRadarFilter} from './radar/radarFilters.mjs'

const router = useRouter()
const tab = ref('all')
const onlyFollowed = ref(false)
const autoRefresh = ref(true)
const loading = ref(false)
const snapshot = ref({events: [], summary: {total: 0, p1: 0, p2: 0}, lastSuccessAt: '', lastError: ''})
let timer

const tabs = [
  {key: 'all', label: '全部'},
  {key: 'opportunity', label: '机会'},
  {key: 'risk', label: '风险'},
  {key: 'capital', label: '大单资金'},
]
const events = computed(() => snapshot.value.events || [])

function code(row) {
  const raw = String(row.code || '').replace(/^(sh|sz|bj)/i, '')
  return raw.startsWith('6') || raw.startsWith('688') ? `${raw}.SH` : raw.startsWith('8') || raw.startsWith('9') ? `${raw}.BJ` : `${raw}.SZ`
}
function openKline(row) {
  EventsEmit('klineSelectStock', {ts_code: code(row), name: row.name})
  router.push({name: 'klineAnalysis'})
}
function openAi(row) {
  router.push({name: 'research', query: {name: 'AI分析报告', stockCode: code(row)}})
}
function follow(row) {
  Follow((String(row.code).startsWith('6') ? 'sh' : 'sz') + String(row.code)).catch(() => {})
}
async function refresh() {
  loading.value = true
  try {
    snapshot.value = await RefreshIntradayRadar(buildRadarFilter({tab: tab.value, onlyFollowed: onlyFollowed.value}))
  } finally {
    loading.value = false
  }
}
function start() {
  clearInterval(timer)
  if (autoRefresh.value) timer = setInterval(refresh, 15000)
}
onMounted(async () => {
  snapshot.value = await GetIntradayRadarSnapshot()
  await refresh()
  start()
})
onBeforeUnmount(() => clearInterval(timer))
</script>

<template>
  <div class="radar">
    <div class="radar__heading">
      <div>
        <p>盘中异动雷达</p>
        <h1>发现异动，带着证据进入分析</h1>
      </div>
      <div class="radar__heading-tools">
        <label><input v-model="autoRefresh" type="checkbox" @change="start"> 自动刷新</label>
        <button type="button" @click="refresh">刷新</button>
      </div>
    </div>

    <details class="radar__how-it-works">
      <summary>这个雷达怎么看</summary>
      <p>{{ radarHowItWorks }}</p>
    </details>

    <div class="radar__filters">
      <button v-for="item in tabs" :key="item.key" type="button" :class="{active: tab === item.key}" @click="tab = item.key; refresh()">{{ item.label }}</button>
      <label><input v-model="onlyFollowed" type="checkbox" @change="refresh"> 仅自选/持仓</label>
      <span>P1 {{ snapshot.summary?.p1 || 0 }} · P2 {{ snapshot.summary?.p2 || 0 }} · {{ snapshot.lastSuccessAt ? '已同步' : '等待同步' }}</span>
    </div>
    <p v-if="snapshot.lastError" class="radar__error">{{ snapshot.lastError }}</p>

    <section class="radar__layout">
      <article class="radar__feed">
        <header><h2>事件流</h2><span>{{ events.length }} 条</span></header>
        <div v-for="row in events" :key="row.code + row.time + row.changeType" class="event">
          <i :class="row.priority === 'P1' ? 'p1' : 'p2'">{{ row.priority }}</i>
          <div class="event__body">
            <b>{{ row.name }} · {{ row.typeName }}</b>
            <small>{{ row.time }} · {{ row.industry || '未分类' }} · {{ Number(row.changeRate || 0).toFixed(2) }}%</small>
            <span>{{ describeRadarEvent(row).trigger }}</span>
          </div>
          <em>{{ row.isHolding ? '持仓' : row.isFollowed ? '自选' : row.category }}</em>
          <div class="event__actions">
            <button type="button" @click="openKline(row)">看K线</button>
            <button type="button" @click="openAi(row)">AI 分析</button>
            <button type="button" @click="follow(row)">关注</button>
          </div>
          <p>{{ describeRadarEvent(row).nextAction }}</p>
        </div>
        <p v-if="!loading && !events.length" class="radar__empty">当前筛选下暂未发现异动；交易时段会每 15 秒刷新一次。</p>
      </article>
      <aside class="radar__pulse">
        <h2>市场脉冲</h2>
        <strong>{{ snapshot.summary?.total || 0 }}</strong><span>条当日异动</span>
        <dl>
          <div><dt>机会</dt><dd>{{ snapshot.summary?.categories?.momentum || 0 }}</dd></div>
          <div><dt>风险</dt><dd>{{ (snapshot.summary?.categories?.risk || 0) + (snapshot.summary?.categories?.breakdown || 0) }}</dd></div>
          <div><dt>涨跌停</dt><dd>{{ snapshot.summary?.categories?.limit || 0 }}</dd></div>
        </dl>
        <RouterLink :to="{name: 'research', query: {name: '异动监控'}}">查看历史监控</RouterLink>
      </aside>
    </section>
  </div>
</template>

<style scoped>
.radar { min-height:100%; padding:22px; background:var(--workspace-canvas); }
.radar__heading { display:flex; justify-content:space-between; align-items:end; gap:16px; }
.radar__heading p { margin:0 0 5px; font-size:12px; color:var(--workspace-text-muted); }
.radar__heading h1 { margin:0; color:var(--workspace-text); font-size:22px; }
.radar__heading-tools { display:flex; align-items:center; gap:10px; font-size:12px; }
.radar__heading button,.radar__filters button { border:1px solid var(--workspace-border); border-radius:var(--workspace-radius-sm); background:var(--workspace-surface); color:var(--workspace-text); padding:8px 11px; cursor:pointer; }
.radar__how-it-works { margin:16px 0; padding:10px 12px; border:1px solid var(--workspace-border); border-radius:var(--workspace-radius-sm); background:var(--workspace-surface-muted); color:var(--workspace-text-muted); font-size:12px; }
.radar__how-it-works summary { color:var(--workspace-text); cursor:pointer; font-weight:650; }
.radar__how-it-works p { margin:8px 0 0; line-height:1.6; }
.radar__filters { display:flex; gap:7px; align-items:center; flex-wrap:wrap; margin:18px 0; }
.radar__filters button.active { border-color:var(--workspace-accent); background:var(--workspace-accent-soft); color:var(--workspace-accent); }
.radar__filters label { margin-left:8px; font-size:12px; }
.radar__filters span { margin-left:auto; color:var(--workspace-text-muted); font-size:12px; }
.radar__error { padding:8px 10px; background:#fff4e5; color:#915b0f; font-size:12px; }
.radar__layout { display:grid; grid-template-columns:minmax(0,1fr) 270px; gap:16px; }
.radar__feed,.radar__pulse { background:var(--workspace-surface); border:1px solid var(--workspace-border); border-radius:var(--workspace-radius-sm); }
.radar__feed header { height:46px; display:flex; align-items:center; padding:0 15px; border-bottom:1px solid var(--workspace-border); }
.radar__feed h2,.radar__pulse h2 { margin:0; color:var(--workspace-text); font-size:15px; }
.radar__feed header span { margin-left:auto; color:var(--workspace-text-muted); font-size:12px; }
.event { display:grid; grid-template-columns:34px minmax(180px,1fr) 48px 154px; gap:10px; align-items:center; padding:12px 15px; border-bottom:1px solid var(--workspace-border); }
.event i { font-style:normal; font-size:11px; padding:3px; color:#fff; background:var(--workspace-accent); border-radius:3px; text-align:center; }
.event i.p2 { background:#d08a1b; }.event__body b,.event__body small,.event__body span { display:block; }.event__body b { color:var(--workspace-text); font-size:13px; }.event__body small,.event__body span,.event p { color:var(--workspace-text-muted); font-size:11px; line-height:1.5; }.event__body span { margin-top:3px; }.event em { font-style:normal; color:var(--workspace-text-muted); font-size:11px; }.event__actions { display:flex; gap:6px; }.event__actions button { border:0; background:transparent; color:var(--workspace-link); font-size:12px; cursor:pointer; padding:0; }.event p { grid-column:2 / -1; margin:0; }
.radar__empty { padding:20px; color:var(--workspace-text-muted); text-align:center; }.radar__pulse { padding:17px; }.radar__pulse > strong { display:block; color:var(--workspace-text); font-size:30px; margin-top:18px; }.radar__pulse > span { color:var(--workspace-text-muted); font-size:12px; }.radar__pulse dl { margin:20px 0; }.radar__pulse dl div { display:flex; justify-content:space-between; border-bottom:1px solid var(--workspace-border); padding:9px 0; font-size:13px; }.radar__pulse dt { color:var(--workspace-text-muted); }.radar__pulse dd { margin:0; color:var(--workspace-text); font-weight:700; }.radar__pulse a { color:var(--workspace-link); font-size:12px; text-decoration:none; }
@media (max-width:960px) { .radar__layout { grid-template-columns:1fr; }.radar__pulse { display:none; } }
@media (max-width:639px) { .radar { padding:14px; }.radar__heading { align-items:flex-start; flex-direction:column; }.radar__heading h1 { font-size:18px; }.event { grid-template-columns:32px minmax(220px,1fr) 42px 150px; min-width:590px; }.radar__feed { overflow:auto; }.radar__filters span { width:100%; margin:3px 0; } }
</style>
