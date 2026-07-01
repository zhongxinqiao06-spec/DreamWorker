package main

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"testing"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/api/runtimeapi"
)

func TestRunPingWritesJSON(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if err := run([]string{"ping", "--trace-id", "tr_cli"}, &stdout, &stderr); err != nil {
		t.Fatalf("run ping: %v\nstderr: %s", err, stderr.String())
	}

	var response runtimeapi.PingResponse
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("decode ping response: %v", err)
	}
	if !response.OK {
		t.Fatal("expected ok response")
	}
	if response.SchemaVersion != runtimeapi.ContractSchemaVersion {
		t.Fatalf("expected schema version %q, got %q", runtimeapi.ContractSchemaVersion, response.SchemaVersion)
	}
	if response.EngineVersion != runtimeapi.EngineVersion {
		t.Fatalf("expected engine version %q, got %q", runtimeapi.EngineVersion, response.EngineVersion)
	}
	if response.TraceID != "tr_cli" {
		t.Fatalf("expected trace id tr_cli, got %q", response.TraceID)
	}
}

func TestPingCommandJSON(t *testing.T) {
	command := exec.Command("go", "run", ".", "ping", "--trace-id", "tr_integration")

	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("go run ping: %v\n%s", err, string(output))
	}

	var response runtimeapi.PingResponse
	if err := json.Unmarshal(output, &response); err != nil {
		t.Fatalf("decode command output: %v\n%s", err, string(output))
	}
	if response.TraceID != "tr_integration" {
		t.Fatalf("expected trace id tr_integration, got %q", response.TraceID)
	}
	if response.SchemaVersion != runtimeapi.ContractSchemaVersion {
		t.Fatalf("expected schema version %q, got %q", runtimeapi.ContractSchemaVersion, response.SchemaVersion)
	}
}
