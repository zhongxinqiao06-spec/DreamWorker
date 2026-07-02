package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/ports"
)

type ChatGatewayMessage = ports.ChatModelMessage
type ModelStreamChunk = ports.ChatModelStreamChunk
type ProviderHealth = ports.ProviderHealth
type ProviderModelDiscoveryResult = ports.ProviderModelDiscoveryResult

type ModelGateway interface {
	DiscoverModels(ctx context.Context, provider ports.ChatModelProvider) ProviderModelDiscoveryResult
	StreamChat(ctx context.Context, provider ports.ChatModelProvider, profile ports.ChatModelProfile, messages []ChatGatewayMessage) <-chan ModelStreamChunk
	HealthCheck(ctx context.Context, provider ports.ChatModelProvider) ProviderHealth
}

func WithModelGateway(gateway ModelGateway) StoreOption {
	return func(store *Store) {
		if gateway != nil {
			store.ModelGateway = gateway
		}
	}
}

type localModelGateway struct{}

func NewLocalModelGateway() ModelGateway {
	return localModelGateway{}
}

func (localModelGateway) DiscoverModels(_ context.Context, provider ports.ChatModelProvider) ProviderModelDiscoveryResult {
	if provider.ProviderID == "provider_local_stub" || provider.DefaultModel == "model_generate_stub" {
		return ProviderModelDiscoveryResult{
			Models:     []string{"model_generate_stub"},
			LatencyMS:  1,
			Discovered: true,
		}
	}
	return ProviderModelDiscoveryResult{
		Models:    append([]string{}, provider.AvailableModels...),
		LatencyMS: 1,
		ErrorCode: "MODEL_PROVIDER_UNAVAILABLE",
		LastError: "real provider adapter is not wired in this runtime",
	}
}

func (localModelGateway) HealthCheck(_ context.Context, provider ports.ChatModelProvider) ProviderHealth {
	if provider.ProviderID == "provider_local_stub" || provider.DefaultModel == "model_generate_stub" {
		return ProviderHealth{
			OK:                provider.Enabled,
			Status:            "connected",
			Message:           "local deterministic streaming provider is ready",
			LatencyMS:         1,
			StreamingVerified: true,
		}
	}
	if !provider.Enabled {
		return ProviderHealth{
			Status:    "unknown",
			Message:   "provider is disabled",
			LatencyMS: 1,
			ErrorCode: "MODEL_PROVIDER_DISABLED",
		}
	}
	return ProviderHealth{
		Status:    "error",
		Message:   "real provider adapter is not wired in this runtime",
		LatencyMS: 1,
		ErrorCode: "MODEL_PROVIDER_UNAVAILABLE",
	}
}

func (localModelGateway) StreamChat(
	ctx context.Context,
	provider ports.ChatModelProvider,
	profile ports.ChatModelProfile,
	messages []ChatGatewayMessage,
) <-chan ModelStreamChunk {
	out := make(chan ModelStreamChunk, 16)
	go func() {
		defer close(out)
		if profile.Model != "model_generate_stub" && provider.ProviderID != "provider_local_stub" {
			out <- ModelStreamChunk{Error: streamError("MODEL_PROVIDER_UNAVAILABLE", "real provider adapter is not wired in this runtime", true)}
			return
		}
		streamLocalStub(ctx, messages, out)
	}()
	return out
}

func streamLocalStub(ctx context.Context, messages []ChatGatewayMessage, out chan<- ModelStreamChunk) {
	last := ""
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			last = messages[i].Content
			break
		}
	}
	reply := fmt.Sprintf("Local streaming model received: %s\n\nRuntime path: PLAN -> GRAPH -> EXECUTE -> OBSERVE -> REPLAN. Configure a real provider in Resource Center to call external streaming models.", last)
	for _, part := range splitForStreaming(reply) {
		select {
		case <-ctx.Done():
			out <- ModelStreamChunk{FinishReason: "cancelled"}
			return
		case out <- ModelStreamChunk{Delta: part}:
		}
	}
	usage := estimateChatUsage(messages, reply)
	out <- ModelStreamChunk{Usage: &usage, FinishReason: "stop"}
}

func splitForStreaming(value string) []string {
	words := strings.Split(value, " ")
	if len(words) <= 1 {
		return []string{value}
	}
	parts := make([]string, 0, len(words))
	for i, word := range words {
		if i == 0 {
			parts = append(parts, word)
		} else {
			parts = append(parts, " "+word)
		}
	}
	return parts
}

func estimateChatUsage(messages []ChatGatewayMessage, output string) ChatModelUsage {
	input := 0
	for _, message := range messages {
		input += estimateTokens(message.Content)
	}
	outputTokens := estimateTokens(output)
	return ChatModelUsage{InputTokens: input, OutputTokens: outputTokens, TotalTokens: input + outputTokens}
}

func estimateTokens(value string) int {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0
	}
	return len([]rune(trimmed))/4 + 1
}

func EstimateTokens(value string) int {
	return estimateTokens(value)
}

func streamError(code string, message string, recoverable bool) *ChatStreamError {
	if code == "" {
		code = "MODEL_STREAM_FAILED"
	}
	if message == "" {
		message = "model stream failed"
	}
	return &ChatStreamError{Code: code, Message: message, Recoverable: recoverable}
}

func toChatModelProvider(provider ModelProviderRecord) ports.ChatModelProvider {
	return ports.ChatModelProvider{
		ProviderID:      provider.ProviderID,
		ProviderType:    string(provider.ProviderType),
		DisplayName:     provider.DisplayName,
		BaseURL:         provider.BaseURL,
		Organization:    provider.Organization,
		Project:         provider.Project,
		DefaultModel:    provider.DefaultModel,
		AvailableModels: append([]string{}, provider.AvailableModels...),
		Enabled:         provider.Enabled,
		APIKey:          provider.APIKey,
		APIKeyOptional:  ProviderAllowsMissingAPIKey(provider),
	}
}

func ToChatModelProvider(provider ModelProviderRecord) ports.ChatModelProvider {
	return toChatModelProvider(provider)
}

func toChatModelProfile(profile ModelProfile) ports.ChatModelProfile {
	return ports.ChatModelProfile{
		ProfileID:      profile.ProfileID,
		DisplayName:    profile.DisplayName,
		ProviderID:     profile.ProviderID,
		Model:          profile.Model,
		Temperature:    profile.Temperature,
		MaxTokens:      profile.MaxTokens,
		ContextWindow:  profile.ContextWindow,
		ResponseFormat: profile.ResponseFormat,
		ToolMode:       profile.ToolMode,
		TimeoutMS:      profile.TimeoutMS,
	}
}

func ToChatModelProfile(profile ModelProfile) ports.ChatModelProfile {
	return toChatModelProfile(profile)
}

func ProviderAllowsMissingAPIKey(provider ModelProviderRecord) bool {
	return provider.ProviderID == "provider_local_stub" ||
		provider.ProviderID == "provider_9router_local" ||
		provider.DefaultModel == "model_generate_stub" ||
		provider.ProviderType == ProviderOllama
}
