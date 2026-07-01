package domain

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestAgentSpecAndTaskGraphValidate(t *testing.T) {
	spec := testAgentSpec()
	if err := spec.Validate(); err != nil {
		t.Fatalf("validate spec: %v", err)
	}

	graph := TaskGraph{
		SchemaVersion: ContractSchemaVersion,
		MissionID:     "msn_001",
		RunID:         "run_001",
		TraceID:       "tr_001",
		Idea:          "idea",
		Tasks: []Task{{
			SchemaVersion:        ContractSchemaVersion,
			TaskID:               "tsk_001",
			Stage:                StageDiscover,
			Goal:                 "generate artifact",
			AssignedAgent:        spec.ID,
			RequiredCapabilities: []string{CapabilityIDModelGenerateStub, CapabilityIDArtifactWrite},
			ExpectedArtifacts:    []string{"dream_brief.md"},
			Budget:               Budget{MaxTokens: 1000, MaxCostUSD: 0},
			Status:               TaskStatusPending,
			TraceID:              "tr_001",
		}},
	}
	if err := graph.Validate(map[string]AgentSpec{spec.ID: spec}); err != nil {
		t.Fatalf("validate graph: %v", err)
	}
}

func TestTaskGraphRejectsForbiddenCapabilityAndCycle(t *testing.T) {
	spec := testAgentSpec()
	graph := TaskGraph{
		SchemaVersion: ContractSchemaVersion,
		MissionID:     "msn_001",
		RunID:         "run_001",
		TraceID:       "tr_001",
		Idea:          "idea",
		Tasks: []Task{
			testTask("tsk_a", spec.ID, []string{"tsk_b"}, []string{CapabilityIDCodeExecutionLike()}),
			testTask("tsk_b", spec.ID, []string{"tsk_a"}, []string{CapabilityIDModelGenerateStub}),
		},
	}

	err := graph.Validate(map[string]AgentSpec{spec.ID: spec})
	if !errors.Is(err, ErrTaskCapabilityNotAllowed) {
		t.Fatalf("expected forbidden capability error, got %v", err)
	}

	graph.Tasks[0].RequiredCapabilities = []string{CapabilityIDModelGenerateStub}
	err = graph.Validate(map[string]AgentSpec{spec.ID: spec})
	if !errors.Is(err, ErrTaskGraphInvalid) {
		t.Fatalf("expected task graph cycle error, got %v", err)
	}
}

func TestNormalizeModelOutputRequiresArtifactsAndNextAction(t *testing.T) {
	valid, err := json.Marshal(NormalizedOutput{
		Artifacts: []ArtifactDraft{{
			FileName:    "dream_brief.md",
			ArtifactID:  "art_dream_brief",
			Kind:        "dream_brief",
			Title:       "Dream Brief",
			ContentType: "text/markdown",
			Content:     "content",
		}},
		NextBestAction: "continue",
	})
	if err != nil {
		t.Fatalf("marshal valid output: %v", err)
	}
	if _, err := NormalizeModelOutput(valid); err != nil {
		t.Fatalf("normalize valid output: %v", err)
	}

	_, err = NormalizeModelOutput(json.RawMessage(`{"artifacts":[],"next_best_action":"continue"}`))
	if !errors.Is(err, ErrNormalizationFailed) {
		t.Fatalf("expected normalization failure, got %v", err)
	}
}

func testAgentSpec() AgentSpec {
	return AgentSpec{
		SchemaVersion:       ContractSchemaVersion,
		ID:                  "product_analyst",
		Role:                "Product analyst",
		InputSchema:         map[string]any{"type": "object"},
		OutputSchema:        map[string]any{"type": "object"},
		AllowedCapabilities: []string{CapabilityIDModelGenerateStub, CapabilityIDArtifactWrite},
		DefaultModelProfile: "stub_reasoning_light",
		Budget:              Budget{MaxTokens: 1000, MaxCostUSD: 0},
		Timeout:             "10s",
		ApprovalPolicy:      ApprovalPolicyOnRisk,
		ExpectedArtifacts:   []string{"dream_brief.md"},
		PromptRef: PromptRef{
			PromptID:      "prm_product_analyst",
			PromptVersion: "v1",
			AgentID:       "product_analyst",
		},
	}
}

func testTask(taskID string, agentID string, dependsOn []string, capabilities []string) Task {
	return Task{
		SchemaVersion:        ContractSchemaVersion,
		TaskID:               taskID,
		Stage:                StageDiscover,
		Goal:                 "goal",
		AssignedAgent:        agentID,
		RequiredCapabilities: capabilities,
		ExpectedArtifacts:    []string{"dream_brief.md"},
		DependsOn:            dependsOn,
		Budget:               Budget{MaxTokens: 1000, MaxCostUSD: 0},
		Status:               TaskStatusPending,
		TraceID:              "tr_001",
	}
}

func CapabilityIDCodeExecutionLike() string {
	return "cap_code_execution"
}
