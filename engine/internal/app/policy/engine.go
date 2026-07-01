package policy

import (
	"context"
	"strings"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/domain"
	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/ports"
)

var _ ports.PolicyEngine = (*Engine)(nil)

type Engine struct{}

func NewEngine() *Engine {
	return &Engine{}
}

func (engine *Engine) Evaluate(
	_ context.Context,
	request domain.PolicyRequest,
) (domain.PolicyDecision, error) {
	decision := domain.PolicyDecision{
		SchemaVersion: domain.ContractSchemaVersion,
		ID:            request.PolicyID,
		TraceID:       request.TraceID,
		Action:        request.Action,
		CapabilityID:  request.CapabilityID,
		Risk:          request.Record.RiskLevel,
	}

	if strings.TrimSpace(request.CapabilityID) == "" || request.Record.Manifest.Metadata.ID == "" {
		decision.Result = domain.PolicyDeny
		decision.Risk = domain.RiskCritical
		decision.Reason = "capability is missing or unregistered"
		return decision, nil
	}
	if request.Record.Lifecycle == domain.CapabilityRevoked {
		decision.Result = domain.PolicyDeny
		decision.Reason = "capability is revoked"
		return decision, nil
	}
	if request.Record.Lifecycle != domain.CapabilityEnabled {
		decision.Result = domain.PolicyDeny
		decision.Reason = "capability is not enabled"
		return decision, nil
	}
	if request.Record.TrustLevel == domain.TrustRemoteUntrusted {
		decision.Result = domain.PolicyDeny
		decision.Reason = "remote_untrusted capability is denied by default"
		return decision, nil
	}
	if hasDeniedAction(request.Record.TrustLevel, request.RiskActions) {
		decision.Result = domain.PolicyDeny
		decision.Risk = maxRisk(decision.Risk, domain.RiskCritical)
		decision.Reason = "risk action denied by trust level"
		return decision, nil
	}
	if hasApprovalAction(request.RiskActions) {
		decision.Result = domain.PolicyRequiresApproval
		decision.Reason = "risk action requires approval"
		return decision, nil
	}
	if request.Record.TrustLevel == domain.TrustTrustedBuiltin && request.Record.RiskLevel == domain.RiskLow {
		decision.Result = domain.PolicyAllow
		decision.Reason = "trusted builtin low-risk capability"
		return decision, nil
	}

	decision.Result = domain.PolicyRequiresApproval
	decision.Reason = "non-low-risk capability requires approval"
	return decision, nil
}

func hasApprovalAction(actions []domain.RiskAction) bool {
	for _, action := range actions {
		switch action {
		case domain.RiskExternalWrite,
			domain.RiskFileWriteOutsideProject,
			domain.RiskNetworkUntrusted,
			domain.RiskPaidCall,
			domain.RiskBrowserAction,
			domain.RiskPublishContent,
			domain.RiskSendEmail,
			domain.RiskInstallSkill:
			return true
		}
	}
	return false
}

func hasDeniedAction(trustLevel domain.TrustLevel, actions []domain.RiskAction) bool {
	for _, action := range actions {
		switch action {
		case domain.RiskSecretAccess, domain.RiskCodeExecution, domain.RiskConnectRemoteMCP:
			if trustLevel == domain.TrustLocalUnverified || trustLevel == domain.TrustRemoteUntrusted {
				return true
			}
		}
	}
	return false
}

func maxRisk(left domain.RiskLevel, right domain.RiskLevel) domain.RiskLevel {
	if riskWeight(right) > riskWeight(left) {
		return right
	}
	return left
}

func riskWeight(risk domain.RiskLevel) int {
	switch risk {
	case domain.RiskLow:
		return 1
	case domain.RiskMedium:
		return 2
	case domain.RiskHigh:
		return 3
	case domain.RiskCritical:
		return 4
	default:
		return 0
	}
}
