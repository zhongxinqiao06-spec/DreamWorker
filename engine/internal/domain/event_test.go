package domain

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDomainEventValidatesContractEnvelope(t *testing.T) {
	event := DomainEvent{
		EventID:       "evt_contract",
		SchemaVersion: EventSchemaVersion,
		TraceID:       "tr_contract",
		MissionID:     "msn_contract",
		RunID:         "run_contract",
		Actor:         "orchestrator",
		Timestamp:     time.Date(2026, 6, 30, 8, 0, 0, 0, time.UTC),
		Type:          "mission.created",
		Payload:       json.RawMessage(`{"title":"AI 项目孵化器"}`),
	}

	if err := event.Validate(); err != nil {
		t.Fatalf("validate event: %v", err)
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}

	var envelope map[string]any
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatalf("unmarshal event: %v", err)
	}
	for _, key := range []string{
		"event_id",
		"schema_version",
		"trace_id",
		"mission_id",
		"run_id",
		"actor",
		"timestamp",
		"type",
		"payload",
	} {
		if _, ok := envelope[key]; !ok {
			t.Fatalf("expected event envelope key %q", key)
		}
	}
}

func TestDomainEventRejectsInvalidPayload(t *testing.T) {
	event := DomainEvent{
		EventID:       "evt_contract",
		SchemaVersion: EventSchemaVersion,
		TraceID:       "tr_contract",
		MissionID:     "msn_contract",
		RunID:         "run_contract",
		Actor:         "orchestrator",
		Timestamp:     time.Date(2026, 6, 30, 8, 0, 0, 0, time.UTC),
		Type:          "mission.created",
		Payload:       json.RawMessage(`[]`),
	}

	if err := event.Validate(); err == nil {
		t.Fatal("expected invalid payload error")
	}
}
