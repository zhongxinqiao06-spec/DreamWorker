import { asString } from '../../../shared/util'
import type { JsonRecord } from '../../../types'
import type { CodeRootService } from './code-root-service'
import type { FileReadService } from './file-read-service'
import type { FileTreeService } from './file-tree-service'
import type { GitStatusService } from './git-status-service'

export class CodeWorkspaceService {
  constructor(
    private readonly roots: CodeRootService,
    private readonly files: FileTreeService,
    private readonly reader: FileReadService,
    private readonly git: GitStatusService
  ) {}

  listFiles(input: JsonRecord): JsonRecord[] {
    return this.files.list(input)
  }

  readFile(input: JsonRecord): JsonRecord {
    return this.reader.read(input)
  }

  fileStatus(input: JsonRecord): JsonRecord {
    const projectId = asString(input.projectId)
    const root = this.roots.resolve(projectId)
    const changes = this.git.changes(root)
    const branch = this.git.branch(root)
    return {
      projectId,
      branch,
      changes,
      clean: changes.length === 0,
      message:
        branch || changes.length > 0
          ? 'git status ready'
          : 'not a git repository or no git executable available'
    }
  }
}
