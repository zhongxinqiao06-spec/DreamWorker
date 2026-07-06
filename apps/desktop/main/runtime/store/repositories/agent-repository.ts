import { sortedValues } from '../../shared/util'
import type { JsonRecord } from '../../types'
import type { WorkspaceStore } from '../workspace-store'

export class AgentRepository {
  constructor(private readonly store: WorkspaceStore) {}

  nextId(): string {
    return this.store.nextId('agent')
  }

  list(): JsonRecord[] {
    return sortedValues(this.store.snapshot.agents, 'agentId')
  }

  get(agentId: string): JsonRecord | undefined {
    return this.store.snapshot.agents[agentId]
  }

  save(agentId: string, agent: JsonRecord): void {
    this.store.snapshot.agents[agentId] = agent
    this.store.save()
  }

  delete(agentId: string): void {
    delete this.store.snapshot.agents[agentId]
    this.store.save()
  }
}
