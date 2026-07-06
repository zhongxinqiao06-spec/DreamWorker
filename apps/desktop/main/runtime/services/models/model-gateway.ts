import type { JsonRecord } from '../../types'
import type { ProviderService } from './provider-service'

export class ModelGateway {
  constructor(private readonly providers: ProviderService) {}

  resolveProviderForCoding(providerId: string, model: string): JsonRecord {
    return this.providers.providerForCoding(providerId, model)
  }
}
