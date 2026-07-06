import { app, BrowserWindow, session } from 'electron'
import { electronApp, is, optimizer } from '@electron-toolkit/utils'
import { join } from 'node:path'
import { openExternalHttpUrl } from './external-url'
import { registerRuntimeIpcHandlers } from './runtime-ipc'
import { createDreamWorkerRuntime, type DreamWorkerRuntime } from './runtime/app'
import { createRuntimePingStubResponse } from './runtime-ping'
import { createMainWindowOptions } from './window-options'

const LOCAL_RENDERER_PROTOCOLS = new Set([
  'file:',
  'data:',
  'blob:',
  'devtools:',
  'chrome:',
  'about:'
])
let rendererNetworkGuardInstalled = false

function isLoopbackHost(hostname: string): boolean {
  const host = hostname.replace(/^\[|\]$/g, '').toLowerCase()
  return (
    host === 'localhost' ||
    host === '0.0.0.0' ||
    host === '::1' ||
    host === '::' ||
    host.startsWith('127.')
  )
}

function shouldBlockRendererRequest(rawUrl: string): boolean {
  try {
    const url = new URL(rawUrl)
    if (LOCAL_RENDERER_PROTOCOLS.has(url.protocol)) {
      return false
    }
    if (
      (url.protocol === 'http:' || url.protocol === 'https:' || url.protocol === 'ws:') &&
      isLoopbackHost(url.hostname)
    ) {
      return false
    }
    return true
  } catch {
    return true
  }
}

function installRendererNetworkGuard(): void {
  if (rendererNetworkGuardInstalled) {
    return
  }
  rendererNetworkGuardInstalled = true
  session.defaultSession.webRequest.onBeforeRequest((details, callback) => {
    callback({ cancel: shouldBlockRendererRequest(details.url) })
  })
}

function createMainWindow(): void {
  installRendererNetworkGuard()
  const mainWindow = new BrowserWindow(
    createMainWindowOptions(
      join(__dirname, '../preload/index.cjs'),
      join(__dirname, '../../build/icon.ico')
    )
  )

  mainWindow.once('ready-to-show', () => {
    mainWindow.maximize()
    mainWindow.show()
  })

  mainWindow.webContents.setWindowOpenHandler(({ url }) => {
    void openExternalHttpUrl(url)
    return { action: 'deny' }
  })

  if (is.dev && process.env.ELECTRON_RENDERER_URL) {
    void mainWindow.loadURL(process.env.ELECTRON_RENDERER_URL)
    registerE2ESmoke(mainWindow)
    return
  }

  void mainWindow.loadFile(join(__dirname, '../renderer/index.html'))
  registerE2ESmoke(mainWindow)
}

let runtime: DreamWorkerRuntime | null = null

function startRuntime(): void {
  try {
    runtime = createDreamWorkerRuntime()
  } catch {
    runtime = null
  }
}

async function pingRuntime() {
  if (!runtime) {
    return createRuntimePingStubResponse()
  }

  try {
    return runtime.ping()
  } catch {
    return createRuntimePingStubResponse()
  }
}

async function requestRuntime<T>(
  path: string,
  init?: {
    readonly method?: 'GET' | 'POST'
    readonly body?: unknown
  }
): Promise<T> {
  if (!runtime) {
    throw new Error('Main 内嵌 Runtime 尚未连接。')
  }
  return runtime.request<T>(path, init)
}

async function streamRuntime(
  path: string,
  init: {
    readonly body: unknown
    readonly streamId: string
  },
  onEvent: Parameters<NonNullable<DreamWorkerRuntime['stream']>>[2]
) {
  if (!runtime) {
    throw new Error('Main 内嵌 Runtime streaming 尚未连接。')
  }
  return runtime.stream(path, init, onEvent)
}

function registerE2ESmoke(mainWindow: BrowserWindow): void {
  if (process.env.DREAMWORKER_E2E_PING !== '1' && process.env.DREAMWORKER_E2E_WORKSPACE !== '1') {
    return
  }

  mainWindow.webContents.once('did-finish-load', () => {
    const script =
      process.env.DREAMWORKER_E2E_WORKSPACE === '1'
        ? `Promise.all([
            window.dreamworker.runtime.ping(),
            window.dreamworker.models.listProviders(),
            window.dreamworker.projects.listProjects(),
            window.dreamworker.chat.listSessions(),
            window.dreamworker.skills.listSkills()
          ]).then(([ping, providers, projects, sessions, skills]) => ({
            ping,
            providerCount: providers.length,
            firstProvider: providers[0],
            projectCount: projects.length,
            sessionCount: sessions.length,
            skillCount: skills.length,
            hasSkillCreator: skills.some((skill) => skill.skillId === 'skill_skillcreator')
          }))`
        : 'window.dreamworker.runtime.ping()'

    void mainWindow.webContents
      .executeJavaScript(script, true)
      .then((result) => {
        const label =
          process.env.DREAMWORKER_E2E_WORKSPACE === '1'
            ? 'DREAMWORKER_E2E_WORKSPACE_RESULT'
            : 'DREAMWORKER_E2E_PING_RESULT'
        console.log(`${label} ${JSON.stringify(result)}`)
        setTimeout(() => {
          app.quit()
        }, 1000)
      })
      .catch((error: unknown) => {
        const message = error instanceof Error ? error.message : String(error)
        console.error(`DREAMWORKER_E2E_PING_ERROR ${message}`)
        app.exit(1)
      })
  })
}

void app.whenReady().then(() => {
  electronApp.setAppUserModelId('dev.dreamworker.desktop')
  startRuntime()
  registerRuntimeIpcHandlers(pingRuntime, requestRuntime, streamRuntime, (streamId) =>
    runtime?.cancelStream(streamId)
  )

  app.on('browser-window-created', (_, window) => {
    optimizer.watchWindowShortcuts(window)
  })

  createMainWindow()

  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createMainWindow()
    }
  })
})

app.on('before-quit', () => {
  runtime?.stop()
  runtime = null
})

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit()
  }
})
