<script setup>
import { ref } from 'vue'
import { callPortfolioApi, formatMoney } from './portfolio-api.mjs'

const emit = defineEmits(['imported'])
const preview = ref(null)
const loading = ref(false)
const error = ref('')
const lastImportedBatch = ref('')

async function selectFile() {
  error.value = ''
  const path = await callPortfolioApi('SelectBankProductCSV')
  if (!path) return
  loading.value = true
  try { preview.value = await callPortfolioApi('PreviewBankProductCSV', path) }
  catch (e) { error.value = e?.message || String(e) }
  finally { loading.value = false }
}

async function importRows() {
  if (!preview.value || preview.value.invalid > 0) return
  loading.value = true
  error.value = ''
  try {
    await callPortfolioApi('ImportBankProductCSV', preview.value.batchId, preview.value.path, preview.value.rows)
    lastImportedBatch.value = preview.value.batchId
    emit('imported')
    preview.value = null
  } catch (e) { error.value = e?.message || String(e) }
  finally { loading.value = false }
}

async function rollback() {
  const batchId = lastImportedBatch.value || preview.value?.batchId
  if (!batchId) return
  loading.value = true
  try { await callPortfolioApi('RollbackAssetImport', batchId); emit('imported'); preview.value = null; lastImportedBatch.value = '' }
  catch (e) { error.value = e?.message || String(e) }
  finally { loading.value = false }
}
</script>

<template>
  <section class="bank-import">
    <header>
      <div><h2>批量导入银行理财</h2><p>支持 UTF-8 与 GB18030 CSV，写入前先逐行校验。</p></div>
      <button type="button" :disabled="loading" @click="selectFile">选择 CSV</button>
    </header>
    <p v-if="error" class="bank-import__error">{{ error }}</p>
    <div v-if="preview" class="bank-import__summary">
      <span>有效 <b>{{ preview.valid }}</b></span><span>异常 <b>{{ preview.invalid }}</b></span><span>批次 {{ preview.batchId }}</span>
      <button type="button" :disabled="loading || preview.invalid > 0" @click="importRows">确认导入</button>
    </div>
    <div v-if="lastImportedBatch" class="bank-import__rollback"><span>最近导入批次 {{ lastImportedBatch }}</span><button type="button" class="secondary" :disabled="loading" @click="rollback">撤销本次导入</button></div>
    <div v-if="preview" class="bank-import__table" tabindex="0">
      <table><thead><tr><th>行</th><th>机构</th><th>产品</th><th>本金/市值</th><th>币种</th><th>校验</th></tr></thead>
      <tbody><tr v-for="row in preview.rows" :key="row.rowNo"><td>{{ row.rowNo }}</td><td>{{ row.institution }}</td><td>{{ row.productName }}<small>{{ row.productCode }}</small></td><td>{{ formatMoney(row.principal || row.quantity * row.nav, row.currency || 'CNY') }}</td><td>{{ row.currency }}</td><td :class="{ invalid: row.errors?.length }">{{ row.errors?.join('；') || '通过' }}</td></tr></tbody></table>
    </div>
  </section>
</template>

<style scoped>
.bank-import{min-width:0;border:1px solid var(--workspace-border);border-radius:var(--workspace-radius-md);background:var(--workspace-surface-solid);overflow:hidden}.bank-import header{display:flex;align-items:center;justify-content:space-between;gap:12px;padding:12px 14px;border-bottom:1px solid var(--workspace-border)}h2{margin:0;font-size:14px}p{margin:3px 0 0;color:var(--workspace-text-muted);font-size:10px}button{min-height:32px;padding:0 11px;border:1px solid var(--workspace-accent);border-radius:var(--workspace-radius-sm);color:white;background:var(--workspace-accent);cursor:pointer}button.secondary{color:var(--workspace-text-secondary);border-color:var(--workspace-border);background:var(--workspace-surface-solid)}button:disabled{opacity:.5;cursor:not-allowed}.bank-import__error{padding:9px 14px;color:var(--workspace-negative)}.bank-import__summary{display:flex;align-items:center;gap:14px;padding:10px 14px;border-bottom:1px solid var(--workspace-border);font-size:11px}.bank-import__summary span:nth-child(3){margin-right:auto;color:var(--workspace-text-muted)}.bank-import__table{max-height:360px;overflow:auto}.bank-import table{width:100%;min-width:720px;border-collapse:collapse;font-size:11px}th,td{padding:9px 11px;border-bottom:1px solid var(--workspace-border);text-align:left;white-space:nowrap}th{position:sticky;top:0;background:var(--workspace-surface-muted)}td small{display:block;color:var(--workspace-text-muted)}td.invalid{color:var(--workspace-negative)}@media(max-width:720px),(min-resolution:2.5dppx){.bank-import__summary{align-items:flex-start;flex-wrap:wrap}.bank-import__summary span:nth-child(3){width:100%}}
.bank-import__rollback{display:flex;align-items:center;gap:14px;padding:10px 14px;border-bottom:1px solid var(--workspace-border);font-size:11px}.bank-import__rollback span{margin-right:auto;color:var(--workspace-text-muted)}
</style>
