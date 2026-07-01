import { app, BrowserWindow, shell } from 'electron'
import { electronApp, is, optimizer } from '@electron-toolkit/utils'
import { join } from 'node:path'
import { startEngineDaemon, type EngineDaemon } from './engine-daemon'
import { registerRuntimeIpcHandlers } from './runtime-ipc'
import { createRuntimePingStubResponse } from './runtime-ping'
import { createMainWindowOptions } from './window-options'

function createMainWindow(): void {
  const mainWindow = new BrowserWindow(
    createMainWindowOptions(join(__dirname, '../preload/index.cjs'))
  )

  mainWindow.once('ready-to-show', () => {
    mainWindow.show()
  })

  mainWindow.webContents.setWindowOpenHandler(({ url }) => {
    void shell.openExternal(url)
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

let engineDaemon: EngineDaemon | null = null

function startEngine(): void {
  try {
    engineDaemon = startEngineDaemon()
    engineDaemon.ready.catch(() => undefined)
  } catch {
    engineDaemon = null
  }
}

async function pingRuntime() {
  if (!engineDaemon) {
    return createRuntimePingStubResponse()
  }

  try {
    return await engineDaemon.ping()
  } catch {
    return createRuntimePingStubResponse()
  }
}

async function requestEngine<T>(
  path: string,
  init?: {
    readonly method?: 'GET' | 'POST'
    readonly body?: unknown
  }
): Promise<T> {
  if (!engineDaemon) {
    throw new Error('Go Engine is not connected.')
  }
  return engineDaemon.request<T>(path, init)
}

async function streamEngine(
  path: string,
  init: {
    readonly body: unknown
    readonly streamId: string
  },
  onEvent: Parameters<NonNullable<EngineDaemon['stream']>>[2]
) {
  if (!engineDaemon) {
    throw new Error('Go Engine streaming is not connected.')
  }
  return engineDaemon.stream(path, init, onEvent)
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
  startEngine()
  registerRuntimeIpcHandlers(pingRuntime, requestEngine, streamEngine, (streamId) =>
    engineDaemon?.cancelStream(streamId)
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
  engineDaemon?.stop()
  engineDaemon = null
})

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit()
  }
})
