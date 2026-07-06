export { newTraceId } from './ids'
export { asRecord, asString, asStringArray, sortedValues } from './json'
export { maskSecret, redactSecrets } from './security'

export function nowISO(): string {
  return new Date().toISOString()
}
