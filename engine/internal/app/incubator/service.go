package incubator

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/domain"
	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/ports"
)

type Service struct {
	store ports.EventStore
	clock ports.Clock
	ids   ports.IdGenerator
	actor string
}

func NewService(store ports.EventStore, clock ports.Clock, ids ports.IdGenerator) *Service {
	return &Service{
		store: store,
		clock: clock,
		ids:   ids,
		actor: "incubator",
	}
}

type CreateMissionCommand struct {
	Title   string
	Idea    string
	TraceID string
}

type StartStageCommand struct {
	MissionID string
	RunID     string
	Stage     domain.StageName
	TraceID   string
}

type RecordHypothesisCommand struct {
	MissionID  string
	RunID      string
	Stage      domain.StageName
	Statement  string
	OwnerAgent string
	Status     string
	TraceID    string
}

type RecordEvidenceCommand struct {
	MissionID      string
	RunID          string
	Stage          domain.StageName
	Source         string
	Summary        string
	Confidence     float64
	Risk           domain.RiskLevel
	NextBestAction string
	Bindings       []domain.EvidenceBinding
	TraceID        string
}

type RecordExperimentCommand struct {
	MissionID   string
	RunID       string
	Stage       domain.StageName
	Goal        string
	Method      string
	Status      string
	EvidenceIDs []string
	TraceID     string
}

type RecordDecisionCommand struct {
	MissionID      string
	RunID          string
	Stage          domain.StageName
	Type           domain.DecisionType
	Confidence     float64
	Reason         string
	EvidenceRefs   []string
	Risks          []string
	NextBestAction string
	TraceID        string
}

type CompleteStageCommand struct {
	MissionID  string
	RunID      string
	Stage      domain.StageName
	DecisionID string
	TraceID    string
}

type CreatePlaceholderStageCommand struct {
	MissionID  string
	RunID      string
	Stage      domain.StageName
	NextAction string
	TraceID    string
}

func (service *Service) CreateMission(
	ctx context.Context,
	command CreateMissionCommand,
) (domain.Mission, error) {
	if strings.TrimSpace(command.Idea) == "" {
		return domain.Mission{}, errors.New("idea is required")
	}
	if strings.TrimSpace(command.TraceID) == "" {
		return domain.Mission{}, errors.New("trace_id is required")
	}

	missionID := service.ids.NewID("msn")
	runID := service.ids.NewID("run")
	stageID := service.ids.NewID("stg")
	title := strings.TrimSpace(command.Title)
	if title == "" {
		title = "未命名 Mission"
	}

	events, err := service.events(
		missionID,
		runID,
		command.TraceID,
		domain.EventMissionCreated,
		domain.MissionCreatedPayload{
			MissionID: missionID,
			RunID:     runID,
			Title:     title,
			Idea:      command.Idea,
		},
		domain.EventStageStarted,
		domain.StageStartedPayload{
			StageID: stageID,
			Stage:   domain.StageDiscover,
		},
	)
	if err != nil {
		return domain.Mission{}, err
	}
	return service.appendAndReplay(ctx, missionID, events)
}

func (service *Service) StartStage(
	ctx context.Context,
	command StartStageCommand,
) (domain.Mission, error) {
	mission, err := service.loadMission(ctx, command.MissionID)
	if err != nil {
		return domain.Mission{}, err
	}
	if err := validateAutomaticStageStart(mission, command.Stage); err != nil {
		return domain.Mission{}, err
	}

	stageID := service.ids.NewID("stg")
	event, err := service.event(
		command.MissionID,
		command.RunID,
		command.TraceID,
		domain.EventStageStarted,
		domain.StageStartedPayload{StageID: stageID, Stage: command.Stage},
	)
	if err != nil {
		return domain.Mission{}, err
	}
	return service.appendAndReplay(ctx, command.MissionID, []domain.DomainEvent{event})
}

func (service *Service) RecordHypothesis(
	ctx context.Context,
	command RecordHypothesisCommand,
) (domain.Mission, error) {
	if strings.TrimSpace(command.Statement) == "" {
		return domain.Mission{}, errors.New("hypothesis statement is required")
	}
	if err := service.ensureStageRunning(ctx, command.MissionID, command.Stage); err != nil {
		return domain.Mission{}, err
	}

	hypothesisID := service.ids.NewID("hyp")
	event, err := service.event(
		command.MissionID,
		command.RunID,
		command.TraceID,
		domain.EventHypothesisRecorded,
		domain.HypothesisRecordedPayload{
			HypothesisID: hypothesisID,
			Stage:        command.Stage,
			Statement:    command.Statement,
			OwnerAgent:   command.OwnerAgent,
			Status:       defaultString(command.Status, "testing"),
		},
	)
	if err != nil {
		return domain.Mission{}, err
	}
	return service.appendAndReplay(ctx, command.MissionID, []domain.DomainEvent{event})
}

func (service *Service) RecordEvidence(
	ctx context.Context,
	command RecordEvidenceCommand,
) (domain.Mission, error) {
	if err := service.ensureStageRunning(ctx, command.MissionID, command.Stage); err != nil {
		return domain.Mission{}, err
	}
	evidenceID := service.ids.NewID("ev")
	bindings := make([]domain.EvidenceBinding, len(command.Bindings))
	for index, binding := range command.Bindings {
		binding.EvidenceID = evidenceID
		bindings[index] = binding
	}
	if err := domain.ValidateEvidenceBindings(evidenceID, bindings); err != nil {
		return domain.Mission{}, err
	}

	event, err := service.event(
		command.MissionID,
		command.RunID,
		command.TraceID,
		domain.EventEvidenceRecorded,
		domain.EvidenceRecordedPayload{
			EvidenceID:     evidenceID,
			Stage:          command.Stage,
			Source:         command.Source,
			Summary:        command.Summary,
			Confidence:     command.Confidence,
			Risk:           command.Risk,
			NextBestAction: command.NextBestAction,
			Bindings:       bindings,
		},
	)
	if err != nil {
		return domain.Mission{}, err
	}
	return service.appendAndReplay(ctx, command.MissionID, []domain.DomainEvent{event})
}

func (service *Service) RecordExperiment(
	ctx context.Context,
	command RecordExperimentCommand,
) (domain.Mission, error) {
	if err := service.ensureStageRunning(ctx, command.MissionID, command.Stage); err != nil {
		return domain.Mission{}, err
	}
	experimentID := service.ids.NewID("exp")
	event, err := service.event(
		command.MissionID,
		command.RunID,
		command.TraceID,
		domain.EventExperimentRecorded,
		domain.ExperimentRecordedPayload{
			ExperimentID: experimentID,
			Stage:        command.Stage,
			Goal:         command.Goal,
			Method:       command.Method,
			Status:       defaultString(command.Status, "planned"),
			EvidenceIDs:  command.EvidenceIDs,
		},
	)
	if err != nil {
		return domain.Mission{}, err
	}
	return service.appendAndReplay(ctx, command.MissionID, []domain.DomainEvent{event})
}

func (service *Service) RecordDecision(
	ctx context.Context,
	command RecordDecisionCommand,
) (domain.Mission, error) {
	if err := service.ensureStageRunning(ctx, command.MissionID, command.Stage); err != nil {
		return domain.Mission{}, err
	}
	decisionID := service.ids.NewID("dec")
	decision := domain.Decision{
		ID:             decisionID,
		MissionID:      command.MissionID,
		Stage:          command.Stage,
		Type:           command.Type,
		Confidence:     command.Confidence,
		Reason:         command.Reason,
		EvidenceRefs:   command.EvidenceRefs,
		Risks:          command.Risks,
		NextBestAction: command.NextBestAction,
		CreatedAt:      service.clock.Now(),
	}
	if err := decision.Validate(); err != nil {
		return domain.Mission{}, err
	}

	event, err := service.event(
		command.MissionID,
		command.RunID,
		command.TraceID,
		domain.EventDecisionRecorded,
		domain.DecisionRecordedPayload{
			DecisionID:     decisionID,
			Stage:          command.Stage,
			DecisionType:   command.Type,
			Confidence:     command.Confidence,
			Reason:         command.Reason,
			EvidenceRefs:   command.EvidenceRefs,
			Risks:          command.Risks,
			NextBestAction: command.NextBestAction,
		},
	)
	if err != nil {
		return domain.Mission{}, err
	}
	return service.appendAndReplay(ctx, command.MissionID, []domain.DomainEvent{event})
}

func (service *Service) CompleteStage(
	ctx context.Context,
	command CompleteStageCommand,
) (domain.Mission, error) {
	mission, err := service.loadMission(ctx, command.MissionID)
	if err != nil {
		return domain.Mission{}, err
	}
	if _, ok := mission.Decisions[command.DecisionID]; !ok {
		return domain.Mission{}, fmt.Errorf("decision %s does not exist", command.DecisionID)
	}
	stage, ok := mission.Stages[command.Stage]
	if !ok || stage.Status != domain.StageStatusRunning {
		return domain.Mission{}, fmt.Errorf("stage %s is not running", command.Stage)
	}

	event, err := service.event(
		command.MissionID,
		command.RunID,
		command.TraceID,
		domain.EventStageCompleted,
		domain.StageCompletedPayload{
			StageID:    stage.ID,
			Stage:      command.Stage,
			DecisionID: command.DecisionID,
		},
	)
	if err != nil {
		return domain.Mission{}, err
	}
	return service.appendAndReplay(ctx, command.MissionID, []domain.DomainEvent{event})
}

func (service *Service) CreatePlaceholderStage(
	ctx context.Context,
	command CreatePlaceholderStageCommand,
) (domain.Mission, error) {
	if _, err := service.loadMission(ctx, command.MissionID); err != nil {
		return domain.Mission{}, err
	}
	if command.Stage.IsMVPAutomatic() {
		return domain.Mission{}, fmt.Errorf("stage %s must run instead of placeholder", command.Stage)
	}
	if !command.Stage.IsValid() {
		return domain.Mission{}, fmt.Errorf("invalid stage %s", command.Stage)
	}
	stageID := service.ids.NewID("stg")
	event, err := service.event(
		command.MissionID,
		command.RunID,
		command.TraceID,
		domain.EventStagePlaceholderMade,
		domain.StagePlaceholderCreatedPayload{
			StageID:    stageID,
			Stage:      command.Stage,
			NextAction: command.NextAction,
		},
	)
	if err != nil {
		return domain.Mission{}, err
	}
	return service.appendAndReplay(ctx, command.MissionID, []domain.DomainEvent{event})
}

func (service *Service) ensureStageRunning(
	ctx context.Context,
	missionID string,
	stageName domain.StageName,
) error {
	mission, err := service.loadMission(ctx, missionID)
	if err != nil {
		return err
	}
	stage, ok := mission.Stages[stageName]
	if !ok || stage.Status != domain.StageStatusRunning {
		return fmt.Errorf("stage %s is not running", stageName)
	}
	return nil
}

func validateAutomaticStageStart(mission domain.Mission, stageName domain.StageName) error {
	if !stageName.IsMVPAutomatic() {
		return fmt.Errorf("stage %s cannot auto-run in MVP", stageName)
	}
	if _, exists := mission.Stages[stageName]; exists {
		return fmt.Errorf("stage %s already exists", stageName)
	}
	previousStage, ok := previousStage(stageName)
	if !ok {
		return nil
	}
	if mission.Stages[previousStage].Status != domain.StageStatusCompleted {
		return fmt.Errorf("previous stage %s must be completed", previousStage)
	}
	return nil
}

func previousStage(stageName domain.StageName) (domain.StageName, bool) {
	for index, stage := range domain.OrderedStages {
		if stage == stageName && index > 0 {
			return domain.OrderedStages[index-1], true
		}
	}
	return "", false
}

func (service *Service) appendAndReplay(
	ctx context.Context,
	missionID string,
	events []domain.DomainEvent,
) (domain.Mission, error) {
	if err := service.store.Append(ctx, events); err != nil {
		return domain.Mission{}, err
	}
	return service.loadMission(ctx, missionID)
}

func (service *Service) loadMission(ctx context.Context, missionID string) (domain.Mission, error) {
	events, err := service.store.LoadMission(ctx, missionID)
	if err != nil {
		return domain.Mission{}, err
	}
	return domain.ReplayMission(events)
}

func (service *Service) events(
	missionID string,
	runID string,
	traceID string,
	eventTypeAndPayload ...any,
) ([]domain.DomainEvent, error) {
	if len(eventTypeAndPayload)%2 != 0 {
		return nil, errors.New("event type and payload pairs are required")
	}
	events := make([]domain.DomainEvent, 0, len(eventTypeAndPayload)/2)
	for index := 0; index < len(eventTypeAndPayload); index += 2 {
		eventType, ok := eventTypeAndPayload[index].(string)
		if !ok {
			return nil, errors.New("event type must be string")
		}
		event, err := service.event(missionID, runID, traceID, eventType, eventTypeAndPayload[index+1])
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}

func (service *Service) event(
	missionID string,
	runID string,
	traceID string,
	eventType string,
	payload any,
) (domain.DomainEvent, error) {
	encodedPayload, err := domain.EncodePayload(payload)
	if err != nil {
		return domain.DomainEvent{}, err
	}
	event := domain.DomainEvent{
		EventID:       service.ids.NewID("evt"),
		SchemaVersion: domain.EventSchemaVersion,
		TraceID:       traceID,
		MissionID:     missionID,
		RunID:         runID,
		Actor:         service.actor,
		Timestamp:     service.clock.Now(),
		Type:          eventType,
		Payload:       encodedPayload,
	}
	if err := event.Validate(); err != nil {
		return domain.DomainEvent{}, err
	}
	return event, nil
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
