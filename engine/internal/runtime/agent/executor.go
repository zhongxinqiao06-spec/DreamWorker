package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/domain"
	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/ports"
)

type Executor struct {
	specs        map[string]domain.AgentSpec
	models       ports.ModelGateway
	capabilities ports.CapabilityInvoker
	events       ports.EventStore
	clock        ports.Clock
	ids          ports.IdGenerator
}

type ExecutionResult struct {
	CompletedTaskIDs []string
	ArtifactIDs      []string
}

func NewExecutor(
	specs []domain.AgentSpec,
	models ports.ModelGateway,
	capabilities ports.CapabilityInvoker,
	events ports.EventStore,
	clock ports.Clock,
	ids ports.IdGenerator,
) (*Executor, error) {
	byID := make(map[string]domain.AgentSpec, len(specs))
	for _, spec := range specs {
		if err := spec.Validate(); err != nil {
			return nil, err
		}
		if _, exists := byID[spec.ID]; exists {
			return nil, fmt.Errorf("%w: duplicate agent %s", domain.ErrAgentSpecInvalid, spec.ID)
		}
		byID[spec.ID] = spec
	}
	return &Executor{
		specs:        byID,
		models:       models,
		capabilities: capabilities,
		events:       events,
		clock:        clock,
		ids:          ids,
	}, nil
}

func (executor *Executor) Execute(
	ctx context.Context,
	graph domain.TaskGraph,
) (ExecutionResult, error) {
	if err := graph.Validate(executor.specs); err != nil {
		return ExecutionResult{}, err
	}
	if err := executor.appendEvent(ctx, graph, "orchestrator", domain.EventAgentTaskGraphCreated, domain.TaskGraphCreatedPayload{
		TaskIDs: taskIDs(graph.Tasks),
	}); err != nil {
		return ExecutionResult{}, err
	}

	var result ExecutionResult
	for _, task := range executionOrder(graph.Tasks) {
		taskResult, err := executor.executeTask(ctx, graph, task)
		if err != nil {
			return result, err
		}
		result.CompletedTaskIDs = append(result.CompletedTaskIDs, task.TaskID)
		result.ArtifactIDs = append(result.ArtifactIDs, taskResult.ArtifactIDs...)
	}
	return result, nil
}

type taskResult struct {
	ArtifactIDs []string
}

func (executor *Executor) executeTask(
	ctx context.Context,
	graph domain.TaskGraph,
	task domain.Task,
) (taskResult, error) {
	spec := executor.specs[task.AssignedAgent]
	taskCtx, cancel := context.WithTimeout(ctx, spec.TimeoutDuration())
	defer cancel()

	if err := executor.appendTaskEvent(taskCtx, graph, task, domain.EventAgentTaskStarted, domain.TaskStatusRunning, ""); err != nil {
		return taskResult{}, err
	}

	capabilityContext, err := executor.invokeTaskCapabilities(taskCtx, graph, task, spec)
	if err != nil {
		return executor.failTask(ctx, graph, task, err)
	}

	requestID := executor.ids.NewID("mdlreq")
	modelRequest := domain.ModelRequest{
		SchemaVersion:     domain.ContractSchemaVersion,
		RequestID:         requestID,
		TraceID:           graph.TraceID,
		MissionID:         graph.MissionID,
		RunID:             graph.RunID,
		TaskID:            task.TaskID,
		AgentID:           spec.ID,
		Stage:             task.Stage,
		Goal:              task.Goal,
		Idea:              graph.Idea,
		ModelProfile:      spec.DefaultModelProfile,
		PromptRef:         spec.PromptRef,
		OutputSchema:      spec.OutputSchema,
		ExpectedArtifacts: task.ExpectedArtifacts,
		CapabilityContext: capabilityContext,
		Budget:            effectiveBudget(task.Budget, spec.Budget),
	}
	if err := executor.appendEvent(taskCtx, graph, spec.ID, domain.EventModelRequested, domain.ModelRequestedPayload{
		RequestID:    requestID,
		TaskID:       task.TaskID,
		AgentID:      spec.ID,
		ModelProfile: spec.DefaultModelProfile,
		PromptRef:    spec.PromptRef,
	}); err != nil {
		return taskResult{}, err
	}

	modelResponse, err := executor.models.Generate(taskCtx, modelRequest)
	if err != nil {
		if appendErr := executor.appendModelFailed(ctx, graph, task, spec, requestID, err); appendErr != nil {
			return taskResult{}, appendErr
		}
		return executor.failTask(ctx, graph, task, err)
	}
	if err := enforceBudget(modelResponse.Usage, modelRequest.Budget); err != nil {
		if appendErr := executor.appendModelFailed(ctx, graph, task, spec, requestID, err); appendErr != nil {
			return taskResult{}, appendErr
		}
		return executor.failTask(ctx, graph, task, err)
	}
	if err := executor.appendEvent(taskCtx, graph, spec.ID, domain.EventModelCompleted, domain.ModelCompletedPayload{
		RequestID:    requestID,
		ResponseID:   modelResponse.ResponseID,
		TaskID:       task.TaskID,
		AgentID:      spec.ID,
		Usage:        modelResponse.Usage,
		FinishReason: modelResponse.FinishReason,
	}); err != nil {
		return taskResult{}, err
	}

	normalized, err := domain.NormalizeModelOutput(modelResponse.StructuredOutput)
	if err != nil {
		if appendErr := executor.appendEvent(ctx, graph, spec.ID, domain.EventNormalizationFailed, domain.NormalizationFailedPayload{
			TaskID:  task.TaskID,
			Reason:  err.Error(),
			RawSize: len(modelResponse.RawOutput),
		}); appendErr != nil {
			return taskResult{}, appendErr
		}
		return executor.failTask(ctx, graph, task, err)
	}

	artifactIDs := artifactIDs(normalized.Artifacts)
	if err := executor.appendEvent(taskCtx, graph, spec.ID, domain.EventOutputNormalized, domain.OutputNormalizedPayload{
		TaskID:         task.TaskID,
		ArtifactIDs:    artifactIDs,
		NextBestAction: normalized.NextBestAction,
	}); err != nil {
		return taskResult{}, err
	}

	if err := executor.writeArtifacts(taskCtx, graph, task, normalized.Artifacts); err != nil {
		return executor.failTask(ctx, graph, task, err)
	}
	if err := executor.appendTaskEvent(taskCtx, graph, task, domain.EventAgentTaskCompleted, domain.TaskStatusCompleted, ""); err != nil {
		return taskResult{}, err
	}
	return taskResult{ArtifactIDs: artifactIDs}, nil
}

func (executor *Executor) invokeTaskCapabilities(
	ctx context.Context,
	graph domain.TaskGraph,
	task domain.Task,
	spec domain.AgentSpec,
) ([]domain.CapabilityContext, error) {
	var contexts []domain.CapabilityContext
	for _, capabilityID := range task.RequiredCapabilities {
		if capabilityID == domain.CapabilityIDArtifactWrite {
			continue
		}
		if !stringIn(spec.AllowedCapabilities, capabilityID) {
			return nil, fmt.Errorf("%w: %s cannot invoke %s", domain.ErrTaskCapabilityNotAllowed, spec.ID, capabilityID)
		}
		input, err := json.Marshal(map[string]any{
			"task_id": task.TaskID,
			"stage":   task.Stage,
			"goal":    task.Goal,
		})
		if err != nil {
			return nil, err
		}
		result, err := executor.capabilities.Invoke(ctx, domain.CapabilityInvocationRequest{
			MissionID:    graph.MissionID,
			RunID:        graph.RunID,
			TraceID:      graph.TraceID,
			Actor:        spec.ID,
			CapabilityID: capabilityID,
			Input:        input,
		})
		if err != nil {
			return nil, err
		}
		if !result.OK {
			return nil, errors.New(result.ErrorMessage)
		}
		contexts = append(contexts, domain.CapabilityContext{
			CapabilityID: capabilityID,
			Output:       result.Output,
		})
	}
	return contexts, nil
}

func (executor *Executor) writeArtifacts(
	ctx context.Context,
	graph domain.TaskGraph,
	task domain.Task,
	artifacts []domain.ArtifactDraft,
) error {
	if len(artifacts) == 0 {
		return nil
	}
	if !stringIn(task.RequiredCapabilities, domain.CapabilityIDArtifactWrite) {
		return fmt.Errorf("%w: task %s cannot write artifacts without %s", domain.ErrTaskCapabilityNotAllowed, task.TaskID, domain.CapabilityIDArtifactWrite)
	}
	for _, artifact := range artifacts {
		input, err := json.Marshal(map[string]any{
			"artifact_id":  artifact.ArtifactID,
			"kind":         artifact.Kind,
			"title":        artifact.Title,
			"version":      1,
			"content_type": artifact.ContentType,
			"file_name":    artifact.FileName,
			"content":      artifact.Content,
		})
		if err != nil {
			return err
		}
		result, err := executor.capabilities.Invoke(ctx, domain.CapabilityInvocationRequest{
			MissionID:    graph.MissionID,
			RunID:        graph.RunID,
			TraceID:      graph.TraceID,
			Actor:        task.AssignedAgent,
			CapabilityID: domain.CapabilityIDArtifactWrite,
			Input:        input,
		})
		if err != nil {
			return err
		}
		if !result.OK {
			return errors.New(result.ErrorMessage)
		}
	}
	return nil
}

func (executor *Executor) failTask(
	ctx context.Context,
	graph domain.TaskGraph,
	task domain.Task,
	taskErr error,
) (taskResult, error) {
	if err := executor.appendTaskEvent(ctx, graph, task, domain.EventAgentTaskFailed, domain.TaskStatusFailed, taskErr.Error()); err != nil {
		return taskResult{}, err
	}
	return taskResult{}, taskErr
}

func (executor *Executor) appendModelFailed(
	ctx context.Context,
	graph domain.TaskGraph,
	task domain.Task,
	spec domain.AgentSpec,
	requestID string,
	modelErr error,
) error {
	return executor.appendEvent(ctx, graph, spec.ID, domain.EventModelFailed, domain.ModelFailedPayload{
		RequestID: requestID,
		TaskID:    task.TaskID,
		AgentID:   spec.ID,
		Code:      modelErrorCode(modelErr),
		Message:   modelErr.Error(),
	})
}

func (executor *Executor) appendTaskEvent(
	ctx context.Context,
	graph domain.TaskGraph,
	task domain.Task,
	eventType string,
	status domain.TaskStatus,
	reason string,
) error {
	return executor.appendEvent(ctx, graph, task.AssignedAgent, eventType, domain.TaskEventPayload{
		TaskID:        task.TaskID,
		Stage:         task.Stage,
		AssignedAgent: task.AssignedAgent,
		Status:        status,
		Reason:        reason,
	})
}

func (executor *Executor) appendEvent(
	ctx context.Context,
	graph domain.TaskGraph,
	actor string,
	eventType string,
	payload any,
) error {
	encoded, err := domain.EncodePayload(payload)
	if err != nil {
		return err
	}
	event := domain.DomainEvent{
		EventID:       executor.ids.NewID("evt"),
		SchemaVersion: domain.EventSchemaVersion,
		TraceID:       graph.TraceID,
		MissionID:     graph.MissionID,
		RunID:         graph.RunID,
		Actor:         actor,
		Timestamp:     executor.clock.Now(),
		Type:          eventType,
		Payload:       encoded,
	}
	if err := event.Validate(); err != nil {
		return err
	}
	return executor.events.Append(ctx, []domain.DomainEvent{event})
}

func taskIDs(tasks []domain.Task) []string {
	ids := make([]string, 0, len(tasks))
	for _, task := range tasks {
		ids = append(ids, task.TaskID)
	}
	return ids
}

func artifactIDs(artifacts []domain.ArtifactDraft) []string {
	ids := make([]string, 0, len(artifacts))
	for _, artifact := range artifacts {
		ids = append(ids, artifact.ArtifactID)
	}
	return ids
}

func executionOrder(tasks []domain.Task) []domain.Task {
	byID := make(map[string]domain.Task, len(tasks))
	for _, task := range tasks {
		byID[task.TaskID] = task
	}
	completed := map[string]bool{}
	ordered := make([]domain.Task, 0, len(tasks))

	for len(ordered) < len(tasks) {
		var ready []domain.Task
		for _, task := range tasks {
			if completed[task.TaskID] || !dependenciesDone(task.DependsOn, completed) {
				continue
			}
			ready = append(ready, byID[task.TaskID])
		}
		sort.Slice(ready, func(i, j int) bool {
			return ready[i].TaskID < ready[j].TaskID
		})
		for _, task := range ready {
			completed[task.TaskID] = true
			ordered = append(ordered, task)
		}
	}
	return ordered
}

func dependenciesDone(dependsOn []string, completed map[string]bool) bool {
	for _, dependency := range dependsOn {
		if !completed[dependency] {
			return false
		}
	}
	return true
}

func effectiveBudget(taskBudget domain.Budget, specBudget domain.Budget) domain.Budget {
	budget := taskBudget
	if specBudget.MaxTokens > 0 && (budget.MaxTokens == 0 || specBudget.MaxTokens < budget.MaxTokens) {
		budget.MaxTokens = specBudget.MaxTokens
	}
	if specBudget.MaxCostUSD >= 0 && (budget.MaxCostUSD < 0 || specBudget.MaxCostUSD < budget.MaxCostUSD) {
		budget.MaxCostUSD = specBudget.MaxCostUSD
	}
	return budget
}

func enforceBudget(usage domain.ModelUsage, budget domain.Budget) error {
	if budget.MaxTokens > 0 && usage.InputTokens+usage.OutputTokens > budget.MaxTokens {
		return domain.ErrModelBudgetExceeded
	}
	if budget.MaxCostUSD >= 0 && usage.CostUSD > budget.MaxCostUSD {
		return domain.ErrModelBudgetExceeded
	}
	return nil
}

func modelErrorCode(err error) string {
	if errors.Is(err, context.Canceled) {
		return "MODEL_CANCELLED"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "MODEL_TIMEOUT"
	}
	if errors.Is(err, domain.ErrModelBudgetExceeded) {
		return "MODEL_BUDGET_EXCEEDED"
	}
	return "MODEL_GENERATE_FAILED"
}

func stringIn(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
