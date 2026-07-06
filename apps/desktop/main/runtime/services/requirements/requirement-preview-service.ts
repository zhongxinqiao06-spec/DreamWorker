import { readFileSync } from 'node:fs'
import { notFound } from '../../kernel/errors'
import { asRecord, asString, newTraceId, nowISO } from '../../shared/util'
import type { JsonRecord } from '../../types'
import type { RequirementImportService } from './requirement-import-service'

export class RequirementPreviewService {
  constructor(private readonly imports: RequirementImportService) {}

  previewRequirementSource(input: JsonRecord): JsonRecord {
    const projectId = asString(input.projectId)
    const sources = this.imports.listRequirementSources(projectId).sources
    const source = Array.isArray(sources)
      ? (sources.find((item) => asRecord(item).sourceId === input.sourceId) as
          JsonRecord | undefined)
      : undefined
    if (!source) {
      throw notFound(
        'REQUIREMENT_SOURCE_NOT_FOUND',
        'requirement source not found',
        'select another source'
      )
    }
    const absolutePath = asString(source.absolutePath)
    const raw = readFileSync(absolutePath)
    const text = raw.toString('utf8')
    return {
      projectId,
      source,
      parser: 'node-inline',
      content: text.slice(0, 20000),
      charCount: text.length,
      truncated: text.length > 20000,
      traceId: newTraceId(),
      createdAt: nowISO()
    }
  }
}
