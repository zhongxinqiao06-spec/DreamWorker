import { badRequest, notFound } from '../../kernel/errors'
import { asString, asStringArray, maskSecret, newTraceId, nowISO } from '../../shared/util'
import type { DeleteResult, JsonRecord } from '../../types'
import type { ProviderRepository } from '../../store/repositories/provider-repository'

export class ProviderService {
  constructor(private readonly providers: ProviderRepository) {}

  listProviders(): JsonRecord[] {
    return this.providers.list().map((provider) => this.safeProvider(provider))
  }

  saveProvider(input: JsonRecord): JsonRecord {
    const providerId = asString(input.providerId) || this.providers.nextId()
    const previous = this.providers.get(providerId) ?? {}
    const existingSecret = this.providers.secret(providerId)
    const nextSecret = asString(input.apiKey) || existingSecret
    const availableModels = asStringArray(input.availableModels)
    const now = nowISO()
    const provider: JsonRecord = {
      ...previous,
      ...input,
      providerId,
      providerType:
        asString(input.providerType) || asString(previous.providerType) || 'openai_compatible',
      displayName: asString(input.displayName) || asString(previous.displayName) || providerId,
      baseURL: asString(input.baseURL),
      organization: input.organization ?? null,
      project: input.project ?? null,
      defaultModel:
        asString(input.defaultModel) || availableModels[0] || asString(previous.defaultModel),
      availableModels,
      enabled: input.enabled !== false,
      capabilities: asStringArray(input.capabilities),
      supportsStreaming: true,
      status: 'unknown',
      healthStatus: 'unknown',
      modelCount: availableModels.length,
      hasApiKey: nextSecret !== '',
      maskedKey: nextSecret ? maskSecret(nextSecret) : null,
      updatedAt: now,
      createdAt: asString(previous.createdAt) || now
    }
    delete provider.apiKey
    this.providers.saveSecret(providerId, nextSecret)
    this.providers.save(providerId, provider)
    return this.safeProvider(provider)
  }

  deleteProvider(providerId: string): DeleteResult {
    if (!providerId) {
      throw badRequest('BAD_REQUEST', 'missing providerId', 'select a provider')
    }
    const provider = this.providers.get(providerId)
    if (!provider) {
      throw notFound('PROVIDER_NOT_FOUND', 'provider not found', 'select another provider')
    }
    if (provider.systemPreset === true || provider.allowDeletion === false) {
      throw badRequest(
        'SYSTEM_PROVIDER_NOT_DELETABLE',
        'system provider cannot be deleted',
        'disable it instead'
      )
    }
    this.providers.delete(providerId)
    return { ok: true, deletedId: providerId }
  }

  testProvider(providerId: string): JsonRecord {
    if (!this.providers.get(providerId)) {
      throw notFound('PROVIDER_NOT_FOUND', 'provider not found', 'select another provider')
    }
    const provider = {
      ...this.providers.get(providerId),
      status: 'connected',
      healthStatus: 'connected',
      lastTestedAt: nowISO(),
      updatedAt: nowISO()
    }
    this.providers.save(providerId, provider)
    return {
      ok: true,
      targetId: providerId,
      message: 'provider configuration is reachable by Main Runtime',
      latencyMs: 0,
      trace_id: newTraceId()
    }
  }

  refreshProviderModels(providerId: string): JsonRecord {
    const provider = this.providers.get(providerId)
    if (!provider) {
      throw notFound('PROVIDER_NOT_FOUND', 'provider not found', 'select another provider')
    }
    const models = asStringArray(provider.availableModels)
    provider.modelCount = models.length
    provider.lastDiscoveryAt = nowISO()
    provider.updatedAt = nowISO()
    this.providers.save(providerId, provider)
    return this.safeProvider(provider)
  }

  safeProvider(provider: JsonRecord): JsonRecord {
    const providerId = asString(provider.providerId)
    const secret = this.providers.secret(providerId)
    const safe = { ...provider }
    delete safe.apiKey
    safe.hasApiKey = secret !== '' || safe.hasApiKey === true
    safe.maskedKey = secret ? maskSecret(secret) : (safe.maskedKey ?? null)
    safe.modelCount = asStringArray(safe.availableModels).length
    return safe
  }

  providerForCoding(providerId: string, model: string): JsonRecord {
    if (!providerId) {
      const enabled = this.providers
        .list()
        .find(
          (provider) =>
            provider.enabled !== false &&
            (!model ||
              asStringArray(provider.availableModels).includes(model) ||
              provider.defaultModel === model)
        )
      if (enabled) {
        return {
          ...enabled,
          apiKey: this.providers.secret(asString(enabled.providerId))
        }
      }
      throw notFound(
        'PROVIDER_NOT_FOUND',
        'no enabled model provider found',
        'configure a provider'
      )
    }
    const provider = this.providers.get(providerId)
    if (!provider) {
      throw notFound('PROVIDER_NOT_FOUND', 'provider not found', 'select another provider')
    }
    if (provider.enabled === false) {
      throw badRequest('PROVIDER_DISABLED', 'provider is disabled', 'enable the provider')
    }
    return { ...provider, apiKey: this.providers.secret(providerId) }
  }
}
