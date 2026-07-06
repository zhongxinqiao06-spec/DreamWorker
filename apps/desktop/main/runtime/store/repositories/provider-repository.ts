import { sortedValues } from '../../shared/util'
import type { JsonRecord } from '../../types'
import type { WorkspaceStore } from '../workspace-store'

export class ProviderRepository {
  constructor(private readonly store: WorkspaceStore) {}

  nextId(): string {
    return this.store.nextId('provider')
  }

  list(): JsonRecord[] {
    this.migrateLegacyNineRouterProvider()
    return sortedValues(this.store.snapshot.providers, 'providerId')
  }

  get(providerId: string): JsonRecord | undefined {
    this.migrateLegacyNineRouterProvider()
    return this.store.snapshot.providers[providerId]
  }

  save(providerId: string, provider: JsonRecord): void {
    this.store.snapshot.providers[providerId] = provider
    this.store.save()
  }

  delete(providerId: string): void {
    delete this.store.snapshot.providers[providerId]
    delete this.store.snapshot.providerSecrets[providerId]
    this.store.save()
  }

  secret(providerId: string): string {
    return this.store.snapshot.providerSecrets[providerId] ?? ''
  }

  saveSecret(providerId: string, secret: string): void {
    if (secret) {
      this.store.snapshot.providerSecrets[providerId] = secret
    }
  }

  private migrateLegacyNineRouterProvider(): void {
    const provider = this.store.snapshot.providers.provider_9router_local
    if (
      provider &&
      (provider.baseURL === 'http://127.0.0.1:9399/v1' ||
        provider.baseURL === 'http://localhost:9399/v1')
    ) {
      this.store.snapshot.providers.provider_9router_local = {
        ...provider,
        baseURL: 'http://127.0.0.1:20128/v1',
        updatedAt: new Date().toISOString()
      }
      this.store.save()
    }
  }
}
