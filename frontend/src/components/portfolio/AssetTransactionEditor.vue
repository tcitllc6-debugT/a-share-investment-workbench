<script setup>
import { computed, reactive, ref, watch } from 'vue'
import { CloseOutline } from '@vicons/ionicons5'
import { NIcon, NModal } from 'naive-ui'
import AssetProductSearch from './AssetProductSearch.vue'
import {
  ASSET_TYPE_MAP,
  createAssetTransactionDraft,
  getActionOptions,
  getAssetFormFields,
  validateAssetTransaction,
} from './asset-form-schema.mjs'

const props = defineProps({
  show: { type: Boolean, default: false },
  assetType: { type: String, required: true },
  position: { type: Object, default: null },
  requestedAction: { type: String, default: '' },
  saving: { type: Boolean, default: false },
})
const emit = defineEmits(['update:show', 'submit'])

const draft = reactive({})
const errors = reactive({})
const selectedProduct = ref(null)
const assetMeta = computed(() => ASSET_TYPE_MAP[props.assetType] || ASSET_TYPE_MAP.stock)
const fields = computed(() => getAssetFormFields(props.assetType))
const commonKeys = new Set(['instrumentCode', 'instrumentName', 'quoteSymbol', 'action', 'quantity', 'price', 'currency', 'fee', 'tradedAt', 'note'])
const identityKeys = new Set(['instrumentCode', 'instrumentName', 'quoteSymbol'])
const commonFields = computed(() => fields.value.filter((field) => {
  if (!commonKeys.has(field.key)) return false
  if (props.assetType === 'bank_product') return field.key !== 'quoteSymbol'
  return !identityKeys.has(field.key)
}))
const specificFields = computed(() => fields.value.filter((field) => !commonKeys.has(field.key)))

const modalVisible = computed({
  get: () => props.show,
  set: (value) => emit('update:show', value),
})

function resolvedAction() {
  const options = getActionOptions(props.assetType)
  if (props.assetType === 'future' && props.position) {
    const metadata = props.position?.metadata || props.position?.Metadata || {}
    const isShort = String(metadata.side || '').includes('空')
    if (props.requestedAction === 'add') return isShort ? 'open_short' : 'open_long'
    if (['reduce', 'close'].includes(props.requestedAction)) return isShort ? 'close_short' : 'close_long'
  }
  if (props.requestedAction === 'add') return options[0]?.value || 'buy'
  if (props.requestedAction === 'reduce') return options[1]?.value || 'sell'
  if (props.requestedAction === 'close') return options.at(-1)?.value || 'close'
  return props.requestedAction
}

function resetDraft() {
  Object.keys(draft).forEach((key) => delete draft[key])
  Object.keys(errors).forEach((key) => delete errors[key])
  Object.assign(draft, createAssetTransactionDraft(props.assetType, props.position, resolvedAction()))
  const positionMetadata = props.position?.metadata || props.position?.Metadata || {}
  Object.assign(draft, positionMetadata)
  selectedProduct.value = props.position ? {
    assetType: props.assetType,
    symbol: draft.instrumentCode,
    name: draft.instrumentName,
    exchange: draft.exchange || '',
    quoteSymbol: draft.quoteSymbol || draft.instrumentCode,
    currency: draft.currency || 'CNY',
    unit: draft.unit || '',
    source: 'local',
    category: `${assetMeta.value.label}持仓`,
    metadata: positionMetadata,
  } : null
  if (props.assetType === 'future' && !draft.multiplier) draft.multiplier = 1
  if (props.assetType === 'precious_metal' && !draft.unit) draft.unit = '克'
  if (props.assetType === 'bank_product' && !draft.valuationMode) draft.valuationMode = '净值计价'
}

function selectProduct(product) {
  selectedProduct.value = product
  draft.instrumentCode = product.symbol || ''
  draft.instrumentName = product.name || ''
  draft.quoteSymbol = product.quoteSymbol || product.symbol || ''
  draft.exchange = product.exchange || ''
  draft.currency = product.currency || draft.currency || 'CNY'
  if (product.unit) draft.unit = product.unit
  draft.searchMetadata = { ...(product.metadata || {}), source: product.source || '', category: product.category || '' }
  for (const [key, value] of Object.entries(product.metadata || {})) {
    if (value !== undefined && value !== null && value !== '') draft[key] = value
  }
  if (product.metadata?.base_currency) draft.baseCurrency = product.metadata.base_currency
  if (product.metadata?.quote_currency) draft.quoteCurrency = product.metadata.quote_currency
  delete errors.product
}

function fieldReadonly(field) {
  return props.assetType !== 'bank_product' && ['baseCurrency', 'quoteCurrency'].includes(field.key)
}

watch(() => [props.show, props.assetType, props.position, props.requestedAction], () => {
  if (props.show) resetDraft()
}, { immediate: true })

function optionValue(option) {
  return typeof option === 'object' ? option.value : option
}

function optionLabel(option) {
  return typeof option === 'object' ? option.label : option
}

function updateField(field, event) {
  const raw = event.target.value
  draft[field.key] = field.type === 'number' ? (raw === '' ? null : Number(raw)) : raw
  if (errors[field.key]) delete errors[field.key]
}

function submit() {
  if (props.assetType !== 'bank_product' && !selectedProduct.value) {
    errors.product = '请先从数据源中选择金融产品'
    requestAnimationFrame(() => document.querySelector('.asset-product-search input')?.focus())
    return
  }
  const nextErrors = validateAssetTransaction(fields.value, draft)
  Object.keys(errors).forEach((key) => delete errors[key])
  Object.assign(errors, nextErrors)
  if (Object.keys(nextErrors).length) {
    requestAnimationFrame(() => document.querySelector('.asset-transaction-editor [aria-invalid="true"]')?.focus())
    return
  }

  const details = {
    ...Object.fromEntries(specificFields.value.map((field) => [field.key, draft[field.key]])),
    ...(draft.searchMetadata || {}),
  }
  emit('submit', {
    ...draft,
    assetType: props.assetType,
    quantity: Number(draft.quantity || 0),
    price: Number(draft.price || 0),
    fee: Number(draft.fee || 0),
    tradedAt: new Date(draft.tradedAt).toISOString(),
    details,
  })
}
</script>

<template>
  <n-modal v-model:show="modalVisible" :mask-closable="!saving" :close-on-esc="!saving">
    <div class="asset-transaction-editor" role="dialog" aria-modal="true" :aria-label="`增加一笔${assetMeta.label}持仓`">
      <header>
        <div>
          <span>{{ assetMeta.shortLabel }}持仓</span>
          <h2>增加一笔{{ assetMeta.label }}持仓</h2>
          <p>先匹配标准产品，再填写成交信息；保存后统一折算为人民币。</p>
        </div>
        <button type="button" class="icon-button" aria-label="关闭交易表单" :disabled="saving" @click="modalVisible = false">
          <n-icon :component="CloseOutline" />
        </button>
      </header>

      <div class="asset-transaction-editor__body">
        <fieldset>
          <legend><b>选择产品</b><span>{{ assetType === 'bank_product' ? '银行理财保留手动录入' : '名称和代码来自公开数据源' }}</span></legend>
          <div v-if="assetType === 'bank_product'" class="manual-product-tip">银行理财保留手动录入，请在下方填写产品名称、代码和发行机构。</div>
          <AssetProductSearch
            v-else
            v-model="selectedProduct"
            :asset-type="assetType"
            :busy="saving"
            :locked="Boolean(position)"
            @select="selectProduct"
          />
          <small v-if="errors.product" class="product-error" role="alert">{{ errors.product }}</small>
        </fieldset>

        <fieldset>
          <legend><b>交易信息</b><span>产品、方向与成交数据</span></legend>
          <div class="form-grid">
            <label v-for="field in commonFields" :key="field.key" :class="{ 'is-wide': field.wide }">
              <span>{{ field.label }}<i v-if="field.required">必填</i></span>
              <select
                v-if="field.type === 'select'"
                :value="draft[field.key]"
                :aria-invalid="Boolean(errors[field.key])"
                @change="updateField(field, $event)"
              >
                <option v-for="option in field.options" :key="optionValue(option)" :value="optionValue(option)">{{ optionLabel(option) }}</option>
              </select>
              <textarea
                v-else-if="field.type === 'textarea'"
                :value="draft[field.key]"
                :placeholder="field.placeholder"
                :aria-invalid="Boolean(errors[field.key])"
                rows="3"
                @input="updateField(field, $event)"
              />
              <input
                v-else
                :type="field.type"
                :value="draft[field.key]"
                :min="field.min"
                :step="field.step"
                :placeholder="field.placeholder"
                :readonly="fieldReadonly(field)"
                :aria-invalid="Boolean(errors[field.key])"
                @input="updateField(field, $event)"
              >
              <small v-if="errors[field.key]" role="alert">{{ errors[field.key] }}</small>
            </label>
          </div>
        </fieldset>

        <fieldset v-if="specificFields.length">
          <legend><b>{{ assetMeta.shortLabel }}专属信息</b><span>用于准确计算市值、收益与到期安排</span></legend>
          <div class="form-grid">
            <label v-for="field in specificFields" :key="field.key" :class="{ 'is-wide': field.wide }">
              <span>{{ field.label }}<i v-if="field.required">必填</i></span>
              <select
                v-if="field.type === 'select'"
                :value="draft[field.key]"
                :aria-invalid="Boolean(errors[field.key])"
                @change="updateField(field, $event)"
              >
                <option value="" disabled>请选择</option>
                <option v-for="option in field.options" :key="optionValue(option)" :value="optionValue(option)">{{ optionLabel(option) }}</option>
              </select>
              <input
                v-else
                :type="field.type"
                :value="draft[field.key]"
                :min="field.min"
                :step="field.step"
                :placeholder="field.placeholder"
                :readonly="fieldReadonly(field)"
                :aria-invalid="Boolean(errors[field.key])"
                @input="updateField(field, $event)"
              >
              <small v-if="errors[field.key]" role="alert">{{ errors[field.key] }}</small>
            </label>
          </div>
        </fieldset>
      </div>

      <footer>
        <p><span aria-hidden="true" />交易日志是持仓计算的唯一账本，修改会同步影响成本与收益。</p>
        <div>
          <button type="button" class="secondary-button" :disabled="saving" @click="modalVisible = false">取消</button>
          <button type="button" class="primary-button" :disabled="saving" @click="submit">{{ saving ? '正在保存…' : '保存持仓' }}</button>
        </div>
      </footer>
    </div>
  </n-modal>
</template>

<style scoped>
.asset-transaction-editor { width:min(760px,calc(100vw - 32px));max-height:min(88dvh,860px);display:flex;flex-direction:column;overflow:hidden;color:var(--workspace-text);background:var(--workspace-surface-elevated);border:1px solid var(--workspace-border-strong);border-radius:var(--workspace-radius-md);box-shadow:var(--workspace-shadow);backdrop-filter:var(--workspace-glass); }
header { flex:0 0 auto;display:flex;align-items:flex-start;justify-content:space-between;gap:18px;padding:17px 19px;border-bottom:1px solid var(--workspace-border);background:color-mix(in srgb,var(--workspace-accent-soft) 38%,var(--workspace-surface-solid)); }
header span { color:var(--workspace-accent);font-size:10px;font-weight:700; } header h2 { margin:3px 0 0;font-size:18px;letter-spacing:0; } header p { margin:4px 0 0;color:var(--workspace-text-muted);font-size:10px;line-height:1.5; }
.icon-button { width:36px;height:36px;display:grid;place-items:center;flex:0 0 auto;padding:0;border:1px solid var(--workspace-border);border-radius:var(--workspace-radius-sm);color:var(--workspace-text-muted);background:var(--workspace-surface-solid);font-size:18px;cursor:pointer; }.icon-button:hover { color:var(--workspace-accent);border-color:var(--workspace-accent); }
.asset-transaction-editor__body { min-height:0;overflow:auto;padding:4px 19px 18px;scrollbar-gutter:stable; }
fieldset { margin:0;padding:17px 0 4px;border:0;border-bottom:1px solid var(--workspace-border); } fieldset:last-child { border-bottom:0; }
legend { width:100%;display:flex;align-items:baseline;gap:9px;padding:0;margin-bottom:12px; } legend b { font-size:13px; } legend span { color:var(--workspace-text-muted);font-size:9px; }
.form-grid { display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:12px 14px; }.form-grid label { min-width:0;display:block; }.form-grid label.is-wide { grid-column:1/-1; }
label>span { min-height:18px;display:flex;align-items:center;justify-content:space-between;gap:8px;color:var(--workspace-text-secondary);font-size:10px;font-weight:650; } label>span i { color:var(--workspace-accent);font-size:8px;font-style:normal;font-weight:500; }
input,select,textarea { width:100%;min-height:40px;margin-top:4px;padding:8px 10px;border:1px solid var(--workspace-border);border-radius:var(--workspace-radius-sm);outline:0;color:var(--workspace-text);background:var(--workspace-surface-solid);font:inherit;font-size:11px;transition:border-color 150ms ease,box-shadow 150ms ease; } textarea { min-height:72px;resize:vertical;line-height:1.55; } input:focus,select:focus,textarea:focus { border-color:var(--workspace-accent);box-shadow:var(--workspace-focus-ring); } [aria-invalid="true"] { border-color:var(--workspace-positive); }
input[readonly]{color:var(--workspace-text-secondary);background:var(--workspace-surface-muted);cursor:default}
label>small { display:block;margin-top:4px;color:var(--workspace-positive);font-size:9px; }
.manual-product-tip{padding:11px 12px;border:1px solid var(--workspace-border);border-radius:var(--workspace-radius-sm);color:var(--workspace-text-secondary);background:var(--workspace-surface-muted);font-size:12px;line-height:1.6}.product-error{display:block;margin-top:6px;color:var(--workspace-negative);font-size:11px}
footer { flex:0 0 auto;display:flex;align-items:center;justify-content:space-between;gap:18px;padding:11px 16px;border-top:1px solid var(--workspace-border);background:var(--workspace-surface-solid); } footer p { min-width:0;display:flex;align-items:center;gap:7px;margin:0;color:var(--workspace-text-muted);font-size:9px;line-height:1.45; } footer p span { width:7px;height:7px;flex:0 0 auto;border-radius:50%;background:var(--workspace-accent);box-shadow:0 0 0 4px var(--workspace-accent-muted); } footer>div { display:flex;gap:8px; }
.secondary-button,.primary-button { min-height:36px;padding:0 14px;border-radius:var(--workspace-radius-sm);font-size:11px;font-weight:650;cursor:pointer; }.secondary-button { border:1px solid var(--workspace-border);color:var(--workspace-text-secondary);background:var(--workspace-surface-solid); }.primary-button { border:1px solid var(--workspace-accent);color:#fff;background:var(--workspace-accent); }.primary-button:hover { background:var(--workspace-accent-hover); }.secondary-button:hover { color:var(--workspace-accent);border-color:var(--workspace-accent); }.secondary-button:disabled,.primary-button:disabled,.icon-button:disabled { opacity:.5;cursor:wait; }
@media (max-width:640px), (min-resolution:2.5dppx) { .asset-transaction-editor { width:min(620px,calc(100vw - 16px));max-height:92dvh; } header { padding:13px 14px; }.asset-transaction-editor__body { padding:2px 14px 14px; }.form-grid { grid-template-columns:minmax(0,1fr); }.form-grid label.is-wide { grid-column:auto; } footer { align-items:stretch;flex-direction:column;gap:9px; } footer>div { justify-content:flex-end; } footer p { max-width:none; } }
</style>
