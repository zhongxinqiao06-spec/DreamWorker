package runtimeapi

import "testing"

func TestPingUsesProvidedTraceID(t *testing.T) {
	response := Ping("tr_unit")

	if !response.OK {
		t.Fatal("expected ping response ok")
	}
	if response.SchemaVersion != ContractSchemaVersion {
		t.Fatalf("expected schema version %q, got %q", ContractSchemaVersion, response.SchemaVersion)
	}
	if response.EngineVersion != EngineVersion {
		t.Fatalf("expected engine version %q, got %q", EngineVersion, response.EngineVersion)
	}
	if response.TraceID != "tr_unit" {
		t.Fatalf("expected trace id tr_unit, got %q", response.TraceID)
	}
}

func TestPingGeneratesTraceID(t *testing.T) {
	response := Ping("")

	if response.TraceID == "" {
		t.Fatal("expected generated trace id")
	}
	if response.SchemaVersion != ContractSchemaVersion {
		t.Fatalf("expected schema version %q, got %q", ContractSchemaVersion, response.SchemaVersion)
	}
	if len(response.TraceID) < 4 || response.TraceID[:3] != "tr_" {
		t.Fatalf("expected trace id prefix tr_, got %q", response.TraceID)
	}
}
