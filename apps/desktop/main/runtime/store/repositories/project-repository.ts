import { sortedValues } from '../../shared/util'
import type { JsonRecord } from '../../types'
import type { WorkspaceStore } from '../workspace-store'

export class ProjectRepository {
  constructor(private readonly store: WorkspaceStore) {}

  nextId(): string {
    return this.store.nextId('project')
  }

  list(): JsonRecord[] {
    return sortedValues(this.store.snapshot.projects, 'projectId')
  }

  get(projectId: string): JsonRecord | undefined {
    return this.store.snapshot.projects[projectId]
  }

  save(projectId: string, project: JsonRecord): void {
    this.store.snapshot.projects[projectId] = project
    this.store.save()
  }

  delete(projectId: string): void {
    delete this.store.snapshot.projects[projectId]
    delete this.store.snapshot.modules[projectId]
    this.store.save()
  }
}
