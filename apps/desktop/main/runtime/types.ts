export type JsonRecord = Record<string, unknown>

export type DeleteResult = {
  readonly ok: boolean
  readonly deletedId: string
}

export type WorkspaceSnapshot = {
  schemaVersion: string
  sequence: number
  providers: Record<string, JsonRecord>
  providerSecrets: Record<string, string>
  profiles: Record<string, JsonRecord>
  agents: Record<string, JsonRecord>
  skills: Record<string, JsonRecord>
  tools: Record<string, JsonRecord>
  mcpServers: Record<string, JsonRecord>
  mcpServerSecrets: Record<string, Record<string, string>>
  mcpTools: Record<string, JsonRecord>
  projects: Record<string, JsonRecord>
  modules: Record<string, Record<string, JsonRecord>>
  sessions: Record<string, JsonRecord>
  messages: Record<string, JsonRecord[]>
  contextSummaries: Record<string, JsonRecord[]>
  settings?: JsonRecord
}

export class RuntimeAppError extends Error {
  readonly status: number
  readonly code: string
  readonly userAction: string

  constructor(status: number, code: string, message: string, userAction: string) {
    super(message)
    this.status = status
    this.code = code
    this.userAction = userAction
  }
}

export function badRequest(code: string, message: string, userAction: string): RuntimeAppError {
  return new RuntimeAppError(400, code, message, userAction)
}

export function notFound(code: string, message: string, userAction: string): RuntimeAppError {
  return new RuntimeAppError(404, code, message, userAction)
}

export function internalError(code: string, message: string, userAction: string): RuntimeAppError {
  return new RuntimeAppError(500, code, message, userAction)
}

export type CodingEngineId = 'claude_agent' | 'codex' | 'opencode'

export type CodingSession = {
  sessionId: string
  projectId: string
  engineId: CodingEngineId
  providerId: string
  model: string
  title: string
  localRootPath: string
  engineThreadId: string
  status: string
  createdAt: string
  updatedAt: string
}

export type CodingStreamEvent = {
  type:
    | 'started'
    | 'delta'
    | 'tool_call'
    | 'shell_output'
    | 'file_changed'
    | 'completed'
    | 'cancelled'
    | 'error'
  streamId: string
  sessionId: string
  engineId: CodingEngineId
  providerId: string
  model: string
  trace_id: string
  sequence: number
  timestamp: string
  delta?: string
  message?: string
  command?: string
  output?: string
  path?: string
  status?: string
  engineThreadId?: string
  toolCall?: {
    callId: string
    toolName: string
    arguments?: unknown
  }
  file?: {
    path: string
    status: string
  }
  error?: {
    code: string
    message: string
    recoverable: boolean
  }
  runtimeAvailable?: boolean
}
