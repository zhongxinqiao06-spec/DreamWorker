package modelstub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/domain"
	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/ports"
)

var _ ports.ModelGateway = (*Gateway)(nil)

type Gateway struct{}

func NewGateway() *Gateway {
	return &Gateway{}
}

func (gateway *Gateway) Generate(
	ctx context.Context,
	request domain.ModelRequest,
) (domain.ModelResponse, error) {
	select {
	case <-ctx.Done():
		return domain.ModelResponse{}, ctx.Err()
	default:
	}

	artifacts := make([]domain.ArtifactDraft, 0, len(request.ExpectedArtifacts))
	for _, fileName := range request.ExpectedArtifacts {
		artifacts = append(artifacts, domain.ArtifactDraft{
			FileName:    fileName,
			ArtifactID:  artifactID(fileName),
			Kind:        artifactKind(fileName),
			Title:       artifactTitle(fileName),
			ContentType: "text/markdown; charset=utf-8",
			Content:     artifactContent(request, fileName),
		})
	}
	output := domain.NormalizedOutput{
		Artifacts:      artifacts,
		EvidenceRefs:   []string{},
		NextBestAction: nextBestAction(request.Stage),
	}
	structured, err := json.Marshal(output)
	if err != nil {
		return domain.ModelResponse{}, fmt.Errorf("marshal stub output: %w", err)
	}

	usage := domain.ModelUsage{
		InputTokens:  estimateTokens(request.Idea + " " + request.Goal),
		OutputTokens: estimateTokens(string(structured)),
		CostUSD:      0,
	}
	if request.Budget.MaxTokens > 0 && usage.InputTokens+usage.OutputTokens > request.Budget.MaxTokens {
		return domain.ModelResponse{}, domain.ErrModelBudgetExceeded
	}
	if request.Budget.MaxCostUSD >= 0 && usage.CostUSD > request.Budget.MaxCostUSD {
		return domain.ModelResponse{}, domain.ErrModelBudgetExceeded
	}

	return domain.ModelResponse{
		SchemaVersion:    domain.ContractSchemaVersion,
		ResponseID:       strings.Replace(request.RequestID, "mdlreq_", "mdlres_", 1),
		RequestID:        request.RequestID,
		TraceID:          request.TraceID,
		Provider:         "stub",
		Model:            "deterministic-agent-runtime-v0.1",
		RawOutput:        string(structured),
		StructuredOutput: structured,
		Usage:            usage,
		FinishReason:     "stop",
	}, nil
}

func (gateway *Gateway) Stream(
	ctx context.Context,
	request domain.ModelRequest,
) (<-chan domain.ModelStreamEvent, error) {
	channel := make(chan domain.ModelStreamEvent, 1)
	response, err := gateway.Generate(ctx, request)
	if err != nil {
		channel <- domain.ModelStreamEvent{
			SchemaVersion: domain.ContractSchemaVersion,
			RequestID:     request.RequestID,
			TraceID:       request.TraceID,
			Done:          true,
			Error: &domain.ModelCallError{
				Code:        modelErrorCode(err),
				Message:     err.Error(),
				Recoverable: errors.Is(err, domain.ErrModelBudgetExceeded),
			},
		}
		close(channel)
		return channel, nil
	}
	channel <- domain.ModelStreamEvent{
		SchemaVersion: domain.ContractSchemaVersion,
		RequestID:     request.RequestID,
		TraceID:       request.TraceID,
		Delta:         response.RawOutput,
		Done:          true,
		Response:      &response,
	}
	close(channel)
	return channel, nil
}

func modelErrorCode(err error) string {
	if errors.Is(err, domain.ErrModelBudgetExceeded) {
		return "MODEL_BUDGET_EXCEEDED"
	}
	return "MODEL_GENERATE_FAILED"
}

func artifactID(fileName string) string {
	name := strings.TrimSuffix(fileName, ".md")
	name = strings.TrimSuffix(name, ".yaml")
	name = strings.ReplaceAll(name, "-", "_")
	return "art_" + name
}

func artifactKind(fileName string) string {
	switch fileName {
	case "dream_brief.md":
		return "dream_brief"
	case "research_pack.md":
		return "research_pack"
	case "blueprint.yaml":
		return "blueprint"
	case "eval_report.yaml":
		return "eval_report"
	default:
		return "other"
	}
}

func artifactTitle(fileName string) string {
	switch fileName {
	case "dream_brief.md":
		return "Dream Brief"
	case "hypotheses.yaml":
		return "Hypotheses"
	case "research_pack.md":
		return "Research Pack"
	case "evidence_graph.yaml":
		return "Evidence Graph"
	case "mvp_scope.md":
		return "MVP Scope"
	case "blueprint.yaml":
		return "Blueprint"
	case "eval_report.yaml":
		return "Eval Report"
	default:
		return fileName
	}
}

func artifactContent(request domain.ModelRequest, fileName string) string {
	return fmt.Sprintf(
		"# %s\n\n- mission: %s\n- stage: %s\n- agent: %s\n- goal: %s\n- prompt: %s@%s\n- next_best_action: %s\n",
		artifactTitle(fileName),
		request.Idea,
		request.Stage,
		request.AgentID,
		request.Goal,
		request.PromptRef.PromptID,
		request.PromptRef.PromptVersion,
		nextBestAction(request.Stage),
	)
}

func nextBestAction(stage domain.StageName) string {
	switch stage {
	case domain.StageDiscover:
		return "进入 Validate，补充竞品和用户证据。"
	case domain.StageValidate:
		return "进入 Shape，收敛 MVP 范围和技术蓝图。"
	case domain.StageShape:
		return "准备 PRD、发布清单和实现计划。"
	default:
		return "等待用户确认下一步。"
	}
}

func estimateTokens(text string) int {
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return 1
	}
	return len(fields)
}
