<script setup>
import { computed, ref } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import {
  GridOutline,
  HelpCircleOutline,
  MenuOutline,
  MoonOutline,
  NewspaperOutline,
  OptionsOutline,
  SearchOutline,
  SettingsOutline,
  SparklesOutline,
  StarOutline,
  SunnyOutline,
  WalletOutline,
} from '@vicons/ionicons5'
import { NIcon } from 'naive-ui'
import { PRIMARY_SECTIONS, getSectionForRoute } from '../../navigation/workspace.mjs'
import brandIconUrl from '../../assets/app-icon.svg'

const props = defineProps({
  marketStatus: { type: String, default: '市场状态同步中' },
  themeMode: { type: String, default: 'light' },
})
const emit = defineEmits(['search', 'toggle-secondary', 'theme-change'])
const route = useRoute()
const query = ref('')
const activeSection = computed(() => getSectionForRoute(route))
const nextTheme = computed(() => props.themeMode === 'dark' ? 'light' : 'dark')
const themeIcon = computed(() => props.themeMode === 'dark' ? SunnyOutline : MoonOutline)
const themeTitle = computed(() => props.themeMode === 'dark' ? '切换到日间主题' : '切换到暗黑主题')

const sectionIcons = {
  overview: GridOutline,
  watchlist: StarOutline,
  market: NewspaperOutline,
  screening: OptionsOutline,
  holding: WalletOutline,
  ai: SparklesOutline,
  settings: SettingsOutline,
  about: HelpCircleOutline,
}

function submitSearch() {
  const symbol = query.value.trim()
  if (symbol) emit('search', symbol)
}
</script>

<template>
  <header class="workspace-top-navigation">
    <RouterLink :to="{ name: 'dashboard' }" class="workspace-top-navigation__brand" aria-label="A股智投工作台总览">
      <span class="workspace-top-navigation__mark" aria-hidden="true">
        <img :src="brandIconUrl" alt="" />
      </span>
      <span class="workspace-top-navigation__brand-copy">
        <strong>A股智投</strong>
        <small>PRISM ATLAS</small>
      </span>
    </RouterLink>

    <nav class="workspace-top-navigation__sections" aria-label="主要工作区">
      <RouterLink
        v-for="section in PRIMARY_SECTIONS"
        :key="section.key"
        :to="{ name: section.routeName }"
        class="workspace-top-navigation__section"
        :class="{ 'is-active': activeSection === section.key }"
        :aria-current="activeSection === section.key ? 'page' : undefined"
      >
        <n-icon :component="sectionIcons[section.key]" aria-hidden="true" />
        <span>{{ section.label }}</span>
      </RouterLink>
    </nav>

    <div class="workspace-top-navigation__tools">
      <form class="workspace-top-navigation__search" @submit.prevent="submitSearch">
        <n-icon :component="SearchOutline" aria-hidden="true" />
        <input v-model="query" aria-label="搜索股票名称或代码" name="stock-search" autocomplete="off" placeholder="搜索股票…" />
        <kbd>Enter</kbd>
      </form>
      <span class="workspace-top-navigation__status" :title="props.marketStatus">
        <i aria-hidden="true" />{{ props.marketStatus }}
      </span>
      <button
        class="workspace-top-navigation__icon-button"
        type="button"
        aria-label="切换日间或暗黑主题"
        :title="themeTitle"
        @click="emit('theme-change', nextTheme)"
      >
        <n-icon :component="themeIcon" />
      </button>
      <button class="workspace-top-navigation__drawer-button" type="button" aria-label="打开当前工作区导航" @click="emit('toggle-secondary')">
        <n-icon :component="MenuOutline" />
      </button>
    </div>
  </header>
</template>

<style scoped>
.workspace-top-navigation {
  position: relative;
  z-index: var(--workspace-layer-nav);
  height: var(--workspace-topbar-height);
  display: flex;
  align-items: stretch;
  gap: 18px;
  padding: 0 18px;
  color: var(--workspace-text);
  background: var(--workspace-surface);
  border-bottom: 1px solid var(--workspace-border);
  box-shadow: 0 1px 0 rgba(255, 255, 255, 0.28);
  backdrop-filter: var(--workspace-glass);
}

.workspace-top-navigation__brand {
  min-width: max-content;
  display: flex;
  align-items: center;
  gap: 10px;
  color: var(--workspace-text);
  text-decoration: none;
}

.workspace-top-navigation__mark {
  width: 34px;
  height: 34px;
  display: grid;
  place-items: center;
  overflow: hidden;
  border-radius: 10px;
  box-shadow: 0 7px 18px rgba(52, 120, 246, 0.2);
}

.workspace-top-navigation__mark img {
  width: 100%;
  height: 100%;
  display: block;
}

.workspace-top-navigation__brand-copy { display: grid; line-height: 1.05; }
.workspace-top-navigation__brand-copy strong { font-size: 14px; font-weight: 720; }
.workspace-top-navigation__brand-copy small { margin-top: 4px; color: var(--workspace-text-faint); font-size: 8px; font-weight: 700; letter-spacing: 0.12em; }

.workspace-top-navigation__sections {
  min-width: 0;
  display: flex;
  align-items: stretch;
  gap: 2px;
  overflow-x: auto;
  scrollbar-width: none;
}
.workspace-top-navigation__sections::-webkit-scrollbar { display: none; }

.workspace-top-navigation__section {
  position: relative;
  min-width: max-content;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 0 10px;
  color: var(--workspace-text-muted);
  font-size: 13px;
  font-weight: 520;
  text-decoration: none;
  transition: color 160ms ease, background-color 160ms ease;
}
.workspace-top-navigation__section .n-icon { font-size: 16px; }
.workspace-top-navigation__section::after {
  content: '';
  position: absolute;
  right: 10px;
  bottom: 0;
  left: 10px;
  height: 3px;
  border-radius: 3px 3px 0 0;
  background: var(--workspace-accent);
  transform: scaleX(0);
  transition: transform 160ms ease;
}
.workspace-top-navigation__section:hover { color: var(--workspace-text); background: var(--workspace-accent-muted); }
.workspace-top-navigation__section.is-active { color: var(--workspace-accent); font-weight: 680; }
.workspace-top-navigation__section.is-active::after { transform: scaleX(1); }

.workspace-top-navigation__tools { margin-left: auto; min-width: 0; display: flex; align-items: center; gap: 8px; }
.workspace-top-navigation__search {
  width: clamp(160px, 16vw, 240px);
  height: 36px;
  display: flex;
  align-items: center;
  gap: 7px;
  padding: 0 9px;
  border: 1px solid var(--workspace-border);
  border-radius: var(--workspace-radius-md);
  color: var(--workspace-text-muted);
  background: var(--workspace-surface-muted);
  transition: border-color 160ms ease, box-shadow 160ms ease, background-color 160ms ease;
}
.workspace-top-navigation__search:focus-within { border-color: var(--workspace-accent); background: var(--workspace-surface-solid); box-shadow: var(--workspace-focus-ring); }
.workspace-top-navigation__search input { min-width: 0; flex: 1; padding: 0; border: 0; outline: 0; color: var(--workspace-text); background: transparent; font-size: 12px; }
.workspace-top-navigation__search kbd { padding: 1px 5px; border: 1px solid var(--workspace-border); border-radius: 4px; color: var(--workspace-text-faint); background: var(--workspace-surface-solid); font-family: var(--workspace-font); font-size: 9px; }

.workspace-top-navigation__status { max-width: 190px; display: flex; align-items: center; gap: 6px; overflow: hidden; color: var(--workspace-text-muted); font-size: 11px; white-space: nowrap; text-overflow: ellipsis; }
.workspace-top-navigation__status i { width: 7px; height: 7px; flex: 0 0 auto; border-radius: 50%; background: var(--workspace-negative); box-shadow: 0 0 0 4px color-mix(in srgb, var(--workspace-negative) 14%, transparent); }

.workspace-top-navigation__icon-button,
.workspace-top-navigation__drawer-button {
  width: 36px;
  height: 36px;
  place-items: center;
  padding: 0;
  border: 1px solid var(--workspace-border);
  border-radius: var(--workspace-radius-md);
  color: var(--workspace-text-secondary);
  background: var(--workspace-surface-solid);
  cursor: pointer;
  transition: color 160ms ease, border-color 160ms ease, background-color 160ms ease, transform 160ms ease;
}
.workspace-top-navigation__icon-button { display: grid; }
.workspace-top-navigation__drawer-button { display: none; }
.workspace-top-navigation__icon-button:hover,
.workspace-top-navigation__drawer-button:hover { color: var(--workspace-accent); border-color: var(--workspace-accent); background: var(--workspace-accent-soft); }
.workspace-top-navigation__icon-button:active,
.workspace-top-navigation__drawer-button:active { transform: translateY(1px); }

@media (max-width: 1360px) {
  .workspace-top-navigation { gap: 9px; padding: 0 12px; }
  .workspace-top-navigation__section { padding: 0 7px; }
  .workspace-top-navigation__status { display: none; }
}
@media (max-width: 1080px) {
  .workspace-top-navigation__drawer-button { display: grid; }
  .workspace-top-navigation__brand-copy { display: none; }
  .workspace-top-navigation__sections { flex: 1; }
  .workspace-top-navigation__search { width: 132px; }
  .workspace-top-navigation__section span { display: none; }
  .workspace-top-navigation__section { min-width: 42px; justify-content: center; padding: 0 8px; }
}
@media (max-width: 719px) {
  .workspace-top-navigation { gap: 7px; padding: 0 9px; }
  .workspace-top-navigation__sections { order: 3; position: absolute; top: var(--workspace-topbar-height); left: 0; width: 100%; height: 46px; padding: 0 10px; background: var(--workspace-surface-elevated); border-bottom: 1px solid var(--workspace-border); backdrop-filter: var(--workspace-glass); }
  .workspace-top-navigation__section { min-width: 46px; }
  .workspace-top-navigation__search { width: 112px; }
  .workspace-top-navigation__search kbd { display: none; }
}
</style>
