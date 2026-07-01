package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/domain"
	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/ports"
)

var _ ports.EventStore = (*EventStore)(nil)

type EventStore struct {
	db *sql.DB
}

func NewEventStore(db *sql.DB) *EventStore {
	return &EventStore{db: db}
}

func (store *EventStore) Append(ctx context.Context, events []domain.DomainEvent) error {
	if len(events) == 0 {
		return nil
	}

	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin append events: %w", err)
	}
	defer rollbackUnlessCommitted(tx)

	statement, err := tx.PrepareContext(ctx, `
INSERT INTO events (
  event_id,
  schema_version,
  trace_id,
  mission_id,
  run_id,
  actor,
  timestamp,
  type,
  payload
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare append events: %w", err)
	}
	defer statement.Close()

	for _, event := range events {
		if err := event.Validate(); err != nil {
			return fmt.Errorf("validate event %s: %w", event.EventID, err)
		}
		if _, err := statement.ExecContext(
			ctx,
			event.EventID,
			event.SchemaVersion,
			event.TraceID,
			event.MissionID,
			event.RunID,
			event.Actor,
			event.Timestamp.UTC().Format(time.RFC3339Nano),
			event.Type,
			string(event.Payload),
		); err != nil {
			return fmt.Errorf("append event %s: %w", event.EventID, err)
		}
	}

	return tx.Commit()
}

func (store *EventStore) LoadMission(ctx context.Context, missionID string) ([]domain.DomainEvent, error) {
	return store.queryEvents(ctx, `
SELECT event_id, schema_version, trace_id, mission_id, run_id, actor, timestamp, type, payload
FROM events
WHERE mission_id = ?
ORDER BY sequence`, missionID)
}

func (store *EventStore) LoadRun(ctx context.Context, runID string) ([]domain.DomainEvent, error) {
	return store.queryEvents(ctx, `
SELECT event_id, schema_version, trace_id, mission_id, run_id, actor, timestamp, type, payload
FROM events
WHERE run_id = ?
ORDER BY sequence`, runID)
}

func (store *EventStore) queryEvents(
	ctx context.Context,
	query string,
	args ...any,
) ([]domain.DomainEvent, error) {
	rows, err := store.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query events: %w", err)
	}
	defer rows.Close()

	var events []domain.DomainEvent
	for rows.Next() {
		event, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan events: %w", err)
	}
	return events, nil
}

func scanEvent(scanner interface {
	Scan(dest ...any) error
}) (domain.DomainEvent, error) {
	var event domain.DomainEvent
	var timestamp string
	var payload string

	if err := scanner.Scan(
		&event.EventID,
		&event.SchemaVersion,
		&event.TraceID,
		&event.MissionID,
		&event.RunID,
		&event.Actor,
		&timestamp,
		&event.Type,
		&payload,
	); err != nil {
		return domain.DomainEvent{}, fmt.Errorf("scan event row: %w", err)
	}

	parsedTimestamp, err := time.Parse(time.RFC3339Nano, timestamp)
	if err != nil {
		return domain.DomainEvent{}, fmt.Errorf("parse event timestamp: %w", err)
	}
	event.Timestamp = parsedTimestamp
	event.Payload = json.RawMessage(payload)
	return event, nil
}
