package sqlite

import (
	"context"
	"errors"
	"testing"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/domain"
)

func TestCapabilityRegistryLifecycle(t *testing.T) {
	ctx := context.Background()
	db := openBootstrappedTestDB(t, ctx)
	registry := NewCapabilityRegistry(db)

	record, err := registry.Discover(ctx, testCapabilityManifest("cap_registry"), domain.TrustTrustedBuiltin)
	if err != nil {
		t.Fatalf("discover capability: %v", err)
	}
	if record.Lifecycle != domain.CapabilityDiscovered {
		t.Fatalf("expected discovered, got %s", record.Lifecycle)
	}

	for _, state := range []domain.LifecycleState{
		domain.CapabilityRegistered,
		domain.CapabilitySchemaValidated,
		domain.CapabilityRiskClassified,
		domain.CapabilityAuthorized,
		domain.CapabilityEnabled,
	} {
		record, err = registry.Transition(ctx, "cap_registry", state)
		if err != nil {
			t.Fatalf("transition to %s: %v", state, err)
		}
	}
	if record.Lifecycle != domain.CapabilityEnabled {
		t.Fatalf("expected enabled, got %s", record.Lifecycle)
	}

	enabled, err := registry.ListEnabled(ctx)
	if err != nil {
		t.Fatalf("list enabled: %v", err)
	}
	if len(enabled) != 1 || enabled[0].Manifest.Metadata.ID != "cap_registry" {
		t.Fatalf("unexpected enabled capabilities: %#v", enabled)
	}
}

func TestCapabilityRegistryRejectsEnableBeforeValidation(t *testing.T) {
	ctx := context.Background()
	db := openBootstrappedTestDB(t, ctx)
	registry := NewCapabilityRegistry(db)

	if _, err := registry.Discover(ctx, testCapabilityManifest("cap_unvalidated"), domain.TrustTrustedBuiltin); err != nil {
		t.Fatalf("discover capability: %v", err)
	}
	if _, err := registry.Transition(ctx, "cap_unvalidated", domain.CapabilityEnabled); err == nil {
		t.Fatal("expected invalid transition error")
	}
}

func TestCapabilityRegistryReturnsNotFound(t *testing.T) {
	ctx := context.Background()
	db := openBootstrappedTestDB(t, ctx)
	registry := NewCapabilityRegistry(db)

	if _, err := registry.Get(ctx, "cap_missing"); !errors.Is(err, domain.ErrCapabilityNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func testCapabilityManifest(id string) domain.CapabilityManifest {
	return domain.CapabilityManifest{
		APIVersion: domain.CapabilityAPIVersion,
		Kind:       domain.CapabilityKindBuiltin,
		Metadata: domain.CapabilityMetadata{
			ID:       id,
			Name:     "Registry Capability",
			Version:  "0.1.0",
			Provider: "test",
		},
		Protocol:      domain.CapabilityProtocol{Type: domain.CapabilityProtocolBuiltin},
		InputSchema:   map[string]any{"type": "object"},
		OutputSchema:  map[string]any{"type": "object"},
		Permissions:   map[string]any{},
		Risk:          domain.CapabilityRisk{Level: domain.RiskLow, Reasons: []string{}},
		Approval:      map[string]any{},
		Runtime:       map[string]any{},
		Observability: map[string]any{},
	}
}
