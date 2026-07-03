import type { BrowserWindowConstructorOptions } from 'electron'

export function createMainWindowOptions(preloadPath: string): BrowserWindowConstructorOptions {
  return {
    width: 1600,
    height: 960,
    minWidth: 1280,
    minHeight: 720,
    show: false,
    title: 'DreamWorker AI 工作台',
    titleBarStyle: 'hidden',
    titleBarOverlay: {
      color: '#f8fafc',
      symbolColor: '#0f172a',
      height: 40
    },
    autoHideMenuBar: true,
    backgroundColor: '#f8fafc',
    webPreferences: {
      preload: preloadPath,
      contextIsolation: true,
      nodeIntegration: false,
      sandbox: true
    }
  }
}
