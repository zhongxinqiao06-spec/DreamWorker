import { badRequest, notFound } from '../../kernel/errors'
import { asString, nowISO } from '../../shared/util'
import type { DeleteResult, JsonRecord } from '../../types'
import type { ToolRepository } from '../../store/repositories/tool-repository'

export class ToolConfigService {
  constructor(private readonly tools: ToolRepository) {}

  listTools(): JsonRecord[] {
    return this.tools.list()
  }

  getTool(toolId: string): JsonRecord {
    const tool = this.tools.get(toolId)
    if (!tool) {
      throw notFound('TOOL_NOT_FOUND', 'tool not found', 'refresh list')
    }
    return tool
  }

  saveTool(input: JsonRecord): JsonRecord {
    const toolId = asString(input.toolId) || this.tools.nextId()
    const previous = this.tools.get(toolId) ?? {}
    const now = nowISO()
    const tool = {
      ...previous,
      ...input,
      toolId,
      createdAt: asString(previous.createdAt) || now,
      updatedAt: now
    }
    this.tools.save(toolId, tool)
    return tool
  }

  setToolEnabled(toolId: string, enabled: boolean): JsonRecord {
    const tool = this.getTool(toolId)
    const updated = { ...tool, enabled, updatedAt: nowISO() }
    this.tools.save(toolId, updated)
    return updated
  }

  deleteTool(toolId: string): DeleteResult {
    if (!toolId) {
      throw badRequest('BAD_REQUEST', 'missing toolId', 'select an item')
    }
    if (!this.tools.get(toolId)) {
      throw notFound('RESOURCE_NOT_FOUND', 'resource not found', 'refresh list')
    }
    this.tools.delete(toolId)
    return { ok: true, deletedId: toolId }
  }
}
