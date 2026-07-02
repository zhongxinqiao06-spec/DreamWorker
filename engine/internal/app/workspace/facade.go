package workspace

import (
	"bufio"
	"context"
	"io"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/extensions"
	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/resources"
)

type AppError = resources.AppError
type ProviderType = resources.ProviderType
type SafeModelProvider = resources.SafeModelProvider
type ModelProviderRecord = resources.ModelProviderRecord
type SaveModelProviderInput = resources.SaveModelProviderInput
type ModelProfile = resources.ModelProfile
type AgentRuntimeConfig = resources.AgentRuntimeConfig
type AgentPlannerConfig = resources.AgentPlannerConfig
type AgentExecutorConfig = resources.AgentExecutorConfig
type AgentConfig = resources.AgentConfig
type SkillConfig = resources.SkillConfig
type ToolConfig = resources.ToolConfig
type MCPServerConfig = resources.MCPServerConfig
type MCPServerRecord = resources.MCPServerRecord
type SaveMCPServerInput = resources.SaveMCPServerInput
type Project = resources.Project
type CreateProjectInput = resources.CreateProjectInput
type UpdateProjectInput = resources.UpdateProjectInput
type ProjectModule = resources.ProjectModule
type ProjectSubmodule = resources.ProjectSubmodule
type ModuleRequest = resources.ModuleRequest
type UpdateModuleConfigInput = resources.UpdateModuleConfigInput
type ChatSession = resources.ChatSession
type CreateChatSessionInput = resources.CreateChatSessionInput
type UpdateChatSessionInput = resources.UpdateChatSessionInput
type ChatMessage = resources.ChatMessage
type SendChatMessageInput = resources.SendChatMessageInput
type CancelChatStreamInput = resources.CancelChatStreamInput
type ChatExecutionStep = resources.ChatExecutionStep
type ChatToolCallPreview = resources.ChatToolCallPreview
type ChatTurnResult = resources.ChatTurnResult
type ChatAuditSummary = resources.ChatAuditSummary
type ChatModelUsage = resources.ChatModelUsage
type ChatContextPack = resources.ChatContextPack
type ChatRuntimeSelection = resources.ChatRuntimeSelection
type ContextBudgetReport = resources.ContextBudgetReport
type ChatContextSummary = resources.ChatContextSummary
type SkillRuntimeDescriptor = resources.SkillRuntimeDescriptor
type ToolRuntimeDescriptor = resources.ToolRuntimeDescriptor
type ToolExecutionRequest = resources.ToolExecutionRequest
type ToolExecutionResult = resources.ToolExecutionResult
type ChatStreamStartResult = resources.ChatStreamStartResult
type ChatStreamEvent = resources.ChatStreamEvent
type ChatStreamError = resources.ChatStreamError
type ChatStreamWarning = resources.ChatStreamWarning
type DeleteResult = resources.DeleteResult
type TestResult = resources.TestResult
type IDRequest = resources.IDRequest
type ChatGatewayMessage = resources.ChatGatewayMessage
type ModelStreamChunk = resources.ModelStreamChunk
type ProviderHealth = resources.ProviderHealth
type ProviderModelDiscoveryResult = resources.ProviderModelDiscoveryResult
type AppSettings = extensions.AppSettings
type UpdateSettingsInput = extensions.UpdateSettingsInput
type ExtensionSpec = extensions.ExtensionSpec
type ExtensionStatus = extensions.ExtensionStatus
type NodeRuntimeInfo = extensions.NodeRuntimeInfo
type ManagedProcess = extensions.ManagedProcess
type ExtensionLogLine = extensions.ExtensionLogLine
type InstallExtensionInput = extensions.InstallExtensionInput
type ExtensionIDRequest = extensions.ExtensionIDRequest
type TailLogsInput = extensions.TailLogsInput
type ExtensionActionResult = extensions.ExtensionActionResult
type ExtensionModelRefreshResult = extensions.ExtensionModelRefreshResult
type ExtensionStreamingResult = extensions.ExtensionStreamingResult

const (
	ProviderOpenAICompatible ProviderType = resources.ProviderOpenAICompatible
	ProviderDeepSeek         ProviderType = resources.ProviderDeepSeek
	ProviderOpenAI           ProviderType = resources.ProviderOpenAI
	ProviderAnthropic        ProviderType = resources.ProviderAnthropic
	ProviderGLM              ProviderType = resources.ProviderGLM
	ProviderVolcano          ProviderType = resources.ProviderVolcano
	ProviderSiliconFlow      ProviderType = resources.ProviderSiliconFlow
	ProviderGemini           ProviderType = resources.ProviderGemini
	ProviderOllama           ProviderType = resources.ProviderOllama
	ProviderCustom           ProviderType = resources.ProviderCustom
)

var BadRequest = resources.BadRequest
var NotFound = resources.NotFound

func writeMCPMessage(writer io.Writer, payload any) error {
	return resources.WriteMCPMessage(writer, payload)
}

func readMCPMessage(reader *bufio.Reader) ([]byte, error) {
	return resources.ReadMCPMessage(reader)
}

func (s *Store) ListProjects() []Project {
	return s.projectStore.ListProjects()
}

func (s *Store) CreateProject(input CreateProjectInput) (Project, *AppError) {
	return s.projectStore.CreateProject(input)
}

func (s *Store) GetProject(projectID string) (Project, *AppError) {
	return s.projectStore.GetProject(projectID)
}

func (s *Store) UpdateProject(input UpdateProjectInput) (Project, *AppError) {
	return s.projectStore.UpdateProject(input)
}

func (s *Store) DeleteProject(projectID string) (DeleteResult, *AppError) {
	return s.projectStore.DeleteProject(projectID)
}

func (s *Store) ListProjectModules(projectID string) ([]ProjectModule, *AppError) {
	return s.projectStore.ListProjectModules(projectID)
}

func (s *Store) GetProjectModule(input ModuleRequest) (ProjectModule, *AppError) {
	return s.projectStore.GetProjectModule(input)
}

func (s *Store) UpdateProjectModuleConfig(input UpdateModuleConfigInput) (ProjectModule, *AppError) {
	return s.projectStore.UpdateProjectModuleConfig(input)
}

func (s *Store) ListChatSessions() []ChatSession {
	return s.chatStore.ListChatSessions()
}

func (s *Store) ListChatMessages(sessionID string) ([]ChatMessage, *AppError) {
	return s.chatStore.ListChatMessages(sessionID)
}

func (s *Store) CreateChatSession(input CreateChatSessionInput) (ChatSession, *AppError) {
	return s.chatStore.CreateChatSession(input)
}

func (s *Store) UpdateChatSession(input UpdateChatSessionInput) (ChatSession, *AppError) {
	return s.chatStore.UpdateChatSession(input)
}

func (s *Store) SendChatMessage(input SendChatMessageInput) (ChatTurnResult, *AppError) {
	return s.chatStore.SendChatMessage(input)
}

func (s *Store) StreamChatMessage(ctx context.Context, input SendChatMessageInput) (<-chan ChatStreamEvent, *AppError) {
	return s.chatStore.StreamChatMessage(ctx, input)
}

func (s *Store) CancelChatStream(input CancelChatStreamInput) (DeleteResult, *AppError) {
	return s.chatStore.CancelChatStream(input)
}

func (s *Store) DeleteChatSession(sessionID string) (DeleteResult, *AppError) {
	return s.chatStore.DeleteChatSession(sessionID)
}
