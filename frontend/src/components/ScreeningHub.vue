<script setup>
import { ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import DuoKongLineSelector from './DuoKongLineSelector.vue'
import AllStockList from './allStockList.vue'
import SelectStock from './SelectStock.vue'
import { normaliseScreeningTab } from '../navigation/tab-sync.mjs'

const route = useRoute()
const router = useRouter()
const active = ref(normaliseScreeningTab(route.query))
const screeningTabs = [
  { key: 'pattern', label: '形态选股' },
  { key: 'duo-kong', label: '多空线选股' },
  { key: 'strategy', label: '策略选股' },
]

watch(() => [route.query.mode, route.query.tab], () => {
  active.value = normaliseScreeningTab(route.query)
})

function updateTab(mode) {
  active.value = mode
  if (route.query.mode !== mode) {
    router.replace({ name: 'screening', query: { ...route.query, mode } })
  }
}
</script>

<template>
  <section class="screening">
    <div class="screening__heading">
      <h1>按形态、趋势和策略筛选 A 股</h1>
    </div>
    <div class="screening__content">
      <AllStockList v-if="active === 'pattern'" />
      <DuoKongLineSelector v-else-if="active === 'duo-kong'" />
      <SelectStock v-else />
    </div>
  </section>
</template>

<style scoped>
.screening { min-height:100%; padding:20px; background:var(--workspace-canvas); }
.screening__heading { display:flex; align-items:center; min-height:36px; margin-bottom:12px; }
.screening__heading h1 { margin:0; color:var(--workspace-text); font-size:20px; font-weight:650; }
.screening__content { min-width:0; }
@media (max-width:719px) { .screening { padding:14px; } .screening__heading h1 { font-size:18px; } }
</style>
