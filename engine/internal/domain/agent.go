package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	CapabilityIDArtifactRead        = "cap_artifact_read"
	CapabilityIDArtifactWrite       = "cap_artifact_write"
	CapabilityIDWebSearchStub       = "cap_web_search_stub"
	CapabilityIDBrowserReadonlyStub = "cap_browser_readonly_stub"
	CapabilityIDModelGenerateStub   = "cap_model_generate_stub"
	CapabilityIDHumanInput          = "cap_human_input"
)

const (
	EventAgentTaskGraphCreated = "agent.task_graph_created"
	EventAgentTaskStarted      = "agent.task_started"
	EventAgentTaskCompleted    = "agent.task_completed"
	EventAgentTaskFailed       = "agent.task_failed"
	EventModelRequested        = "model.requested"
	EventModelCompleted        = "model.completed"
	EventModelFailed           = "model.failed"
	EventOutputNormalized      = "agent.output_normalized"
	EventNormalizationFailed   = "agent.normalization_failed"
	EventEvalReportCreated     = "eval.report_created"
)

var (
	ErrAgentSpecInvalid         = errors.New("agent spec is invalid")
	ErrTaskGraphInvalid         = errors.New("task graph is invalid")
	ErrTaskCapabilityNotAllowed = errors.New("task capability is not allowed")
	ErrModelBudgetExceeded      = errors.New("model budget exceeded")
	ErrNormalizationFailed      = errors.New("agent output normalization failed")
)

type ApprovalPolicy string

const (
	ApprovalPolicyNever  ApprovalPolicy = "never"
	ApprovalPolicyOnRisk ApprovalPolicy = "on_risk"
	ApprovalPolicyAlways ApprovalPolicy = "always"
)

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

type PromptRef struct {
	PromptID      string `json:"prompt_id"`
	PromptVersion string `json:"prompt_version"`
	AgentID       string `json:"agent_id"`
}

type PromptSpec struct {
	PromptRef
	Changelog string
	Text      string
}

type Budget struct {
	MaxTokens  int     `json:"max_tokens"`
	MaxCostUSD float64 `json:"max_cost_usd"`
}

type ModelProfile struct {
	ID           string `json:"id"`
	Provider     string `json:"provider"`
	Model        string `json:"model"`
	SupportsJSON bool   `json:"supports_json"`
}

type AgentSpec struct {
	SchemaVersion       string         `json:"schema_version"`
	ID                  string         `json:"id"`
	Role                string         `json:"role"`
	InputSchema         map[string]any `json:"input_schema"`
	OutputSchema        map[string]any `json:"output_schema"`
	AllowedCapabilities []string       `json:"allowed_capabilities"`
	DefaultModelProfile string         `json:"default_model_profile"`
	Budget              Budget         `json:"budget"`
	Timeout             string         `json:"timeout"`
	ApprovalPolicy      ApprovalPolicy `json:"approval_policy"`
	ExpectedArtifacts   []string       `json:"expected_artifacts"`
	PromptRef           PromptRef      `json:"prompt_ref"`
}

type Task struct {
	SchemaVersion        string     `json:"schema_version"`
	TaskID               string     `json:"task_id"`
	Stage                StageName  `json:"stage"`
	Goal                 string     `json:"goal"`
	AssignedAgent        string     `json:"assigned_agent"`
	RequiredCapabilities []string   `json:"required_capabilities"`
	ExpectedArtifacts    []string   `json:"expected_artifacts"`
	DependsOn            []string   `json:"depends_on"`
	Budget               Budget     `json:"budget"`
	Status               TaskStatus `json:"status"`
	TraceID              string     `json:"trace_id"`
}

type TaskGraph struct {
	SchemaVersion string `json:"schema_version"`
	MissionID     string `json:"mission_id"`
	RunID         string `json:"run_id"`
	TraceID       string `json:"trace_id"`
	Idea          string `json:"idea"`
	Tasks         []Task `json:"tasks"`
}

type ModelRequest struct {
	SchemaVersion     string              `json:"schema_version"`
	RequestID         string              `json:"request_id"`
	TraceID           string              `json:"trace_id"`
	MissionID         string              `json:"mission_id"`
	RunID             string              `json:"run_id"`
	TaskID            string              `json:"task_id"`
	AgentID           string              `json:"agent_id"`
	Stage             StageName           `json:"stage"`
	Goal              string              `json:"goal"`
	Idea              string              `json:"idea"`
	ModelProfile      string              `json:"model_profile"`
	PromptRef         PromptRef           `json:"prompt_ref"`
	OutputSchema      map[string]any      `json:"output_schema"`
	ExpectedArtifacts []string            `json:"expected_artifacts"`
	CapabilityContext []CapabilityContext `json:"capability_context"`
	Budget            Budget              `json:"budget"`
}

type CapabilityContext struct {
	CapabilityID string          `json:"capability_id"`
	Output       json.RawMessage `json:"output"`
}

type ModelResponse struct {
	SchemaVersion    string          `json:"schema_version"`
	ResponseID       string          `json:"response_id"`
	RequestID        string          `json:"request_id"`
	TraceID          string          `json:"trace_id"`
	Provider         string          `json:"provider"`
	Model            string          `json:"model"`
	RawOutput        string          `json:"raw_output"`
	StructuredOutput json.RawMessage `json:"structured_output"`
	Usage            ModelUsage      `json:"usage"`
	FinishReason     string          `json:"finish_reason"`
}

type ModelStreamEvent struct {
	SchemaVersion string          `json:"schema_version"`
	RequestID     string          `json:"request_id"`
	TraceID       string          `json:"trace_id"`
	Delta         string          `json:"delta,omitempty"`
	Done          bool            `json:"done"`
	Response      *ModelResponse  `json:"response,omitempty"`
	Error         *ModelCallError `json:"error,omitempty"`
}

type ModelCallError struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Recoverable bool   `json:"recoverable"`
}

type ModelUsage struct {
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	CostUSD      float64 `json:"cost_usd"`
}

type ArtifactDraft struct {
	FileName    string `json:"file_name"`
	ArtifactID  string `json:"artifact_id"`
	Kind        string `json:"kind"`
	Title       string `json:"title"`
	ContentType string `json:"content_type"`
	Content     string `json:"content"`
}

type NormalizedOutput struct {
	Artifacts      []ArtifactDraft `json:"artifacts"`
	EvidenceRefs   []string        `json:"evidence_refs,omitempty"`
	NextBestAction string          `json:"next_best_action"`
}

type EvalReport struct {
	SchemaVersion        string    `json:"schema_version"`
	ReportID             string    `json:"report_id"`
	MissionID            string    `json:"mission_id"`
	RunID                string    `json:"run_id"`
	TraceID              string    `json:"trace_id"`
	ArtifactScore        float64   `json:"artifact_score"`
	EvidenceQualityScore float64   `json:"evidence_quality_score"`
	HallucinationRisk    RiskLevel `json:"hallucination_risk"`
	ActionabilityScore   float64   `json:"actionability_score"`
	NextBestAction       string    `json:"next_best_action"`
	CheckedArtifacts     []string  `json:"checked_artifacts"`
}

type TaskGraphCreatedPayload struct {
	TaskIDs []string `json:"task_ids"`
}

type TaskEventPayload struct {
	TaskID        string     `json:"task_id"`
	Stage         StageName  `json:"stage"`
	AssignedAgent string     `json:"assigned_agent"`
	Status        TaskStatus `json:"status"`
	Reason        string     `json:"reason,omitempty"`
}

type ModelRequestedPayload struct {
	RequestID    string    `json:"request_id"`
	TaskID       string    `json:"task_id"`
	AgentID      string    `json:"agent_id"`
	ModelProfile string    `json:"model_profile"`
	PromptRef    PromptRef `json:"prompt_ref"`
}

type ModelCompletedPayload struct {
	RequestID    string     `json:"request_id"`
	ResponseID   string     `json:"response_id"`
	TaskID       string     `json:"task_id"`
	AgentID      string     `json:"agent_id"`
	Usage        ModelUsage `json:"usage"`
	FinishReason string     `json:"finish_reason"`
}

type ModelFailedPayload struct {
	RequestID string `json:"request_id"`
	TaskID    string `json:"task_id"`
	AgentID   string `json:"agent_id"`
	Code      string `json:"code"`
	Message   string `json:"message"`
}

type OutputNormalizedPayload struct {
	TaskID         string   `json:"task_id"`
	ArtifactIDs    []string `json:"artifact_ids"`
	NextBestAction string   `json:"next_best_action"`
}

type NormalizationFailedPayload struct {
	TaskID  string `json:"task_id"`
	Reason  string `json:"reason"`
	RawSize int    `json:"raw_size"`
}

type EvalReportCreatedPayload struct {
	Report EvalReport `json:"report"`
}

func (spec AgentSpec) Validate() error {
	switch {
	case spec.SchemaVersion != ContractSchemaVersion:
		return fmt.Errorf("%w: schema_version must be %s", ErrAgentSpecInvalid, ContractSchemaVersion)
	case strings.TrimSpace(spec.ID) == "":
		return fmt.Errorf("%w: id is required", ErrAgentSpecInvalid)
	case strings.TrimSpace(spec.Role) == "":
		return fmt.Errorf("%w: role is required", ErrAgentSpecInvalid)
	case spec.InputSchema == nil:
		return fmt.Errorf("%w: input_schema is required", ErrAgentSpecInvalid)
	case spec.OutputSchema == nil:
		return fmt.Errorf("%w: output_schema is required", ErrAgentSpecInvalid)
	case len(spec.AllowedCapabilities) == 0:
		return fmt.Errorf("%w: allowed_capabilities is required", ErrAgentSpecInvalid)
	case strings.TrimSpace(spec.DefaultModelProfile) == "":
		return fmt.Errorf("%w: default_model_profile is required", ErrAgentSpecInvalid)
	case spec.Budget.MaxTokens <= 0:
		return fmt.Errorf("%w: budget.max_tokens must be positive", ErrAgentSpecInvalid)
	case spec.Budget.MaxCostUSD < 0:
		return fmt.Errorf("%w: budget.max_cost_usd cannot be negative", ErrAgentSpecInvalid)
	case spec.TimeoutDuration() <= 0:
		return fmt.Errorf("%w: timeout must be a positive duration", ErrAgentSpecInvalid)
	case !spec.ApprovalPolicy.IsValid():
		return fmt.Errorf("%w: approval_policy is invalid", ErrAgentSpecInvalid)
	case spec.PromptRef.PromptID == "" || spec.PromptRef.PromptVersion == "" || spec.PromptRef.AgentID != spec.ID:
		return fmt.Errorf("%w: prompt_ref is invalid", ErrAgentSpecInvalid)
	default:
		return nil
	}
}

func (spec AgentSpec) TimeoutDuration() time.Duration {
	duration, err := time.ParseDuration(spec.Timeout)
	if err != nil {
		return 0
	}
	return duration
}

func (policy ApprovalPolicy) IsValid() bool {
	switch policy {
	case ApprovalPolicyNever, ApprovalPolicyOnRisk, ApprovalPolicyAlways:
		return true
	default:
		return false
	}
}

func (task Task) Validate() error {
	switch {
	case task.SchemaVersion != ContractSchemaVersion:
		return fmt.Errorf("%w: task %s schema_version must be %s", ErrTaskGraphInvalid, task.TaskID, ContractSchemaVersion)
	case !strings.HasPrefix(task.TaskID, "tsk_"):
		return fmt.Errorf("%w: task_id must start with tsk_", ErrTaskGraphInvalid)
	case !task.Stage.IsValid():
		return fmt.Errorf("%w: task %s stage is invalid", ErrTaskGraphInvalid, task.TaskID)
	case strings.TrimSpace(task.Goal) == "":
		return fmt.Errorf("%w: task %s goal is required", ErrTaskGraphInvalid, task.TaskID)
	case strings.TrimSpace(task.AssignedAgent) == "":
		return fmt.Errorf("%w: task %s assigned_agent is required", ErrTaskGraphInvalid, task.TaskID)
	case task.Budget.MaxTokens <= 0:
		return fmt.Errorf("%w: task %s budget.max_tokens must be positive", ErrTaskGraphInvalid, task.TaskID)
	case task.Budget.MaxCostUSD < 0:
		return fmt.Errorf("%w: task %s budget.max_cost_usd cannot be negative", ErrTaskGraphInvalid, task.TaskID)
	case !task.Status.IsValid():
		return fmt.Errorf("%w: task %s status is invalid", ErrTaskGraphInvalid, task.TaskID)
	case !strings.HasPrefix(task.TraceID, "tr_"):
		return fmt.Errorf("%w: task %s trace_id must start with tr_", ErrTaskGraphInvalid, task.TaskID)
	default:
		return nil
	}
}

func (status TaskStatus) IsValid() bool {
	switch status {
	case TaskStatusPending, TaskStatusRunning, TaskStatusCompleted, TaskStatusFailed, TaskStatusCancelled:
		return true
	default:
		return false
	}
}

func (graph TaskGraph) Validate(specs map[string]AgentSpec) error {
	switch {
	case graph.SchemaVersion != ContractSchemaVersion:
		return fmt.Errorf("%w: schema_version must be %s", ErrTaskGraphInvalid, ContractSchemaVersion)
	case !strings.HasPrefix(graph.MissionID, "msn_"):
		return fmt.Errorf("%w: mission_id must start with msn_", ErrTaskGraphInvalid)
	case !strings.HasPrefix(graph.RunID, "run_"):
		return fmt.Errorf("%w: run_id must start with run_", ErrTaskGraphInvalid)
	case !strings.HasPrefix(graph.TraceID, "tr_"):
		return fmt.Errorf("%w: trace_id must start with tr_", ErrTaskGraphInvalid)
	case len(graph.Tasks) == 0:
		return fmt.Errorf("%w: tasks are required", ErrTaskGraphInvalid)
	}

	seen := map[string]Task{}
	for _, task := range graph.Tasks {
		if err := task.Validate(); err != nil {
			return err
		}
		if task.TraceID != graph.TraceID {
			return fmt.Errorf("%w: task %s trace_id must match graph trace_id", ErrTaskGraphInvalid, task.TaskID)
		}
		spec, ok := specs[task.AssignedAgent]
		if !ok {
			return fmt.Errorf("%w: unknown agent %s", ErrTaskGraphInvalid, task.AssignedAgent)
		}
		for _, capabilityID := range task.RequiredCapabilities {
			if !contains(spec.AllowedCapabilities, capabilityID) {
				return fmt.Errorf("%w: %s cannot use %s", ErrTaskCapabilityNotAllowed, task.AssignedAgent, capabilityID)
			}
		}
		if _, exists := seen[task.TaskID]; exists {
			return fmt.Errorf("%w: duplicate task %s", ErrTaskGraphInvalid, task.TaskID)
		}
		seen[task.TaskID] = task
	}

	for _, task := range graph.Tasks {
		for _, dependency := range task.DependsOn {
			if _, ok := seen[dependency]; !ok {
				return fmt.Errorf("%w: task %s depends on missing task %s", ErrTaskGraphInvalid, task.TaskID, dependency)
			}
		}
	}
	if hasCycle(graph.Tasks) {
		return fmt.Errorf("%w: task graph has cycle", ErrTaskGraphInvalid)
	}
	return nil
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func hasCycle(tasks []Task) bool {
	visiting := map[string]bool{}
	visited := map[string]bool{}
	byID := make(map[string]Task, len(tasks))
	for _, task := range tasks {
		byID[task.TaskID] = task
	}

	var visit func(string) bool
	visit = func(taskID string) bool {
		if visiting[taskID] {
			return true
		}
		if visited[taskID] {
			return false
		}
		visiting[taskID] = true
		for _, dependency := range byID[taskID].DependsOn {
			if visit(dependency) {
				return true
			}
		}
		visiting[taskID] = false
		visited[taskID] = true
		return false
	}

	for _, task := range tasks {
		if visit(task.TaskID) {
			return true
		}
	}
	return false
}

func NormalizeModelOutput(raw json.RawMessage) (NormalizedOutput, error) {
	var output NormalizedOutput
	if err := json.Unmarshal(raw, &output); err != nil {
		return NormalizedOutput{}, fmt.Errorf("%w: %v", ErrNormalizationFailed, err)
	}
	if len(output.Artifacts) == 0 {
		return NormalizedOutput{}, fmt.Errorf("%w: artifacts are required", ErrNormalizationFailed)
	}
	for _, artifact := range output.Artifacts {
		if strings.TrimSpace(artifact.ArtifactID) == "" ||
			strings.TrimSpace(artifact.FileName) == "" ||
			strings.TrimSpace(artifact.Kind) == "" ||
			strings.TrimSpace(artifact.Title) == "" ||
			strings.TrimSpace(artifact.Content) == "" {
			return NormalizedOutput{}, fmt.Errorf("%w: artifact draft is incomplete", ErrNormalizationFailed)
		}
	}
	if strings.TrimSpace(output.NextBestAction) == "" {
		return NormalizedOutput{}, fmt.Errorf("%w: next_best_action is required", ErrNormalizationFailed)
	}
	return output, nil
}
