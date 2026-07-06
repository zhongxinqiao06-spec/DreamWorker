import type { JsonRecord } from '../../types'
import { ArtifactWriter, type ArtifactWriteRequest } from './artifact-writer'

export class ArtifactService {
  constructor(private readonly writer = new ArtifactWriter()) {}

  write(request: ArtifactWriteRequest): JsonRecord {
    return this.writer.write(request)
  }
}
