package tools

import (
	"fmt"
	"strings"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/resources"
)

func BuildChatToolCalls(state *resources.Store, agent resources.AgentConfig) []resources.ChatToolCallPreview {
	result := make([]resources.ChatToolCallPreview, 0, len(agent.EnabledTools))
	for _, toolID := range agent.EnabledTools {
		tool := state.Tools[toolID]
		risk := fallback(tool.RiskLevel, "low")
		name := fallback(tool.DisplayName, toolID)
		result = append(result, resources.ChatToolCallPreview{
			CallID:           "call_" + toolID,
			ToolID:           toolID,
			DisplayName:      name,
			RiskLevel:        risk,
			ApprovalRequired: !IsAutoExecutableRisk(risk),
			Status:           "preview",
			Summary:          "模型可选择该工具；低风险工具可自动执行，需要外部副作用的工具会被拦截等待确认。",
		})
	}
	return result
}

func (r Registry) lookupTool(toolID string) (resources.ToolConfig, bool) {
	r.State.Mu.Lock()
	defer r.State.Mu.Unlock()
	if tool, ok := r.State.Tools[toolID]; ok && tool.Enabled {
		return tool, true
	}
	if binding, ok := r.State.MCPTools[toolID]; ok {
		if tool, ok := r.State.Tools[binding.ToolID]; ok && tool.Enabled {
			return tool, true
		}
	}
	return resources.ToolConfig{}, false
}

func UpsertToolCall(calls []resources.ChatToolCallPreview, call resources.ChatToolCallPreview) []resources.ChatToolCallPreview {
	for index, existing := range calls {
		if existing.CallID == call.CallID || existing.ToolID == call.ToolID {
			calls[index] = call
			return calls
		}
	}
	return append(calls, call)
}

func FormatToolResultForModel(result resources.ToolExecutionResult) string {
	status := fallback(result.Status, "completed")
	summary := result.OutputSummary
	if summary == "" {
		summary = result.ErrorMessage
	}
	return fmt.Sprintf("- %s status=%s error=%s summary=%s",
		result.ToolID,
		status,
		result.ErrorCode,
		resources.RedactSecrets(summary),
	)
}

func fallback(value string, fallbackValue string) string {
	if strings.TrimSpace(value) == "" {
		return fallbackValue
	}
	return value
}
