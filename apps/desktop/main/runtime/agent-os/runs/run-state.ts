export type AgentRunKind = 'chat' | 'requirement' | 'coding' | 'document' | 'workflow'

export type AgentRunStatus =
  'queued' | 'running' | 'waiting_approval' | 'completed' | 'failed' | 'cancelled'

export type AgentRun = {
  runId: string
  projectId: string
  moduleId?: string
  sessionId?: string
  kind: AgentRunKind
  status: AgentRunStatus
  traceId: string
  createdAt: string
  updatedAt: string
}
