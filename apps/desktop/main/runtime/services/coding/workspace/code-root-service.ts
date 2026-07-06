import type { ProjectDirectoryService } from '../../projects/project-directory-service'

export class CodeRootService {
  constructor(private readonly projectDirectory: ProjectDirectoryService) {}

  resolve(projectId: string): string {
    return this.projectDirectory.codeRoot(projectId)
  }
}
