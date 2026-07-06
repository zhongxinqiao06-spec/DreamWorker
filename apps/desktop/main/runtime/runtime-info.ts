import { newTraceId } from './shared/util'
import type { JsonRecord } from './types'

export const engineVersion = '0.1.0-main-runtime'
export const contractSchemaVersion = '0.1'

export function runtimePing(traceId = newTraceId()): JsonRecord {
  return {
    schema_version: contractSchemaVersion,
    ok: true,
    engineVersion,
    trace_id: traceId,
    runtime: 'desktop-main-runtime'
  }
}
