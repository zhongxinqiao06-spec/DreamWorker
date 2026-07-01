package policy

import (
	"context"
	"testing"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/domain"
)

func TestPolicyAllowsTrustedBuiltinLowRisk(t *testing.T) {
	decision := evaluate(t, record(domain.CapabilityEnabled, domain.TrustTrustedBuiltin, domain.RiskLow, nil))

	assertDecision(t, decision, domain.PolicyAllow)
}

func TestPolicyDeniesRevokedCapability(t *testing.T) {
	decision := evaluate(t, record(domain.CapabilityRevoked, domain.TrustTrustedBuiltin, domain.RiskLow, nil))

	assertDecision(t, decision, domain.PolicyDeny)
}

func TestPolicyRequiresApprovalForExternalWrite(t *testing.T) {
	decision := evaluate(t, record(
		domain.CapabilityEnabled,
		domain.TrustTrustedBuiltin,
		domain.RiskMedium,
		[]domain.RiskAction{domain.RiskExternalWrite},
	))

	assertDecision(t, decision, domain.PolicyRequiresApproval)
}

func TestPolicyDeniesSecretAccessForUnverifiedCapability(t *testing.T) {
	decision := evaluate(t, record(
		domain.CapabilityEnabled,
		domain.TrustLocalUnverified,
		domain.RiskCritical,
		[]domain.RiskAction{domain.RiskSecretAccess},
	))

	assertDecision(t, decision, domain.PolicyDeny)
}

func TestPolicyDeniesRemoteUntrustedByDefault(t *testing.T) {
	decision := evaluate(t, record(domain.CapabilityEnabled, domain.TrustRemoteUntrusted, domain.RiskLow, nil))

	assertDecision(t, decision, domain.PolicyDeny)
}

func evaluate(t *testing.T, capability domain.CapabilityRecord) domain.PolicyDecision {
	t.Helper()
	decision, err := NewEngine().Evaluate(context.Background(), domain.PolicyRequest{
		PolicyID:     "pol_001",
		TraceID:      "tr_policy",
		Action:       "invoke_capability",
		Actor:        "test",
		CapabilityID: capability.Manifest.Metadata.ID,
		Record:       capability,
		RiskActions:  capability.RiskActions,
	})
	if err != nil {
		t.Fatalf("evaluate policy: %v", err)
	}
	return decision
}

func assertDecision(t *testing.T, decision domain.PolicyDecision, expected domain.PolicyResult) {
	t.Helper()
	if decision.Result != expected {
		t.Fatalf("expected %s, got %s: %#v", expected, decision.Result, decision)
	}
	if decision.SchemaVersion != domain.ContractSchemaVersion {
		t.Fatalf("expected schema version %s, got %s", domain.ContractSchemaVersion, decision.SchemaVersion)
	}
	if decision.Reason == "" {
		t.Fatal("expected policy reason")
	}
}

func record(
	state domain.LifecycleState,
	trust domain.TrustLevel,
	risk domain.RiskLevel,
	actions []domain.RiskAction,
) domain.CapabilityRecord {
	return domain.CapabilityRecord{
		Manifest: domain.CapabilityManifest{
			APIVersion: domain.CapabilityAPIVersion,
			Kind:       domain.CapabilityKindBuiltin,
			Metadata: domain.CapabilityMetadata{
				ID:       "cap_policy",
				Name:     "Policy Capability",
				Version:  "0.1.0",
				Provider: "test",
			},
			Protocol:      domain.CapabilityProtocol{Type: domain.CapabilityProtocolBuiltin},
			InputSchema:   map[string]any{"type": "object"},
			OutputSchema:  map[string]any{"type": "object"},
			Permissions:   map[string]any{},
			Risk:          domain.CapabilityRisk{Level: risk},
			Approval:      map[string]any{},
			Runtime:       map[string]any{},
			Observability: map[string]any{},
		},
		Lifecycle:   state,
		TrustLevel:  trust,
		RiskLevel:   risk,
		RiskActions: actions,
	}
}
