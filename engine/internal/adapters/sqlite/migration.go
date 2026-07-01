package sqlite

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

var ErrDestructiveMigrationRequiresBackup = errors.New("destructive migration requires backup hook")

type Migration struct {
	Version        string
	SQL            string
	NonDestructive bool
}

func DefaultMigrations() []Migration {
	return []Migration{
		{
			Version:        "0001_engine_foundation",
			NonDestructive: true,
			SQL: `
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
`,
		},
		{
			Version:        "0002_capability_policy_runtime",
			NonDestructive: true,
			SQL: `
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
`,
		},
	}
}

func RunMigrations(ctx context.Context, db *sql.DB, migrations []Migration) error {
	if err := ensureMigrationTable(ctx, db); err != nil {
		return err
	}

	for _, migration := range migrations {
		if migration.Version == "" {
			return errors.New("migration version is required")
		}
		if !migration.NonDestructive {
			return ErrDestructiveMigrationRequiresBackup
		}
		if err := runMigration(ctx, db, migration); err != nil {
			return err
		}
	}
	return nil
}

func ensureMigrationTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
  version TEXT PRIMARY KEY,
  checksum TEXT NOT NULL,
  applied_at TEXT NOT NULL,
  non_destructive INTEGER NOT NULL
)`)
	if err != nil {
		return fmt.Errorf("ensure schema_migrations table: %w", err)
	}
	return nil
}

func runMigration(ctx context.Context, db *sql.DB, migration Migration) error {
	checksum := checksumSQL(migration.SQL)
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", migration.Version, err)
	}
	defer rollbackUnlessCommitted(tx)

	var existingChecksum string
	err = tx.QueryRowContext(
		ctx,
		"SELECT checksum FROM schema_migrations WHERE version = ?",
		migration.Version,
	).Scan(&existingChecksum)
	if err == nil {
		if existingChecksum != checksum {
			return fmt.Errorf("migration %s checksum mismatch", migration.Version)
		}
		return tx.Commit()
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("read migration %s: %w", migration.Version, err)
	}

	if _, err := tx.ExecContext(ctx, migration.SQL); err != nil {
		return fmt.Errorf("apply migration %s: %w", migration.Version, err)
	}
	if _, err := tx.ExecContext(
		ctx,
		"INSERT INTO schema_migrations (version, checksum, applied_at, non_destructive) VALUES (?, ?, ?, ?)",
		migration.Version,
		checksum,
		time.Now().UTC().Format(time.RFC3339Nano),
		1,
	); err != nil {
		return fmt.Errorf("record migration %s: %w", migration.Version, err)
	}
	return tx.Commit()
}

func rollbackUnlessCommitted(tx *sql.Tx) {
	_ = tx.Rollback()
}

func checksumSQL(sqlText string) string {
	sum := sha256.Sum256([]byte(sqlText))
	return hex.EncodeToString(sum[:])
}
