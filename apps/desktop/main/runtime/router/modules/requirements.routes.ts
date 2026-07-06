import { asString } from '../../shared/util'
import type { RuntimeContext } from '../../bootstrap/runtime-context'
import { post, type RuntimeRoute } from '../route'

export function requirementRoutes(context: RuntimeContext): RuntimeRoute[] {
  const { store } = context
  return [
    post('/projects/requirements/import-files', (body) => store.importRequirementFiles(body)),
    post('/projects/requirements/sources', (body) =>
      store.listRequirementSources(asString(body.projectId))
    ),
    post('/projects/requirements/preview-source', (body) => store.previewRequirementSource(body)),
    post('/projects/requirements/run', (body) => store.runRequirementAnalysis(body))
  ]
}
