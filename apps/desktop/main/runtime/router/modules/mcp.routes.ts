import { asString } from '../../shared/util'
import type { RuntimeContext } from '../../bootstrap/runtime-context'
import { get, post, type RuntimeRoute } from '../route'

export function mcpRoutes(context: RuntimeContext): RuntimeRoute[] {
  const { store } = context
  return [
    get('/mcp/servers', () => store.listMcpServers()),
    post('/mcp/servers/save', (body) => store.saveMcpServer(body)),
    post('/mcp/servers/delete', (body) => store.deleteMcpServer(asString(body.serverId))),
    post('/mcp/servers/test', (body) => store.testMcpServer(asString(body.serverId))),
    post('/mcp/servers/refresh-tools', (body) => store.refreshMcpTools(asString(body.serverId)))
  ]
}
