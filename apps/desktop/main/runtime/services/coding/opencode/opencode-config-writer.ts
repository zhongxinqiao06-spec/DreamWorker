import { mkdirSync, writeFileSync } from 'node:fs'
import { join, resolve } from 'node:path'
import { asString } from '../../../shared/util'
import type { JsonRecord } from '../../../types'

export function openCodeProviderId(provider: JsonRecord): string {
  const providerType = asString(provider.providerType)
  if (providerType === 'anthropic') {
    return 'anthropic'
  }
  if (providerType === 'openai') {
    return 'openai'
  }
  if (providerType === 'ollama') {
    return 'ollama'
  }
  return 'dreamworker'
}

export function openCodeConfig(provider: JsonRecord, model: string): JsonRecord {
  const providerId = openCodeProviderId(provider)
  const baseURL = asString(provider.baseURL)
  const apiKey = asString(provider.apiKey)
  const modelName = model || asString(provider.defaultModel)
  const config: JsonRecord = {
    autoupdate: false,
    share: 'disabled',
    model: `${providerId}/${modelName}`,
    permission: {
      edit: 'allow',
      bash: 'allow'
    },
    tools: {
      write: true,
      edit: true,
      bash: true
    },
    provider: {}
  }
  if (providerId === 'dreamworker') {
    config.provider = {
      dreamworker: {
        npm: '@ai-sdk/openai-compatible',
        name: asString(provider.displayName) || 'DreamWorker Provider',
        options: {
          baseURL,
          apiKey: apiKey ? '{env:DREAMWORKER_OPENCODE_API_KEY}' : undefined
        },
        models: {
          [modelName]: {
            name: modelName
          }
        }
      }
    }
    return config
  }
  return config
}

export function openCodeProviderEnv(provider: JsonRecord): NodeJS.ProcessEnv {
  const apiKey = asString(provider.apiKey)
  const providerType = asString(provider.providerType)
  const env: NodeJS.ProcessEnv = {
    DREAMWORKER_OPENCODE_API_KEY: apiKey
  }
  if (providerType === 'openai' || openCodeProviderId(provider) === 'dreamworker') {
    env.OPENAI_API_KEY = apiKey
  }
  if (providerType === 'anthropic') {
    env.ANTHROPIC_API_KEY = apiKey
  }
  return env
}

export function writeOpenCodeConfig(cwd: string, config: JsonRecord): string {
  const serialized = JSON.stringify(config, null, 2)
  const configDir = resolve(cwd, '..', '..')
  mkdirSync(configDir, { recursive: true })
  const configPath = join(configDir, 'opencode.json')
  writeFileSync(configPath, `${serialized}\n`, 'utf8')
  return configPath
}
