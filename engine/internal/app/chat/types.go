package chat

import (
	"sort"
	"strings"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/resources"
)

type Store struct {
	*resources.Store
}

func NewStore(state *resources.Store) *Store {
	return &Store{Store: state}
}

type AppError = resources.AppError
type AgentConfig = resources.AgentConfig
type ChatAuditSummary = resources.ChatAuditSummary
type ChatContextPack = resources.ChatContextPack
type ChatContextSummary = resources.ChatContextSummary
type ChatExecutionStep = resources.ChatExecutionStep
type ChatGatewayMessage = resources.ChatGatewayMessage
type ChatMessage = resources.ChatMessage
type ChatModelUsage = resources.ChatModelUsage
type ChatRuntimeSelection = resources.ChatRuntimeSelection
type ChatSession = resources.ChatSession
type ChatStreamError = resources.ChatStreamError
type ChatStreamEvent = resources.ChatStreamEvent
type ChatStreamWarning = resources.ChatStreamWarning
type ChatToolCallPreview = resources.ChatToolCallPreview
type ChatTurnResult = resources.ChatTurnResult
type ContextBudgetReport = resources.ContextBudgetReport
type CreateChatSessionInput = resources.CreateChatSessionInput
type DeleteResult = resources.DeleteResult
type ModelProfile = resources.ModelProfile
type ModelProviderRecord = resources.ModelProviderRecord
type SendChatMessageInput = resources.SendChatMessageInput
type SkillRuntimeDescriptor = resources.SkillRuntimeDescriptor
type ToolExecutionResult = resources.ToolExecutionResult
type ToolExecutionRequest = resources.ToolExecutionRequest
type ToolRuntimeDescriptor = resources.ToolRuntimeDescriptor
type UpdateChatSessionInput = resources.UpdateChatSessionInput
type CancelChatStreamInput = resources.CancelChatStreamInput

var BadRequest = resources.BadRequest
var NotFound = resources.NotFound

func sortedValues[T any](items map[string]T, key func(T) string) []T {
	values := make([]T, 0, len(items))
	for _, value := range items {
		values = append(values, value)
	}
	sort.Slice(values, func(i, j int) bool {
		return key(values[i]) < key(values[j])
	})
	return values
}

func fallback(value string, fallbackValue string) string {
	if strings.TrimSpace(value) == "" {
		return fallbackValue
	}
	return value
}
