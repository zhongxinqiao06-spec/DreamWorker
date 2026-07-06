import { spawn, spawnSync, type ChildProcess } from 'node:child_process'
import { createHash } from 'node:crypto'
import {
  existsSync,
  mkdirSync,
  readdirSync,
  readFileSync,
  writeFileSync,
  statSync,
  realpathSync
} from 'node:fs'
import { dirname, isAbsolute, join, relative, resolve, sep } from 'node:path'
import { fileURLToPath } from 'node:url'
import {
  badRequest,
  notFound,
  type CodingEngineId,
  type CodingSession,
  type CodingStreamEvent,
  type DeleteResult,
  type JsonRecord
} from '../types'
import { asString, newTraceId, nowISO, redactSecrets } from '../shared/util'
import type { WorkspaceStore } from '../store/workspace-store'

const maxReadFileBytes = 512 * 1024

const engineDescriptors = [
  {
    engineId: 'claude_agent',
    displayName: 'Claude Agent',
    description: 'Anthropic Claude Agent SDK, cwd scoped to project workspace/code.',
    supportedProviderTypes: ['anthropic'],
    preferredProviderIds: ['provider_anthropic'],
    directWrite: true,
    streaming: true
  },
  {
    engineId: 'codex',
    displayName: 'Codex',
    description: 'OpenAI Codex SDK thread run with workspace-write sandbox.',
    supportedProviderTypes: [
      'openai',
      'openai_compatible',
      'deepseek',
      'siliconflow',
      'glm',
      'custom'
    ],
    preferredProviderIds: ['provider_9router_local', 'provider_openai'],
    directWrite: true,
    streaming: false
  },
  {
    engineId: 'opencode',
    displayName: 'OpenCode',
    description: 'OpenCode SDK/CLI managed by the Main Runtime process.',
    supportedProviderTypes: [
      'openai',
      'openai_compatible',
      'deepseek',
      'siliconflow',
      'glm',
      'ollama',
      'custom'
    ],
    preferredProviderIds: ['provider_9router_local'],
    directWrite: true,
    streaming: true
  }
] as const

export class CodingService {
  private readonly sessions = new Map<string, CodingSession>()
  private readonly abortControllers = new Map<string, AbortController>()

  constructor(private readonly store: WorkspaceStore) {}

  dispose(): void {
    for (const controller of this.abortControllers.values()) {
      controller.abort()
    }
    this.abortControllers.clear()
    openCodeManager.stop()
  }

  async listEngines(): Promise<JsonRecord> {
    const engineStatuses = await Promise.all([
      runtimeStatusFor('claude_agent', '@anthropic-ai/claude-agent-sdk', 'claude'),
      runtimeStatusFor('codex', '@openai/codex-sdk', 'codex'),
      runtimeStatusFor('opencode', '@opencode-ai/sdk', 'opencode')
    ])
    const available = engineStatuses.some((status) => status.installed === true)
    return {
      runtimeDir: runtimeRoot(),
      nodeBin: process.execPath,
      adapterPath: fileURLToPath(import.meta.url),
      available,
      message: available
        ? 'Node coding runtime is ready'
        : 'Node coding runtime SDK packages are missing',
      engines: engineDescriptors,
      engineStatuses
    }
  }

  createSession(input: JsonRecord): CodingSession {
    const projectId = asString(input.projectId)
    if (!projectId) {
      throw badRequest('BAD_REQUEST', 'missing projectId', 'select a project')
    }
    const engineId = normalizeEngine(asString(input.engineId))
    const provider = this.store.providerForCoding(asString(input.providerId), asString(input.model))
    const model = asString(input.model) || asString(provider.defaultModel)
    const codeRoot = this.store.projectCodeRoot(projectId)
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

  getSession(input: JsonRecord): CodingSession {
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

  cancelTurn(input: JsonRecord): DeleteResult {
    const streamId = asString(input.streamId)
    if (!streamId) {
      throw badRequest('BAD_REQUEST', 'missing streamId', 'select an active coding turn')
    }
    const controller = this.abortControllers.get(streamId)
    if (!controller) {
      throw notFound(
        'CODING_STREAM_NOT_FOUND',
        'coding stream not found',
        'the turn may already be finished'
      )
    }
    controller.abort()
    this.abortControllers.delete(streamId)
    return { ok: true, deletedId: streamId }
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
    const session = this.resolveSessionForTurn(input)
    const provider = this.store.providerForCoding(session.providerId, session.model)
    const streamId = asString(input.streamId) || this.store.nextId('coding_stream')
    const traceId = newTraceId()
    const abort = new AbortController()
    this.abortControllers.set(streamId, abort)
    session.status = 'running'
    session.updatedAt = nowISO()
    this.sessions.set(session.sessionId, session)

    let sequence = 0
    const base = (): Omit<CodingStreamEvent, 'type' | 'sequence' | 'timestamp'> => ({
      streamId,
      sessionId: session.sessionId,
      engineId: session.engineId,
      providerId: asString(provider.providerId),
      model: session.model,
      trace_id: traceId
    })
    const event = (
      eventInput: Omit<
        CodingStreamEvent,
        | 'streamId'
        | 'sessionId'
        | 'engineId'
        | 'providerId'
        | 'model'
        | 'trace_id'
        | 'sequence'
        | 'timestamp'
      >
    ): CodingStreamEvent => ({
      ...base(),
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

      if (session.engineId === 'opencode') {
        yield* this.runOpenCodeTurn(prompt, session, provider, event, abort.signal)
      } else {
        yield event({
          type: 'delta',
          delta:
            session.engineId === 'codex'
              ? 'Codex SDK is now hosted inside the Main Runtime. '
              : 'Claude Agent SDK is now hosted inside the Main Runtime. '
        })
        yield event({
          type: 'completed',
          message: 'Main Runtime coding service completed the turn handoff.'
        })
      }
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
      this.abortControllers.delete(streamId)
      const current = this.sessions.get(session.sessionId)
      if (current) {
        current.status = 'ready'
        current.updatedAt = nowISO()
        this.sessions.set(current.sessionId, current)
      }
    }
  }

  listFiles(input: JsonRecord): JsonRecord[] {
    const root = this.store.projectCodeRoot(asString(input.projectId))
    const limitValue = typeof input.limit === 'number' ? input.limit : 500
    const limit = limitValue > 0 && limitValue <= 1000 ? limitValue : 500
    const query = asString(input.query).toLowerCase()
    const gitStatus = gitStatusMap(root)
    const entries: JsonRecord[] = []
    walkProject(root, (path, isDir) => {
      if (entries.length >= limit) {
        return false
      }
      const rel = relative(root, path).replaceAll(sep, '/')
      if (query && !rel.toLowerCase().includes(query)) {
        return true
      }
      const stats = statSync(path)
      entries.push({
        path: rel,
        name: path.split(/[\\/]/).pop() || rel,
        isDir,
        size: Number(stats.size),
        modifiedAt: stats.mtime.toISOString(),
        gitStatus: gitStatus.get(rel) ?? ''
      })
      return true
    })
    return entries.sort((left, right) => {
      if (left.isDir !== right.isDir) {
        return left.isDir ? -1 : 1
      }
      return asString(left.path).localeCompare(asString(right.path))
    })
  }

  readFile(input: JsonRecord): JsonRecord {
    const projectId = asString(input.projectId)
    const root = this.store.projectCodeRoot(projectId)
    const path = safeProjectPath(root, asString(input.path))
    const stats = statSync(path)
    if (stats.isDirectory()) {
      throw badRequest('FILE_IS_DIRECTORY', 'path is a directory', 'select a file')
    }
    const raw = readFileSync(path)
    const truncated = raw.length > maxReadFileBytes
    const payload = truncated ? raw.subarray(0, maxReadFileBytes) : raw
    const content = payload.toString('utf8')
    const rel = relative(root, path).replaceAll(sep, '/')
    return {
      projectId,
      path: rel,
      content,
      size: Number(stats.size),
      truncated,
      mimeType: 'text/plain'
    }
  }

  fileStatus(input: JsonRecord): JsonRecord {
    const projectId = asString(input.projectId)
    const root = this.store.projectCodeRoot(projectId)
    const changes = gitChanges(root)
    const branch = gitBranch(root)
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

  private resolveSessionForTurn(input: JsonRecord): CodingSession {
    const sessionId = asString(input.sessionId)
    if (sessionId) {
      const session = this.getSession({ sessionId })
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
      session.localRootPath = this.store.projectCodeRoot(session.projectId)
      this.sessions.set(session.sessionId, session)
      return session
    }
    return this.createSession(input)
  }

  private async *runOpenCodeTurn(
    prompt: string,
    session: CodingSession,
    provider: JsonRecord,
    event: (
      eventInput: Omit<
        CodingStreamEvent,
        | 'streamId'
        | 'sessionId'
        | 'engineId'
        | 'providerId'
        | 'model'
        | 'trace_id'
        | 'sequence'
        | 'timestamp'
      >
    ) => CodingStreamEvent,
    signal: AbortSignal
  ): AsyncGenerator<CodingStreamEvent> {
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
    await openCodeManager.ensureServer(session.localRootPath, provider, session.model)
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
      const created = openCodeManager.createSession(
        session.localRootPath,
        provider,
        session.model,
        session.title
      )
      session.engineThreadId = created.sessionId
      session.updatedAt = nowISO()
      this.sessions.set(session.sessionId, session)
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

    const admitted = openCodeManager.prompt(session.engineThreadId, prompt)
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
        openCodeManager.interrupt(session.engineThreadId)
        yield event({ type: 'cancelled', message: 'OpenCode turn cancelled' })
        return
      }
      const state = openCodeManager.readTurnState(session.engineThreadId, admitted.messageId)
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

type OpenCodePromptResult = {
  messageId: string
  delivery: string
}

type OpenCodeTurnState = {
  text: string
  done: boolean
  error: string
  activityCount: number
  changes: Array<{ path: string; status: string }>
}

type OpenCodeCommand = {
  command: string
  argsPrefix: string[]
  displayPath: string
}

class OpenCodeManager {
  private process: ChildProcess | null = null
  private configHash = ''
  private cwd = ''
  private providerId = ''
  private model = ''
  private env: NodeJS.ProcessEnv = process.env
  private lastOutput = ''
  private readonly port = 43000 + Math.floor(Math.random() * 1000)

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
    const data = toRecord(result.data)
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
    const data = toRecord(result.data)
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
    const command = resolveOpenCodeCommand()
    if (!command) {
      throw new Error('OpenCode CLI was not found')
    }
    const commandArgs = [...command.argsPrefix, 'api', operation, ...args]
    if (data !== undefined) {
      commandArgs.push('--data', JSON.stringify(data))
    }
    const result = spawnSync(command.command, commandArgs, {
      cwd: this.cwd || process.cwd(),
      env: this.env,
      encoding: 'utf8',
      timeout: timeoutMs,
      windowsHide: true
    })
    if (result.status !== 0) {
      throw new Error(
        result.stderr.trim() || result.stdout.trim() || `OpenCode API ${operation} failed`
      )
    }
    const stdout = result.stdout.trim()
    if (!stdout) {
      return {}
    }
    let payload: JsonRecord
    try {
      const parsed = JSON.parse(stdout) as unknown
      payload = Array.isArray(parsed) ? { data: parsed } : toRecord(parsed)
    } catch (error) {
      throw new Error(`OpenCode API ${operation} returned invalid JSON: ${stdout.slice(0, 400)}`, {
        cause: error
      })
    }
    const tag = asString(payload._tag)
    if (tag.endsWith('Error') || tag === 'UnauthorizedError') {
      throw new Error(asString(payload.message) || `OpenCode API ${operation} failed with ${tag}`)
    }
    return payload
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

const openCodeManager = new OpenCodeManager()

async function runtimeStatusFor(
  engineId: CodingEngineId,
  packageName: string,
  key: string
): Promise<JsonRecord> {
  try {
    await import(packageName)
    if (engineId === 'opencode') {
      const command = resolveOpenCodeCommand()
      if (!command) {
        return {
          engineId,
          packageName,
          installed: true,
          executable: false,
          status: 'error',
          message: 'OpenCode SDK is installed but CLI binary was not found',
          key
        }
      }
      const version = spawnSync(command.command, [...command.argsPrefix, '--version'], {
        encoding: 'utf8',
        timeout: 5000,
        windowsHide: true
      })
      if (version.status !== 0) {
        return {
          engineId,
          packageName,
          installed: true,
          executable: false,
          status: 'error',
          message:
            version.stderr.trim() || version.stdout.trim() || 'OpenCode CLI is not executable',
          key
        }
      }
    }
    return {
      engineId,
      packageName,
      installed: true,
      executable: true,
      status: 'ready',
      message: `${packageName} is installed in Main Runtime`,
      key
    }
  } catch (error) {
    return {
      engineId,
      packageName,
      installed: false,
      executable: false,
      status: 'missing',
      message: error instanceof Error ? error.message : `${packageName} is missing`,
      key
    }
  }
}

function runtimeRoot(): string {
  return resolve(dirname(fileURLToPath(import.meta.url)), '..')
}

function normalizeEngine(value: string): CodingEngineId {
  if (value === 'codex' || value === 'opencode' || value === 'claude_agent') {
    return value
  }
  return 'claude_agent'
}

function resolveOpenCodeCli(): string {
  return resolveOpenCodeCommand()?.displayPath ?? ''
}

function resolveOpenCodeCommand(): OpenCodeCommand | null {
  const executable = resolveOpenCodeExecutable()
  if (executable) {
    return { command: executable, argsPrefix: [], displayPath: executable }
  }
  const packageDir = findNodePackageDir('@opencode-ai/cli')
  if (!packageDir) {
    return null
  }
  const cliPath = join(packageDir, 'bin', 'lildax')
  const unpackedCliPath = existingAsarAwarePath(cliPath)
  return unpackedCliPath
    ? { command: process.execPath, argsPrefix: [unpackedCliPath], displayPath: unpackedCliPath }
    : null
}

function resolveOpenCodeExecutable(): string {
  const packageDir = findNodePackageDir('@opencode-ai/cli')
  if (!packageDir) {
    return ''
  }
  const nodeModulesDir = findAncestorNodeModules(packageDir)
  if (!nodeModulesDir) {
    return ''
  }
  const platform = process.platform === 'win32' ? 'windows' : process.platform
  const arch = process.arch === 'arm64' ? 'arm64' : process.arch === 'arm' ? 'arm' : 'x64'
  const binary = process.platform === 'win32' ? 'lildax.exe' : 'lildax'
  const packageNames = [
    `cli-${platform}-${arch}`,
    `cli-${platform}-${arch}-baseline`,
    `cli-${platform}-${arch}-musl`,
    `cli-${platform}-${arch}-baseline-musl`
  ]
  for (const packageName of packageNames) {
    const candidate = join(nodeModulesDir, '@opencode-ai', packageName, 'bin', binary)
    const executable = existingAsarAwarePath(candidate)
    if (executable) {
      return executable
    }
  }
  return ''
}

function existingAsarAwarePath(path: string): string {
  if (existsSync(path)) {
    return path
  }
  const unpacked = path.replace(`${sep}app.asar${sep}`, `${sep}app.asar.unpacked${sep}`)
  return unpacked !== path && existsSync(unpacked) ? unpacked : ''
}

function findAncestorNodeModules(path: string): string {
  let current = resolve(path)
  for (;;) {
    if (current.endsWith(`${sep}node_modules`)) {
      return current
    }
    const parent = dirname(current)
    if (parent === current) {
      return ''
    }
    current = parent
  }
}

function findNodePackageDir(packageName: string): string {
  const segments = packageName.split('/')
  const starts = [runtimeRoot(), process.cwd(), dirname(fileURLToPath(import.meta.url))]
  for (const start of starts) {
    let current = resolve(start)
    for (;;) {
      const candidate = join(current, 'node_modules', ...segments)
      if (existsSync(join(candidate, 'package.json'))) {
        return candidate
      }
      const parent = dirname(current)
      if (parent === current) {
        break
      }
      current = parent
    }
  }
  return ''
}

function openCodeProviderId(provider: JsonRecord): string {
  const providerType = asString(provider.providerType)
  if (providerType === 'anthropic') {
    return 'anthropic'
  }
  if (providerType === 'openai') {
    return 'openai'
  }
  if (providerType === 'ollama') {
    return 'ollama'
  }
  return 'dreamworker'
}

function openCodeConfig(provider: JsonRecord, model: string): JsonRecord {
  const providerId = openCodeProviderId(provider)
  const baseURL = asString(provider.baseURL)
  const apiKey = asString(provider.apiKey)
  const modelName = model || asString(provider.defaultModel)
  const config: JsonRecord = {
    autoupdate: false,
    share: 'disabled',
    model: `${providerId}/${modelName}`,
    permission: {
      edit: 'allow',
      bash: 'allow'
    },
    tools: {
      write: true,
      edit: true,
      bash: true
    },
    provider: {}
  }
  if (providerId === 'dreamworker') {
    config.provider = {
      dreamworker: {
        npm: '@ai-sdk/openai-compatible',
        name: asString(provider.displayName) || 'DreamWorker Provider',
        options: {
          baseURL,
          apiKey: apiKey ? '{env:DREAMWORKER_OPENCODE_API_KEY}' : undefined
        },
        models: {
          [modelName]: {
            name: modelName
          }
        }
      }
    }
    return config
  }
  return config
}

function openCodeProviderEnv(provider: JsonRecord): NodeJS.ProcessEnv {
  const apiKey = asString(provider.apiKey)
  const providerType = asString(provider.providerType)
  const env: NodeJS.ProcessEnv = {
    DREAMWORKER_OPENCODE_API_KEY: apiKey
  }
  if (providerType === 'openai' || openCodeProviderId(provider) === 'dreamworker') {
    env.OPENAI_API_KEY = apiKey
  }
  if (providerType === 'anthropic') {
    env.ANTHROPIC_API_KEY = apiKey
  }
  return env
}

function writeOpenCodeConfig(_cwd: string, config: JsonRecord): string {
  const serialized = JSON.stringify(config, null, 2)
  const configDir = resolve(_cwd, '..', '..')
  mkdirSync(configDir, { recursive: true })
  const configPath = join(configDir, 'opencode.json')
  writeFileSync(configPath, `${serialized}\n`, 'utf8')
  return configPath
}

function nodePathForOpenCode(): string {
  const packageDir = findNodePackageDir('@opencode-ai/cli')
  const nodeModulesDir = packageDir ? findAncestorNodeModules(packageDir) : ''
  return [nodeModulesDir, process.env.NODE_PATH]
    .filter(Boolean)
    .join(process.platform === 'win32' ? ';' : ':')
}

function readDurationEnv(name: string, fallback: number): number {
  const raw = Number(process.env[name])
  return Number.isFinite(raw) && raw >= 1000 ? raw : fallback
}

function normalizeOpenCodeMessage(value: unknown): JsonRecord {
  const record = toRecord(value)
  const info = toRecord(record.info)
  if (Object.keys(info).length > 0) {
    return { ...info, parts: Array.isArray(record.parts) ? record.parts : [] }
  }
  return record
}

function extractOpenCodeMessageText(message: JsonRecord): string {
  if (typeof message.text === 'string') {
    return message.text
  }
  const parts = Array.isArray(message.parts) ? message.parts : []
  for (const part of parts) {
    const record = toRecord(part)
    if (typeof record.text === 'string') {
      return record.text
    }
  }
  const content = Array.isArray(message.content) ? message.content : []
  const chunks: string[] = []
  for (const part of content) {
    const record = toRecord(part)
    if (typeof record.text === 'string') {
      chunks.push(record.text)
    }
    if (typeof record.content === 'string') {
      chunks.push(record.content)
    }
    if (typeof record.output === 'string') {
      chunks.push(record.output)
    }
  }
  return chunks.join('')
}

function openCodeMessageError(message: JsonRecord): string {
  const error = toRecord(message.error)
  const data = toRecord(error.data)
  const text = asString(error.message) || asString(data.message) || asString(error.name)
  return text ? redactSecrets(text) : ''
}

function normalizeDiff(payload: JsonRecord): Array<{ path: string; status: string }> {
  const data = payload.data
  const record = toRecord(data)
  let items: unknown[] = []
  if (Array.isArray(data)) {
    items = data
  } else if (Array.isArray(record.files)) {
    items = record.files
  } else if (Array.isArray(record.changes)) {
    items = record.changes
  }
  return items
    .map((item) => toRecord(item))
    .map((item) => ({
      path: asString(item.path) || asString(item.file) || asString(item.filename),
      status: asString(item.status) || asString(item.type) || 'modified'
    }))
    .filter((item) => item.path !== '')
}

function toRecord(value: unknown): JsonRecord {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
    ? (value as JsonRecord)
    : {}
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolveSleep) => setTimeout(resolveSleep, ms))
}

function safeProjectPath(root: string, raw: string): string {
  const normalized = raw.trim().replaceAll('\\', '/')
  if (!normalized || normalized.startsWith('/') || normalized.includes('\0')) {
    throw badRequest(
      'PATH_OUTSIDE_PROJECT',
      'file path must be relative to the project',
      'select a project file'
    )
  }
  const joined = resolve(root, normalized)
  const rel = relative(root, joined)
  if (rel === '..' || rel.startsWith(`..${sep}`) || isAbsolute(rel)) {
    throw badRequest(
      'PATH_OUTSIDE_PROJECT',
      'file path escapes the project root',
      'select a file inside the project'
    )
  }
  if (!existsSync(joined)) {
    throw notFound('FILE_NOT_FOUND', 'file not found', 'select another project file')
  }
  const realRoot = realpathSync(root)
  const realPath = realpathSync(joined)
  const realRel = relative(realRoot, realPath)
  if (realRel === '..' || realRel.startsWith(`..${sep}`) || isAbsolute(realRel)) {
    throw badRequest(
      'PATH_OUTSIDE_PROJECT',
      'resolved file path escapes the project root',
      'select a file inside the project'
    )
  }
  return realPath
}

function walkProject(root: string, visit: (path: string, isDir: boolean) => boolean): void {
  const stack = [root]
  while (stack.length > 0) {
    const current = stack.pop()
    if (!current) {
      continue
    }
    for (const entry of readdirSync(current, { withFileTypes: true })) {
      if (entry.isDirectory() && shouldSkipDir(entry.name)) {
        continue
      }
      const fullPath = join(current, entry.name)
      if (!visit(fullPath, entry.isDirectory())) {
        return
      }
      if (entry.isDirectory()) {
        stack.push(fullPath)
      }
    }
  }
}

function shouldSkipDir(name: string): boolean {
  return new Set([
    '.git',
    'node_modules',
    'dist',
    'out',
    'release',
    'coverage',
    '.cache',
    '.vite',
    'tmp'
  ]).has(name)
}

function gitChanges(root: string): JsonRecord[] {
  const result = spawnSync('git', ['status', '--short', '--porcelain=v1'], {
    cwd: root,
    encoding: 'utf8',
    windowsHide: true
  })
  if (result.status !== 0 || !result.stdout.trim()) {
    return []
  }
  return result.stdout
    .trim()
    .split(/\r?\n/)
    .filter(Boolean)
    .map((line) => {
      const status = line.slice(0, 2).trim()
      const path = line.slice(3).trim().split(' -> ').pop() ?? ''
      return { path: path.replaceAll('\\', '/'), status }
    })
}

function gitStatusMap(root: string): Map<string, string> {
  return new Map(gitChanges(root).map((change) => [asString(change.path), asString(change.status)]))
}

function gitBranch(root: string): string {
  const result = spawnSync('git', ['branch', '--show-current'], {
    cwd: root,
    encoding: 'utf8',
    windowsHide: true
  })
  return result.status === 0 ? result.stdout.trim() : ''
}
