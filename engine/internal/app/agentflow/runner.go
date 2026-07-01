package agentflow

import (
	"context"
	"fmt"

	runtimeagent "github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/runtime/agent"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/domain"
	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/ports"
)

type Runner struct {
	executor *runtimeagent.Executor
	events   ports.EventStore
	clock    ports.Clock
	ids      ports.IdGenerator
}

type RunResult struct {
	Execution runtimeagent.ExecutionResult
	Eval      domain.EvalReport
}

type GoldenTask struct {
	ID   string
	Idea string
}

type GoldenTaskResult struct {
	TaskID string
	Result RunResult
}

func NewRunner(
	executor *runtimeagent.Executor,
	events ports.EventStore,
	clock ports.Clock,
	ids ports.IdGenerator,
) *Runner {
	return &Runner{
		executor: executor,
		events:   events,
		clock:    clock,
		ids:      ids,
	}
}

func (runner *Runner) Run(ctx context.Context, graph domain.TaskGraph) (RunResult, error) {
	execution, err := runner.executor.Execute(ctx, graph)
	if err != nil {
		return RunResult{}, err
	}
	report := Evaluate(graph, execution.ArtifactIDs, runner.ids.NewID("eval"))
	if err := runner.appendEvalEvent(ctx, graph, report); err != nil {
		return RunResult{}, err
	}
	return RunResult{Execution: execution, Eval: report}, nil
}

func (runner *Runner) RunGoldenTasks(ctx context.Context) ([]GoldenTaskResult, error) {
	tasks := GoldenTasks()
	results := make([]GoldenTaskResult, 0, len(tasks))
	for index, task := range tasks {
		missionID := fmt.Sprintf("msn_golden_%03d", index+1)
		runID := fmt.Sprintf("run_golden_%03d", index+1)
		traceID := fmt.Sprintf("tr_golden_%03d", index+1)
		result, err := runner.Run(ctx, MVPTaskGraph(missionID, runID, traceID, task.Idea))
		if err != nil {
			return nil, err
		}
		results = append(results, GoldenTaskResult{TaskID: task.ID, Result: result})
	}
	return results, nil
}

func Evaluate(graph domain.TaskGraph, artifactIDs []string, reportID string) domain.EvalReport {
	expected := expectedArtifactCount(graph.Tasks)
	completeness := ratio(len(artifactIDs), expected)
	evidenceScore := 0.7
	if containsArtifact(artifactIDs, "art_research_pack") && containsArtifact(artifactIDs, "art_evidence_graph") {
		evidenceScore = 0.86
	}
	hallucinationRisk := domain.RiskMedium
	if evidenceScore >= 0.8 && completeness >= 0.9 {
		hallucinationRisk = domain.RiskLow
	}
	actionability := 0.78
	if containsArtifact(artifactIDs, "art_blueprint") {
		actionability = 0.88
	}
	return domain.EvalReport{
		SchemaVersion:        domain.ContractSchemaVersion,
		ReportID:             reportID,
		MissionID:            graph.MissionID,
		RunID:                graph.RunID,
		TraceID:              graph.TraceID,
		ArtifactScore:        completeness,
		EvidenceQualityScore: evidenceScore,
		HallucinationRisk:    hallucinationRisk,
		ActionabilityScore:   actionability,
		NextBestAction:       "进入 PR-07 UI 工作台接入前，保持 stub 模式跑通端到端黄金任务。",
		CheckedArtifacts:     append([]string{}, artifactIDs...),
	}
}

func (runner *Runner) appendEvalEvent(ctx context.Context, graph domain.TaskGraph, report domain.EvalReport) error {
	payload, err := domain.EncodePayload(domain.EvalReportCreatedPayload{Report: report})
	if err != nil {
		return err
	}
	event := domain.DomainEvent{
		EventID:       runner.ids.NewID("evt"),
		SchemaVersion: domain.EventSchemaVersion,
		TraceID:       graph.TraceID,
		MissionID:     graph.MissionID,
		RunID:         graph.RunID,
		Actor:         AgentEvaluator,
		Timestamp:     runner.clock.Now(),
		Type:          domain.EventEvalReportCreated,
		Payload:       payload,
	}
	if err := event.Validate(); err != nil {
		return err
	}
	return runner.events.Append(ctx, []domain.DomainEvent{event})
}

func expectedArtifactCount(tasks []domain.Task) int {
	total := 0
	for _, task := range tasks {
		total += len(task.ExpectedArtifacts)
	}
	return total
}

func ratio(value int, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(value) / float64(total)
}

func containsArtifact(artifactIDs []string, artifactID string) bool {
	for _, candidate := range artifactIDs {
		if candidate == artifactID {
			return true
		}
	}
	return false
}
