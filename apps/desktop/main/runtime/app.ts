import { CodingService } from './coding/coding-service'
import { engineVersion, runtimePing } from './runtime-info'
import { WorkspaceStore } from './store/workspace-store'
import { RuntimeAppError, badRequest, type CodingStreamEvent, type JsonRecord } from './types'
import { asRecord, asString, newTraceId, nowISO } from './shared/util'
import type { ChatStreamEvent, RuntimePingResponse } from '../../shared/dreamworker-api'

type RuntimeRoute = {
  readonly method: 'GET' | 'POST'
  readonly path: string
  readonly handler: (body: JsonRecord) => unknown | Promise<unknown>
}

export type RuntimeRequestInit = {
  readonly method?: 'GET' | 'POST'
  readonly body?: unknown
}

export type RuntimeStreamInit = {
  readonly body: unknown
  readonly streamId: string
}

export type RuntimeStreamEvent = ChatStreamEvent | CodingStreamEvent

export class DreamWorkerRuntime {
  private readonly store: WorkspaceStore
  private readonly coding: CodingService
  private readonly routes: readonly RuntimeRoute[]

  constructor(configDir?: string) {
    this.store = new WorkspaceStore(configDir)
    this.coding = new CodingService(this.store)
    this.routes = createRoutes(this.store, this.coding)
  }

  ping(): RuntimePingResponse {
    return runtimePing() as unknown as RuntimePingResponse
  }

  async request<T>(path: string, init: RuntimeRequestInit = {}): Promise<T> {
    const method = init.method ?? 'GET'
    const route = this.routes.find(
      (candidate) => candidate.path === path && candidate.method === method
    )
    if (!route) {
      throw badRequest(
        'ROUTE_NOT_FOUND',
        `Runtime route not found: ${method} ${path}`,
        '刷新应用后重试。'
      )
    }
    const body = method === 'POST' ? asRecord(init.body ?? {}) : {}
    return (await route.handler(body)) as T
  }

  async stream(
    path: string,
    init: RuntimeStreamInit,
    onEvent: (event: RuntimeStreamEvent) => void
  ): Promise<{ readonly streamId: string }> {
    if (path === '/chat/messages/stream') {
      for await (const event of this.chatStream(asRecord(init.body))) {
        onEvent(event)
      }
      return { streamId: init.streamId }
    }
    if (path === '/coding/turns/stream') {
      for await (const event of this.coding.streamTurn(asRecord(init.body))) {
        onEvent(event)
      }
      return { streamId: init.streamId }
    }
    throw badRequest(
      'STREAM_ROUTE_NOT_FOUND',
      `Runtime stream route not found: ${path}`,
      '刷新应用后重试。'
    )
  }

  cancelStream(streamId: string): void {
    try {
      this.coding.cancelTurn({ streamId })
    } catch {
      // The IPC cancel route also performs a typed cancel. This hook is best-effort cleanup.
    }
  }

  stop(): void {
    this.coding.dispose()
    this.store.close()
  }

  private async *chatStream(body: JsonRecord): AsyncGenerator<ChatStreamEvent> {
    const streamId = asString(body.streamId) || `stream_${Date.now()}`
    const sessionId = asString(body.sessionId)
    const traceId = newTraceId()
    const started: ChatStreamEvent = {
      type: 'started',
      streamId,
      sessionId,
      messageId: '',
      trace_id: traceId,
      sequence: 1,
      timestamp: nowISO()
    }
    yield started
    yield {
      ...started,
      type: 'token_delta',
      sequence: 2,
      timestamp: nowISO(),
      delta: 'Main Runtime 已在 Main 进程内接管 Chat stream；模型网关接入后会输出真实模型增量。'
    }

    let result: unknown
    try {
      result = this.store.sendChatMessage({
        ...body,
        content: asString(body.content) || asString(body.prompt)
      })
    } catch {
      result = undefined
    }
    const completed: ChatStreamEvent = {
      ...started,
      type: 'completed',
      sequence: 3,
      timestamp: nowISO()
    }
    if (result) {
      yield {
        ...completed,
        result: result as NonNullable<ChatStreamEvent['result']>
      }
      return
    }
    yield completed
  }
}

export function createDreamWorkerRuntime(configDir?: string): DreamWorkerRuntime {
  return new DreamWorkerRuntime(configDir)
}

export function runtimeErrorToJson(error: unknown): JsonRecord {
  if (error instanceof RuntimeAppError) {
    return {
      code: error.code,
      message: error.message,
      recoverable: error.status < 500,
      user_action: error.userAction,
      trace_id: newTraceId()
    }
  }
  const message = error instanceof Error ? error.message : 'Runtime request failed.'
  return {
    code: 'RUNTIME_REQUEST_FAILED',
    message,
    recoverable: true,
    user_action: '请查看本地日志后重试。',
    trace_id: newTraceId()
  }
}

export function runtimeSummary(): JsonRecord {
  return {
    engineVersion,
    runtime: 'desktop-main-runtime',
    transport: 'direct-ipc'
  }
}

function createRoutes(store: WorkspaceStore, coding: CodingService): RuntimeRoute[] {
  return [
    get('/models/providers', () => store.listProviders()),
    post('/models/providers/save', (body) => store.saveProvider(body)),
    post('/models/providers/delete', (body) => store.deleteProvider(asString(body.providerId))),
    post('/models/providers/test', (body) => store.testProvider(asString(body.providerId))),
    post('/models/providers/refresh-models', (body) =>
      store.refreshProviderModels(asString(body.providerId))
    ),
    get('/models/profiles', () => store.listProfiles()),
    post('/models/profiles/save', (body) => store.saveProfile(body)),
    post('/models/profiles/delete', (body) => store.deleteProfile(asString(body.profileId))),
    get('/settings', () => store.getSettings()),
    post('/settings/update', (body) => store.updateSettings(body)),
    post('/settings/reset-extension', () => store.resetExtensionSettings()),
    get('/extensions', () => store.listExtensions()),
    post('/extensions/status', (body) => store.extensionStatus(asString(body.extensionId))),
    post('/extensions/detect', (body) => store.extensionAction(body, '检测')),
    post('/extensions/install', (body) => store.extensionAction(body, '安装')),
    post('/extensions/start', (body) => store.extensionAction(body, '启动')),
    post('/extensions/stop', (body) => store.extensionAction(body, '停止')),
    post('/extensions/restart', (body) => store.extensionAction(body, '重启')),
    post('/extensions/test', (body) => store.extensionAction(body, '测试')),
    post('/extensions/refresh-models', (body) => ({
      ok: true,
      extensionId: asString(body.extensionId) || '9router',
      models: [],
      status: store.extensionStatus(asString(body.extensionId))
    })),
    post('/extensions/verify-streaming', (body) => ({
      ok: true,
      extensionId: asString(body.extensionId) || '9router',
      message: 'Main 内嵌 Runtime stream bridge 可用。',
      latencyMs: 0,
      status: store.extensionStatus(asString(body.extensionId))
    })),
    post('/extensions/logs/tail', () => []),
    post('/extensions/logs/clear', (body) => store.extensionAction(body, '清理日志')),
    get('/agents', () => store.listAgents()),
    post('/agents/get', (body) => store.getAgent(asString(body.agentId))),
    post('/agents/save', (body) => store.saveAgent(body)),
    post('/agents/duplicate', (body) => store.duplicateAgent(asString(body.agentId))),
    post('/agents/delete', (body) => store.deleteAgent(asString(body.agentId))),
    get('/skills', () => store.listSkills()),
    post('/skills/get', (body) => store.getSkill(asString(body.skillId))),
    post('/skills/save', (body) => store.saveSkill(body)),
    post('/skills/delete', (body) => store.deleteSkill(asString(body.skillId))),
    get('/tools', () => store.listTools()),
    post('/tools/get', (body) => store.getTool(asString(body.toolId))),
    post('/tools/save', (body) => store.saveTool(body)),
    post('/tools/set-enabled', (body) =>
      store.setToolEnabled(asString(body.toolId), body.enabled === true)
    ),
    post('/tools/delete', (body) => store.deleteTool(asString(body.toolId))),
    get('/mcp/servers', () => store.listMcpServers()),
    post('/mcp/servers/save', (body) => store.saveMcpServer(body)),
    post('/mcp/servers/delete', (body) => store.deleteMcpServer(asString(body.serverId))),
    post('/mcp/servers/test', (body) => store.testMcpServer(asString(body.serverId))),
    post('/mcp/servers/refresh-tools', (body) => store.refreshMcpTools(asString(body.serverId))),
    get('/projects', () => store.listProjects()),
    post('/projects/create', (body) => store.createProject(body)),
    post('/projects/get', (body) => store.getProject(asString(body.projectId))),
    post('/projects/update', (body) => store.updateProject(body)),
    post('/projects/delete', (body) => store.deleteProject(asString(body.projectId))),
    post('/projects/local-directory/validate', (body) =>
      store.validateLocalDirectory(asString(body.projectId))
    ),
    post('/projects/local-directory/initialize', (body) =>
      store.initializeLocalDirectory(asString(body.projectId))
    ),
    post('/projects/export-manifest', (body) =>
      store.exportProjectManifest(asString(body.projectId))
    ),
    post('/projects/modules', (body) => store.listProjectModules(asString(body.projectId))),
    post('/projects/modules/get', (body) => store.getProjectModule(body)),
    post('/projects/modules/update-config', (body) => store.updateProjectModuleConfig(body)),
    post('/projects/requirements/import-files', (body) => store.importRequirementFiles(body)),
    post('/projects/requirements/sources', (body) =>
      store.listRequirementSources(asString(body.projectId))
    ),
    post('/projects/requirements/preview-source', (body) => store.previewRequirementSource(body)),
    post('/projects/requirements/run', (body) => store.runRequirementAnalysis(body)),
    get('/chat/sessions', () => store.listChatSessions()),
    post('/chat/sessions/create', (body) => store.createChatSession(body)),
    post('/chat/sessions/update', (body) => store.updateChatSession(body)),
    post('/chat/messages', (body) => store.listChatMessages(asString(body.sessionId))),
    post('/chat/messages/send', (body) => store.sendChatMessage(body)),
    post('/chat/messages/cancel', (body) => ({ ok: true, deletedId: asString(body.streamId) })),
    post('/chat/images/generate', (body) =>
      store.sendChatMessage({ ...body, content: asString(body.prompt) || 'generate image' })
    ),
    post('/chat/sessions/delete', (body) => store.deleteChatSession(asString(body.sessionId))),
    get('/coding/engines', () => coding.listEngines()),
    post('/coding/sessions/create', (body) => coding.createSession(body)),
    post('/coding/sessions/get', (body) => coding.getSession(body)),
    post('/coding/files/list', (body) => coding.listFiles(body)),
    post('/coding/files/read', (body) => coding.readFile(body)),
    post('/coding/files/status', (body) => coding.fileStatus(body)),
    post('/coding/turns/cancel', (body) => coding.cancelTurn(body))
  ]
}

function get(path: string, handler: RuntimeRoute['handler']): RuntimeRoute {
  return { method: 'GET', path, handler }
}

function post(path: string, handler: RuntimeRoute['handler']): RuntimeRoute {
  return { method: 'POST', path, handler }
}
