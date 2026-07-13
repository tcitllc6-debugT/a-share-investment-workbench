<script setup>
import {h, onBeforeUnmount, onMounted, ref} from "vue";
import {
  AddPrompt,
  DelPrompt,
  ExportConfig,
  GetConfig,
  GetPromptTemplates,
  SendDingDingMessageByType,
  SendFeishuMessageByType,
  StartFeishuBot,
  StopFeishuBot,
  GetFeishuBotStatus,
  UpdateGeneralConfig,
  GetAiConfigs,
  CheckManagedUpdate,
  ApplyManagedUpdate
} from "../../wailsjs/go/main/App";
import {NTag, NTooltip, NIcon, useMessage} from "naive-ui";
import {data, models} from "../../wailsjs/go/models";
import {EventsEmit} from "../../wailsjs/runtime";
import {HelpCircleFilledIcon, HelpIcon} from "tdesign-icons-vue-next";
import {applyThemeMode, persistThemeMode} from '../utils/theme-mode.mjs'

const message = useMessage()
const checkingUpdate = ref(false)
const applyingUpdate = ref(false)
const updateResult = ref(null)

const formRef = ref(null)
const formValue = ref({
  ID: 1,
  tushareToken: '',
  iwencaiApiKey: '',
  emApiKey: '',
  dingPush: {
    enable: false,
    dingRobot: ''
  },
  feishuPush: {
    enable: false,
    feishuRobot: '',
    feishuSecret: ''
  },
  feishuBot: {
    enable: false,
    appId: '',
    appSecret: '',
    aiConfigId: 0,
    sysPromptId: 0,
    enableTools: true,
    thinking: false,
    agentMode: 'react',
    status: 'stopped'
  },
  localPush: {
    enable: true,
  },
  updateBasicInfoOnStart: false,
  refreshInterval: 1,
  enableDanmu: false,
  browserPath: '',
  enableNews: false,
  darkTheme: true,
  enableFund: false,
  enablePushNews: true,
  enableOnlyPushRedNews: true,
  httpProxy:"",
  httpProxyEnabled:false,
  enableAgent: false,
  qgqpBId: '',
  updateChannel: 'release',
})

const updateChannelOptions = [
  { label: 'Release（稳定版）', value: 'release' },
  { label: 'Pre-release（预发布版）', value: 'pre' },
  { label: 'Dev（开发版）', value: 'dev' },
]

const promptTemplates = ref([])
const aiConfigChoices = ref([])

onMounted(() => {
  GetConfig().then(res => {
    formValue.value.ID = res.ID
    formValue.value.tushareToken = res.tushareToken
    formValue.value.iwencaiApiKey = res.iwencaiApiKey || ''
    formValue.value.emApiKey = res.emApiKey || ''
    formValue.value.dingPush = {
      enable: res.dingPushEnable,
      dingRobot: res.dingRobot
    }
    formValue.value.feishuPush = {
      enable: res.feishuPushEnable,
      feishuRobot: res.feishuRobot,
      feishuSecret: res.feishuSecret
    }
    formValue.value.feishuBot = {
      enable: res.feishuBotEnable,
      appId: res.feishuAppId || '',
      appSecret: res.feishuAppSecret || '',
      aiConfigId: res.feishuBotAiConfigId || 0,
      sysPromptId: res.feishuBotSysPromptId || 0,
      enableTools: res.feishuBotEnableTools !== false,
      thinking: res.feishuBotThinking === true,
      agentMode: res.feishuBotAgentMode || 'react',
      status: 'stopped'
    }
    formValue.value.localPush = {
      enable: res.localPushEnable,
    }
    formValue.value.updateBasicInfoOnStart = res.updateBasicInfoOnStart
    formValue.value.refreshInterval = res.refreshInterval
    formValue.value.enableDanmu = res.enableDanmu
    formValue.value.browserPath = res.browserPath
    formValue.value.enableNews = res.enableNews
    formValue.value.darkTheme = res.darkTheme
    formValue.value.enableFund = res.enableFund
    formValue.value.enablePushNews = res.enablePushNews
    formValue.value.enableOnlyPushRedNews = res.enableOnlyPushRedNews
    formValue.value.httpProxy=res.httpProxy;
    formValue.value.httpProxyEnabled=res.httpProxyEnabled;
    formValue.value.enableAgent = res.enableAgent;
    formValue.value.qgqpBId = res.qgqpBId;
    formValue.value.updateChannel = res.updateChannel || 'release';

  })

  GetPromptTemplates("", "").then(res => {
    promptTemplates.value = res
  })
  GetAiConfigs().then((res) => {
    aiConfigChoices.value = res || []
  })
})
onBeforeUnmount(() => {
  message.destroyAll()
})

function saveConfig() {
  console.log('开始保存设置', formValue.value);
  let config = new data.SettingConfig({
    ID: formValue.value.ID,
    dingPushEnable: formValue.value.dingPush.enable,
    dingRobot: formValue.value.dingPush.dingRobot,
    feishuPushEnable: formValue.value.feishuPush.enable,
    feishuRobot: formValue.value.feishuPush.feishuRobot,
    feishuSecret: formValue.value.feishuPush.feishuSecret,
    feishuBotEnable: formValue.value.feishuBot.enable,
    feishuAppId: formValue.value.feishuBot.appId,
    feishuAppSecret: formValue.value.feishuBot.appSecret,
    feishuBotAiConfigId: formValue.value.feishuBot.aiConfigId,
    feishuBotSysPromptId: formValue.value.feishuBot.sysPromptId,
    feishuBotEnableTools: formValue.value.feishuBot.enableTools,
    feishuBotThinking: formValue.value.feishuBot.thinking,
    feishuBotAgentMode: formValue.value.feishuBot.agentMode,
    localPushEnable: formValue.value.localPush.enable,
    updateBasicInfoOnStart: formValue.value.updateBasicInfoOnStart,
    refreshInterval: formValue.value.refreshInterval,
    tushareToken: formValue.value.tushareToken,
    iwencaiApiKey: formValue.value.iwencaiApiKey,
    emApiKey: formValue.value.emApiKey,
    enableDanmu: formValue.value.enableDanmu,
    browserPath: formValue.value.browserPath,
    enableNews: formValue.value.enableNews,
    darkTheme: formValue.value.darkTheme,
    enableFund: formValue.value.enableFund,
    enablePushNews: formValue.value.enablePushNews,
    enableOnlyPushRedNews: formValue.value.enableOnlyPushRedNews,
    httpProxy:formValue.value.httpProxy,
    httpProxyEnabled:formValue.value.httpProxyEnabled,
    enableAgent: formValue.value.enableAgent,
    qgqpBId: formValue.value.qgqpBId,
    updateChannel: formValue.value.updateChannel,
  })

  return UpdateGeneralConfig(config).then(res => {
    if (res === '保存成功！') {
      message.success(res)
    } else {
      message.error(res)
    }
    EventsEmit("updateSettings", config);
  })
}


function getHeight() {
  return document.documentElement.clientHeight
}

function sendTestNotice() {
  let markdown = "### A股智投工作台测试\n" + new Date()
  let msg = '{' +
      '     "msgtype": "markdown",' +
      '     "markdown": {' +
      '         "title":"A股智投工作台 ' + new Date() + '",' +
      '         "text": "' + markdown + '"' +
      '     },' +
      '      "at": {' +
      '          "isAtAll": true' +
      '      }' +
      ' }'

  SendDingDingMessageByType(msg, "test-" + new Date().getTime(), 1).then(res => {
    message.info(res)
  })
}

function sendFeishuTestNotice() {
  let markdown = "### A股智投工作台飞书测试\n" + new Date()
  // 飞书卡片 JSON 2.0 协议：schema="2.0" + body.elements + markdown 元素
  // 文档：https://open.feishu.cn/document/feishu-cards/card-json-v2-components/content-components/rich-text
  let msg = JSON.stringify({
    msg_type: "interactive",
    card: {
      schema: "2.0",
      header: {
        title: {
          tag: "plain_text",
          content: "A股智投工作台飞书测试 " + new Date()
        }
      },
      body: {
        elements: [
          {
            tag: "markdown",
            content: '<at id=all></at>\n' + markdown
          }
        ]
      }
    }
  })

  SendFeishuMessageByType(msg, "test-feishu-" + new Date().getTime(), 1).then(res => {
    message.info(res)
  })
}

// 飞书应用机器人控制函数（与 feishuPush 自定义机器人推送独立）
async function startFeishuBot() {
  // 先保存配置，确保最新 appID/appSecret/aiConfigId 已写入
  await saveConfig()
  try {
    const res = await StartFeishuBot()
    if (res && res.includes('失败')) {
      message.error(res)
    } else {
      message.success(res)
      formValue.value.feishuBot.status = 'running'
    }
  } catch (e) {
    message.error('启动失败：' + e)
  }
}

async function stopFeishuBot() {
  try {
    const res = await StopFeishuBot()
    message.success(res)
    formValue.value.feishuBot.status = 'stopped'
  } catch (e) {
    message.error('停止失败：' + e)
  }
}

async function refreshFeishuBotStatus() {
  try {
    const res = await GetFeishuBotStatus()
    formValue.value.feishuBot.status = res
    message.info('当前状态：' + res)
  } catch (e) {
    formValue.value.feishuBot.status = 'stopped'
  }
}

// 飞书机器人只读取脱敏后的 AI 配置名称与 ID。
function aiConfigOptions() {
  return aiConfigChoices.value.map(c => ({
    label: `${c.name || '未命名'} [${c.modelName || '未指定'}]`,
    value: Number(c.ID) || 0
  }))
}

// 飞书应用机器人「系统提示词」下拉选项：从提示词模板加载，0=默认（不使用系统提示词）
function promptTemplateOptions() {
  const opts = [{ label: '默认（不使用系统提示词）', value: 0 }]
  if (!promptTemplates.value || promptTemplates.value.length === 0) {
    return opts
  }
  promptTemplates.value.forEach(t => {
    const id = Number(t.ID ?? t.id) || 0
    if (id === 0) return
    const name = t.name ?? '未命名'
    const type = t.type ? ` [${t.type}]` : ''
    opts.push({ label: `${name}${type}`, value: id })
  })
  return opts
}

// 飞书应用机器人「Agent 模式」下拉选项
function agentModeOptions() {
  return [
    { label: '快速模式（React，多轮工具调用）', value: 'react' },
    { label: '规划模式（先规划再执行）', value: 'plan_execute' },
    { label: '深度模式（任务规划+子Agent委派）', value: 'deepagents' },
    { label: '直接模式（跳过Agent框架，直连OpenAI+工具调用）', value: 'direct' },
    { label: '自动（根据问题复杂度判断）', value: '' }
  ]
}

function exportConfig() {
  ExportConfig().then(res => {
    message.info(res)
  })
}

function previewTheme(isDark) {
  const mode = isDark ? 'dark' : 'light'
  applyThemeMode(mode)
  persistThemeMode(mode)
  EventsEmit('themePreview', mode)
}

async function checkForUpdates() {
  if (checkingUpdate.value) return
  checkingUpdate.value = true
  updateResult.value = null
  try {
    const result = await CheckManagedUpdate()
    updateResult.value = result || {available: false}
    if (result?.error) {
      message.error('检查更新失败：' + result.error)
    } else if (result?.available) {
      message.success(`发现新版本 v${result.version || ''}`)
    } else {
      message.info('当前已是最新版本')
    }
  } catch (error) {
    updateResult.value = {available: false, error: error?.message || String(error)}
    message.error('检查更新失败：' + (error?.message || error))
  } finally {
    checkingUpdate.value = false
  }
}

async function applyAvailableUpdate() {
  if (applyingUpdate.value) return
  applyingUpdate.value = true
  try {
    const result = await ApplyManagedUpdate()
    if (result?.error) {
      message.error('更新未启动：' + result.error)
      return
    }
    message.success(result?.message || '更新程序已启动')
  } catch (error) {
    message.error('更新未启动：' + (error?.message || error))
  } finally {
    applyingUpdate.value = false
  }
}

function importConfig() {
  let input = document.createElement('input');
  input.type = 'file';
  input.accept = '.json';
  input.onchange = (e) => {
    let file = e.target.files[0];
    let reader = new FileReader();
    reader.onload = (e) => {
      let config = JSON.parse(e.target.result);
      formValue.value.ID = config.ID
      formValue.value.tushareToken = config.tushareToken
      formValue.value.iwencaiApiKey = config.iwencaiApiKey || ''
      formValue.value.emApiKey = config.emApiKey || ''
      formValue.value.dingPush = {
        enable: config.dingPushEnable,
        dingRobot: config.dingRobot
      }
      formValue.value.feishuPush = {
        enable: config.feishuPushEnable,
        feishuRobot: config.feishuRobot,
        feishuSecret: config.feishuSecret
      }
      formValue.value.feishuBot = {
        enable: config.feishuBotEnable,
        appId: config.feishuAppId || '',
        appSecret: config.feishuAppSecret || '',
        aiConfigId: config.feishuBotAiConfigId || 0,
        sysPromptId: config.feishuBotSysPromptId || 0,
        enableTools: config.feishuBotEnableTools !== false,
        thinking: config.feishuBotThinking === true,
        agentMode: config.feishuBotAgentMode || 'react',
        status: formValue.value.feishuBot.status || 'stopped'
      }
      formValue.value.localPush = {
        enable: config.localPushEnable,
      }
      formValue.value.updateBasicInfoOnStart = config.updateBasicInfoOnStart
      formValue.value.refreshInterval = config.refreshInterval
      formValue.value.enableDanmu = config.enableDanmu
      formValue.value.browserPath = config.browserPath
      formValue.value.enableNews = config.enableNews
      formValue.value.darkTheme = config.darkTheme
      formValue.value.enableFund = config.enableFund
      formValue.value.enablePushNews = config.enablePushNews
      formValue.value.enableOnlyPushRedNews = config.enableOnlyPushRedNews
      formValue.value.httpProxy=config.httpProxy
      formValue.value.httpProxyEnabled=config.httpProxyEnabled
      formValue.value.enableAgent = config.enableAgent
      formValue.value.qgqpBId = config.qgqpBId
      formValue.value.updateChannel = config.updateChannel || 'release'
    };
    reader.readAsText(file);
  };
  input.click();
}


window.onerror = function (event, source, lineno, colno, error) {
  EventsEmit("frontendError", {
    page: "settings.vue",
    message: event,
    source: source,
    lineno: lineno,
    colno: colno,
    error: error ? error.stack : null
  });
  return true;
};

const showManagePromptsModal = ref(false)
const promptTypeOptions = [
  {label: "模型系统Prompt", value: '模型系统Prompt'},
  {label: "模型用户Prompt", value: '模型用户Prompt'},]
const formPromptRef = ref(null)
const formPrompt = ref({
  ID: 0,
  Name: '',
  Content: '',
  Type: '',
})

function managePrompts() {
  formPrompt.value.ID = 0
  showManagePromptsModal.value = true
}

function savePrompt() {
  AddPrompt(formPrompt.value).then(res => {
    message.success(res)
    GetPromptTemplates("", "").then(res => {
      promptTemplates.value = res
    })
    showManagePromptsModal.value = false
  })
}

function editPrompt(prompt) {
  formPrompt.value.ID = prompt.ID
  formPrompt.value.Name = prompt.name
  formPrompt.value.Content = prompt.content
  formPrompt.value.Type = prompt.type
  showManagePromptsModal.value = true
}

function deletePrompt(ID) {
  DelPrompt(ID).then(res => {
    message.success(res)
    GetPromptTemplates("", "").then(res => {
      promptTemplates.value = res
    })
  })
}
</script>

<template>
  <n-flex class="settings-workspace" justify="left" style="text-align: left; --wails-draggable:no-drag">
    <n-form ref="formRef" :label-placement="'left'" :label-align="'left'">
      <n-space vertical size="large">
        <n-card :title="() => h(NTag, { type: 'primary', bordered: false }, () => '基础设置')" size="small">
          <n-grid :cols="24" :x-gap="24" style="text-align: left">
<!--            <n-form-item-gi :span="10" label="Tushare Token：" path="tushareToken">
              <n-input type="text" placeholder="Tushare api token" v-model:value="formValue.tushareToken" clearable/>
            </n-form-item-gi>-->
            <n-form-item-gi :span="4" label="启动时更新基础信息：" path="updateBasicInfoOnStart">
              <n-switch v-model:value="formValue.updateBasicInfoOnStart"/>
            </n-form-item-gi>
            <n-form-item-gi :span="4" label="数据刷新间隔：" path="refreshInterval">
              <n-input-number v-model:value="formValue.refreshInterval" placeholder="请输入数据刷新间隔(秒)">
                <template #suffix>秒</template>
              </n-input-number>
            </n-form-item-gi>
            <n-form-item-gi :span="7" label="外观主题：" path="darkTheme">
              <n-radio-group v-model:value="formValue.darkTheme" size="small" @update:value="previewTheme">
                <n-radio-button :value="false">日间主题</n-radio-button>
                <n-radio-button :value="true">暗黑主题</n-radio-button>
              </n-radio-group>
            </n-form-item-gi>
            <n-form-item-gi :span="7" label="更新通道：" path="updateChannel">
              <n-select v-model:value="formValue.updateChannel" :options="updateChannelOptions" />
              <n-tooltip placement="top">
                <template #trigger>
                  <n-icon color="#0e7a0d" size="20">
                    <HelpCircleFilledIcon />
                  </n-icon>
                </template>
                <template #default>
                  <n-gradient-text :type="'warning'">
                  <div style="max-width: 400px;text-align: left">
                    更新通道说明：<br>
                    <b>Release（稳定版）</b>：仅接收正式发布版本，稳定性最高<br>
                    <b>Pre-release（预发布版）</b>：包含预发布版本，可提前体验新功能<br>
                    <b>Dev（开发版）</b>：包含所有可用版本，获取最新开发进度
                  </div>
                  </n-gradient-text>
                </template>
              </n-tooltip>
            </n-form-item-gi>
            <n-form-item-gi :span="10" label="浏览器安装路径：" path="browserPath">
              <n-input type="text" placeholder="浏览器安装路径" v-model:value="formValue.browserPath" clearable/>
            </n-form-item-gi>
           <n-form-item-gi :span="3" label="指数基金：" path="enableFund">
              <n-switch v-model:value="formValue.enableFund"/>
            </n-form-item-gi>
            <!--      <n-form-item-gi :span="3" label="AI智能体：" path="enableAgent">
                   <n-switch v-model:value="formValue.enableAgent"/>
                 </n-form-item-gi>-->
            <n-form-item-gi :span="11" label="东财唯一标识：" path="qgqpBId">
              <n-input type="text" placeholder="东财唯一标识" v-model:value="formValue.qgqpBId" clearable/>
              <n-tooltip placement="top">
                <template #trigger>
                  <n-icon color="#0e7a0d" size="20">
                    <HelpCircleFilledIcon />
                  </n-icon>
                </template>
                <template #default>
                  <n-gradient-text :type="'warning'">
                  <div style="max-width: 400px;text-align: left">
                    获取方法：<br>
                    打开浏览器,访问东财网站，<br>
                    按F12打开开发人员工具-》网络面板，<br>
                    随便点开一个请求，复制请求cookie中qgqp_b_id对应的值。
                  </div>
                  </n-gradient-text>
                </template>
              </n-tooltip>
            </n-form-item-gi>

            <n-form-item-gi :span="11" label="问财API密钥：" path="iwencaiApiKey">
              <n-input type="password" placeholder="同花顺问财开放平台API Key" v-model:value="formValue.iwencaiApiKey" clearable show-password-on="click"/>
              <n-tooltip placement="top">
                <template #trigger>
                  <n-icon color="#0e7a0d" size="20">
                    <HelpCircleFilledIcon />
                  </n-icon>
                </template>
                <template #default>
                  <n-gradient-text :type="'warning'">
                  <div style="max-width: 400px;text-align: left">
                    获取方法：<br>
                    访问同花顺问财开放平台：<br>
                    <a href="https://open.iwencai.com" target="_blank" style="color: #63e2b7">https://www.iwencai.com/skillhub</a><br>
                    注册并登录后，在控制台获取API Key。<br>
                    配置后可使用问财智能选股、行情查询、研报搜索等功能。
                  </div>
                  </n-gradient-text>
                </template>
              </n-tooltip>
            </n-form-item-gi>

            <n-form-item-gi :span="11" label="东财AI密钥：" path="emApiKey">
              <n-input type="password" placeholder="东方财富AI SaaS API Key" v-model:value="formValue.emApiKey" clearable show-password-on="click"/>
              <n-tooltip placement="top">
                <template #trigger>
                  <n-icon color="#0e7a0d" size="20">
                    <HelpCircleFilledIcon />
                  </n-icon>
                </template>
                <template #default>
                  <n-gradient-text :type="'warning'">
                  <div style="max-width: 400px;text-align: left">
                    获取方法：<br>
                    访问东方财富妙想AI平台获取API Key。
                    https://ai.eastmoney.com/mxClaw<br>
                    配置后可使用个股业绩点评功能。
                  </div>
                  </n-gradient-text>
                </template>
              </n-tooltip>
            </n-form-item-gi>

          </n-grid>
          <div class="settings-update-panel">
            <div>
              <strong>应用更新</strong>
              <span v-if="!updateResult">仅在你点击后联网检查，不再在启动时弹出更新提示。</span>
              <span v-else-if="updateResult.error">检查失败：{{ updateResult.error }}</span>
              <span v-else-if="updateResult.available">发现 v{{ updateResult.version }}：{{ updateResult.notes || '包含新的功能与修复。' }}</span>
              <span v-else>当前已是最新版本。</span>
            </div>
            <n-space>
              <n-button secondary type="primary" :loading="checkingUpdate" @click="checkForUpdates">检查更新</n-button>
              <n-button v-if="updateResult?.available" type="primary" :loading="applyingUpdate" @click="applyAvailableUpdate">下载并更新</n-button>
            </n-space>
          </div>
        </n-card>

        <n-card :title="() => h(NTag, { type: 'primary', bordered: false }, () => '通知设置')" size="small">
          <n-grid :cols="24" :x-gap="24" style="text-align: left">
            <n-form-item-gi :span="3" label="钉钉推送：" path="dingPush.enable">
              <n-switch v-model:value="formValue.dingPush.enable"/>
            </n-form-item-gi>
            <n-form-item-gi :span="3" label="飞书推送：" path="feishuPush.enable">
              <n-switch v-model:value="formValue.feishuPush.enable"/>
            </n-form-item-gi>
            <n-form-item-gi :span="3" label="本地推送：" path="localPush.enable">
              <n-switch v-model:value="formValue.localPush.enable"/>
            </n-form-item-gi>
            <n-form-item-gi :span="3" label="弹幕功能：" path="enableDanmu">
              <n-switch v-model:value="formValue.enableDanmu"/>
            </n-form-item-gi>
            <n-form-item-gi :span="3" label="显示滚动快讯：" path="enableNews">
              <n-switch v-model:value="formValue.enableNews"/>
            </n-form-item-gi>
            <n-form-item-gi :span="3" label="市场资讯提醒：" path="enablePushNews">
              <n-switch v-model:value="formValue.enablePushNews"/>
            </n-form-item-gi>
            <n-form-item-gi v-if="formValue.enablePushNews" :span="4" label="只提醒红字或关注个股的新闻：" path="enableOnlyPushRedNews">
              <n-switch v-model:value="formValue.enableOnlyPushRedNews"/>
            </n-form-item-gi>

            <n-form-item-gi :span="22" v-if="formValue.dingPush.enable" label="钉钉机器人接口地址："
                            path="dingPush.dingRobot">
              <n-input placeholder="请输入钉钉机器人接口地址" v-model:value="formValue.dingPush.dingRobot"/>
              <n-button type="primary" @click="sendTestNotice">发送测试通知</n-button>
            </n-form-item-gi>

            <n-form-item-gi :span="22" v-if="formValue.feishuPush.enable" label="飞书机器人接口地址："
                            path="feishuPush.feishuRobot">
              <n-input placeholder="请输入飞书自定义机器人 Webhook 地址（https://open.feishu.cn/open-apis/bot/v2/hook/...）"
                       v-model:value="formValue.feishuPush.feishuRobot"/>
            </n-form-item-gi>
            <n-form-item-gi :span="22" v-if="formValue.feishuPush.enable" label="飞书签名校验 Secret："
                            path="feishuPush.feishuSecret">
              <n-input type="password" show-password-on="click"
                       placeholder="可选：填写机器人安全设置的签名校验 Secret，留空则不签名"
                       v-model:value="formValue.feishuPush.feishuSecret"/>
              <n-button type="primary" @click="sendFeishuTestNotice">发送测试通知</n-button>
            </n-form-item-gi>

          </n-grid>
        </n-card>

        <n-card :title="() => h(NTag, { type: 'info', bordered: false }, () => '飞书应用机器人（AI对话）')" size="small">
          <n-grid :cols="24" :x-gap="24" style="text-align: left">
            <n-form-item-gi :span="24" label="启用飞书应用机器人" path="feishuBot.enable">
              <n-switch v-model:value="formValue.feishuBot.enable"/>
              <n-text depth="3" style="margin-left: 12px; font-size: 12px;">
                通过长连接接收飞书用户消息，由 AI Agent 自动回复（与上方自定义机器人推送完全独立）
              </n-text>
            </n-form-item-gi>

            <n-form-item-gi :span="12" v-if="formValue.feishuBot.enable" label="App ID"
                            path="feishuBot.appId">
              <n-input placeholder="飞书应用 App ID（cli_xxx）" v-model:value="formValue.feishuBot.appId"/>
            </n-form-item-gi>
            <n-form-item-gi :span="12" v-if="formValue.feishuBot.enable" label="App Secret"
                            path="feishuBot.appSecret">
              <n-input type="password" show-password-on="click"
                       placeholder="飞书应用 App Secret" v-model:value="formValue.feishuBot.appSecret"/>
            </n-form-item-gi>

            <n-form-item-gi :span="8" v-if="formValue.feishuBot.enable" label="AI 配置"
                            path="feishuBot.aiConfigId">
              <n-select v-model:value="formValue.feishuBot.aiConfigId"
                        :options="aiConfigOptions()"
                        filterable
                        placeholder="选择上方 AI 模型服务配置"/>
            </n-form-item-gi>
            <n-form-item-gi :span="8" v-if="formValue.feishuBot.enable" label="系统提示词"
                            path="feishuBot.sysPromptId">
              <n-select v-model:value="formValue.feishuBot.sysPromptId"
                        :options="promptTemplateOptions()"
                        filterable
                        placeholder="选择系统提示词模板（可选）"/>
            </n-form-item-gi>
            <n-form-item-gi :span="8" v-if="formValue.feishuBot.enable" label="Agent 模式"
                            path="feishuBot.agentMode">
              <n-select v-model:value="formValue.feishuBot.agentMode"
                        :options="agentModeOptions()"/>
            </n-form-item-gi>
            <n-form-item-gi :span="6" v-if="formValue.feishuBot.enable" label="启用工具调用"
                            path="feishuBot.enableTools">
              <n-switch v-model:value="formValue.feishuBot.enableTools"/>
              <n-text depth="3" style="margin-left: 12px; font-size: 12px;">
                关闭后走单轮对话（不调用工具）
              </n-text>
            </n-form-item-gi>
            <n-form-item-gi :span="6" v-if="formValue.feishuBot.enable" label="深度思考"
                            path="feishuBot.thinking">
              <n-switch v-model:value="formValue.feishuBot.thinking"/>
              <n-text depth="3" style="margin-left: 12px; font-size: 12px;">
                推理模型启用后输出思考过程
              </n-text>
            </n-form-item-gi>

            <n-form-item-gi :span="24" v-if="formValue.feishuBot.enable">
              <n-space>
                <n-button type="primary" @click="startFeishuBot">启动机器人</n-button>
                <n-button type="warning" @click="stopFeishuBot">停止机器人</n-button>
                <n-button @click="refreshFeishuBotStatus">查询状态</n-button>
                <n-tag :type="formValue.feishuBot.status === 'running' ? 'success' : 'default'">
                  {{ formValue.feishuBot.status === 'running' ? '运行中' : '已停止' }}
                </n-tag>
              </n-space>
            </n-form-item-gi>

            <n-form-item-gi :span="24" v-if="formValue.feishuBot.enable">
              <n-gradient-text type="info">
                <div style="font-size: 12px; line-height: 1.6;">
                  配置步骤：<br>
                  1. 在飞书开放平台创建企业自建应用，获取 App ID 和 App Secret<br>
                  2. 启用「机器人」能力，添加事件订阅 im.message.receive_v1<br>
                  3. 事件订阅页选择「使用长连接接收事件」（无需公网 IP）<br>
                  4. 应用需发布并通过审核；权限需添加「获取用户发给机器人的单聊消息」「获取群组中所有消息」或「获取用户在群组中@机器人的消息」<br>
                  5. 单聊直接回复；群聊需 @机器人才会回复；多轮对话记忆按 sessionID 自动隔离<br>
                  文档：https://open.feishu.cn/document/server-side-sdk/golang-sdk-guide/handle-events
                </div>
              </n-gradient-text>
            </n-form-item-gi>
          </n-grid>
        </n-card>

        <n-card :title="() => h(NTag, { type: 'primary', bordered: false }, () => '设置操作')" size="small">
          <n-flex justify="center" :wrap="true">
            <n-button type="primary" strong @click="saveConfig">保存设置</n-button>
            <n-button type="success" secondary :loading="checkingUpdate" @click="checkForUpdates">检查更新</n-button>
            <n-button type="info" @click="exportConfig">导出配置</n-button>
            <n-button type="error" @click="importConfig">导入配置</n-button>
          </n-flex>
        </n-card>
      </n-space>
    </n-form>
  </n-flex>

  <n-modal v-model:show="showManagePromptsModal" closable :mask-closable="false">
    <n-card style="width: 800px; height: 600px; text-align: left" :bordered="false"
            :title="(formPrompt.ID > 0 ? '修改' : '添加') + '提示词'" size="huge" role="dialog" aria-modal="true">
      <n-form ref="formPromptRef" :label-placement="'left'" :label-align="'left'">
        <n-form-item label="名称">
          <n-input v-model:value="formPrompt.Name" placeholder="请输入提示词名称"/>
        </n-form-item>
        <n-form-item label="类型">
          <n-select v-model:value="formPrompt.Type" :options="promptTypeOptions" placeholder="请选择提示词类型"/>
        </n-form-item>
        <n-form-item label="内容">
          <n-input v-model:value="formPrompt.Content" type="textarea" :show-count="true" placeholder="请输入prompt"
                   :autosize="{ minRows: 12, maxRows: 12, }"/>
        </n-form-item>
      </n-form>
      <template #footer>
        <n-flex justify="end">
          <n-button type="primary" @click="savePrompt">保存</n-button>
          <n-button type="warning" @click="showManagePromptsModal = false">取消</n-button>
        </n-flex>
      </template>
    </n-card>
  </n-modal>
</template>

<style scoped>
.settings-workspace { width:100%; padding:14px; color:var(--workspace-text); }
.settings-workspace :deep(.n-form), .settings-workspace :deep(.n-space) { width:100%; }
.settings-update-panel { display:flex; align-items:center; justify-content:space-between; gap:16px; margin-top:4px; padding:13px 14px; border:1px solid var(--workspace-border); border-radius:var(--workspace-radius-md); color:var(--workspace-text); background:var(--workspace-surface-muted); }
.settings-update-panel > div { min-width:0; display:grid; gap:3px; }
.settings-update-panel strong { font-size:13px; }
.settings-update-panel span { color:var(--workspace-text-muted); font-size:11px; line-height:1.5; }
.cardHeaderClass {
  font-size: 16px;
  font-weight: bold;
  color: red;
}
@media (max-width:1080px) {
  .settings-workspace { padding:10px; }
  .settings-update-panel { align-items:flex-start; flex-direction:column; }
  .settings-workspace :deep(.n-grid) { min-width:820px; }
  .settings-workspace :deep(.n-card__content) { overflow-x:auto; }
}
</style>
