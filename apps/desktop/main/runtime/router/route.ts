import type { ChatStreamEvent } from '../../../shared/dreamworker-api'
import type { CodingStreamEvent, JsonRecord } from '../types'

export type RuntimeHttpMethod = 'GET' | 'POST'

export type RuntimeRoute = {
  readonly method: RuntimeHttpMethod
  readonly path: string
  readonly handler: (body: JsonRecord) => unknown | Promise<unknown>
}

export type RuntimeStreamEvent = ChatStreamEvent | CodingStreamEvent

export type RuntimeStreamRoute = {
  readonly path: string
  readonly handler: (
    body: JsonRecord,
    onEvent: (event: RuntimeStreamEvent) => void
  ) => Promise<void>
}

export function get(path: string, handler: RuntimeRoute['handler']): RuntimeRoute {
  return { method: 'GET', path, handler }
}

export function post(path: string, handler: RuntimeRoute['handler']): RuntimeRoute {
  return { method: 'POST', path, handler }
}

export function stream(path: string, handler: RuntimeStreamRoute['handler']): RuntimeStreamRoute {
  return { path, handler }
}
