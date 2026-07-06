import type { JsonRecord } from '../../types'

export function createProfile(
  profileId: string,
  displayName: string,
  providerId: string,
  model: string,
  timestamp: string,
  overrides: JsonRecord = {}
): JsonRecord {
  return {
    profileId,
    displayName,
    providerId,
    model,
    temperature: 0.4,
    maxTokens: 4096,
    contextWindow: 128000,
    responseFormat: 'text',
    toolMode: 'auto',
    fallbackProfileId: 'profile_stub',
    timeoutMs: 120000,
    purpose: '默认模型配置',
    enabled: true,
    createdAt: timestamp,
    updatedAt: timestamp,
    ...overrides
  }
}
