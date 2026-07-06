import { notFound } from '../../kernel/errors'
import { asRecord, asString, maskSecret, newTraceId, nowISO } from '../../shared/util'
import type { DeleteResult, JsonRecord } from '../../types'
import type { McpRepository } from '../../store/repositories/mcp-repository'

export class McpService {
  constructor(private readonly mcp: McpRepository) {}

  listMcpServers(): JsonRecord[] {
    return this.mcp.listServers().map((server) => this.safeMcpServer(server))
  }

  saveMcpServer(input: JsonRecord): JsonRecord {
    const serverId = asString(input.serverId) || this.mcp.nextId()
    const secrets = asRecord(input.secrets)
    const secretMap = Object.fromEntries(
      Object.entries(secrets).filter(
        (entry): entry is [string, string] => typeof entry[1] === 'string'
      )
    )
    const previous = this.mcp.getServer(serverId) ?? {}
    const now = nowISO()
    const server: JsonRecord = {
      ...previous,
      ...input,
      serverId,
      args: Array.isArray(input.args) ? input.args : [],
      envKeys: Object.keys(secretMap),
      hasSecrets: Object.keys(secretMap).length > 0,
      maskedSecrets: Object.entries(secretMap).map(([key, value]) => `${key}=${maskSecret(value)}`),
      createdAt: asString(previous.createdAt) || now,
      updatedAt: now
    }
    delete server.secrets
    this.mcp.saveSecrets(serverId, secretMap)
    this.mcp.saveServer(serverId, server)
    return this.safeMcpServer(server)
  }

  deleteMcpServer(serverId: string): DeleteResult {
    if (!this.mcp.getServer(serverId)) {
      throw notFound('MCP_SERVER_NOT_FOUND', 'MCP server not found', 'select another MCP server')
    }
    this.mcp.deleteServer(serverId)
    return { ok: true, deletedId: serverId }
  }

  testMcpServer(serverId: string): JsonRecord {
    if (!this.mcp.getServer(serverId)) {
      throw notFound('MCP_SERVER_NOT_FOUND', 'MCP server not found', 'select another MCP server')
    }
    return {
      ok: true,
      targetId: serverId,
      message: 'MCP config accepted by Main Runtime',
      latencyMs: 0,
      trace_id: newTraceId()
    }
  }

  refreshMcpTools(serverId: string): JsonRecord[] {
    if (!this.mcp.getServer(serverId)) {
      throw notFound('MCP_SERVER_NOT_FOUND', 'MCP server not found', 'select another MCP server')
    }
    return []
  }

  private safeMcpServer(server: JsonRecord): JsonRecord {
    const safe = { ...server }
    delete safe.secrets
    return safe
  }
}
