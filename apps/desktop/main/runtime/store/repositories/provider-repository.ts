import { sortedValues } from '../../shared/util'
import type { JsonRecord } from '../../types'
import type { WorkspaceStore } from '../workspace-store'

export class ProviderRepository {
  constructor(private readonly store: WorkspaceStore) {}

  nextId(): string {
    return this.store.nextId('provider')
  }

  list(): JsonRecord[] {
    return sortedValues(this.store.snapshot.providers, 'providerId')
  }

  get(providerId: string): JsonRecord | undefined {
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
}
