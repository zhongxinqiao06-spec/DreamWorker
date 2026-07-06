import { nowISO } from '../../../shared/util'
import type { CodingStreamEvent } from '../../../types'
import { resolveOpenCodeCli } from '../opencode/opencode-cli-resolver'
import { OpenCodeManager } from '../opencode/opencode-manager'
import type { CodingEngine, CodingEngineTurn } from './coding-engine'

export class OpenCodeEngine implements CodingEngine {
  readonly engineId = 'opencode'

  constructor(private readonly manager = new OpenCodeManager()) {}

  dispose(): void {
    this.manager.stop()
  }

  async *streamTurn(turn: CodingEngineTurn): AsyncGenerator<CodingStreamEvent> {
    const { prompt, session, provider, event, signal, updateSession } = turn
    const cli = resolveOpenCodeCli()
    if (!cli) {
      yield event({
        type: 'error',
        message: 'OpenCode CLI is not available in the Main Runtime package.',
        error: {
          code: 'OPENCODE_CLI_MISSING',
          message: 'OpenCode CLI is not available in the Main Runtime package.',
          recoverable: true
        }
      })
      return
    }

    await this.manager.ensureServer(session.localRootPath, provider, session.model)
    yield event({
      type: 'tool_call',
      message: 'OpenCode managed server is ready',
      toolCall: {
        callId: `${session.sessionId}_opencode`,
        toolName: 'opencode.server',
        arguments: { cli, cwd: session.localRootPath, auth: 'cli-api' }
      }
    })
    if (signal.aborted) {
      yield event({ type: 'cancelled', message: 'coding turn cancelled' })
      return
    }

    if (!session.engineThreadId) {
      const created = this.manager.createSession(
        session.localRootPath,
        provider,
        session.model,
        session.title
      )
      session.engineThreadId = created.sessionId
      session.updatedAt = nowISO()
      updateSession(session)
      yield event({
        type: 'tool_call',
        message: 'OpenCode session created',
        engineThreadId: session.engineThreadId,
        toolCall: {
          callId: `${session.sessionId}_session`,
          toolName: 'opencode.session.create',
          arguments: { sessionId: session.engineThreadId }
        }
      })
    }

    const admitted = this.manager.prompt(session.engineThreadId, prompt)
    yield event({
      type: 'tool_call',
      message: 'OpenCode prompt admitted',
      toolCall: {
        callId: admitted.messageId || `${session.sessionId}_prompt`,
        toolName: 'opencode.session.prompt',
        arguments: { delivery: admitted.delivery, messageId: admitted.messageId }
      }
    })
    yield event({
      type: 'tool_call',
      message: 'OpenCode event stream is tracked through authenticated CLI API polling',
      toolCall: {
        callId: `${session.sessionId}_event`,
        toolName: 'opencode.event.poll',
        arguments: { sessionId: session.engineThreadId, source: 'session.message + session.diff' }
      }
    })

    let emittedText = ''
    const emittedChanges = new Set<string>()
    const startedAt = Date.now()
    let lastActivityAt = startedAt
    let lastActivityCount = 0
    const turnTimeoutMs = readDurationEnv('DREAMWORKER_OPENCODE_TURN_TIMEOUT_MS', 180000)
    const idleTimeoutMs = readDurationEnv('DREAMWORKER_OPENCODE_IDLE_TIMEOUT_MS', 60000)
    for (;;) {
      if (signal.aborted) {
        this.manager.interrupt(session.engineThreadId)
        yield event({ type: 'cancelled', message: 'OpenCode turn cancelled' })
        return
      }
      const state = this.manager.readTurnState(session.engineThreadId, admitted.messageId)
      if (state.activityCount !== lastActivityCount) {
        lastActivityAt = Date.now()
        lastActivityCount = state.activityCount
      }
      if (state.text.startsWith(emittedText) && state.text.length > emittedText.length) {
        const delta = state.text.slice(emittedText.length)
        emittedText = state.text
        lastActivityAt = Date.now()
        yield event({ type: 'delta', delta })
      } else if (state.text && state.text !== emittedText) {
        emittedText = state.text
        lastActivityAt = Date.now()
        yield event({ type: 'delta', delta: state.text })
      }
      for (const change of state.changes) {
        const key = `${change.status}:${change.path}`
        if (emittedChanges.has(key)) {
          continue
        }
        emittedChanges.add(key)
        lastActivityAt = Date.now()
        yield event({
          type: 'file_changed',
          path: change.path,
          status: change.status,
          file: change
        })
      }
      if (state.done) {
        if (state.error) {
          yield event({
            type: 'error',
            message: state.error,
            error: { code: 'OPENCODE_TURN_FAILED', message: state.error, recoverable: true }
          })
          return
        }
        break
      }
      if (Date.now() - lastActivityAt > idleTimeoutMs) {
        const message = `OpenCode turn is still running but produced no new session events for ${Math.round(idleTimeoutMs / 1000)}s.`
        yield event({
          type: 'error',
          message,
          error: { code: 'OPENCODE_EVENT_IDLE_TIMEOUT', message, recoverable: true }
        })
        return
      }
      if (Date.now() - startedAt > turnTimeoutMs) {
        const message = 'OpenCode turn timed out while waiting for session events.'
        yield event({
          type: 'error',
          message,
          error: { code: 'OPENCODE_TURN_TIMEOUT', message, recoverable: true }
        })
        return
      }
      await sleep(900)
    }

    if (!emittedText) {
      yield event({ type: 'delta', delta: 'OpenCode completed without assistant text output.' })
    }
    yield event({
      type: 'completed',
      message: 'OpenCode turn completed.',
      engineThreadId: session.engineThreadId
    })
  }
}

function readDurationEnv(name: string, fallback: number): number {
  const raw = Number(process.env[name])
  return Number.isFinite(raw) && raw >= 1000 ? raw : fallback
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolveSleep) => setTimeout(resolveSleep, ms))
}
