package workspace

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func (s *Store) resolveChatModelBindingLocked(
	session ChatSession,
	agent AgentConfig,
) (ModelProfile, ModelProviderRecord, string, *AppError) {
	seen := map[string]bool{}
	var skipped []string
	var candidates []string
	appendCandidate := func(profileID string) {
		profileID = strings.TrimSpace(profileID)
		if profileID == "" || seen[profileID] {
			return
		}
		seen[profileID] = true
		candidates = append(candidates, profileID)
	}
	appendCandidate(session.ModelProfileID)
	appendCandidate(agent.ModelProfileID)
	if session.ProjectID != nil {
		if project, ok := s.projects[*session.ProjectID]; ok {
			appendCandidate(project.DefaultModelProfileID)
		}
	}
	appendCandidate("profile_fast")
	appendCandidate("profile_openai")
	appendCandidate("profile_anthropic")
	appendCandidate("profile_deepseek")
	appendCandidate("profile_ollama")
	appendCandidate("profile_stub")

	for index := 0; index < len(candidates); index++ {
		profileID := candidates[index]
		profile, ok := s.profiles[profileID]
		if !ok || !profile.Enabled {
			skipped = append(skipped, profileID+":profile_unavailable")
			continue
		}
		profile = ensureModelProfileDefaults(profile)
		if profile.FallbackProfileID != nil {
			appendCandidate(*profile.FallbackProfileID)
		}
		provider, ok := s.providers[profile.ProviderID]
		if !ok || !provider.Enabled {
			skipped = append(skipped, profileID+":provider_unavailable")
			continue
		}
		if !providerHasUsableCredential(provider) {
			skipped = append(skipped, profileID+":credential_missing")
			continue
		}
		fallbackPath := ""
		if len(skipped) > 0 || profile.ProfileID != session.ModelProfileID {
			fallbackPath = strings.Join(append(skipped, profile.ProfileID+":selected"), " -> ")
		}
		return profile, provider, fallbackPath, nil
	}
	return ModelProfile{}, ModelProviderRecord{}, "", BadRequest(
		"MODEL_PROVIDER_UNAVAILABLE",
		"no usable model profile is available",
		"configure a provider key or select the offline stub profile",
	)
}

func providerHasUsableCredential(provider ModelProviderRecord) bool {
	if provider.ProviderID == "provider_local_stub" || provider.DefaultModel == "model_generate_stub" {
		return true
	}
	if provider.ProviderType == ProviderOllama {
		return true
	}
	if provider.APIKey == "" || provider.APIKey == "sk-local-demo" {
		return false
	}
	return true
}

func firstContextSummary(pack ChatContextPack) *ChatContextSummary {
	if len(pack.Summaries) == 0 {
		return nil
	}
	summary := pack.Summaries[0]
	return &summary
}

func (s *Store) handleChatToolCall(
	ctx context.Context,
	session ChatSession,
	agent AgentConfig,
	traceID string,
	request ToolExecutionRequest,
) (ChatToolCallPreview, ToolExecutionResult) {
	startedAt := time.Now()
	callID := fallback(request.CallID, "call_"+s.nextRuntimeID("tool"))
	toolID := normalizeToolID(request.ToolID)
	if toolID == "" {
		toolID = normalizeToolID(request.DisplayName)
	}
	tool, ok := s.lookupTool(toolID)
	if !ok {
		call := ChatToolCallPreview{
			CallID:      callID,
			ToolID:      toolID,
			DisplayName: fallback(request.DisplayName, toolID),
			RiskLevel:   "high",
			Status:      "blocked",
			Summary:     "Tool is not enabled for this agent or was not discovered.",
			Arguments:   redactSecrets(request.Arguments),
			ErrorCode:   "TOOL_NOT_AVAILABLE",
		}
		return call, ToolExecutionResult{
			CallID:       callID,
			ToolID:       toolID,
			Status:       "blocked",
			ErrorCode:    "TOOL_NOT_AVAILABLE",
			ErrorMessage: "tool is not enabled or discovered",
			LatencyMS:    latencyMS(startedAt),
		}
	}
	risk := fallback(tool.RiskLevel, "low")
	call := ChatToolCallPreview{
		CallID:           callID,
		ToolID:           tool.ToolID,
		DisplayName:      fallback(tool.DisplayName, tool.ToolID),
		RiskLevel:        risk,
		ApprovalRequired: !isAutoExecutableRisk(risk),
		Status:           "running",
		Summary:          "Tool call was requested by the model and evaluated by policy.",
		Arguments:        redactSecrets(request.Arguments),
	}
	if !agentAllowsTool(agent, tool.ToolID) && !strings.HasPrefix(tool.ToolID, "mcp_") {
		call.Status = "blocked"
		call.ErrorCode = "TOOL_NOT_ENABLED_FOR_AGENT"
		call.Summary = "Tool is not enabled for the current agent."
		return call, ToolExecutionResult{
			CallID:       callID,
			ToolID:       tool.ToolID,
			Status:       "blocked",
			ErrorCode:    "TOOL_NOT_ENABLED_FOR_AGENT",
			ErrorMessage: "tool is not enabled for the current agent",
			LatencyMS:    latencyMS(startedAt),
		}
	}
	if !isAutoExecutableRisk(risk) {
		call.Status = "blocked"
		call.ErrorCode = "APPROVAL_REQUIRED"
		call.Summary = "Tool has side effects or medium/high risk and requires approval before execution."
		return call, ToolExecutionResult{
			CallID:       callID,
			ToolID:       tool.ToolID,
			Status:       "blocked",
			ErrorCode:    "APPROVAL_REQUIRED",
			ErrorMessage: "approval is required before executing this tool",
			LatencyMS:    latencyMS(startedAt),
		}
	}
	result := s.executeLowRiskTool(ctx, session, traceID, tool, request.Arguments)
	result.CallID = callID
	result.ToolID = tool.ToolID
	result.LatencyMS = latencyMS(startedAt)
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

func (s *Store) lookupTool(toolID string) (ToolConfig, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if tool, ok := s.tools[toolID]; ok && tool.Enabled {
		return tool, true
	}
	if binding, ok := s.mcpTools[toolID]; ok {
		if tool, ok := s.tools[binding.ToolID]; ok && tool.Enabled {
			return tool, true
		}
	}
	return ToolConfig{}, false
}

func (s *Store) executeLowRiskTool(
	ctx context.Context,
	session ChatSession,
	traceID string,
	tool ToolConfig,
	arguments string,
) ToolExecutionResult {
	_ = session
	if binding, ok := s.lookupMCPBinding(tool.ToolID); ok {
		result, err := s.callMCPTool(ctx, binding, arguments)
		if err != nil {
			return ToolExecutionResult{
				Status:       "blocked",
				ErrorCode:    "MCP_TOOL_FAILED",
				ErrorMessage: redactSecrets(err.Error()),
			}
		}
		return ToolExecutionResult{Status: "completed", OutputSummary: result}
	}
	switch tool.ToolID {
	case "tool_model_generate_stub":
		return ToolExecutionResult{
			Status:        "completed",
			OutputSummary: "Deterministic model helper executed inside the Engine for trace " + traceID + ".",
		}
	case "tool_human_input":
		return ToolExecutionResult{
			Status:       "blocked",
			ErrorCode:    "HUMAN_INPUT_REQUIRED",
			ErrorMessage: "human input requires an explicit UI handoff",
		}
	case "tool_artifact_read":
		return ToolExecutionResult{
			Status:        "completed",
			OutputSummary: "Artifact read tool is available for project-scoped artifacts; no artifact_id was provided.",
		}
	default:
		return ToolExecutionResult{
			Status:        "completed",
			OutputSummary: fmt.Sprintf("%s executed with policy-gated low-risk runtime.", tool.ToolID),
		}
	}
}

func (s *Store) lookupMCPBinding(toolID string) (MCPToolBinding, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	binding, ok := s.mcpTools[toolID]
	return binding, ok
}

func agentAllowsTool(agent AgentConfig, toolID string) bool {
	for _, enabled := range agent.EnabledTools {
		if enabled == toolID {
			return true
		}
	}
	return false
}

func normalizeToolID(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	value = strings.ReplaceAll(value, ".", "_")
	value = strings.ReplaceAll(value, "-", "_")
	switch value {
	case "artifact_read", "cap_artifact_read":
		return "tool_artifact_read"
	case "artifact_write", "cap_artifact_write":
		return "tool_artifact_write"
	case "web_search", "web_search_stub", "cap_web_search_stub":
		return "tool_web_search_stub"
	case "browser_readonly", "browser_readonly_stub", "cap_browser_readonly_stub":
		return "tool_browser_readonly_stub"
	case "model_generate", "model_generate_stub", "cap_model_generate_stub":
		return "tool_model_generate_stub"
	case "human_input", "cap_human_input":
		return "tool_human_input"
	default:
		return value
	}
}

func upsertToolCall(calls []ChatToolCallPreview, call ChatToolCallPreview) []ChatToolCallPreview {
	for index, existing := range calls {
		if existing.CallID == call.CallID || existing.ToolID == call.ToolID {
			calls[index] = call
			return calls
		}
	}
	return append(calls, call)
}

func formatToolResultForModel(result ToolExecutionResult) string {
	status := fallback(result.Status, "completed")
	summary := result.OutputSummary
	if summary == "" {
		summary = result.ErrorMessage
	}
	return fmt.Sprintf("- %s status=%s error=%s summary=%s",
		result.ToolID,
		status,
		result.ErrorCode,
		redactSecrets(summary),
	)
}

func (s *Store) nextRuntimeID(prefix string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.nextIDLocked(prefix)
}

func decodeToolArguments(arguments string) map[string]any {
	var value map[string]any
	if err := json.Unmarshal([]byte(arguments), &value); err != nil {
		return map[string]any{}
	}
	return value
}
