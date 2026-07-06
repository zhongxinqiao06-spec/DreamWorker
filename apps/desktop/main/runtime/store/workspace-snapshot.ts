import { internalError, type WorkspaceSnapshot } from '../types'
import { nowISO } from '../shared/util'
import { createDefaultSnapshot, ensureSnapshotShape } from './defaults'
import type { WorkspaceDb } from './workspace-db'

const workspaceSnapshotKey = 'workspace'

export class WorkspaceSnapshotStore {
  constructor(private readonly db: WorkspaceDb) {}

  load(): WorkspaceSnapshot {
    const row = this.db.connection
      .prepare('SELECT payload FROM workspace_state WHERE key = ?')
      .get(workspaceSnapshotKey) as { payload?: string } | undefined
    if (!row?.payload) {
      const seeded = createDefaultSnapshot()
      this.save(seeded)
      return seeded
    }
    try {
      return ensureSnapshotShape(JSON.parse(row.payload) as WorkspaceSnapshot)
    } catch {
      throw internalError(
        'WORKSPACE_PERSIST_DECODE_FAILED',
        'failed to decode workspace state',
        'check workspace database'
      )
    }
  }

  save(snapshot: WorkspaceSnapshot): void {
    const payload = JSON.stringify(snapshot)
    this.db.connection
      .prepare(
        `INSERT INTO workspace_state (key, payload, updated_at)
VALUES (?, ?, ?)
ON CONFLICT(key) DO UPDATE SET payload = excluded.payload, updated_at = excluded.updated_at`
      )
      .run(workspaceSnapshotKey, payload, nowISO())
  }
}
