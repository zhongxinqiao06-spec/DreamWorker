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

export function sortedValues<T extends Record<string, unknown>>(
  items: Record<string, T>,
  key: keyof T
): T[] {
  return Object.values(items).sort((left, right) =>
    String(left[key] ?? '').localeCompare(String(right[key] ?? ''))
  )
}
