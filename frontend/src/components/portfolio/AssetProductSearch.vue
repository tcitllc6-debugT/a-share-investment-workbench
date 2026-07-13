<script setup>
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { CheckmarkCircleOutline, SearchOutline } from '@vicons/ionicons5'
import { NIcon } from 'naive-ui'
import { callPortfolioApi, readValue } from './portfolio-api.mjs'

const props = defineProps({
  assetType: { type: String, required: true },
  modelValue: { type: Object, default: null },
  busy: { type: Boolean, default: false },
  locked: { type: Boolean, default: false },
})
const emit = defineEmits(['update:modelValue', 'select'])

const keyword = ref('')
const results = ref([])
const loading = ref(false)
const error = ref('')
let timer = null
let requestSequence = 0

const selectedProduct = computed(() => props.modelValue)

const sourceNames = {
  eastmoney: '东方财富',
  binance: '币安行情',
  frankfurter: '公开汇率',
  local: '本地基础资料',
}

function valueOf(item, ...keys) {
  return readValue(item, ...keys)
}

function normaliseProduct(item) {
  return {
    assetType: valueOf(item, 'assetType', 'AssetType') || props.assetType,
    symbol: valueOf(item, 'symbol', 'Symbol') || '',
    name: valueOf(item, 'name', 'Name') || '',
    exchange: valueOf(item, 'exchange', 'Exchange') || '',
    quoteSymbol: valueOf(item, 'quoteSymbol', 'QuoteSymbol') || valueOf(item, 'symbol', 'Symbol') || '',
    currency: valueOf(item, 'currency', 'Currency') || 'CNY',
    unit: valueOf(item, 'unit', 'Unit') || '',
    source: valueOf(item, 'source', 'Source') || '',
    category: valueOf(item, 'category', 'Category') || '',
    metadata: valueOf(item, 'metadata', 'Metadata') || {},
  }
}

function sourceLabel(source) {
  return sourceNames[String(source || '').toLowerCase()] || '公开数据源'
}

function backendAssetType(value) {
  if (value === 'future') return 'futures'
  if (value === 'bank_product') return 'bank_wealth'
  return value
}

async function searchProducts(value) {
  const query = String(value || '').trim()
  const sequence = ++requestSequence
  if (!query) {
    results.value = []
    error.value = ''
    loading.value = false
    return
  }
  loading.value = true
  error.value = ''
  try {
    const response = await callPortfolioApi('SearchAssetInstruments', backendAssetType(props.assetType), query, 20)
    if (sequence !== requestSequence) return
    results.value = (Array.isArray(response) ? response : []).map(normaliseProduct)
  } catch (searchError) {
    if (sequence !== requestSequence) return
    results.value = []
    error.value = searchError?.message || '暂时无法搜索，请稍后重试'
  } finally {
    if (sequence === requestSequence) loading.value = false
  }
}

watch(keyword, (value) => {
  if (timer) clearTimeout(timer)
  timer = setTimeout(() => searchProducts(value), 250)
})

watch(() => props.assetType, () => {
  keyword.value = ''
  results.value = []
  error.value = ''
})

function choose(item) {
  if (props.busy) return
  const product = normaliseProduct(item)
  emit('update:modelValue', product)
  emit('select', product)
  keyword.value = ''
  results.value = []
}

function clearSelection() {
  if (props.busy || props.locked) return
  emit('update:modelValue', null)
  keyword.value = ''
  results.value = []
  requestAnimationFrame(() => document.querySelector('.asset-product-search input')?.focus())
}

onBeforeUnmount(() => {
  if (timer) clearTimeout(timer)
  requestSequence += 1
})
</script>

<template>
  <div class="asset-product-search">
    <div v-if="selectedProduct" class="selected-product">
      <n-icon :component="CheckmarkCircleOutline" aria-hidden="true" />
      <div>
        <b>{{ selectedProduct.name }}</b>
        <span>{{ selectedProduct.symbol }}<i v-if="selectedProduct.exchange"> · {{ selectedProduct.exchange }}</i></span>
        <small>{{ selectedProduct.category || '标准金融产品' }} · {{ sourceLabel(selectedProduct.source) }}</small>
      </div>
      <button type="button" :disabled="busy || locked" @click="clearSelection">重新选择</button>
    </div>

    <template v-else>
      <label class="search-input">
        <span>搜索并选择金融产品<i>必选</i></span>
        <div>
          <n-icon :component="SearchOutline" aria-hidden="true" />
          <input v-model="keyword" type="search" autocomplete="off" placeholder="输入产品名称、代码或简称">
          <em v-if="loading">搜索中…</em>
        </div>
        <small>产品代码由数据源自动匹配，选择后再填写成交信息。</small>
      </label>

      <div v-if="error" class="search-message is-error" role="alert">{{ error }}</div>
      <div v-else-if="keyword && !loading && results.length === 0" class="search-message">没有匹配结果，请更换关键词。</div>
      <div v-if="results.length" class="search-results" role="listbox" aria-label="金融产品搜索结果">
        <button v-for="item in results" :key="`${item.source}-${item.quoteSymbol}-${item.symbol}`" type="button" role="option" @click="choose(item)">
          <span><b>{{ item.name }}</b><small>{{ item.symbol }}</small></span>
          <span><i>{{ item.category || '金融产品' }}</i><small>{{ item.exchange || sourceLabel(item.source) }}</small></span>
        </button>
      </div>
    </template>
  </div>
</template>

<style scoped>
.asset-product-search{position:relative}.selected-product{min-height:72px;display:grid;grid-template-columns:30px minmax(0,1fr) auto;align-items:center;gap:11px;padding:12px;border:1px solid color-mix(in srgb,var(--workspace-accent) 38%,var(--workspace-border));border-radius:var(--workspace-radius-sm);background:color-mix(in srgb,var(--workspace-accent-soft) 45%,var(--workspace-surface-solid))}.selected-product>.n-icon{color:var(--workspace-accent);font-size:25px}.selected-product b,.selected-product span,.selected-product small{display:block;letter-spacing:0}.selected-product b{font-size:14px;font-weight:700}.selected-product span{margin-top:2px;color:var(--workspace-text-secondary);font-size:12px}.selected-product span i{font-style:normal}.selected-product small{margin-top:3px;color:var(--workspace-text-muted);font-size:11px}.selected-product button{min-height:34px;padding:0 11px;border:1px solid var(--workspace-border);border-radius:var(--workspace-radius-xs);color:var(--workspace-accent);background:var(--workspace-surface-solid);font-size:12px;cursor:pointer}.search-input>span{min-height:20px;display:flex;align-items:center;justify-content:space-between;color:var(--workspace-text-secondary);font-size:12px;font-weight:650}.search-input>span i{color:var(--workspace-accent);font-size:10px;font-style:normal}.search-input>div{min-height:44px;display:flex;align-items:center;gap:8px;margin-top:5px;padding:0 11px;border:1px solid var(--workspace-border-strong);border-radius:var(--workspace-radius-sm);background:var(--workspace-surface-solid);transition:border-color 150ms ease,box-shadow 150ms ease}.search-input>div:focus-within{border-color:var(--workspace-accent);box-shadow:var(--workspace-focus-ring)}.search-input .n-icon{flex:0 0 auto;color:var(--workspace-text-muted);font-size:18px}.search-input input{min-width:0;flex:1;border:0;outline:0;color:var(--workspace-text);background:transparent;font:inherit;font-size:13px}.search-input em{flex:0 0 auto;color:var(--workspace-accent);font-size:11px;font-style:normal}.search-input>small{display:block;margin-top:5px;color:var(--workspace-text-muted);font-size:11px}.search-results{position:absolute;z-index:80;top:75px;left:0;right:0;max-height:286px;overflow:auto;padding:5px;border:1px solid var(--workspace-border-strong);border-radius:var(--workspace-radius-sm);background:var(--workspace-surface-elevated);box-shadow:var(--workspace-shadow)}.search-results button{width:100%;display:grid;grid-template-columns:minmax(0,1fr) auto;align-items:center;gap:14px;padding:10px;border:0;border-radius:var(--workspace-radius-xs);color:var(--workspace-text);background:transparent;text-align:left;cursor:pointer}.search-results button:hover,.search-results button:focus-visible{outline:0;background:var(--workspace-accent-soft)}.search-results span:last-child{text-align:right}.search-results b,.search-results small,.search-results i{display:block;letter-spacing:0}.search-results b{font-size:13px}.search-results small{margin-top:2px;color:var(--workspace-text-muted);font-size:11px}.search-results i{color:var(--workspace-text-secondary);font-size:11px;font-style:normal}.search-message{margin-top:7px;padding:8px 10px;border-radius:var(--workspace-radius-xs);color:var(--workspace-text-muted);background:var(--workspace-surface-muted);font-size:11px}.search-message.is-error{color:var(--workspace-negative)}@media(max-width:640px),(min-resolution:2.5dppx){.selected-product{grid-template-columns:26px minmax(0,1fr)}.selected-product button{grid-column:2;justify-self:start}.search-results{position:static;margin-top:7px}.search-input input{font-size:14px}}
</style>
