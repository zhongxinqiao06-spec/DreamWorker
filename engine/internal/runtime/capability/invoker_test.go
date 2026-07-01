package capability

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	sqliteadapter "github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/adapters/sqlite"
	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/policy"
	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/domain"
)

func TestInvokeRejectsUnregisteredCapability(t *testing.T) {
	invoker := newTestInvoker()
	result, err := invoker.Invoke(context.Background(), invocation("cap_missing"))

	if !errors.Is(err, domain.ErrCapabilityNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
	if result.ErrorCode != "CAPABILITY_NOT_FOUND" {
		t.Fatalf("expected not found result, got %#v", result)
	}
	if len(invoker.events.eventsByMission["msn_001"]) != 1 {
		t.Fatalf("expected failed event, got %d", len(invoker.events.eventsByMission["msn_001"]))
	}
}

func TestInvokeRejectsRevokedCapability(t *testing.T) {
	invoker := newTestInvoker()
	invoker.registry.records["cap_revoked"] = capabilityRecord(
		"cap_revoked",
		domain.CapabilityRevoked,
		domain.TrustTrustedBuiltin,
		domain.RiskLow,
		nil,
	)

	result, err := invoker.Invoke(context.Background(), invocation("cap_revoked"))

	if !errors.Is(err, domain.ErrCapabilityRevoked) {
		t.Fatalf("expected revoked error, got %v", err)
	}
	if result.ErrorCode != "CAPABILITY_REVOKED" {
		t.Fatalf("expected revoked result, got %#v", result)
	}
	if invoker.handlerCalls != 0 {
		t.Fatal("expected revoked capability to avoid handler execution")
	}
}

func TestInvokeRequiresApprovalForHighRiskAction(t *testing.T) {
	invoker := newTestInvoker()
	invoker.registry.records["cap_external"] = capabilityRecord(
		"cap_external",
		domain.CapabilityEnabled,
		domain.TrustTrustedBuiltin,
		domain.RiskMedium,
		[]domain.RiskAction{domain.RiskExternalWrite},
	)

	result, err := invoker.Invoke(context.Background(), invocation("cap_external"))

	if !errors.Is(err, domain.ErrApprovalRequired) {
		t.Fatalf("expected approval required error, got %v", err)
	}
	if result.Approval == nil || result.Approval.Status != domain.ApprovalPending {
		t.Fatalf("expected pending approval, got %#v", result.Approval)
	}
	if invoker.handlerCalls != 0 {
		t.Fatal("expected approval gate to block handler execution")
	}
}

func TestInvokeContinuesAfterApprovedApproval(t *testing.T) {
	invoker := newTestInvoker()
	invoker.registry.records["cap_external"] = capabilityRecord(
		"cap_external",
		domain.CapabilityEnabled,
		domain.TrustTrustedBuiltin,
		domain.RiskMedium,
		[]domain.RiskAction{domain.RiskExternalWrite},
	)
	invoker.approvals.approvals["apr_approved"] = domain.ApprovalRequest{
		ApprovalID:   "apr_approved",
		CapabilityID: "cap_external",
		Status:       domain.ApprovalApproved,
	}
	request := invocation("cap_external")
	request.ApprovalID = "apr_approved"

	result, err := invoker.Invoke(context.Background(), request)

	if err != nil {
		t.Fatalf("invoke with approval: %v", err)
	}
	if !result.OK {
		t.Fatalf("expected ok result, got %#v", result)
	}
	if invoker.handlerCalls != 1 {
		t.Fatalf("expected one handler call, got %d", invoker.handlerCalls)
	}
}

func TestInvokeStopsRejectedApproval(t *testing.T) {
	invoker := newTestInvoker()
	invoker.registry.records["cap_external"] = capabilityRecord(
		"cap_external",
		domain.CapabilityEnabled,
		domain.TrustTrustedBuiltin,
		domain.RiskMedium,
		[]domain.RiskAction{domain.RiskExternalWrite},
	)
	invoker.approvals.approvals["apr_rejected"] = domain.ApprovalRequest{
		ApprovalID:   "apr_rejected",
		CapabilityID: "cap_external",
		Status:       domain.ApprovalRejected,
	}
	request := invocation("cap_external")
	request.ApprovalID = "apr_rejected"

	result, err := invoker.Invoke(context.Background(), request)

	if err == nil {
		t.Fatal("expected rejected approval error")
	}
	if result.ErrorCode != "APPROVAL_REJECTED" {
		t.Fatalf("expected rejected approval result, got %#v", result)
	}
	if invoker.handlerCalls != 0 {
		t.Fatal("expected rejected approval to avoid handler execution")
	}
}

func TestArtifactWriteBuiltinStaysInsideProject(t *testing.T) {
	ctx := context.Background()
	db, err := sqliteadapter.Open(ctx, filepath.Join(t.TempDir(), "engine.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	if err := sqliteadapter.Bootstrap(ctx, db); err != nil {
		t.Fatalf("bootstrap sqlite: %v", err)
	}
	projectDir := t.TempDir()
	artifactStore, err := sqliteadapter.NewArtifactStore(db, projectDir)
	if err != nil {
		t.Fatalf("new artifact store: %v", err)
	}
	registry := &fakeRegistry{records: map[string]domain.CapabilityRecord{
		BuiltinArtifactWrite: capabilityRecord(
			BuiltinArtifactWrite,
			domain.CapabilityEnabled,
			domain.TrustTrustedBuiltin,
			domain.RiskLow,
			nil,
		),
	}}
	events := &memoryEventStore{eventsByMission: map[string][]domain.DomainEvent{}}
	invoker := NewInvoker(
		registry,
		policy.NewEngine(),
		&fakeApprovalStore{approvals: map[string]domain.ApprovalRequest{}},
		events,
		fakeClock{},
		newIDs(),
		BuiltinHandlers(artifactStore),
	)
	input, _ := json.Marshal(map[string]any{
		"artifact_id":  "art_boundary",
		"kind":         "dream_brief",
		"title":        "Dream Brief",
		"version":      1,
		"content_type": "text/markdown",
		"file_name":    "../evil.md",
		"content":      "# nope",
	})
	request := invocation(BuiltinArtifactWrite)
	request.Input = input

	result, err := invoker.Invoke(ctx, request)

	if err == nil {
		t.Fatal("expected artifact write boundary error")
	}
	if result.ErrorCode != "ARTIFACT_WRITE_FAILED" {
		t.Fatalf("expected artifact write failure, got %#v", result)
	}
	if _, statErr := os.Stat(filepath.Join(projectDir, "evil.md")); !os.IsNotExist(statErr) {
		t.Fatalf("expected no file outside project artifacts, stat=%v", statErr)
	}
}

func TestRegisterBuiltinsEnablesMVPBuiltins(t *testing.T) {
	ctx := context.Background()
	db, err := sqliteadapter.Open(ctx, filepath.Join(t.TempDir(), "engine.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	if err := sqliteadapter.Bootstrap(ctx, db); err != nil {
		t.Fatalf("bootstrap sqlite: %v", err)
	}
	registry := sqliteadapter.NewCapabilityRegistry(db)

	if err := RegisterBuiltins(ctx, registry); err != nil {
		t.Fatalf("register builtins: %v", err)
	}
	enabled, err := registry.ListEnabled(ctx)
	if err != nil {
		t.Fatalf("list enabled: %v", err)
	}
	if len(enabled) != 6 {
		t.Fatalf("expected six enabled MVP builtins, got %d", len(enabled))
	}
}

type testInvoker struct {
	*Invoker
	registry     *fakeRegistry
	approvals    *fakeApprovalStore
	events       *memoryEventStore
	handlerCalls int
}

func newTestInvoker() *testInvoker {
	registry := &fakeRegistry{records: map[string]domain.CapabilityRecord{}}
	approvals := &fakeApprovalStore{approvals: map[string]domain.ApprovalRequest{}}
	events := &memoryEventStore{eventsByMission: map[string][]domain.DomainEvent{}}
	wrapper := &testInvoker{registry: registry, approvals: approvals, events: events}
	wrapper.Invoker = NewInvoker(
		registry,
		policy.NewEngine(),
		approvals,
		events,
		fakeClock{},
		newIDs(),
		map[string]domain.CapabilityHandler{
			"cap_external": func(request domain.CapabilityInvocationRequest) (domain.CapabilityInvocationResult, error) {
				wrapper.handlerCalls++
				return domain.CapabilityInvocationResult{OK: true, Output: JSONOutput(map[string]string{"ok": "true"})}, nil
			},
		},
	)
	return wrapper
}

type fakeRegistry struct {
	records map[string]domain.CapabilityRecord
}

func (registry *fakeRegistry) Discover(
	context.Context,
	domain.CapabilityManifest,
	domain.TrustLevel,
) (domain.CapabilityRecord, error) {
	return domain.CapabilityRecord{}, nil
}

func (registry *fakeRegistry) Transition(
	context.Context,
	string,
	domain.LifecycleState,
) (domain.CapabilityRecord, error) {
	return domain.CapabilityRecord{}, nil
}

func (registry *fakeRegistry) Get(_ context.Context, capabilityID string) (domain.CapabilityRecord, error) {
	record, ok := registry.records[capabilityID]
	if !ok {
		return domain.CapabilityRecord{}, domain.ErrCapabilityNotFound
	}
	return record, nil
}

func (registry *fakeRegistry) ListEnabled(context.Context) ([]domain.CapabilityRecord, error) {
	return nil, nil
}

type fakeApprovalStore struct {
	approvals map[string]domain.ApprovalRequest
}

func (store *fakeApprovalStore) GetApproval(
	_ context.Context,
	_ string,
	approvalID string,
) (domain.ApprovalRequest, error) {
	approval, ok := store.approvals[approvalID]
	if !ok {
		return domain.ApprovalRequest{}, domain.ErrApprovalRequired
	}
	return approval, nil
}

type memoryEventStore struct {
	eventsByMission map[string][]domain.DomainEvent
}

func (store *memoryEventStore) Append(_ context.Context, events []domain.DomainEvent) error {
	for _, event := range events {
		store.eventsByMission[event.MissionID] = append(store.eventsByMission[event.MissionID], event)
	}
	return nil
}

func (store *memoryEventStore) LoadMission(_ context.Context, missionID string) ([]domain.DomainEvent, error) {
	return append([]domain.DomainEvent{}, store.eventsByMission[missionID]...), nil
}

func (store *memoryEventStore) LoadRun(_ context.Context, runID string) ([]domain.DomainEvent, error) {
	var events []domain.DomainEvent
	for _, missionEvents := range store.eventsByMission {
		for _, event := range missionEvents {
			if event.RunID == runID {
				events = append(events, event)
			}
		}
	}
	return events, nil
}

type fakeClock struct{}

func (fakeClock) Now() time.Time {
	return time.Date(2026, 6, 30, 8, 0, 0, 0, time.UTC)
}

type deterministicIDs struct {
	counts map[string]int
}

func newIDs() *deterministicIDs {
	return &deterministicIDs{counts: map[string]int{}}
}

func (ids *deterministicIDs) NewID(prefix string) string {
	ids.counts[prefix]++
	switch prefix {
	case "pol":
		return "pol_001"
	case "apr":
		return "apr_001"
	default:
		return prefix + "_001"
	}
}

func invocation(capabilityID string) domain.CapabilityInvocationRequest {
	return domain.CapabilityInvocationRequest{
		MissionID:    "msn_001",
		RunID:        "run_001",
		TraceID:      "tr_invocation",
		Actor:        "test",
		CapabilityID: capabilityID,
		Input:        json.RawMessage(`{}`),
	}
}

func capabilityRecord(
	id string,
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
				ID:       id,
				Name:     id,
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
