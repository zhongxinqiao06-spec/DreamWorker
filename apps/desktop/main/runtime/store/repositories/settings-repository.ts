import type { JsonRecord } from '../../types'
import { defaultSettings } from '../defaults'
import type { WorkspaceStore } from '../workspace-store'

export class SettingsRepository {
  constructor(private readonly store: WorkspaceStore) {}

  get(): JsonRecord {
    this.store.snapshot.settings = {
      ...defaultSettings(),
      ...(this.store.snapshot.settings ?? {})
    }
    return this.store.snapshot.settings
  }

  save(settings: JsonRecord): void {
    this.store.snapshot.settings = settings
    this.store.save()
  }

  reset(): JsonRecord {
    const settings = defaultSettings()
    this.save(settings)
    return settings
  }
}
