import { contextBridge, ipcRenderer } from 'electron'
import { createDreamWorkerApi } from './bridge'

contextBridge.exposeInMainWorld(
  'dreamworker',
  createDreamWorkerApi(
    (channel, ...args) => ipcRenderer.invoke(channel, ...args),
    (channel, listener) => {
      const wrapped = (_event: Electron.IpcRendererEvent, payload: unknown): void => {
        listener(payload)
      }
      ipcRenderer.on(channel, wrapped)
      return () => {
        ipcRenderer.removeListener(channel, wrapped)
      }
    }
  )
)
