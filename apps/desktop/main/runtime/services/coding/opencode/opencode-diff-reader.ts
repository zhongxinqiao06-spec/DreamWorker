import { asRecord, asString } from '../../../shared/util'
import type { JsonRecord } from '../../../types'

export function normalizeDiff(payload: JsonRecord): Array<{ path: string; status: string }> {
  const data = payload.data
  const record = asRecord(data)
  let items: unknown[] = []
  if (Array.isArray(data)) {
    items = data
  } else if (Array.isArray(record.files)) {
    items = record.files
  } else if (Array.isArray(record.changes)) {
    items = record.changes
  }
  return items
    .map((item) => asRecord(item))
    .map((item) => ({
      path: asString(item.path) || asString(item.file) || asString(item.filename),
      status: asString(item.status) || asString(item.type) || 'modified'
    }))
    .filter((item) => item.path !== '')
}
