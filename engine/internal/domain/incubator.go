package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type StageName string

const (
	StageDiscover StageName = "Discover"
	StageValidate StageName = "Validate"
	StageShape    StageName = "Shape"
	StageBuild    StageName = "Build"
	StageLaunch   StageName = "Launch"
	StageLearn    StageName = "Learn"
)

var OrderedStages = []StageName{
	StageDiscover,
	StageValidate,
	StageShape,
	StageBuild,
	StageLaunch,
	StageLearn,
}

type StageStatus string

const (
	StageStatusPending   StageStatus = "pending"
	StageStatusRunning   StageStatus = "running"
	StageStatusCompleted StageStatus = "completed"
	StageStatusPaused    StageStatus = "paused"
	StageStatusCancelled StageStatus = "cancelled"
)

type DecisionType string

const (
	DecisionContinue DecisionType = "continue"
	DecisionPivot    DecisionType = "pivot"
	DecisionPause    DecisionType = "pause"
	DecisionKill     DecisionType = "kill"
	DecisionAskUser  DecisionType = "ask_user"
)

type RiskLevel string

const (
	RiskLow      RiskLevel = "low"
	RiskMedium   RiskLevel = "medium"
	RiskHigh     RiskLevel = "high"
	RiskCritical RiskLevel = "critical"
)

type EvidenceTargetKind string

const (
	EvidenceTargetHypothesis           EvidenceTargetKind = "Hypothesis"
	EvidenceTargetArtifact             EvidenceTargetKind = "Artifact"
	EvidenceTargetRun                  EvidenceTargetKind = "Run"
	EvidenceTargetCapabilityInvocation EvidenceTargetKind = "CapabilityInvocation"
	EvidenceTargetDecision             EvidenceTargetKind = "Decision"
)

type Mission struct {
	ID            string
	Title         string
	Idea          string
	CurrentStage  StageName
	Stages        map[StageName]Stage
	Hypotheses    map[string]Hypothesis
	Evidence      map[string]Evidence
	Experiments   map[string]Experiment
	Decisions     map[string]Decision
	Blueprints    map[string]Blueprint
	Runs          map[string]Run
	EvidenceGraph EvidenceGraph
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Stage struct {
	ID          string
	MissionID   string
	Name        StageName
	Status      StageStatus
	Placeholder bool
	NextAction  string
	StartedAt   time.Time
	CompletedAt time.Time
}

type Hypothesis struct {
	ID           string
	MissionID    string
	Stage        StageName
	Statement    string
	OwnerAgent   string
	Status       string
	EvidenceRefs []string
	CreatedAt    time.Time
}

type Evidence struct {
	ID             string
	MissionID      string
	Stage          StageName
	Source         string
	Summary        string
	Confidence     float64
	Risk           RiskLevel
	NextBestAction string
	Bindings       []EvidenceBinding
	CreatedAt      time.Time
}

type EvidenceBinding struct {
	EvidenceID string
	TargetKind EvidenceTargetKind
	TargetID   string
}

type Experiment struct {
	ID          string
	MissionID   string
	Stage       StageName
	Goal        string
	Method      string
	Status      string
	EvidenceIDs []string
	CreatedAt   time.Time
}

type Decision struct {
	ID             string
	MissionID      string
	Stage          StageName
	Type           DecisionType
	Confidence     float64
	Reason         string
	EvidenceRefs   []string
	Risks          []string
	NextBestAction string
	CreatedAt      time.Time
}

type Blueprint struct {
	ID        string
	MissionID string
	Title     string
	CreatedAt time.Time
}

type Run struct {
	ID        string
	MissionID string
	Status    string
	TraceID   string
	CreatedAt time.Time
}

type EvidenceGraph struct {
	ByEvidence   map[string][]EvidenceBinding
	ByHypothesis map[string][]string
	ByDecision   map[string][]string
}

func NewEmptyMission(id string) Mission {
	return Mission{
		ID:            id,
		Stages:        map[StageName]Stage{},
		Hypotheses:    map[string]Hypothesis{},
		Evidence:      map[string]Evidence{},
		Experiments:   map[string]Experiment{},
		Decisions:     map[string]Decision{},
		Blueprints:    map[string]Blueprint{},
		Runs:          map[string]Run{},
		EvidenceGraph: NewEvidenceGraph(),
	}
}

func NewEvidenceGraph() EvidenceGraph {
	return EvidenceGraph{
		ByEvidence:   map[string][]EvidenceBinding{},
		ByHypothesis: map[string][]string{},
		ByDecision:   map[string][]string{},
	}
}

func (stage StageName) IsValid() bool {
	for _, candidate := range OrderedStages {
		if stage == candidate {
			return true
		}
	}
	return false
}

func (stage StageName) IsMVPAutomatic() bool {
	return stage == StageDiscover || stage == StageValidate || stage == StageShape
}

func NextStage(stage StageName) (StageName, bool) {
	for index, candidate := range OrderedStages {
		if stage == candidate && index+1 < len(OrderedStages) {
			return OrderedStages[index+1], true
		}
	}
	return "", false
}

func ValidateEvidenceBindings(evidenceID string, bindings []EvidenceBinding) error {
	if len(bindings) == 0 {
		return errors.New("evidence must bind at least one target")
	}
	for _, binding := range bindings {
		if binding.EvidenceID != evidenceID {
			return fmt.Errorf("evidence binding id mismatch: %s", binding.EvidenceID)
		}
		if !binding.TargetKind.IsValid() {
			return fmt.Errorf("invalid evidence target kind: %s", binding.TargetKind)
		}
		if strings.TrimSpace(binding.TargetID) == "" {
			return errors.New("evidence binding target id is required")
		}
	}
	return nil
}

func (kind EvidenceTargetKind) IsValid() bool {
	switch kind {
	case EvidenceTargetHypothesis,
		EvidenceTargetArtifact,
		EvidenceTargetRun,
		EvidenceTargetCapabilityInvocation,
		EvidenceTargetDecision:
		return true
	default:
		return false
	}
}

func (decision Decision) Validate() error {
	switch {
	case strings.TrimSpace(decision.ID) == "":
		return errors.New("decision id is required")
	case !decision.Stage.IsValid():
		return errors.New("decision stage is invalid")
	case !decision.Type.IsValid():
		return errors.New("decision type is invalid")
	case decision.Confidence < 0 || decision.Confidence > 1:
		return errors.New("decision confidence must be between 0 and 1")
	case strings.TrimSpace(decision.Reason) == "":
		return errors.New("decision reason is required")
	case len(decision.EvidenceRefs) == 0 && decision.Type != DecisionAskUser:
		return errors.New("decision requires evidence refs unless ask_user")
	case strings.TrimSpace(decision.NextBestAction) == "":
		return errors.New("decision next_best_action is required")
	default:
		return nil
	}
}

func (decisionType DecisionType) IsValid() bool {
	switch decisionType {
	case DecisionContinue, DecisionPivot, DecisionPause, DecisionKill, DecisionAskUser:
		return true
	default:
		return false
	}
}

func (graph EvidenceGraph) Add(binding EvidenceBinding) {
	graph.ByEvidence[binding.EvidenceID] = append(graph.ByEvidence[binding.EvidenceID], binding)
	switch binding.TargetKind {
	case EvidenceTargetHypothesis:
		graph.ByHypothesis[binding.TargetID] = append(graph.ByHypothesis[binding.TargetID], binding.EvidenceID)
	case EvidenceTargetDecision:
		graph.ByDecision[binding.TargetID] = append(graph.ByDecision[binding.TargetID], binding.EvidenceID)
	}
}
