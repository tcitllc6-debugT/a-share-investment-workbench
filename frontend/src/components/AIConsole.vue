<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import ResearchReport from './researchReport.vue'
import AiRecommendStocksList from './aiRecommendStocksList.vue'
import AIConfiguration from './AIConfiguration.vue'
import PromptTemplateList from './promptTemplateList.vue'

const route = useRoute()
const router = useRouter()
const view = computed(() => {
  if (route.query.view === 'recommendations') return 'recommendations'
  if (route.query.view === 'templates') return 'templates'
  if (route.query.view === 'configuration') return 'configuration'
  return 'analysis'
})

const viewCopy = computed(() => ({
  analysis: '复盘研究结论，把证据和判断留在可追溯的记录里',
  recommendations: '查看 AI 曾经给出的候选股、依据和后续表现',
  templates: '管理本地提示词模板，让每次分析按你熟悉的框架展开',
  configuration: '连接模型服务，指定分析时默认使用的模型',
}[view.value]))

function switchView(nextView) {
  router.replace({ name: 'aiConsole', query: { ...route.query, view: nextView } })
}

</script>

<template>
  <section class="ai-console">
    <div class="ai-console__intro">
      <div><h1>{{ viewCopy }}</h1><p v-if="view === 'templates'">本地提示词模板</p></div>
      <button v-if="!['configuration', 'templates'].includes(view)" type="button" @click="switchView('configuration')">配置模型服务</button>
    </div>
    <ResearchReport v-if="view === 'analysis'" />
    <AiRecommendStocksList v-else-if="view === 'recommendations'" />
    <PromptTemplateList v-else-if="view === 'templates'" />
    <AIConfiguration v-else />
  </section>
</template>

<style scoped>
.ai-console { min-height:100%; padding:20px; }
.ai-console__intro { display:flex; align-items:center; justify-content:space-between; gap:16px; min-height:36px; margin:0 0 12px; }
.ai-console__intro h1 { margin:0; color:var(--workspace-text); font-size:20px; font-weight:650; }
.ai-console__intro p { margin:4px 0 0; color:var(--workspace-text-secondary); font-size:12px; }
.ai-console__intro button { min-height:32px; padding:0 11px; border:1px solid var(--workspace-border); border-radius:var(--workspace-radius-sm); color:var(--workspace-accent); background:var(--workspace-surface-solid); font-size:12px; cursor:pointer; }
.ai-console__intro button:hover { border-color:var(--workspace-accent); background:var(--workspace-accent-soft); }
@media (max-width:719px) { .ai-console { padding:14px; } .ai-console__intro { align-items:flex-start; flex-direction:column; gap:8px; } .ai-console__intro h1 { font-size:18px; } }
</style>
