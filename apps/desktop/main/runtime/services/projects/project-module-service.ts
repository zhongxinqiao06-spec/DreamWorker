import { notFound } from '../../kernel/errors'
import { asRecord, asString } from '../../shared/util'
import type { JsonRecord } from '../../types'
import type { ProjectModuleRepository } from '../../store/repositories/project-module-repository'
import type { ProjectService } from './project-service'

export class ProjectModuleService {
  constructor(
    private readonly projects: ProjectService,
    private readonly modules: ProjectModuleRepository
  ) {}

  listProjectModules(projectId: string): JsonRecord[] {
    this.projects.requireProject(projectId)
    return this.modules.list(projectId)
  }

  getProjectModule(input: JsonRecord): JsonRecord {
    const projectId = asString(input.projectId)
    const moduleId = asString(input.moduleId)
    this.projects.requireProject(projectId)
    const module = this.modules.get(projectId, moduleId)
    if (!module) {
      throw notFound('MODULE_NOT_FOUND', 'module not found', 'select another module')
    }
    return module
  }

  updateProjectModuleConfig(input: JsonRecord): JsonRecord {
    const projectId = asString(input.projectId)
    const moduleId = asString(input.moduleId)
    const module = this.getProjectModule({ projectId, moduleId })
    const updated = { ...module, config: { ...asRecord(module.config), ...asRecord(input.config) } }
    this.modules.save(projectId, moduleId, updated)
    return updated
  }
}
