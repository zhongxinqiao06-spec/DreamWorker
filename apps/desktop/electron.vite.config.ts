import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { defineConfig, externalizeDepsPlugin } from 'electron-vite'
import vue from '@vitejs/plugin-vue'

const appRoot = dirname(fileURLToPath(import.meta.url))

export default defineConfig({
  main: {
    plugins: [externalizeDepsPlugin()],
    build: {
      rollupOptions: {
        input: {
          index: resolve(appRoot, 'main/index.ts')
        }
      }
    }
  },
  preload: {
    plugins: [externalizeDepsPlugin()],
    build: {
      rollupOptions: {
        input: {
          index: resolve(appRoot, 'preload/index.ts')
        }
      }
    }
  },
  renderer: {
    root: resolve(appRoot, 'renderer'),
    plugins: [vue()],
    build: {
      rollupOptions: {
        input: resolve(appRoot, 'renderer/index.html')
      }
    }
  }
})
