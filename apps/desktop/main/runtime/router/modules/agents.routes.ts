import { asString } from '../../shared/util'
import type { RuntimeContext } from '../../bootstrap/runtime-context'
import { get, post, type RuntimeRoute } from '../route'

export function agentRoutes(context: RuntimeContext): RuntimeRoute[] {
  const { store } = context
  return [
    get('/agents', () => store.listAgents()),
    post('/agents/get', (body) => store.getAgent(asString(body.agentId))),
    post('/agents/save', (body) => store.saveAgent(body)),
    post('/agents/duplicate', (body) => store.duplicateAgent(asString(body.agentId))),
    post('/agents/delete', (body) => store.deleteAgent(asString(body.agentId)))
  ]
}
