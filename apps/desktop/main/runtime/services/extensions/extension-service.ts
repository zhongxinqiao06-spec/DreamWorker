import { spawn, spawnSync, type ChildProcess } from 'node:child_process'
import { existsSync, mkdirSync } from 'node:fs'
import { createRequire } from 'node:module'
import { dirname, join, resolve } from 'node:path'
import { asString, nowISO } from '../../shared/util'
import type { JsonRecord } from '../../types'
import type { SettingsService } from '../settings/settings-service'

const nineRouterExtensionId = 'extension_9router'
const require = createRequire(import.meta.url)

type NineRouterCommand = {
  command: string
  args: string[]
  displayPath: string
  env: NodeJS.ProcessEnv
  mode: 'server' | 'cli'
}

type ExtensionLogLine = {
  timestamp: string
  stream: 'stdout' | 'stderr' | 'system'
  line: string
}

type NineRouterHealth = {
  connected: boolean
  dashboardConnected: boolean
  modelsConnected: boolean
  models: string[]
  errorCode?: string
  message?: string
  modelErrorCode?: string
  modelErrorMessage?: string
}

export class ExtensionService {
  private process: ChildProcess | null = null
  private startedAt = ''
  private stoppedAt = ''
  private logs: ExtensionLogLine[] = []

  constructor(
    private readonly settings: SettingsService,
    private readonly configDir: string
  ) {}

  listExtensions(): JsonRecord[] {
    const settings = this.settings.getSettings()
    return [
      {
        extensionId: nineRouterExtensionId,
        name: '9Router',
        kind: 'model_router',
        runtimeKind: 'node',
        description: '本地 OpenAI-compatible 模型路由。',
        install: {
          packageName: settings.nineRouterManagedPackageName,
          packageVersion: settings.nineRouterManagedInstallVersion,
          runtimeDir: '',
          logDir: '',
          configDir: this.configDir
        },
        process: {
          defaultCommand: settings.nineRouterManagedCommand,
          defaultArgs: [],
          port: dashboardPort(asString(settings.nineRouterDashboardURL)) ?? 20128,
          env: []
        },
        health: {
          dashboardURL: settings.nineRouterDashboardURL,
          baseURL: settings.nineRouterBaseURL,
          modelsPath: '/models',
          chatPath: '/chat/completions'
        },
        providerBridge: {
          providerId: 'provider_9router_local',
          providerType: 'openai_compatible',
          displayName: '9Router 本地模型路由',
          baseURL: settings.nineRouterBaseURL,
          defaultModel: settings.nineRouterDefaultModel,
          sortOrder: 10,
          systemPreset: true,
          allowDeletion: false
        },
        capabilities: ['model_routing', 'openai_compatible'],
        security: {
          riskLevel: 'medium',
          allowedHosts: ['127.0.0.1', 'localhost'],
          secretKeys: [],
          envAllowList: [],
          managedRequiresExplicitEnable: true
        },
        systemPreset: true,
        enabled: settings.enableNineRouterIntegration
      }
    ]
  }

  async extensionStatus(extensionId = nineRouterExtensionId): Promise<JsonRecord> {
    const settings = this.settings.getSettings()
    const nodeVersion = process.version
    const command = asString(settings.nineRouterManagedCommand) || '9router'
    const commandInfo = resolveNineRouterCommand(command)
    const npmVersion = commandVersion('npm')
    const health = await probeNineRouter(
      asString(settings.nineRouterBaseURL),
      asString(settings.nineRouterDashboardURL)
    )
    const running = this.process !== null && this.process.exitCode === null
    const normalizedExtensionId = normalizeExtensionId(extensionId)
    return {
      extensionId: normalizedExtensionId,
      installed: Boolean(commandInfo),
      installSource: commandInfo ? 'bundled_or_path' : 'missing',
      nodeAvailable: true,
      npmAvailable: npmVersion !== '',
      nodeVersion,
      npmVersion,
      command,
      runMode: settings.nineRouterRunMode ?? 'managed',
      processState: health.connected ? 'running' : running ? 'starting' : 'stopped',
      pid: running ? this.process?.pid : undefined,
      startedByDreamWorker: running,
      baseURL: settings.nineRouterBaseURL,
      dashboardURL: settings.nineRouterDashboardURL,
      dashboardConnected: health.dashboardConnected,
      modelsConnected: health.modelsConnected,
      healthStatus: health.connected ? 'connected' : 'disconnected',
      modelCount: health.models.length,
      models: health.models,
      streamingVerified: false,
      hasApiKey: false,
      logDir: '',
      workDir: this.runtimeDir(),
      lastStartedAt: this.startedAt || undefined,
      lastStoppedAt: this.stoppedAt || undefined,
      lastCheckedAt: nowISO(),
      lastErrorCode: health.connected ? undefined : health.errorCode,
      lastErrorMessage: health.connected ? undefined : health.message,
      modelErrorCode: health.modelErrorCode,
      modelErrorMessage: health.modelErrorMessage,
      runtime: {
        nodeAvailable: true,
        npmAvailable: npmVersion !== '',
        nodeVersion,
        npmVersion,
        commandAvailable: Boolean(commandInfo),
        command,
        managedLocalBin: commandInfo?.displayPath,
        installSource: commandInfo ? 'bundled_or_path' : 'missing',
        lastErrorCode: commandInfo ? undefined : 'COMMAND_NOT_FOUND',
        lastErrorMessage: commandInfo ? undefined : `未找到 9Router 命令：${command}`
      }
    }
  }

  async extensionAction(input: JsonRecord, verb: string): Promise<JsonRecord> {
    const extensionId = normalizeExtensionId(asString(input.extensionId))
    if (verb === '安装') {
      const command = asString(this.settings.getSettings().nineRouterManagedCommand) || '9router'
      const commandInfo = resolveNineRouterCommand(command)
      if (commandInfo) {
        this.appendLog('system', `内置 9Router 已就绪：${commandInfo.displayPath}`)
        return {
          ok: true,
          extensionId,
          message: '内置 9Router 已就绪，可以直接启动。',
          status: await this.extensionStatus(extensionId)
        }
      }
      this.appendLog('system', '未找到内置或系统 9Router 命令。')
      return {
        ok: false,
        extensionId,
        message: '未找到 9Router：请确认 desktop 依赖已安装，或切换到外部服务模式。',
        status: await this.extensionStatus(extensionId)
      }
    }
    if (verb === '启动') {
      return this.start(extensionId)
    }
    if (verb === '停止') {
      return this.stop(extensionId)
    }
    if (verb === '重启') {
      await this.stop(extensionId)
      return this.start(extensionId)
    }
    if (verb === '检测' || verb === '测试') {
      const status = await this.extensionStatus(extensionId)
      return {
        ok: status.healthStatus === 'connected',
        extensionId,
        message:
          status.healthStatus === 'connected'
            ? '9Router 控制台和模型接口已连接。'
            : status.lastErrorMessage || '9Router 服务未连接。',
        status
      }
    }
    return {
      ok: true,
      extensionId,
      message: `Main Runtime handled ${verb}.`,
      status: await this.extensionStatus(extensionId)
    }
  }

  async refreshModels(input: JsonRecord): Promise<JsonRecord> {
    const extensionId = normalizeExtensionId(asString(input.extensionId))
    const status = await this.extensionStatus(extensionId)
    return {
      ok: status.healthStatus === 'connected',
      extensionId,
      models: status.models,
      status
    }
  }

  async verifyStreaming(input: JsonRecord): Promise<JsonRecord> {
    const extensionId = normalizeExtensionId(asString(input.extensionId))
    const status = await this.extensionStatus(extensionId)
    return {
      ok: status.healthStatus === 'connected',
      extensionId,
      message:
        status.healthStatus === 'connected'
          ? 'Main 内嵌 Runtime stream bridge 可用。'
          : status.lastErrorMessage || '9Router 服务未连接，无法验证流式输出。',
      latencyMs: 0,
      status
    }
  }

  tailLogs(): JsonRecord[] {
    return [...this.logs.slice(-160)]
  }

  stopManagedProcess(): void {
    if (this.process && this.process.exitCode === null) {
      this.process.kill()
      this.appendLog('system', 'Runtime 退出时已停止 9Router。')
    }
    this.process = null
    this.stoppedAt = nowISO()
  }

  private async start(extensionId: string): Promise<JsonRecord> {
    const settings = this.settings.getSettings()
    const command = asString(settings.nineRouterManagedCommand) || '9router'
    const commandInfo = resolveNineRouterCommand(command)
    if (!commandInfo) {
      this.appendLog('system', `未找到 9Router 命令：${command}`)
      return {
        ok: false,
        extensionId,
        message: `未找到 9Router 命令：${command}`,
        status: await this.extensionStatus(extensionId)
      }
    }
    if (this.process && this.process.exitCode === null) {
      return {
        ok: true,
        extensionId,
        message: '9Router 已由 DreamWorker 启动。',
        status: await this.extensionStatus(extensionId)
      }
    }
    const existingHealth = await probeNineRouter(
      asString(settings.nineRouterBaseURL),
      asString(settings.nineRouterDashboardURL)
    )
    if (existingHealth.connected) {
      this.appendLog('system', '9Router 已在当前地址运行，复用现有服务。')
      return {
        ok: true,
        extensionId,
        message: '9Router 已在当前地址运行，已复用现有服务。',
        status: await this.extensionStatus(extensionId)
      }
    }
    const cwd = this.runtimeDir()
    mkdirSync(cwd, { recursive: true })
    const port = dashboardPort(asString(settings.nineRouterDashboardURL)) ?? 20128
    const startArgs =
      commandInfo.mode === 'server'
        ? commandInfo.args
        : [
            ...commandInfo.args,
            '--tray',
            '--skip-update',
            '--host',
            '127.0.0.1',
            '--port',
            String(port),
            '--no-browser'
          ]
    const env =
      commandInfo.mode === 'server'
        ? {
            ...commandInfo.env,
            PORT: String(port),
            HOSTNAME: '127.0.0.1'
          }
        : commandInfo.env
    const processCwd = commandInfo.mode === 'server' ? dirname(commandInfo.displayPath) : cwd
    this.process = spawn(commandInfo.command, startArgs, {
      cwd: processCwd,
      env,
      stdio: ['ignore', 'pipe', 'pipe'],
      windowsHide: true
    })
    this.startedAt = nowISO()
    this.appendLog('system', `启动 9Router：${commandInfo.displayPath} ${startArgs.join(' ')}`)
    this.process.stdout?.on('data', (chunk) => this.appendLog('stdout', String(chunk)))
    this.process.stderr?.on('data', (chunk) => this.appendLog('stderr', String(chunk)))
    this.process.once('exit', (code) => {
      this.stoppedAt = nowISO()
      this.appendLog('system', `9Router 进程退出：${code ?? 'unknown'}`)
    })
    await waitForNineRouter(
      asString(settings.nineRouterBaseURL),
      asString(settings.nineRouterDashboardURL),
      managedTimeoutMs(settings.nineRouterManagedTimeoutMs)
    )
    const status = await this.extensionStatus(extensionId)
    return {
      ok: status.healthStatus === 'connected' || status.processState === 'starting',
      extensionId,
      message:
        status.healthStatus === 'connected'
          ? '9Router 已启动并连接。'
          : '9Router 已尝试启动，控制台仍在等待服务就绪。',
      status
    }
  }

  private async stop(extensionId: string): Promise<JsonRecord> {
    this.stopManagedProcess()
    return {
      ok: true,
      extensionId,
      message: '9Router 已停止。',
      status: await this.extensionStatus(extensionId)
    }
  }

  private runtimeDir(): string {
    const configured = asString(this.settings.getSettings().nineRouterManagedWorkDir)
    return configured || join(this.configDir, 'extensions', '9router')
  }

  private appendLog(stream: ExtensionLogLine['stream'], text: string): void {
    for (const line of text.split(/\r?\n/).filter(Boolean)) {
      this.logs.push({ timestamp: nowISO(), stream, line })
    }
    this.logs = this.logs.slice(-400)
  }
}

function normalizeExtensionId(extensionId: string): string {
  return extensionId === '9router' || !extensionId ? nineRouterExtensionId : extensionId
}

function resolveNineRouterCommand(command: string): NineRouterCommand | null {
  const bundled = resolveBundledNineRouter()
  if (bundled) {
    return bundled
  }
  const commandPath = resolveCommand(command)
  if (!commandPath) {
    return null
  }
  if (process.platform === 'win32' && commandPath.toLowerCase().endsWith('.cmd')) {
    return {
      command: 'cmd.exe',
      args: ['/d', '/s', '/c', `"${commandPath}"`],
      displayPath: commandPath,
      env: process.env,
      mode: 'cli'
    }
  }
  return { command: commandPath, args: [], displayPath: commandPath, env: process.env, mode: 'cli' }
}

function resolveBundledNineRouter(): NineRouterCommand | null {
  try {
    const packageDir = dirname(require.resolve('9router/package.json'))
    const appDir = join(packageDir, 'app')
    const customServerPath = join(appDir, 'custom-server.js')
    const serverPath = existsSync(customServerPath) ? customServerPath : join(appDir, 'server.js')
    if (!existsSync(serverPath)) {
      return null
    }
    const env = runtimeEnvForBundledNineRouter(packageDir)
    return {
      command: process.execPath,
      args: ['--max-old-space-size=6144', serverPath],
      displayPath: serverPath,
      env,
      mode: 'server'
    }
  } catch {
    return null
  }
}

function runtimeEnvForBundledNineRouter(packageDir: string): NodeJS.ProcessEnv {
  const baseEnv = { ...process.env, ELECTRON_RUN_AS_NODE: '1' }
  try {
    const sqliteRuntime = require(join(packageDir, 'hooks', 'sqliteRuntime.js')) as {
      ensureSqliteRuntime?: (options?: { silent?: boolean }) => void
      buildEnvWithRuntime?: (env: NodeJS.ProcessEnv) => NodeJS.ProcessEnv
    }
    sqliteRuntime.ensureSqliteRuntime?.({ silent: true })
    return sqliteRuntime.buildEnvWithRuntime?.(baseEnv) ?? baseEnv
  } catch {
    return baseEnv
  }
}

function resolveCommand(command: string): string {
  if (!command) {
    return ''
  }
  if (existsSync(command)) {
    return resolve(command)
  }
  const where = process.platform === 'win32' ? 'where.exe' : 'which'
  const result = spawnSync(where, [command], { encoding: 'utf8' })
  return result.status === 0 ? result.stdout.split(/\r?\n/)[0]?.trim() || '' : ''
}

function commandVersion(command: string): string {
  const executable =
    process.platform === 'win32' && !command.endsWith('.cmd') ? `${command}.cmd` : command
  const result = spawnSync(executable, ['--version'], { encoding: 'utf8' })
  return result.status === 0 ? result.stdout.trim() || result.stderr.trim() : ''
}

function dashboardPort(value: string): number | null {
  try {
    const url = new URL(value)
    return url.port ? Number(url.port) : null
  } catch {
    return null
  }
}

async function probeNineRouter(baseURL: string, dashboardURL: string): Promise<NineRouterHealth> {
  const dashboardEndpoint = dashboardEndpointFor(dashboardURL)
  const modelEndpointURL = modelEndpoint(baseURL)
  const [dashboard, models] = await Promise.all([
    probeHttpOk(dashboardEndpoint, 2500),
    probeModels(modelEndpointURL, 2500)
  ])
  const connected = dashboard.ok || models.ok
  const health: NineRouterHealth = {
    connected,
    dashboardConnected: dashboard.ok,
    modelsConnected: models.ok,
    models: models.models
  }
  if (!connected) {
    const errorCode = dashboard.errorCode || models.errorCode
    if (errorCode) {
      health.errorCode = errorCode
    }
    health.message = dashboard.message || models.message || '9Router 服务未连接。'
  }
  if (!models.ok) {
    if (models.errorCode) {
      health.modelErrorCode = models.errorCode
    }
    if (models.message) {
      health.modelErrorMessage = models.message
    }
  }
  return health
}

async function waitForNineRouter(
  baseURL: string,
  dashboardURL: string,
  timeoutMs: number
): Promise<NineRouterHealth> {
  const startedAt = Date.now()
  let lastHealth = await probeNineRouter(baseURL, dashboardURL)
  while (!lastHealth.connected && Date.now() - startedAt < timeoutMs) {
    await sleep(400)
    lastHealth = await probeNineRouter(baseURL, dashboardURL)
  }
  return lastHealth
}

async function probeHttpOk(
  endpoint: string,
  timeoutMs: number
): Promise<{ ok: boolean; errorCode?: string; message?: string }> {
  const controller = new AbortController()
  const timeout = setTimeout(() => controller.abort(), timeoutMs)
  try {
    const response = await fetch(endpoint, { signal: controller.signal })
    if (response.ok) {
      return { ok: true }
    }
    return {
      ok: false,
      errorCode: `HTTP_${response.status}`,
      message: `9Router 控制台返回 HTTP ${response.status}：${endpoint}`
    }
  } catch (error) {
    return {
      ok: false,
      errorCode: 'DASHBOARD_UNREACHABLE',
      message: error instanceof Error ? error.message : String(error)
    }
  } finally {
    clearTimeout(timeout)
  }
}

async function probeModels(
  endpoint: string,
  timeoutMs: number
): Promise<{ ok: boolean; models: string[]; errorCode?: string; message?: string }> {
  const controller = new AbortController()
  const timeout = setTimeout(() => controller.abort(), timeoutMs)
  try {
    const response = await fetch(endpoint, { signal: controller.signal })
    if (!response.ok) {
      return {
        ok: false,
        models: [],
        errorCode: `HTTP_${response.status}`,
        message: `9Router 模型接口返回 HTTP ${response.status}：${endpoint}`
      }
    }
    const payload = (await response.json()) as JsonRecord
    const data = Array.isArray(payload.data) ? payload.data : []
    return {
      ok: true,
      models: data
        .map((item) => (typeof item === 'object' && item ? asString((item as JsonRecord).id) : ''))
        .filter(Boolean)
    }
  } catch (error) {
    return {
      ok: false,
      models: [],
      errorCode: 'MODELS_UNREACHABLE',
      message: error instanceof Error ? error.message : String(error)
    }
  } finally {
    clearTimeout(timeout)
  }
}

function managedTimeoutMs(value: unknown): number {
  const parsed = Number(value)
  return Number.isFinite(parsed) && parsed >= 1000 ? parsed : 15000
}

function dashboardEndpointFor(value: string): string {
  try {
    const url = new URL(value.trim() || 'http://127.0.0.1:20128')
    if (url.pathname === '' || url.pathname === '/') {
      url.pathname = '/dashboard'
    }
    return url.toString()
  } catch {
    return 'http://127.0.0.1:20128/dashboard'
  }
}

function modelEndpoint(baseURL: string): string {
  try {
    const url = new URL(baseURL)
    const pathname = url.pathname.replace(/\/$/, '')
    url.pathname = pathname.endsWith('/v1') ? `${pathname}/models` : `${pathname}/v1/models`
    return url.toString()
  } catch {
    return 'http://127.0.0.1:20128/v1/models'
  }
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolveSleep) => setTimeout(resolveSleep, ms))
}
