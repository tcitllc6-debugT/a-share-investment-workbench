<script setup>
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import TradingRecordManager from './TradingRecordManager.vue'
import PortfolioOverview from './portfolio/PortfolioOverview.vue'
import AssetPositionTable from './portfolio/AssetPositionTable.vue'
import AssetTransactionEditor from './portfolio/AssetTransactionEditor.vue'
import BankProductImport from './portfolio/BankProductImport.vue'
import { ASSET_TYPE_MAP } from './portfolio/asset-form-schema.mjs'
import { callPortfolioApi } from './portfolio/portfolio-api.mjs'

const route = useRoute()
const router = useRouter()
const validViews = new Set(['overview', 'stock', 'fund', 'bond', 'precious_metal', 'forex', 'future', 'crypto', 'bank_product'])
const currentView = computed(() => validViews.has(String(route.query.assetType || '')) ? String(route.query.assetType) : 'overview')
const currentMeta = computed(() => ASSET_TYPE_MAP[currentView.value] || null)
const overview = ref({})
const positions = ref([])
const loading = ref(false)
const saving = ref(false)
const status = ref('')
const editorShow = ref(false)
const editorPosition = ref(null)
const editorAction = ref('')

function backendAssetType(value) {
  if (value === 'future') return 'futures'
  if (value === 'bank_product') return 'bank_wealth'
  return value
}

async function load() {
  loading.value = true
  status.value = ''
  try {
    overview.value = await callPortfolioApi('GetPortfolioOverview') || {}
    if (!['overview', 'stock'].includes(currentView.value)) {
      positions.value = await callPortfolioApi('GetAssetPositions', backendAssetType(currentView.value)) || []
    } else positions.value = []
  } catch (error) { status.value = error?.message || String(error) }
  finally { loading.value = false }
}

function selectAsset(assetType) {
  const uiType = assetType === 'futures' ? 'future' : assetType === 'bank_wealth' ? 'bank_product' : assetType
  router.push({ name: 'holdingCenter', query: { ...route.query, assetType: uiType } })
}

function openEditor(detail = {}) {
  editorPosition.value = detail.position || null
  editorAction.value = detail.action || ''
  editorShow.value = true
}

function transactionPayload(payload) {
  return {
    id: Number(payload.id || 0),
    instrumentId: Number(payload.instrumentId || 0),
    assetType: backendAssetType(payload.assetType),
    symbol: payload.instrumentCode,
    name: payload.instrumentName,
    exchange: payload.exchange || '',
    quoteSymbol: payload.quoteSymbol || payload.instrumentCode,
    unit: payload.unit || '',
    direction: payload.action,
    quantity: Number(payload.quantity || 0),
    price: Number(payload.price || 0),
    currency: payload.currency || 'CNY',
    fee: Number(payload.fee || 0),
    fxRateToCny: payload.currency === 'CNY' ? 1 : 0,
    tradeTime: payload.tradedAt,
    note: payload.note || '',
    metadata: payload.details || {},
  }
}

async function saveTransaction(payload) {
  saving.value = true
  status.value = ''
  try {
    const input = transactionPayload(payload)
    await callPortfolioApi(input.id ? 'UpdateAssetTransaction' : 'AddAssetTransaction', input)
    editorShow.value = false
    status.value = '交易已写入统一账本'
    await refreshQuotes()
  } catch (error) { status.value = error?.message || String(error) }
  finally { saving.value = false }
}

async function refreshQuotes() {
  loading.value = true
  status.value = '正在更新公开行情与汇率…'
  try {
    const types = currentView.value === 'overview' ? [] : [backendAssetType(currentView.value)]
    const result = await callPortfolioApi('RefreshAssetQuotes', types)
    const failed = Number(result?.failed || 0)
    const detail = failed && Array.isArray(result?.errors) && result.errors.length ? `；${result.errors.slice(0, 2).join('；')}` : ''
    status.value = `报价更新 ${result?.updated || 0} 项，失败 ${failed} 项${detail}`
    await load()
  } catch (error) { status.value = error?.message || String(error); loading.value = false }
}

async function updateManualQuote(position) {
  const value = window.prompt('输入最新净值或价格', String(position?.latestPrice || position?.averageCost || ''))
  if (value === null) return
  const price = Number(value)
  if (!Number.isFinite(price) || price <= 0) { status.value = '请输入大于 0 的有效净值'; return }
  try {
    await callPortfolioApi('SaveManualAssetQuote', Number(position.instrumentId || 0), price, position.currency || 'CNY')
    status.value = '手动净值已更新'
    await load()
  } catch (error) { status.value = error?.message || String(error) }
}

watch(currentView, load, { immediate: true })
</script>

<template>
  <section class="holding-center">
    <div class="holding-center__intro">
      <div><p>金融产品持仓</p><h1>{{ currentView === 'overview' ? '全部资产概览' : `${currentMeta?.label || ''}持仓` }}</h1><span>查看各类资产的市值、成本与收益，统一折算为人民币</span></div>
      <div class="holding-center__actions">
        <button type="button" :disabled="loading" @click="refreshQuotes">更新报价</button>
        <button v-if="!['overview','stock'].includes(currentView)" type="button" class="primary" @click="openEditor()">增加一笔持仓</button>
      </div>
    </div>
    <p v-if="status" class="holding-center__status">{{ status }}</p>
    <PortfolioOverview v-if="currentView === 'overview'" :overview="overview" :loading="loading" @select-asset="selectAsset" />
    <TradingRecordManager v-else-if="currentView === 'stock'" />
    <template v-else>
      <AssetPositionTable :asset-type="currentView" :positions="positions" :loading="loading" @transact="openEditor" @refresh-quotes="refreshQuotes" @manual-quote="updateManualQuote" />
      <BankProductImport v-if="currentView === 'bank_product'" @imported="load" />
    </template>
    <AssetTransactionEditor v-if="!['overview','stock'].includes(currentView)" v-model:show="editorShow" :asset-type="currentView" :position="editorPosition" :requested-action="editorAction" :saving="saving" @submit="saveTransaction" />
  </section>
</template>

<style scoped>
.holding-center{min-height:100%;min-width:0;padding:20px;overflow:auto;color:var(--workspace-text);font-family:var(--workspace-font),"Microsoft YaHei UI","PingFang SC",sans-serif;letter-spacing:0;-webkit-font-smoothing:antialiased;text-rendering:optimizeLegibility}.holding-center__intro{display:flex;align-items:flex-end;justify-content:space-between;gap:16px;margin:0 0 14px}.holding-center__intro p{margin:0 0 4px;color:var(--workspace-text-secondary);font-size:12px;font-weight:650}.holding-center__intro h1{margin:0;color:var(--workspace-text);font-size:21px;font-weight:750}.holding-center__intro span{display:block;margin-top:4px;color:var(--workspace-text-secondary);font-size:12px;line-height:1.5}.holding-center__actions{display:flex;gap:8px}.holding-center__actions button{min-height:36px;padding:0 12px;border:1px solid var(--workspace-border);border-radius:var(--workspace-radius-sm);color:var(--workspace-text-secondary);background:var(--workspace-surface-solid);font-size:12px;font-weight:650;cursor:pointer}.holding-center__actions button.primary{color:white;border-color:var(--workspace-accent);background:var(--workspace-accent)}.holding-center__status{margin:0 0 10px;padding:9px 11px;border:1px solid var(--workspace-border);color:var(--workspace-text-secondary);background:var(--workspace-surface-muted);font-size:12px}.holding-center :deep(.bank-import){margin-top:12px}@media(max-width:719px),(min-resolution:2.5dppx){.holding-center{padding:14px}.holding-center__intro{align-items:flex-start;flex-direction:column}.holding-center__actions{width:100%;overflow-x:auto}.holding-center__actions button{flex:0 0 auto}}
</style>
