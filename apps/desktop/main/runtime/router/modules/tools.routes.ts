import { asString } from '../../shared/util'
import type { RuntimeContext } from '../../bootstrap/runtime-context'
import { get, post, type RuntimeRoute } from '../route'

export function toolRoutes(context: RuntimeContext): RuntimeRoute[] {
  return [
    get('/tools', () => context.tools.listTools()),
    post('/tools/get', (body) => context.tools.getTool(asString(body.toolId))),
    post('/tools/save', (body) => context.tools.saveTool(body)),
    post('/tools/set-enabled', (body) =>
      context.tools.setToolEnabled(asString(body.toolId), body.enabled === true)
    ),
    post('/tools/delete', (body) => context.tools.deleteTool(asString(body.toolId)))
  ]
}
