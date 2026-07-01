package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const CapabilityAPIVersion = "capability.dreamworker.dev/v1"

var (
	ErrCapabilityNotFound          = errors.New("capability not found")
	ErrCapabilityNotEnabled        = errors.New("capability is not enabled")
	ErrCapabilityRevoked           = errors.New("capability is revoked")
	ErrInvalidCapabilityTransition = errors.New("invalid capability lifecycle transition")
	ErrApprovalRequired            = errors.New("approval required")
	ErrApprovalRejected            = errors.New("approval rejected")
)

type CapabilityKind string

const (
	CapabilityKindBuiltin  CapabilityKind = "builtin"
	CapabilityKindMCPTool  CapabilityKind = "mcp_tool"
	CapabilityKindA2AAgent CapabilityKind = "a2a_agent"
	CapabilityKindSkill    CapabilityKind = "skill"
	CapabilityKindOpenAPI  CapabilityKind = "openapi"
	CapabilityKindBrowser  CapabilityKind = "browser"
	CapabilityKindHuman    CapabilityKind = "human"
	CapabilityKindWebhook  CapabilityKind = "webhook"
	CapabilityKindModel    CapabilityKind = "model"
)

type CapabilityProtocolType string

const (
	CapabilityProtocolBuiltin CapabilityProtocolType = "builtin"
	CapabilityProtocolMCP     CapabilityProtocolType = "mcp"
	CapabilityProtocolA2A     CapabilityProtocolType = "a2a"
	CapabilityProtocolSkill   CapabilityProtocolType = "skill"
	CapabilityProtocolOpenAPI CapabilityProtocolType = "openapi"
	CapabilityProtocolBrowser CapabilityProtocolType = "browser"
	CapabilityProtocolHuman   CapabilityProtocolType = "human"
	CapabilityProtocolWebhook CapabilityProtocolType = "webhook"
	CapabilityProtocolModel   CapabilityProtocolType = "model"
)

type LifecycleState string

const (
	CapabilityDiscovered      LifecycleState = "discovered"
	CapabilityRegistered      LifecycleState = "registered"
	CapabilitySchemaValidated LifecycleState = "schema_validated"
	CapabilityRiskClassified  LifecycleState = "risk_classified"
	CapabilityAuthorized      LifecycleState = "authorized"
	CapabilityEnabled         LifecycleState = "enabled"
	CapabilityDisabled        LifecycleState = "disabled"
	CapabilityRevoked         LifecycleState = "revoked"
	CapabilityDeprecated      LifecycleState = "deprecated"
)

type TrustLevel string

const (
	TrustTrustedBuiltin  TrustLevel = "trusted_builtin"
	TrustVerifiedPartner TrustLevel = "verified_partner"
	TrustCommunity       TrustLevel = "community"
	TrustLocalUnverified TrustLevel = "local_unverified"
	TrustRemoteUntrusted TrustLevel = "remote_untrusted"
)

type RiskAction string

const (
	RiskExternalWrite           RiskAction = "external_write"
	RiskFileWriteOutsideProject RiskAction = "file_write_outside_project"
	RiskSecretAccess            RiskAction = "secret_access"
	RiskNetworkUntrusted        RiskAction = "network_untrusted"
	RiskPaidCall                RiskAction = "paid_call"
	RiskCodeExecution           RiskAction = "code_execution"
	RiskBrowserAction           RiskAction = "browser_action"
	RiskPublishContent          RiskAction = "publish_content"
	RiskSendEmail               RiskAction = "send_email"
	RiskInstallSkill            RiskAction = "install_skill"
	RiskConnectRemoteMCP        RiskAction = "connect_remote_mcp"
)

type PolicyResult string

const (
	PolicyAllow            PolicyResult = "allow"
	PolicyDeny             PolicyResult = "deny"
	PolicyRequiresApproval PolicyResult = "requires_approval"
)

type ApprovalStatus string

const (
	ApprovalPending   ApprovalStatus = "pending"
	ApprovalApproved  ApprovalStatus = "approved"
	ApprovalRejected  ApprovalStatus = "rejected"
	ApprovalEdited    ApprovalStatus = "edited"
	ApprovalCancelled ApprovalStatus = "cancelled"
)

type CapabilityManifest struct {
	APIVersion    string             `json:"apiVersion"`
	Kind          CapabilityKind     `json:"kind"`
	Metadata      CapabilityMetadata `json:"metadata"`
	Protocol      CapabilityProtocol `json:"protocol"`
	InputSchema   map[string]any     `json:"inputSchema"`
	OutputSchema  map[string]any     `json:"outputSchema"`
	Permissions   map[string]any     `json:"permissions"`
	Risk          CapabilityRisk     `json:"risk"`
	Approval      map[string]any     `json:"approval"`
	Runtime       map[string]any     `json:"runtime"`
	Observability map[string]any     `json:"observability"`
}

type CapabilityMetadata struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	Provider string `json:"provider"`
}

type CapabilityProtocol struct {
	Type      CapabilityProtocolType `json:"type"`
	Transport string                 `json:"transport,omitempty"`
}

type CapabilityRisk struct {
	Level   RiskLevel `json:"level"`
	Reasons []string  `json:"reasons"`
}

type CapabilityRecord struct {
	Manifest       CapabilityManifest
	Lifecycle      LifecycleState
	TrustLevel     TrustLevel
	RiskLevel      RiskLevel
	RiskActions    []RiskAction
	CreatedAt      time.Time
	UpdatedAt      time.Time
	LastTransition string
}

func (manifest CapabilityManifest) Validate() error {
	switch {
	case manifest.APIVersion != CapabilityAPIVersion:
		return fmt.Errorf("capability apiVersion must be %s", CapabilityAPIVersion)
	case !manifest.Kind.IsValid():
		return fmt.Errorf("invalid capability kind %s", manifest.Kind)
	case strings.TrimSpace(manifest.Metadata.ID) == "":
		return errors.New("capability metadata.id is required")
	case !strings.HasPrefix(manifest.Metadata.ID, "cap_"):
		return errors.New("capability metadata.id must start with cap_")
	case strings.TrimSpace(manifest.Metadata.Name) == "":
		return errors.New("capability metadata.name is required")
	case strings.TrimSpace(manifest.Metadata.Version) == "":
		return errors.New("capability metadata.version is required")
	case strings.TrimSpace(manifest.Metadata.Provider) == "":
		return errors.New("capability metadata.provider is required")
	case !manifest.Protocol.Type.IsValid():
		return fmt.Errorf("invalid capability protocol %s", manifest.Protocol.Type)
	case manifest.InputSchema == nil:
		return errors.New("capability inputSchema is required")
	case manifest.OutputSchema == nil:
		return errors.New("capability outputSchema is required")
	case manifest.Permissions == nil:
		return errors.New("capability permissions is required")
	case !manifest.Risk.Level.IsValid():
		return fmt.Errorf("invalid capability risk level %s", manifest.Risk.Level)
	case manifest.Approval == nil:
		return errors.New("capability approval is required")
	case manifest.Runtime == nil:
		return errors.New("capability runtime is required")
	case manifest.Observability == nil:
		return errors.New("capability observability is required")
	default:
		return nil
	}
}

func (kind CapabilityKind) IsValid() bool {
	switch kind {
	case CapabilityKindBuiltin,
		CapabilityKindMCPTool,
		CapabilityKindA2AAgent,
		CapabilityKindSkill,
		CapabilityKindOpenAPI,
		CapabilityKindBrowser,
		CapabilityKindHuman,
		CapabilityKindWebhook,
		CapabilityKindModel:
		return true
	default:
		return false
	}
}

func (protocol CapabilityProtocolType) IsValid() bool {
	switch protocol {
	case CapabilityProtocolBuiltin,
		CapabilityProtocolMCP,
		CapabilityProtocolA2A,
		CapabilityProtocolSkill,
		CapabilityProtocolOpenAPI,
		CapabilityProtocolBrowser,
		CapabilityProtocolHuman,
		CapabilityProtocolWebhook,
		CapabilityProtocolModel:
		return true
	default:
		return false
	}
}

func (risk RiskLevel) IsValid() bool {
	switch risk {
	case RiskLow, RiskMedium, RiskHigh, RiskCritical:
		return true
	default:
		return false
	}
}

func (state LifecycleState) CanTransitionTo(next LifecycleState) bool {
	if state == CapabilityRevoked {
		return false
	}
	if next == CapabilityRevoked || next == CapabilityDeprecated || next == CapabilityDisabled {
		return true
	}

	allowed := map[LifecycleState][]LifecycleState{
		"":                        {CapabilityDiscovered},
		CapabilityDiscovered:      {CapabilityRegistered},
		CapabilityRegistered:      {CapabilitySchemaValidated},
		CapabilitySchemaValidated: {CapabilityRiskClassified},
		CapabilityRiskClassified:  {CapabilityAuthorized},
		CapabilityAuthorized:      {CapabilityEnabled},
		CapabilityDisabled:        {CapabilityEnabled, CapabilityRevoked},
		CapabilityDeprecated:      {CapabilityRevoked},
	}
	for _, candidate := range allowed[state] {
		if candidate == next {
			return true
		}
	}
	return false
}

func (record CapabilityRecord) CanInvoke() error {
	switch {
	case record.Lifecycle == CapabilityRevoked:
		return ErrCapabilityRevoked
	case record.Lifecycle != CapabilityEnabled:
		return ErrCapabilityNotEnabled
	default:
		return nil
	}
}

func (manifest CapabilityManifest) MarshalJSONRaw() (json.RawMessage, error) {
	data, err := json.Marshal(manifest)
	if err != nil {
		return nil, fmt.Errorf("marshal capability manifest: %w", err)
	}
	return data, nil
}
