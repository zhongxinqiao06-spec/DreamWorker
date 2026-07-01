package decisiongate

import (
	"testing"
	"time"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/domain"
)

func TestEvaluateAsksUserWhenEvidenceIsMissing(t *testing.T) {
	decision := evaluate(t, Input{
		DecisionID: "dec_001",
		MissionID:  "msn_001",
		Stage:      domain.StageDiscover,
	})

	assertDecision(t, decision, domain.DecisionAskUser)
	if len(decision.EvidenceRefs) != 0 {
		t.Fatalf("expected no evidence refs, got %#v", decision.EvidenceRefs)
	}
}

func TestEvaluatePausesLowConfidenceHighRiskStage(t *testing.T) {
	decision := evaluate(t, Input{
		DecisionID: "dec_001",
		MissionID:  "msn_001",
		Stage:      domain.StageValidate,
		Evidence: []domain.Evidence{
			evidence("ev_001", 0.45, domain.RiskHigh),
		},
	})

	assertDecision(t, decision, domain.DecisionPause)
}

func TestEvaluateAsksUserForConflictingEvidence(t *testing.T) {
	decision := evaluate(t, Input{
		DecisionID: "dec_001",
		MissionID:  "msn_001",
		Stage:      domain.StageValidate,
		Evidence: []domain.Evidence{
			evidence("ev_001", 0.9, domain.RiskLow),
			evidence("ev_002", 0.85, domain.RiskHigh),
		},
	})

	assertDecision(t, decision, domain.DecisionAskUser)
	if decision.Risks[0] != "conflicting_evidence" {
		t.Fatalf("expected conflicting evidence risk, got %#v", decision.Risks)
	}
}

func TestEvaluateContinuesWithEnoughEvidence(t *testing.T) {
	decision := evaluate(t, Input{
		DecisionID: "dec_001",
		MissionID:  "msn_001",
		Stage:      domain.StageShape,
		Evidence: []domain.Evidence{
			evidence("ev_001", 0.82, domain.RiskMedium),
		},
	})

	assertDecision(t, decision, domain.DecisionContinue)
}

func TestEvaluateHonorsUserPivotAndStop(t *testing.T) {
	pivot := evaluate(t, Input{
		DecisionID:        "dec_001",
		MissionID:         "msn_001",
		Stage:             domain.StageValidate,
		Evidence:          []domain.Evidence{evidence("ev_001", 0.8, domain.RiskMedium)},
		UserSteering:      UserSteeringPivot,
		NewHypothesisSeed: "转向项目孵化器操作系统",
	})
	assertDecision(t, pivot, domain.DecisionPivot)

	kill := evaluate(t, Input{
		DecisionID:   "dec_002",
		MissionID:    "msn_001",
		Stage:        domain.StageValidate,
		Evidence:     []domain.Evidence{evidence("ev_001", 0.8, domain.RiskMedium)},
		UserSteering: UserSteeringStop,
	})
	assertDecision(t, kill, domain.DecisionKill)
}

func evaluate(t *testing.T, input Input) domain.Decision {
	t.Helper()
	input.Now = time.Date(2026, 6, 30, 8, 0, 0, 0, time.UTC)
	decision, err := Evaluate(input)
	if err != nil {
		t.Fatalf("evaluate decision: %v", err)
	}
	return decision
}

func assertDecision(t *testing.T, decision domain.Decision, expected domain.DecisionType) {
	t.Helper()
	if decision.Type != expected {
		t.Fatalf("expected %s decision, got %s", expected, decision.Type)
	}
	if err := decision.Validate(); err != nil {
		t.Fatalf("expected valid decision: %v", err)
	}
	if decision.NextBestAction == "" {
		t.Fatal("expected next_best_action")
	}
}

func evidence(id string, confidence float64, risk domain.RiskLevel) domain.Evidence {
	return domain.Evidence{
		ID:             id,
		MissionID:      "msn_001",
		Stage:          domain.StageValidate,
		Source:         "fixture",
		Summary:        "证据摘要",
		Confidence:     confidence,
		Risk:           risk,
		NextBestAction: "继续验证。",
	}
}
