#!/usr/bin/env node
import { createRequire } from 'node:module'
import { createInterface } from 'node:readline'

type EngineId = 'claude_agent' | 'codex' | 'opencode'

type RpcRequest = {
  readonly id?: string
  readonly method?: string
  readonly params?: unknown
}

type CodingProvider = {
  readonly providerId: string
  readonly providerType: string
  readonly displayName: string
  readonly baseURL: string
  readonly apiKey?: string
}

type TurnParams = {
  readonly streamId: string
  readonly sessionId: string
  readonly engineId: EngineId
  readonly provider: CodingProvider
  readonly model: string
  readonly prompt: string
  readonly cwd: string
  readonly engineThreadId?: string
}

type SessionParams = {
  readonly sessionId?: string
  readonly engineId: EngineId
  readonly providerId: string
  readonly model: string
  readonly cwd: string
}

type CodingEvent = {
  readonly type:
    | 'started'
    | 'delta'
    | 'tool_call'
    | 'shell_output'
    | 'file_changed'
    | 'completed'
    | 'cancelled'
    | 'error'
  readonly delta?: string
  readonly message?: string
  readonly callId?: string
  readonly toolName?: string
  readonly arguments?: unknown
  readonly command?: string
  readonly output?: string
  readonly path?: string
  readonly status?: string
  readonly engineThreadId?: string
  readonly error?: {
    readonly code: string
    readonly message: string
    readonly recoverable: boolean
  }
}

type RpcEnvelope =
  | { readonly id: string; readonly result: unknown }
  | { readonly id: string; readonly error: { readonly code: string; readonly message: string } }
  | { readonly id: string; readonly event: CodingEvent }

const requireFromHere = createRequire(import.meta.url)
const activeAbortControllers = new Map<string, AbortController>()

const engines = [
  {
    engineId: 'claude_agent',
    displayName: 'Claude Agent',
    description: 'Anthropic Claude Agent SDK with acceptEdits project-write mode.',
    supportedProviderTypes: ['anthropic'],
    preferredProviderIds: ['provider_anthropic'],
    directWrite: true,
    streaming: true
  },
  {
    engineId: 'codex',
    displayName: 'Codex',
    description: 'OpenAI Codex SDK thread runs with workspace-write sandbox.',
    supportedProviderTypes: ['openai', 'openai_compatible', 'deepseek', 'siliconflow', 'glm', 'custom'],
    preferredProviderIds: ['provider_9router_local', 'provider_openai'],
    directWrite: true,
    streaming: false
  },
  {
    engineId: 'opencode',
    displayName: 'OpenCode',
    description: 'OpenCode SDK server/client session prompt runtime.',
    supportedProviderTypes: ['openai', 'openai_compatible', 'deepseek', 'siliconflow', 'glm', 'ollama', 'custom'],
    preferredProviderIds: ['provider_9router_local'],
    directWrite: true,
    streaming: true
  }
] as const

async function main(): Promise<void> {
  if (process.argv.includes('--health-check')) {
    await healthCheck()
    return
  }

  const input = createInterface({ input: process.stdin, crlfDelay: Infinity })
  for await (const line of input) {
    if (!line.trim()) {
      continue
    }
    void handleLine(line)
  }
}

async function handleLine(line: string): Promise<void> {
  let request: RpcRequest
  try {
    request = JSON.parse(line) as RpcRequest
  } catch (error) {
    write({
      id: '',
      error: {
        code: 'BAD_JSON',
        message: error instanceof Error ? error.message : 'invalid JSON-RPC line'
      }
    })
    return
  }

  const id = request.id ?? ''
  try {
    const result = await dispatch(id, request.method ?? '', request.params)
    if (result !== undefined) {
      write({ id, result })
    }
  } catch (error) {
    write({
      id,
      error: {
        code: 'ADAPTER_ERROR',
        message: error instanceof Error ? error.message : String(error)
      }
    })
  }
}

async function dispatch(id: string, method: string, params: unknown): Promise<unknown> {
  switch (method) {
    case 'engine.list':
      return engines
    case 'session.create':
      return createSession(assertRecord(params))
    case 'turn.run':
      await runTurn(id, assertTurnParams(params))
      return undefined
    case 'turn.cancel':
      return cancelTurn(assertRecord(params))
    case 'files.list':
    case 'files.read':
    case 'files.status':
      return {
        ok: false,
        reason: 'files are served by DreamWorker Go Engine to enforce project-root boundaries'
      }
    default:
      throw new Error(`unsupported method: ${method}`)
  }
}

function createSession(params: Record<string, unknown>) {
  const session = params as Partial<SessionParams>
  return {
    sessionId: session.sessionId ?? `coding_${Date.now()}_${Math.random().toString(36).slice(2)}`,
    engineId: session.engineId,
    providerId: session.providerId,
    model: session.model,
    cwd: session.cwd,
    engineThreadId: ''
  }
}

function cancelTurn(params: Record<string, unknown>) {
  const streamId = typeof params.streamId === 'string' ? params.streamId : ''
  const controller = activeAbortControllers.get(streamId)
  if (controller) {
    controller.abort()
    activeAbortControllers.delete(streamId)
  }
  return { ok: true, deletedId: streamId }
}

async function runTurn(id: string, params: TurnParams): Promise<void> {
  const abortController = new AbortController()
  activeAbortControllers.set(params.streamId, abortController)
  try {
    applyProviderEnvironment(params.provider)
    emit(id, { type: 'started', message: `${params.engineId} started` })
    switch (params.engineId) {
      case 'claude_agent':
        await runClaude(id, params, abortController.signal)
        break
      case 'codex':
        await runCodex(id, params, abortController.signal)
        break
      case 'opencode':
        await runOpenCode(id, params, abortController.signal)
        break
      default:
        throw new Error(`unsupported engine: ${params.engineId}`)
    }
  } catch (error) {
    if (abortController.signal.aborted) {
      emit(id, { type: 'cancelled', message: 'turn cancelled' })
      return
    }
    emit(id, {
      type: 'error',
      error: {
        code: 'ENGINE_RUN_FAILED',
        message: error instanceof Error ? error.message : String(error),
        recoverable: true
      }
    })
  } finally {
    activeAbortControllers.delete(params.streamId)
  }
}

async function runClaude(id: string, params: TurnParams, signal: AbortSignal): Promise<void> {
  const sdk = (await import('@anthropic-ai/claude-agent-sdk')) as Record<string, unknown>
  const query = sdk.query
  if (typeof query !== 'function') {
    throw new Error('Claude Agent SDK query() export was not found')
  }

  const options = {
    cwd: params.cwd,
    model: params.model,
    permissionMode: 'acceptEdits',
    allowedTools: ['Read', 'Edit', 'Write', 'Bash', 'Glob', 'Grep'],
    abortController: signal
  }

  for await (const message of query({ prompt: params.prompt, options } as never) as AsyncIterable<unknown>) {
    if (signal.aborted) {
      emit(id, { type: 'cancelled', message: 'turn cancelled' })
      return
    }
    emitClaudeMessage(id, message)
  }
  emit(id, { type: 'completed', message: 'Claude Agent turn completed' })
}

async function runCodex(id: string, params: TurnParams, signal: AbortSignal): Promise<void> {
  const sdk = (await import('@openai/codex-sdk')) as Record<string, unknown>
  const CodexCtor = sdk.Codex
  if (typeof CodexCtor !== 'function') {
    throw new Error('Codex SDK Codex export was not found')
  }

  const codex = new (CodexCtor as new (config?: unknown) => {
    startThread: (options?: unknown) => unknown
    resumeThread: (threadId: string) => unknown
  })({
    cwd: params.cwd,
    model: params.model
  })
  const thread = params.engineThreadId
    ? codex.resumeThread(params.engineThreadId)
    : codex.startThread({ model: params.model, sandbox: 'workspace_write', cwd: params.cwd })
  if (signal.aborted) {
    emit(id, { type: 'cancelled', message: 'turn cancelled' })
    return
  }
  const result = await callThreadRun(thread, params.prompt, {
    model: params.model,
    sandbox: 'workspace_write',
    cwd: params.cwd,
    signal
  })
  if (signal.aborted) {
    emit(id, { type: 'cancelled', message: 'turn cancelled' })
    return
  }
  const threadId = readString(result, 'threadId') || readString(thread, 'id') || params.engineThreadId
  const text = extractResultText(result)
  if (text) {
    emit(id, { type: 'delta', delta: text })
  }
  emit(id, {
    type: 'completed',
    message: 'Codex turn completed',
    ...(threadId ? { engineThreadId: threadId } : {})
  })
}

async function runOpenCode(id: string, params: TurnParams, signal: AbortSignal): Promise<void> {
  const sdk = (await import('@opencode-ai/sdk')) as Record<string, unknown>
  const createOpencode = sdk.createOpencode
  if (typeof createOpencode !== 'function') {
    throw new Error('OpenCode SDK createOpencode() export was not found')
  }

  const runtime = (await createOpencode({
    cwd: params.cwd,
    config: {
      model: modelNameForOpenCode(params),
      tools: {
        write: true,
        edit: true,
        bash: true,
        read: true
      }
    }
  } as never)) as {
    client?: Record<string, unknown>
    close?: () => Promise<void> | void
  }
  try {
    const client = runtime.client
    if (!client) {
      throw new Error('OpenCode SDK did not return a client')
    }
    const session = await callOpenCode(client, ['session', 'create'], { body: { title: 'DreamWorker Coding Agent' } })
    const sessionId = readNestedString(session, ['data', 'id']) || readString(session, 'id') || params.sessionId
    if (signal.aborted) {
      emit(id, { type: 'cancelled', message: 'turn cancelled' })
      return
    }
    const result = await callOpenCode(client, ['session', 'prompt'], {
      path: { id: sessionId },
      body: {
        model: {
          providerID: openCodeProviderId(params.provider),
          modelID: params.model
        },
        parts: [{ type: 'text', text: params.prompt }]
      }
    })
    const text = extractResultText(result)
    if (text) {
      emit(id, { type: 'delta', delta: text })
    }
    emit(id, { type: 'completed', message: 'OpenCode turn completed' })
  } finally {
    await runtime.close?.()
  }
}

function emitClaudeMessage(id: string, message: unknown): void {
  if (!isRecord(message)) {
    emit(id, { type: 'delta', delta: String(message) })
    return
  }
  const kind = readString(message, 'type')
  if (kind === 'assistant' || kind === 'partial_assistant') {
    const text = extractResultText(message)
    if (text) {
      emit(id, { type: 'delta', delta: text })
    }
    return
  }
  if (kind === 'result') {
    const text = extractResultText(message)
    if (text) {
      emit(id, { type: 'delta', delta: text })
    }
    return
  }
  if (kind.toLowerCase().includes('tool')) {
    emit(id, {
      type: 'tool_call',
      toolName: readString(message, 'name') || kind,
      arguments: message
    })
    return
  }
  if (kind.toLowerCase().includes('file')) {
    emit(id, {
      type: 'file_changed',
      path: readString(message, 'path') || readString(message, 'filePath'),
      status: kind
    })
  }
}

function applyProviderEnvironment(provider: CodingProvider): void {
  const apiKey = provider.apiKey ?? ''
  if (provider.providerType === 'anthropic') {
    if (apiKey) {
      process.env.ANTHROPIC_API_KEY = apiKey
    }
    return
  }
  if (apiKey) {
    process.env.OPENAI_API_KEY = apiKey
  }
  if (provider.baseURL) {
    process.env.OPENAI_BASE_URL = provider.baseURL
    process.env.OPENCODE_BASE_URL = provider.baseURL
  }
}

async function callThreadRun(thread: unknown, prompt: string, options: unknown): Promise<unknown> {
  if (!isRecord(thread) || typeof thread.run !== 'function') {
    throw new Error('Codex thread run() was not available')
  }
  return thread.run(prompt, options)
}

async function callOpenCode(client: Record<string, unknown>, path: readonly string[], payload: unknown): Promise<unknown> {
  let target: unknown = client
  for (const segment of path) {
    if (!isRecord(target)) {
      throw new Error(`OpenCode client path missing: ${path.join('.')}`)
    }
    target = target[segment]
  }
  if (typeof target !== 'function') {
    throw new Error(`OpenCode client method missing: ${path.join('.')}`)
  }
  return target(payload)
}

function modelNameForOpenCode(params: TurnParams): string {
  return `${openCodeProviderId(params.provider)}/${params.model}`
}

function openCodeProviderId(provider: CodingProvider): string {
  if (provider.providerType === 'anthropic') {
    return 'anthropic'
  }
  if (provider.providerType === 'openai') {
    return 'openai'
  }
  if (provider.providerType === 'ollama') {
    return 'ollama'
  }
  return provider.providerId.includes('9router') ? 'openai' : provider.providerType
}

function extractResultText(value: unknown): string {
  if (typeof value === 'string') {
    return value
  }
  if (!isRecord(value)) {
    return ''
  }
  for (const key of ['final_response', 'finalResponse', 'text', 'message', 'content', 'summary']) {
    const direct = value[key]
    if (typeof direct === 'string' && direct.trim()) {
      return direct
    }
  }
  const data = value.data
  if (isRecord(data)) {
    const info = data.info
    if (isRecord(info)) {
      const infoText = extractResultText(info)
      if (infoText) {
        return infoText
      }
    }
    const dataText = extractResultText(data)
    if (dataText) {
      return dataText
    }
  }
  const content = value.content
  if (Array.isArray(content)) {
    return content.map(extractResultText).filter(Boolean).join('')
  }
  return ''
}

function readNestedString(value: unknown, path: readonly string[]): string {
  let current: unknown = value
  for (const segment of path) {
    if (!isRecord(current)) {
      return ''
    }
    current = current[segment]
  }
  return typeof current === 'string' ? current : ''
}

function readString(value: unknown, key: string): string {
  return isRecord(value) && typeof value[key] === 'string' ? value[key] : ''
}

function assertTurnParams(value: unknown): TurnParams {
  const record = assertRecord(value)
  const provider = assertRecord(record.provider)
  return {
    streamId: requiredString(record, 'streamId'),
    sessionId: requiredString(record, 'sessionId'),
    engineId: requiredEngineId(requiredString(record, 'engineId')),
    provider: {
      providerId: requiredString(provider, 'providerId'),
      providerType: requiredString(provider, 'providerType'),
      displayName: requiredString(provider, 'displayName'),
      baseURL: typeof provider.baseURL === 'string' ? provider.baseURL : '',
      ...(typeof provider.apiKey === 'string' ? { apiKey: provider.apiKey } : {})
    },
    model: requiredString(record, 'model'),
    prompt: requiredString(record, 'prompt'),
    cwd: requiredString(record, 'cwd'),
    ...(typeof record.engineThreadId === 'string' ? { engineThreadId: record.engineThreadId } : {})
  }
}

function requiredEngineId(value: string): EngineId {
  if (value === 'claude_agent' || value === 'codex' || value === 'opencode') {
    return value
  }
  throw new Error(`unsupported engineId: ${value}`)
}

function requiredString(record: Record<string, unknown>, key: string): string {
  const value = record[key]
  if (typeof value !== 'string' || !value.trim()) {
    throw new Error(`missing ${key}`)
  }
  return value
}

function assertRecord(value: unknown): Record<string, unknown> {
  if (!isRecord(value)) {
    throw new Error('params must be an object')
  }
  return value
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}

function emit(id: string, event: CodingEvent): void {
  write({ id, event })
}

function write(payload: RpcEnvelope): void {
  process.stdout.write(`${JSON.stringify(payload)}\n`)
}

async function healthCheck(): Promise<void> {
  const checks = {
    claude: await checkClaude(),
    codex: await checkImport('@openai/codex-sdk', 'Codex'),
    opencode: await checkImport('@opencode-ai/sdk', 'createOpencode')
  }
  const ok = Object.values(checks).every((check) => check.ok)
  process.stdout.write(`${JSON.stringify({ ok, checks })}\n`)
  if (!ok) {
    process.exitCode = 1
  }
}

async function checkClaude(): Promise<{ ok: boolean; message: string }> {
  const imported = await checkImport('@anthropic-ai/claude-agent-sdk', 'query')
  if (!imported.ok) {
    return imported
  }
  const binary = claudeNativeBinarySubpath()
  if (!binary) {
    return { ok: false, message: `unsupported platform ${process.platform}/${process.arch}` }
  }
  try {
    requireFromHere.resolve(binary)
    return { ok: true, message: 'Claude Agent SDK import and native binary resolved' }
  } catch (error) {
    return {
      ok: false,
      message: error instanceof Error ? error.message : `native binary not found: ${binary}`
    }
  }
}

async function checkImport(pkg: string, exportName: string): Promise<{ ok: boolean; message: string }> {
  try {
    const mod = (await import(pkg)) as Record<string, unknown>
    if (typeof mod[exportName] !== 'function') {
      return { ok: false, message: `${pkg} missing ${exportName} export` }
    }
    return { ok: true, message: `${pkg} resolved` }
  } catch (error) {
    return {
      ok: false,
      message: error instanceof Error ? error.message : `${pkg} import failed`
    }
  }
}

function claudeNativeBinarySubpath(): string {
  const arch = process.arch === 'x64' ? 'x64' : process.arch === 'arm64' ? 'arm64' : ''
  if (!arch) {
    return ''
  }
  if (process.platform === 'win32') {
    return `@anthropic-ai/claude-agent-sdk-win32-${arch}/claude.exe`
  }
  if (process.platform === 'darwin') {
    return `@anthropic-ai/claude-agent-sdk-darwin-${arch}/claude`
  }
  if (process.platform === 'linux') {
    return `@anthropic-ai/claude-agent-sdk-linux-${arch}/claude`
  }
  return ''
}

void main().catch((error) => {
  process.stderr.write(`${error instanceof Error ? error.stack : String(error)}\n`)
  process.exitCode = 1
})

