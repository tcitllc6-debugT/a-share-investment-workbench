<script setup>
import { computed, nextTick, onMounted, ref } from 'vue'
import { MdPreview } from 'md-editor-v3'
import 'md-editor-v3/lib/preview.css'
import { CheckUpdate, GetConfig, GetUserManual, GetVersionInfo } from '../../wailsjs/go/main/App'

const productName = 'A股智投工作台'
const versionInfo = ref('')
const updateLog = ref('')
const icon = ref('')
const darkTheme = ref(false)
const showManual = ref(false)
const manualContent = ref('')
const manualId = 'manual-preview'
const theme = computed(() => darkTheme.value ? 'dark' : 'light')

function openManual() {
  if (!manualContent.value) {
    GetUserManual().then(res => {
      manualContent.value = res
      showManual.value = true
      nextTick()
    })
    return
  }
  showManual.value = true
}

onMounted(() => {
  document.title = `关于 ${productName}`
  GetConfig().then(res => {
    darkTheme.value = !!res.darkTheme
  })
  GetVersionInfo().then(res => {
    versionInfo.value = res.version
    updateLog.value = res.content
    icon.value = res.icon
  })
})
</script>

<template>
  <n-space vertical size="large" style="--wails-draggable:no-drag">
    <n-card size="large">
      <n-space vertical align="center" :size="16">
        <n-image v-if="icon" width="96" :src="icon" />
        <n-gradient-text type="info" :size="46">{{ productName }}</n-gradient-text>
        <n-tag type="success" :bordered="false">v{{ versionInfo || 'local' }}</n-tag>
        <n-text depth="3" style="max-width: 760px; text-align: center; line-height: 1.8">
          本项目是基于 go-stock 的本地优先复刻版，面向 A 股行情查看、AI 辅助分析和桌面工作流。
          默认不依赖原作者的社区、赞助、设备绑定服务；更新由本项目自己的签名发布通道管理。
        </n-text>
        <n-flex justify="center">
          <n-button size="small" type="info" tertiary @click="CheckUpdate(1)">检查更新</n-button>
          <n-button size="small" type="success" tertiary @click="openManual">查看用户手册</n-button>
        </n-flex>
      </n-space>
    </n-card>

    <n-card title="本地优先说明" size="small">
      <n-space vertical style="text-align: left; line-height: 1.9">
        <n-text>远程提示词广场默认为空；如需远程服务，请在设置中填写你自己的 API 地址。</n-text>
        <n-text>分享功能默认导出到本地 Markdown 文件，不上传到原作者服务。</n-text>
        <n-text>自动更新只接受本项目维护者签名的发布清单和校验通过的更新包。</n-text>
        <n-text>所有行情、新闻、AI 模型等外部能力应通过可替换 provider 或用户配置接入。</n-text>
      </n-space>
    </n-card>

    <n-card title="开源说明" size="small">
      <n-space vertical style="text-align: left; line-height: 1.9">
        <n-text>本项目包含基于 ArvinLovegood/go-stock 修改的代码。</n-text>
        <n-text>上游项目：https://github.com/ArvinLovegood/go-stock</n-text>
        <n-text>本项目按 GNU GPL v3.0 发布。</n-text>
        <n-text>修改内容与对应源码：https://github.com/tcitllc6-debugT/a-share-investment-workbench</n-text>
        <n-text>本项目与上游作者无隶属、授权或官方合作关系。</n-text>
        <n-text>技术栈：Wails、Vue、Naive UI、Go，以及用户自行配置的 AI 模型服务。</n-text>
      </n-space>
    </n-card>

    <n-card title="风险提示" size="small">
      <n-alert type="warning" :bordered="false">
        本软件仅用于学习、研究和个人工具流。行情数据和 AI 分析结果仅供参考，不构成任何投资建议或交易决策依据。
        投资有风险，入市需谨慎。
      </n-alert>
      <p v-if="updateLog" style="text-align: left; margin-top: 16px">构建说明：{{ updateLog }}</p>
    </n-card>

    <n-modal v-model:show="showManual" preset="card" title="用户手册" style="width: min(980px, 92vw)">
      <MdPreview
        style="text-align: left;"
        :id="manualId"
        v-model="manualContent"
        :theme="theme"
        :preview-theme="'github'"
      />
    </n-modal>
  </n-space>
</template>
