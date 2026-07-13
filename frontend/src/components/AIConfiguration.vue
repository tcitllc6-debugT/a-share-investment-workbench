<script setup>
import { computed, onMounted, ref } from 'vue'
import { useDialog, useMessage } from 'naive-ui'
import {
  AddOutline,
  CheckmarkCircleOutline,
  FlaskOutline,
  KeyOutline,
  OptionsOutline,
  RefreshOutline,
  SaveOutline,
  ServerOutline,
  ShieldCheckmarkOutline,
  TrashOutline,
} from '@vicons/ionicons5'
import {
  GetAIWorkspaceConfig,
  SaveAIWorkspaceConfig,
  TestAIConnection,
} from '../../wailsjs/go/main/App'
import {
  AI_PROVIDER_PRESETS,
  connectionResultTone,
  createAIService,
  displayServiceName,
  endpointChanged,
  providerForService,
  validateAIService,
} from './ai-configuration.mjs'

const message = useMessage()
const dialog = useDialog()
const loading = ref(true)
const saving = ref(false)
const selectedIndex = ref(0)
const advancedOpen = ref(false)
const researchOpen = ref(false)
const errors = ref({})
const stockPromptHint = '支持 {{stockName}}、{{stockCode}}'
const stockPromptPlaceholder = '{{stockName}}[{{stockCode}}]分析和总结…'
let clientKeySeed = 0

const workspace = ref({
  enabled: false,
  prompt: '',
  questionTemplate: '{{stockName}}[{{stockCode}}]分析和总结',
  crawlTimeOut: 60,
  kDays: 60,
  httpProxy: '',
  httpProxyEnabled: false,
  services: [],
})

const selectedService = computed(() => workspace.value.services[selectedIndex.value] || null)
const providerOptions = AI_PROVIDER_PRESETS.map((item) => ({
  label: item.name,
  value: item.key,
}))
const enabledServices = computed(() => workspace.value.services.filter((item) => item.modelName && item.baseUrl))
const defaultResearch = computed(() => workspace.value.services.find((item) => item.isDefaultResearch))
const defaultAgent = computed(() => workspace.value.services.find((item) => item.isDefaultAgent))

function hydrateService(service) {
  const provider = providerForService(service)
  return {
    ...createAIService(provider?.key),
    ...service,
    providerKey: provider?.key || 'custom',
    apiKey: '',
    _originalBaseUrl: service.ID ? service.baseUrl : '',
    _clientKey: 'ai-service-' + (++clientKeySeed),
    _modelOptions: service.modelName ? [{ label: service.modelName, value: service.modelName }] : [],
    _connection: null,
  }
}

async function loadWorkspace() {
  loading.value = true
  try {
    const result = await GetAIWorkspaceConfig()
    workspace.value = {
      enabled: result?.enabled === true,
      prompt: result?.prompt || '',
      questionTemplate: result?.questionTemplate || '{{stockName}}[{{stockCode}}]分析和总结',
      crawlTimeOut: result?.crawlTimeOut || 60,
      kDays: result?.kDays || 60,
      httpProxy: result?.httpProxy || '',
      httpProxyEnabled: result?.httpProxyEnabled === true,
      services: (result?.services || []).map(hydrateService),
    }
    selectedIndex.value = Math.min(selectedIndex.value, Math.max(0, workspace.value.services.length - 1))
  } catch (error) {
    message.error('AI 配置加载失败：' + (error?.message || String(error)))
  } finally {
    loading.value = false
  }
}

function addService(providerKey = 'deepseek') {
  const service = hydrateService(createAIService(providerKey))
  if (!workspace.value.services.length) {
    service.isDefaultResearch = true
    service.isDefaultAgent = true
    workspace.value.enabled = true
  }
  workspace.value.services.push(service)
  selectedIndex.value = workspace.value.services.length - 1
  errors.value = {}
}

function selectService(index) {
  selectedIndex.value = index
  errors.value = {}
}

function applyProvider(providerKey) {
  const current = selectedService.value
  if (!current) return
  const preset = createAIService(providerKey)
  Object.assign(current, {
    providerKey,
    name: current.ID ? current.name : preset.name,
    baseUrl: preset.baseUrl,
    modelName: preset.modelName,
    maxTokens: preset.maxTokens,
    thinking: preset.thinking,
    _modelOptions: preset._modelOptions,
    _connection: null,
  })
}

function setDefault(kind, index, checked) {
  if (!checked) return
  const field = kind === 'research' ? 'isDefaultResearch' : 'isDefaultAgent'
  workspace.value.services.forEach((item, itemIndex) => {
    item[field] = itemIndex === index
  })
}

function removeSelected() {
  const service = selectedService.value
  if (!service) return
  dialog.warning({
    title: '删除模型服务',
    content: '确定删除“' + displayServiceName(service) + '”吗？保存后使用它的 AI 任务需要重新选择模型。',
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: () => {
      workspace.value.services.splice(selectedIndex.value, 1)
      selectedIndex.value = Math.max(0, selectedIndex.value - 1)
      if (!workspace.value.services.length) {
        workspace.value.enabled = false
      }
      if (workspace.value.services.length && !workspace.value.services.some((item) => item.isDefaultResearch)) {
        workspace.value.services[0].isDefaultResearch = true
      }
      if (workspace.value.services.length && !workspace.value.services.some((item) => item.isDefaultAgent)) {
        workspace.value.services[0].isDefaultAgent = true
      }
    },
  })
}

function servicePayload(service) {
  const { _clientKey, _modelOptions, _connection, _originalBaseUrl, providerKey, ...payload } = service
  return payload
}

async function testConnection(service = selectedService.value) {
  if (!service || service._testing) return
  const serviceErrors = validateAIService(service, { requireModel: false })
  errors.value = serviceErrors
  if (Object.keys(serviceErrors).length) {
    message.warning('请先完成密钥和接口地址')
    return
  }
  service._testing = true
  try {
    const result = await TestAIConnection(servicePayload(service))
    service._connection = result
    if (result?.models?.length) {
      service._modelOptions = result.models.map((id) => ({ label: id, value: id }))
      if (!service.modelName || !result.models.includes(service.modelName)) {
        service.modelName = result.models[0]
      }
    }
    const tone = connectionResultTone(result)
    if (tone === 'success') message.success(result.message)
    else if (tone === 'warning') message.warning(result.message)
    else message.error(result?.message || '连接测试失败')
  } catch (error) {
    service._connection = { success: false, code: 'client_error', message: error?.message || String(error) }
    message.error('连接测试失败：' + (error?.message || String(error)))
  } finally {
    service._testing = false
  }
}

async function fetchModels() {
  await testConnection()
}

function validateWorkspace() {
  if (workspace.value.enabled && !workspace.value.services.length) {
    message.warning('启用 AI 前至少添加一个模型服务')
    return false
  }
  for (let index = 0; index < workspace.value.services.length; index += 1) {
    const serviceErrors = validateAIService(workspace.value.services[index])
    if (Object.keys(serviceErrors).length) {
      selectedIndex.value = index
      errors.value = serviceErrors
      message.warning('请完善“' + displayServiceName(workspace.value.services[index]) + '”')
      return false
    }
  }
  if (workspace.value.services.length && !workspace.value.services.some((item) => item.isDefaultResearch)) {
    workspace.value.services[0].isDefaultResearch = true
  }
  if (workspace.value.services.length && !workspace.value.services.some((item) => item.isDefaultAgent)) {
    workspace.value.services[0].isDefaultAgent = true
  }
  return true
}

async function saveWorkspace() {
  if (!validateWorkspace() || saving.value) return
  saving.value = true
  try {
    const payload = {
      ...workspace.value,
      services: workspace.value.services.map(servicePayload),
    }
    const result = await SaveAIWorkspaceConfig(payload)
    if (result !== '保存成功！') throw new Error(result)
    message.success('AI 配置已保存')
    await loadWorkspace()
  } catch (error) {
    message.error('保存失败：' + (error?.message || String(error)))
  } finally {
    saving.value = false
  }
}

onMounted(loadWorkspace)
</script>

<template>
  <section class="ai-configuration" :aria-busy="loading">
    <div class="ai-configuration__status">
      <div class="ai-status-copy">
        <span class="ai-status-copy__icon"><n-icon :component="ShieldCheckmarkOutline" /></span>
        <div>
          <strong>AI 能力中心</strong>
          <p>模型密钥只保存在本机，不会随连接测试发送股票、自选或持仓数据。</p>
        </div>
      </div>
      <div class="ai-status-metrics">
        <span><b>{{ workspace.enabled ? '已启用' : '未启用' }}</b><small>AI 研判</small></span>
        <span><b>{{ enabledServices.length }}</b><small>可用服务</small></span>
        <span><b>{{ defaultResearch ? displayServiceName(defaultResearch) : '未设置' }}</b><small>默认研究模型</small></span>
        <span><b>{{ defaultAgent ? displayServiceName(defaultAgent) : '未设置' }}</b><small>默认智能体模型</small></span>
      </div>
    </div>

    <div class="ai-configuration__layout">
      <aside class="ai-service-list" aria-label="模型服务列表">
        <div class="ai-section-heading">
          <div><span>模型服务</span><small>一个服务对应一个模型</small></div>
          <n-dropdown :options="providerOptions" trigger="click" @select="addService">
            <n-button type="primary" size="small" :disabled="loading">
              <template #icon><n-icon :component="AddOutline" /></template>
              添加
            </n-button>
          </n-dropdown>
        </div>

        <button
          v-for="(service, index) in workspace.services"
          :key="service._clientKey"
          type="button"
          class="ai-service-item"
          :class="{ 'is-active': index === selectedIndex }"
          @click="selectService(index)"
        >
          <span class="ai-service-item__glyph"><n-icon :component="ServerOutline" /></span>
          <span class="ai-service-item__body">
            <b>{{ displayServiceName(service) }}</b>
            <small>{{ service.modelName || '还没有选择模型' }}</small>
            <em>
              <i v-if="service.isDefaultResearch">研究</i>
              <i v-if="service.isDefaultAgent">智能体</i>
              <i v-if="service._connection?.success" class="is-ok">连接正常</i>
            </em>
          </span>
        </button>

        <div v-if="!workspace.services.length && !loading" class="ai-service-empty">
          <n-icon :component="ServerOutline" />
          <strong>还没有模型服务</strong>
          <p>从 DeepSeek、通义千问或本机 Ollama 开始，通常 1 分钟即可完成。</p>
          <n-button type="primary" size="small" @click="addService('deepseek')">添加第一个服务</n-button>
        </div>
      </aside>

      <main class="ai-service-editor">
        <template v-if="selectedService">
          <header class="ai-service-editor__header">
            <div>
              <span>当前模型服务</span>
              <h2>{{ displayServiceName(selectedService) }}</h2>
            </div>
            <n-button quaternary type="error" size="small" @click="removeSelected">
              <template #icon><n-icon :component="TrashOutline" /></template>
              删除
            </n-button>
          </header>

          <div class="ai-form-section">
            <div class="ai-form-section__title">
              <span class="step-index">1</span>
              <div><h3>选择服务商</h3><p>选择后会自动填好常用接口地址。</p></div>
            </div>
            <n-select
              v-model:value="selectedService.providerKey"
              :options="providerOptions"
              aria-label="选择服务商"
              @update:value="applyProvider"
            />
          </div>

          <div class="ai-form-section">
            <div class="ai-form-section__title">
              <span class="step-index">2</span>
              <div><h3>连接账户</h3><p>密钥保存在本机数据库，页面重新打开后只显示“已保存”。</p></div>
            </div>
            <div class="form-grid form-grid--two">
              <label>
                <span>服务名称</span>
                <n-input v-model:value="selectedService.name" name="ai-service-name" autocomplete="off" placeholder="例如：DeepSeek 日常研究…" />
              </label>
              <label>
                <span>接口地址</span>
                <n-input v-model:value="selectedService.baseUrl" name="ai-base-url" autocomplete="off" placeholder="https://api.example.com/v1…" />
                <small v-if="errors.baseUrl" class="field-error">{{ errors.baseUrl }}</small>
              </label>
              <label class="form-grid__wide">
                <span>API 密钥
                  <em v-if="endpointChanged(selectedService)">接口已变更，需重新输入</em>
                  <em v-else-if="selectedService.hasApiKey">已保存，留空不会修改</em>
                </span>
                <n-input
                  v-model:value="selectedService.apiKey"
                  type="password"
                  name="ai-api-key"
                  autocomplete="off"
                  show-password-on="click"
                  :placeholder="endpointChanged(selectedService) ? '接口已变更，请输入新服务的 API 密钥…' : (selectedService.hasApiKey ? '已安全保存；输入新密钥才会替换…' : '粘贴服务商提供的 API 密钥…')"
                >
                  <template #prefix><n-icon :component="KeyOutline" /></template>
                </n-input>
                <small v-if="errors.apiKey" class="field-error">{{ errors.apiKey }}</small>
              </label>
            </div>
          </div>

          <div class="ai-form-section">
            <div class="ai-form-section__title">
              <span class="step-index">3</span>
              <div><h3>选择模型并测试</h3><p>先获取模型；服务不支持时也可以手动输入。</p></div>
            </div>
            <div class="model-row">
              <n-select
                v-model:value="selectedService.modelName"
                :options="selectedService._modelOptions"
                filterable
                tag
                aria-label="模型名称"
                placeholder="选择或输入模型名称…"
              />
              <n-button secondary :loading="selectedService._testing" @click="fetchModels">
                <template #icon><n-icon :component="RefreshOutline" /></template>
                获取模型
              </n-button>
              <n-button type="primary" secondary :loading="selectedService._testing" @click="testConnection()">
                <template #icon><n-icon :component="FlaskOutline" /></template>
                测试连接
              </n-button>
            </div>
            <small v-if="errors.modelName" class="field-error">{{ errors.modelName }}</small>
            <div
              v-if="selectedService._connection"
              class="connection-result"
              :class="'is-' + connectionResultTone(selectedService._connection)"
              aria-live="polite"
            >
              <n-icon :component="selectedService._connection.success ? CheckmarkCircleOutline : OptionsOutline" />
              <span>
                <b>{{ selectedService._connection.message }}</b>
                <small v-if="selectedService._connection.latencyMs >= 0">
                  耗时 {{ selectedService._connection.latencyMs }} ms
                  <template v-if="selectedService._connection.models?.length"> · 找到 {{ selectedService._connection.models.length }} 个模型</template>
                </small>
              </span>
            </div>
          </div>

          <div class="ai-form-section">
            <div class="ai-form-section__title">
              <span class="step-index">4</span>
              <div><h3>指定用途</h3><p>研究模型负责分析，智能体模型需要支持工具调用。</p></div>
            </div>
            <div class="purpose-options">
              <label>
                <n-checkbox
                  :checked="selectedService.isDefaultResearch"
                  @update:checked="(checked) => setDefault('research', selectedIndex, checked)"
                />
                <span><b>默认研究模型</b><small>自选股研判、市场总结、研究报告</small></span>
              </label>
              <label>
                <n-checkbox
                  :checked="selectedService.isDefaultAgent"
                  @update:checked="(checked) => setDefault('agent', selectedIndex, checked)"
                />
                <span><b>默认智能体模型</b><small>AI 助手、工具调用、自动化任务</small></span>
              </label>
            </div>
          </div>

          <n-collapse-transition :show="advancedOpen">
            <div class="ai-form-section ai-form-section--advanced">
              <div class="form-grid form-grid--three">
                <label>
                  <span>回答风格 <em>越低越稳定</em></span>
                  <n-slider v-model:value="selectedService.temperature" :min="0" :max="2" :step="0.1" />
                  <small>{{ selectedService.temperature }}</small>
                </label>
                <label>
                  <span>最长回答</span>
                  <n-input-number v-model:value="selectedService.maxTokens" :min="256" :max="262144" :step="1024" />
                </label>
                <label>
                  <span>响应超时（秒）</span>
                  <n-input-number v-model:value="selectedService.timeOut" :min="10" :max="900" :step="10" />
                </label>
              </div>
              <div class="toggle-row">
                <label><n-switch v-model:value="selectedService.thinking" /><span><b>深度思考</b><small>仅在模型明确支持推理时开启</small></span></label>
                <label><n-switch v-model:value="selectedService.httpProxyEnabled" /><span><b>模型代理</b><small>仅影响这个模型服务</small></span></label>
              </div>
              <n-input v-if="selectedService.httpProxyEnabled" v-model:value="selectedService.httpProxy" placeholder="http://127.0.0.1:7890…" />
            </div>
          </n-collapse-transition>
          <button class="disclosure-button" type="button" @click="advancedOpen = !advancedOpen">
            <n-icon :component="OptionsOutline" />高级设置{{ advancedOpen ? '收起' : '展开' }}
          </button>
        </template>

        <div v-else-if="!loading" class="editor-empty">
          <n-icon :component="ServerOutline" />
          <h2>添加一个模型服务开始配置</h2>
          <p>推荐第一次使用选择 DeepSeek；希望完全本地运行可以选择 Ollama。</p>
        </div>
      </main>
    </div>

    <section class="research-preferences">
      <button type="button" class="research-preferences__header" @click="researchOpen = !researchOpen">
        <span><n-icon :component="OptionsOutline" /><b>研究偏好</b><small>诊股数据范围、默认提示词和爬虫代理</small></span>
        <em>{{ researchOpen ? '收起' : '展开' }}</em>
      </button>
      <n-collapse-transition :show="researchOpen">
        <div class="research-preferences__body">
          <div class="toggle-row">
            <label><n-switch v-model:value="workspace.enabled" /><span><b>启用 AI 研判</b><small>关闭后保留配置，但不显示 AI 分析能力</small></span></label>
            <label><n-switch v-model:value="workspace.httpProxyEnabled" /><span><b>爬虫代理</b><small>仅用于资讯采集，不影响行情数据</small></span></label>
          </div>
          <div class="form-grid form-grid--three">
            <label><span>资讯采集超时（秒）</span><n-input-number v-model:value="workspace.crawlTimeOut" :min="30" :max="300" /></label>
            <label><span>K 线历史天数</span><n-input-number v-model:value="workspace.kDays" :min="30" :max="120" /></label>
            <label v-if="workspace.httpProxyEnabled"><span>爬虫代理地址</span><n-input v-model:value="workspace.httpProxy" placeholder="http://127.0.0.1:7890…" /></label>
          </div>
          <div class="prompt-grid">
            <label><span>默认系统提示词 <em>定义 AI 的分析原则</em></span><n-input v-model:value="workspace.prompt" type="textarea" :autosize="{ minRows: 4, maxRows: 8 }" placeholder="例如：你是严谨的 A 股研究助理…" /></label>
            <label><span>默认个股分析提示词 <em>{{ stockPromptHint }}</em></span><n-input v-model:value="workspace.questionTemplate" type="textarea" :autosize="{ minRows: 4, maxRows: 8 }" :placeholder="stockPromptPlaceholder" /></label>
          </div>
        </div>
      </n-collapse-transition>
    </section>

    <footer class="ai-configuration__footer">
      <div><n-icon :component="ShieldCheckmarkOutline" /><span>密钥只保存在本机；连接测试只访问模型服务的模型列表。</span></div>
      <n-button type="primary" :loading="saving" @click="saveWorkspace">
        <template #icon><n-icon :component="SaveOutline" /></template>
        保存 AI 配置
      </n-button>
    </footer>
  </section>
</template>

<style scoped>
.ai-configuration { display:flex; flex-direction:column; gap:14px; min-height:0; max-height:calc(100vh - 176px); overflow-y:auto; padding:2px 2px 24px; color:var(--workspace-text); }
.ai-configuration__status, .ai-service-list, .ai-service-editor, .research-preferences, .ai-configuration__footer { border:1px solid var(--workspace-border); background:var(--workspace-surface-solid); box-shadow:var(--workspace-shadow-sm); }
.ai-configuration__status { display:flex; align-items:center; justify-content:space-between; gap:18px; padding:15px 18px; border-radius:var(--workspace-radius-md); }
.ai-status-copy { display:flex; align-items:center; gap:12px; min-width:0; }
.ai-status-copy__icon { display:grid; place-items:center; flex:0 0 auto; width:36px; height:36px; border-radius:10px; color:var(--workspace-accent); background:var(--workspace-accent-soft); font-size:20px; }
.ai-status-copy strong { display:block; font-size:14px; }.ai-status-copy p { margin:3px 0 0; color:var(--workspace-text-muted); font-size:11px; line-height:1.5; }
.ai-status-metrics { display:grid; grid-template-columns:repeat(4,minmax(92px,1fr)); min-width:min(620px,58%); }
.ai-status-metrics span { min-width:0; padding:0 14px; border-left:1px solid var(--workspace-border); }.ai-status-metrics b,.ai-status-metrics small{display:block;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.ai-status-metrics b{font-size:12px}.ai-status-metrics small{margin-top:4px;color:var(--workspace-text-muted);font-size:10px}
.ai-configuration__layout { display:grid; grid-template-columns:310px minmax(0,1fr); gap:14px; align-items:start; }
.ai-service-list,.ai-service-editor{border-radius:var(--workspace-radius-md)}.ai-service-list{position:sticky;top:0;padding:12px;max-height:calc(100vh - 280px);overflow-y:auto}.ai-service-editor{min-width:0;padding:18px}
.ai-section-heading,.ai-service-editor__header{display:flex;align-items:center;justify-content:space-between;gap:12px}.ai-section-heading{padding:2px 2px 12px}.ai-section-heading span{display:block;font-size:13px;font-weight:700}.ai-section-heading small{display:block;margin-top:2px;color:var(--workspace-text-muted);font-size:10px}
.ai-service-item{display:flex;align-items:flex-start;gap:10px;width:100%;margin:0 0 7px;padding:11px;border:1px solid transparent;border-radius:var(--workspace-radius-sm);color:var(--workspace-text);background:transparent;text-align:left;cursor:pointer}.ai-service-item:hover{background:var(--workspace-surface-hover)}.ai-service-item.is-active{border-color:var(--workspace-accent);background:var(--workspace-accent-soft)}
.ai-service-item__glyph{display:grid;place-items:center;flex:0 0 auto;width:30px;height:30px;border-radius:8px;color:var(--workspace-accent);background:var(--workspace-surface-muted)}.ai-service-item__body{min-width:0;flex:1}.ai-service-item__body b,.ai-service-item__body small{display:block;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.ai-service-item__body b{font-size:12px}.ai-service-item__body small{margin-top:3px;color:var(--workspace-text-muted);font-size:10px}.ai-service-item__body em{display:flex;gap:4px;margin-top:7px}.ai-service-item__body i{padding:2px 5px;border-radius:4px;color:var(--workspace-accent);background:var(--workspace-accent-soft);font-size:9px;font-style:normal}.ai-service-item__body i.is-ok{color:var(--workspace-negative)}
.ai-service-empty,.editor-empty{display:flex;align-items:center;flex-direction:column;padding:34px 18px;color:var(--workspace-text-muted);text-align:center}.ai-service-empty>.n-icon,.editor-empty>.n-icon{font-size:28px;color:var(--workspace-accent)}.ai-service-empty strong,.editor-empty h2{margin:10px 0 0;color:var(--workspace-text);font-size:14px}.ai-service-empty p,.editor-empty p{max-width:420px;margin:6px 0 14px;font-size:11px;line-height:1.6}
.ai-service-editor__header{padding-bottom:14px;border-bottom:1px solid var(--workspace-border)}.ai-service-editor__header span{color:var(--workspace-text-muted);font-size:10px}.ai-service-editor__header h2{margin:3px 0 0;font-size:17px}
.ai-form-section{padding:18px 0;border-bottom:1px solid var(--workspace-border)}.ai-form-section__title{display:flex;align-items:flex-start;gap:10px;margin-bottom:12px}.step-index{display:grid;place-items:center;flex:0 0 auto;width:22px;height:22px;border-radius:7px;color:var(--workspace-accent);background:var(--workspace-accent-soft);font-size:11px;font-weight:700}.ai-form-section__title h3{margin:1px 0 0;font-size:13px}.ai-form-section__title p{margin:3px 0 0;color:var(--workspace-text-muted);font-size:10px}
.form-grid{display:grid;gap:13px}.form-grid--two{grid-template-columns:repeat(2,minmax(0,1fr))}.form-grid--three{grid-template-columns:repeat(3,minmax(0,1fr))}.form-grid__wide{grid-column:1/-1}.form-grid label,.prompt-grid label{min-width:0}.form-grid label>span,.prompt-grid label>span{display:flex;align-items:center;justify-content:space-between;gap:8px;margin-bottom:6px;font-size:11px;font-weight:650}.form-grid label em,.prompt-grid label em{color:var(--workspace-text-muted);font-size:9px;font-style:normal;font-weight:400}.field-error{display:block;margin-top:4px;color:var(--workspace-positive);font-size:10px}
.model-row{display:grid;grid-template-columns:minmax(0,1fr) auto auto;gap:8px}.connection-result{display:flex;align-items:flex-start;gap:8px;margin-top:10px;padding:10px 12px;border-radius:var(--workspace-radius-sm);background:var(--workspace-surface-muted)}.connection-result.is-success{color:var(--workspace-negative)}.connection-result.is-warning{color:var(--workspace-warning)}.connection-result.is-error{color:var(--workspace-positive)}.connection-result span{min-width:0}.connection-result b,.connection-result small{display:block}.connection-result b{font-size:11px}.connection-result small{margin-top:3px;color:var(--workspace-text-muted);font-size:9px}
.purpose-options,.toggle-row{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:10px}.purpose-options label,.toggle-row label{display:flex;align-items:flex-start;gap:9px;padding:11px;border:1px solid var(--workspace-border);border-radius:var(--workspace-radius-sm);background:var(--workspace-surface-muted)}.purpose-options b,.purpose-options small,.toggle-row b,.toggle-row small{display:block}.purpose-options b,.toggle-row b{font-size:11px}.purpose-options small,.toggle-row small{margin-top:3px;color:var(--workspace-text-muted);font-size:9px;line-height:1.5}
.ai-form-section--advanced{border:0}.disclosure-button{display:flex;align-items:center;justify-content:center;gap:6px;width:100%;padding:10px;border:0;color:var(--workspace-accent);background:transparent;font-size:11px;cursor:pointer}.disclosure-button:hover{background:var(--workspace-accent-soft)}
.research-preferences{border-radius:var(--workspace-radius-md);overflow:hidden}.research-preferences__header{display:flex;align-items:center;justify-content:space-between;width:100%;padding:14px 18px;border:0;color:var(--workspace-text);background:transparent;text-align:left;cursor:pointer}.research-preferences__header>span{display:flex;align-items:center;gap:8px}.research-preferences__header b{font-size:12px}.research-preferences__header small{color:var(--workspace-text-muted);font-size:10px}.research-preferences__header em{color:var(--workspace-accent);font-size:10px;font-style:normal}.research-preferences__body{padding:0 18px 18px}.research-preferences__body .form-grid{margin-top:12px}.prompt-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:13px;margin-top:13px}
.ai-configuration__footer{position:sticky;bottom:0;z-index:4;display:flex;align-items:center;justify-content:space-between;gap:16px;padding:10px 14px;border-radius:var(--workspace-radius-md);backdrop-filter:var(--workspace-glass)}.ai-configuration__footer>div{display:flex;align-items:center;gap:7px;color:var(--workspace-text-muted);font-size:10px}.ai-configuration__footer>div .n-icon{color:var(--workspace-accent)}
@media (max-width:1100px){.ai-configuration__status{align-items:flex-start;flex-direction:column}.ai-status-metrics{width:100%;min-width:0}.ai-status-metrics span:first-child{border-left:0}.ai-configuration__layout{grid-template-columns:240px minmax(0,1fr)}.form-grid--three{grid-template-columns:repeat(2,minmax(0,1fr))}}
@media (max-width:720px){.ai-configuration{max-height:none}.ai-status-metrics{grid-template-columns:repeat(2,minmax(0,1fr));gap:10px}.ai-status-metrics span{padding:0 8px}.ai-configuration__layout{grid-template-columns:1fr}.ai-service-list{position:static;max-height:none}.form-grid--two,.form-grid--three,.prompt-grid,.purpose-options,.toggle-row{grid-template-columns:1fr}.model-row{grid-template-columns:1fr}.ai-configuration__footer{align-items:stretch;flex-direction:column}.ai-configuration__footer .n-button{width:100%}}
</style>
