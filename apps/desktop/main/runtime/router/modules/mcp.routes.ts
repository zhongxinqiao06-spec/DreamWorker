import { asString } from '../../shared/util'
import type { RuntimeContext } from '../../bootstrap/runtime-context'
import { get, post, type RuntimeRoute } from '../route'

export function mcpRoutes(context: RuntimeContext): RuntimeRoute[] {
  return [
    get('/mcp/servers', () => context.mcp.listMcpServers()),
    post('/mcp/servers/save', (body) => context.mcp.saveMcpServer(body)),
    post('/mcp/servers/delete', (body) => context.mcp.deleteMcpServer(asString(body.serverId))),
    post('/mcp/servers/test', (body) => context.mcp.testMcpServer(asString(body.serverId))),
    post('/mcp/servers/refresh-tools', (body) =>
      context.mcp.refreshMcpTools(asString(body.serverId))
    )
  ]
}
