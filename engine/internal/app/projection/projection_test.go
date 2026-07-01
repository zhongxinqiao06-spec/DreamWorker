package projection

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/domain"
)

func TestReplayMission(t *testing.T) {
	store := fakeEventStore{
		missionEvents: []domain.DomainEvent{
			projectionEvent("evt_001", "run_001", "mission.created"),
			projectionEvent("evt_002", "run_001", "mission.updated"),
		},
	}

	projection, err := ReplayMission(context.Background(), store, "msn_001")
	if err != nil {
		t.Fatalf("replay mission: %v", err)
	}

	if projection.MissionID != "msn_001" {
		t.Fatalf("expected mission id msn_001, got %q", projection.MissionID)
	}
	if projection.EventCount != 2 {
		t.Fatalf("expected two events, got %d", projection.EventCount)
	}
	if projection.LastEventID != "evt_002" {
		t.Fatalf("expected last event evt_002, got %q", projection.LastEventID)
	}
	if projection.LastEventType != "mission.updated" {
		t.Fatalf("expected last event type mission.updated, got %q", projection.LastEventType)
	}
}

func TestReplayRun(t *testing.T) {
	store := fakeEventStore{
		runEvents: []domain.DomainEvent{
			projectionEvent("evt_001", "run_001", "run.started"),
			projectionEvent("evt_002", "run_001", "run.completed"),
		},
	}

	projection, err := ReplayRun(context.Background(), store, "run_001")
	if err != nil {
		t.Fatalf("replay run: %v", err)
	}

	if projection.RunID != "run_001" {
		t.Fatalf("expected run id run_001, got %q", projection.RunID)
	}
	if projection.EventCount != 2 {
		t.Fatalf("expected two events, got %d", projection.EventCount)
	}
	if projection.LastEventID != "evt_002" {
		t.Fatalf("expected last event evt_002, got %q", projection.LastEventID)
	}
	if projection.LastEventType != "run.completed" {
		t.Fatalf("expected last event type run.completed, got %q", projection.LastEventType)
	}
}

type fakeEventStore struct {
	missionEvents []domain.DomainEvent
	runEvents     []domain.DomainEvent
}

func (store fakeEventStore) Append(context.Context, []domain.DomainEvent) error {
	return nil
}

func (store fakeEventStore) LoadMission(context.Context, string) ([]domain.DomainEvent, error) {
	return store.missionEvents, nil
}

func (store fakeEventStore) LoadRun(context.Context, string) ([]domain.DomainEvent, error) {
	return store.runEvents, nil
}

func projectionEvent(eventID string, runID string, eventType string) domain.DomainEvent {
	payload, _ := json.Marshal(map[string]string{"title": "AI 项目孵化器"})
	return domain.DomainEvent{
		EventID:       eventID,
		SchemaVersion: domain.EventSchemaVersion,
		TraceID:       "tr_projection",
		MissionID:     "msn_001",
		RunID:         runID,
		Actor:         "orchestrator",
		Timestamp:     time.Date(2026, 6, 30, 8, 0, 0, 0, time.UTC),
		Type:          eventType,
		Payload:       payload,
	}
}
