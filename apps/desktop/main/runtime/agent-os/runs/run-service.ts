import { newTraceId, nowISO } from '../../shared/util'
import type { AgentRun, AgentRunKind, AgentRunStatus } from './run-state'

export class RunService {
  private readonly runs = new Map<string, AgentRun>()

  create(input: {
    readonly runId: string
    readonly projectId: string
    readonly kind: AgentRunKind
    readonly moduleId?: string
    readonly sessionId?: string
  }): AgentRun {
    const timestamp = nowISO()
    const run: AgentRun = {
      runId: input.runId,
      projectId: input.projectId,
      kind: input.kind,
      status: 'queued',
      traceId: newTraceId(),
      createdAt: timestamp,
      updatedAt: timestamp
    }
    if (input.moduleId) {
      run.moduleId = input.moduleId
    }
    if (input.sessionId) {
      run.sessionId = input.sessionId
    }
    this.runs.set(run.runId, run)
    return run
  }

  updateStatus(runId: string, status: AgentRunStatus): AgentRun | undefined {
    const run = this.runs.get(runId)
    if (!run) {
      return undefined
    }
    const nextRun = { ...run, status, updatedAt: nowISO() }
    this.runs.set(runId, nextRun)
    return nextRun
  }

  list(): AgentRun[] {
    return [...this.runs.values()]
  }
}
