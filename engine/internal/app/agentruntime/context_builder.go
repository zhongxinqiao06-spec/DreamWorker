package agentruntime

import (
	"fmt"
	"sort"
	"strings"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/resources"
)

const contextSummaryVersion = 1

func (r Runtime) BuildChatContext(
	session resources.ChatSession,
	assistantMessageID string,
	agent resources.AgentConfig,
	profile resources.ModelProfile,
	provider resources.ModelProviderRecord,
	providerFallback string,
) (resources.ChatContextPack, *resources.ChatStreamWarning) {
	return BuildChatContext(r.State, session, assistantMessageID, agent, profile, provider, providerFallback)
}

func BuildChatContext(
	state *resources.Store,
	session resources.ChatSession,
	assistantMessageID string,
	agent resources.AgentConfig,
	profile resources.ModelProfile,
	provider resources.ModelProviderRecord,
	providerFallback string,
) (resources.ChatContextPack, *resources.ChatStreamWarning) {
	_ = provider
	system := buildSystemPrompt(state, session, agent)
	skills := resolveSkillRuntime(state, agent)
	tools := resolveToolRuntime(state, agent)
	mcpServers := resolveMCPServerRuntime(state, agent)
	history := completedChatHistory(state, session.SessionID, assistantMessageID)

	contextWindow := profile.ContextWindow
	if contextWindow <= 0 {
		contextWindow = agent.RuntimeConfig.ContextWindow
	}
	if contextWindow <= 0 {
		contextWindow = 128000
	}
	maxOutput := profile.MaxTokens
	if maxOutput <= 0 {
		maxOutput = agent.RuntimeConfig.MaxTokens
	}
	if maxOutput <= 0 {
		maxOutput = 4096
	}
	inputBudget := contextWindow - maxOutput
	if inputBudget < 1024 {
		inputBudget = max(256, contextWindow/2)
	}

	pack := resources.ChatContextPack{
		SystemPrompt:     system,
		Skills:           skills,
		Tools:            tools,
		MCPServers:       mcpServers,
		ProviderFallback: providerFallback,
		Budget: resources.ContextBudgetReport{
			ContextWindow:     contextWindow,
			MaxOutputTokens:   maxOutput,
			InputBudgetTokens: inputBudget,
		},
	}

	systemTokens := resources.EstimateTokens(system)
	pack.Budget.SystemTokens = systemTokens
	recentBudget := inputBudget - systemTokens - 512
	if recentBudget < 256 {
		recentBudget = 256
	}

	historyTokens := estimateGatewayMessages(history)
	if historyTokens <= recentBudget && len(history) <= 32 {
		pack.Messages = append([]resources.ChatGatewayMessage{{Role: "system", Content: system}}, history...)
		pack.Budget.RecentMessageCount = len(history)
		pack.Budget.RecentMessageTokens = historyTokens
		pack.Budget.EstimatedTokens = systemTokens + historyTokens
		if providerFallback != "" {
			pack.Budget.Warnings = append(pack.Budget.Warnings, "MODEL_PROFILE_FALLBACK")
		}
		return pack, nil
	}

	recent, compacted := splitHistoryForBudget(history, recentBudget)
	summary := ensureContextSummary(state, session.SessionID, compacted)
	if summary.Content != "" {
		pack.Summaries = []resources.ChatContextSummary{summary}
		summaryMessage := resources.ChatGatewayMessage{
			Role:    "system",
			Content: "历史对话摘要：\n" + summary.Content,
		}
		pack.Messages = append(pack.Messages, resources.ChatGatewayMessage{Role: "system", Content: system}, summaryMessage)
	} else {
		pack.Messages = append(pack.Messages, resources.ChatGatewayMessage{Role: "system", Content: system})
	}
	pack.Messages = append(pack.Messages, recent...)
	pack.Budget.Compacted = len(compacted) > 0
	pack.Budget.CompactedCount = len(compacted)
	pack.Budget.RecentMessageCount = len(recent)
	pack.Budget.RecentMessageTokens = estimateGatewayMessages(recent)
	pack.Budget.SummaryTokens = summary.TokenEstimate
	pack.Budget.EstimatedTokens = estimateGatewayMessages(pack.Messages)
	if pack.Budget.EstimatedTokens > inputBudget {
		pack.Budget.Warnings = append(pack.Budget.Warnings, "CONTEXT_BUDGET_EXCEEDED")
	}
	if providerFallback != "" {
		pack.Budget.Warnings = append(pack.Budget.Warnings, "MODEL_PROFILE_FALLBACK")
	}
	warning := &resources.ChatStreamWarning{
		Code:    "CONTEXT_COMPACTED",
		Message: fmt.Sprintf("已将 %d 条历史消息压缩为上下文摘要。", len(compacted)),
	}
	return pack, warning
}

func buildSystemPrompt(state *resources.Store, session resources.ChatSession, agent resources.AgentConfig) string {
	system := strings.TrimSpace(resources.RedactSecrets(agent.SystemPrompt))
	if system == "" {
		system = "你是 DreamWorker Agent。"
	}
	parts := []string{
		system,
		"运行环境：AI OS + Agent Runtime + 项目孵化系统。回答要清晰、可执行，并基于已提供的上下文。",
	}
	if session.ProjectID != nil {
		if project, ok := state.Projects[*session.ProjectID]; ok {
			parts = append(parts, fmt.Sprintf("项目：%s\n项目状态：%s\n项目目标：%s",
				resources.RedactSecrets(project.Title),
				resources.RedactSecrets(project.Status),
				resources.RedactSecrets(project.Description),
			))
			if modules, ok := state.Modules[project.ProjectID]; ok {
				moduleSummaries := make([]string, 0, len(modules))
				for _, module := range modules {
					moduleSummaries = append(moduleSummaries, fmt.Sprintf("%s=%s", module.ModuleID, resources.RedactSecrets(module.Summary)))
				}
				sort.Strings(moduleSummaries)
				if len(moduleSummaries) > 0 {
					parts = append(parts, "项目模块："+strings.Join(moduleSummaries, " | "))
				}
			}
		}
	}
	if len(agent.EnabledSkills) > 0 {
		parts = append(parts, "已启用 Skill："+strings.Join(agent.EnabledSkills, ", "))
	}
	if len(agent.EnabledTools) > 0 {
		parts = append(parts, "可用工具受策略控制；低风险工具可自动执行，写入、命令、网络和外部副作用需要审批。")
	}
	if len(agent.EnabledMCPServers) > 0 {
		parts = append(parts, "MCP 服务通过 Engine 暴露工具；不得复述密钥或环境变量值。")
	}
	return strings.Join(parts, "\n\n")
}

func completedChatHistory(state *resources.Store, sessionID string, assistantMessageID string) []resources.ChatGatewayMessage {
	history := state.Messages[sessionID]
	result := make([]resources.ChatGatewayMessage, 0, len(history))
	for _, message := range history {
		if message.MessageID == assistantMessageID || !chatMessageHasModelContent(message) {
			continue
		}
		if message.Role != "user" && message.Role != "assistant" {
			continue
		}
		if message.Role == "assistant" && message.Status != "completed" {
			continue
		}
		result = append(result, chatMessageToGatewayMessage(message))
	}
	return result
}

func resolveSkillRuntime(state *resources.Store, agent resources.AgentConfig) []resources.SkillRuntimeDescriptor {
	result := make([]resources.SkillRuntimeDescriptor, 0, len(agent.EnabledSkills))
	for _, skillID := range agent.EnabledSkills {
		skill, ok := state.Skills[skillID]
		if !ok || !skill.Enabled {
			continue
		}
		instruction := strings.TrimSpace(skill.Instructions)
		if instruction == "" {
			instruction = strings.TrimSpace(skill.Description)
		}
		if instruction == "" {
			instruction = "仅在它能直接改善当前任务时使用。"
		}
		result = append(result, resources.SkillRuntimeDescriptor{
			SkillID:              skill.SkillID,
			DisplayName:          skill.DisplayName,
			Instruction:          resources.RedactSecrets(instruction),
			RequiredCapabilities: append([]string{}, skill.RequiredCapabilities...),
			OutputArtifacts:      append([]string{}, skill.OutputArtifacts...),
			RuntimePolicy:        "instructions_and_allowed_tools",
		})
	}
	return result
}

func resolveToolRuntime(state *resources.Store, agent resources.AgentConfig) []resources.ToolRuntimeDescriptor {
	result := make([]resources.ToolRuntimeDescriptor, 0, len(agent.EnabledTools))
	for _, toolID := range agent.EnabledTools {
		tool, ok := state.Tools[toolID]
		if !ok || !tool.Enabled {
			continue
		}
		risk := fallback(tool.RiskLevel, "low")
		result = append(result, resources.ToolRuntimeDescriptor{
			ToolID:           tool.ToolID,
			DisplayName:      tool.DisplayName,
			Description:      resources.RedactSecrets(tool.Description),
			RiskLevel:        risk,
			AutoExecutable:   IsAutoExecutableRisk(risk),
			ApprovalRequired: !IsAutoExecutableRisk(risk),
		})
	}
	return result
}

func resolveMCPServerRuntime(state *resources.Store, agent resources.AgentConfig) []string {
	result := make([]string, 0, len(agent.EnabledMCPServers))
	for _, serverID := range agent.EnabledMCPServers {
		server, ok := state.Servers[serverID]
		if !ok || !server.Enabled {
			continue
		}
		result = append(result, server.ServerID)
	}
	return sortedStrings(result)
}

func splitHistoryForBudget(history []resources.ChatGatewayMessage, budget int) ([]resources.ChatGatewayMessage, []resources.ChatGatewayMessage) {
	recentTokens := 0
	start := len(history)
	for index := len(history) - 1; index >= 0; index-- {
		nextTokens := gatewayMessageTokenEstimate(history[index])
		if start != len(history) && recentTokens+nextTokens > budget {
			break
		}
		recentTokens += nextTokens
		start = index
	}
	recent := append([]resources.ChatGatewayMessage{}, history[start:]...)
	compacted := append([]resources.ChatGatewayMessage{}, history[:start]...)
	return recent, compacted
}

func ensureContextSummary(state *resources.Store, sessionID string, messages []resources.ChatGatewayMessage) resources.ChatContextSummary {
	if len(messages) == 0 {
		return resources.ChatContextSummary{}
	}
	sourceIDs := make([]string, 0, len(messages))
	var sourceBuilder strings.Builder
	for index, message := range messages {
		sourceID := fmt.Sprintf("%s_%03d", message.Role, index)
		sourceIDs = append(sourceIDs, sourceID)
		sourceBuilder.WriteString(message.Role)
		sourceBuilder.WriteString(":")
		sourceBuilder.WriteString(gatewayMessageText(message))
		sourceBuilder.WriteString("\n")
	}
	hash := ContentHash(sourceBuilder.String())
	for _, existing := range state.ContextSummaries[sessionID] {
		if existing.ContentHash == hash {
			return existing
		}
	}
	content := deterministicContextSummary(messages)
	now := state.Now()
	summary := resources.ChatContextSummary{
		SummaryID:        "ctx_" + state.NextIDLocked("ctx"),
		SessionID:        sessionID,
		SourceMessageIDs: sourceIDs,
		Content:          content,
		ContentHash:      hash,
		TokenEstimate:    resources.EstimateTokens(content),
		CreatedBy:        "deterministic_extractive",
		ContextVersion:   contextSummaryVersion,
		CreatedAt:        now,
	}
	state.ContextSummaries[sessionID] = append(state.ContextSummaries[sessionID], summary)
	return summary
}

func deterministicContextSummary(messages []resources.ChatGatewayMessage) string {
	if len(messages) == 0 {
		return ""
	}
	lines := make([]string, 0, min(len(messages), 12))
	start := 0
	if len(messages) > 12 {
		start = len(messages) - 12
	}
	for _, message := range messages[start:] {
		content := strings.Join(strings.Fields(resources.RedactSecrets(gatewayMessageText(message))), " ")
		runes := []rune(content)
		if len(runes) > 220 {
			content = string(runes[:220]) + "..."
		}
		lines = append(lines, fmt.Sprintf("- %s: %s", message.Role, content))
	}
	return strings.Join(lines, "\n")
}

func estimateGatewayMessages(messages []resources.ChatGatewayMessage) int {
	total := 0
	for _, message := range messages {
		total += gatewayMessageTokenEstimate(message)
	}
	return total
}

func chatMessageHasModelContent(message resources.ChatMessage) bool {
	if strings.TrimSpace(message.Content) != "" {
		return true
	}
	for _, part := range message.Parts {
		if strings.TrimSpace(part.Text) != "" || strings.TrimSpace(part.DataURL) != "" || strings.TrimSpace(part.URL) != "" {
			return true
		}
	}
	return false
}

func chatMessageToGatewayMessage(message resources.ChatMessage) resources.ChatGatewayMessage {
	gateway := resources.ChatGatewayMessage{
		Role:    message.Role,
		Content: resources.RedactSecrets(message.Content),
	}
	if message.Role != "user" {
		return gateway
	}
	parts := make([]resources.ChatGatewayContentPart, 0, len(message.Parts))
	for _, part := range message.Parts {
		switch part.Type {
		case "text":
			text := strings.TrimSpace(resources.RedactSecrets(part.Text))
			if text != "" {
				parts = append(parts, resources.ChatGatewayContentPart{Type: "text", Text: text})
			}
		case "image", "image_url":
			url := strings.TrimSpace(part.DataURL)
			if url == "" {
				url = strings.TrimSpace(part.URL)
			}
			if url == "" {
				continue
			}
			parts = append(parts, resources.ChatGatewayContentPart{
				Type: "image_url",
				ImageURL: &resources.ChatGatewayImageURL{
					URL:    url,
					Detail: strings.TrimSpace(part.Detail),
				},
			})
		}
	}
	if len(parts) > 0 {
		gateway.Parts = parts
	}
	return gateway
}

func gatewayMessageText(message resources.ChatGatewayMessage) string {
	text := strings.TrimSpace(message.Content)
	for _, part := range message.Parts {
		if part.Type == "text" && strings.TrimSpace(part.Text) != "" && !strings.Contains(text, strings.TrimSpace(part.Text)) {
			if text != "" {
				text += "\n"
			}
			text += strings.TrimSpace(part.Text)
		}
	}
	if count := gatewayImageCount(message); count > 0 {
		if text != "" {
			text += "\n"
		}
		if count == 1 {
			text += "[image attached]"
		} else {
			text += fmt.Sprintf("[%d images attached]", count)
		}
	}
	return text
}

func gatewayImageCount(message resources.ChatGatewayMessage) int {
	count := 0
	for _, part := range message.Parts {
		if part.Type == "image_url" && part.ImageURL != nil && strings.TrimSpace(part.ImageURL.URL) != "" {
			count++
		}
	}
	return count
}

func gatewayMessageTokenEstimate(message resources.ChatGatewayMessage) int {
	return resources.EstimateTokens(gatewayMessageText(message)) + gatewayImageCount(message)*180
}

func IsAutoExecutableRisk(risk string) bool {
	return risk == "" || risk == "low"
}
