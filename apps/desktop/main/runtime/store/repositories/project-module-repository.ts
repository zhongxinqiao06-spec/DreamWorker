import { sortedValues } from '../../shared/util'
import type { JsonRecord } from '../../types'
import { createProjectModules } from '../defaults'
import type { WorkspaceStore } from '../workspace-store'

export class ProjectModuleRepository {
  constructor(private readonly store: WorkspaceStore) {}

  ensure(projectId: string): Record<string, JsonRecord> {
    if (!this.store.snapshot.modules[projectId]) {
      this.store.snapshot.modules[projectId] = createProjectModules(projectId)
      this.store.save()
    }
    return this.store.snapshot.modules[projectId] ?? {}
  }

  replace(projectId: string, modules: Record<string, JsonRecord>): void {
    this.store.snapshot.modules[projectId] = modules
    this.store.save()
  }

  list(projectId: string): JsonRecord[] {
    return sortedValues(this.ensure(projectId), 'moduleId')
  }

  get(projectId: string, moduleId: string): JsonRecord | undefined {
    return this.ensure(projectId)[moduleId]
  }

  save(projectId: string, moduleId: string, module: JsonRecord): void {
    this.ensure(projectId)[moduleId] = module
    this.store.save()
  }
}
