<script setup>
import {computed} from 'vue'
import {NModal} from 'naive-ui'
import StockLightweightKlineChart from './StockLightweightKlineChart.vue'

const props = defineProps({
  show: {type: Boolean, default: false},
  code: {type: String, default: ''},
  stockName: {type: String, default: ''},
  darkTheme: {type: Boolean, default: false},
  chartHeight: {type: Number, default: 500},
})

const emit = defineEmits(['update:show'])

const title = computed(() => {
  const identity = [props.stockName, props.code].filter(Boolean).join(' · ')
  return `${identity || '股票'} — 多周期K线`
})
</script>

<template>
  <NModal
    :show="show"
    preset="card"
    :title="title"
    :mask-closable="true"
    class="stock-kline-modal"
    :style="{
      width: 'min(1100px, 96vw)',
      maxWidth: '96vw',
      boxSizing: 'border-box',
    }"
    :content-style="{
      maxHeight: 'min(85vh, 820px)',
      overflowY: 'auto',
      overflowX: 'hidden',
      minWidth: 0,
      boxSizing: 'border-box',
    }"
    @update:show="emit('update:show', $event)"
  >
    <StockLightweightKlineChart
      v-if="show && code"
      :key="code"
      :code="code"
      :stock-name="stockName"
      :dark-theme="darkTheme"
      :chart-height="chartHeight"
    />
  </NModal>
</template>

<style scoped>
:deep(.stock-kline-modal) {
  min-width: 0;
}
</style>
