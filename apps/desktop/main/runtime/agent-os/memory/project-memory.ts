import type { JsonRecord } from '../../types'

export type ProjectMemoryRecord = {
  projectId: string
  summary: string
  facts: JsonRecord[]
  updatedAt: string
}
