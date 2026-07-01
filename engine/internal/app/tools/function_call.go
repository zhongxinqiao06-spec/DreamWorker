package tools

import (
	"context"
	"strings"
	"time"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/resources"
)

func (r Registry) HandleChatToolCall(
	ctx context.Context,
	session resources.ChatSession,
	agent resources.AgentConfig,
	traceID string,
	request resources.ToolExecutionRequest,
) (resources.ChatToolCallPreview, resources.ToolExecutionResult) {
	startedAt := time.Now()
	callID := fallback(request.CallID, "call_"+r.nextRuntimeID("tool"))
	toolID := normalizeToolID(request.ToolID)
	if toolID == "" {
		toolID = normalizeToolID(request.DisplayName)
	}
	tool, ok := r.lookupTool(toolID)
	if !ok {
		call := resources.ChatToolCallPreview{
			CallID:      callID,
			ToolID:      toolID,
			DisplayName: fallback(request.DisplayName, toolID),
			RiskLevel:   "high",
			Status:      "blocked",
			Summary:     "工具未启用或未发现。",
			Arguments:   resources.RedactSecrets(request.Arguments),
			ErrorCode:   "TOOL_NOT_AVAILABLE",
		}
		return call, resources.ToolExecutionResult{
			CallID:       callID,
			ToolID:       toolID,
			Status:       "blocked",
			ErrorCode:    "TOOL_NOT_AVAILABLE",
			ErrorMessage: "tool is not enabled or discovered",
			LatencyMS:    resources.LatencyMS(startedAt),
		}
	}
	risk := fallback(tool.RiskLevel, "low")
	call := resources.ChatToolCallPreview{
		CallID:           callID,
		ToolID:           tool.ToolID,
		DisplayName:      fallback(tool.DisplayName, tool.ToolID),
		RiskLevel:        risk,
		ApprovalRequired: !IsAutoExecutableRisk(risk),
		Status:           "running",
		Summary:          "模型请求了工具调用，Engine 正在进行策略检查。",
		Arguments:        resources.RedactSecrets(request.Arguments),
	}
	if !agentAllowsTool(agent, tool.ToolID) && !strings.HasPrefix(tool.ToolID, "mcp_") {
		call.Status = "blocked"
		call.ErrorCode = "TOOL_NOT_ENABLED_FOR_AGENT"
		call.Summary = "当前 Agent 未启用该工具。"
		return call, resources.ToolExecutionResult{
			CallID:       callID,
			ToolID:       tool.ToolID,
			Status:       "blocked",
			ErrorCode:    "TOOL_NOT_ENABLED_FOR_AGENT",
			ErrorMessage: "tool is not enabled for the current agent",
			LatencyMS:    resources.LatencyMS(startedAt),
		}
	}
	if !IsAutoExecutableRisk(risk) {
		call.Status = "blocked"
		call.ErrorCode = "APPROVAL_REQUIRED"
		call.Summary = "该工具存在外部副作用或中高风险，需要确认后执行。"
		return call, resources.ToolExecutionResult{
			CallID:       callID,
			ToolID:       tool.ToolID,
			Status:       "blocked",
			ErrorCode:    "APPROVAL_REQUIRED",
			ErrorMessage: "approval is required before executing this tool",
			LatencyMS:    resources.LatencyMS(startedAt),
		}
	}
	result := r.executeLowRiskTool(ctx, session, traceID, tool, request.Arguments)
	result.CallID = callID
	result.ToolID = tool.ToolID
	result.LatencyMS = resources.LatencyMS(startedAt)
	if result.Status == "" {
		result.Status = "completed"
	}
	call.Status = result.Status
	call.ResultSummary = result.OutputSummary
	call.ErrorCode = result.ErrorCode
	if result.Status == "completed" {
		call.Summary = result.OutputSummary
	}
	return call, result
}

func (r Registry) nextRuntimeID(prefix string) string {
	r.State.Mu.Lock()
	defer r.State.Mu.Unlock()
	return r.State.NextIDLocked(prefix)
}
