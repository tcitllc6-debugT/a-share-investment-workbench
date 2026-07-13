<script setup>
import { computed, onMounted, ref } from 'vue'
import { ApplyManagedUpdate, CheckManagedUpdate, GetVersionInfo, OpenURL } from '../../wailsjs/go/main/App'

const repositoryUrl = 'https://github.com/tcitllc6-debugT/a-share-investment-workbench'
const upstreamProjectUrl = 'https://github.com/ArvinLovegood/go-stock'
const version = ref('--')
const checking = ref(false)
const update = ref(null)
const status = ref('')

const updateSummary = computed(() => {
  if (checking.value) return '正在检查已签名的更新包...'
  if (update.value?.available) return `发现 v${update.value.version}，已可在顶部提醒中一键更新。`
  return status.value || '当前版本已连接到安全更新通道。'
})

async function checkUpdate() {
  checking.value = true
  status.value = ''
  try {
    const result = await CheckManagedUpdate()
    update.value = result || null
    status.value = result?.error
      ? '暂时无法连接更新源，稍后会自动重试。'
      : result?.available
        ? `发现新版本 v${result.version}，启动后会显示更新提醒。`
        : '已是最新版本。'
  } catch (_) {
    status.value = '检查更新失败，稍后会自动重试。'
  } finally {
    checking.value = false
  }
}

async function applyUpdate() {
  checking.value = true
  try {
    const result = await ApplyManagedUpdate()
    status.value = result?.started
      ? '更新包已验证，应用将自动重启并完成替换。'
      : result?.error || result?.message || '暂时无法启动更新，请稍后重试。'
  } catch (error) {
    status.value = error?.message || '暂时无法启动更新，请稍后重试。'
  } finally {
    checking.value = false
  }
}

onMounted(async () => {
  try {
    const info = await GetVersionInfo()
    version.value = info?.version || '--'
  } catch (_) {
    version.value = '--'
  }
})
</script>

<template>
  <section class="about-workspace">
    <div class="about-workspace__hero">
      <div class="about-workspace__mark">A</div>
      <div>
        <span class="about-workspace__eyebrow">A 股智投工作台</span>
        <h1>把研究、筛选和持仓放回同一个桌面。</h1>
        <p>这是一个基于多个开源项目进行二次开发、统一交互与本地化优化的 A 股桌面研究工具。</p>
      </div>
      <div class="about-workspace__version"><span>当前版本</span><strong>v{{ version }}</strong></div>
    </div>

    <div class="about-workspace__grid">
      <article class="about-workspace__card">
        <span class="about-workspace__label">项目地址</span>
        <h2>持续迭代的本地优先投资工作台</h2>
        <p>自选股、K 线分析、市场资讯、多维选股、盘中异动、统一持仓和 AI 投研都围绕日常 A 股研究流程组织；核心数据默认保存在你的电脑上。</p>
        <div class="about-workspace__actions">
          <button type="button" class="about-workspace__primary" @click="OpenURL(repositoryUrl)">打开 GitHub 仓库</button>
          <button type="button" class="about-workspace__secondary" :disabled="checking" @click="checkUpdate">{{ checking ? '检查中...' : '检查更新' }}</button>
        </div>
      </article>

      <article class="about-workspace__card">
        <span class="about-workspace__label">自动更新</span>
        <h2>GitHub 发布，应用内提醒</h2>
        <p>每次合并到发布分支后，发布工作流会构建并签名更新包。应用启动完成后静默检查；有新版时会出现一键更新提醒。</p>
        <p class="about-workspace__status">{{ updateSummary }}</p>
        <button v-if="update?.available" type="button" class="about-workspace__update" :disabled="checking" @click="applyUpdate">立即更新</button>
      </article>
    </div>

    <section class="about-workspace__features">
      <div class="about-workspace__section-heading"><span class="about-workspace__label">主要能力</span><h2>专为 A 股研究的日常工作流</h2></div>
      <div class="about-workspace__feature-grid">
        <article><span>01</span><h3>看盘与自选</h3><p>自选分组、多周期 K 线、资金、公告、研报与 AI 分析入口。</p></article>
        <article><span>02</span><h3>市场与雷达</h3><p>快讯、指数、行业资金、热点和盘中异动统一观察。</p></article>
        <article><span>03</span><h3>多维选股</h3><p>形态选股、多空线、策略选股和多条件过滤。</p></article>
        <article><span>04</span><h3>统一持仓</h3><p>股票、基金、债券、贵金属、外汇、期货、加密资产与理财折算为人民币。</p></article>
        <article><span>05</span><h3>AI 控制台</h3><p>研究记录、推荐记录、提示词模板与模型配置可追溯管理。</p></article>
        <article><span>06</span><h3>本地优先</h3><p>持仓和研究账本优先保存在本机，更新只替换程序文件。</p></article>
      </div>
    </section>

    <section class="about-workspace__notice">
      <span class="about-workspace__label">许可与归属</span>
      <h2>开源说明</h2>
      <p>本项目包含基于 <button type="button" class="about-workspace__text-link" @click="OpenURL(upstreamProjectUrl)">ArvinLovegood/go-stock</button> 修改的代码。</p>
      <p>本项目按 GNU GPL v3.0 发布。修改内容与对应源码：<button type="button" class="about-workspace__text-link" @click="OpenURL(repositoryUrl)">A股智投工作台 GitHub 仓库</button>。</p>
      <p>本项目与上游作者无隶属、授权或官方合作关系。</p>
    </section>

    <p class="about-workspace__footnote">更新包会先校验签名和文件哈希，验证通过后才会替换程序；本地持仓、交易日志和设置不会被覆盖。</p>
  </section>
</template>

<style scoped>
.about-workspace { min-height:100%; padding:28px; color:var(--workspace-text); }
.about-workspace__hero { display:grid; grid-template-columns:auto minmax(0, 1fr) auto; align-items:center; gap:18px; padding:28px; border:1px solid var(--workspace-border); border-radius:var(--workspace-radius-md); background:linear-gradient(135deg, color-mix(in srgb, var(--workspace-accent-soft) 68%, var(--workspace-surface-solid)), var(--workspace-surface-solid)); box-shadow:var(--workspace-shadow-sm); }
.about-workspace__mark { width:56px; height:56px; display:grid; place-items:center; border-radius:16px; color:#fff; background:linear-gradient(145deg, #75a7ff, var(--workspace-accent)); box-shadow:0 14px 30px color-mix(in srgb, var(--workspace-accent) 24%, transparent); font-size:25px; font-weight:760; }
.about-workspace__eyebrow, .about-workspace__label { display:block; color:var(--workspace-accent); font-size:11px; font-weight:760; letter-spacing:.08em; }
.about-workspace__hero h1 { margin:5px 0 0; font-size:25px; line-height:1.22; letter-spacing:0; }.about-workspace__hero p { max-width:760px; margin:8px 0 0; color:var(--workspace-text-secondary); font-size:13px; line-height:1.65; }
.about-workspace__version { min-width:112px; padding-left:22px; border-left:1px solid var(--workspace-border); text-align:right; }.about-workspace__version span { display:block; color:var(--workspace-text-faint); font-size:11px; }.about-workspace__version strong { display:block; margin-top:5px; font-family:var(--workspace-number-font); font-size:16px; }
.about-workspace__grid { display:grid; grid-template-columns:1.35fr 1fr; gap:14px; margin-top:14px; }.about-workspace__card { min-height:204px; padding:22px; border:1px solid var(--workspace-border); border-radius:var(--workspace-radius-md); background:var(--workspace-surface-solid); box-shadow:var(--workspace-shadow-sm); }.about-workspace__card h2,.about-workspace__section-heading h2 { margin:7px 0 0; font-size:18px; letter-spacing:0; }.about-workspace__card p { margin:10px 0 0; color:var(--workspace-text-secondary); font-size:13px; line-height:1.7; }.about-workspace__status { color:var(--workspace-text-muted) !important; }
.about-workspace__actions { display:flex; flex-wrap:wrap; gap:9px; margin-top:18px; }.about-workspace__actions button { min-height:34px; padding:0 12px; border-radius:var(--workspace-radius-sm); font-size:12px; cursor:pointer; }.about-workspace__primary { border:1px solid var(--workspace-accent); color:#fff; background:var(--workspace-accent); }.about-workspace__secondary { border:1px solid var(--workspace-border-strong); color:var(--workspace-accent); background:var(--workspace-surface-solid); }.about-workspace__actions button:disabled { cursor:wait; opacity:.65; }
.about-workspace__update { min-height:34px; margin-top:14px; padding:0 12px; border:1px solid var(--workspace-accent); border-radius:var(--workspace-radius-sm); color:#fff; background:var(--workspace-accent); font-size:12px; cursor:pointer; }.about-workspace__update:disabled { cursor:wait; opacity:.65; }
.about-workspace__features { margin-top:14px; padding:22px; border:1px solid var(--workspace-border); border-radius:var(--workspace-radius-md); background:var(--workspace-surface-solid); box-shadow:var(--workspace-shadow-sm); }.about-workspace__feature-grid { display:grid; grid-template-columns:repeat(3, minmax(0, 1fr)); gap:10px; margin-top:18px; }.about-workspace__feature-grid article { min-height:140px; padding:16px; border:1px solid var(--workspace-border); border-radius:var(--workspace-radius-sm); background:var(--workspace-surface-muted); }.about-workspace__feature-grid span { color:var(--workspace-accent); font-family:var(--workspace-number-font); font-size:11px; font-weight:700; }.about-workspace__feature-grid h3 { margin:10px 0 0; font-size:14px; }.about-workspace__feature-grid p { margin:7px 0 0; color:var(--workspace-text-secondary); font-size:12px; line-height:1.65; }.about-workspace__footnote { margin:14px 2px 0; color:var(--workspace-text-faint); font-size:11px; line-height:1.65; }
.about-workspace__notice { margin-top:14px; padding:18px 22px; border:1px solid var(--workspace-border); border-radius:var(--workspace-radius-md); background:var(--workspace-surface-muted); }.about-workspace__notice h2 { margin:6px 0 0; font-size:16px; letter-spacing:0; }.about-workspace__notice p { margin:7px 0 0; color:var(--workspace-text-secondary); font-size:12px; line-height:1.65; }.about-workspace__text-link { padding:0; border:0; color:var(--workspace-accent); background:transparent; font:inherit; cursor:pointer; text-decoration:underline; text-underline-offset:2px; }
@media (max-width:1000px) { .about-workspace__grid { grid-template-columns:1fr; }.about-workspace__feature-grid { grid-template-columns:repeat(2, minmax(0, 1fr)); } }
@media (max-width:680px) { .about-workspace { padding:14px; }.about-workspace__hero { grid-template-columns:auto minmax(0, 1fr); padding:20px; }.about-workspace__version { grid-column:1 / -1; padding:12px 0 0; border-top:1px solid var(--workspace-border); border-left:0; text-align:left; }.about-workspace__feature-grid { grid-template-columns:1fr; }.about-workspace__hero h1 { font-size:21px; } }
</style>
