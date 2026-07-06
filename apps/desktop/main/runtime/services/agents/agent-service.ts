import { badRequest, notFound } from '../../kernel/errors'
import { asString, nowISO } from '../../shared/util'
import type { DeleteResult, JsonRecord } from '../../types'
import type { AgentRepository } from '../../store/repositories/agent-repository'

export class AgentService {
  constructor(private readonly agents: AgentRepository) {}

  listAgents(): JsonRecord[] {
    return this.agents.list()
  }

  getAgent(agentId: string): JsonRecord {
    const agent = this.agents.get(agentId)
    if (!agent) {
      throw notFound('AGENT_NOT_FOUND', 'agent not found', 'refresh list')
    }
    return agent
  }

  saveAgent(input: JsonRecord): JsonRecord {
    const agentId = asString(input.agentId) || this.agents.nextId()
    const previous = this.agents.get(agentId) ?? {}
    const now = nowISO()
    const agent = {
      ...previous,
      ...input,
      agentId,
      createdAt: asString(previous.createdAt) || now,
      updatedAt: now
    }
    this.agents.save(agentId, agent)
    return agent
  }

  duplicateAgent(agentId: string): JsonRecord {
    const agent = this.getAgent(agentId)
    const duplicated = {
      ...agent,
      agentId: this.agents.nextId(),
      displayName: `${asString(agent.displayName) || agentId} Copy`,
      builtIn: false,
      createdAt: nowISO(),
      updatedAt: nowISO()
    }
    this.agents.save(asString(duplicated.agentId), duplicated)
    return duplicated
  }

  deleteAgent(agentId: string): DeleteResult {
    if (!agentId) {
      throw badRequest('BAD_REQUEST', 'missing agentId', 'select an item')
    }
    if (!this.agents.get(agentId)) {
      throw notFound('RESOURCE_NOT_FOUND', 'resource not found', 'refresh list')
    }
    this.agents.delete(agentId)
    return { ok: true, deletedId: agentId }
  }
}
