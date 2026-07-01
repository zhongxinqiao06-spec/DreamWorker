package workspace

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

const contextSummaryVersion = 1

var secretPattern = regexp.MustCompile(`(?i)(sk-[a-z0-9._+=:/-]+|bearer\s+[a-z0-9._+=:/-]+|api[_-]?key\s*[:=]\s*[\S]+|token\s*[:=]\s*[\S]+)`)

func (s *Store) buildChatContextLocked(
	session ChatSession,
	assistantMessageID string,
	agent AgentConfig,
	profile ModelProfile,
	provider ModelProviderRecord,
	providerFallback string,
) (ChatContextPack, *ChatStreamWarning) {
	system := s.buildSystemPromptLocked(session, agent)
	skills := s.resolveSkillRuntimeLocked(agent)
	tools := s.resolveToolRuntimeLocked(agent)
	mcpServers := s.resolveMCPServerRuntimeLocked(agent)

	history := s.completedChatHistoryLocked(session.SessionID, assistantMessageID)
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

	pack := ChatContextPack{
		SystemPrompt:     system,
		Skills:           skills,
		Tools:            tools,
		MCPServers:       mcpServers,
		ProviderFallback: providerFallback,
		Budget: ContextBudgetReport{
			ContextWindow:     contextWindow,
			MaxOutputTokens:   maxOutput,
			InputBudgetTokens: inputBudget,
		},
	}

	systemTokens := estimateTokens(system)
	pack.Budget.SystemTokens = systemTokens
	recentBudget := inputBudget - systemTokens - 512
	if recentBudget < 256 {
		recentBudget = 256
	}

	historyTokens := estimateGatewayMessages(history)
	if historyTokens <= recentBudget && len(history) <= 32 {
		pack.Messages = append([]ChatGatewayMessage{{Role: "system", Content: system}}, history...)
		pack.Budget.RecentMessageCount = len(history)
		pack.Budget.RecentMessageTokens = historyTokens
		pack.Budget.EstimatedTokens = systemTokens + historyTokens
		return pack, nil
	}

	recent, compacted := splitHistoryForBudget(history, recentBudget)
	summary := s.ensureContextSummaryLocked(session.SessionID, compacted)
	if summary.Content != "" {
		pack.Summaries = []ChatContextSummary{summary}
		summaryMessage := ChatGatewayMessage{
			Role:    "system",
			Content: "Compressed prior conversation summary:\n" + summary.Content,
		}
		pack.Messages = append(pack.Messages, ChatGatewayMessage{Role: "system", Content: system}, summaryMessage)
	} else {
		pack.Messages = append(pack.Messages, ChatGatewayMessage{Role: "system", Content: system})
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
	warning := &ChatStreamWarning{
		Code:    "CONTEXT_COMPACTED",
		Message: fmt.Sprintf("Compacted %d prior messages into a reusable context summary.", len(compacted)),
	}
	return pack, warning
}

func (s *Store) buildSystemPromptLocked(session ChatSession, agent AgentConfig) string {
	system := strings.TrimSpace(redactSecrets(agent.SystemPrompt))
	if system == "" {
		system = "You are a DreamWorker agent."
	}
	parts := []string{
		system,
		"Runtime: AI OS + Agent Runtime + project incubation system. Keep answers clear, actionable, and grounded in the provided context.",
	}
	if session.ProjectID != nil {
		if project, ok := s.projects[*session.ProjectID]; ok {
			parts = append(parts, fmt.Sprintf("Project: %s\nProject status: %s\nProject goal: %s",
				redactSecrets(project.Title),
				redactSecrets(project.Status),
				redactSecrets(project.Description),
			))
			if modules, ok := s.modules[project.ProjectID]; ok {
				moduleSummaries := make([]string, 0, len(modules))
				for _, module := range modules {
					moduleSummaries = append(moduleSummaries, fmt.Sprintf("%s=%s", module.ModuleID, redactSecrets(module.Summary)))
				}
				sort.Strings(moduleSummaries)
				if len(moduleSummaries) > 0 {
					parts = append(parts, "Project modules: "+strings.Join(moduleSummaries, " | "))
				}
			}
		}
	}
	if len(agent.EnabledSkills) > 0 {
		parts = append(parts, "Enabled skills: "+strings.Join(agent.EnabledSkills, ", "))
	}
	if len(agent.EnabledTools) > 0 {
		parts = append(parts, "Available tools are policy-gated. Low-risk tools may execute automatically; write, shell, network, and external side effects require approval.")
	}
	if len(agent.EnabledMCPServers) > 0 {
		parts = append(parts, "MCP servers may expose tools through the Engine. Secrets and environment values must never be repeated.")
	}
	return strings.Join(parts, "\n\n")
}

func (s *Store) completedChatHistoryLocked(sessionID string, assistantMessageID string) []ChatGatewayMessage {
	history := s.messages[sessionID]
	result := make([]ChatGatewayMessage, 0, len(history))
	for _, message := range history {
		if message.MessageID == assistantMessageID || strings.TrimSpace(message.Content) == "" {
			continue
		}
		if message.Role != "user" && message.Role != "assistant" {
			continue
		}
		if message.Role == "assistant" && message.Status != "completed" {
			continue
		}
		result = append(result, ChatGatewayMessage{Role: message.Role, Content: redactSecrets(message.Content)})
	}
	return result
}

func (s *Store) resolveSkillRuntimeLocked(agent AgentConfig) []SkillRuntimeDescriptor {
	result := make([]SkillRuntimeDescriptor, 0, len(agent.EnabledSkills))
	for _, skillID := range agent.EnabledSkills {
		skill, ok := s.skills[skillID]
		if !ok || !skill.Enabled {
			continue
		}
		instruction := strings.TrimSpace(skill.Instructions)
		if instruction == "" {
			instruction = strings.TrimSpace(skill.Description)
		}
		if instruction == "" {
			instruction = "Use this skill only when it directly improves the current task."
		}
		result = append(result, SkillRuntimeDescriptor{
			SkillID:              skill.SkillID,
			DisplayName:          skill.DisplayName,
			Instruction:          redactSecrets(instruction),
			RequiredCapabilities: append([]string{}, skill.RequiredCapabilities...),
			OutputArtifacts:      append([]string{}, skill.OutputArtifacts...),
			RuntimePolicy:        "instructions_and_allowed_tools",
		})
	}
	return result
}

func (s *Store) resolveToolRuntimeLocked(agent AgentConfig) []ToolRuntimeDescriptor {
	result := make([]ToolRuntimeDescriptor, 0, len(agent.EnabledTools))
	for _, toolID := range agent.EnabledTools {
		tool, ok := s.tools[toolID]
		if !ok || !tool.Enabled {
			continue
		}
		risk := fallback(tool.RiskLevel, "low")
		result = append(result, ToolRuntimeDescriptor{
			ToolID:           tool.ToolID,
			DisplayName:      tool.DisplayName,
			Description:      redactSecrets(tool.Description),
			RiskLevel:        risk,
			AutoExecutable:   isAutoExecutableRisk(risk),
			ApprovalRequired: !isAutoExecutableRisk(risk),
		})
	}
	return result
}

func (s *Store) resolveMCPServerRuntimeLocked(agent AgentConfig) []string {
	result := make([]string, 0, len(agent.EnabledMCPServers))
	for _, serverID := range agent.EnabledMCPServers {
		server, ok := s.servers[serverID]
		if !ok || !server.Enabled {
			continue
		}
		result = append(result, server.ServerID)
	}
	sort.Strings(result)
	return result
}

func splitHistoryForBudget(history []ChatGatewayMessage, budget int) ([]ChatGatewayMessage, []ChatGatewayMessage) {
	recentTokens := 0
	start := len(history)
	for index := len(history) - 1; index >= 0; index-- {
		nextTokens := estimateTokens(history[index].Content)
		if start != len(history) && recentTokens+nextTokens > budget {
			break
		}
		recentTokens += nextTokens
		start = index
	}
	recent := append([]ChatGatewayMessage{}, history[start:]...)
	compacted := append([]ChatGatewayMessage{}, history[:start]...)
	return recent, compacted
}

func (s *Store) ensureContextSummaryLocked(sessionID string, messages []ChatGatewayMessage) ChatContextSummary {
	if len(messages) == 0 {
		return ChatContextSummary{}
	}
	sourceIDs := make([]string, 0, len(messages))
	var sourceBuilder strings.Builder
	for index, message := range messages {
		sourceID := fmt.Sprintf("%s_%03d", message.Role, index)
		sourceIDs = append(sourceIDs, sourceID)
		sourceBuilder.WriteString(message.Role)
		sourceBuilder.WriteString(":")
		sourceBuilder.WriteString(message.Content)
		sourceBuilder.WriteString("\n")
	}
	hash := contentHash(sourceBuilder.String())
	for _, existing := range s.contextSummaries[sessionID] {
		if existing.ContentHash == hash {
			return existing
		}
	}
	content := deterministicContextSummary(messages)
	now := s.now()
	summary := ChatContextSummary{
		SummaryID:        "ctx_" + s.nextIDLocked("ctx"),
		SessionID:        sessionID,
		SourceMessageIDs: sourceIDs,
		Content:          content,
		ContentHash:      hash,
		TokenEstimate:    estimateTokens(content),
		CreatedBy:        "deterministic_extractive",
		ContextVersion:   contextSummaryVersion,
		CreatedAt:        now,
	}
	s.contextSummaries[sessionID] = append(s.contextSummaries[sessionID], summary)
	return summary
}

func deterministicContextSummary(messages []ChatGatewayMessage) string {
	if len(messages) == 0 {
		return ""
	}
	lines := make([]string, 0, min(len(messages), 12))
	start := 0
	if len(messages) > 12 {
		start = len(messages) - 12
	}
	for _, message := range messages[start:] {
		content := strings.Join(strings.Fields(redactSecrets(message.Content)), " ")
		runes := []rune(content)
		if len(runes) > 220 {
			content = string(runes[:220]) + "..."
		}
		lines = append(lines, fmt.Sprintf("- %s: %s", message.Role, content))
	}
	return strings.Join(lines, "\n")
}

func estimateGatewayMessages(messages []ChatGatewayMessage) int {
	total := 0
	for _, message := range messages {
		total += estimateTokens(message.Content)
	}
	return total
}

func redactSecrets(value string) string {
	return secretPattern.ReplaceAllString(value, "[redacted]")
}

func isAutoExecutableRisk(risk string) bool {
	return risk == "" || risk == "low"
}
