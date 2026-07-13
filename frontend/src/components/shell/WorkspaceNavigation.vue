<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  AnalyticsOutline,
  BarChartSharp,
  CloseOutline,
  DiamondOutline,
  EaselSharp,
  Flag,
  Flame,
  FlaskOutline,
  HomeOutline,
  InformationOutline,
  NewspaperOutline,
  Pulse,
  ReorderTwoOutline,
  SearchOutline,
  SettingsOutline,
  SparklesOutline,
  StarOutline,
  StatsChartOutline,
  TimeOutline,
  Wallet,
  WarningOutline,
} from '@vicons/ionicons5'
import { NIcon } from 'naive-ui'
import { buildSecondaryLocation, getSectionForRoute, getSecondaryItems } from '../../navigation/workspace.mjs'
import { normaliseMarketTab, normaliseScreeningTab } from '../../navigation/tab-sync.mjs'

const props = defineProps({
  presentation: { type: String, default: 'wide' },
  secondaryOpen: { type: Boolean, default: false },
})
const emit = defineEmits(['navigate', 'close'])
const route = useRoute()
const router = useRouter()
const sectionKey = computed(() => getSectionForRoute(route))
const items = computed(() => getSecondaryItems(sectionKey.value))
const headings = {
  ai: '智能投研', screening: '选股工具', market: '市场资讯', holding: '账户与持仓',
  watchlist: '股票自选', settings: '应用设置', overview: '工作总览', about: '关于应用',
}

const itemIcons = {
  home: HomeOutline,
  all: StarOutline,
  kline: Pulse,
  telegraph: NewspaperOutline,
  'global-index': StatsChartOutline,
  'major-index': BarChartSharp,
  'industry-rank': Flag,
  'stock-flow': AnalyticsOutline,
  'sector-flow': DiamondOutline,
  'concept-flow': ReorderTwoOutline,
  'dragon-tiger': Flame,
  'stock-research': EaselSharp,
  'company-notice': InformationOutline,
  'industry-research': FlaskOutline,
  'hot-now': WarningOutline,
  'featured-sites': SearchOutline,
  pattern: SearchOutline,
  'duo-kong': Pulse,
  strategy: AnalyticsOutline,
  radar: TimeOutline,
  'portfolio-overview': Wallet,
  'stock-holdings': AnalyticsOutline,
  'fund-holdings': StatsChartOutline,
  'bond-holdings': InformationOutline,
  'metal-holdings': DiamondOutline,
  'forex-holdings': ReorderTwoOutline,
  'future-holdings': Pulse,
  'crypto-holdings': SparklesOutline,
  'bank-holdings': Wallet,
  analysis: SparklesOutline,
  recommendations: StarOutline,
  configuration: SettingsOutline,
  general: SettingsOutline,
}

function isActive(item) {
  if (route.name !== item.routeName) return false
  if (!item.queryKey) return true
  if (sectionKey.value === 'holding' && item.queryKey === 'assetType') {
    return (route.query.assetType || 'overview') === (item.queryValue || item.key)
  }
  if (sectionKey.value === 'market' && item.queryKey === 'tab') {
    return item.label === normaliseMarketTab(route.query)
  }
  if (sectionKey.value === 'screening' && item.queryKey === 'mode') {
    return item.key === normaliseScreeningTab(route.query)
  }
  return route.query[item.queryKey] === (item.queryValue || item.key)
}

function navigate(item) {
  router.push(buildSecondaryLocation(sectionKey.value, item.key, route.query))
  emit('navigate')
}
</script>

<template>
  <aside class="workspace-navigation" :class="[`workspace-navigation--${presentation}`, { 'secondary-open': secondaryOpen }]">
    <div class="workspace-navigation__heading">
      <div>
        <small>WORKSPACE</small>
        <p>{{ headings[sectionKey] }}</p>
      </div>
      <button class="workspace-navigation__close" type="button" aria-label="关闭当前工作区导航" @click="emit('close')">
        <n-icon :component="CloseOutline" />
      </button>
    </div>
    <nav class="workspace-navigation__list" :aria-label="`${headings[sectionKey]}二级导航`">
      <button
        v-for="item in items"
        :key="item.key"
        type="button"
        class="workspace-navigation__item"
        :class="{ 'is-active': isActive(item) }"
        :aria-current="isActive(item) ? 'page' : undefined"
        @click="navigate(item)"
      >
        <n-icon :component="itemIcons[item.key] || InformationOutline" aria-hidden="true" />
        <span>{{ item.label }}</span>
        <i v-if="isActive(item)" aria-hidden="true" />
      </button>
    </nav>
    <div class="workspace-navigation__footer">
      <span aria-hidden="true"><n-icon :component="DiamondOutline" /></span>
      <p><strong>本地数据舱</strong><small>核心数据优先保存在本机</small></p>
    </div>
  </aside>
</template>

<style scoped>
.workspace-navigation {
  width: var(--workspace-secondary-wide);
  min-width: var(--workspace-secondary-wide);
  height: 100%;
  padding: 16px 10px 12px;
  display: flex;
  flex-direction: column;
  gap: 14px;
  overflow-y: auto;
  color: var(--workspace-text);
  background: var(--workspace-surface);
  border-right: 1px solid var(--workspace-border);
  backdrop-filter: var(--workspace-glass);
}
.workspace-navigation--standard { width: var(--workspace-secondary-standard); min-width: var(--workspace-secondary-standard); }
.workspace-navigation__heading { min-height: 42px; display: flex; align-items: center; justify-content: space-between; padding: 0 9px; }
.workspace-navigation__heading div { min-width: 0; }
.workspace-navigation__heading small { display: block; color: var(--workspace-accent); font-size: 9px; font-weight: 760; letter-spacing: 0.12em; }
.workspace-navigation__heading p { margin: 2px 0 0; color: var(--workspace-text); font-size: 14px; font-weight: 700; }
.workspace-navigation__close { display: none; width: 34px; height: 34px; place-items: center; padding: 0; border: 1px solid var(--workspace-border); border-radius: var(--workspace-radius-md); color: var(--workspace-text-muted); background: var(--workspace-surface-solid); cursor: pointer; }
.workspace-navigation__list { display: grid; gap: 4px; }
.workspace-navigation__item {
  position: relative;
  min-height: 40px;
  display: grid;
  grid-template-columns: 20px minmax(0, 1fr) 6px;
  align-items: center;
  gap: 9px;
  padding: 8px 10px;
  border: 1px solid transparent;
  border-radius: var(--workspace-radius-md);
  color: var(--workspace-text-muted);
  background: transparent;
  text-align: left;
  font-size: 13px;
  cursor: pointer;
  transition: color 160ms ease, border-color 160ms ease, background-color 160ms ease, transform 160ms ease;
}
.workspace-navigation__item .n-icon { font-size: 17px; }
.workspace-navigation__item:hover { color: var(--workspace-text); border-color: var(--workspace-border); background: var(--workspace-surface-muted); }
.workspace-navigation__item:active { transform: translateY(1px); }
.workspace-navigation__item.is-active { color: var(--workspace-accent); border-color: color-mix(in srgb, var(--workspace-accent) 22%, transparent); background: var(--workspace-accent-soft); font-weight: 680; }
.workspace-navigation__item i { width: 6px; height: 6px; border-radius: 50%; background: var(--workspace-accent); box-shadow: 0 0 0 4px var(--workspace-accent-muted); }
.workspace-navigation__footer { margin-top: auto; display: flex; align-items: center; gap: 9px; padding: 12px 9px 0; border-top: 1px solid var(--workspace-border); color: var(--workspace-text-muted); }
.workspace-navigation__footer > span { width: 30px; height: 30px; display: grid; place-items: center; border-radius: var(--workspace-radius-md); color: var(--workspace-accent); background: var(--workspace-accent-soft); }
.workspace-navigation__footer p { margin: 0; display: grid; }
.workspace-navigation__footer strong { color: var(--workspace-text-secondary); font-size: 11px; font-weight: 680; }
.workspace-navigation__footer small { margin-top: 2px; color: var(--workspace-text-faint); font-size: 9px; }

@media (max-width: 1080px) {
  .workspace-navigation { position: fixed; inset: var(--workspace-topbar-height) auto 0 0; z-index: var(--workspace-layer-drawer); width: 296px; min-width: 296px; transform: translateX(-101%); box-shadow: var(--workspace-shadow); transition: transform 180ms ease; }
  .workspace-navigation.secondary-open { transform: translateX(0); }
  .workspace-navigation__close { display: grid; }
}
@media (max-width: 719px) { .workspace-navigation { inset: calc(var(--workspace-topbar-height) + 46px) auto 0 0; } }
</style>
