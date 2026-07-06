import { createRuntimeContext } from './bootstrap/create-runtime'
import type { RuntimeContext } from './bootstrap/runtime-context'
import { createRuntimeRouter } from './router/create-runtime-router'
import type { RuntimeStreamEvent } from './router/route'
import {
  RuntimeRouter,
  type RuntimeRequestInit,
  type RuntimeStreamInit
} from './router/runtime-router'
import { engineVersion, runtimePing } from './runtime-info'
import { newTraceId } from './shared/util'
import { RuntimeAppError, type JsonRecord } from './types'
import type { RuntimePingResponse } from '../../shared/dreamworker-api'

export type { RuntimeRequestInit, RuntimeStreamInit } from './router/runtime-router'
export type { RuntimeStreamEvent } from './router/route'

export class DreamWorkerRuntime {
  private readonly ctx: RuntimeContext
  private readonly router: RuntimeRouter

  constructor(configDir?: string) {
    this.ctx = createRuntimeContext(configDir)
    this.router = createRuntimeRouter(this.ctx)
  }

  ping(): RuntimePingResponse {
    return runtimePing() as unknown as RuntimePingResponse
  }

  request<T>(path: string, init: RuntimeRequestInit = {}): Promise<T> {
    return this.router.request<T>(path, init)
  }

  stream(
    path: string,
    init: RuntimeStreamInit,
    onEvent: (event: RuntimeStreamEvent) => void
  ): Promise<{ readonly streamId: string }> {
    return this.router.stream(path, init, onEvent)
  }

  cancelStream(streamId: string): void {
    try {
      this.ctx.coding.cancelTurn({ streamId })
    } catch {
      // The IPC cancel route also performs a typed cancel. This hook is best-effort cleanup.
    }
  }

  stop(): void {
    this.ctx.lifecycle.stop()
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
