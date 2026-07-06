import type { JsonRecord } from '../../types'

export type DefaultProviderSeed = {
  readonly providers: Record<string, JsonRecord>
  readonly providerSecrets: Record<string, string>
  readonly deepseekModel: string
  readonly deepseekProModel: string
}

export function createDefaultProviders(timestamp: string): DefaultProviderSeed {
  const deepseekModel =
    process.env.DEEPSEEK_FAST_MODEL || process.env.DEEPSEEK_MODEL || 'deepseek-v4-flash'
  const deepseekProModel = process.env.DEEPSEEK_PRO_MODEL || 'deepseek-v4-pro'
  const deepseekKey = process.env.DEEPSEEK_API_KEY || ''

  return {
    deepseekModel,
    deepseekProModel,
    providerSecrets: deepseekKey ? { provider_deepseek: deepseekKey } : {},
    providers: {
      provider_deepseek: {
        providerId: 'provider_deepseek',
        providerType: 'deepseek',
        displayName: 'DeepSeek 兼容服务',
        baseURL: process.env.DEEPSEEK_BASE_URL || 'https://api.deepseek.com',
        organization: null,
        project: null,
        defaultModel: deepseekModel,
        availableModels: unique([
          deepseekModel,
          deepseekProModel,
          'deepseek-v4-flash',
          'deepseek-v4-pro',
          'deepseek-chat',
          'deepseek-reasoner'
        ]),
        enabled: true,
        status: 'unknown',
        capabilities: ['chat', 'tools', 'json_schema'],
        supportsStreaming: true,
        healthStatus: 'unknown',
        modelCount: 6,
        latencyMs: 0,
        lastDiscoveryAt: null,
        lastStreamAt: null,
        lastErrorCode: null,
        streamingVerified: false,
        hasApiKey: deepseekKey !== '',
        maskedKey: deepseekKey ? maskInline(deepseekKey) : null,
        lastTestedAt: null,
        lastError: null,
        createdAt: timestamp,
        updatedAt: timestamp
      },
      provider_9router_local: {
        providerId: 'provider_9router_local',
        providerType: 'openai_compatible',
        displayName: '9Router 本地模型路由',
        baseURL: 'http://127.0.0.1:9399/v1',
        organization: null,
        project: null,
        defaultModel: process.env.NINE_ROUTER_DEFAULT_MODEL || deepseekModel,
        availableModels: [process.env.NINE_ROUTER_DEFAULT_MODEL || deepseekModel],
        enabled: true,
        status: 'unknown',
        capabilities: ['chat', 'tools', 'json_schema'],
        supportsStreaming: true,
        healthStatus: 'unknown',
        modelCount: 1,
        latencyMs: 0,
        lastDiscoveryAt: null,
        lastStreamAt: null,
        lastErrorCode: null,
        streamingVerified: false,
        hasApiKey: false,
        maskedKey: null,
        lastTestedAt: null,
        lastError: null,
        createdAt: timestamp,
        updatedAt: timestamp,
        systemPreset: true,
        allowDeletion: false
      },
      provider_local_stub: {
        providerId: 'provider_local_stub',
        providerType: 'openai_compatible',
        displayName: '本地 Stub 模型',
        baseURL: 'http://127.0.0.1/model-stub',
        organization: null,
        project: null,
        defaultModel: 'model_generate_stub',
        availableModels: ['model_generate_stub'],
        enabled: true,
        status: 'connected',
        capabilities: ['chat', 'tools', 'image_generation', 'json_schema'],
        supportsStreaming: true,
        healthStatus: 'connected',
        modelCount: 1,
        latencyMs: 0,
        lastDiscoveryAt: null,
        lastStreamAt: null,
        lastErrorCode: null,
        streamingVerified: true,
        hasApiKey: false,
        maskedKey: null,
        lastTestedAt: null,
        lastError: null,
        createdAt: timestamp,
        updatedAt: timestamp
      }
    }
  }
}

function unique(values: string[]): string[] {
  return [...new Set(values.filter(Boolean))]
}

function maskInline(value: string): string {
  return value.length <= 8 ? '***' : `${value.slice(0, 4)}...${value.slice(-4)}`
}
