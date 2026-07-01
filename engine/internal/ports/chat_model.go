package ports

import "context"

type ChatModelMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatModelUsage struct {
	InputTokens  int     `json:"inputTokens"`
	OutputTokens int     `json:"outputTokens"`
	TotalTokens  int     `json:"totalTokens"`
	CostUSD      float64 `json:"costUsd"`
}

type ChatStreamError struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Recoverable bool   `json:"recoverable"`
}

type ToolExecutionRequest struct {
	CallID      string `json:"callId"`
	ToolID      string `json:"toolId"`
	DisplayName string `json:"displayName"`
	Arguments   string `json:"arguments"`
}

type ChatModelStreamChunk struct {
	Delta          string
	ReasoningDelta string
	ToolCall       *ToolExecutionRequest
	Usage          *ChatModelUsage
	FinishReason   string
	LatencyMS      int
	Error          *ChatStreamError
}

type ChatModelProvider struct {
	ProviderID      string
	ProviderType    string
	DisplayName     string
	BaseURL         string
	Organization    *string
	Project         *string
	DefaultModel    string
	AvailableModels []string
	Enabled         bool
	APIKey          string
}

type ChatModelProfile struct {
	ProfileID      string
	DisplayName    string
	ProviderID     string
	Model          string
	Temperature    float64
	MaxTokens      int
	ContextWindow  int
	ResponseFormat string
	ToolMode       string
	TimeoutMS      int
}

type ProviderHealth struct {
	OK                 bool
	Status             string
	Message            string
	LatencyMS          int
	ErrorCode          string
	StreamingVerified  bool
	LastDiscoveredAt   string
	DiscoveredModelCnt int
}

type ProviderModelDiscoveryResult struct {
	Models     []string
	LatencyMS  int
	ErrorCode  string
	LastError  string
	Discovered bool
}

type ChatModelGateway interface {
	DiscoverModels(ctx context.Context, provider ChatModelProvider) ProviderModelDiscoveryResult
	StreamChat(ctx context.Context, provider ChatModelProvider, profile ChatModelProfile, messages []ChatModelMessage) <-chan ChatModelStreamChunk
	HealthCheck(ctx context.Context, provider ChatModelProvider) ProviderHealth
}
