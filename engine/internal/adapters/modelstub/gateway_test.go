package modelstub

import (
	"context"
	"errors"
	"testing"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/domain"
)

func TestGenerateReturnsDeterministicNormalizedOutput(t *testing.T) {
	response, err := NewGateway().Generate(context.Background(), testRequest())
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if response.SchemaVersion != domain.ContractSchemaVersion {
		t.Fatalf("expected schema version %s, got %s", domain.ContractSchemaVersion, response.SchemaVersion)
	}
	normalized, err := domain.NormalizeModelOutput(response.StructuredOutput)
	if err != nil {
		t.Fatalf("normalize stub response: %v", err)
	}
	if len(normalized.Artifacts) != 1 || normalized.Artifacts[0].ArtifactID != "art_dream_brief" {
		t.Fatalf("unexpected artifacts %#v", normalized.Artifacts)
	}
}

func TestGenerateHonorsBudgetAndCancellation(t *testing.T) {
	request := testRequest()
	request.Budget.MaxTokens = 1
	if _, err := NewGateway().Generate(context.Background(), request); !errors.Is(err, domain.ErrModelBudgetExceeded) {
		t.Fatalf("expected budget error, got %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := NewGateway().Generate(ctx, testRequest()); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
}

func TestStreamReturnsSingleFinalEvent(t *testing.T) {
	stream, err := NewGateway().Stream(context.Background(), testRequest())
	if err != nil {
		t.Fatalf("stream: %v", err)
	}
	event := <-stream
	if !event.Done || event.Response == nil {
		t.Fatalf("expected final response event, got %#v", event)
	}
}

func testRequest() domain.ModelRequest {
	return domain.ModelRequest{
		SchemaVersion:     domain.ContractSchemaVersion,
		RequestID:         "mdlreq_001",
		TraceID:           "tr_001",
		MissionID:         "msn_001",
		RunID:             "run_001",
		TaskID:            "tsk_001",
		AgentID:           "product_analyst",
		Stage:             domain.StageDiscover,
		Goal:              "Generate brief",
		Idea:              "AI project incubator",
		ModelProfile:      "stub_reasoning_light",
		ExpectedArtifacts: []string{"dream_brief.md"},
		Budget:            domain.Budget{MaxTokens: 1000, MaxCostUSD: 0},
		PromptRef: domain.PromptRef{
			PromptID:      "prm_product_analyst",
			PromptVersion: "v1",
			AgentID:       "product_analyst",
		},
	}
}
