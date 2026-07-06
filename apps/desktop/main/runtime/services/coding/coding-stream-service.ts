import { CancellationRegistry } from '../../kernel/cancellation'
import { badRequest } from '../../kernel/errors'
import { asString, newTraceId, nowISO, redactSecrets } from '../../shared/util'
import type { CodingStreamEvent, DeleteResult, JsonRecord } from '../../types'
import type { WorkspaceStore } from '../../store/workspace-store'
import type { CodingEngineRegistry } from './engines/engine-registry'
import type { CodingEventInput } from './engines/coding-engine'
import type { CodingSessionService } from './coding-session-service'

export class CodingStreamService {
  private readonly cancellations = new CancellationRegistry()

  constructor(
    private readonly store: WorkspaceStore,
    private readonly sessions: CodingSessionService,
    private readonly engines: CodingEngineRegistry
  ) {}

  cancelTurn(input: JsonRecord): DeleteResult {
    return this.cancellations.cancel(input)
  }

  dispose(): void {
    this.cancellations.abortAll()
  }

  async *streamTurn(input: JsonRecord): AsyncGenerator<CodingStreamEvent> {
    const prompt = asString(input.prompt)
    if (!prompt) {
      throw badRequest(
        'BAD_REQUEST',
        'prompt is required',
        'enter an instruction for the coding agent'
      )
    }
    const session = this.sessions.resolveForTurn(input)
    const provider = this.store.providerForCoding(session.providerId, session.model)
    const streamId = asString(input.streamId) || this.store.nextId('coding_stream')
    const traceId = newTraceId()
    const abort = this.cancellations.start(streamId)
    this.sessions.markRunning(session)

    let sequence = 0
    const event = (eventInput: CodingEventInput): CodingStreamEvent => ({
      streamId,
      sessionId: session.sessionId,
      engineId: session.engineId,
      providerId: asString(provider.providerId),
      model: session.model,
      trace_id: traceId,
      ...eventInput,
      sequence: ++sequence,
      timestamp: nowISO()
    })

    try {
      yield event({
        type: 'started',
        message: `${session.engineId} started`,
        runtimeAvailable: true
      })
      yield event({
        type: 'tool_call',
        message: 'Main Runtime selected project code root',
        toolCall: {
          callId: `${streamId}_workspace`,
          toolName: 'workspace.code_root',
          arguments: { cwd: session.localRootPath }
        }
      })

      if (abort.signal.aborted) {
        yield event({ type: 'cancelled', message: 'coding turn cancelled' })
        return
      }

      const engine = this.engines.get(session.engineId)
      yield* engine.streamTurn({
        prompt,
        session,
        provider,
        signal: abort.signal,
        event,
        updateSession: (nextSession) => this.sessions.save(nextSession)
      })
    } catch (error) {
      yield event({
        type: 'error',
        message: error instanceof Error ? redactSecrets(error.message) : 'coding turn failed',
        error: {
          code: 'CODING_TURN_FAILED',
          message: error instanceof Error ? redactSecrets(error.message) : 'coding turn failed',
          recoverable: true
        }
      })
    } finally {
      this.cancellations.complete(streamId)
      this.sessions.markReady(session.sessionId)
    }
  }
}
