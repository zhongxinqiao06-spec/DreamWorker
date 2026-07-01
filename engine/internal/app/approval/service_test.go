package approval

import (
	"context"
	"testing"
	"time"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/domain"
)

func TestApprovalRequestAndResolveReplay(t *testing.T) {
	service, _ := newTestService()
	ctx := context.Background()

	request, err := service.Request(ctx, RequestCommand{
		MissionID:    "msn_001",
		RunID:        "run_001",
		TraceID:      "tr_approval",
		CapabilityID: "cap_external_write",
		Risk:         domain.RiskHigh,
		Reason:       "external write requires approval",
		DiffSummary:  "将写入外部系统。",
	})
	if err != nil {
		t.Fatalf("request approval: %v", err)
	}
	if request.Status != domain.ApprovalPending {
		t.Fatalf("expected pending, got %s", request.Status)
	}

	resolved, err := service.Resolve(ctx, ResolveCommand{
		MissionID:  "msn_001",
		RunID:      "run_001",
		TraceID:    "tr_approval",
		ApprovalID: request.ApprovalID,
		Status:     domain.ApprovalApproved,
		Reason:     "用户批准。",
	})
	if err != nil {
		t.Fatalf("resolve approval: %v", err)
	}
	if resolved.Status != domain.ApprovalApproved {
		t.Fatalf("expected approved, got %s", resolved.Status)
	}

	replayed, err := service.GetApproval(ctx, "msn_001", request.ApprovalID)
	if err != nil {
		t.Fatalf("get replayed approval: %v", err)
	}
	if replayed.Status != domain.ApprovalApproved {
		t.Fatalf("expected replayed approved, got %s", replayed.Status)
	}
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

func (ids *deterministicIDs) NewID(prefix string) string {
	ids.counts[prefix]++
	return prefix + "_001"
}

func newTestService() (*Service, *memoryEventStore) {
	store := &memoryEventStore{eventsByMission: map[string][]domain.DomainEvent{}}
	return NewService(store, fakeClock{}, &deterministicIDs{counts: map[string]int{}}), store
}
