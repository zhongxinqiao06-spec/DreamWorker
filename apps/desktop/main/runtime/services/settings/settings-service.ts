import type { JsonRecord } from '../../types'
import type { SettingsRepository } from '../../store/repositories/settings-repository'

export class SettingsService {
  constructor(private readonly settings: SettingsRepository) {}

  getSettings(): JsonRecord {
    return this.settings.get()
  }

  updateSettings(input: JsonRecord): JsonRecord {
    const settings = { ...this.getSettings(), ...input }
    this.settings.save(settings)
    return settings
  }

  resetExtensionSettings(): JsonRecord {
    return this.settings.reset()
  }
}
