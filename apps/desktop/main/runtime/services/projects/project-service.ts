import { badRequest, notFound } from '../../kernel/errors'
import { asString, nowISO } from '../../shared/util'
import type { DeleteResult, JsonRecord } from '../../types'
import { createProject, createProjectModules, touch } from '../../store/defaults'
import type { ProjectModuleRepository } from '../../store/repositories/project-module-repository'
import type { ProjectRepository } from '../../store/repositories/project-repository'
import type { ProjectDirectoryService } from './project-directory-service'

export class ProjectService {
  constructor(
    private readonly projects: ProjectRepository,
    private readonly modules: ProjectModuleRepository,
    private readonly directories: ProjectDirectoryService
  ) {}

  listProjects(): JsonRecord[] {
    return this.projects.list()
  }

  createProject(input: JsonRecord): JsonRecord {
    const now = nowISO()
    const projectId = this.projects.nextId()
    const project = createProject(projectId, now, {
      title: asString(input.title) || '新项目',
      description: asString(input.description),
      localRootPath: typeof input.localRootPath === 'string' ? input.localRootPath : null
    })
    this.projects.save(projectId, project)
    this.modules.replace(projectId, createProjectModules(projectId))
    if (project.localRootPath) {
      this.directories.initialize(projectId)
    }
    return this.projects.get(projectId) ?? project
  }

  getProject(projectId: string): JsonRecord {
    const project = this.projects.get(projectId)
    if (!project) {
      throw notFound('PROJECT_NOT_FOUND', 'project not found', 'project not found')
    }
    return project
  }

  updateProject(input: JsonRecord): JsonRecord {
    const projectId = asString(input.projectId)
    const previous = this.getProject(projectId)
    const project = touch({ ...previous, ...input, projectId })
    this.projects.save(projectId, project)
    this.modules.ensure(projectId)
    return project
  }

  deleteProject(projectId: string): DeleteResult {
    if (!this.projects.get(projectId)) {
      throw notFound('PROJECT_NOT_FOUND', 'project not found', 'refresh project list')
    }
    this.projects.delete(projectId)
    return { ok: true, deletedId: projectId }
  }

  requireProject(projectId: string): void {
    if (!projectId) {
      throw badRequest('BAD_REQUEST', 'missing projectId', 'select a project')
    }
    this.getProject(projectId)
  }
}
