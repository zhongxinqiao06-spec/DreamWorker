import type { DreamWorkerApi } from '../../../shared/dreamworker-api'

declare global {
  interface Window {
    readonly dreamworker: DreamWorkerApi
  }
}

export {}
