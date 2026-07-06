import type { JsonRecord } from '../../types'
import { defaultSettings } from '../defaults'
import type { WorkspaceStore } from '../workspace-store'

export class SettingsRepository {
  constructor(private readonly store: WorkspaceStore) {}

  get(): JsonRecord {
    const result = migrateSettings({
      ...defaultSettings(),
      ...(this.store.snapshot.settings ?? {})
    })
    this.store.snapshot.settings = result.settings
    if (result.changed) {
      this.store.save()
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

function migrateSettings(settings: JsonRecord): { settings: JsonRecord; changed: boolean } {
  let changed = false
  const next = { ...settings }
  if (
    next.nineRouterBaseURL === 'http://127.0.0.1:9399/v1' ||
    next.nineRouterBaseURL === 'http://localhost:9399/v1'
  ) {
    next.nineRouterBaseURL = 'http://127.0.0.1:20128/v1'
    changed = true
  }
  if (
    next.nineRouterDashboardURL === 'http://127.0.0.1:9399' ||
    next.nineRouterDashboardURL === 'http://localhost:9399'
  ) {
    next.nineRouterDashboardURL = 'http://127.0.0.1:20128'
    changed = true
  }
  if (next.nineRouterManagedPackageName === '@9router/cli') {
    next.nineRouterManagedPackageName = '9router'
    changed = true
  }
  if (next.nineRouterManagedInstallVersion === 'latest') {
    next.nineRouterManagedInstallVersion = '0.5.18'
    changed = true
  }
  return { settings: changed ? next : settings, changed }
}
