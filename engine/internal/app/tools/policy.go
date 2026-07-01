package tools

import (
	"strings"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/resources"
)

func IsAutoExecutableRisk(risk string) bool {
	return risk == "" || risk == "low"
}

func agentAllowsTool(agent resources.AgentConfig, toolID string) bool {
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
