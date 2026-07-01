package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

const (
	EventMissionCreated       = "mission.created"
	EventStageStarted         = "stage.started"
	EventHypothesisRecorded   = "hypothesis.recorded"
	EventEvidenceRecorded     = "evidence.recorded"
	EventExperimentRecorded   = "experiment.recorded"
	EventDecisionRecorded     = "decision.recorded"
	EventStageCompleted       = "stage.completed"
	EventStagePlaceholderMade = "stage.placeholder_created"
)

type MissionCreatedPayload struct {
	MissionID string `json:"mission_id"`
	RunID     string `json:"run_id"`
	Title     string `json:"title"`
	Idea      string `json:"idea"`
}

type StageStartedPayload struct {
	StageID string    `json:"stage_id"`
	Stage   StageName `json:"stage"`
}

type HypothesisRecordedPayload struct {
	HypothesisID string    `json:"hypothesis_id"`
	Stage        StageName `json:"stage"`
	Statement    string    `json:"statement"`
	OwnerAgent   string    `json:"owner_agent"`
	Status       string    `json:"status"`
}

type EvidenceRecordedPayload struct {
	EvidenceID     string            `json:"evidence_id"`
	Stage          StageName         `json:"stage"`
	Source         string            `json:"source"`
	Summary        string            `json:"summary"`
	Confidence     float64           `json:"confidence"`
	Risk           RiskLevel         `json:"risk"`
	NextBestAction string            `json:"next_best_action"`
	Bindings       []EvidenceBinding `json:"bindings"`
}

type ExperimentRecordedPayload struct {
	ExperimentID string    `json:"experiment_id"`
	Stage        StageName `json:"stage"`
	Goal         string    `json:"goal"`
	Method       string    `json:"method"`
	Status       string    `json:"status"`
	EvidenceIDs  []string  `json:"evidence_ids"`
}

type DecisionRecordedPayload struct {
	DecisionID     string       `json:"decision_id"`
	Stage          StageName    `json:"stage"`
	DecisionType   DecisionType `json:"decision_type"`
	Confidence     float64      `json:"confidence"`
	Reason         string       `json:"reason"`
	EvidenceRefs   []string     `json:"evidence_refs"`
	Risks          []string     `json:"risks"`
	NextBestAction string       `json:"next_best_action"`
}

type StageCompletedPayload struct {
	StageID    string    `json:"stage_id"`
	Stage      StageName `json:"stage"`
	DecisionID string    `json:"decision_id"`
}

type StagePlaceholderCreatedPayload struct {
	StageID    string    `json:"stage_id"`
	Stage      StageName `json:"stage"`
	NextAction string    `json:"next_action"`
}

func EncodePayload(payload any) (json.RawMessage, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode event payload: %w", err)
	}
	return json.RawMessage(data), nil
}

func ReplayMission(events []DomainEvent) (Mission, error) {
	if len(events) == 0 {
		return Mission{}, errors.New("cannot replay mission without events")
	}

	var mission Mission
	for _, event := range events {
		if err := event.Validate(); err != nil {
			return Mission{}, fmt.Errorf("invalid event %s: %w", event.EventID, err)
		}
		if event.Type != EventMissionCreated && mission.ID == "" {
			return Mission{}, fmt.Errorf("first mission event must be %s", EventMissionCreated)
		}
		if err := applyMissionEvent(&mission, event); err != nil {
			return Mission{}, err
		}
	}
	return mission, nil
}

func applyMissionEvent(mission *Mission, event DomainEvent) error {
	switch event.Type {
	case EventMissionCreated:
		var payload MissionCreatedPayload
		if err := decodePayload(event, &payload); err != nil {
			return err
		}
		*mission = NewEmptyMission(payload.MissionID)
		mission.Title = payload.Title
		mission.Idea = payload.Idea
		mission.CreatedAt = event.Timestamp
		mission.UpdatedAt = event.Timestamp
		mission.Runs[payload.RunID] = Run{
			ID:        payload.RunID,
			MissionID: payload.MissionID,
			Status:    "running",
			TraceID:   event.TraceID,
			CreatedAt: event.Timestamp,
		}
	case EventStageStarted:
		var payload StageStartedPayload
		if err := decodePayload(event, &payload); err != nil {
			return err
		}
		if !payload.Stage.IsValid() {
			return fmt.Errorf("invalid stage: %s", payload.Stage)
		}
		stage := Stage{
			ID:        payload.StageID,
			MissionID: event.MissionID,
			Name:      payload.Stage,
			Status:    StageStatusRunning,
			StartedAt: event.Timestamp,
		}
		mission.Stages[payload.Stage] = stage
		mission.CurrentStage = payload.Stage
		mission.UpdatedAt = event.Timestamp
	case EventHypothesisRecorded:
		var payload HypothesisRecordedPayload
		if err := decodePayload(event, &payload); err != nil {
			return err
		}
		mission.Hypotheses[payload.HypothesisID] = Hypothesis{
			ID:         payload.HypothesisID,
			MissionID:  event.MissionID,
			Stage:      payload.Stage,
			Statement:  payload.Statement,
			OwnerAgent: payload.OwnerAgent,
			Status:     payload.Status,
			CreatedAt:  event.Timestamp,
		}
		mission.UpdatedAt = event.Timestamp
	case EventEvidenceRecorded:
		var payload EvidenceRecordedPayload
		if err := decodePayload(event, &payload); err != nil {
			return err
		}
		if err := ValidateEvidenceBindings(payload.EvidenceID, payload.Bindings); err != nil {
			return fmt.Errorf("invalid evidence bindings: %w", err)
		}
		evidence := Evidence{
			ID:             payload.EvidenceID,
			MissionID:      event.MissionID,
			Stage:          payload.Stage,
			Source:         payload.Source,
			Summary:        payload.Summary,
			Confidence:     payload.Confidence,
			Risk:           payload.Risk,
			NextBestAction: payload.NextBestAction,
			Bindings:       payload.Bindings,
			CreatedAt:      event.Timestamp,
		}
		mission.Evidence[payload.EvidenceID] = evidence
		for _, binding := range payload.Bindings {
			mission.EvidenceGraph.Add(binding)
			if binding.TargetKind == EvidenceTargetHypothesis {
				hypothesis := mission.Hypotheses[binding.TargetID]
				hypothesis.EvidenceRefs = appendUnique(hypothesis.EvidenceRefs, payload.EvidenceID)
				mission.Hypotheses[binding.TargetID] = hypothesis
			}
		}
		mission.UpdatedAt = event.Timestamp
	case EventExperimentRecorded:
		var payload ExperimentRecordedPayload
		if err := decodePayload(event, &payload); err != nil {
			return err
		}
		mission.Experiments[payload.ExperimentID] = Experiment{
			ID:          payload.ExperimentID,
			MissionID:   event.MissionID,
			Stage:       payload.Stage,
			Goal:        payload.Goal,
			Method:      payload.Method,
			Status:      payload.Status,
			EvidenceIDs: payload.EvidenceIDs,
			CreatedAt:   event.Timestamp,
		}
		mission.UpdatedAt = event.Timestamp
	case EventDecisionRecorded:
		var payload DecisionRecordedPayload
		if err := decodePayload(event, &payload); err != nil {
			return err
		}
		decision := Decision{
			ID:             payload.DecisionID,
			MissionID:      event.MissionID,
			Stage:          payload.Stage,
			Type:           payload.DecisionType,
			Confidence:     payload.Confidence,
			Reason:         payload.Reason,
			EvidenceRefs:   payload.EvidenceRefs,
			Risks:          payload.Risks,
			NextBestAction: payload.NextBestAction,
			CreatedAt:      event.Timestamp,
		}
		if err := decision.Validate(); err != nil {
			return fmt.Errorf("invalid decision: %w", err)
		}
		mission.Decisions[payload.DecisionID] = decision
		mission.EvidenceGraph.ByDecision[payload.DecisionID] = append([]string{}, payload.EvidenceRefs...)
		mission.UpdatedAt = event.Timestamp
	case EventStageCompleted:
		var payload StageCompletedPayload
		if err := decodePayload(event, &payload); err != nil {
			return err
		}
		stage := mission.Stages[payload.Stage]
		stage.Status = StageStatusCompleted
		stage.CompletedAt = event.Timestamp
		mission.Stages[payload.Stage] = stage
		mission.UpdatedAt = event.Timestamp
	case EventStagePlaceholderMade:
		var payload StagePlaceholderCreatedPayload
		if err := decodePayload(event, &payload); err != nil {
			return err
		}
		stage := Stage{
			ID:          payload.StageID,
			MissionID:   event.MissionID,
			Name:        payload.Stage,
			Status:      StageStatusPending,
			Placeholder: true,
			NextAction:  payload.NextAction,
			StartedAt:   event.Timestamp,
		}
		mission.Stages[payload.Stage] = stage
		mission.UpdatedAt = event.Timestamp
	default:
		return fmt.Errorf("unsupported mission event type %s", event.Type)
	}
	return nil
}

func decodePayload(event DomainEvent, target any) error {
	if err := json.Unmarshal(event.Payload, target); err != nil {
		return fmt.Errorf("decode %s payload: %w", event.Type, err)
	}
	return nil
}

func appendUnique(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}
