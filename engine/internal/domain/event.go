package domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

const ContractSchemaVersion = "0.1"
const EventSchemaVersion = ContractSchemaVersion

type DomainEvent struct {
	EventID       string          `json:"event_id"`
	SchemaVersion string          `json:"schema_version"`
	TraceID       string          `json:"trace_id"`
	MissionID     string          `json:"mission_id"`
	RunID         string          `json:"run_id"`
	Actor         string          `json:"actor"`
	Timestamp     time.Time       `json:"timestamp"`
	Type          string          `json:"type"`
	Payload       json.RawMessage `json:"payload"`
}

func (event DomainEvent) Validate() error {
	switch {
	case !strings.HasPrefix(event.EventID, "evt_"):
		return errors.New("event_id must start with evt_")
	case event.SchemaVersion != EventSchemaVersion:
		return errors.New("schema_version must be 0.1")
	case !strings.HasPrefix(event.TraceID, "tr_"):
		return errors.New("trace_id must start with tr_")
	case !strings.HasPrefix(event.MissionID, "msn_"):
		return errors.New("mission_id must start with msn_")
	case !strings.HasPrefix(event.RunID, "run_"):
		return errors.New("run_id must start with run_")
	case strings.TrimSpace(event.Actor) == "":
		return errors.New("actor is required")
	case event.Timestamp.IsZero():
		return errors.New("timestamp is required")
	case !strings.Contains(event.Type, "."):
		return errors.New("type must use dotted event name")
	case !isJSONObject(event.Payload):
		return errors.New("payload must be a JSON object")
	default:
		return nil
	}
}

func isJSONObject(payload json.RawMessage) bool {
	trimmed := bytes.TrimSpace(payload)
	return len(trimmed) >= 2 && trimmed[0] == '{' && json.Valid(trimmed)
}
