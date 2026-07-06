import type { JsonRecord } from '../../types'

export type GraphEdge = {
  id: string
  from: string
  to: string
  kind: string
  data: JsonRecord
}
