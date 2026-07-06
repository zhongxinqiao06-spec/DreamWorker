import { mkdirSync } from 'node:fs'
import { homedir } from 'node:os'
import { join, resolve } from 'node:path'
import { DatabaseSync } from 'node:sqlite'

export class WorkspaceDb {
  readonly configDir: string
  readonly connection: DatabaseSync

  constructor(configDir = defaultConfigDir()) {
    this.configDir = configDir
    mkdirSync(this.configDir, { recursive: true })
    this.connection = new DatabaseSync(join(this.configDir, 'workspace.db'))
    bootstrapDatabase(this.connection)
  }

  close(): void {
    this.connection.close()
  }
}

export function defaultConfigDir(): string {
  const configured = process.env.DREAMWORKER_CONFIG_DIR?.trim()
  if (configured) {
    return resolve(configured)
  }
  if (process.platform === 'win32') {
    return join(process.env.APPDATA || join(homedir(), 'AppData', 'Roaming'), 'DreamWorker')
  }
  if (process.platform === 'darwin') {
    return join(homedir(), 'Library', 'Application Support', 'DreamWorker')
  }
  return join(process.env.XDG_CONFIG_HOME || join(homedir(), '.config'), 'DreamWorker')
}

function bootstrapDatabase(db: DatabaseSync): void {
  db.exec(`
CREATE TABLE IF NOT EXISTS schema_migrations (
  version TEXT PRIMARY KEY,
  checksum TEXT NOT NULL,
  applied_at TEXT NOT NULL,
  non_destructive INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS events (
  sequence INTEGER PRIMARY KEY AUTOINCREMENT,
  event_id TEXT NOT NULL UNIQUE,
  schema_version TEXT NOT NULL,
  trace_id TEXT NOT NULL,
  mission_id TEXT NOT NULL,
  run_id TEXT NOT NULL,
  actor TEXT NOT NULL,
  timestamp TEXT NOT NULL,
  type TEXT NOT NULL,
  payload TEXT NOT NULL,
  inserted_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE INDEX IF NOT EXISTS idx_events_mission_id_sequence ON events (mission_id, sequence);
CREATE INDEX IF NOT EXISTS idx_events_run_id_sequence ON events (run_id, sequence);
CREATE INDEX IF NOT EXISTS idx_events_trace_id ON events (trace_id);
CREATE INDEX IF NOT EXISTS idx_events_type ON events (type);
CREATE TABLE IF NOT EXISTS artifacts (
  artifact_id TEXT NOT NULL,
  version INTEGER NOT NULL,
  schema_version TEXT NOT NULL,
  mission_id TEXT NOT NULL,
  run_id TEXT,
  kind TEXT NOT NULL,
  title TEXT NOT NULL,
  uri TEXT NOT NULL,
  content_type TEXT,
  path TEXT NOT NULL,
  trace_id TEXT NOT NULL,
  created_at TEXT NOT NULL,
  PRIMARY KEY (artifact_id, version)
);
CREATE INDEX IF NOT EXISTS idx_artifacts_mission_id ON artifacts (mission_id);
CREATE INDEX IF NOT EXISTS idx_artifacts_run_id ON artifacts (run_id);
CREATE TABLE IF NOT EXISTS capabilities (
  capability_id TEXT PRIMARY KEY,
  manifest TEXT NOT NULL,
  lifecycle TEXT NOT NULL,
  trust_level TEXT NOT NULL,
  risk_level TEXT NOT NULL,
  risk_actions TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  last_transition TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_capabilities_lifecycle ON capabilities (lifecycle);
CREATE INDEX IF NOT EXISTS idx_capabilities_trust_level ON capabilities (trust_level);
CREATE TABLE IF NOT EXISTS workspace_state (
  key TEXT PRIMARY KEY,
  payload TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
`)
}
