import { dialog, ipcMain, shell } from 'electron'
import nodePath from 'node:path'
import {
  CHANNELS,
  type ChatStreamEvent,
  type CodingStreamEvent,
  type ProjectDirectoryCheck,
  type ProjectLocalDirectoryActionResult,
  type RequirementImportResult,
  type RuntimePingResponse
} from '../shared/dreamworker-api'
import { openExternalHttpUrl } from './external-url'
import { createRuntimePingStubResponse } from './runtime-ping'

export type RuntimePingProvider = () => Promise<RuntimePingResponse> | RuntimePingResponse

export type RuntimeRequestProvider = <T>(
  path: string,
  init?: {
    readonly method?: 'GET' | 'POST'
    readonly body?: unknown
  }
) => Promise<T>

export type RuntimeStreamProvider = (
  path: string,
  init: {
    readonly body: unknown
    readonly streamId: string
  },
  onEvent: (event: ChatStreamEvent | CodingStreamEvent) => void
) => Promise<{ readonly streamId: string }>

type RuntimeRoute = {
  readonly channel: string
  readonly path: string
  readonly method: 'GET' | 'POST'
}

const RUNTIME_ROUTES: readonly RuntimeRoute[] = [
  { channel: CHANNELS.modelsListProviders, path: '/models/providers', method: 'GET' },
  { channel: CHANNELS.modelsSaveProvider, path: '/models/providers/save', method: 'POST' },
  { channel: CHANNELS.modelsDeleteProvider, path: '/models/providers/delete', method: 'POST' },
  { channel: CHANNELS.modelsTestProvider, path: '/models/providers/test', method: 'POST' },
  {
    channel: CHANNELS.modelsRefreshProviderModels,
    path: '/models/providers/refresh-models',
    method: 'POST'
  },
  { channel: CHANNELS.modelsListProfiles, path: '/models/profiles', method: 'GET' },
  { channel: CHANNELS.modelsSaveProfile, path: '/models/profiles/save', method: 'POST' },
  { channel: CHANNELS.modelsDeleteProfile, path: '/models/profiles/delete', method: 'POST' },
  { channel: CHANNELS.settingsGet, path: '/settings', method: 'GET' },
  { channel: CHANNELS.settingsUpdate, path: '/settings/update', method: 'POST' },
  { channel: CHANNELS.settingsResetExtension, path: '/settings/reset-extension', method: 'POST' },
  { channel: CHANNELS.extensionsList, path: '/extensions', method: 'GET' },
  { channel: CHANNELS.extensionsGetStatus, path: '/extensions/status', method: 'POST' },
  { channel: CHANNELS.extensionsDetect, path: '/extensions/detect', method: 'POST' },
  { channel: CHANNELS.extensionsInstall, path: '/extensions/install', method: 'POST' },
  { channel: CHANNELS.extensionsStart, path: '/extensions/start', method: 'POST' },
  { channel: CHANNELS.extensionsStop, path: '/extensions/stop', method: 'POST' },
  { channel: CHANNELS.extensionsRestart, path: '/extensions/restart', method: 'POST' },
  { channel: CHANNELS.extensionsTest, path: '/extensions/test', method: 'POST' },
  {
    channel: CHANNELS.extensionsRefreshModels,
    path: '/extensions/refresh-models',
    method: 'POST'
  },
  {
    channel: CHANNELS.extensionsVerifyStreaming,
    path: '/extensions/verify-streaming',
    method: 'POST'
  },
  { channel: CHANNELS.extensionsTailLogs, path: '/extensions/logs/tail', method: 'POST' },
  { channel: CHANNELS.extensionsClearLogs, path: '/extensions/logs/clear', method: 'POST' },
  { channel: CHANNELS.agentsList, path: '/agents', method: 'GET' },
  { channel: CHANNELS.agentsGet, path: '/agents/get', method: 'POST' },
  { channel: CHANNELS.agentsSave, path: '/agents/save', method: 'POST' },
  { channel: CHANNELS.agentsDuplicate, path: '/agents/duplicate', method: 'POST' },
  { channel: CHANNELS.agentsDelete, path: '/agents/delete', method: 'POST' },
  { channel: CHANNELS.skillsList, path: '/skills', method: 'GET' },
  { channel: CHANNELS.skillsGet, path: '/skills/get', method: 'POST' },
  { channel: CHANNELS.skillsSave, path: '/skills/save', method: 'POST' },
  { channel: CHANNELS.skillsDelete, path: '/skills/delete', method: 'POST' },
  { channel: CHANNELS.toolsList, path: '/tools', method: 'GET' },
  { channel: CHANNELS.toolsGet, path: '/tools/get', method: 'POST' },
  { channel: CHANNELS.toolsSave, path: '/tools/save', method: 'POST' },
  { channel: CHANNELS.toolsSetEnabled, path: '/tools/set-enabled', method: 'POST' },
  { channel: CHANNELS.toolsDelete, path: '/tools/delete', method: 'POST' },
  { channel: CHANNELS.mcpListServers, path: '/mcp/servers', method: 'GET' },
  { channel: CHANNELS.mcpSaveServer, path: '/mcp/servers/save', method: 'POST' },
  { channel: CHANNELS.mcpDeleteServer, path: '/mcp/servers/delete', method: 'POST' },
  { channel: CHANNELS.mcpTestServer, path: '/mcp/servers/test', method: 'POST' },
  { channel: CHANNELS.mcpRefreshTools, path: '/mcp/servers/refresh-tools', method: 'POST' },
  { channel: CHANNELS.projectsList, path: '/projects', method: 'GET' },
  { channel: CHANNELS.projectsCreate, path: '/projects/create', method: 'POST' },
  { channel: CHANNELS.projectsGet, path: '/projects/get', method: 'POST' },
  { channel: CHANNELS.projectsUpdate, path: '/projects/update', method: 'POST' },
  { channel: CHANNELS.projectsDelete, path: '/projects/delete', method: 'POST' },
  {
    channel: CHANNELS.projectsValidateLocalDirectory,
    path: '/projects/local-directory/validate',
    method: 'POST'
  },
  {
    channel: CHANNELS.projectsInitializeLocalDirectory,
    path: '/projects/local-directory/initialize',
    method: 'POST'
  },
  { channel: CHANNELS.projectsExportManifest, path: '/projects/export-manifest', method: 'POST' },
  { channel: CHANNELS.projectsListModules, path: '/projects/modules', method: 'POST' },
  { channel: CHANNELS.projectsGetModule, path: '/projects/modules/get', method: 'POST' },
  {
    channel: CHANNELS.projectsUpdateModuleConfig,
    path: '/projects/modules/update-config',
    method: 'POST'
  },
  {
    channel: CHANNELS.projectsListRequirementSources,
    path: '/projects/requirements/sources',
    method: 'POST'
  },
  {
    channel: CHANNELS.projectsPreviewRequirementSource,
    path: '/projects/requirements/preview-source',
    method: 'POST'
  },
  {
    channel: CHANNELS.projectsRunRequirementAnalysis,
    path: '/projects/requirements/run',
    method: 'POST'
  },
  { channel: CHANNELS.chatListSessions, path: '/chat/sessions', method: 'GET' },
  { channel: CHANNELS.chatCreateSession, path: '/chat/sessions/create', method: 'POST' },
  { channel: CHANNELS.chatUpdateSession, path: '/chat/sessions/update', method: 'POST' },
  { channel: CHANNELS.chatGetMessages, path: '/chat/messages', method: 'POST' },
  { channel: CHANNELS.chatSendMessage, path: '/chat/messages/send', method: 'POST' },
  { channel: CHANNELS.chatDeleteSession, path: '/chat/sessions/delete', method: 'POST' },
  { channel: CHANNELS.codingListEngines, path: '/coding/engines', method: 'GET' },
  { channel: CHANNELS.codingCreateSession, path: '/coding/sessions/create', method: 'POST' },
  { channel: CHANNELS.codingGetSession, path: '/coding/sessions/get', method: 'POST' },
  { channel: CHANNELS.codingListFiles, path: '/coding/files/list', method: 'POST' },
  { channel: CHANNELS.codingReadFile, path: '/coding/files/read', method: 'POST' },
  { channel: CHANNELS.codingFileStatus, path: '/coding/files/status', method: 'POST' }
]

export function registerRuntimeIpcHandlers(
  runtimePingProvider: RuntimePingProvider = createRuntimePingStubResponse,
  runtimeRequestProvider?: RuntimeRequestProvider,
  runtimeStreamProvider?: RuntimeStreamProvider,
  runtimeStreamCancelProvider?: (streamId: string) => void
): void {
  ipcMain.handle(CHANNELS.runtimePing, () => runtimePingProvider())

  ipcMain.handle(CHANNELS.systemOpenExternal, (_event, payload: unknown) => {
    const url = isRecord(payload) && typeof payload.url === 'string' ? payload.url : ''
    return openExternalHttpUrl(url)
  })

  ipcMain.handle(CHANNELS.projectsPickLocalDirectory, async () => {
    const result = await dialog.showOpenDialog({
      title: '选择项目本地目录',
      properties: ['openDirectory', 'createDirectory']
    })
    if (result.canceled) {
      return null
    }
    return result.filePaths[0] ?? null
  })

  ipcMain.handle(CHANNELS.projectsOpenLocalDirectory, async (_event, payload: unknown) => {
    if (!runtimeRequestProvider) {
      throw new Error('Main Runtime 尚未连接。')
    }
    const projectId =
      isRecord(payload) && typeof payload.projectId === 'string' ? payload.projectId : ''
    const check = await runtimeRequestProvider<ProjectDirectoryCheck>(
      '/projects/local-directory/validate',
      { method: 'POST', body: { projectId } }
    )
    if (!check.localRootPath || !check.exists) {
      return {
        ok: false,
        projectId,
        localRootPath: check.localRootPath,
        message: check.message,
        check
      } satisfies ProjectLocalDirectoryActionResult
    }
    const openError = await shell.openPath(check.localRootPath)
    return {
      ok: openError === '',
      projectId,
      localRootPath: check.localRootPath,
      message: openError || '已打开项目本地目录。',
      check
    } satisfies ProjectLocalDirectoryActionResult
  })

  ipcMain.handle(CHANNELS.projectsImportRequirementFiles, async (_event, payload: unknown) => {
    if (!runtimeRequestProvider) {
      throw new Error('Main Runtime 尚未连接。')
    }
    const projectId =
      isRecord(payload) && typeof payload.projectId === 'string' ? payload.projectId : ''
    const result = await dialog.showOpenDialog({
      title: '导入需求文件',
      properties: ['openFile', 'multiSelections'],
      filters: [{ name: '需求文件', extensions: ['docx', 'pdf'] }]
    })
    if (result.canceled || result.filePaths.length === 0) {
      return null
    }
    return runtimeRequestProvider<RequirementImportResult>('/projects/requirements/import-files', {
      method: 'POST',
      body: { projectId, filePaths: result.filePaths }
    })
  })

  ipcMain.handle(CHANNELS.projectsOpenRequirementOutputFile, async (_event, payload: unknown) => {
    const absolutePath =
      isRecord(payload) && typeof payload.absolutePath === 'string' ? payload.absolutePath : ''
    if (!absolutePath || !nodePath.isAbsolute(absolutePath)) {
      return { ok: false, path: absolutePath, message: '缺少有效的本地文件路径。' }
    }
    const allowedOutputFiles = new Set([
      'feature_list.xlsx',
      'requirements_spec.docx',
      'requirements_spec.pdf',
      'requirements_analysis.json'
    ])
    const fileName = nodePath.basename(absolutePath).toLowerCase()
    const normalizedPath = absolutePath.replaceAll('\\', '/').toLowerCase()
    if (!allowedOutputFiles.has(fileName) || !normalizedPath.includes('/artifacts/product/')) {
      return { ok: false, path: absolutePath, message: '仅允许打开需求分析生成的文档产物。' }
    }
    const openError = await shell.openPath(absolutePath)
    return {
      ok: openError === '',
      path: absolutePath,
      message: openError || '已打开文件。'
    }
  })

  for (const route of RUNTIME_ROUTES) {
    ipcMain.handle(route.channel, (_event, payload: unknown) => {
      if (!runtimeRequestProvider) {
        throw new Error('Main Runtime 尚未连接，无法读取工作台数据。')
      }

      if (route.method === 'GET') {
        return runtimeRequestProvider(route.path, { method: 'GET' })
      }

      return runtimeRequestProvider(route.path, { method: 'POST', body: payload ?? {} })
    })
  }

  ipcMain.handle(CHANNELS.chatStartStream, async (event, payload: unknown) => {
    if (!runtimeStreamProvider || !isRecord(payload) || typeof payload.streamId !== 'string') {
      throw new Error('Main Runtime streaming 尚不可用。')
    }
    const streamId = payload.streamId
    return runtimeStreamProvider(
      '/chat/messages/stream',
      { body: payload, streamId },
      (streamEvent) => {
        event.sender.send(CHANNELS.chatStreamEvent, streamEvent)
      }
    )
  })

  ipcMain.handle(CHANNELS.chatCancelStream, (_event, payload: unknown) => {
    if (!runtimeRequestProvider) {
      throw new Error('Main Runtime 尚未连接。')
    }
    if (isRecord(payload) && typeof payload.streamId === 'string') {
      runtimeStreamCancelProvider?.(payload.streamId)
    }
    return runtimeRequestProvider('/chat/messages/cancel', { method: 'POST', body: payload ?? {} })
  })

  ipcMain.handle(CHANNELS.codingStartTurn, async (event, payload: unknown) => {
    if (!runtimeStreamProvider || !isRecord(payload) || typeof payload.streamId !== 'string') {
      throw new Error('Main Runtime streaming 尚不可用。')
    }
    const streamId = payload.streamId
    return runtimeStreamProvider(
      '/coding/turns/stream',
      { body: payload, streamId },
      (streamEvent) => {
        if (streamEvent.type === 'failed') {
          event.sender.send(CHANNELS.codingStreamEvent, {
            type: 'error',
            streamId,
            sessionId: '',
            engineId: 'codex',
            providerId: '',
            model: '',
            trace_id: streamId,
            sequence: 0,
            timestamp: new Date().toISOString(),
            message: streamEvent.error?.message ?? 'coding stream failed',
            error: streamEvent.error
          })
          return
        }
        event.sender.send(CHANNELS.codingStreamEvent, streamEvent)
      }
    )
  })

  ipcMain.handle(CHANNELS.codingCancelTurn, (_event, payload: unknown) => {
    if (!runtimeRequestProvider) {
      throw new Error('Main Runtime 尚未连接。')
    }
    if (isRecord(payload) && typeof payload.streamId === 'string') {
      runtimeStreamCancelProvider?.(payload.streamId)
    }
    return runtimeRequestProvider('/coding/turns/cancel', { method: 'POST', body: payload ?? {} })
  })
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}
