import type { WorkspaceSnapshot } from '../types'
import { defaultConfigDir, WorkspaceDb } from './workspace-db'
import { WorkspaceSnapshotStore } from './workspace-snapshot'

export class WorkspaceStore {
  readonly configDir: string
  readonly db: WorkspaceDb
  private readonly snapshots: WorkspaceSnapshotStore
  snapshot: WorkspaceSnapshot

  constructor(configDir = defaultConfigDir()) {
    this.db = new WorkspaceDb(configDir)
    this.configDir = this.db.configDir
    this.snapshots = new WorkspaceSnapshotStore(this.db)
    this.snapshot = this.snapshots.load()
  }

  close(): void {
    this.db.close()
  }

  nextId(prefix: string): string {
    this.snapshot.sequence += 1
    return `${prefix}_${String(this.snapshot.sequence).padStart(3, '0')}`
  }

  save(): void {
    this.snapshots.save(this.snapshot)
  }
}
