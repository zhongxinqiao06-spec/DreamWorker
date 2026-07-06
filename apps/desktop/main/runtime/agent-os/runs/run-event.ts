import type { JsonRecord } from '../../types'
import type { AgentRunStatus } from './run-state'

export type AgentRunEvent = {
  eventId: string
  runId: string
  type: string
  status?: AgentRunStatus
  payload?: JsonRecord
  traceId: string
  createdAt: string
}
