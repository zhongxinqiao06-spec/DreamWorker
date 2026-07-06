import type { JsonRecord } from '../../types'

export type ToolRouteRequest = {
  toolName: string
  input: JsonRecord
}

export class ToolRouter {
  route(request: ToolRouteRequest): JsonRecord {
    return {
      routed: false,
      toolName: request.toolName,
      input: request.input,
      message: 'tool routing is reserved for the Agent OS executor'
    }
  }
}
