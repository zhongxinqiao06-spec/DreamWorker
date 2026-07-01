package agentflow

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/adapters/modelstub"
	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/domain"
	runtimeagent "github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/runtime/agent"
)

func TestBuiltinAgentSpecsAndTaskGraphValidate(t *testing.T) {
	specs := BuiltinAgentSpecs()
	if len(specs) != 6 {
		t.Fatalf("expected 6 builtin agents, got %d", len(specs))
	}
	byID := map[string]domain.AgentSpec{}
	for _, spec := range specs {
		if err := spec.Validate(); err != nil {
			t.Fatalf("validate %s: %v", spec.ID, err)
		}
		byID[spec.ID] = spec
	}
	graph := MVPTaskGraph("msn_001", "run_001", "tr_001", "AI project incubator")
	if err := graph.Validate(byID); err != nil {
		t.Fatalf("validate MVP task graph: %v", err)
	}
}

func TestRunnerCreatesArtifactsAndEvalReport(t *testing.T) {
	runner, store := newTestRunner(t)
	result, err := runner.Run(
		context.Background(),
		MVPTaskGraph("msn_001", "run_001", "tr_001", "我想做一个面向独立开发者的 AI 项目孵化工具。"),
	)
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if len(result.Execution.CompletedTaskIDs) != 4 {
		t.Fatalf("expected 4 completed tasks, got %#v", result.Execution.CompletedTaskIDs)
	}
	if len(result.Execution.ArtifactIDs) != 7 {
		t.Fatalf("expected 7 artifact ids, got %#v", result.Execution.ArtifactIDs)
	}
	if result.Eval.ArtifactScore != 1 || result.Eval.HallucinationRisk != domain.RiskLow {
		t.Fatalf("unexpected eval report %#v", result.Eval)
	}
	if !store.hasEvent(domain.EventEvalReportCreated) {
		t.Fatalf("expected eval report event, got %#v", store.eventTypes())
	}
}

func TestGoldenTasksRunDeterministically(t *testing.T) {
	runner, _ := newTestRunner(t)
	results, err := runner.RunGoldenTasks(context.Background())
	if err != nil {
		t.Fatalf("run golden tasks: %v", err)
	}
	if len(results) != 5 {
		t.Fatalf("expected 5 golden task results, got %d", len(results))
	}
	for _, result := range results {
		if len(result.Result.Execution.ArtifactIDs) != 7 {
			t.Fatalf("%s expected 7 artifacts, got %#v", result.TaskID, result.Result.Execution.ArtifactIDs)
		}
		if result.Result.Eval.ActionabilityScore < 0.8 {
			t.Fatalf("%s expected actionable eval, got %#v", result.TaskID, result.Result.Eval)
		}
	}
}

func newTestRunner(t *testing.T) (*Runner, *memoryEventStore) {
	t.Helper()
	store := &memoryEventStore{}
	ids := newDeterministicIDGenerator()
	executor, err := runtimeagent.NewExecutor(
		BuiltinAgentSpecs(),
		modelstub.NewGateway(),
		&recordingCapabilityInvoker{},
		store,
		fakeClock{},
		ids,
	)
	if err != nil {
		t.Fatalf("new executor: %v", err)
	}
	return NewRunner(executor, store, fakeClock{}, ids), store
}

type recordingCapabilityInvoker struct{}

func (invoker *recordingCapabilityInvoker) Invoke(
	_ context.Context,
	request domain.CapabilityInvocationRequest,
) (domain.CapabilityInvocationResult, error) {
	return domain.CapabilityInvocationResult{
		CapabilityID: request.CapabilityID,
		TraceID:      request.TraceID,
		OK:           true,
		Output:       json.RawMessage(`{"ok":true}`),
	}, nil
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
