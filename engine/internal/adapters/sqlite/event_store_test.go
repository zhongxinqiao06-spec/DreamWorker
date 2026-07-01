package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/domain"
)

func TestOpenEnablesWAL(t *testing.T) {
	ctx := context.Background()
	db := openBootstrappedTestDB(t, ctx)

	mode, err := JournalMode(ctx, db)
	if err != nil {
		t.Fatalf("read journal mode: %v", err)
	}
	if mode != "wal" {
		t.Fatalf("expected WAL journal mode, got %q", mode)
	}
}

func TestRunMigrationsIsIdempotent(t *testing.T) {
	ctx := context.Background()
	db := openBootstrappedTestDB(t, ctx)

	if err := Bootstrap(ctx, db); err != nil {
		t.Fatalf("run migrations twice: %v", err)
	}

	var count int
	if err := db.QueryRowContext(
		ctx,
		"SELECT count(*) FROM schema_migrations WHERE version = ?",
		"0001_engine_foundation",
	).Scan(&count); err != nil {
		t.Fatalf("count migrations: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one migration record, got %d", count)
	}
}

func TestRunMigrationsRejectsDestructiveMigration(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t, ctx)

	err := RunMigrations(ctx, db, []Migration{{
		Version:        "9999_destructive",
		SQL:            "DROP TABLE events",
		NonDestructive: false,
	}})

	if !errors.Is(err, ErrDestructiveMigrationRequiresBackup) {
		t.Fatalf("expected destructive migration error, got %v", err)
	}
}

func TestEventStoreAppendAndLoad(t *testing.T) {
	ctx := context.Background()
	db := openBootstrappedTestDB(t, ctx)
	store := NewEventStore(db)

	first := sampleEvent("evt_001", "run_001", "mission.created")
	second := sampleEvent("evt_002", "run_001", "mission.updated")

	if err := store.Append(ctx, []domain.DomainEvent{first, second}); err != nil {
		t.Fatalf("append events: %v", err)
	}

	missionEvents, err := store.LoadMission(ctx, "msn_001")
	if err != nil {
		t.Fatalf("load mission events: %v", err)
	}
	if len(missionEvents) != 2 {
		t.Fatalf("expected two mission events, got %d", len(missionEvents))
	}
	if missionEvents[0].EventID != "evt_001" || missionEvents[1].EventID != "evt_002" {
		t.Fatalf("events not loaded in append order: %#v", missionEvents)
	}

	runEvents, err := store.LoadRun(ctx, "run_001")
	if err != nil {
		t.Fatalf("load run events: %v", err)
	}
	if len(runEvents) != 2 {
		t.Fatalf("expected two run events, got %d", len(runEvents))
	}
}

func TestEventStoreRejectsDuplicateEventID(t *testing.T) {
	ctx := context.Background()
	db := openBootstrappedTestDB(t, ctx)
	store := NewEventStore(db)
	event := sampleEvent("evt_duplicate", "run_001", "mission.created")

	if err := store.Append(ctx, []domain.DomainEvent{event}); err != nil {
		t.Fatalf("append first event: %v", err)
	}
	if err := store.Append(ctx, []domain.DomainEvent{event}); err == nil {
		t.Fatal("expected duplicate event_id error")
	}
}

func TestEventStoreRejectsInvalidEvent(t *testing.T) {
	ctx := context.Background()
	db := openBootstrappedTestDB(t, ctx)
	store := NewEventStore(db)
	event := sampleEvent("bad_id", "run_001", "mission.created")

	if err := store.Append(ctx, []domain.DomainEvent{event}); err == nil {
		t.Fatal("expected invalid event error")
	}
}

func TestEventStoreTraceIndexExists(t *testing.T) {
	ctx := context.Background()
	db := openBootstrappedTestDB(t, ctx)

	var indexName string
	if err := db.QueryRowContext(
		ctx,
		"SELECT name FROM sqlite_master WHERE type = 'index' AND name = 'idx_events_trace_id'",
	).Scan(&indexName); err != nil {
		t.Fatalf("read trace index: %v", err)
	}
	if indexName != "idx_events_trace_id" {
		t.Fatalf("expected trace index, got %q", indexName)
	}
}

func openBootstrappedTestDB(t *testing.T, ctx context.Context) *sql.DB {
	t.Helper()
	db := openTestDB(t, ctx)
	if err := Bootstrap(ctx, db); err != nil {
		t.Fatalf("bootstrap sqlite: %v", err)
	}
	return db
}

func openTestDB(t *testing.T, ctx context.Context) *sql.DB {
	t.Helper()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "engine.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	return db
}

func sampleEvent(eventID string, runID string, eventType string) domain.DomainEvent {
	payload, _ := json.Marshal(map[string]string{"title": "AI 项目孵化器"})
	return domain.DomainEvent{
		EventID:       eventID,
		SchemaVersion: domain.EventSchemaVersion,
		TraceID:       "tr_001",
		MissionID:     "msn_001",
		RunID:         runID,
		Actor:         "orchestrator",
		Timestamp:     time.Date(2026, 6, 30, 8, 0, 0, 0, time.UTC),
		Type:          eventType,
		Payload:       payload,
	}
}
