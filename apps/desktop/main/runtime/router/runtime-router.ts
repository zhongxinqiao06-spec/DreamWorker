import { badRequest } from '../kernel/errors'
import { asRecord } from '../shared/util'
import type { JsonRecord } from '../types'
import type { RuntimeRoute, RuntimeStreamEvent, RuntimeStreamRoute } from './route'

export type RuntimeRequestInit = {
  readonly method?: 'GET' | 'POST'
  readonly body?: unknown
}

export type RuntimeStreamInit = {
  readonly body: unknown
  readonly streamId: string
}

export class RuntimeRouter {
  constructor(
    private readonly routes: readonly RuntimeRoute[],
    private readonly streamRoutes: readonly RuntimeStreamRoute[]
  ) {}

  async request<T>(path: string, init: RuntimeRequestInit = {}): Promise<T> {
    const method = init.method ?? 'GET'
    const route = this.routes.find(
      (candidate) => candidate.path === path && candidate.method === method
    )
    if (!route) {
      throw badRequest(
        'ROUTE_NOT_FOUND',
        `Runtime route not found: ${method} ${path}`,
        '刷新应用后重试。'
      )
    }
    const body: JsonRecord = method === 'POST' ? asRecord(init.body ?? {}) : {}
    return (await route.handler(body)) as T
  }

  async stream(
    path: string,
    init: RuntimeStreamInit,
    onEvent: (event: RuntimeStreamEvent) => void
  ): Promise<{ readonly streamId: string }> {
    const route = this.streamRoutes.find((candidate) => candidate.path === path)
    if (!route) {
      throw badRequest(
        'STREAM_ROUTE_NOT_FOUND',
        `Runtime stream route not found: ${path}`,
        '刷新应用后重试。'
      )
    }
    await route.handler(asRecord(init.body), onEvent)
    return { streamId: init.streamId }
  }
}

export function flattenRoutes(groups: readonly RuntimeRoute[][]): RuntimeRoute[] {
  return groups.flat()
}

export function flattenStreamRoutes(groups: readonly RuntimeStreamRoute[][]): RuntimeStreamRoute[] {
  return groups.flat()
}
