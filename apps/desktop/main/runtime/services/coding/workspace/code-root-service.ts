import type { WorkspaceStore } from '../../../store/workspace-store'

export class CodeRootService {
  constructor(private readonly store: WorkspaceStore) {}

  resolve(projectId: string): string {
    return this.store.projectCodeRoot(projectId)
  }
}
