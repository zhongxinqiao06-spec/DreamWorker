package incubator

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"testing"
	"time"

	sqliteadapter "github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/adapters/sqlite"
	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/decisiongate"
	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/domain"
)

func TestCreateMissionStartsDiscoverAndAppendsEvents(t *testing.T) {
	service, store := newTestService()

	mission, err := service.CreateMission(context.Background(), CreateMissionCommand{
		Title:   "AI 项目孵化器",
		Idea:    "我想做一个面向独立开发者的 AI 项目孵化工具。",
		TraceID: "tr_fixture",
	})
	if err != nil {
		t.Fatalf("create mission: %v", err)
	}

	if mission.ID != "msn_001" {
		t.Fatalf("expected mission id msn_001, got %q", mission.ID)
	}
	if mission.CurrentStage != domain.StageDiscover {
		t.Fatalf("expected current stage Discover, got %q", mission.CurrentStage)
	}
	if mission.Stages[domain.StageDiscover].Status != domain.StageStatusRunning {
		t.Fatalf("expected Discover running, got %q", mission.Stages[domain.StageDiscover].Status)
	}
	if len(store.eventsByMission["msn_001"]) != 2 {
		t.Fatalf("expected mission.created and stage.started events, got %d", len(store.eventsByMission["msn_001"]))
	}
}

func TestServiceRecordsStageObjectsThroughEventStore(t *testing.T) {
	service, store := newTestService()
	ctx := context.Background()
	mission := createMission(t, ctx, service)
	runID := onlyRunID(t, mission)

	mission, err := service.RecordHypothesis(ctx, RecordHypothesisCommand{
		MissionID:  mission.ID,
		RunID:      runID,
		Stage:      domain.StageDiscover,
		Statement:  "独立开发者需要结构化 AI 项目孵化流程。",
		OwnerAgent: "product_analyst",
		TraceID:    "tr_fixture",
	})
	if err != nil {
		t.Fatalf("record hypothesis: %v", err)
	}
	if _, ok := mission.Hypotheses["hyp_001"]; !ok {
		t.Fatal("expected hypothesis hyp_001")
	}

	mission, err = service.RecordEvidence(ctx, RecordEvidenceCommand{
		MissionID:      mission.ID,
		RunID:          runID,
		Stage:          domain.StageDiscover,
		Source:         "seed",
		Summary:        "用户输入明确要求从想法生成可执行蓝图。",
		Confidence:     0.82,
		Risk:           domain.RiskMedium,
		NextBestAction: "继续验证付费意愿。",
		Bindings: []domain.EvidenceBinding{{
			TargetKind: domain.EvidenceTargetHypothesis,
			TargetID:   "hyp_001",
		}},
		TraceID: "tr_fixture",
	})
	if err != nil {
		t.Fatalf("record evidence: %v", err)
	}
	if mission.EvidenceGraph.ByHypothesis["hyp_001"][0] != "ev_001" {
		t.Fatalf("expected evidence graph binding, got %#v", mission.EvidenceGraph.ByHypothesis)
	}

	mission, err = service.RecordExperiment(ctx, RecordExperimentCommand{
		MissionID:   mission.ID,
		RunID:       runID,
		Stage:       domain.StageDiscover,
		Goal:        "验证目标用户是否有该问题。",
		Method:      "访谈 5 个独立开发者。",
		EvidenceIDs: []string{"ev_001"},
		TraceID:     "tr_fixture",
	})
	if err != nil {
		t.Fatalf("record experiment: %v", err)
	}
	if _, ok := mission.Experiments["exp_001"]; !ok {
		t.Fatal("expected experiment exp_001")
	}

	mission, err = service.RecordDecision(ctx, RecordDecisionCommand{
		MissionID:      mission.ID,
		RunID:          runID,
		Stage:          domain.StageDiscover,
		Type:           domain.DecisionContinue,
		Confidence:     0.82,
		Reason:         "证据足够进入下一阶段。",
		EvidenceRefs:   []string{"ev_001"},
		Risks:          []string{"market_uncertainty"},
		NextBestAction: "进入 Validate。",
		TraceID:        "tr_fixture",
	})
	if err != nil {
		t.Fatalf("record decision: %v", err)
	}
	if mission.Decisions["dec_001"].Type != domain.DecisionContinue {
		t.Fatalf("expected continue decision, got %q", mission.Decisions["dec_001"].Type)
	}

	mission, err = service.CompleteStage(ctx, CompleteStageCommand{
		MissionID:  mission.ID,
		RunID:      runID,
		Stage:      domain.StageDiscover,
		DecisionID: "dec_001",
		TraceID:    "tr_fixture",
	})
	if err != nil {
		t.Fatalf("complete stage: %v", err)
	}
	if mission.Stages[domain.StageDiscover].Status != domain.StageStatusCompleted {
		t.Fatalf("expected Discover completed, got %q", mission.Stages[domain.StageDiscover].Status)
	}
	if len(store.eventsByMission["msn_001"]) != 7 {
		t.Fatalf("expected seven appended events, got %d", len(store.eventsByMission["msn_001"]))
	}
}

func TestStageTransitionRules(t *testing.T) {
	service, _ := newTestService()
	ctx := context.Background()
	mission := createMission(t, ctx, service)
	runID := onlyRunID(t, mission)

	if _, err := service.StartStage(ctx, StartStageCommand{
		MissionID: mission.ID,
		RunID:     runID,
		Stage:     domain.StageValidate,
		TraceID:   "tr_fixture",
	}); err == nil {
		t.Fatal("expected Validate to require completed Discover")
	}

	if _, err := service.StartStage(ctx, StartStageCommand{
		MissionID: mission.ID,
		RunID:     runID,
		Stage:     domain.StageBuild,
		TraceID:   "tr_fixture",
	}); err == nil {
		t.Fatal("expected Build auto-run to be rejected")
	}
}

func TestEvidenceRequiresBindingBeforeAppend(t *testing.T) {
	service, store := newTestService()
	ctx := context.Background()
	mission := createMission(t, ctx, service)
	runID := onlyRunID(t, mission)
	before := len(store.eventsByMission[mission.ID])

	if _, err := service.RecordEvidence(ctx, RecordEvidenceCommand{
		MissionID:      mission.ID,
		RunID:          runID,
		Stage:          domain.StageDiscover,
		Source:         "seed",
		Summary:        "缺少绑定。",
		Confidence:     0.4,
		Risk:           domain.RiskHigh,
		NextBestAction: "补充绑定。",
		TraceID:        "tr_fixture",
	}); err == nil {
		t.Fatal("expected evidence binding error")
	}
	if len(store.eventsByMission[mission.ID]) != before {
		t.Fatal("expected invalid evidence to avoid appending events")
	}
}

func TestDiscoverValidateShapeFixtureFlow(t *testing.T) {
	service, _ := newTestService()
	ctx := context.Background()
	mission := createMission(t, ctx, service)
	runID := onlyRunID(t, mission)

	mission = runStageFixture(t, ctx, service, mission, runID, domain.StageDiscover, "hyp_001")
	mission = startStage(t, ctx, service, mission, runID, domain.StageValidate)
	mission = runStageFixture(t, ctx, service, mission, runID, domain.StageValidate, "hyp_002")
	mission = startStage(t, ctx, service, mission, runID, domain.StageShape)
	mission = runStageFixture(t, ctx, service, mission, runID, domain.StageShape, "hyp_003")

	for _, stage := range []domain.StageName{domain.StageBuild, domain.StageLaunch, domain.StageLearn} {
		var err error
		mission, err = service.CreatePlaceholderStage(ctx, CreatePlaceholderStageCommand{
			MissionID:  mission.ID,
			RunID:      runID,
			Stage:      stage,
			NextAction: "后续阶段接入 Agent Runtime 后执行。",
			TraceID:    "tr_fixture",
		})
		if err != nil {
			t.Fatalf("create %s placeholder: %v", stage, err)
		}
		if !mission.Stages[stage].Placeholder {
			t.Fatalf("expected %s placeholder", stage)
		}
	}

	for _, stage := range []domain.StageName{domain.StageDiscover, domain.StageValidate, domain.StageShape} {
		if mission.Stages[stage].Status != domain.StageStatusCompleted {
			t.Fatalf("expected %s completed, got %s", stage, mission.Stages[stage].Status)
		}
	}
	if len(mission.Hypotheses) != 3 || len(mission.Evidence) != 3 || len(mission.Decisions) != 3 {
		t.Fatalf("expected 3 hypotheses/evidence/decisions, got %d/%d/%d", len(mission.Hypotheses), len(mission.Evidence), len(mission.Decisions))
	}
}

func TestServicePersistsTransitionsThroughSQLiteEventStore(t *testing.T) {
	ctx := context.Background()
	db, err := sqliteadapter.Open(ctx, filepath.Join(t.TempDir(), "engine.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	if err := sqliteadapter.Bootstrap(ctx, db); err != nil {
		t.Fatalf("bootstrap sqlite: %v", err)
	}

	store := sqliteadapter.NewEventStore(db)
	service := NewService(store, fakeClock{}, newDeterministicIDGenerator())
	mission := createMission(t, ctx, service)
	runID := onlyRunID(t, mission)
	mission = runStageFixture(t, ctx, service, mission, runID, domain.StageDiscover, "hyp_001")

	events, err := store.LoadMission(ctx, mission.ID)
	if err != nil {
		t.Fatalf("load persisted mission events: %v", err)
	}
	replayed, err := domain.ReplayMission(events)
	if err != nil {
		t.Fatalf("replay persisted mission: %v", err)
	}
	if replayed.Stages[domain.StageDiscover].Status != domain.StageStatusCompleted {
		t.Fatalf("expected persisted Discover completed, got %s", replayed.Stages[domain.StageDiscover].Status)
	}
	if replayed.Decisions["dec_001"].Type != domain.DecisionContinue {
		t.Fatalf("expected persisted continue decision, got %s", replayed.Decisions["dec_001"].Type)
	}
}

func runStageFixture(
	t *testing.T,
	ctx context.Context,
	service *Service,
	mission domain.Mission,
	runID string,
	stage domain.StageName,
	hypothesisID string,
) domain.Mission {
	t.Helper()
	var err error
	mission, err = service.RecordHypothesis(ctx, RecordHypothesisCommand{
		MissionID:  mission.ID,
		RunID:      runID,
		Stage:      stage,
		Statement:  fmt.Sprintf("%s 阶段假设。", stage),
		OwnerAgent: "product_analyst",
		TraceID:    "tr_fixture",
	})
	if err != nil {
		t.Fatalf("record %s hypothesis: %v", stage, err)
	}
	mission, err = service.RecordEvidence(ctx, RecordEvidenceCommand{
		MissionID:      mission.ID,
		RunID:          runID,
		Stage:          stage,
		Source:         "fixture",
		Summary:        fmt.Sprintf("%s 阶段证据。", stage),
		Confidence:     0.82,
		Risk:           domain.RiskMedium,
		NextBestAction: "进入下一阶段。",
		Bindings: []domain.EvidenceBinding{{
			TargetKind: domain.EvidenceTargetHypothesis,
			TargetID:   hypothesisID,
		}},
		TraceID: "tr_fixture",
	})
	if err != nil {
		t.Fatalf("record %s evidence: %v", stage, err)
	}

	evidence := mission.Evidence[fmt.Sprintf("ev_%03d", len(mission.Evidence))]
	decisionDraft, err := decisiongate.Evaluate(decisiongate.Input{
		DecisionID: fmt.Sprintf("dec_%03d", len(mission.Decisions)+1),
		MissionID:  mission.ID,
		Stage:      stage,
		Evidence:   []domain.Evidence{evidence},
		Now:        fixedTime(),
	})
	if err != nil {
		t.Fatalf("evaluate %s decision: %v", stage, err)
	}
	mission, err = service.RecordDecision(ctx, RecordDecisionCommand{
		MissionID:      mission.ID,
		RunID:          runID,
		Stage:          stage,
		Type:           decisionDraft.Type,
		Confidence:     decisionDraft.Confidence,
		Reason:         decisionDraft.Reason,
		EvidenceRefs:   decisionDraft.EvidenceRefs,
		Risks:          decisionDraft.Risks,
		NextBestAction: decisionDraft.NextBestAction,
		TraceID:        "tr_fixture",
	})
	if err != nil {
		t.Fatalf("record %s decision: %v", stage, err)
	}

	decisionID := fmt.Sprintf("dec_%03d", len(mission.Decisions))
	mission, err = service.CompleteStage(ctx, CompleteStageCommand{
		MissionID:  mission.ID,
		RunID:      runID,
		Stage:      stage,
		DecisionID: decisionID,
		TraceID:    "tr_fixture",
	})
	if err != nil {
		t.Fatalf("complete %s stage: %v", stage, err)
	}
	return mission
}

func startStage(
	t *testing.T,
	ctx context.Context,
	service *Service,
	mission domain.Mission,
	runID string,
	stage domain.StageName,
) domain.Mission {
	t.Helper()
	nextMission, err := service.StartStage(ctx, StartStageCommand{
		MissionID: mission.ID,
		RunID:     runID,
		Stage:     stage,
		TraceID:   "tr_fixture",
	})
	if err != nil {
		t.Fatalf("start %s: %v", stage, err)
	}
	return nextMission
}

func createMission(t *testing.T, ctx context.Context, service *Service) domain.Mission {
	t.Helper()
	mission, err := service.CreateMission(ctx, CreateMissionCommand{
		Title:   "AI 项目孵化器",
		Idea:    "我想做一个面向独立开发者的 AI 项目孵化工具。",
		TraceID: "tr_fixture",
	})
	if err != nil {
		t.Fatalf("create mission: %v", err)
	}
	return mission
}

func newTestService() (*Service, *memoryEventStore) {
	store := &memoryEventStore{eventsByMission: map[string][]domain.DomainEvent{}}
	return NewService(store, fakeClock{}, newDeterministicIDGenerator()), store
}

type memoryEventStore struct {
	eventsByMission map[string][]domain.DomainEvent
}

func (store *memoryEventStore) Append(_ context.Context, events []domain.DomainEvent) error {
	for _, event := range events {
		store.eventsByMission[event.MissionID] = append(store.eventsByMission[event.MissionID], event)
	}
	return nil
}

func (store *memoryEventStore) LoadMission(_ context.Context, missionID string) ([]domain.DomainEvent, error) {
	return append([]domain.DomainEvent{}, store.eventsByMission[missionID]...), nil
}

func (store *memoryEventStore) LoadRun(_ context.Context, runID string) ([]domain.DomainEvent, error) {
	var events []domain.DomainEvent
	for _, missionEvents := range store.eventsByMission {
		for _, event := range missionEvents {
			if event.RunID == runID {
				events = append(events, event)
			}
		}
	}
	return events, nil
}

type fakeClock struct{}

func (fakeClock) Now() time.Time {
	return fixedTime()
}

func fixedTime() time.Time {
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

func onlyRunID(t *testing.T, mission domain.Mission) string {
	t.Helper()
	ids := make([]string, 0, len(mission.Runs))
	for id := range mission.Runs {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	if len(ids) != 1 {
		t.Fatalf("expected one run id, got %#v", ids)
	}
	return ids[0]
}
