import { sortedValues } from '../../shared/util'
import type { JsonRecord } from '../../types'
import type { WorkspaceStore } from '../workspace-store'

export class McpRepository {
  constructor(private readonly store: WorkspaceStore) {}

  nextId(): string {
    return this.store.nextId('mcp')
  }

  listServers(): JsonRecord[] {
    return sortedValues(this.store.snapshot.mcpServers, 'serverId')
  }

  getServer(serverId: string): JsonRecord | undefined {
    return this.store.snapshot.mcpServers[serverId]
  }

  saveServer(serverId: string, server: JsonRecord): void {
    this.store.snapshot.mcpServers[serverId] = server
    this.store.save()
  }

  deleteServer(serverId: string): void {
    delete this.store.snapshot.mcpServerSecrets[serverId]
    delete this.store.snapshot.mcpServers[serverId]
    this.store.save()
  }

  saveSecrets(serverId: string, secrets: Record<string, string>): void {
    if (Object.keys(secrets).length > 0) {
      this.store.snapshot.mcpServerSecrets[serverId] = secrets
    }
  }
}
