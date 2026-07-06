import { spawn, type ChildProcess } from 'node:child_process'
import { createHash } from 'node:crypto'
import { asRecord, asString } from '../../../shared/util'
import type { JsonRecord } from '../../../types'
import { OpenCodeApiClient } from './opencode-api-client'
import { resolveOpenCodeCommand, nodePathForOpenCode } from './opencode-cli-resolver'
import {
  openCodeConfig,
  openCodeProviderEnv,
  openCodeProviderId,
  writeOpenCodeConfig
} from './opencode-config-writer'
import { normalizeDiff } from './opencode-diff-reader'
import {
  extractOpenCodeMessageText,
  normalizeOpenCodeMessage,
  openCodeMessageError
} from './opencode-event-normalizer'

export type OpenCodePromptResult = {
  messageId: string
  delivery: string
}

export type OpenCodeTurnState = {
  text: string
  done: boolean
  error: string
  activityCount: number
  changes: Array<{ path: string; status: string }>
}

export class OpenCodeManager {
  private process: ChildProcess | null = null
  private configHash = ''
  private cwd = ''
  private providerId = ''
  private model = ''
  private env: NodeJS.ProcessEnv = process.env
  private lastOutput = ''
  private readonly port = 43000 + Math.floor(Math.random() * 1000)
  private readonly api = new OpenCodeApiClient()

  async ensureServer(cwd: string, provider: JsonRecord, model: string): Promise<void> {
    const config = openCodeConfig(provider, model)
    const configHash = createHash('sha256').update(JSON.stringify(config)).digest('hex')
    if (
      this.process &&
      !this.process.killed &&
      this.configHash === configHash &&
      this.cwd === cwd
    ) {
      return
    }
    this.stop()
    const command = resolveOpenCodeCommand()
    if (!command) {
      throw new Error('OpenCode CLI was not found')
    }
    this.env = {
      ...process.env,
      ...openCodeProviderEnv(provider),
      NODE_PATH: nodePathForOpenCode(),
      OPENCODE_CONFIG: writeOpenCodeConfig(cwd, config),
      OPENCODE_CONFIG_CONTENT: JSON.stringify(config)
    }
    this.cwd = cwd
    this.providerId = openCodeProviderId(provider)
    this.model = model
    this.lastOutput = ''
    this.process = spawn(
      command.command,
      [...command.argsPrefix, 'serve', '--hostname=127.0.0.1', `--port=${this.port}`, '--register'],
      {
        cwd,
        env: this.env,
        stdio: ['ignore', 'pipe', 'pipe'],
        windowsHide: true
      }
    )
    this.process.stdout?.on('data', (chunk) => {
      this.lastOutput = `${this.lastOutput}${String(chunk)}`.slice(-4000)
    })
    this.process.stderr?.on('data', (chunk) => {
      this.lastOutput = `${this.lastOutput}${String(chunk)}`.slice(-4000)
    })
    try {
      await this.waitForServer()
      this.configHash = configHash
    } catch (error) {
      this.stop()
      throw error
    }
  }

  stop(): void {
    if (this.process && !this.process.killed) {
      this.process.kill()
    }
    this.process = null
    this.configHash = ''
    this.cwd = ''
    this.providerId = ''
    this.model = ''
  }

  createSession(
    cwd: string,
    provider: JsonRecord,
    model: string,
    title: string
  ): { sessionId: string } {
    const result = this.apiJson(
      'v2.session.create',
      [],
      {
        directory: cwd,
        title,
        model: {
          providerID: openCodeProviderId(provider),
          id: model
        },
        permission: {
          edit: 'allow',
          bash: 'allow'
        },
        metadata: {
          dreamworker: true
        }
      },
      15000
    )
    const data = asRecord(result.data)
    const sessionId = asString(data.id)
    if (!sessionId) {
      throw new Error('OpenCode did not return a session id')
    }
    return { sessionId }
  }

  prompt(sessionId: string, prompt: string): OpenCodePromptResult {
    const result = this.apiJson(
      'v2.session.prompt',
      ['--param', `sessionID=${sessionId}`],
      {
        prompt: { text: prompt },
        delivery: 'queue'
      },
      15000
    )
    const data = asRecord(result.data)
    return {
      messageId: asString(data.id),
      delivery: asString(data.delivery) || 'queue'
    }
  }

  interrupt(sessionId: string): void {
    if (!sessionId) {
      return
    }
    try {
      this.apiJson('v2.session.interrupt', ['--param', `sessionID=${sessionId}`], undefined, 5000)
    } catch {
      // Best effort cancellation; the DreamWorker stream still stops immediately.
    }
  }

  readTurnState(sessionId: string, userMessageId: string): OpenCodeTurnState {
    const messagesPayload = this.apiJson(
      'GET',
      [`/api/session/${sessionId}/message`],
      undefined,
      15000
    )
    const messages = Array.isArray(messagesPayload.data)
      ? messagesPayload.data.map((value) => normalizeOpenCodeMessage(value))
      : []
    const userIndex = messages.findIndex((message) => asString(message.id) === userMessageId)
    const candidates = userIndex > 0 ? messages.slice(0, userIndex) : messages
    const assistant = candidates.find(
      (message) => asString(message.type) === 'assistant' || asString(message.role) === 'assistant'
    )
    const text = assistant ? extractOpenCodeMessageText(assistant) : ''
    const finish = assistant ? asString(assistant.finish) : ''
    const error = assistant ? openCodeMessageError(assistant) : ''
    const changes = this.readDiff(sessionId)
    return {
      text,
      done: finish !== '' || error !== '',
      error,
      activityCount: messages.length + changes.length,
      changes
    }
  }

  private readDiff(sessionId: string): Array<{ path: string; status: string }> {
    try {
      const payload = this.apiJson('GET', [`/api/session/${sessionId}/diff`], undefined, 15000)
      return normalizeDiff(payload)
    } catch {
      return []
    }
  }

  private apiJson(operation: string, args: string[], data: unknown, timeoutMs: number): JsonRecord {
    return this.api.apiJson(operation, args, data, timeoutMs, this.cwd || process.cwd(), this.env)
  }

  private async waitForServer(): Promise<void> {
    const deadline = Date.now() + 15000
    let lastError = ''
    while (Date.now() < deadline) {
      if (this.process && this.process.exitCode !== null) {
        throw new Error(
          `OpenCode server exited with code ${this.process.exitCode}: ${this.lastOutput.trim()}`
        )
      }
      try {
        const response = await fetch(`http://127.0.0.1:${this.port}/openapi.json`)
        if (response.ok) {
          return
        }
        lastError = `HTTP ${response.status}`
      } catch (error) {
        lastError = error instanceof Error ? error.message : String(error)
      }
      await sleep(300)
    }
    throw new Error(`OpenCode server did not become ready: ${lastError || this.lastOutput.trim()}`)
  }
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolveSleep) => setTimeout(resolveSleep, ms))
}
