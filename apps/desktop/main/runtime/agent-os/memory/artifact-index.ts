import type { JsonRecord } from '../../types'

export type ArtifactIndexEntry = {
  artifactId: string
  projectId: string
  path: string
  kind: string
  metadata: JsonRecord
  updatedAt: string
}
