<script setup>
import { markRaw } from 'vue'
import {
  BarChartOutline,
  CalendarOutline,
  GlobeOutline,
  NewspaperOutline,
  OpenOutline,
  PulseOutline,
  TrendingUpOutline,
} from '@vicons/ionicons5'
import { OpenURL } from '../../wailsjs/go/main/App'

const sites = [
  {
    name: '选股通',
    description: 'A 股题材、事件与产业链线索',
    url: 'https://xuangutong.com.cn',
    icon: markRaw(TrendingUpOutline),
  },
  {
    name: '东方财富大盘星图',
    description: '全市场涨跌分布与板块热度',
    url: 'https://quote.eastmoney.com/stockhotmap/',
    icon: markRaw(BarChartOutline),
  },
  {
    name: '财联社行情',
    description: '指数、板块和个股行情数据',
    url: 'https://www.cls.cn/quotation',
    icon: markRaw(PulseOutline),
  },
  {
    name: '财经热榜',
    description: '聚合当日财经热点与讨论趋势',
    url: 'https://tophub.today/c/finance',
    icon: markRaw(NewspaperOutline),
  },
  {
    name: '金十期货',
    description: '国内外期货品种行情与资讯',
    url: 'https://qihuo.jin10.com/',
    icon: markRaw(GlobeOutline),
  },
  {
    name: '金十数据日历',
    description: '全球宏观事件与重要经济数据',
    url: 'https://datas.jin10.com/#/category/52005',
    icon: markRaw(CalendarOutline),
  },
]

function openSite(site) {
  OpenURL(site.url)
}
</script>

<template>
  <section class="curated-sites">
    <header class="curated-sites__header">
      <div>
        <h3>财经站点</h3>
        <p>行情、资讯与宏观数据入口</p>
      </div>
    </header>

    <n-grid cols="1 720:2 1180:3" :x-gap="12" :y-gap="12" responsive="self">
      <n-grid-item v-for="site in sites" :key="site.name">
        <n-card class="site-card" size="small" hoverable>
          <div class="site-card__body">
            <span class="site-card__icon">
              <n-icon :component="site.icon" :size="22" />
            </span>
            <div class="site-card__copy">
              <strong>{{ site.name }}</strong>
              <span>{{ site.description }}</span>
            </div>
            <n-button quaternary type="primary" size="small" @click="openSite(site)">
              <template #icon><n-icon :component="OpenOutline" /></template>
              打开
            </n-button>
          </div>
        </n-card>
      </n-grid-item>
    </n-grid>
  </section>
</template>

<style scoped>
.curated-sites {
  padding: 14px;
}

.curated-sites__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
}

.curated-sites__header h3 {
  margin: 0;
  font-size: 17px;
  font-weight: 700;
  letter-spacing: 0;
}

.curated-sites__header p {
  margin: 4px 0 0;
  color: var(--n-text-color-3);
  font-size: 12px;
}

.site-card {
  height: 100%;
  border-radius: 8px;
}

.site-card__body {
  display: grid;
  grid-template-columns: 38px minmax(0, 1fr) auto;
  gap: 10px;
  align-items: center;
  min-height: 48px;
}

.site-card__icon {
  display: grid;
  width: 36px;
  height: 36px;
  place-items: center;
  border: 1px solid rgba(49, 119, 246, 0.2);
  border-radius: 8px;
  color: #3177f6;
  background: rgba(49, 119, 246, 0.08);
}

.site-card__copy {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 3px;
}

.site-card__copy strong {
  overflow: hidden;
  font-size: 14px;
  font-weight: 650;
  letter-spacing: 0;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.site-card__copy span {
  overflow: hidden;
  color: var(--n-text-color-3);
  font-size: 12px;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
