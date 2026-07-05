package coding

import "github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/resources"

type AppError = resources.AppError
type DeleteResult = resources.DeleteResult

type EngineID string

const (
	EngineClaudeAgent EngineID = "claude_agent"
	EngineCodex       EngineID = "codex"
	EngineOpenCode    EngineID = "opencode"
)

type EngineDescriptor struct {
	EngineID               EngineID `json:"engineId"`
	DisplayName            string   `json:"displayName"`
	Description            string   `json:"description"`
	SupportedProviderTypes []string `json:"supportedProviderTypes"`
	PreferredProviderIDs   []string `json:"preferredProviderIds"`
	DirectWrite            bool     `json:"directWrite"`
	Streaming              bool     `json:"streaming"`
}

type RuntimeStatus struct {
	RuntimeDir  string             `json:"runtimeDir"`
	NodeBin     string             `json:"nodeBin"`
	AdapterPath string             `json:"adapterPath"`
	Available   bool               `json:"available"`
	Message     string             `json:"message"`
	Engines     []EngineDescriptor `json:"engines"`
}

type CreateSessionInput struct {
	ProjectID  string   `json:"projectId"`
	EngineID   EngineID `json:"engineId"`
	ProviderID string   `json:"providerId"`
	Model      string   `json:"model"`
	Title      string   `json:"title"`
}

type Session struct {
	SessionID      string   `json:"sessionId"`
	ProjectID      string   `json:"projectId"`
	EngineID       EngineID `json:"engineId"`
	ProviderID     string   `json:"providerId"`
	Model          string   `json:"model"`
	Title          string   `json:"title"`
	LocalRootPath  string   `json:"localRootPath"`
	EngineThreadID string   `json:"engineThreadId"`
	Status         string   `json:"status"`
	CreatedAt      string   `json:"createdAt"`
	UpdatedAt      string   `json:"updatedAt"`
}

type TurnInput struct {
	SessionID  string   `json:"sessionId"`
	ProjectID  string   `json:"projectId"`
	EngineID   EngineID `json:"engineId"`
	ProviderID string   `json:"providerId"`
	Model      string   `json:"model"`
	Prompt     string   `json:"prompt"`
	StreamID   string   `json:"streamId"`
}

type CancelTurnInput struct {
	StreamID string `json:"streamId"`
}

type ToolCall struct {
	CallID    string `json:"callId"`
	ToolName  string `json:"toolName"`
	Arguments any    `json:"arguments,omitempty"`
}

type FileChange struct {
	Path   string `json:"path"`
	Status string `json:"status"`
}

type StreamError struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Recoverable bool   `json:"recoverable"`
}

type StreamEvent struct {
	Type             string       `json:"type"`
	StreamID         string       `json:"streamId"`
	SessionID        string       `json:"sessionId"`
	EngineID         EngineID     `json:"engineId"`
	ProviderID       string       `json:"providerId"`
	Model            string       `json:"model"`
	TraceID          string       `json:"trace_id"`
	Sequence         int          `json:"sequence"`
	Timestamp        string       `json:"timestamp"`
	Delta            string       `json:"delta,omitempty"`
	Message          string       `json:"message,omitempty"`
	Command          string       `json:"command,omitempty"`
	Output           string       `json:"output,omitempty"`
	Path             string       `json:"path,omitempty"`
	Status           string       `json:"status,omitempty"`
	EngineThreadID   string       `json:"engineThreadId,omitempty"`
	ToolCall         *ToolCall    `json:"toolCall,omitempty"`
	File             *FileChange  `json:"file,omitempty"`
	Error            *StreamError `json:"error,omitempty"`
	RuntimeAvailable bool         `json:"runtimeAvailable,omitempty"`
}

type ListFilesInput struct {
	ProjectID string `json:"projectId"`
	Query     string `json:"query"`
	Limit     int    `json:"limit"`
}

type FileEntry struct {
	Path       string `json:"path"`
	Name       string `json:"name"`
	IsDir      bool   `json:"isDir"`
	Size       int64  `json:"size"`
	ModifiedAt string `json:"modifiedAt"`
	GitStatus  string `json:"gitStatus,omitempty"`
}

type ReadFileInput struct {
	ProjectID string `json:"projectId"`
	Path      string `json:"path"`
}

type ReadFileResult struct {
	ProjectID string `json:"projectId"`
	Path      string `json:"path"`
	Content   string `json:"content"`
	Size      int64  `json:"size"`
	Truncated bool   `json:"truncated"`
	MimeType  string `json:"mimeType"`
}

type FileStatusInput struct {
	ProjectID string `json:"projectId"`
}

type FileStatus struct {
	ProjectID string       `json:"projectId"`
	Branch    string       `json:"branch"`
	Changes   []FileChange `json:"changes"`
	Clean     bool         `json:"clean"`
	Message   string       `json:"message"`
}
