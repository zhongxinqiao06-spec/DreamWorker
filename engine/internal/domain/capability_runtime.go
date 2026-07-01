package domain

import "encoding/json"

const (
	EventCapabilityInvocationRequested = "capability.invocation_requested"
	EventCapabilityPolicyEvaluated     = "capability.policy_evaluated"
	EventCapabilityInvocationSucceeded = "capability.invocation_succeeded"
	EventCapabilityInvocationFailed    = "capability.invocation_failed"
	EventApprovalRequested             = "approval.requested"
	EventApprovalResolved              = "approval.resolved"
)

type PolicyRequest struct {
	PolicyID     string
	TraceID      string
	Action       string
	Actor        string
	CapabilityID string
	Record       CapabilityRecord
	RiskActions  []RiskAction
}

type PolicyDecision struct {
	SchemaVersion string       `json:"schema_version"`
	ID            string       `json:"id"`
	TraceID       string       `json:"trace_id"`
	Action        string       `json:"action"`
	CapabilityID  string       `json:"capability_id,omitempty"`
	Result        PolicyResult `json:"result"`
	Risk          RiskLevel    `json:"risk"`
	Reason        string       `json:"reason"`
}

type ApprovalRequest struct {
	SchemaVersion string         `json:"schema_version"`
	ApprovalID    string         `json:"approval_id"`
	TraceID       string         `json:"trace_id"`
	CapabilityID  string         `json:"capability_id"`
	Risk          RiskLevel      `json:"risk"`
	Status        ApprovalStatus `json:"status"`
	Reason        string         `json:"reason"`
	DiffSummary   string         `json:"diff_summary,omitempty"`
}

type ApprovalResolution struct {
	ApprovalID string         `json:"approval_id"`
	Status     ApprovalStatus `json:"status"`
	Reason     string         `json:"reason"`
}

type CapabilityInvocationRequest struct {
	MissionID    string
	RunID        string
	TraceID      string
	Actor        string
	CapabilityID string
	Input        json.RawMessage
	ApprovalID   string
}

type CapabilityInvocationResult struct {
	CapabilityID string
	TraceID      string
	OK           bool
	Output       json.RawMessage
	ErrorCode    string
	ErrorMessage string
	Approval     *ApprovalRequest
}

type CapabilityHandler func(CapabilityInvocationRequest) (CapabilityInvocationResult, error)

type CapabilityInvocationRequestedPayload struct {
	CapabilityID   string          `json:"capability_id"`
	LifecycleState LifecycleState  `json:"lifecycle_state"`
	TrustLevel     TrustLevel      `json:"trust_level"`
	RiskActions    []RiskAction    `json:"risk_actions"`
	Input          json.RawMessage `json:"input"`
	ApprovalID     string          `json:"approval_id,omitempty"`
}

type CapabilityPolicyEvaluatedPayload struct {
	CapabilityID string         `json:"capability_id"`
	Decision     PolicyDecision `json:"decision"`
}

type CapabilityInvocationFinishedPayload struct {
	CapabilityID string          `json:"capability_id"`
	OK           bool            `json:"ok"`
	Output       json.RawMessage `json:"output,omitempty"`
	ErrorCode    string          `json:"error_code,omitempty"`
	ErrorMessage string          `json:"error_message,omitempty"`
}
