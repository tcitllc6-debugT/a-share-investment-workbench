<script setup>
import {computed, h, onBeforeMount, onBeforeUnmount, onMounted, reactive, ref} from 'vue'
import {
  Follow,
  CancelDuoKongLineScan,
  GetAllConcepts,
  GetAllIndustries,
  GetConfig,
  GetDuoKongLineScanTask,
  GetDuoKongLineSelections,
  StartDuoKongLineScan
} from '../../wailsjs/go/main/App'
import {
  NButton,
  NCard,
  NCheckbox,
  NDataTable,
  NEllipsis,
  NInput,
  NInputNumber,
  NModal,
  NProgress,
  NRadioButton,
  NRadioGroup,
  NSelect,
  NSpace,
  NTag,
  NText,
  useMessage
} from 'naive-ui'
import StockLightweightKlineChart from './StockLightweightKlineChart.vue'

const message = useMessage()
const loading = ref(false)
const rows = ref([])
const scanTask = ref(null)
const multiKlineModalShow = ref(false)
const multiKlineCode = ref('')
const multiKlineName = ref('')
let pollTimer = null
const viewportWidth = ref(window.innerWidth)
const isCompact = computed(() => viewportWidth.value < 760)

function updateViewportWidth() {
  viewportWidth.value = window.innerWidth
}

const editorDataRef = reactive({
  darkTheme: false
})

onBeforeMount(() => {
  GetConfig().then(result => {
    editorDataRef.darkTheme = !!result?.darkTheme
  }).catch(() => {
    editorDataRef.darkTheme = false
  })
})

const form = reactive({
  signal: 'fine_cross_up',
  keyword: '',
  minDeviation: 0,
  marketBoards: ['ALL'],
  industries: [],
  concepts: [],
  onlyFirstSignalToday: false,
  minAboveDays: null,
  keepAboveAfterCross: false,
})

const signalGroups = [
  {label: '事件', options: [
    {label: '细线上穿粗线', value: 'fine_cross_up', type: 'success'},
    {label: '细线下穿粗线', value: 'fine_cross_down', type: 'error'},
  ]},
  {label: '状态', options: [
    {label: '细线在粗线上方', value: 'fine_above_thick', type: 'info'},
    {label: '细线在粗线下方', value: 'fine_below_thick', type: 'warning'},
  ]},
  {label: '趋势', options: [
    {label: '双线上行', value: 'both_rising', type: 'success'},
    {label: '双线下行', value: 'both_falling', type: 'error'},
    {label: '细线拐头向上', value: 'fine_turn_up', type: 'success'},
    {label: '细线拐头向下', value: 'fine_turn_down', type: 'error'},
    {label: '粗线拐头向上', value: 'thick_turn_up', type: 'success'},
    {label: '粗线拐头向下', value: 'thick_turn_down', type: 'error'},
  ]},
]
const signalOptions = signalGroups.flatMap(group => group.options)

const boardOptions = [
  {label: '全部市场', value: 'ALL'},
  {label: '科创板', value: 'STAR'},
  {label: '创业板', value: 'CHINEXT'},
  {label: '主板', value: 'MAIN'},
]

const industryOptions = ref([])
const conceptOptions = ref([])

const signalLabel = (value) => signalOptions.find(item => item.value === value)?.label || value
const signalType = (value) => signalOptions.find(item => item.value === value)?.type || 'default'

function formatNumber(value, digits = 2) {
  const n = Number(value)
  if (Number.isNaN(n)) return '--'
  return n.toFixed(digits)
}

function toEastMoneyCode(row) {
  const code = String(row.code || '').trim()
  if (/\.(SH|SZ|BJ|HK|US|SS)$/i.test(code)) return code.toUpperCase()

  const raw = String(row.kLineCode || '').trim()
  const lower = raw.toLowerCase()
  if (lower.startsWith('sh')) return `${lower.slice(2)}.SH`
  if (lower.startsWith('sz')) return `${lower.slice(2)}.SZ`
  if (lower.startsWith('bj')) return `${lower.slice(2)}.BJ`
  if (lower.startsWith('hk')) return `${lower.slice(2).toUpperCase()}.HK`
  if (lower.startsWith('us')) return `${lower.slice(2).toUpperCase()}.US`
  return raw.toUpperCase()
}

function openKLine(row) {
  const code = toEastMoneyCode(row)
  if (!code) {
    message.warning('当前代码暂不支持本地K线图')
    return
  }
  multiKlineCode.value = code
  multiKlineName.value = row.name || ''
  multiKlineModalShow.value = true
}

function openMultiPeriod(row) {
  openKLine(row)
}

function followStock(row) {
  const code = row.kLineCode || ''
  if (!code) {
    message.warning('股票代码为空，无法加入自选')
    return
  }
  Follow(code).then(res => {
    if (res === '关注成功') {
      message.success('已加入自选')
    } else {
      message.info(res)
    }
  }).catch(err => {
    message.error(`加入自选失败：${err}`)
  })
}

const columns = [
  {
    title: '代码',
    key: 'code',
    width: 110,
    fixed: 'left',
    render(row) {
      return h(NButton, {text: true, type: 'primary', onClick: () => openKLine(row)}, {default: () => row.code || row.kLineCode})
    }
  },
  {
    title: '名称',
    key: 'name',
    width: 120,
    fixed: 'left',
    render(row) {
      return h(NText, {strong: true}, {default: () => row.name || '--'})
    }
  },
  {
    title: '条件',
    key: 'signal',
    width: 130,
    render(row) {
      return h(NTag, {type: signalType(row.signal), bordered: false, size: 'small'}, {default: () => signalLabel(row.signal)})
    }
  },
  {title: '信号日', key: 'signalDate', width: 110},
  {
    title: '距今',
    key: 'signalAge',
    width: 80,
    sorter: (left, right) => Number(left.signalAge || 0) - Number(right.signalAge || 0),
    render(row) {
      return h(NText, {depth: 3}, {default: () => `${Number(row.signalAge || 0)}日`})
    }
  },
  {
    title: '收盘',
    key: 'close',
    width: 90,
    render(row) {
      return h(NText, {type: 'info'}, {default: () => formatNumber(row.close)})
    }
  },
  {
    title: '细线',
    key: 'fine',
    width: 90,
    render(row) {
      return h(NText, {type: row.fine >= row.thick ? 'error' : 'success'}, {default: () => formatNumber(row.fine)})
    }
  },
  {
    title: '粗线',
    key: 'thick',
    width: 90,
    render(row) {
      return h(NText, {type: 'warning'}, {default: () => formatNumber(row.thick)})
    }
  },
  {
    title: '偏离率',
    key: 'deviation',
    width: 90,
    sorter: (left, right) => Number(left.deviation || 0) - Number(right.deviation || 0),
    render(row) {
      const value = Number(row.deviation || 0)
      return h(NText, {type: value >= 0 ? 'error' : 'success'}, {default: () => `${formatNumber(value)}%`})
    }
  },
  {
    title: '量比',
    key: 'volumeRatio',
    width: 80,
    sorter: (left, right) => Number(left.volumeRatio || 0) - Number(right.volumeRatio || 0),
    render(row) {
      return h(NText, {type: Number(row.volumeRatio || 0) >= 1 ? 'error' : 'default'}, {default: () => formatNumber(row.volumeRatio)})
    }
  },
  {
    title: '行业',
    key: 'industry',
    width: 140,
    render(row) {
      return h(NEllipsis, {tooltip: true}, {default: () => row.industry || '--'})
    }
  },
  {
    title: '概念',
    key: 'concept',
    width: 220,
    render(row) {
      return h(NEllipsis, {tooltip: true}, {default: () => row.concept || '--'})
    }
  },
  {
    title: '操作',
    key: 'actions',
    width: 210,
    fixed: 'right',
    render(row) {
      return h(NSpace, {size: 6}, {
        default: () => [
          h(NButton, {size: 'tiny', type: 'primary', secondary: true, onClick: () => followStock(row)}, {default: () => '自选'}),
          h(NButton, {size: 'tiny', type: 'warning', secondary: true, onClick: () => openKLine(row)}, {default: () => '日K'}),
          h(NButton, {size: 'tiny', type: 'success', secondary: true, onClick: () => openMultiPeriod(row)}, {default: () => '多周期'})
        ]
      })
    }
  }
]

const compactColumnKeys = new Set(['code', 'name', 'signal', 'close', 'actions'])
const visibleColumns = computed(() => isCompact.value
  ? columns.filter(column => compactColumnKeys.has(column.key))
  : columns
)

function buildCondition() {
  const marketBoards = form.marketBoards.includes('ALL') ? ['ALL'] : form.marketBoards
  return {
    signal: form.signal,
    days: 120,
    signalDays: 3,
    limit: 0,
    keyword: form.keyword || '',
    minDeviation: Number(form.minDeviation) || 0,
    marketBoard: marketBoards[0] || 'ALL',
    marketBoards,
    industries: form.industries,
    concepts: form.concepts,
    onlyFirstSignalToday: form.onlyFirstSignalToday,
    minAboveDays: Number(form.minAboveDays) || 0,
    keepAboveAfterCross: form.keepAboveAfterCross,
  }
}

function stopPolling() {
  if (pollTimer) {
    clearTimeout(pollTimer)
    pollTimer = null
  }
}

function updateTask(task) {
  scanTask.value = task
  if (task?.results) {
    rows.value = task.results
  }
}

async function pollTask(taskId) {
  try {
    const task = await GetDuoKongLineScanTask(taskId)
    updateTask(task)
    if (!task) {
      throw new Error('扫描任务不存在')
    }
    if (task.status === 'done') {
      loading.value = false
      if ((task.results || []).length === 0) {
        message.info('没有筛到符合条件的股票')
      } else {
        message.success(`扫描完成，共 ${task.results.length} 条结果`)
      }
      return
    }
    if (task.status === 'error') {
      loading.value = false
      message.error(`多空线扫描失败：${task.error || '未知错误'}`)
      return
    }
    pollTimer = setTimeout(() => pollTask(taskId), 800)
  } catch (err) {
    loading.value = false
    message.error(`读取扫描进度失败：${err}`)
  }
}

async function scan() {
  stopPolling()
  loading.value = true
  rows.value = []
  scanTask.value = {status: 'running', progress: 0, total: 0, results: []}
  try {
    const condition = buildCondition()
    const taskId = await StartDuoKongLineScan(condition)
    if (!taskId) {
      const result = await GetDuoKongLineSelections(condition)
      rows.value = result || []
      loading.value = false
      return
    }
    pollTask(taskId)
  } catch (err) {
    loading.value = false
    message.error(`多空线扫描失败：${err}`)
  }
}

async function cancelScan() {
  const taskId = scanTask.value?.id
  if (!taskId) return
  const cancelled = await CancelDuoKongLineScan(taskId)
  if (cancelled) {
    message.info('正在取消扫描')
  }
}

function progressPercentage() {
  const total = Number(scanTask.value?.total || 0)
  const progress = Number(scanTask.value?.progress || 0)
  if (total <= 0) return loading.value ? 1 : 0
  return Math.min(100, Math.round(progress / total * 100))
}

async function loadFilterOptions() {
  const [industries, concepts] = await Promise.all([
    GetAllIndustries().catch(() => []),
    GetAllConcepts().catch(() => []),
  ])
  industryOptions.value = (industries || []).map(value => ({label: value, value}))
  conceptOptions.value = (concepts || []).map(value => ({label: value, value}))
}

onMounted(async () => {
	window.addEventListener('resize', updateViewportWidth)
  await loadFilterOptions()
  scan()
})
onBeforeUnmount(() => {
  window.removeEventListener('resize', updateViewportWidth)
  stopPolling()
})
</script>

<template>
  <div class="duo-kong-selector">
    <n-card size="small" class="filter-panel">
      <n-space vertical size="small">
        <div class="panel-title">
          <n-text strong>多空线选股</n-text>
          <n-text depth="3">默认扫描全市场，筛出最近 3 个交易日内出现所选信号的股票</n-text>
        </div>

        <n-radio-group v-model:value="form.signal" name="duo-kong-signal">
          <n-space class="signal-row" wrap :size="0">
            <n-radio-button
              v-for="item in signalOptions"
              :key="item.value"
              :value="item.value"
            >
              {{ item.label }}
            </n-radio-button>
          </n-space>
        </n-radio-group>

        <div class="filter-row">
          <label>板块</label>
          <n-select v-model:value="form.marketBoards" :options="boardOptions" multiple clearable />
          <label>行业</label>
          <n-select v-model:value="form.industries" :options="industryOptions" multiple filterable clearable />
          <label>概念</label>
          <n-select v-model:value="form.concepts" :options="conceptOptions" multiple filterable clearable />
          <label>最小偏离</label>
          <n-input-number v-model:value="form.minDeviation" :min="0" :max="50" :step="0.5">
            <template #suffix>%</template>
          </n-input-number>
        </div>

        <n-space class="behavior-row" size="large" wrap>
          <n-checkbox v-model:checked="form.onlyFirstSignalToday">只看今天首次触发</n-checkbox>
          <n-checkbox v-model:checked="form.keepAboveAfterCross">上穿后未跌回</n-checkbox>
          <n-space size="small" align="center">
            <label>连续站上</label>
            <n-input-number v-model:value="form.minAboveDays" :min="0" :max="120" placeholder="最少天数" style="width: 120px">
              <template #suffix>日</template>
            </n-input-number>
          </n-space>
          <n-text depth="3">固定计算 120 日K，信号窗口固定最近 3 个交易日；默认排除 ST、停牌、新股和北交所。</n-text>
        </n-space>

        <div class="filter-row search-row">
          <n-input v-model:value="form.keyword" clearable placeholder="代码、名称、行业、概念" @keyup.enter="scan" />
          <n-button type="primary" :loading="loading" @click="scan">开始扫描</n-button>
          <n-button v-if="loading" tertiary type="warning" @click="cancelScan">取消</n-button>
        </div>

        <div v-if="loading || scanTask" class="scan-progress">
          <n-progress
            type="line"
            :percentage="progressPercentage()"
            :show-indicator="false"
            processing
          />
          <n-text depth="3">
            已扫描 {{ scanTask?.progress || 0 }} / {{ scanTask?.total || 0 }}，当前命中 {{ rows.length }} 只
          </n-text>
        </div>
      </n-space>
    </n-card>

    <n-data-table
      class="duo-kong-results"
      size="small"
      striped
      flex-height
      :loading="loading"
      :columns="visibleColumns"
      :data="rows"
      :scroll-x="isCompact ? 760 : 1800"
      :pagination="{ pageSize: 20 }"
      style="--wails-draggable:no-drag"
    />

    <n-modal
      v-model:show="multiKlineModalShow"
      :title="(multiKlineName || '') + ' — 多周期K线'"
      preset="card"
      style="width: min(1100px, 96vw); max-width: 96vw; box-sizing: border-box"
      :content-style="{
        maxHeight: 'min(88dvh, 820px)',
        overflowY: 'auto',
        overflowX: 'hidden',
        minWidth: 0,
        boxSizing: 'border-box',
      }"
    >
      <StockLightweightKlineChart
        v-if="multiKlineModalShow && multiKlineCode"
        :key="'duokong-' + multiKlineCode"
        :code="multiKlineCode"
        :stock-name="multiKlineName"
        :dark-theme="editorDataRef.darkTheme"
        :chart-height="isCompact ? 320 : 500"
      />
    </n-modal>
  </div>
</template>

<style scoped>
.duo-kong-selector {
  display: grid;
  gap: 10px;
}

.filter-panel {
  --wails-draggable: no-drag;
}

.panel-title {
  display: flex;
  align-items: baseline;
  gap: 10px;
}

.filter-row {
  display: grid;
  grid-template-columns: auto minmax(105px, 0.8fr) auto minmax(120px, 1fr) auto minmax(160px, 1.35fr) auto minmax(95px, 0.55fr);
  align-items: center;
  gap: 8px;
}

.signal-row :deep(.n-radio-button) {
  white-space: nowrap;
}

.behavior-row {
  align-items: center;
}

.search-row {
  grid-template-columns: minmax(220px, 420px) auto;
  justify-content: start;
}

.scan-progress {
  display: grid;
  grid-template-columns: minmax(180px, 360px) auto;
  align-items: center;
  gap: 10px;
}

.duo-kong-results {
  height: clamp(320px, calc(100dvh - 315px), 720px);
}

label {
  color: var(--n-text-color-2);
  white-space: nowrap;
}

@media (max-width: 900px) {
  .panel-title {
    display: grid;
    gap: 4px;
  }

  .filter-row,
  .search-row,
  .scan-progress {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 759px) {
  .duo-kong-selector {
    gap: 8px;
  }

  .signal-row {
    display: flex;
    overflow-x: auto;
    padding-bottom: 2px;
  }

  .behavior-row {
    align-items: flex-start;
  }

  .duo-kong-results {
    height: clamp(280px, calc(100dvh - 420px), 480px);
  }
}
</style>
