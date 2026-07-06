import { randomBytes } from 'node:crypto'

export function nowISO(): string {
  return new Date().toISOString()
}

export function newTraceId(): string {
  return `tr_${new Date().toISOString().replace(/[-:.]/g, '').slice(0, 17)}_${randomBytes(4).toString('hex')}`
}

export function asRecord(value: unknown): Record<string, unknown> {
  if (typeof value === 'object' && value !== null && !Array.isArray(value)) {
    return value as Record<string, unknown>
  }
  return {}
}

export function asString(value: unknown): string {
  return typeof value === 'string' ? value : ''
}

export function asStringArray(value: unknown): string[] {
  if (!Array.isArray(value)) {
    return []
  }
  return value.filter((item): item is string => typeof item === 'string')
}

export function maskSecret(value: string): string {
  if (!value) {
    return ''
  }
  if (value.length <= 8) {
    return '***'
  }
  return `${value.slice(0, 4)}...${value.slice(-4)}`
}

export function sortedValues<T extends Record<string, unknown>>(
  items: Record<string, T>,
  key: keyof T
): T[] {
  return Object.values(items).sort((left, right) =>
    String(left[key] ?? '').localeCompare(String(right[key] ?? ''))
  )
}

export function redactSecrets(message: string): string {
  return message.replace(/(sk-[A-Za-z0-9_-]{8,}|[A-Za-z0-9_-]{32,})/g, '***')
}
