package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/resources"
)

const workspaceSnapshotKey = "workspace"

type WorkspaceStatePersistence struct {
	db *sql.DB
}

func WorkspacePersistenceOptions(configDir string) ([]resources.StoreOption, error) {
	persistence, err := OpenWorkspaceStatePersistence(context.Background(), configDir)
	if err != nil || persistence == nil {
		return nil, err
	}
	options := []resources.StoreOption{
		resources.WithWorkspacePersistence(persistence.Save),
		resources.WithWorkspacePersistenceClose(persistence.Close),
	}
	snapshot, ok, err := persistence.Load(context.Background())
	if err != nil {
		_ = persistence.Close()
		return nil, err
	}
	if ok {
		options = append(options, resources.WithInitialWorkspaceSnapshot(snapshot))
	}
	return options, nil
}

func OpenWorkspaceStatePersistence(ctx context.Context, configDir string) (*WorkspaceStatePersistence, error) {
	if configDir == "" {
		return nil, nil
	}
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return nil, err
	}
	db, err := Open(ctx, filepath.Join(configDir, "workspace.db"))
	if err != nil {
		return nil, err
	}
	if err := Bootstrap(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &WorkspaceStatePersistence{db: db}, nil
}

func (p *WorkspaceStatePersistence) Load(ctx context.Context) (resources.WorkspaceSnapshot, bool, error) {
	var payload string
	err := p.db.QueryRowContext(
		ctx,
		"SELECT payload FROM workspace_state WHERE key = ?",
		workspaceSnapshotKey,
	).Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) {
		return resources.WorkspaceSnapshot{}, false, nil
	}
	if err != nil {
		return resources.WorkspaceSnapshot{}, false, err
	}
	var snapshot resources.WorkspaceSnapshot
	if err := json.Unmarshal([]byte(payload), &snapshot); err != nil {
		return resources.WorkspaceSnapshot{}, false, err
	}
	return snapshot, true, nil
}

func (p *WorkspaceStatePersistence) Save(store *resources.Store) *resources.AppError {
	snapshot := store.CaptureWorkspaceSnapshotLocked()
	payload, err := json.Marshal(snapshot)
	if err != nil {
		return resources.Internal("WORKSPACE_PERSIST_ENCODE_FAILED", "failed to encode workspace state", "check workspace resource configuration")
	}
	if _, err := p.db.ExecContext(
		context.Background(),
		`INSERT INTO workspace_state (key, payload, updated_at)
VALUES (?, ?, ?)
ON CONFLICT(key) DO UPDATE SET payload = excluded.payload, updated_at = excluded.updated_at`,
		workspaceSnapshotKey,
		string(payload),
		store.Now(),
	); err != nil {
		return resources.Internal("WORKSPACE_PERSIST_WRITE_FAILED", "failed to write workspace SQLite state", "check DreamWorker config directory permissions")
	}
	return nil
}

func (p *WorkspaceStatePersistence) Close() error {
	if p == nil || p.db == nil {
		return nil
	}
	return p.db.Close()
}
