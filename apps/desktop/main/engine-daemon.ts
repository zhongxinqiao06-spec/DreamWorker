import { spawn, spawnSync, type ChildProcessByStdio } from 'node:child_process'
import { randomBytes } from 'node:crypto'
import { existsSync, readFileSync } from 'node:fs'
import { request as httpRequest, type ClientRequest } from 'node:http'
import { request as httpsRequest } from 'node:https'
import { join, resolve } from 'node:path'
import type { Readable } from 'node:stream'
import type { ChatStreamEvent, RuntimePingResponse } from '../shared/dreamworker-api'

export type EngineReadyMessage = {
  readonly ok: true
  readonly event: 'engine.ready'
  readonly baseUrl: string
  readonly engineVersion: string
  readonly trace_id: string
}

export type EngineLaunchCommand = {
  readonly command: string
  readonly args: readonly string[]
  readonly cwd?: string
  readonly env?: Record<string, string>
}

export type EngineDaemon = {
  readonly token: string
  readonly ready: Promise<EngineReadyMessage>
  readonly ping: () => Promise<RuntimePingResponse>
  readonly request: <T>(
    path: string,
    init?: {
      readonly method?: 'GET' | 'POST'
      readonly body?: unknown
    }
  ) => Promise<T>
  readonly stream: (
    path: string,
    init: {
      readonly body: unknown
      readonly streamId: string
    },
    onEvent: (event: ChatStreamEvent) => void
  ) => Promise<{ readonly streamId: string }>
  readonly cancelStream: (streamId: string) => void
  readonly stop: () => void
}

type FetchResponse = {
  readonly ok: boolean
  readonly status: number
  readonly json: () => Promise<unknown>
}

type FetchLike = (
  url: string,
  init: {
    readonly method?: string
    readonly headers: Record<string, string>
    readonly body?: string
  }
) => Promise<FetchResponse>

type EngineChildProcess = ChildProcessByStdio<null, Readable, Readable>

export const ENGINE_READY_TIMEOUT_MS = 60000
const activeStreamRequests = new Map<string, ClientRequest>()
const cancelledStreamRequests = new Set<string>()

export function createEngineToken(): string {
  return randomBytes(24).toString('hex')
}

export function resolveEngineLaunchCommand(
  token: string,
  rootDir = process.cwd()
): EngineLaunchCommand {
  const projectRoot = resolveProjectRoot(rootDir)
  const runtimeEnv = resolveEngineRuntimeEnv(projectRoot)
  const configuredPath = process.env.DREAMWORKER_ENGINE_PATH
  if (configuredPath) {
    return {
      command: configuredPath,
      args: ['serve', '--token', token],
      env: runtimeEnv
    }
  }

  const sourceLaunchCommand = resolveGoRunEngineLaunchCommand(token, projectRoot, runtimeEnv)
  if (isElectronViteDevSession() && sourceLaunchCommand) {
    return sourceLaunchCommand
  }

  const executableName =
    process.platform === 'win32' ? 'dreamworker-engine.exe' : 'dreamworker-engine'
  const explicitRoot = resolve(rootDir)
  const explicitBinaryPath = join(explicitRoot, 'engine', 'bin', executableName)
  if (existsSync(explicitBinaryPath)) {
    return {
      command: explicitBinaryPath,
      args: ['serve', '--token', token],
      env: resolveEngineRuntimeEnv(explicitRoot)
    }
  }

  const binaryPath = join(projectRoot, 'engine', 'bin', executableName)
  if (existsSync(binaryPath)) {
    return {
      command: binaryPath,
      args: ['serve', '--token', token],
      env: runtimeEnv
    }
  }

  const resourcesPath = process.resourcesPath
  if (resourcesPath) {
    const packagedBinaryPath = join(resourcesPath, 'engine', executableName)
    if (existsSync(packagedBinaryPath)) {
      return {
        command: packagedBinaryPath,
        args: ['serve', '--token', token],
        env: resolveEngineRuntimeEnv(resourcesPath)
      }
    }
  }

  return (
    sourceLaunchCommand ?? {
      command: 'go',
      args: ['run', './cmd/dreamworker-engine', 'serve', '--token', token],
      cwd: join(projectRoot, 'engine'),
      env: runtimeEnv
    }
  )
}

function resolveEngineRuntimeEnv(rootDir: string): Record<string, string> {
  const env = readDotEnvLocal(rootDir)
  const agentDir = join(rootDir, '.agent')
  if (existsSync(agentDir)) {
    env.DREAMWORKER_AGENT_DIR = agentDir
  }
  return env
}

function resolveGoRunEngineLaunchCommand(
  token: string,
  projectRoot: string,
  runtimeEnv: Record<string, string>
): EngineLaunchCommand | null {
  const engineDir = join(projectRoot, 'engine')
  if (!existsSync(join(engineDir, 'go.mod'))) {
    return null
  }

  return {
    command: 'go',
    args: ['run', './cmd/dreamworker-engine', 'serve', '--token', token],
    cwd: engineDir,
    env: runtimeEnv
  }
}

function isElectronViteDevSession(): boolean {
  return Boolean(process.env.ELECTRON_RENDERER_URL)
}

function readDotEnvLocal(rootDir: string): Record<string, string> {
  const envPath = join(rootDir, '.env.local')
  if (!existsSync(envPath)) {
    return {}
  }
  const env: Record<string, string> = {}
  for (const rawLine of readFileSync(envPath, 'utf8').split(/\r?\n/)) {
    const line = rawLine.trim()
    if (!line || line.startsWith('#')) {
      continue
    }
    const separatorIndex = line.indexOf('=')
    if (separatorIndex <= 0) {
      continue
    }
    const key = line.slice(0, separatorIndex).trim()
    let value = line.slice(separatorIndex + 1).trim()
    if (!/^[A-Za-z_][A-Za-z0-9_]*$/.test(key)) {
      continue
    }
    if (
      (value.startsWith('"') && value.endsWith('"')) ||
      (value.startsWith("'") && value.endsWith("'"))
    ) {
      value = value.slice(1, -1)
    }
    env[key] = value
  }
  return env
}

export function resolveProjectRoot(startDir = process.cwd()): string {
  const candidates = [
    process.env.DREAMWORKER_PROJECT_ROOT,
    startDir,
    join(startDir, '..'),
    join(startDir, '..', '..'),
    join(startDir, '..', '..', '..'),
    join(__dirname, '..', '..'),
    join(__dirname, '..', '..', '..'),
    join(__dirname, '..', '..', '..', '..')
  ].filter(
    (candidate): candidate is string => typeof candidate === 'string' && candidate.length > 0
  )

  for (const candidate of candidates) {
    const projectRoot = resolve(candidate)
    if (existsSync(join(projectRoot, 'engine', 'go.mod'))) {
      return projectRoot
    }
  }

  return resolve(startDir)
}

export function startEngineDaemon(
  options: {
    readonly token?: string
    readonly rootDir?: string
    readonly launchCommand?: EngineLaunchCommand
  } = {}
): EngineDaemon {
  const token = options.token ?? createEngineToken()
  const launchCommand = options.launchCommand ?? resolveEngineLaunchCommand(token, options.rootDir)
  const child = spawn(launchCommand.command, [...launchCommand.args], {
    cwd: launchCommand.cwd,
    env: { ...process.env, ...launchCommand.env },
    stdio: ['ignore', 'pipe', 'pipe'],
    windowsHide: true
  })
  const ready = waitForReadyMessage(child)

  return {
    token,
    ready,
    ping: async () => {
      const readyMessage = await ready
      return pingEngineDaemon(readyMessage.baseUrl, token)
    },
    request: async <T>(
      path: string,
      init: {
        readonly method?: 'GET' | 'POST'
        readonly body?: unknown
      } = {}
    ) => {
      const readyMessage = await ready
      return requestEngineDaemon<T>(readyMessage.baseUrl, token, path, init)
    },
    stream: async (path, init, onEvent) => {
      const readyMessage = await ready
      startEngineDaemonStream(readyMessage.baseUrl, token, path, init.body, init.streamId, onEvent)
      return { streamId: init.streamId }
    },
    cancelStream: (streamId: string) => {
      cancelEngineDaemonStream(streamId)
    },
    stop: () => {
      stopEngineChild(child)
    }
  }
}

export function startEngineDaemonStream(
  baseUrl: string,
  token: string,
  path: string,
  body: unknown,
  streamId: string,
  onEvent: (event: ChatStreamEvent) => void
): void {
  const endpoint = new URL(path, baseUrl)
  const payload = JSON.stringify(body)
  const request = (endpoint.protocol === 'https:' ? httpsRequest : httpRequest)(
    endpoint,
    {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${token}`,
        'Content-Type': 'application/json; charset=utf-8',
        Accept: 'text/event-stream',
        'Content-Length': Buffer.byteLength(payload).toString()
      }
    },
    (response) => {
      response.on('end', () => {
        activeStreamRequests.delete(streamId)
      })
      response.on('close', () => {
        activeStreamRequests.delete(streamId)
      })
      if ((response.statusCode ?? 500) < 200 || (response.statusCode ?? 500) >= 300) {
        let errorBody = ''
        response.setEncoding('utf8')
        response.on('data', (chunk: string) => {
          errorBody += chunk
        })
        response.on('end', () => {
          onEvent(createFailedStreamEvent(streamId, `engine stream failed: ${errorBody}`))
        })
        return
      }
      parseSseNodeStream(response, onEvent)
    }
  )
  request.on('error', (error) => {
    activeStreamRequests.delete(streamId)
    if (cancelledStreamRequests.delete(streamId)) {
      return
    }
    onEvent(createFailedStreamEvent(streamId, error.message))
  })
  activeStreamRequests.set(streamId, request)
  request.write(payload)
  request.end()
}

export function cancelEngineDaemonStream(streamId: string): void {
  const request = activeStreamRequests.get(streamId)
  if (!request) {
    return
  }
  activeStreamRequests.delete(streamId)
  cancelledStreamRequests.add(streamId)
  request.destroy()
}

function createFailedStreamEvent(streamId: string, message: string): ChatStreamEvent {
  return {
    type: 'failed',
    streamId,
    sessionId: '',
    messageId: '',
    trace_id: streamId,
    sequence: 0,
    timestamp: new Date().toISOString(),
    error: {
      code: 'ENGINE_STREAM_FAILED',
      message,
      recoverable: true
    }
  }
}

function parseSseNodeStream(stream: Readable, onEvent: (event: ChatStreamEvent) => void): void {
  let buffer = ''
  stream.setEncoding('utf8')
  stream.on('data', (chunk: string) => {
    buffer += chunk
    let boundary = buffer.indexOf('\n\n')
    while (boundary >= 0) {
      const frame = buffer.slice(0, boundary)
      buffer = buffer.slice(boundary + 2)
      emitSseFrame(frame, onEvent)
      boundary = buffer.indexOf('\n\n')
    }
  })
  stream.on('end', () => {
    if (buffer.trim()) {
      emitSseFrame(buffer, onEvent)
    }
  })
}

function emitSseFrame(frame: string, onEvent: (event: ChatStreamEvent) => void): void {
  const data = frame
    .split(/\r?\n/)
    .filter((line) => line.startsWith('data:'))
    .map((line) => line.slice(5).trim())
    .join('\n')
  if (!data) {
    return
  }
  try {
    onEvent(JSON.parse(data) as ChatStreamEvent)
  } catch {
    // Ignore malformed local stream frames.
  }
}

export async function pingEngineDaemon(
  baseUrl: string,
  token: string,
  fetchLike: FetchLike = globalThis.fetch as FetchLike
): Promise<RuntimePingResponse> {
  const response = await fetchLike(`${baseUrl}/runtime/ping`, {
    headers: {
      Authorization: `Bearer ${token}`
    }
  })

  if (!response.ok) {
    throw new Error(`engine ping failed with status ${response.status}`)
  }

  return parseRuntimePingResponse(await response.json())
}

export async function requestEngineDaemon<T>(
  baseUrl: string,
  token: string,
  path: string,
  init: {
    readonly method?: 'GET' | 'POST'
    readonly body?: unknown
  } = {},
  fetchLike: FetchLike = globalThis.fetch as FetchLike
): Promise<T> {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${token}`
  }
  const requestInit: {
    method?: string
    headers: Record<string, string>
    body?: string
  } = {
    method: init.method ?? 'GET',
    headers
  }

  if (init.body !== undefined) {
    headers['Content-Type'] = 'application/json; charset=utf-8'
    requestInit.body = JSON.stringify(init.body)
  }

  const response = await fetchLike(`${baseUrl}${path}`, requestInit)
  if (!response.ok) {
    throw new Error(`engine request failed with status ${response.status}`)
  }

  return (await response.json()) as T
}

export function parseEngineReadyLine(line: string): EngineReadyMessage | null {
  let payload: unknown
  try {
    payload = JSON.parse(line)
  } catch {
    return null
  }

  if (!isRecord(payload)) {
    return null
  }

  if (
    payload.ok === true &&
    payload.event === 'engine.ready' &&
    typeof payload.baseUrl === 'string' &&
    typeof payload.engineVersion === 'string' &&
    typeof payload.trace_id === 'string'
  ) {
    return {
      ok: true,
      event: 'engine.ready',
      baseUrl: payload.baseUrl,
      engineVersion: payload.engineVersion,
      trace_id: payload.trace_id
    }
  }

  return null
}

function waitForReadyMessage(child: EngineChildProcess): Promise<EngineReadyMessage> {
  return new Promise((resolve, reject) => {
    let settled = false
    let stdoutBuffer = ''
    let stderrBuffer = ''
    const timeout = setTimeout(() => {
      if (!settled) {
        settled = true
        child.kill()
        reject(new Error('engine ready timeout'))
      }
    }, ENGINE_READY_TIMEOUT_MS)

    const settleReady = (readyMessage: EngineReadyMessage): void => {
      if (!settled) {
        settled = true
        clearTimeout(timeout)
        resolve(readyMessage)
      }
    }

    const settleError = (message: string): void => {
      if (!settled) {
        settled = true
        clearTimeout(timeout)
        reject(new Error(message))
      }
    }

    child.stdout.on('data', (chunk: Buffer) => {
      stdoutBuffer += chunk.toString('utf8')
      const lines = stdoutBuffer.split(/\r?\n/)
      stdoutBuffer = lines.pop() ?? ''
      for (const line of lines) {
        const readyMessage = parseEngineReadyLine(line)
        if (readyMessage) {
          settleReady(readyMessage)
          return
        }
      }
    })

    child.stderr.on('data', (chunk: Buffer) => {
      stderrBuffer += chunk.toString('utf8')
    })

    child.on('error', (error) => {
      settleError(`engine process failed: ${error.message}`)
    })

    child.on('exit', (code) => {
      if (!settled) {
        settleError(`engine exited before ready: ${code ?? 'unknown'} ${stderrBuffer.trim()}`)
      }
    })
  })
}

function parseRuntimePingResponse(payload: unknown): RuntimePingResponse {
  if (!isRecord(payload)) {
    throw new Error('engine ping response is not an object')
  }

  if (
    payload.schema_version === '0.1' &&
    payload.ok === true &&
    typeof payload.engineVersion === 'string' &&
    typeof payload.trace_id === 'string'
  ) {
    return {
      schema_version: '0.1',
      ok: true,
      engineVersion: payload.engineVersion,
      trace_id: payload.trace_id
    }
  }

  throw new Error('engine ping response is invalid')
}

function stopEngineChild(child: EngineChildProcess): void {
  if (child.killed) {
    return
  }

  if (process.platform === 'win32' && child.pid) {
    spawnSync('taskkill.exe', ['/PID', String(child.pid), '/T', '/F'], {
      stdio: 'ignore',
      windowsHide: true
    })
    return
  }

  child.kill()
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}
