package resources

import (
	"net/http"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/ports"
)

type AppError struct {
	Status     int
	Code       string
	Message    string
	UserAction string
}

func BadRequest(code string, message string, userAction string) *AppError {
	return &AppError{Status: http.StatusBadRequest, Code: code, Message: message, UserAction: userAction}
}

func NotFound(code string, message string, userAction string) *AppError {
	return &AppError{Status: http.StatusNotFound, Code: code, Message: message, UserAction: userAction}
}

type ProviderType string

const (
	ProviderOpenAICompatible ProviderType = "openai_compatible"
	ProviderDeepSeek         ProviderType = "deepseek"
	ProviderOpenAI           ProviderType = "openai"
	ProviderAnthropic        ProviderType = "anthropic"
	ProviderGLM              ProviderType = "glm"
	ProviderVolcano          ProviderType = "volcano"
	ProviderSiliconFlow      ProviderType = "siliconflow"
	ProviderGemini           ProviderType = "gemini"
	ProviderOllama           ProviderType = "ollama"
	ProviderCustom           ProviderType = "custom"
)

type SafeModelProvider struct {
	ProviderID        string       `json:"providerId"`
	ProviderType      ProviderType `json:"providerType"`
	DisplayName       string       `json:"displayName"`
	BaseURL           string       `json:"baseURL"`
	Organization      *string      `json:"organization"`
	Project           *string      `json:"project"`
	DefaultModel      string       `json:"defaultModel"`
	AvailableModels   []string     `json:"availableModels"`
	Enabled           bool         `json:"enabled"`
	Status            string       `json:"status"`
	Capabilities      []string     `json:"capabilities"`
	SupportsStream    bool         `json:"supportsStreaming"`
	HealthStatus      string       `json:"healthStatus"`
	ModelCount        int          `json:"modelCount"`
	LatencyMS         int          `json:"latencyMs"`
	LastDiscoveryAt   *string      `json:"lastDiscoveryAt"`
	LastStreamAt      *string      `json:"lastStreamAt"`
	LastErrorCode     *string      `json:"lastErrorCode"`
	StreamingVerified bool         `json:"streamingVerified"`
	HasAPIKey         bool         `json:"hasApiKey"`
	MaskedKey         *string      `json:"maskedKey"`
	LastTestedAt      *string      `json:"lastTestedAt"`
	LastError         *string      `json:"lastError"`
	CreatedAt         string       `json:"createdAt"`
	UpdatedAt         string       `json:"updatedAt"`
}

type ModelProviderRecord struct {
	SafeModelProvider
	APIKey string `json:"-"`
}

type SaveModelProviderInput struct {
	ProviderID      string       `json:"providerId"`
	ProviderType    ProviderType `json:"providerType"`
	DisplayName     string       `json:"displayName"`
	BaseURL         string       `json:"baseURL"`
	Organization    *string      `json:"organization"`
	Project         *string      `json:"project"`
	DefaultModel    string       `json:"defaultModel"`
	AvailableModels []string     `json:"availableModels"`
	Enabled         bool         `json:"enabled"`
	Capabilities    []string     `json:"capabilities"`
	APIKey          string       `json:"apiKey"`
}

type ModelProfile struct {
	ProfileID         string  `json:"profileId"`
	DisplayName       string  `json:"displayName"`
	ProviderID        string  `json:"providerId"`
	Model             string  `json:"model"`
	Temperature       float64 `json:"temperature"`
	MaxTokens         int     `json:"maxTokens"`
	ContextWindow     int     `json:"contextWindow"`
	ResponseFormat    string  `json:"responseFormat"`
	ToolMode          string  `json:"toolMode"`
	FallbackProfileID *string `json:"fallbackProfileId"`
	TimeoutMS         int     `json:"timeoutMs"`
	Purpose           string  `json:"purpose"`
	Enabled           bool    `json:"enabled"`
	CreatedAt         string  `json:"createdAt"`
	UpdatedAt         string  `json:"updatedAt"`
}

type AgentRuntimeConfig struct {
	ContextWindow int     `json:"contextWindow"`
	Temperature   float64 `json:"temperature"`
	MaxTokens     int     `json:"maxTokens"`
}

type AgentPlannerConfig struct {
	Enabled  bool   `json:"enabled"`
	Strategy string `json:"strategy"`
}

type AgentExecutorConfig struct {
	TimeoutMS   int    `json:"timeoutMs"`
	RetryPolicy string `json:"retryPolicy"`
}

type AgentConfig struct {
	AgentID           string              `json:"agentId"`
	DisplayName       string              `json:"displayName"`
	Role              string              `json:"role"`
	Description       string              `json:"description"`
	SystemPrompt      string              `json:"systemPrompt"`
	ModelProfileID    string              `json:"modelProfileId"`
	EnabledSkills     []string            `json:"enabledSkills"`
	EnabledTools      []string            `json:"enabledTools"`
	EnabledMCPServers []string            `json:"enabledMcpServers"`
	RuntimeConfig     AgentRuntimeConfig  `json:"runtimeConfig"`
	Planner           AgentPlannerConfig  `json:"planner"`
	Executor          AgentExecutorConfig `json:"executor"`
	MemoryScope       string              `json:"memoryScope"`
	Enabled           bool                `json:"enabled"`
	BuiltIn           bool                `json:"builtIn"`
	CreatedAt         string              `json:"createdAt"`
	UpdatedAt         string              `json:"updatedAt"`
}

type SkillConfig struct {
	SkillID              string   `json:"skillId"`
	CommandName          string   `json:"commandName"`
	DisplayName          string   `json:"displayName"`
	Description          string   `json:"description"`
	WhenToUse            string   `json:"whenToUse"`
	Instructions         string   `json:"instructions"`
	Category             string   `json:"category"`
	Version              string   `json:"version"`
	Enabled              bool     `json:"enabled"`
	BuiltIn              bool     `json:"builtIn"`
	SourcePath           string   `json:"sourcePath"`
	RequiredCapabilities []string `json:"requiredCapabilities"`
	OutputArtifacts      []string `json:"outputArtifacts"`
}

type ToolConfig struct {
	ToolID      string `json:"toolId"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	Category    string `json:"category"`
	RiskLevel   string `json:"riskLevel"`
	Enabled     bool   `json:"enabled"`
	BuiltIn     bool   `json:"builtIn"`
}

type MCPServerConfig struct {
	ServerID      string   `json:"serverId"`
	DisplayName   string   `json:"displayName"`
	Command       string   `json:"command"`
	Args          []string `json:"args"`
	EnvKeys       []string `json:"envKeys"`
	URL           *string  `json:"url"`
	TrustLevel    string   `json:"trustLevel"`
	Enabled       bool     `json:"enabled"`
	HasSecrets    bool     `json:"hasSecrets"`
	MaskedSecrets []string `json:"maskedSecrets"`
	CreatedAt     string   `json:"createdAt"`
	UpdatedAt     string   `json:"updatedAt"`
}

type MCPServerRecord struct {
	MCPServerConfig
	Secrets map[string]string `json:"-"`
}

type SaveMCPServerInput struct {
	ServerID    string            `json:"serverId"`
	DisplayName string            `json:"displayName"`
	Command     string            `json:"command"`
	Args        []string          `json:"args"`
	URL         *string           `json:"url"`
	TrustLevel  string            `json:"trustLevel"`
	Enabled     bool              `json:"enabled"`
	Secrets     map[string]string `json:"secrets"`
}

type Project struct {
	ProjectID             string   `json:"projectId"`
	Title                 string   `json:"title"`
	Description           string   `json:"description"`
	Status                string   `json:"status"`
	DefaultModelProfileID string   `json:"defaultModelProfileId"`
	EnabledAgents         []string `json:"enabledAgents"`
	EnabledSkills         []string `json:"enabledSkills"`
	EnabledTools          []string `json:"enabledTools"`
	EnabledMCPServers     []string `json:"enabledMcpServers"`
	CreatedAt             string   `json:"createdAt"`
	UpdatedAt             string   `json:"updatedAt"`
}

type CreateProjectInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type UpdateProjectInput struct {
	ProjectID             string    `json:"projectId"`
	Title                 *string   `json:"title"`
	Description           *string   `json:"description"`
	Status                *string   `json:"status"`
	DefaultModelProfileID *string   `json:"defaultModelProfileId"`
	EnabledAgents         *[]string `json:"enabledAgents"`
	EnabledSkills         *[]string `json:"enabledSkills"`
	EnabledTools          *[]string `json:"enabledTools"`
	EnabledMCPServers     *[]string `json:"enabledMcpServers"`
}

type ProjectModule struct {
	ProjectID         string             `json:"projectId"`
	ModuleID          string             `json:"moduleId"`
	DisplayName       string             `json:"displayName"`
	Status            string             `json:"status"`
	Summary           string             `json:"summary"`
	DefaultAgents     []string           `json:"defaultAgents"`
	EnabledSkills     []string           `json:"enabledSkills"`
	EnabledTools      []string           `json:"enabledTools"`
	EnabledMCPServers []string           `json:"enabledMcpServers"`
	OutputArtifacts   []string           `json:"outputArtifacts"`
	NextBestAction    string             `json:"nextBestAction"`
	Submodules        []ProjectSubmodule `json:"submodules"`
	Config            map[string]any     `json:"config"`
}

type ProjectSubmodule struct {
	ProjectID       string         `json:"projectId"`
	ModuleID        string         `json:"moduleId"`
	SubmoduleID     string         `json:"submoduleId"`
	DisplayName     string         `json:"displayName"`
	Status          string         `json:"status"`
	Summary         string         `json:"summary"`
	DefaultAgents   []string       `json:"defaultAgents"`
	EnabledSkills   []string       `json:"enabledSkills"`
	EnabledTools    []string       `json:"enabledTools"`
	OutputArtifacts []string       `json:"outputArtifacts"`
	NextBestAction  string         `json:"nextBestAction"`
	Config          map[string]any `json:"config"`
}

type ModuleRequest struct {
	ProjectID string `json:"projectId"`
	ModuleID  string `json:"moduleId"`
}

type UpdateModuleConfigInput struct {
	ProjectID string         `json:"projectId"`
	ModuleID  string         `json:"moduleId"`
	Config    map[string]any `json:"config"`
}

type ChatSession struct {
	SessionID      string  `json:"sessionId"`
	ProjectID      *string `json:"projectId"`
	Title          string  `json:"title"`
	AgentID        string  `json:"agentId"`
	ModelProfileID string  `json:"modelProfileId"`
	MessageCount   int     `json:"messageCount"`
	CreatedAt      string  `json:"createdAt"`
	UpdatedAt      string  `json:"updatedAt"`
}

type CreateChatSessionInput struct {
	ProjectID      *string `json:"projectId"`
	Title          string  `json:"title"`
	AgentID        string  `json:"agentId"`
	ModelProfileID string  `json:"modelProfileId"`
}

type UpdateChatSessionInput struct {
	SessionID      string  `json:"sessionId"`
	ProjectID      *string `json:"projectId"`
	Title          string  `json:"title"`
	AgentID        string  `json:"agentId"`
	ModelProfileID string  `json:"modelProfileId"`
}

type ChatMessage struct {
	MessageID      string          `json:"messageId"`
	AttemptID      string          `json:"attemptId"`
	SessionID      string          `json:"sessionId"`
	Role           string          `json:"role"`
	Content        string          `json:"content"`
	Status         string          `json:"status"`
	ProviderID     string          `json:"providerId"`
	Model          string          `json:"model"`
	Usage          *ChatModelUsage `json:"usage"`
	FinishReason   string          `json:"finishReason"`
	RuntimeSummary string          `json:"runtimeSummary"`
	TraceID        string          `json:"trace_id"`
	CreatedAt      string          `json:"createdAt"`
}

type SendChatMessageInput struct {
	SessionID        string `json:"sessionId"`
	Content          string `json:"content"`
	StreamID         string `json:"streamId"`
	RetryOfMessageID string `json:"retryOfMessageId"`
}

type CancelChatStreamInput struct {
	StreamID string `json:"streamId"`
}

type ChatExecutionStep struct {
	StepID      string `json:"stepId"`
	Phase       string `json:"phase"`
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	Status      string `json:"status"`
	StartedAt   string `json:"startedAt"`
	CompletedAt string `json:"completedAt"`
}

type ChatToolCallPreview struct {
	CallID           string `json:"callId"`
	ToolID           string `json:"toolId"`
	DisplayName      string `json:"displayName"`
	RiskLevel        string `json:"riskLevel"`
	ApprovalRequired bool   `json:"approvalRequired"`
	Status           string `json:"status"`
	Summary          string `json:"summary"`
	Arguments        string `json:"arguments,omitempty"`
	ResultSummary    string `json:"resultSummary,omitempty"`
	ErrorCode        string `json:"errorCode,omitempty"`
}

type ChatTurnResult struct {
	Session         ChatSession           `json:"session"`
	Messages        []ChatMessage         `json:"messages"`
	ExecutionSteps  []ChatExecutionStep   `json:"executionSteps"`
	ToolCalls       []ChatToolCallPreview `json:"toolCalls"`
	ContextSummary  *ChatContextSummary   `json:"contextSummary"`
	ContextBudget   ContextBudgetReport   `json:"contextBudget"`
	AuditSummary    ChatAuditSummary      `json:"auditSummary"`
	ProviderStatus  string                `json:"providerStatus"`
	RuntimeSnapshot string                `json:"runtimeSnapshot"`
	RuntimeSummary  string                `json:"runtimeSummary"`
}

type ChatAuditSummary struct {
	ContentHash  string          `json:"contentHash"`
	ProviderID   string          `json:"providerId"`
	Model        string          `json:"model"`
	LatencyMS    int             `json:"latencyMs"`
	ErrorCode    string          `json:"errorCode"`
	Usage        *ChatModelUsage `json:"usage"`
	FinishReason string          `json:"finishReason"`
}

type ChatModelUsage = ports.ChatModelUsage

type ChatContextPack struct {
	SystemPrompt     string                   `json:"systemPrompt"`
	Messages         []ChatGatewayMessage     `json:"messages"`
	Summaries        []ChatContextSummary     `json:"summaries"`
	Skills           []SkillRuntimeDescriptor `json:"skills"`
	Tools            []ToolRuntimeDescriptor  `json:"tools"`
	MCPServers       []string                 `json:"mcpServers"`
	Budget           ContextBudgetReport      `json:"budget"`
	ProviderFallback string                   `json:"providerFallback"`
}

type ChatRuntimeSelection struct {
	Summary    string                   `json:"summary"`
	Skills     []SkillRuntimeDescriptor `json:"skills"`
	Tools      []ToolRuntimeDescriptor  `json:"tools"`
	MCPServers []string                 `json:"mcpServers"`
}

type ContextBudgetReport struct {
	ContextWindow       int      `json:"contextWindow"`
	MaxOutputTokens     int      `json:"maxOutputTokens"`
	InputBudgetTokens   int      `json:"inputBudgetTokens"`
	EstimatedTokens     int      `json:"estimatedTokens"`
	SystemTokens        int      `json:"systemTokens"`
	RecentMessageTokens int      `json:"recentMessageTokens"`
	SummaryTokens       int      `json:"summaryTokens"`
	RecentMessageCount  int      `json:"recentMessageCount"`
	CompactedCount      int      `json:"compactedCount"`
	Compacted           bool     `json:"compacted"`
	Warnings            []string `json:"warnings"`
}

type ChatContextSummary struct {
	SummaryID        string   `json:"summaryId"`
	SessionID        string   `json:"sessionId"`
	SourceMessageIDs []string `json:"sourceMessageIds"`
	Content          string   `json:"content"`
	ContentHash      string   `json:"contentHash"`
	TokenEstimate    int      `json:"tokenEstimate"`
	CreatedBy        string   `json:"createdBy"`
	ContextVersion   int      `json:"contextVersion"`
	CreatedAt        string   `json:"createdAt"`
}

type SkillRuntimeDescriptor struct {
	SkillID              string   `json:"skillId"`
	DisplayName          string   `json:"displayName"`
	Instruction          string   `json:"instruction"`
	RequiredCapabilities []string `json:"requiredCapabilities"`
	OutputArtifacts      []string `json:"outputArtifacts"`
	RuntimePolicy        string   `json:"runtimePolicy"`
}

type ToolRuntimeDescriptor struct {
	ToolID           string `json:"toolId"`
	DisplayName      string `json:"displayName"`
	Description      string `json:"description"`
	RiskLevel        string `json:"riskLevel"`
	AutoExecutable   bool   `json:"autoExecutable"`
	ApprovalRequired bool   `json:"approvalRequired"`
}

type ToolExecutionRequest = ports.ToolExecutionRequest

type ToolExecutionResult struct {
	CallID        string `json:"callId"`
	ToolID        string `json:"toolId"`
	Status        string `json:"status"`
	OutputSummary string `json:"outputSummary"`
	ErrorCode     string `json:"errorCode"`
	ErrorMessage  string `json:"errorMessage"`
	LatencyMS     int    `json:"latencyMs"`
}

type ChatStreamStartResult struct {
	StreamID string `json:"streamId"`
}

type ChatStreamEvent struct {
	Type             string                `json:"type"`
	StreamID         string                `json:"streamId"`
	SessionID        string                `json:"sessionId"`
	MessageID        string                `json:"messageId"`
	TraceID          string                `json:"trace_id"`
	Sequence         int                   `json:"sequence"`
	Timestamp        string                `json:"timestamp"`
	Delta            string                `json:"delta,omitempty"`
	ReasoningDelta   string                `json:"reasoningDelta,omitempty"`
	Step             *ChatExecutionStep    `json:"step,omitempty"`
	ToolCall         *ChatToolCallPreview  `json:"toolCall,omitempty"`
	ToolResult       *ToolExecutionResult  `json:"toolResult,omitempty"`
	RuntimeSelection *ChatRuntimeSelection `json:"runtimeSelection,omitempty"`
	ContextBudget    *ContextBudgetReport  `json:"contextBudget,omitempty"`
	ContextSummary   *ChatContextSummary   `json:"contextSummary,omitempty"`
	Usage            *ChatModelUsage       `json:"usage,omitempty"`
	Result           *ChatTurnResult       `json:"result,omitempty"`
	Error            *ChatStreamError      `json:"error,omitempty"`
	Warning          *ChatStreamWarning    `json:"warning,omitempty"`
	ProviderID       string                `json:"providerId,omitempty"`
	Model            string                `json:"model,omitempty"`
	FinishReason     string                `json:"finishReason,omitempty"`
	AttemptID        string                `json:"attemptId,omitempty"`
	LatencyMS        int                   `json:"latencyMs,omitempty"`
}

type ChatStreamError = ports.ChatStreamError

type ChatStreamWarning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type DeleteResult struct {
	OK        bool   `json:"ok"`
	DeletedID string `json:"deletedId"`
}

type TestResult struct {
	OK        bool   `json:"ok"`
	TargetID  string `json:"targetId"`
	Message   string `json:"message"`
	LatencyMS int    `json:"latencyMs"`
	TraceID   string `json:"trace_id"`
}

type IDRequest struct {
	ProviderID string `json:"providerId"`
	ProfileID  string `json:"profileId"`
	AgentID    string `json:"agentId"`
	SkillID    string `json:"skillId"`
	ToolID     string `json:"toolId"`
	ServerID   string `json:"serverId"`
	ProjectID  string `json:"projectId"`
	SessionID  string `json:"sessionId"`
	Enabled    bool   `json:"enabled"`
}
