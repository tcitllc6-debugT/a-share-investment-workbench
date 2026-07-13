export const AI_PROVIDER_PRESETS = [
  { key: 'deepseek', name: 'DeepSeek', baseUrl: 'https://api.deepseek.com', defaultModel: 'deepseek-chat', requiresKey: true, description: '适合个股研究与深度推理' },
  { key: 'siliconflow', name: '硅基流动', baseUrl: 'https://api.siliconflow.cn/v1', defaultModel: '', requiresKey: true, description: '一个密钥使用多种开源模型' },
  { key: 'bigmodel', name: '智谱 AI', baseUrl: 'https://open.bigmodel.cn/api/paas/v4', defaultModel: 'glm-4-flash', requiresKey: true, description: 'GLM 系列中文模型' },
  { key: 'volcengine', name: '豆包 / 火山引擎', baseUrl: 'https://ark.cn-beijing.volces.com/api/v3', defaultModel: '', requiresKey: true, description: '填写推理接入点名称作为模型' },
  { key: 'dashscope', name: '通义千问 / 百炼', baseUrl: 'https://dashscope.aliyuncs.com/compatible-mode/v1', defaultModel: 'qwen-plus', requiresKey: true, description: '通义千问兼容接口' },
  { key: 'moonshot', name: 'Kimi / 月之暗面', baseUrl: 'https://api.moonshot.cn/v1', defaultModel: 'moonshot-v1-32k', requiresKey: true, description: '长文本理解能力' },
  { key: 'openai', name: 'OpenAI', baseUrl: 'https://api.openai.com/v1', defaultModel: 'gpt-4o-mini', requiresKey: true, description: 'OpenAI 官方接口' },
  { key: 'openrouter', name: 'OpenRouter', baseUrl: 'https://openrouter.ai/api/v1', defaultModel: '', requiresKey: true, description: '统一接入多个海外模型' },
  { key: 'ollama', name: '本机 Ollama', baseUrl: 'http://127.0.0.1:11434/v1', defaultModel: 'qwen3', requiresKey: false, description: '模型与数据都在本机运行' },
  { key: 'custom', name: '自定义兼容接口', baseUrl: '', defaultModel: '', requiresKey: false, description: '适用于 LM Studio 等 OpenAI 兼容服务' },
]

export function providerForService(service = {}) {
  const baseUrl = String(service.baseUrl || '').replace(/\/$/, '')
  return AI_PROVIDER_PRESETS.find((item) => item.baseUrl && item.baseUrl === baseUrl)
    || (baseUrl.includes('127.0.0.1:11434') || baseUrl.includes('localhost:11434')
      ? AI_PROVIDER_PRESETS.find((item) => item.key === 'ollama')
      : AI_PROVIDER_PRESETS.find((item) => item.key === 'custom'))
}

export function createAIService(providerKey = 'deepseek') {
  const provider = AI_PROVIDER_PRESETS.find((item) => item.key === providerKey) || AI_PROVIDER_PRESETS[0]
  return {
    ID: 0,
    name: provider.name,
    providerKey: provider.key,
    baseUrl: provider.baseUrl,
    apiKey: '',
    hasApiKey: false,
    modelName: provider.defaultModel,
    maxTokens: 8192,
    temperature: 0.2,
    timeOut: 300,
    thinking: false,
    httpProxyEnabled: false,
    httpProxy: '',
    isDefaultResearch: false,
    isDefaultAgent: false,
    _modelOptions: provider.defaultModel ? [{ label: provider.defaultModel, value: provider.defaultModel }] : [],
    _connection: null,
  }
}

function normaliseEndpoint(value) {
  return String(value || '').trim().replace(/\/+$/, '')
}

export function endpointChanged(service = {}) {
  const original = normaliseEndpoint(service._originalBaseUrl)
  return Boolean(service.ID && service.hasApiKey && original && normaliseEndpoint(service.baseUrl) !== original)
}

export function validateAIService(service = {}, { requireModel = true } = {}) {
  const errors = {}
  const baseUrl = String(service.baseUrl || '').trim()
  const provider = providerForService(service)
  if (!baseUrl) {
    errors.baseUrl = '请选择服务商或填写接口地址'
  } else {
    try {
      const parsed = new URL(baseUrl)
      if (!['http:', 'https:'].includes(parsed.protocol)) throw new Error('protocol')
    } catch {
      errors.baseUrl = '接口地址格式不正确'
    }
  }
  const hasNewKey = Boolean(String(service.apiKey || '').trim())
  if (endpointChanged(service) && provider?.key !== 'ollama' && !hasNewKey) {
    errors.apiKey = '接口地址已更改，请重新输入 API 密钥'
  } else if (provider?.requiresKey && !service.hasApiKey && !hasNewKey) {
    errors.apiKey = '请输入服务商提供的 API 密钥'
  }
  if (requireModel && !String(service.modelName || '').trim()) {
    errors.modelName = '请选择或输入模型名称'
  }
  return errors
}

export function connectionResultTone(result) {
  if (result?.success) return 'success'
  if (result?.reachable && result?.manualModelAllowed) return 'warning'
  if (result?.code === 'invalid_key') return 'error'
  return 'error'
}

export function displayServiceName(service = {}) {
  return String(service.name || '').trim() || providerForService(service)?.name || '未命名模型服务'
}

export function preferredAIService(services = [], purpose = 'research') {
  if (!Array.isArray(services) || !services.length) return null
  const field = purpose === 'agent' ? 'isDefaultAgent' : 'isDefaultResearch'
  return services.find((service) => service?.[field]) || services.find(Boolean) || null
}
