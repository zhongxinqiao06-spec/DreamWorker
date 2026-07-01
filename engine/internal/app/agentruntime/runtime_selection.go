package agentruntime

import (
	"fmt"
	"strings"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/resources"
)

func BuildRuntimeSelection(contextPack resources.ChatContextPack) resources.ChatRuntimeSelection {
	skillNames := make([]string, 0, len(contextPack.Skills))
	for _, skill := range contextPack.Skills {
		skillNames = append(skillNames, fallback(skill.DisplayName, skill.SkillID))
	}
	toolNames := make([]string, 0, len(contextPack.Tools))
	for _, tool := range contextPack.Tools {
		toolNames = append(toolNames, fallback(tool.DisplayName, tool.ToolID))
	}
	summaryParts := []string{
		fmt.Sprintf("%d 个 Skill", len(contextPack.Skills)),
		fmt.Sprintf("%d 个工具", len(contextPack.Tools)),
		fmt.Sprintf("上下文约 %d token", contextPack.Budget.EstimatedTokens),
	}
	if len(skillNames) > 0 {
		summaryParts = append(summaryParts, "Skill: "+strings.Join(skillNames, "、"))
	}
	if len(toolNames) > 0 {
		summaryParts = append(summaryParts, "工具: "+strings.Join(toolNames, "、"))
	}
	return resources.ChatRuntimeSelection{
		Summary:    strings.Join(summaryParts, " / "),
		Skills:     append([]resources.SkillRuntimeDescriptor{}, contextPack.Skills...),
		Tools:      append([]resources.ToolRuntimeDescriptor{}, contextPack.Tools...),
		MCPServers: append([]string{}, contextPack.MCPServers...),
	}
}

func FirstContextSummary(pack resources.ChatContextPack) *resources.ChatContextSummary {
	if len(pack.Summaries) == 0 {
		return nil
	}
	summary := pack.Summaries[0]
	return &summary
}

func BuildRuntimeSummary(
	agent resources.AgentConfig,
	session resources.ChatSession,
	provider resources.ModelProviderRecord,
	profile resources.ModelProfile,
) string {
	return fmt.Sprintf("Agent=%s / Provider=%s / Model=%s / ModelProfile=%s / Memory=%s / Planner=%s / Timeout=%dms",
		fallback(agent.AgentID, session.AgentID),
		provider.ProviderID,
		profile.Model,
		session.ModelProfileID,
		fallback(agent.MemoryScope, "short_term"),
		fallback(agent.Planner.Strategy, "plan-execute"),
		agent.Executor.TimeoutMS,
	)
}
