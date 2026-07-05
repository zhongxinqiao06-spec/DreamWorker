import { beforeEach, describe, expect, it, vi } from 'vitest'
import { CHANNELS, type ProjectDirectoryCheck } from '../shared/dreamworker-api'
import { registerRuntimeIpcHandlers } from './runtime-ipc'

const electronMock = vi.hoisted(() => {
  const handlers = new Map<string, (...args: unknown[]) => unknown>()
  return {
    handlers,
    handle: vi.fn((channel: string, handler: (...args: unknown[]) => unknown) => {
      handlers.set(channel, handler)
    }),
    showOpenDialog: vi.fn(),
    openPath: vi.fn(),
    openExternal: vi.fn()
  }
})

vi.mock('electron', () => ({
  ipcMain: {
    handle: electronMock.handle
  },
  dialog: {
    showOpenDialog: electronMock.showOpenDialog
  },
  shell: {
    openPath: electronMock.openPath,
    openExternal: electronMock.openExternal
  }
}))

describe('runtime ipc project directory handlers', () => {
  beforeEach(() => {
    electronMock.handlers.clear()
    electronMock.handle.mockClear()
    electronMock.showOpenDialog.mockReset()
    electronMock.openPath.mockReset()
    electronMock.openExternal.mockReset()
  })

  it('opens a creatable directory picker for project local roots', async () => {
    registerRuntimeIpcHandlers()
    electronMock.showOpenDialog.mockResolvedValue({
      canceled: false,
      filePaths: ['C:\\DreamWorkerProjects\\picked']
    })

    const handler = electronMock.handlers.get(CHANNELS.projectsPickLocalDirectory)
    await expect(handler?.()).resolves.toBe('C:\\DreamWorkerProjects\\picked')

    expect(electronMock.showOpenDialog).toHaveBeenCalledWith(
      expect.objectContaining({
        properties: ['openDirectory', 'createDirectory']
      })
    )
  })

  it('returns null when project directory picking is cancelled', async () => {
    registerRuntimeIpcHandlers()
    electronMock.showOpenDialog.mockResolvedValue({ canceled: true, filePaths: [] })

    const handler = electronMock.handlers.get(CHANNELS.projectsPickLocalDirectory)
    await expect(handler?.()).resolves.toBeNull()
  })

  it('imports requirement documents through the project requirement picker', async () => {
    const importResult = {
      projectId: 'project_001',
      runId: 'req_001',
      sources: [],
      message: '已导入 2 个需求文件。'
    }
    const requestEngine = vi.fn().mockResolvedValue(importResult)
    registerRuntimeIpcHandlers(undefined, requestEngine)
    electronMock.showOpenDialog.mockResolvedValue({
      canceled: false,
      filePaths: ['C:\\input\\requirements.docx', 'C:\\input\\scope.pdf']
    })

    const handler = electronMock.handlers.get(CHANNELS.projectsImportRequirementFiles)
    await expect(handler?.(null, { projectId: 'project_001' })).resolves.toEqual(importResult)

    expect(electronMock.showOpenDialog).toHaveBeenCalledWith(
      expect.objectContaining({
        properties: ['openFile', 'multiSelections'],
        filters: [{ name: '需求文件', extensions: ['docx', 'pdf'] }]
      })
    )
    expect(requestEngine).toHaveBeenCalledWith('/projects/requirements/import-files', {
      method: 'POST',
      body: {
        projectId: 'project_001',
        filePaths: ['C:\\input\\requirements.docx', 'C:\\input\\scope.pdf']
      }
    })
  })

  it('routes requirement source preview to the engine', async () => {
    const previewResult = {
      projectId: 'project_001',
      source: { sourceId: 'src_imported', fileName: 'requirements.docx' },
      parser: 'mineru',
      content: '解析文本',
      charCount: 4,
      truncated: false,
      traceId: 'tr_preview',
      createdAt: '2026-07-01T00:00:00Z'
    }
    const requestEngine = vi.fn().mockResolvedValue(previewResult)
    registerRuntimeIpcHandlers(undefined, requestEngine)

    const handler = electronMock.handlers.get(CHANNELS.projectsPreviewRequirementSource)
    await expect(
      handler?.(null, { projectId: 'project_001', sourceId: 'src_imported' })
    ).resolves.toEqual(previewResult)

    expect(requestEngine).toHaveBeenCalledWith('/projects/requirements/preview-source', {
      method: 'POST',
      body: { projectId: 'project_001', sourceId: 'src_imported' }
    })
  })

  it('opens only http urls through the system external channel', async () => {
    registerRuntimeIpcHandlers()
    electronMock.openExternal.mockResolvedValue(undefined)

    const handler = electronMock.handlers.get(CHANNELS.systemOpenExternal)
    await expect(handler?.(null, { url: 'http://localhost:20128/dashboard' })).resolves.toEqual({
      ok: true,
      url: 'http://localhost:20128/dashboard',
      message: null
    })
    await expect(handler?.(null, { url: 'file:///C:/secret.txt' })).resolves.toEqual({
      ok: false,
      url: 'file:///C:/secret.txt',
      message: 'Only http and https URLs can be opened externally.'
    })

    expect(electronMock.openExternal).toHaveBeenCalledTimes(1)
    expect(electronMock.openExternal).toHaveBeenCalledWith('http://localhost:20128/dashboard')
  })

  it('validates the project directory before opening it', async () => {
    const check: ProjectDirectoryCheck = {
      projectId: 'project_001',
      localRootPath: 'C:\\DreamWorkerProjects\\project_001',
      status: 'valid',
      lastCheckedAt: '2026-07-01T00:00:00Z',
      exists: true,
      readable: true,
      writable: true,
      dreamworkerInitialized: true,
      requiredDirectories: [],
      message: '本地目录可用，项目结构完整。'
    }
    const requestEngine = vi.fn().mockResolvedValue(check)
    registerRuntimeIpcHandlers(undefined, requestEngine)
    electronMock.openPath.mockResolvedValue('')

    const handler = electronMock.handlers.get(CHANNELS.projectsOpenLocalDirectory)
    await expect(handler?.(null, { projectId: 'project_001' })).resolves.toEqual({
      ok: true,
      projectId: 'project_001',
      localRootPath: 'C:\\DreamWorkerProjects\\project_001',
      message: expect.any(String),
      check
    })

    expect(requestEngine).toHaveBeenCalledWith('/projects/local-directory/validate', {
      method: 'POST',
      body: { projectId: 'project_001' }
    })
    expect(electronMock.openPath).toHaveBeenCalledWith('C:\\DreamWorkerProjects\\project_001')
  })

  it('does not open a missing project directory', async () => {
    const check: ProjectDirectoryCheck = {
      projectId: 'project_001',
      localRootPath: 'C:\\DreamWorkerProjects\\missing',
      status: 'missing',
      lastCheckedAt: '2026-07-01T00:00:00Z',
      exists: false,
      readable: false,
      writable: false,
      dreamworkerInitialized: false,
      requiredDirectories: [],
      message: '本地目录不存在。'
    }
    const requestEngine = vi.fn().mockResolvedValue(check)
    registerRuntimeIpcHandlers(undefined, requestEngine)

    const handler = electronMock.handlers.get(CHANNELS.projectsOpenLocalDirectory)
    await expect(handler?.(null, { projectId: 'project_001' })).resolves.toEqual({
      ok: false,
      projectId: 'project_001',
      localRootPath: 'C:\\DreamWorkerProjects\\missing',
      message: '本地目录不存在。',
      check
    })

    expect(electronMock.openPath).not.toHaveBeenCalled()
  })
})
