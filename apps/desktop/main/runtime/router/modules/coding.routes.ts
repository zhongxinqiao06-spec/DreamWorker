import { asRecord } from '../../shared/util'
import type { RuntimeContext } from '../../bootstrap/runtime-context'
import { get, post, stream, type RuntimeRoute, type RuntimeStreamRoute } from '../route'

export function codingRoutes(context: RuntimeContext): RuntimeRoute[] {
  return [
    get('/coding/engines', () => context.coding.listEngines()),
    post('/coding/sessions/create', (body) => context.coding.createSession(body)),
    post('/coding/sessions/get', (body) => context.coding.getSession(body)),
    post('/coding/files/list', (body) => context.coding.listFiles(body)),
    post('/coding/files/read', (body) => context.coding.readFile(body)),
    post('/coding/files/status', (body) => context.coding.fileStatus(body)),
    post('/coding/turns/cancel', (body) => context.coding.cancelTurn(body))
  ]
}

export function codingStreamRoutes(context: RuntimeContext): RuntimeStreamRoute[] {
  return [
    stream('/coding/turns/stream', async (body, onEvent) => {
      for await (const event of context.coding.streamTurn(asRecord(body))) {
        onEvent(event)
      }
    })
  ]
}
