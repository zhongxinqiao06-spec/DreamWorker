import { asString } from '../../shared/util'
import type { RuntimeContext } from '../../bootstrap/runtime-context'
import { post, type RuntimeRoute } from '../route'

export function requirementRoutes(context: RuntimeContext): RuntimeRoute[] {
  return [
    post('/projects/requirements/import-files', (body) =>
      context.requirementImports.importRequirementFiles(body)
    ),
    post('/projects/requirements/sources', (body) =>
      context.requirementImports.listRequirementSources(asString(body.projectId))
    ),
    post('/projects/requirements/preview-source', (body) =>
      context.requirementPreview.previewRequirementSource(body)
    ),
    post('/projects/requirements/run', (body) =>
      context.requirementAnalysis.runRequirementAnalysis(body)
    )
  ]
}
