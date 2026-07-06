import { mkdirSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'
import { asString, asStringArray, newTraceId, nowISO } from '../../shared/util'
import type { JsonRecord } from '../../types'
import type { WorkspaceStore } from '../../store/workspace-store'
import type { ProjectDirectoryService } from '../projects/project-directory-service'
import type { ProjectService } from '../projects/project-service'

export class RequirementAnalysisService {
  constructor(
    private readonly store: WorkspaceStore,
    private readonly projects: ProjectService,
    private readonly projectDirectory: ProjectDirectoryService
  ) {}

  runRequirementAnalysis(input: JsonRecord): JsonRecord {
    const project = this.projects.getProject(asString(input.projectId))
    const projectTitle = asString(project.title)
    const runId = this.store.nextId('requirements_run')
    const root = this.projectDirectory.root(asString(project.projectId))
    const outputDir = join(root, 'artifacts', 'product')
    mkdirSync(outputDir, { recursive: true })
    const analysis = {
      projectTitle,
      summary: asString(input.prompt) || 'Main Runtime generated requirement analysis placeholder.',
      sources: asStringArray(input.sourceIds),
      roles: ['用户'],
      features: [],
      nonFunctionalRequirements: [],
      risks: [],
      openQuestions: []
    }
    const analysisPath = join(outputDir, 'requirements_analysis.json')
    writeFileSync(analysisPath, `${JSON.stringify(analysis, null, 2)}\n`)
    return {
      runId,
      projectId: asString(project.projectId),
      status: 'completed',
      sources: [],
      featureCount: 0,
      outputFiles: [
        {
          kind: 'analysis_json',
          fileName: 'requirements_analysis.json',
          relativePath: 'artifacts/product/requirements_analysis.json',
          absolutePath: analysisPath
        }
      ],
      warnings: ['Main Runtime has not yet connected the document parser pipeline.'],
      traceId: newTraceId(),
      createdAt: nowISO(),
      analysis
    }
  }
}
