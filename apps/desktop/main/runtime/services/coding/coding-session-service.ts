import { badRequest, notFound } from '../../kernel/errors'
import { asString, nowISO } from '../../shared/util'
import type { CodingSession, JsonRecord } from '../../types'
import type { WorkspaceStore } from '../../store/workspace-store'
import type { ProviderService } from '../models/provider-service'
import type { ProjectDirectoryService } from '../projects/project-directory-service'
import { normalizeEngine } from './engines/coding-engine'

export class CodingSessionService {
  private readonly sessions = new Map<string, CodingSession>()

  constructor(
    private readonly store: WorkspaceStore,
    private readonly providers: ProviderService,
    private readonly projectDirectory: ProjectDirectoryService
  ) {}

  create(input: JsonRecord): CodingSession {
    const projectId = asString(input.projectId)
    if (!projectId) {
      throw badRequest('BAD_REQUEST', 'missing projectId', 'select a project')
    }
    const engineId = normalizeEngine(asString(input.engineId))
    const provider = this.providers.providerForCoding(
      asString(input.providerId),
      asString(input.model)
    )
    const model = asString(input.model) || asString(provider.defaultModel)
    const codeRoot = this.projectDirectory.codeRoot(projectId)
    const now = nowISO()
    const session: CodingSession = {
      sessionId: this.store.nextId('coding'),
      projectId,
      engineId,
      providerId: asString(provider.providerId),
      model,
      title: asString(input.title) || 'Coding Agent',
      localRootPath: codeRoot,
      engineThreadId: '',
      status: 'ready',
      createdAt: now,
      updatedAt: now
    }
    this.sessions.set(session.sessionId, session)
    return session
  }

  get(input: JsonRecord): CodingSession {
    const sessionId = asString(input.sessionId)
    const session = this.sessions.get(sessionId)
    if (!session) {
      throw notFound(
        'CODING_SESSION_NOT_FOUND',
        'coding session not found',
        'create a new coding session'
      )
    }
    return session
  }

  resolveForTurn(input: JsonRecord): CodingSession {
    const sessionId = asString(input.sessionId)
    if (sessionId) {
      const session = this.get({ sessionId })
      const engineId = asString(input.engineId)
      const providerId = asString(input.providerId)
      const model = asString(input.model)
      if (engineId) {
        session.engineId = normalizeEngine(engineId)
      }
      if (providerId) {
        session.providerId = providerId
      }
      if (model) {
        session.model = model
      }
      session.localRootPath = this.projectDirectory.codeRoot(session.projectId)
      this.save(session)
      return session
    }
    return this.create(input)
  }

  markRunning(session: CodingSession): void {
    session.status = 'running'
    session.updatedAt = nowISO()
    this.save(session)
  }

  markReady(sessionId: string): void {
    const current = this.sessions.get(sessionId)
    if (current) {
      current.status = 'ready'
      current.updatedAt = nowISO()
      this.save(current)
    }
  }

  save(session: CodingSession): void {
    this.sessions.set(session.sessionId, session)
  }
}
