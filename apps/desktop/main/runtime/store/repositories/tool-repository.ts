import { sortedValues } from '../../shared/util'
import type { JsonRecord } from '../../types'
import type { WorkspaceStore } from '../workspace-store'

export class ToolRepository {
  constructor(private readonly store: WorkspaceStore) {}

  nextId(): string {
    return this.store.nextId('tool')
  }

  list(): JsonRecord[] {
    return sortedValues(this.store.snapshot.tools, 'toolId')
  }

  get(toolId: string): JsonRecord | undefined {
    return this.store.snapshot.tools[toolId]
  }

  save(toolId: string, tool: JsonRecord): void {
    this.store.snapshot.tools[toolId] = tool
    this.store.save()
  }

  delete(toolId: string): void {
    delete this.store.snapshot.tools[toolId]
    this.store.save()
  }
}
