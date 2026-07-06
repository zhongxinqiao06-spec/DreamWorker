import type { JsonRecord } from '../../types'

export type ArtifactWriteRequest = {
  projectId: string
  path: string
  content: string
  metadata?: JsonRecord
}

export class ArtifactWriter {
  write(request: ArtifactWriteRequest): JsonRecord {
    return {
      accepted: false,
      projectId: request.projectId,
      path: request.path,
      metadata: request.metadata ?? {},
      message: 'artifact writing is reserved for the Agent OS artifact service'
    }
  }
}
