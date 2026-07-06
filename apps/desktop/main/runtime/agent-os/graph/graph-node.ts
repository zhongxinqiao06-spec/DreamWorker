import type { JsonRecord } from '../../types'

export type GraphNode = {
  id: string
  kind: string
  label: string
  data: JsonRecord
}
