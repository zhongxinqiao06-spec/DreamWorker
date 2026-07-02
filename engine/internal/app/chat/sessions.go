package chat

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/resources"
)

func (s *Store) ListChatSessions() []ChatSession {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	return sortedValues(s.Sessions, func(item ChatSession) string { return item.Title })
}

func (s *Store) CreateChatSession(input CreateChatSessionInput) (ChatSession, *AppError) {
	if strings.TrimSpace(input.Title) == "" {
		return ChatSession{}, BadRequest("BAD_REQUEST", "chat title is required", "enter a title")
	}
	s.Mu.Lock()
	defer s.Mu.Unlock()
	now := s.Now()
	modelProfileID, providerID, model := s.ensureChatModelDefaultsLocked(input.ModelProfileID, input.ProviderID, input.Model)
	session := ChatSession{
		SessionID:      s.NextIDLocked("chat"),
		ProjectID:      input.ProjectID,
		Title:          input.Title,
		AgentID:        fallback(input.AgentID, "agent_general_assistant"),
		ModelProfileID: modelProfileID,
		ProviderID:     providerID,
		Model:          model,
		MessageCount:   0,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	s.Sessions[session.SessionID] = session
	s.Messages[session.SessionID] = []ChatMessage{}
	return session, nil
}

func (s *Store) UpdateChatSession(input UpdateChatSessionInput) (ChatSession, *AppError) {
	if input.SessionID == "" {
		return ChatSession{}, BadRequest("BAD_REQUEST", "missing sessionId", "select a chat session")
	}
	s.Mu.Lock()
	defer s.Mu.Unlock()
	session, ok := s.Sessions[input.SessionID]
	if !ok {
		return ChatSession{}, NotFound("SESSION_NOT_FOUND", "session not found", "create a new chat session")
	}
	if input.ProjectID != nil {
		if _, ok := s.Projects[*input.ProjectID]; !ok {
			return ChatSession{}, NotFound("PROJECT_NOT_FOUND", "project not found", "select another project")
		}
	}
	if input.AgentID != "" {
		if _, ok := s.Agents[input.AgentID]; !ok {
			return ChatSession{}, NotFound("AGENT_NOT_FOUND", "agent not found", "refresh agents")
		}
		session.AgentID = input.AgentID
	}
	if input.ModelProfileID != "" {
		profile, ok := s.Profiles[input.ModelProfileID]
		if !ok {
			return ChatSession{}, NotFound("MODEL_PROFILE_NOT_FOUND", "model profile not found", "refresh profiles")
		}
		session.ModelProfileID = input.ModelProfileID
		session.ProviderID = profile.ProviderID
		session.Model = profile.Model
	}
	if input.ProviderID != "" || input.Model != "" {
		modelProfileID, providerID, model := s.ensureChatModelDefaultsLocked(session.ModelProfileID, input.ProviderID, input.Model)
		session.ModelProfileID = modelProfileID
		session.ProviderID = providerID
		session.Model = model
	} else if session.ProviderID == "" || session.Model == "" {
		modelProfileID, providerID, model := s.ensureChatModelDefaultsLocked(session.ModelProfileID, "", "")
		session.ModelProfileID = modelProfileID
		session.ProviderID = providerID
		session.Model = model
	}
	if strings.TrimSpace(input.Title) != "" {
		session.Title = strings.TrimSpace(input.Title)
	}
	session.ProjectID = input.ProjectID
	session.UpdatedAt = s.Now()
	s.Sessions[input.SessionID] = session
	return session, nil
}

func (s *Store) ensureChatModelDefaultsLocked(profileID string, providerID string, model string) (string, string, string) {
	providerID = strings.TrimSpace(providerID)
	model = strings.TrimSpace(model)
	if providerID != "" {
		if provider, ok := s.Providers[providerID]; ok {
			if model == "" {
				model = provider.DefaultModel
			}
			nextProfileID := resources.ProfileIDForProviderModel(providerID, model)
			if _, ok := s.Profiles[nextProfileID]; !ok {
				s.Profiles[nextProfileID] = resources.ProfileFromProviderModel(provider, model, s.Now())
			}
			return nextProfileID, providerID, model
		}
	}
	profileID = fallback(profileID, "profile_fast")
	if profile, ok := s.Profiles[profileID]; ok {
		return profile.ProfileID, profile.ProviderID, profile.Model
	}
	if provider, ok := s.Providers["provider_deepseek"]; ok {
		model = provider.DefaultModel
		return resources.ProfileIDForProviderModel(provider.ProviderID, model), provider.ProviderID, model
	}
	return "profile_fast", "provider_deepseek", "deepseek-v4-flash"
}

func (s *Store) SendChatMessage(input SendChatMessageInput) (ChatTurnResult, *AppError) {
	events, appErr := s.StreamChatMessage(context.Background(), input)
	if appErr != nil {
		return ChatTurnResult{}, appErr
	}
	var failed *ChatStreamError
	for event := range events {
		switch event.Type {
		case "completed":
			if event.Result != nil {
				return *event.Result, nil
			}
		case "failed":
			failed = event.Error
		case "cancelled":
			return ChatTurnResult{}, BadRequest("CHAT_STREAM_CANCELLED", "chat stream was cancelled", "retry the message")
		}
	}
	if failed != nil {
		return ChatTurnResult{}, BadRequest(failed.Code, failed.Message, "check provider configuration and retry")
	}
	return ChatTurnResult{}, BadRequest("CHAT_STREAM_FAILED", "chat stream ended without a final result", "retry the message")
}

func (s *Store) StreamChatMessage(ctx context.Context, input SendChatMessageInput) (<-chan ChatStreamEvent, *AppError) {
	if input.SessionID == "" || (strings.TrimSpace(input.Content) == "" && strings.TrimSpace(input.RetryOfMessageID) == "") {
		return nil, BadRequest("BAD_REQUEST", "message content is required", "enter a message for the agent")
	}

	s.Mu.Lock()
	session, ok := s.Sessions[input.SessionID]
	if !ok {
		s.Mu.Unlock()
		return nil, NotFound("SESSION_NOT_FOUND", "session not found", "create a new chat session")
	}
	agent, ok := s.Agents[session.AgentID]
	if !ok || !agent.Enabled {
		s.Mu.Unlock()
		return nil, NotFound("AGENT_NOT_FOUND", "agent is unavailable", "select another agent")
	}
	profile, provider, providerFallback, bindErr := s.resolveChatModelBindingLocked(session, agent)
	if bindErr != nil {
		s.Mu.Unlock()
		return nil, bindErr
	}
	streamID := strings.TrimSpace(input.StreamID)
	if streamID == "" {
		streamID = s.NextIDLocked("stream")
	}
	traceID := s.TraceID()
	now := s.Now()
	userMessage, retryErr := s.resolveUserMessageForAttemptLocked(session, input, traceID, now)
	if retryErr != nil {
		s.Mu.Unlock()
		return nil, retryErr
	}
	attemptID := "attempt_" + s.NextIDLocked("msg")
	assistantMessage := ChatMessage{
		MessageID:      "assistant_" + s.NextIDLocked("msg"),
		AttemptID:      attemptID,
		SessionID:      session.SessionID,
		Role:           "assistant",
		Status:         "streaming",
		ProviderID:     provider.ProviderID,
		Model:          profile.Model,
		RuntimeSummary: buildRuntimeSummary(agent, session, provider, profile),
		TraceID:        traceID,
		CreatedAt:      now,
	}
	if strings.TrimSpace(input.RetryOfMessageID) == "" {
		s.Messages[session.SessionID] = append(s.Messages[session.SessionID], userMessage, assistantMessage)
	} else {
		s.Messages[session.SessionID] = append(s.Messages[session.SessionID], assistantMessage)
	}
	session.MessageCount = len(s.Messages[session.SessionID])
	session.UpdatedAt = now
	s.Sessions[session.SessionID] = session
	contextPack, warning := s.buildChatContextLocked(session, assistantMessage.MessageID, agent, profile, provider, providerFallback)
	toolCalls := s.buildChatToolCallsLocked(agent)
	streamCtx, cancel := context.WithCancel(ctx)
	s.Streams[streamID] = cancel
	s.Mu.Unlock()

	out := make(chan ChatStreamEvent, 32)
	go s.runChatStream(streamCtx, streamID, traceID, session, agent, profile, provider, assistantMessage.MessageID, attemptID, contextPack, warning, toolCalls, out)
	return out, nil
}

func (s *Store) CancelChatStream(input CancelChatStreamInput) (DeleteResult, *AppError) {
	if input.StreamID == "" {
		return DeleteResult{}, BadRequest("BAD_REQUEST", "missing streamId", "select an active stream")
	}
	s.Mu.Lock()
	cancel, ok := s.Streams[input.StreamID]
	if ok {
		delete(s.Streams, input.StreamID)
	}
	s.Mu.Unlock()
	if !ok {
		return DeleteResult{}, NotFound("STREAM_NOT_FOUND", "stream not found", "the stream may already be finished")
	}
	cancel()
	return DeleteResult{OK: true, DeletedID: input.StreamID}, nil
}

func (s *Store) runChatStream(
	ctx context.Context,
	streamID string,
	traceID string,
	session ChatSession,
	agent AgentConfig,
	profile ModelProfile,
	provider ModelProviderRecord,
	messageID string,
	attemptID string,
	contextPack ChatContextPack,
	warning *ChatStreamWarning,
	toolCalls []ChatToolCallPreview,
	out chan<- ChatStreamEvent,
) {
	defer close(out)
	defer func() {
		s.Mu.Lock()
		delete(s.Streams, streamID)
		s.Mu.Unlock()
	}()
	seq := 0
	emit := func(event ChatStreamEvent) {
		seq++
		event.StreamID = streamID
		event.SessionID = session.SessionID
		event.MessageID = messageID
		event.TraceID = traceID
		event.AttemptID = attemptID
		event.Sequence = seq
		event.Timestamp = s.Now()
		out <- event
	}

	runtimeSelection := buildRuntimeSelection(contextPack)
	emit(ChatStreamEvent{
		Type:             "started",
		ProviderID:       provider.ProviderID,
		Model:            profile.Model,
		ContextBudget:    &contextPack.Budget,
		RuntimeSelection: &runtimeSelection,
	})
	if warning != nil {
		event := ChatStreamEvent{Type: "context_compacted", Warning: warning, ContextBudget: &contextPack.Budget}
		if len(contextPack.Summaries) > 0 {
			event.ContextSummary = &contextPack.Summaries[0]
		}
		emit(event)
	}

	planStep := chatStep("step_plan", "PLAN", "Plan", "Bind session, agent, model profile and project context pack.", "completed", s.Now())
	emit(ChatStreamEvent{Type: "step", Step: &planStep})
	graphStep := chatStep("step_graph", "GRAPH", "Build task graph", "Create context, model and policy-gated tool tasks for this turn.", "completed", s.Now())
	emit(ChatStreamEvent{Type: "step", Step: &graphStep})
	executeStep := chatStep("step_execute", "EXECUTE", "Stream model response", "Call provider through ModelGateway streaming adapter.", "running", s.Now())
	emit(ChatStreamEvent{Type: "step", Step: &executeStep})
	for _, toolCall := range toolCalls {
		call := toolCall
		emit(ChatStreamEvent{Type: "tool_call_delta", ToolCall: &call})
	}

	var builder strings.Builder
	var usage *ChatModelUsage
	finishReason := ""
	startedAt := time.Now()
	modelMessages := append([]ChatGatewayMessage{}, contextPack.Messages...)
	for toolRound := 0; toolRound < 4; toolRound++ {
		toolResultMessages := []string{}
		finishReason = ""
		for chunk := range s.ModelGateway.StreamChat(ctx, toChatModelProvider(provider), toChatModelProfile(profile), modelMessages) {
			if ctx.Err() != nil {
				result := s.completeStreamMessage(session.SessionID, messageID, builder.String(), "cancelled", usage, "cancelled", latencyMS(startedAt), "", agent, provider, profile, toolCalls, contextPack)
				emit(ChatStreamEvent{Type: "cancelled", FinishReason: "cancelled", Result: &result, LatencyMS: latencyMS(startedAt)})
				return
			}
			if chunk.Error != nil {
				result := s.completeStreamMessage(session.SessionID, messageID, builder.String(), "failed", usage, "error", latencyMS(startedAt), chunk.Error.Code, agent, provider, profile, toolCalls, contextPack)
				emit(ChatStreamEvent{Type: "failed", Error: chunk.Error, FinishReason: "error", Result: &result, LatencyMS: latencyMS(startedAt)})
				return
			}
			if chunk.ToolCall != nil {
				call, result := s.handleChatToolCall(ctx, session, agent, traceID, *chunk.ToolCall)
				toolCalls = upsertToolCall(toolCalls, call)
				switch result.Status {
				case "blocked":
					emit(ChatStreamEvent{Type: "tool_blocked", ToolCall: &call, ToolResult: &result})
				default:
					running := call
					running.Status = "running"
					emit(ChatStreamEvent{Type: "tool_started", ToolCall: &running})
					emit(ChatStreamEvent{Type: "tool_result", ToolCall: &call, ToolResult: &result})
				}
				toolResultMessages = append(toolResultMessages, formatToolResultForModel(result))
				continue
			}
			if chunk.Delta != "" {
				builder.WriteString(chunk.Delta)
				emit(ChatStreamEvent{Type: "token_delta", Delta: chunk.Delta})
			}
			if chunk.ReasoningDelta != "" {
				emit(ChatStreamEvent{Type: "reasoning_delta", ReasoningDelta: chunk.ReasoningDelta})
			}
			if chunk.Usage != nil {
				usage = mergeUsage(usage, chunk.Usage)
				emit(ChatStreamEvent{Type: "usage", Usage: usage})
			}
			if chunk.FinishReason != "" {
				finishReason = chunk.FinishReason
			}
		}
		if len(toolResultMessages) == 0 || finishReason != "tool_calls" {
			break
		}
		if toolRound >= 2 {
			result := s.completeStreamMessage(session.SessionID, messageID, builder.String(), "failed", usage, "tool_loop_limit", latencyMS(startedAt), "TOOL_LOOP_LIMIT", agent, provider, profile, toolCalls, contextPack)
			emit(ChatStreamEvent{Type: "failed", Error: &ChatStreamError{Code: "TOOL_LOOP_LIMIT", Message: "tool loop limit reached", Recoverable: true}, FinishReason: "tool_loop_limit", Result: &result, LatencyMS: latencyMS(startedAt)})
			return
		}
		modelMessages = append(modelMessages, ChatGatewayMessage{
			Role:    "system",
			Content: "Tool execution results for the next model step:\n" + strings.Join(toolResultMessages, "\n"),
		})
	}
	if ctx.Err() != nil {
		result := s.completeStreamMessage(session.SessionID, messageID, builder.String(), "cancelled", usage, "cancelled", latencyMS(startedAt), "", agent, provider, profile, toolCalls, contextPack)
		emit(ChatStreamEvent{Type: "cancelled", FinishReason: "cancelled", Result: &result, LatencyMS: latencyMS(startedAt)})
		return
	}
	if finishReason == "" {
		finishReason = "stop"
	}
	observeStep := chatStep("step_observe", "OBSERVE", "Persist observation", "Persist assistant response, usage and runtime metadata.", "completed", s.Now())
	emit(ChatStreamEvent{Type: "step", Step: &observeStep})
	replanStep := chatStep("step_replan", "REPLAN", "Ready for steering", "Wait for user steering or project handoff.", "ready", s.Now())
	emit(ChatStreamEvent{Type: "step", Step: &replanStep})
	result := s.completeStreamMessage(session.SessionID, messageID, builder.String(), "completed", usage, finishReason, latencyMS(startedAt), "", agent, provider, profile, toolCalls, contextPack)
	emit(ChatStreamEvent{Type: "completed", Usage: usage, Result: &result, ProviderID: provider.ProviderID, Model: profile.Model, FinishReason: finishReason, LatencyMS: latencyMS(startedAt)})
}

func buildRuntimeSelection(contextPack ChatContextPack) ChatRuntimeSelection {
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
	return ChatRuntimeSelection{
		Summary:    strings.Join(summaryParts, " / "),
		Skills:     append([]SkillRuntimeDescriptor{}, contextPack.Skills...),
		Tools:      append([]ToolRuntimeDescriptor{}, contextPack.Tools...),
		MCPServers: append([]string{}, contextPack.MCPServers...),
	}
}

func (s *Store) buildModelMessagesLocked(session ChatSession, assistantMessageID string, agent AgentConfig) ([]ChatGatewayMessage, *ChatStreamWarning) {
	system := strings.TrimSpace(agent.SystemPrompt)
	if system == "" {
		system = "You are a DreamWorker agent."
	}
	system += "\n\nRuntime: AI OS + Agent Runtime + project incubation system. Keep answers clear and actionable."
	if session.ProjectID != nil {
		if project, ok := s.Projects[*session.ProjectID]; ok {
			system += fmt.Sprintf("\nProject: %s\nProject status: %s", project.Title, project.Status)
		}
	}
	if len(agent.EnabledSkills) > 0 {
		system += "\nEnabled skills: " + strings.Join(agent.EnabledSkills, ", ")
	}
	if len(agent.EnabledTools) > 0 {
		system += "\nAvailable tools are policy-gated previews: " + strings.Join(agent.EnabledTools, ", ")
	}
	if len(agent.EnabledMCPServers) > 0 {
		system += "\nMCP servers are policy-gated and never expose secrets."
	}
	result := []ChatGatewayMessage{{Role: "system", Content: system}}
	history := s.Messages[session.SessionID]
	start := 0
	warning := (*ChatStreamWarning)(nil)
	if len(history) > 24 {
		start = len(history) - 24
		warning = &ChatStreamWarning{Code: "CONTEXT_TRIMMED", Message: "Older chat messages were trimmed to fit the context window."}
	}
	for _, message := range history[start:] {
		if message.MessageID == assistantMessageID || strings.TrimSpace(message.Content) == "" {
			continue
		}
		if message.Role != "user" && message.Role != "assistant" {
			continue
		}
		if message.Role == "assistant" && message.Status != "completed" {
			continue
		}
		result = append(result, ChatGatewayMessage{Role: message.Role, Content: message.Content})
	}
	return result, warning
}

func (s *Store) completeStreamMessage(
	sessionID string,
	messageID string,
	content string,
	status string,
	usage *ChatModelUsage,
	finishReason string,
	latencyMS int,
	errorCode string,
	agent AgentConfig,
	provider ModelProviderRecord,
	profile ModelProfile,
	toolCalls []ChatToolCallPreview,
	contextPack ChatContextPack,
) ChatTurnResult {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	session := s.Sessions[sessionID]
	runtimeSummary := buildRuntimeSummary(agent, session, provider, profile)
	messages := s.Messages[sessionID]
	for index, message := range messages {
		if message.MessageID == messageID {
			message.Content = content
			message.Status = status
			message.ProviderID = provider.ProviderID
			message.Model = profile.Model
			message.Usage = usage
			message.FinishReason = finishReason
			message.RuntimeSummary = runtimeSummary
			messages[index] = message
			break
		}
	}
	now := s.Now()
	session.MessageCount = len(messages)
	session.UpdatedAt = now
	s.Messages[sessionID] = messages
	s.Sessions[sessionID] = session
	provider.LastStreamAt = &now
	provider.LatencyMS = latencyMS
	provider.StreamingVerified = status == "completed" || provider.StreamingVerified
	if errorCode != "" {
		provider.LastErrorCode = &errorCode
		provider.HealthStatus = "error"
		provider.Status = "error"
	} else if status == "completed" {
		provider.LastError = nil
		provider.LastErrorCode = nil
		provider.HealthStatus = "connected"
		provider.Status = "connected"
	}
	s.Providers[provider.ProviderID] = provider
	steps := buildCompletedChatExecutionSteps(now, agent)
	audit := ChatAuditSummary{
		ContentHash:  contentHash(content),
		ProviderID:   provider.ProviderID,
		Model:        profile.Model,
		LatencyMS:    latencyMS,
		ErrorCode:    errorCode,
		Usage:        usage,
		FinishReason: finishReason,
	}
	return ChatTurnResult{
		Session:         session,
		Messages:        append([]ChatMessage{}, messages...),
		ExecutionSteps:  steps,
		ToolCalls:       append([]ChatToolCallPreview{}, toolCalls...),
		ContextSummary:  firstContextSummary(contextPack),
		ContextBudget:   contextPack.Budget,
		AuditSummary:    audit,
		ProviderStatus:  provider.Safe().Status,
		RuntimeSnapshot: runtimeSummary,
		RuntimeSummary:  runtimeSummary,
	}
}

func (s *Store) resolveUserMessageForAttemptLocked(session ChatSession, input SendChatMessageInput, traceID string, now string) (ChatMessage, *AppError) {
	retryOf := strings.TrimSpace(input.RetryOfMessageID)
	if retryOf != "" {
		for _, message := range s.Messages[session.SessionID] {
			if message.MessageID == retryOf && message.Role == "user" {
				return message, nil
			}
		}
		return ChatMessage{}, NotFound("CHAT_MESSAGE_NOT_FOUND", "retry source message was not found", "send a new message")
	}
	return ChatMessage{
		MessageID: "user_" + s.NextIDLocked("msg"),
		SessionID: session.SessionID,
		Role:      "user",
		Content:   strings.TrimSpace(input.Content),
		Status:    "completed",
		TraceID:   traceID,
		CreatedAt: now,
	}, nil
}

func (s *Store) DeleteChatSession(sessionID string) (DeleteResult, *AppError) {
	if sessionID == "" {
		return DeleteResult{}, BadRequest("BAD_REQUEST", "missing sessionId", "select a chat session")
	}
	s.Mu.Lock()
	defer s.Mu.Unlock()
	delete(s.Sessions, sessionID)
	delete(s.Messages, sessionID)
	return DeleteResult{OK: true, DeletedID: sessionID}, nil
}

func buildCompletedChatExecutionSteps(timestamp string, agent AgentConfig) []ChatExecutionStep {
	strategy := fallback(agent.Planner.Strategy, "plan-execute")
	return []ChatExecutionStep{
		chatStep("step_plan", "PLAN", "Plan", "Bound agent, model profile and project context.", "completed", timestamp),
		chatStep("step_graph", "GRAPH", "Task graph", fmt.Sprintf("Created observable graph with %s strategy.", strategy), "completed", timestamp),
		chatStep("step_execute", "EXECUTE", "Model stream", "Streamed provider response through the Engine.", "completed", timestamp),
		chatStep("step_observe", "OBSERVE", "Persist result", "Recorded final response, usage and trace metadata.", "completed", timestamp),
		chatStep("step_replan", "REPLAN", "Ready", "Ready for steering, retry or project handoff.", "ready", timestamp),
	}
}

func chatStep(stepID string, phase string, title string, summary string, status string, timestamp string) ChatExecutionStep {
	completedAt := ""
	if status == "completed" || status == "ready" {
		completedAt = timestamp
	}
	return ChatExecutionStep{
		StepID:      stepID,
		Phase:       phase,
		Title:       title,
		Summary:     summary,
		Status:      status,
		StartedAt:   timestamp,
		CompletedAt: completedAt,
	}
}

func (s *Store) buildChatToolCallsLocked(agent AgentConfig) []ChatToolCallPreview {
	result := make([]ChatToolCallPreview, 0, len(agent.EnabledTools))
	for _, toolID := range agent.EnabledTools {
		tool := s.Tools[toolID]
		risk := fallback(tool.RiskLevel, "low")
		name := fallback(tool.DisplayName, toolID)
		result = append(result, ChatToolCallPreview{
			CallID:           "call_" + toolID,
			ToolID:           toolID,
			DisplayName:      name,
			RiskLevel:        risk,
			ApprovalRequired: risk == "high" || risk == "critical",
			Status:           "preview",
			Summary:          "Tool is available to the model. Low-risk calls can execute through the Engine; side-effecting calls are blocked for approval.",
		})
	}
	return result
}

func buildRuntimeSummary(agent AgentConfig, session ChatSession, provider ModelProviderRecord, profile ModelProfile) string {
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

func mergeUsage(current *ChatModelUsage, next *ChatModelUsage) *ChatModelUsage {
	if next == nil {
		return current
	}
	if current == nil {
		copy := *next
		if copy.TotalTokens == 0 {
			copy.TotalTokens = copy.InputTokens + copy.OutputTokens
		}
		return &copy
	}
	if next.InputTokens > 0 {
		current.InputTokens = next.InputTokens
	}
	if next.OutputTokens > 0 {
		current.OutputTokens = next.OutputTokens
	}
	if next.TotalTokens > 0 {
		current.TotalTokens = next.TotalTokens
	} else {
		current.TotalTokens = current.InputTokens + current.OutputTokens
	}
	if next.CostUSD > 0 {
		current.CostUSD = next.CostUSD
	}
	return current
}

func latencyMS(startedAt time.Time) int {
	return int(time.Since(startedAt).Milliseconds())
}

func contentHash(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}
