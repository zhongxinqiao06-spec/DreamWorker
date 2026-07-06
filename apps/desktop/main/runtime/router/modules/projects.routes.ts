import { asString } from '../../shared/util'
import type { RuntimeContext } from '../../bootstrap/runtime-context'
import { get, post, type RuntimeRoute } from '../route'

export function projectRoutes(context: RuntimeContext): RuntimeRoute[] {
  const { store } = context
  return [
    get('/projects', () => store.listProjects()),
    post('/projects/create', (body) => store.createProject(body)),
    post('/projects/get', (body) => store.getProject(asString(body.projectId))),
    post('/projects/update', (body) => store.updateProject(body)),
    post('/projects/delete', (body) => store.deleteProject(asString(body.projectId))),
    post('/projects/local-directory/validate', (body) =>
      store.validateLocalDirectory(asString(body.projectId))
    ),
    post('/projects/local-directory/initialize', (body) =>
      store.initializeLocalDirectory(asString(body.projectId))
    ),
    post('/projects/export-manifest', (body) =>
      store.exportProjectManifest(asString(body.projectId))
    ),
    post('/projects/modules', (body) => store.listProjectModules(asString(body.projectId))),
    post('/projects/modules/get', (body) => store.getProjectModule(body)),
    post('/projects/modules/update-config', (body) => store.updateProjectModuleConfig(body))
  ]
}
