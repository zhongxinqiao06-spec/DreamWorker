import type { JsonRecord } from '../../types'

export class WorkingMemory {
  private readonly values = new Map<string, JsonRecord>()

  set(key: string, value: JsonRecord): void {
    this.values.set(key, value)
  }

  get(key: string): JsonRecord | undefined {
    return this.values.get(key)
  }

  snapshot(): Record<string, JsonRecord> {
    return Object.fromEntries(this.values)
  }
}
