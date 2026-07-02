package chat

import (
	"context"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/agentruntime"
	runtimetools "github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/tools"
	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/ports"
)

func (s *Store) resolveChatModelBindingLocked(
	session ChatSession,
	agent AgentConfig,
) (ModelProfile, ModelProviderRecord, string, *AppError) {
	return agentruntime.ResolveChatModelBinding(s.Store, session, agent)
}

func (s *Store) buildChatContextLocked(
	session ChatSession,
	assistantMessageID string,
	agent AgentConfig,
	profile ModelProfile,
	provider ModelProviderRecord,
	providerFallback string,
) (ChatContextPack, *ChatStreamWarning) {
	return agentruntime.BuildChatContext(s.Store, session, assistantMessageID, agent, profile, provider, providerFallback)
}

func (s *Store) handleChatToolCall(
	ctx context.Context,
	session ChatSession,
	agent AgentConfig,
	traceID string,
	request ToolExecutionRequest,
) (ChatToolCallPreview, ToolExecutionResult) {
	return runtimetools.NewRegistry(s.Store).HandleChatToolCall(ctx, session, agent, traceID, request)
}

func toChatModelProvider(provider ModelProviderRecord) ports.ChatModelProvider {
	return agentruntime.ToChatModelProvider(provider)
}

func toChatModelProfile(profile ModelProfile) ports.ChatModelProfile {
	return agentruntime.ToChatModelProfile(profile)
}

func upsertToolCall(calls []ChatToolCallPreview, call ChatToolCallPreview) []ChatToolCallPreview {
	return runtimetools.UpsertToolCall(calls, call)
}

func formatToolResultForModel(result ToolExecutionResult) string {
	return runtimetools.FormatToolResultForModel(result)
}

func firstContextSummary(pack ChatContextPack) *ChatContextSummary {
	return agentruntime.FirstContextSummary(pack)
}
