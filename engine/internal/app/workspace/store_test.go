package workspace

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/ports"
)

func newTestStore() *Store {
	return NewStore(
		WithClock(func() string { return "2026-07-01T00:00:00Z" }),
		WithTraceID(func() string { return "tr_test" }),
	)
}

type testGateway struct {
	chunks []ModelStreamChunk
}

func (gateway testGateway) DiscoverModels(_ context.Context, provider ports.ChatModelProvider) ProviderModelDiscoveryResult {
	return ProviderModelDiscoveryResult{Models: append([]string{}, provider.AvailableModels...), Discovered: true}
}

func (gateway testGateway) HealthCheck(_ context.Context, _ ports.ChatModelProvider) ProviderHealth {
	return ProviderHealth{OK: true, Status: "connected", Message: "test gateway ready", StreamingVerified: true}
}

func (gateway testGateway) StreamChat(_ context.Context, _ ports.ChatModelProvider, _ ports.ChatModelProfile, _ []ChatGatewayMessage) <-chan ModelStreamChunk {
	out := make(chan ModelStreamChunk, len(gateway.chunks))
	go func() {
		defer close(out)
		for _, chunk := range gateway.chunks {
			out <- chunk
		}
	}()
	return out
}

func TestProvidersNeverExposeRawAPIKey(t *testing.T) {
	store := newTestStore()

	provider, appErr := store.SaveProvider(SaveModelProviderInput{
		ProviderID:      "provider_custom",
		ProviderType:    ProviderDeepSeek,
		DisplayName:     "Test Provider",
		BaseURL:         "https://api.example.com",
		DefaultModel:    "deepseek-chat",
		AvailableModels: []string{"deepseek-chat"},
		Enabled:         true,
		APIKey:          "sk-test-secret",
	})
	if appErr != nil {
		t.Fatalf("save provider: %v", appErr)
	}

	payload, err := json.Marshal(provider)
	if err != nil {
		t.Fatalf("marshal provider: %v", err)
	}
	if strings.Contains(string(payload), "sk-test-secret") {
		t.Fatalf("safe provider leaked raw api key: %s", payload)
	}
	if provider.MaskedKey == nil || *provider.MaskedKey != "sk-t...cret" {
		t.Fatalf("expected masked key, got %#v", provider.MaskedKey)
	}
}

func TestMCPServerMasksSecrets(t *testing.T) {
	store := newTestStore()

	server, appErr := store.SaveMCPServer(SaveMCPServerInput{
		ServerID:    "mcp_custom",
		DisplayName: "Remote MCP",
		Command:     "dreamworker-mcp",
		TrustLevel:  "remote_untrusted",
		Enabled:     true,
		Secrets: map[string]string{
			"MCP_TOKEN": "token-secret-value",
		},
	})
	if appErr != nil {
		t.Fatalf("save mcp server: %v", appErr)
	}

	payload, err := json.Marshal(server)
	if err != nil {
		t.Fatalf("marshal server: %v", err)
	}
	if strings.Contains(string(payload), "token-secret-value") {
		t.Fatalf("mcp server leaked raw secret: %s", payload)
	}
	if len(server.EnvKeys) != 1 || server.EnvKeys[0] != "MCP_TOKEN" {
		t.Fatalf("expected env key summary, got %#v", server.EnvKeys)
	}
	if len(server.MaskedSecrets) != 1 || server.MaskedSecrets[0] != "MCP_TOKEN=toke...alue" {
		t.Fatalf("expected masked secret, got %#v", server.MaskedSecrets)
	}
}

func TestProjectModulesCarryProjectID(t *testing.T) {
	store := newTestStore()
	project, appErr := store.CreateProject(CreateProjectInput{
		Title:       "Test project",
		Description: "module isolation",
	})
	if appErr != nil {
		t.Fatalf("create project: %v", appErr)
	}

	modules, appErr := store.ListProjectModules(project.ProjectID)
	if appErr != nil {
		t.Fatalf("list modules: %v", appErr)
	}
	if len(modules) != 4 {
		t.Fatalf("expected four modules, got %d", len(modules))
	}
	for _, module := range modules {
		if module.ProjectID != project.ProjectID {
			t.Fatalf("module %s has project id %q, want %q", module.ModuleID, module.ProjectID, project.ProjectID)
		}
		if len(module.Submodules) != 4 {
			t.Fatalf("module %s expected four submodules, got %d", module.ModuleID, len(module.Submodules))
		}
		for _, submodule := range module.Submodules {
			if submodule.ProjectID != project.ProjectID {
				t.Fatalf("submodule %s has project id %q, want %q", submodule.SubmoduleID, submodule.ProjectID, project.ProjectID)
			}
			if submodule.ModuleID != module.ModuleID {
				t.Fatalf("submodule %s has module id %q, want %q", submodule.SubmoduleID, submodule.ModuleID, module.ModuleID)
			}
		}
	}
}

func TestDeleteProjectClearsModulesAndChatProjectRefs(t *testing.T) {
	store := newTestStore()
	project, appErr := store.CreateProject(CreateProjectInput{
		Title:       "Delete me",
		Description: "delete boundary",
	})
	if appErr != nil {
		t.Fatalf("create project: %v", appErr)
	}
	session, appErr := store.CreateChatSession(CreateChatSessionInput{
		ProjectID:      &project.ProjectID,
		Title:          "Project chat",
		AgentID:        "agent_general_assistant",
		ModelProfileID: "profile_fast",
	})
	if appErr != nil {
		t.Fatalf("create chat session: %v", appErr)
	}

	result, appErr := store.DeleteProject(project.ProjectID)
	if appErr != nil {
		t.Fatalf("delete project: %v", appErr)
	}
	if !result.OK || result.DeletedID != project.ProjectID {
		t.Fatalf("unexpected delete result: %#v", result)
	}
	if _, appErr := store.GetProject(project.ProjectID); appErr == nil {
		t.Fatalf("expected deleted project to be missing")
	}
	if _, appErr := store.ListProjectModules(project.ProjectID); appErr == nil {
		t.Fatalf("expected deleted project modules to be missing")
	}
	updatedSession := store.sessions[session.SessionID]
	if updatedSession.ProjectID != nil {
		t.Fatalf("expected chat project ref to be cleared, got %#v", updatedSession.ProjectID)
	}
}

func TestChatSendMessageAggregatesLocalStreamingTurn(t *testing.T) {
	store := newTestStore()
	session, appErr := store.CreateChatSession(CreateChatSessionInput{
		Title:          "Test chat",
		AgentID:        "agent_evaluator",
		ModelProfileID: "profile_stub",
	})
	if appErr != nil {
		t.Fatalf("create session: %v", appErr)
	}

	turn, appErr := store.SendChatMessage(SendChatMessageInput{
		SessionID: session.SessionID,
		Content:   "hello",
	})
	if appErr != nil {
		t.Fatalf("send message: %v", appErr)
	}
	if turn.Session.MessageCount != 2 {
		t.Fatalf("expected two messages, got %d", turn.Session.MessageCount)
	}
	if turn.Messages[0].TraceID != "tr_test" || turn.Messages[1].TraceID != "tr_test" {
		t.Fatalf("expected trace id to propagate, got %#v", turn.Messages)
	}
	if !strings.Contains(turn.Messages[1].Content, "Local streaming model received") {
		t.Fatalf("expected local streaming response, got %q", turn.Messages[1].Content)
	}
	if len(turn.ExecutionSteps) != 5 {
		t.Fatalf("expected five execution steps, got %#v", turn.ExecutionSteps)
	}
	if turn.ExecutionSteps[0].Phase != "PLAN" || turn.ExecutionSteps[4].Phase != "REPLAN" {
		t.Fatalf("expected plan to replan phases, got %#v", turn.ExecutionSteps)
	}
	if len(turn.ToolCalls) == 0 || turn.ToolCalls[0].Status != "preview" {
		t.Fatalf("expected tool call previews, got %#v", turn.ToolCalls)
	}
	if !strings.Contains(turn.RuntimeSummary, "Planner=plan-execute") {
		t.Fatalf("expected runtime summary, got %q", turn.RuntimeSummary)
	}
}

func TestChatStreamEmitsTokenDeltasAndCompletion(t *testing.T) {
	store := NewStore(
		WithClock(func() string { return "2026-07-01T00:00:00Z" }),
		WithTraceID(func() string { return "tr_test" }),
		WithModelGateway(testGateway{chunks: []ModelStreamChunk{
			{ReasoningDelta: "先识别上下文"},
			{Delta: "stream ok", Usage: &ChatModelUsage{InputTokens: 3, OutputTokens: 2, TotalTokens: 5}, FinishReason: "stop"},
		}}),
	)
	session, appErr := store.CreateChatSession(CreateChatSessionInput{
		Title:          "stream",
		AgentID:        "agent_evaluator",
		ModelProfileID: "profile_stub",
	})
	if appErr != nil {
		t.Fatalf("create session: %v", appErr)
	}

	events, appErr := store.StreamChatMessage(context.Background(), SendChatMessageInput{
		SessionID: session.SessionID,
		Content:   "hello",
		StreamID:  "stream_test",
	})
	if appErr != nil {
		t.Fatalf("stream message: %v", appErr)
	}
	var started bool
	var deltaCount int
	var reasoning string
	var runtimeSelection *ChatRuntimeSelection
	var completed *ChatTurnResult
	for event := range events {
		if event.StreamID != "stream_test" || event.TraceID != "tr_test" {
			t.Fatalf("unexpected event identity: %#v", event)
		}
		switch event.Type {
		case "started":
			started = true
			runtimeSelection = event.RuntimeSelection
		case "reasoning_delta":
			reasoning += event.ReasoningDelta
		case "token_delta":
			deltaCount++
		case "completed":
			completed = event.Result
		}
	}
	if !started || deltaCount == 0 || completed == nil {
		t.Fatalf("expected started, deltas and completion, started=%v deltas=%d completed=%#v", started, deltaCount, completed)
	}
	if reasoning == "" {
		t.Fatalf("expected reasoning delta")
	}
	if runtimeSelection == nil || len(runtimeSelection.Skills) == 0 || len(runtimeSelection.Tools) == 0 {
		t.Fatalf("expected skill and tool runtime selection, got %#v", runtimeSelection)
	}
	if completed.Session.MessageCount != 2 {
		t.Fatalf("expected persisted user and assistant messages, got %#v", completed.Session)
	}
}

func TestChatRetryCreatesAssistantAttemptWithoutDuplicatingUserMessage(t *testing.T) {
	store := NewStore(
		WithClock(func() string { return "2026-07-01T00:00:00Z" }),
		WithTraceID(func() string { return "tr_test" }),
		WithModelGateway(testGateway{chunks: []ModelStreamChunk{
			{Delta: "retry ok", FinishReason: "stop"},
		}}),
	)
	session, appErr := store.CreateChatSession(CreateChatSessionInput{
		Title:          "retry",
		AgentID:        "agent_evaluator",
		ModelProfileID: "profile_stub",
	})
	if appErr != nil {
		t.Fatalf("create session: %v", appErr)
	}
	first, appErr := store.SendChatMessage(SendChatMessageInput{
		SessionID: session.SessionID,
		Content:   "hello",
	})
	if appErr != nil {
		t.Fatalf("send message: %v", appErr)
	}
	retried, appErr := store.SendChatMessage(SendChatMessageInput{
		SessionID:        session.SessionID,
		RetryOfMessageID: first.Messages[0].MessageID,
	})
	if appErr != nil {
		t.Fatalf("retry message: %v", appErr)
	}
	userCount := 0
	for _, message := range retried.Messages {
		if message.Role == "user" {
			userCount++
		}
	}
	if userCount != 1 || retried.Session.MessageCount != 3 {
		t.Fatalf("expected one user and two attempts, users=%d messages=%d", userCount, retried.Session.MessageCount)
	}
	if retried.Messages[2].AttemptID == "" || retried.Messages[2].Status != "completed" {
		t.Fatalf("expected completed assistant attempt, got %#v", retried.Messages[2])
	}
}

func TestChatStreamFailureKeepsPartialAssistantAttempt(t *testing.T) {
	store := NewStore(
		WithClock(func() string { return "2026-07-01T00:00:00Z" }),
		WithTraceID(func() string { return "tr_test" }),
		WithModelGateway(testGateway{chunks: []ModelStreamChunk{
			{Delta: "partial"},
			{Error: &ChatStreamError{Code: "RATE_LIMIT", Message: "rate limited", Recoverable: true}},
		}}),
	)
	session, appErr := store.CreateChatSession(CreateChatSessionInput{
		Title:          "failed",
		AgentID:        "agent_evaluator",
		ModelProfileID: "profile_stub",
	})
	if appErr != nil {
		t.Fatalf("create session: %v", appErr)
	}
	events, appErr := store.StreamChatMessage(context.Background(), SendChatMessageInput{
		SessionID: session.SessionID,
		Content:   "hello",
	})
	if appErr != nil {
		t.Fatalf("stream message: %v", appErr)
	}
	var failed *ChatTurnResult
	for event := range events {
		if event.Type == "failed" {
			failed = event.Result
		}
	}
	if failed == nil {
		t.Fatalf("expected failed event result")
	}
	assistant := failed.Messages[len(failed.Messages)-1]
	if assistant.Status != "failed" || assistant.Content != "partial" || assistant.AttemptID == "" {
		t.Fatalf("expected failed partial attempt, got %#v", assistant)
	}
	if failed.AuditSummary.ErrorCode != "RATE_LIMIT" {
		t.Fatalf("expected audit error code, got %#v", failed.AuditSummary)
	}
}

func TestChatStreamConsumesInjectedModelGateway(t *testing.T) {
	store := NewStore(
		WithClock(func() string { return "2026-07-01T00:00:00Z" }),
		WithTraceID(func() string { return "tr_test" }),
		WithModelGateway(testGateway{chunks: []ModelStreamChunk{
			{Delta: "Hello"},
			{Delta: " world", Usage: &ChatModelUsage{InputTokens: 3, OutputTokens: 2, TotalTokens: 5}, FinishReason: "stop"},
		}}),
	)
	if _, appErr := store.SaveProvider(SaveModelProviderInput{
		ProviderID:      "provider_real",
		ProviderType:    ProviderOpenAICompatible,
		DisplayName:     "Real compatible",
		BaseURL:         "https://api.example.com/v1",
		DefaultModel:    "test-model",
		AvailableModels: []string{"test-model"},
		Enabled:         true,
		Capabilities:    []string{"chat", "tools"},
		APIKey:          "sk-real-test",
	}); appErr != nil {
		t.Fatalf("save provider: %v", appErr)
	}
	if _, appErr := store.SaveProfile(ModelProfile{
		ProfileID:   "profile_real",
		DisplayName: "Real stream",
		ProviderID:  "provider_real",
		Model:       "test-model",
		Temperature: 0,
		MaxTokens:   128,
		Purpose:     "test",
		Enabled:     true,
	}); appErr != nil {
		t.Fatalf("save profile: %v", appErr)
	}
	session, appErr := store.CreateChatSession(CreateChatSessionInput{
		Title:          "real stream",
		AgentID:        "agent_evaluator",
		ModelProfileID: "profile_real",
	})
	if appErr != nil {
		t.Fatalf("create session: %v", appErr)
	}
	turn, appErr := store.SendChatMessage(SendChatMessageInput{
		SessionID: session.SessionID,
		Content:   "hello",
	})
	if appErr != nil {
		t.Fatalf("send message: %v", appErr)
	}
	if got := turn.Messages[1].Content; got != "Hello world" {
		t.Fatalf("expected streamed content, got %q", got)
	}
	if turn.Messages[1].Usage == nil || turn.Messages[1].Usage.TotalTokens != 5 {
		t.Fatalf("expected usage from SSE, got %#v", turn.Messages[1].Usage)
	}
}

func TestUpdateChatSessionPersistsRuntimeBindings(t *testing.T) {
	store := newTestStore()
	project, appErr := store.CreateProject(CreateProjectInput{
		Title:       "Binding test",
		Description: "chat binding test",
	})
	if appErr != nil {
		t.Fatalf("create project: %v", appErr)
	}
	session, appErr := store.CreateChatSession(CreateChatSessionInput{
		Title:          "Before",
		AgentID:        "agent_general_assistant",
		ModelProfileID: "profile_fast",
	})
	if appErr != nil {
		t.Fatalf("create session: %v", appErr)
	}

	updated, appErr := store.UpdateChatSession(UpdateChatSessionInput{
		SessionID:      session.SessionID,
		ProjectID:      &project.ProjectID,
		Title:          "After",
		AgentID:        "agent_evaluator",
		ModelProfileID: "profile_stub",
	})
	if appErr != nil {
		t.Fatalf("update session: %v", appErr)
	}
	if updated.ProjectID == nil || *updated.ProjectID != project.ProjectID {
		t.Fatalf("expected project binding, got %#v", updated.ProjectID)
	}
	if updated.AgentID != "agent_evaluator" || updated.ModelProfileID != "profile_stub" {
		t.Fatalf("expected agent/model binding, got %#v", updated)
	}
	if updated.Title != "After" {
		t.Fatalf("expected title update, got %q", updated.Title)
	}
}

func TestProviderRefreshModelsPreservesCatalogOnRealProviderFailure(t *testing.T) {
	store := newTestStore()

	provider, appErr := store.RefreshProviderModels("provider_deepseek")
	if appErr != nil {
		t.Fatalf("refresh provider models: %v", appErr)
	}
	if provider.Status != "error" {
		t.Fatalf("expected missing-key provider refresh to mark error, got %s", provider.Status)
	}
	if len(provider.AvailableModels) == 0 || provider.AvailableModels[0] != "deepseek-v4-flash" {
		t.Fatalf("expected existing deepseek models to be preserved, got %#v", provider.AvailableModels)
	}
	if len(provider.Capabilities) == 0 || provider.Capabilities[0] != "chat" {
		t.Fatalf("expected provider capabilities, got %#v", provider.Capabilities)
	}
}

func TestChatContextCompactsLongHistory(t *testing.T) {
	store := newTestStore()
	if _, appErr := store.SaveProfile(ModelProfile{
		ProfileID:      "profile_tiny_context",
		DisplayName:    "Tiny context",
		ProviderID:     "provider_local_stub",
		Model:          "model_generate_stub",
		Temperature:    0,
		MaxTokens:      128,
		ContextWindow:  768,
		ResponseFormat: "text",
		ToolMode:       "none",
		TimeoutMS:      30000,
		Purpose:        "test compaction",
		Enabled:        true,
	}); appErr != nil {
		t.Fatalf("save profile: %v", appErr)
	}
	session, appErr := store.CreateChatSession(CreateChatSessionInput{
		Title:          "compact",
		AgentID:        "agent_evaluator",
		ModelProfileID: "profile_tiny_context",
	})
	if appErr != nil {
		t.Fatalf("create session: %v", appErr)
	}
	store.mu.Lock()
	for index := 0; index < 36; index++ {
		store.messages[session.SessionID] = append(store.messages[session.SessionID], ChatMessage{
			MessageID: "hist_" + string(rune('a'+index%26)),
			SessionID: session.SessionID,
			Role:      "user",
			Content:   strings.Repeat("context message ", 16),
			Status:    "completed",
			CreatedAt: "2026-07-01T00:00:00Z",
		})
	}
	store.mu.Unlock()

	events, appErr := store.StreamChatMessage(context.Background(), SendChatMessageInput{
		SessionID: session.SessionID,
		Content:   "final question",
	})
	if appErr != nil {
		t.Fatalf("stream: %v", appErr)
	}
	var compacted bool
	var completed *ChatTurnResult
	for event := range events {
		if event.Type == "context_compacted" {
			compacted = true
			if event.ContextBudget == nil || !event.ContextBudget.Compacted {
				t.Fatalf("expected compacted budget, got %#v", event.ContextBudget)
			}
		}
		if event.Type == "completed" {
			completed = event.Result
		}
	}
	if !compacted {
		t.Fatalf("expected context_compacted event")
	}
	if completed == nil || completed.ContextSummary == nil || !completed.ContextBudget.Compacted {
		t.Fatalf("expected completed context summary, got %#v", completed)
	}
}

func TestChatStreamExecutesLowRiskToolCall(t *testing.T) {
	store := NewStore(
		WithClock(func() string { return "2026-07-01T00:00:00Z" }),
		WithTraceID(func() string { return "tr_test" }),
		WithModelGateway(testGateway{chunks: []ModelStreamChunk{
			{ToolCall: &ToolExecutionRequest{CallID: "call_model", ToolID: "tool_model_generate_stub", Arguments: `{"prompt":"hello"}`}},
			{Delta: "done", FinishReason: "stop"},
		}}),
	)
	session, appErr := store.CreateChatSession(CreateChatSessionInput{
		Title:          "tool",
		AgentID:        "agent_evaluator",
		ModelProfileID: "profile_stub",
	})
	if appErr != nil {
		t.Fatalf("create session: %v", appErr)
	}
	events, appErr := store.StreamChatMessage(context.Background(), SendChatMessageInput{
		SessionID: session.SessionID,
		Content:   "use tool",
	})
	if appErr != nil {
		t.Fatalf("stream: %v", appErr)
	}
	var sawStarted bool
	var sawResult bool
	var completed *ChatTurnResult
	for event := range events {
		if event.Type == "tool_started" {
			sawStarted = true
		}
		if event.Type == "tool_result" && event.ToolResult != nil && event.ToolResult.Status == "completed" {
			sawResult = true
		}
		if event.Type == "completed" {
			completed = event.Result
		}
	}
	if !sawStarted || !sawResult {
		t.Fatalf("expected tool_started and tool_result, started=%v result=%v", sawStarted, sawResult)
	}
	if completed == nil || len(completed.ToolCalls) == 0 || completed.ToolCalls[len(completed.ToolCalls)-1].Status != "completed" {
		t.Fatalf("expected completed tool call in result, got %#v", completed)
	}
}

func TestChatStreamBlocksMediumRiskToolCall(t *testing.T) {
	store := NewStore(
		WithClock(func() string { return "2026-07-01T00:00:00Z" }),
		WithTraceID(func() string { return "tr_test" }),
		WithModelGateway(testGateway{chunks: []ModelStreamChunk{
			{ToolCall: &ToolExecutionRequest{CallID: "call_write", ToolID: "tool_artifact_write", Arguments: `{"content":"x"}`}},
			{Delta: "blocked", FinishReason: "stop"},
		}}),
	)
	session, appErr := store.CreateChatSession(CreateChatSessionInput{
		Title:          "tool block",
		AgentID:        "agent_opportunity_scout",
		ModelProfileID: "profile_stub",
	})
	if appErr != nil {
		t.Fatalf("create session: %v", appErr)
	}
	events, appErr := store.StreamChatMessage(context.Background(), SendChatMessageInput{
		SessionID: session.SessionID,
		Content:   "write artifact",
	})
	if appErr != nil {
		t.Fatalf("stream: %v", appErr)
	}
	var blocked *ToolExecutionResult
	for event := range events {
		if event.Type == "tool_blocked" {
			blocked = event.ToolResult
		}
	}
	if blocked == nil || blocked.ErrorCode != "APPROVAL_REQUIRED" {
		t.Fatalf("expected approval block, got %#v", blocked)
	}
}

func TestDefaultProfileFallsBackToOfflineStubWithoutDemoKey(t *testing.T) {
	t.Setenv("DEEPSEEK_API_KEY", "")
	store := newTestStore()
	session, appErr := store.CreateChatSession(CreateChatSessionInput{
		Title:          "fallback",
		AgentID:        "agent_general_assistant",
		ModelProfileID: "profile_fast",
	})
	if appErr != nil {
		t.Fatalf("create session: %v", appErr)
	}
	turn, appErr := store.SendChatMessage(SendChatMessageInput{
		SessionID: session.SessionID,
		Content:   "hello",
	})
	if appErr != nil {
		t.Fatalf("send: %v", appErr)
	}
	assistant := turn.Messages[len(turn.Messages)-1]
	if assistant.ProviderID != "provider_local_stub" || assistant.Model != "model_generate_stub" {
		t.Fatalf("expected offline stub fallback, got provider=%s model=%s", assistant.ProviderID, assistant.Model)
	}
}

func TestRefreshMCPToolsDiscoversStdioTools(t *testing.T) {
	if os.Getenv("GO_WANT_MCP_HELPER_PROCESS") == "1" {
		runMCPHelperProcess()
		return
	}
	store := newTestStore()
	server, appErr := store.SaveMCPServer(SaveMCPServerInput{
		ServerID:    "mcp_fake",
		DisplayName: "Fake MCP",
		Command:     os.Args[0],
		Args:        []string{"-test.run=TestRefreshMCPToolsDiscoversStdioTools"},
		TrustLevel:  "trusted_builtin",
		Enabled:     true,
		Secrets: map[string]string{
			"GO_WANT_MCP_HELPER_PROCESS": "1",
		},
	})
	if appErr != nil {
		t.Fatalf("save mcp: %v", appErr)
	}
	if !server.HasSecrets || len(server.MaskedSecrets) != 1 {
		t.Fatalf("expected masked helper env, got %#v", server)
	}
	tools, appErr := store.RefreshMCPTools("mcp_fake")
	if appErr != nil {
		t.Fatalf("refresh tools: %v", appErr)
	}
	if len(tools) != 1 || tools[0].ToolID != "mcp_mcp_fake_echo" {
		t.Fatalf("expected discovered echo tool, got %#v", tools)
	}
}

func runMCPHelperProcess() {
	reader := bufio.NewReader(os.Stdin)
	for {
		payload, err := readMCPMessage(reader)
		if err != nil {
			os.Exit(0)
		}
		var request struct {
			ID     int    `json:"id"`
			Method string `json:"method"`
		}
		_ = json.Unmarshal(payload, &request)
		if request.ID == 0 {
			continue
		}
		switch request.Method {
		case "initialize":
			_ = writeMCPMessage(os.Stdout, map[string]any{
				"jsonrpc": "2.0",
				"id":      request.ID,
				"result":  map[string]any{"protocolVersion": "2024-11-05", "capabilities": map[string]any{}},
			})
		case "tools/list":
			_ = writeMCPMessage(os.Stdout, map[string]any{
				"jsonrpc": "2.0",
				"id":      request.ID,
				"result": map[string]any{
					"tools": []map[string]any{{"name": "echo", "description": "Echo tool"}},
				},
			})
		default:
			_ = writeMCPMessage(os.Stdout, map[string]any{
				"jsonrpc": "2.0",
				"id":      request.ID,
				"result": map[string]any{
					"content": []map[string]any{{"type": "text", "text": "ok"}},
				},
			})
		}
	}
}

func TestSeedAgentsExposeRuntimeConfig(t *testing.T) {
	store := newTestStore()
	agent, appErr := store.GetAgent("agent_general_assistant")
	if appErr != nil {
		t.Fatalf("get agent: %v", appErr)
	}
	if !agent.Planner.Enabled || agent.Planner.Strategy != "plan-execute" {
		t.Fatalf("expected planner config, got %#v", agent.Planner)
	}
	if agent.RuntimeConfig.ContextWindow == 0 || agent.Executor.TimeoutMS == 0 || agent.MemoryScope == "" {
		t.Fatalf("expected runtime defaults, got %#v", agent)
	}
}

func TestAgentSkillsLoadFromAgentDirectory(t *testing.T) {
	agentDir := t.TempDir()
	skillDir := filepath.Join(agentDir, "skills", "skillcreator")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("mkdir skill: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(`---
name: Skill Creator
description: Creates reusable skills in standard SKILL.md format.
when_to_use: Use when the user asks to create or install a skill.
allowed-tools: artifact_write, human_question
category: general
version: 0.1.0
output-artifacts: SKILL.md
dreamworker-built-in: true
---

## Instructions

Create a complete skill file.
`), 0o644); err != nil {
		t.Fatalf("write skill: %v", err)
	}
	store := NewStore(
		WithAgentDir(agentDir),
		WithClock(func() string { return "2026-07-01T00:00:00Z" }),
		WithTraceID(func() string { return "tr_test" }),
	)
	skill, appErr := store.GetSkill("skill_skillcreator")
	if appErr != nil {
		t.Fatalf("get skillcreator: %v", appErr)
	}
	if skill.CommandName != "skillcreator" || !strings.Contains(skill.Instructions, "complete skill") {
		t.Fatalf("expected loaded skillcreator metadata, got %#v", skill)
	}
	if len(skill.RequiredCapabilities) != 2 || skill.RequiredCapabilities[0] != "cap_artifact_write" {
		t.Fatalf("expected mapped capabilities, got %#v", skill.RequiredCapabilities)
	}
}

func TestSaveSkillWritesAgentSkillFile(t *testing.T) {
	agentDir := t.TempDir()
	store := NewStore(
		WithAgentDir(agentDir),
		WithClock(func() string { return "2026-07-01T00:00:00Z" }),
		WithTraceID(func() string { return "tr_test" }),
	)
	skill, appErr := store.SaveSkill(SkillConfig{
		SkillID:              "skill_research_brief",
		DisplayName:          "Research Brief",
		Description:          "Creates a research brief.",
		WhenToUse:            "Use when research must become a reusable brief.",
		Instructions:         "## Instructions\n\nWrite a brief.",
		Category:             "explore",
		Version:              "0.1.0",
		RequiredCapabilities: []string{"cap_artifact_write", "cap_human_input"},
		OutputArtifacts:      []string{"research_brief.md"},
	})
	if appErr != nil {
		t.Fatalf("save skill: %v", appErr)
	}
	if skill.CommandName != "research-brief" {
		t.Fatalf("expected command from skill id, got %#v", skill)
	}
	data, err := os.ReadFile(filepath.Join(agentDir, "skills", "research-brief", "SKILL.md"))
	if err != nil {
		t.Fatalf("read written skill: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "allowed-tools: artifact_write, human_question") ||
		!strings.Contains(content, "dreamworker-built-in: false") {
		t.Fatalf("unexpected skill file content: %s", content)
	}
}
