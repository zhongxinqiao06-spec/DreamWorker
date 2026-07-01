package generated

import (
	"encoding/json"
	"testing"
)

func TestRuntimePingResponseRoundTrip(t *testing.T) {
	response := RuntimePingResponse{
		SchemaVersion: "0.1",
		OK:            true,
		EngineVersion: "0.1.0",
		TraceID:       "tr_contract",
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal runtime ping response: %v", err)
	}

	var decoded RuntimePingResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal runtime ping response: %v", err)
	}

	if decoded.SchemaVersion != "0.1" {
		t.Fatalf("expected schema version 0.1, got %q", decoded.SchemaVersion)
	}
	if !decoded.OK {
		t.Fatal("expected ok response")
	}
	if decoded.EngineVersion != "0.1.0" {
		t.Fatalf("expected engine version 0.1.0, got %q", decoded.EngineVersion)
	}
	if decoded.TraceID != "tr_contract" {
		t.Fatalf("expected trace id tr_contract, got %q", decoded.TraceID)
	}
}

func TestDreamWorkerErrorRoundTrip(t *testing.T) {
	errorEnvelope := DreamWorkerError{
		Code:        "ENGINE_NOT_CONNECTED",
		Message:     "Go Engine 尚未连接，后续阶段会接入本地引擎。",
		Recoverable: true,
		UserAction:  "等待引擎接入后重试。",
		TraceID:     "tr_error",
	}

	data, err := json.Marshal(errorEnvelope)
	if err != nil {
		t.Fatalf("marshal error envelope: %v", err)
	}

	var decoded DreamWorkerError
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error envelope: %v", err)
	}

	if decoded.Code != "ENGINE_NOT_CONNECTED" {
		t.Fatalf("expected code ENGINE_NOT_CONNECTED, got %q", decoded.Code)
	}
	if !decoded.Recoverable {
		t.Fatal("expected recoverable error")
	}
	if decoded.UserAction == "" {
		t.Fatal("expected user action")
	}
	if decoded.TraceID != "tr_error" {
		t.Fatalf("expected trace id tr_error, got %q", decoded.TraceID)
	}
}

func TestEventEnvelopeRoundTrip(t *testing.T) {
	event := EventEnvelope{
		EventID:       "evt_contract",
		SchemaVersion: "0.1",
		TraceID:       "tr_contract",
		MissionID:     "msn_contract",
		RunID:         "run_contract",
		Actor:         "orchestrator",
		Timestamp:     "2026-06-30T00:00:00Z",
		Type:          "mission.created",
		Payload: map[string]any{
			"title": "AI 项目孵化器",
		},
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal event envelope: %v", err)
	}

	var decoded EventEnvelope
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal event envelope: %v", err)
	}

	if decoded.SchemaVersion != "0.1" {
		t.Fatalf("expected schema version 0.1, got %q", decoded.SchemaVersion)
	}
	if decoded.EventID != "evt_contract" {
		t.Fatalf("expected event id evt_contract, got %q", decoded.EventID)
	}
	if decoded.Payload["title"] != "AI 项目孵化器" {
		t.Fatalf("expected payload title, got %#v", decoded.Payload["title"])
	}
}
