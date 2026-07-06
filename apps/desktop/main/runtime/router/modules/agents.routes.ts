import { asString } from '../../shared/util'
import type { RuntimeContext } from '../../bootstrap/runtime-context'
import { get, post, type RuntimeRoute } from '../route'

export function agentRoutes(context: RuntimeContext): RuntimeRoute[] {
  return [
    get('/agents', () => context.agents.listAgents()),
    post('/agents/get', (body) => context.agents.getAgent(asString(body.agentId))),
    post('/agents/save', (body) => context.agents.saveAgent(body)),
    post('/agents/duplicate', (body) => context.agents.duplicateAgent(asString(body.agentId))),
    post('/agents/delete', (body) => context.agents.deleteAgent(asString(body.agentId)))
  ]
}
