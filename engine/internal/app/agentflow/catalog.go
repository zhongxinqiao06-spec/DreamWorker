package agentflow

import (
	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/domain"
)

const (
	AgentOrchestrator      = "orchestrator"
	AgentProductAnalyst    = "product_analyst"
	AgentCompetitorAnalyst = "competitor_analyst"
	AgentTechArchitect     = "tech_architect"
	AgentGrowthAgent       = "growth_agent"
	AgentEvaluator         = "evaluator"
)

func BuiltinAgentSpecs() []domain.AgentSpec {
	return []domain.AgentSpec{
		agentSpec(
			AgentOrchestrator,
			"Plan and mediate a bounded project incubation task graph.",
			[]string{domain.CapabilityIDModelGenerateStub, domain.CapabilityIDHumanInput},
			[]string{},
		),
		agentSpec(
			AgentProductAnalyst,
			"Analyze target users, pain points, product hypotheses and MVP scope.",
			[]string{domain.CapabilityIDModelGenerateStub, domain.CapabilityIDArtifactWrite, domain.CapabilityIDHumanInput},
			[]string{"dream_brief.md", "hypotheses.yaml", "mvp_scope.md"},
		),
		agentSpec(
			AgentCompetitorAnalyst,
			"Collect competitor and alternative evidence through approved read-only stubs.",
			[]string{
				domain.CapabilityIDModelGenerateStub,
				domain.CapabilityIDWebSearchStub,
				domain.CapabilityIDBrowserReadonlyStub,
				domain.CapabilityIDArtifactWrite,
			},
			[]string{"research_pack.md", "evidence_graph.yaml"},
		),
		agentSpec(
			AgentTechArchitect,
			"Shape the technical blueprint, architecture choices and buildable MVP boundary.",
			[]string{domain.CapabilityIDModelGenerateStub, domain.CapabilityIDArtifactWrite},
			[]string{"blueprint.yaml"},
		),
		agentSpec(
			AgentGrowthAgent,
			"Prepare launch positioning, channel assumptions and feedback loops.",
			[]string{domain.CapabilityIDModelGenerateStub, domain.CapabilityIDArtifactWrite},
			[]string{"launch_checklist.md"},
		),
		agentSpec(
			AgentEvaluator,
			"Score artifact completeness, evidence quality, hallucination risk and actionability.",
			[]string{domain.CapabilityIDModelGenerateStub, domain.CapabilityIDArtifactRead, domain.CapabilityIDArtifactWrite},
			[]string{"eval_report.yaml"},
		),
	}
}

func PromptCatalog() []domain.PromptSpec {
	specs := BuiltinAgentSpecs()
	prompts := make([]domain.PromptSpec, 0, len(specs))
	for _, spec := range specs {
		prompts = append(prompts, domain.PromptSpec{
			PromptRef: spec.PromptRef,
			Changelog: "v1: deterministic MVP prompt contract for PR-06.",
			Text:      "Use the assigned task goal, approved capability context and required output schema to produce normalized incubation artifacts.",
		})
	}
	return prompts
}

func MVPTaskGraph(missionID string, runID string, traceID string, idea string) domain.TaskGraph {
	return domain.TaskGraph{
		SchemaVersion: domain.ContractSchemaVersion,
		MissionID:     missionID,
		RunID:         runID,
		TraceID:       traceID,
		Idea:          idea,
		Tasks: []domain.Task{
			task(
				"tsk_discover_brief",
				domain.StageDiscover,
				"Generate Dream Brief and first hypotheses from the idea.",
				AgentProductAnalyst,
				[]string{domain.CapabilityIDModelGenerateStub, domain.CapabilityIDArtifactWrite},
				[]string{"dream_brief.md", "hypotheses.yaml"},
				nil,
				traceID,
			),
			task(
				"tsk_validate_research",
				domain.StageValidate,
				"Generate Research Pack and Evidence Graph from approved read-only stubs.",
				AgentCompetitorAnalyst,
				[]string{
					domain.CapabilityIDWebSearchStub,
					domain.CapabilityIDBrowserReadonlyStub,
					domain.CapabilityIDModelGenerateStub,
					domain.CapabilityIDArtifactWrite,
				},
				[]string{"research_pack.md", "evidence_graph.yaml"},
				[]string{"tsk_discover_brief"},
				traceID,
			),
			task(
				"tsk_shape_blueprint",
				domain.StageShape,
				"Generate MVP Scope and Blueprint from validated evidence.",
				AgentTechArchitect,
				[]string{domain.CapabilityIDModelGenerateStub, domain.CapabilityIDArtifactWrite},
				[]string{"mvp_scope.md", "blueprint.yaml"},
				[]string{"tsk_validate_research"},
				traceID,
			),
			task(
				"tsk_eval_report",
				domain.StageShape,
				"Generate Eval Report for artifact quality, evidence quality and next best action.",
				AgentEvaluator,
				[]string{domain.CapabilityIDModelGenerateStub, domain.CapabilityIDArtifactWrite},
				[]string{"eval_report.yaml"},
				[]string{"tsk_shape_blueprint"},
				traceID,
			),
		},
	}
}

func GoldenTasks() []GoldenTask {
	return []GoldenTask{
		{
			ID:   "golden_001",
			Idea: "我想做一个面向独立开发者的 AI 项目孵化工具。",
		},
		{
			ID:   "golden_002",
			Idea: "我想做一个帮助独立开发者生成宣发素材的 AI 工具。",
		},
		{
			ID:   "golden_003",
			Idea: "我想做一个面向小团队的本地优先研究助手。",
		},
		{
			ID:   "golden_004",
			Idea: "我想做一个 SaaS 客服团队的新员工 AI 上手助手。",
		},
		{
			ID:   "golden_005",
			Idea: "我想做一个把开发日志变成发布包和增长素材的工具。",
		},
	}
}

func agentSpec(id string, role string, capabilities []string, artifacts []string) domain.AgentSpec {
	return domain.AgentSpec{
		SchemaVersion:       domain.ContractSchemaVersion,
		ID:                  id,
		Role:                role,
		InputSchema:         map[string]any{"type": "object"},
		OutputSchema:        normalizedOutputSchema(),
		AllowedCapabilities: capabilities,
		DefaultModelProfile: "stub_reasoning_light",
		Budget: domain.Budget{
			MaxTokens:  20000,
			MaxCostUSD: 0,
		},
		Timeout:           "120s",
		ApprovalPolicy:    domain.ApprovalPolicyOnRisk,
		ExpectedArtifacts: artifacts,
		PromptRef: domain.PromptRef{
			PromptID:      "prm_" + id,
			PromptVersion: "v1",
			AgentID:       id,
		},
	}
}

func task(
	taskID string,
	stage domain.StageName,
	goal string,
	agentID string,
	capabilities []string,
	artifacts []string,
	dependsOn []string,
	traceID string,
) domain.Task {
	return domain.Task{
		SchemaVersion:        domain.ContractSchemaVersion,
		TaskID:               taskID,
		Stage:                stage,
		Goal:                 goal,
		AssignedAgent:        agentID,
		RequiredCapabilities: capabilities,
		ExpectedArtifacts:    artifacts,
		DependsOn:            append([]string{}, dependsOn...),
		Budget: domain.Budget{
			MaxTokens:  12000,
			MaxCostUSD: 0,
		},
		Status:  domain.TaskStatusPending,
		TraceID: traceID,
	}
}

func normalizedOutputSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"artifacts", "next_best_action"},
	}
}
