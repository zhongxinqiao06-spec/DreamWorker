package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"
)

func Open(ctx context.Context, path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}
	db.SetMaxOpenConns(1)

	if err := configureDatabase(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func Bootstrap(ctx context.Context, db *sql.DB) error {
	return RunMigrations(ctx, db, DefaultMigrations())
}

func JournalMode(ctx context.Context, db *sql.DB) (string, error) {
	var mode string
	if err := db.QueryRowContext(ctx, "PRAGMA journal_mode").Scan(&mode); err != nil {
		return "", fmt.Errorf("read sqlite journal mode: %w", err)
	}
	return strings.ToLower(mode), nil
}

func configureDatabase(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("enable sqlite foreign keys: %w", err)
	}
	if _, err := db.ExecContext(ctx, "PRAGMA busy_timeout = 5000"); err != nil {
		return fmt.Errorf("set sqlite busy timeout: %w", err)
	}

	var mode string
	if err := db.QueryRowContext(ctx, "PRAGMA journal_mode = WAL").Scan(&mode); err != nil {
		return fmt.Errorf("enable sqlite WAL: %w", err)
	}
	if strings.ToLower(mode) != "wal" {
		return fmt.Errorf("sqlite WAL unavailable, journal_mode=%s", mode)
	}
	return nil
}
