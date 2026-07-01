package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/domain"
)

func TestExecuteRunsThroughCapabilityInvokerAndWritesEvents(t *testing.T) {
	store := &memoryEventStore{}
	capabilities := &recordingCapabilityInvoker{}
	executor := newTestExecutor(t, store, capabilities, modelResponseForArtifacts(t, []domain.ArtifactDraft{{
		FileName:    "dream_brief.md",
		ArtifactID:  "art_dream_brief",
		Kind:        "dream_brief",
		Title:       "Dream Brief",
		ContentType: "text/markdown",
		Content:     "brief",
	}}))

	result, err := executor.Execute(context.Background(), testGraph())
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if len(result.CompletedTaskIDs) != 1 || result.CompletedTaskIDs[0] != "tsk_001" {
		t.Fatalf("unexpected completed tasks %#v", result.CompletedTaskIDs)
	}
	if !capabilities.called(domain.CapabilityIDModelGenerateStub) {
		t.Fatal("expected model_generate_stub invocation through CapabilityInvoker")
	}
	if !capabilities.called(domain.CapabilityIDArtifactWrite) {
		t.Fatal("expected artifact write through CapabilityInvoker")
	}
	if !store.hasEvent(domain.EventModelRequested) || !store.hasEvent(domain.EventOutputNormalized) || !store.hasEvent(domain.EventAgentTaskCompleted) {
		t.Fatalf("expected model/normalization/task events, got %#v", store.eventTypes())
	}

	payload := store.payloadFor(t, domain.EventModelRequested)
	if !strings.Contains(payload, "prompt_ref") {
		t.Fatalf("expected prompt_ref in event payload, got %s", payload)
	}
	if strings.Contains(payload, "Use the assigned task goal") {
		t.Fatalf("model event leaked raw prompt text: %s", payload)
	}
}

func TestExecuteRejectsNormalizationFailureBeforeArtifactWrite(t *testing.T) {
	store := &memoryEventStore{}
	capabilities := &recordingCapabilityInvoker{}
	executor := newTestExecutor(t, store, capabilities, domain.ModelResponse{
		SchemaVersion:    domain.ContractSchemaVersion,
		ResponseID:       "mdlres_001",
		RequestID:        "mdlreq_001",
		TraceID:          "tr_001",
		Provider:         "stub",
		Model:            "stub",
		RawOutput:        `{"artifacts":[],"next_best_action":"continue"}`,
		StructuredOutput: json.RawMessage(`{"artifacts":[],"next_best_action":"continue"}`),
		Usage:            domain.ModelUsage{InputTokens: 1, OutputTokens: 1, CostUSD: 0},
		FinishReason:     "stop",
	})

	_, err := executor.Execute(context.Background(), testGraph())
	if !errors.Is(err, domain.ErrNormalizationFailed) {
		t.Fatalf("expected normalization failure, got %v", err)
	}
	if capabilities.called(domain.CapabilityIDArtifactWrite) {
		t.Fatal("artifact write must not run after normalization failure")
	}
	if !store.hasEvent(domain.EventNormalizationFailed) || !store.hasEvent(domain.EventAgentTaskFailed) {
		t.Fatalf("expected normalization failed and task failed events, got %#v", store.eventTypes())
	}
}

func TestExecuteStopsOnBudgetExceeded(t *testing.T) {
	store := &memoryEventStore{}
	capabilities := &recordingCapabilityInvoker{}
	response := modelResponseForArtifacts(t, []domain.ArtifactDraft{{
		FileName:    "dream_brief.md",
		ArtifactID:  "art_dream_brief",
		Kind:        "dream_brief",
		Title:       "Dream Brief",
		ContentType: "text/markdown",
		Content:     "brief",
	}})
	response.Usage = domain.ModelUsage{InputTokens: 1000, OutputTokens: 1000, CostUSD: 0}
	executor := newTestExecutor(t, store, capabilities, response)

	graph := testGraph()
	graph.Tasks[0].Budget.MaxTokens = 10
	_, err := executor.Execute(context.Background(), graph)
	if !errors.Is(err, domain.ErrModelBudgetExceeded) {
		t.Fatalf("expected budget error, got %v", err)
	}
	if capabilities.called(domain.CapabilityIDArtifactWrite) {
		t.Fatal("artifact write must not run after budget failure")
	}
}

func newTestExecutor(
	t *testing.T,
	store *memoryEventStore,
	capabilities *recordingCapabilityInvoker,
	response domain.ModelResponse,
) *Executor {
	t.Helper()
	executor, err := NewExecutor(
		[]domain.AgentSpec{testSpec()},
		&fixedModelGateway{response: response},
		capabilities,
		store,
		fakeClock{},
		newDeterministicIDGenerator(),
	)
	if err != nil {
		t.Fatalf("new executor: %v", err)
	}
	return executor
}

func testSpec() domain.AgentSpec {
	return domain.AgentSpec{
		SchemaVersion:       domain.ContractSchemaVersion,
		ID:                  "product_analyst",
		Role:                "Product analyst",
		InputSchema:         map[string]any{"type": "object"},
		OutputSchema:        map[string]any{"type": "object"},
		AllowedCapabilities: []string{domain.CapabilityIDModelGenerateStub, domain.CapabilityIDArtifactWrite},
		DefaultModelProfile: "stub_reasoning_light",
		Budget:              domain.Budget{MaxTokens: 10000, MaxCostUSD: 0},
		Timeout:             "10s",
		ApprovalPolicy:      domain.ApprovalPolicyOnRisk,
		ExpectedArtifacts:   []string{"dream_brief.md"},
		PromptRef: domain.PromptRef{
			PromptID:      "prm_product_analyst",
			PromptVersion: "v1",
			AgentID:       "product_analyst",
		},
	}
}

func testGraph() domain.TaskGraph {
	return domain.TaskGraph{
		SchemaVersion: domain.ContractSchemaVersion,
		MissionID:     "msn_001",
		RunID:         "run_001",
		TraceID:       "tr_001",
		Idea:          "AI project incubator",
		Tasks: []domain.Task{{
			SchemaVersion:        domain.ContractSchemaVersion,
			TaskID:               "tsk_001",
			Stage:                domain.StageDiscover,
			Goal:                 "Generate brief",
			AssignedAgent:        "product_analyst",
			RequiredCapabilities: []string{domain.CapabilityIDModelGenerateStub, domain.CapabilityIDArtifactWrite},
			ExpectedArtifacts:    []string{"dream_brief.md"},
			Budget:               domain.Budget{MaxTokens: 10000, MaxCostUSD: 0},
			Status:               domain.TaskStatusPending,
			TraceID:              "tr_001",
		}},
	}
}

func modelResponseForArtifacts(t *testing.T, artifacts []domain.ArtifactDraft) domain.ModelResponse {
	t.Helper()
	output, err := json.Marshal(domain.NormalizedOutput{
		Artifacts:      artifacts,
		NextBestAction: "continue",
	})
	if err != nil {
		t.Fatalf("marshal model response: %v", err)
	}
	return domain.ModelResponse{
		SchemaVersion:    domain.ContractSchemaVersion,
		ResponseID:       "mdlres_001",
		RequestID:        "mdlreq_001",
		TraceID:          "tr_001",
		Provider:         "stub",
		Model:            "stub",
		RawOutput:        string(output),
		StructuredOutput: output,
		Usage:            domain.ModelUsage{InputTokens: 1, OutputTokens: 1, CostUSD: 0},
		FinishReason:     "stop",
	}
}

type fixedModelGateway struct {
	response domain.ModelResponse
}

func (gateway *fixedModelGateway) Generate(_ context.Context, request domain.ModelRequest) (domain.ModelResponse, error) {
	response := gateway.response
	response.RequestID = request.RequestID
	response.TraceID = request.TraceID
	return response, nil
}

func (gateway *fixedModelGateway) Stream(_ context.Context, request domain.ModelRequest) (<-chan domain.ModelStreamEvent, error) {
	channel := make(chan domain.ModelStreamEvent, 1)
	response := gateway.response
	response.RequestID = request.RequestID
	response.TraceID = request.TraceID
	channel <- domain.ModelStreamEvent{
		SchemaVersion: domain.ContractSchemaVersion,
		RequestID:     request.RequestID,
		TraceID:       request.TraceID,
		Done:          true,
		Response:      &response,
	}
	close(channel)
	return channel, nil
}

type recordingCapabilityInvoker struct {
	requests []domain.CapabilityInvocationRequest
}

func (invoker *recordingCapabilityInvoker) Invoke(
	_ context.Context,
	request domain.CapabilityInvocationRequest,
) (domain.CapabilityInvocationResult, error) {
	invoker.requests = append(invoker.requests, request)
	return domain.CapabilityInvocationResult{
		CapabilityID: request.CapabilityID,
		TraceID:      request.TraceID,
		OK:           true,
		Output:       json.RawMessage(`{"ok":true}`),
	}, nil
}

func (invoker *recordingCapabilityInvoker) called(capabilityID string) bool {
	for _, request := range invoker.requests {
		if request.CapabilityID == capabilityID {
			return true
		}
	}
	return false
}

type memoryEventStore struct {
	events []domain.DomainEvent
}

func (store *memoryEventStore) Append(_ context.Context, events []domain.DomainEvent) error {
	store.events = append(store.events, events...)
	return nil
}

func (store *memoryEventStore) LoadMission(_ context.Context, missionID string) ([]domain.DomainEvent, error) {
	var events []domain.DomainEvent
	for _, event := range store.events {
		if event.MissionID == missionID {
			events = append(events, event)
		}
	}
	return events, nil
}

func (store *memoryEventStore) LoadRun(_ context.Context, runID string) ([]domain.DomainEvent, error) {
	var events []domain.DomainEvent
	for _, event := range store.events {
		if event.RunID == runID {
			events = append(events, event)
		}
	}
	return events, nil
}

func (store *memoryEventStore) hasEvent(eventType string) bool {
	for _, event := range store.events {
		if event.Type == eventType {
			return true
		}
	}
	return false
}

func (store *memoryEventStore) payloadFor(t *testing.T, eventType string) string {
	t.Helper()
	for _, event := range store.events {
		if event.Type == eventType {
			return string(event.Payload)
		}
	}
	t.Fatalf("event %s not found", eventType)
	return ""
}

func (store *memoryEventStore) eventTypes() []string {
	types := make([]string, 0, len(store.events))
	for _, event := range store.events {
		types = append(types, event.Type)
	}
	return types
}

type fakeClock struct{}

func (fakeClock) Now() time.Time {
	return time.Date(2026, 6, 30, 8, 0, 0, 0, time.UTC)
}

type deterministicIDGenerator struct {
	counts map[string]int
}

func newDeterministicIDGenerator() *deterministicIDGenerator {
	return &deterministicIDGenerator{counts: map[string]int{}}
}

func (generator *deterministicIDGenerator) NewID(prefix string) string {
	generator.counts[prefix]++
	return fmt.Sprintf("%s_%03d", prefix, generator.counts[prefix])
}
