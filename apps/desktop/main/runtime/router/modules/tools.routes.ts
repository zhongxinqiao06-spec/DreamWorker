import { asString } from '../../shared/util'
import type { RuntimeContext } from '../../bootstrap/runtime-context'
import { get, post, type RuntimeRoute } from '../route'

export function toolRoutes(context: RuntimeContext): RuntimeRoute[] {
  const { store } = context
  return [
    get('/tools', () => store.listTools()),
    post('/tools/get', (body) => store.getTool(asString(body.toolId))),
    post('/tools/save', (body) => store.saveTool(body)),
    post('/tools/set-enabled', (body) =>
      store.setToolEnabled(asString(body.toolId), body.enabled === true)
    ),
    post('/tools/delete', (body) => store.deleteTool(asString(body.toolId)))
  ]
}
