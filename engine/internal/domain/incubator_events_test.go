package domain

import (
	"encoding/json"
	"testing"
	"time"
)

func TestReplayMissionRebuildsIncubatorState(t *testing.T) {
	events := []DomainEvent{
		incubatorEvent(t, "evt_001", EventMissionCreated, MissionCreatedPayload{
			MissionID: "msn_001",
			RunID:     "run_001",
			Title:     "AI 项目孵化器",
			Idea:      "我想做一个面向独立开发者的 AI 项目孵化工具。",
		}),
		incubatorEvent(t, "evt_002", EventStageStarted, StageStartedPayload{
			StageID: "stg_001",
			Stage:   StageDiscover,
		}),
		incubatorEvent(t, "evt_003", EventHypothesisRecorded, HypothesisRecordedPayload{
			HypothesisID: "hyp_001",
			Stage:        StageDiscover,
			Statement:    "独立开发者需要更结构化的 AI 项目孵化流程。",
			OwnerAgent:   "product_analyst",
			Status:       "testing",
		}),
		incubatorEvent(t, "evt_004", EventEvidenceRecorded, EvidenceRecordedPayload{
			EvidenceID:     "ev_001",
			Stage:          StageDiscover,
			Source:         "seed",
			Summary:        "用户明确希望从想法生成可执行蓝图。",
			Confidence:     0.82,
			Risk:           RiskMedium,
			NextBestAction: "继续验证付费意愿。",
			Bindings: []EvidenceBinding{{
				EvidenceID: "ev_001",
				TargetKind: EvidenceTargetHypothesis,
				TargetID:   "hyp_001",
			}},
		}),
		incubatorEvent(t, "evt_005", EventDecisionRecorded, DecisionRecordedPayload{
			DecisionID:     "dec_001",
			Stage:          StageDiscover,
			DecisionType:   DecisionContinue,
			Confidence:     0.82,
			Reason:         "已有足够证据进入 Validate。",
			EvidenceRefs:   []string{"ev_001"},
			Risks:          []string{"market_uncertainty"},
			NextBestAction: "进入 Validate 并补充竞品证据。",
		}),
		incubatorEvent(t, "evt_006", EventStageCompleted, StageCompletedPayload{
			StageID:    "stg_001",
			Stage:      StageDiscover,
			DecisionID: "dec_001",
		}),
	}

	mission, err := ReplayMission(events)
	if err != nil {
		t.Fatalf("replay mission: %v", err)
	}

	if mission.ID != "msn_001" {
		t.Fatalf("expected mission id msn_001, got %q", mission.ID)
	}
	if mission.CurrentStage != StageDiscover {
		t.Fatalf("expected current stage Discover, got %q", mission.CurrentStage)
	}
	if mission.Stages[StageDiscover].Status != StageStatusCompleted {
		t.Fatalf("expected Discover completed, got %q", mission.Stages[StageDiscover].Status)
	}
	if len(mission.Hypotheses["hyp_001"].EvidenceRefs) != 1 {
		t.Fatalf("expected hypothesis evidence ref, got %#v", mission.Hypotheses["hyp_001"].EvidenceRefs)
	}
	if mission.EvidenceGraph.ByHypothesis["hyp_001"][0] != "ev_001" {
		t.Fatalf("expected evidence graph binding, got %#v", mission.EvidenceGraph.ByHypothesis)
	}
	if mission.Decisions["dec_001"].Type != DecisionContinue {
		t.Fatalf("expected continue decision, got %q", mission.Decisions["dec_001"].Type)
	}
}

func TestReplayMissionRejectsEvidenceWithoutBinding(t *testing.T) {
	events := []DomainEvent{
		incubatorEvent(t, "evt_001", EventMissionCreated, MissionCreatedPayload{
			MissionID: "msn_001",
			RunID:     "run_001",
			Title:     "AI 项目孵化器",
			Idea:      "seed",
		}),
		incubatorEvent(t, "evt_002", EventEvidenceRecorded, EvidenceRecordedPayload{
			EvidenceID:     "ev_001",
			Stage:          StageDiscover,
			Source:         "seed",
			Summary:        "缺少绑定。",
			Confidence:     0.5,
			Risk:           RiskMedium,
			NextBestAction: "补充绑定。",
		}),
	}

	if _, err := ReplayMission(events); err == nil {
		t.Fatal("expected evidence binding error")
	}
}

func TestBuildLaunchLearnAreNotMVPAutomaticStages(t *testing.T) {
	for _, stage := range []StageName{StageBuild, StageLaunch, StageLearn} {
		if stage.IsMVPAutomatic() {
			t.Fatalf("expected %s to be placeholder-only in MVP", stage)
		}
	}
}

func incubatorEvent(t *testing.T, eventID string, eventType string, payload any) DomainEvent {
	t.Helper()
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	event := DomainEvent{
		EventID:       eventID,
		SchemaVersion: EventSchemaVersion,
		TraceID:       "tr_001",
		MissionID:     "msn_001",
		RunID:         "run_001",
		Actor:         "test",
		Timestamp:     time.Date(2026, 6, 30, 8, 0, 0, 0, time.UTC),
		Type:          eventType,
		Payload:       data,
	}
	if err := event.Validate(); err != nil {
		t.Fatalf("validate event: %v", err)
	}
	return event
}
