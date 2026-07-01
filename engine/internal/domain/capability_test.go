package domain

import "testing"

func TestCapabilityManifestValidation(t *testing.T) {
	manifest := testCapabilityManifest("cap_test")

	if err := manifest.Validate(); err != nil {
		t.Fatalf("validate capability manifest: %v", err)
	}
}

func TestCapabilityLifecycleTransitions(t *testing.T) {
	state := CapabilityDiscovered
	for _, next := range []LifecycleState{
		CapabilityRegistered,
		CapabilitySchemaValidated,
		CapabilityRiskClassified,
		CapabilityAuthorized,
		CapabilityEnabled,
	} {
		if !state.CanTransitionTo(next) {
			t.Fatalf("expected transition %s -> %s", state, next)
		}
		state = next
	}
	if CapabilityRegistered.CanTransitionTo(CapabilityEnabled) {
		t.Fatal("expected schema_validated/risk_classified/authorized before enabled")
	}
	if CapabilityRevoked.CanTransitionTo(CapabilityEnabled) {
		t.Fatal("expected revoked to be terminal")
	}
}

func TestCapabilityCanInvokeRequiresEnabled(t *testing.T) {
	record := CapabilityRecord{Manifest: testCapabilityManifest("cap_test"), Lifecycle: CapabilityRegistered}
	if err := record.CanInvoke(); err == nil {
		t.Fatal("expected non-enabled capability to be rejected")
	}
	record.Lifecycle = CapabilityRevoked
	if err := record.CanInvoke(); err != ErrCapabilityRevoked {
		t.Fatalf("expected revoked error, got %v", err)
	}
	record.Lifecycle = CapabilityEnabled
	if err := record.CanInvoke(); err != nil {
		t.Fatalf("expected enabled capability to be invokable: %v", err)
	}
}

func testCapabilityManifest(id string) CapabilityManifest {
	return CapabilityManifest{
		APIVersion: CapabilityAPIVersion,
		Kind:       CapabilityKindBuiltin,
		Metadata: CapabilityMetadata{
			ID:       id,
			Name:     "Test Capability",
			Version:  "0.1.0",
			Provider: "test",
		},
		Protocol:      CapabilityProtocol{Type: CapabilityProtocolBuiltin},
		InputSchema:   map[string]any{"type": "object"},
		OutputSchema:  map[string]any{"type": "object"},
		Permissions:   map[string]any{},
		Risk:          CapabilityRisk{Level: RiskLow, Reasons: []string{}},
		Approval:      map[string]any{},
		Runtime:       map[string]any{},
		Observability: map[string]any{},
	}
}
