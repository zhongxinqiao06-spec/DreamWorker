import { asString } from '../../shared/util'
import type { RuntimeContext } from '../../bootstrap/runtime-context'
import { get, post, type RuntimeRoute } from '../route'

export function projectRoutes(context: RuntimeContext): RuntimeRoute[] {
  return [
    get('/projects', () => context.projects.listProjects()),
    post('/projects/create', (body) => context.projects.createProject(body)),
    post('/projects/get', (body) => context.projects.getProject(asString(body.projectId))),
    post('/projects/update', (body) => context.projects.updateProject(body)),
    post('/projects/delete', (body) => context.projects.deleteProject(asString(body.projectId))),
    post('/projects/local-directory/validate', (body) =>
      context.projectDirectory.validate(asString(body.projectId))
    ),
    post('/projects/local-directory/initialize', (body) =>
      context.projectDirectory.initialize(asString(body.projectId))
    ),
    post('/projects/export-manifest', (body) =>
      context.projectDirectory.exportManifest(asString(body.projectId))
    ),
    post('/projects/modules', (body) =>
      context.projectModules.listProjectModules(asString(body.projectId))
    ),
    post('/projects/modules/get', (body) => context.projectModules.getProjectModule(body)),
    post('/projects/modules/update-config', (body) =>
      context.projectModules.updateProjectModuleConfig(body)
    )
  ]
}
