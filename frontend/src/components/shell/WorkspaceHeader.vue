<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { PRIMARY_SECTIONS, getSectionForRoute } from '../../navigation/workspace.mjs'

defineEmits(['search'])
const route = useRoute()
const section = computed(() => PRIMARY_SECTIONS.find((item) => item.key === getSectionForRoute(route)))
const sectionBadges = { overview: '总览', watchlist: '自选', market: '资讯', screening: '选股', holding: '持仓', ai: '智能', settings: '设置', about: '关于' }
const sectionBadge = computed(() => sectionBadges[section.value?.key] || '总览')
const context = computed(() => {
  if (route.name === 'klineAnalysis') return '多周期价格结构与指标共振分析'
  if (route.name === 'radar') return '盘中异动发现、提醒与一键跳转分析'
  if (route.name === 'screening') return '从全市场筛选值得进一步研究的候选股'
  if (route.name === 'holdingCenter') return '统一查看各类金融产品的市值、成本与收益'
  if (route.name === 'aiConsole') return '管理 AI 研究证据、报告与股票推荐记录'
  if (route.name === 'about') return '了解项目、功能与安全的自动更新机制'
  return '本地优先的 A 股研究与决策工作区'
})
</script>

<template>
  <header class="workspace-header">
    <div class="workspace-header__context">
      <span class="workspace-header__eyebrow">{{ sectionBadge }}</span>
      <div>
        <strong>{{ section?.label || '总览' }}</strong>
        <span>{{ context }}</span>
      </div>
    </div>
    <div class="workspace-header__actions"><slot name="actions" /></div>
  </header>
</template>

<style scoped>
.workspace-header {
  min-height: var(--workspace-toolbar-height);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 0 20px;
  color: var(--workspace-text);
  background: color-mix(in srgb, var(--workspace-surface) 92%, transparent);
  border-bottom: 1px solid var(--workspace-border);
  backdrop-filter: var(--workspace-glass);
}
.workspace-header__context { min-width: 0; display: flex; align-items: center; gap: 11px; }
.workspace-header__eyebrow { min-width: 58px; padding: 4px 8px; border: 1px solid color-mix(in srgb, var(--workspace-accent) 24%, transparent); border-radius: 4px; color: var(--workspace-accent); background: var(--workspace-accent-soft); font-size: 11px; font-weight: 760; text-align: center; }
.workspace-header__context > div { min-width: 0; display: flex; align-items: baseline; gap: 10px; }
.workspace-header__context strong { color: var(--workspace-text); font-size: 15px; font-weight: 720; white-space: nowrap; }
.workspace-header__context div span { overflow: hidden; color: var(--workspace-text-muted); font-size: 11px; white-space: nowrap; text-overflow: ellipsis; }
.workspace-header__actions { display: flex; align-items: center; gap: 8px; }
@media (max-width: 719px) {
  .workspace-header { padding: 46px 12px 0; min-height: calc(var(--workspace-toolbar-height) + 46px); }
  .workspace-header__context div span { display: none; }
  .workspace-header__eyebrow { display: none; }
}
</style>
