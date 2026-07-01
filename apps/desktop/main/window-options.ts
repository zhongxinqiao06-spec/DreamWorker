import type { BrowserWindowConstructorOptions } from 'electron'

export function createMainWindowOptions(preloadPath: string): BrowserWindowConstructorOptions {
  return {
    width: 1240,
    height: 820,
    minWidth: 960,
    minHeight: 640,
    show: false,
    title: 'DreamWorker 项目孵化器',
    autoHideMenuBar: true,
    backgroundColor: '#111318',
    webPreferences: {
      preload: preloadPath,
      contextIsolation: true,
      nodeIntegration: false,
      sandbox: true
    }
  }
}
