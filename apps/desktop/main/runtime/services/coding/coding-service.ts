import type { WorkspaceStore } from '../../store/workspace-store'
import type { CodingSession, CodingStreamEvent, DeleteResult, JsonRecord } from '../../types'
import { CodeRootService } from './workspace/code-root-service'
import { CodeWorkspaceService } from './workspace/code-workspace-service'
import { FileReadService } from './workspace/file-read-service'
import { FileTreeService } from './workspace/file-tree-service'
import { GitStatusService } from './workspace/git-status-service'
import { CodingSessionService } from './coding-session-service'
import { CodingStreamService } from './coding-stream-service'
import { CodingEngineRegistry } from './engines/engine-registry'

export class CodingService {
  private readonly sessions: CodingSessionService
  private readonly engines: CodingEngineRegistry
  private readonly files: CodeWorkspaceService
  private readonly streams: CodingStreamService

  constructor(store: WorkspaceStore)
  constructor(
    sessions: CodingSessionService,
    engines: CodingEngineRegistry,
    files: CodeWorkspaceService,
    streams: CodingStreamService
  )
  constructor(
    first: WorkspaceStore | CodingSessionService,
    engines?: CodingEngineRegistry,
    files?: CodeWorkspaceService,
    streams?: CodingStreamService
  ) {
    if (first instanceof CodingSessionService) {
      if (!engines || !files || !streams) {
        throw new Error('CodingService requires sessions, engines, files and streams')
      }
      this.sessions = first
      this.engines = engines
      this.files = files
      this.streams = streams
      return
    }

    const store = first
    const roots = new CodeRootService(store)
    const git = new GitStatusService()
    const tree = new FileTreeService(roots, git)
    const reader = new FileReadService(roots)
    this.sessions = new CodingSessionService(store)
    this.engines = new CodingEngineRegistry()
    this.files = new CodeWorkspaceService(roots, tree, reader, git)
    this.streams = new CodingStreamService(store, this.sessions, this.engines)
  }

  dispose(): void {
    this.streams.dispose()
    this.engines.dispose()
  }

  listEngines(): Promise<JsonRecord> {
    return this.engines.list()
  }

  createSession(input: JsonRecord): CodingSession {
    return this.sessions.create(input)
  }

  getSession(input: JsonRecord): CodingSession {
    return this.sessions.get(input)
  }

  cancelTurn(input: JsonRecord): DeleteResult {
    return this.streams.cancelTurn(input)
  }

  streamTurn(input: JsonRecord): AsyncGenerator<CodingStreamEvent> {
    return this.streams.streamTurn(input)
  }

  listFiles(input: JsonRecord): JsonRecord[] {
    return this.files.listFiles(input)
  }

  readFile(input: JsonRecord): JsonRecord {
    return this.files.readFile(input)
  }

  fileStatus(input: JsonRecord): JsonRecord {
    return this.files.fileStatus(input)
  }
}
