package runtimeapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRuntimePingHandlerRequiresToken(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/runtime/ping", nil)
	recorder := httptest.NewRecorder()

	RuntimePingHandler("secret-token")(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if response.Code != "UNAUTHORIZED" {
		t.Fatalf("expected unauthorized code, got %q", response.Code)
	}
	if !response.Recoverable {
		t.Fatal("expected recoverable unauthorized response")
	}
	if response.UserAction == "" {
		t.Fatal("expected user action")
	}
	if response.TraceID == "" {
		t.Fatal("expected trace id")
	}
}

func TestRuntimePingHandlerReturnsPing(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/runtime/ping", nil)
	request.Header.Set("Authorization", "Bearer secret-token")
	recorder := httptest.NewRecorder()

	RuntimePingHandler("secret-token")(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var response PingResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if !response.OK {
		t.Fatal("expected ok response")
	}
	if response.SchemaVersion != ContractSchemaVersion {
		t.Fatalf("expected schema version %q, got %q", ContractSchemaVersion, response.SchemaVersion)
	}
	if response.EngineVersion != EngineVersion {
		t.Fatalf("expected engine version %q, got %q", EngineVersion, response.EngineVersion)
	}
	if response.TraceID == "" {
		t.Fatal("expected trace id")
	}
}
