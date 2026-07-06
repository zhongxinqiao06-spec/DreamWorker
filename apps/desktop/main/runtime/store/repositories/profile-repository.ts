import { sortedValues } from '../../shared/util'
import type { JsonRecord } from '../../types'
import type { WorkspaceStore } from '../workspace-store'

export class ProfileRepository {
  constructor(private readonly store: WorkspaceStore) {}

  nextId(): string {
    return this.store.nextId('profile')
  }

  list(): JsonRecord[] {
    return sortedValues(this.store.snapshot.profiles, 'profileId')
  }

  get(profileId: string): JsonRecord | undefined {
    return this.store.snapshot.profiles[profileId]
  }

  save(profileId: string, profile: JsonRecord): void {
    this.store.snapshot.profiles[profileId] = profile
    this.store.save()
  }

  delete(profileId: string): void {
    delete this.store.snapshot.profiles[profileId]
    this.store.save()
  }
}
