package approval

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/domain"
	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/ports"
)

var _ ports.ApprovalStore = (*Service)(nil)

type Service struct {
	store ports.EventStore
	clock ports.Clock
	ids   ports.IdGenerator
}

func NewService(store ports.EventStore, clock ports.Clock, ids ports.IdGenerator) *Service {
	return &Service{store: store, clock: clock, ids: ids}
}

type RequestCommand struct {
	MissionID    string
	RunID        string
	TraceID      string
	CapabilityID string
	Risk         domain.RiskLevel
	Reason       string
	DiffSummary  string
}

type ResolveCommand struct {
	MissionID  string
	RunID      string
	TraceID    string
	ApprovalID string
	Status     domain.ApprovalStatus
	Reason     string
}

func (service *Service) Request(ctx context.Context, command RequestCommand) (domain.ApprovalRequest, error) {
	if strings.TrimSpace(command.CapabilityID) == "" {
		return domain.ApprovalRequest{}, errors.New("capability id is required")
	}
	approval := domain.ApprovalRequest{
		SchemaVersion: domain.ContractSchemaVersion,
		ApprovalID:    service.ids.NewID("apr"),
		TraceID:       command.TraceID,
		CapabilityID:  command.CapabilityID,
		Risk:          command.Risk,
		Status:        domain.ApprovalPending,
		Reason:        command.Reason,
		DiffSummary:   command.DiffSummary,
	}
	event, err := service.event(
		command.MissionID,
		command.RunID,
		command.TraceID,
		domain.EventApprovalRequested,
		approval,
	)
	if err != nil {
		return domain.ApprovalRequest{}, err
	}
	if err := service.store.Append(ctx, []domain.DomainEvent{event}); err != nil {
		return domain.ApprovalRequest{}, err
	}
	return approval, nil
}

func (service *Service) Resolve(ctx context.Context, command ResolveCommand) (domain.ApprovalRequest, error) {
	current, err := service.GetApproval(ctx, command.MissionID, command.ApprovalID)
	if err != nil {
		return domain.ApprovalRequest{}, err
	}
	if current.Status != domain.ApprovalPending {
		return domain.ApprovalRequest{}, fmt.Errorf("approval %s is not pending", command.ApprovalID)
	}
	resolution := domain.ApprovalResolution{
		ApprovalID: command.ApprovalID,
		Status:     command.Status,
		Reason:     command.Reason,
	}
	event, err := service.event(
		command.MissionID,
		command.RunID,
		command.TraceID,
		domain.EventApprovalResolved,
		resolution,
	)
	if err != nil {
		return domain.ApprovalRequest{}, err
	}
	if err := service.store.Append(ctx, []domain.DomainEvent{event}); err != nil {
		return domain.ApprovalRequest{}, err
	}
	current.Status = command.Status
	current.Reason = command.Reason
	return current, nil
}

func (service *Service) GetApproval(
	ctx context.Context,
	missionID string,
	approvalID string,
) (domain.ApprovalRequest, error) {
	events, err := service.store.LoadMission(ctx, missionID)
	if err != nil {
		return domain.ApprovalRequest{}, err
	}
	var found bool
	var approval domain.ApprovalRequest
	for _, event := range events {
		switch event.Type {
		case domain.EventApprovalRequested:
			var request domain.ApprovalRequest
			if err := decodeEventPayload(event, &request); err != nil {
				return domain.ApprovalRequest{}, err
			}
			if request.ApprovalID == approvalID {
				approval = request
				found = true
			}
		case domain.EventApprovalResolved:
			var resolution domain.ApprovalResolution
			if err := decodeEventPayload(event, &resolution); err != nil {
				return domain.ApprovalRequest{}, err
			}
			if resolution.ApprovalID == approvalID && found {
				approval.Status = resolution.Status
				approval.Reason = resolution.Reason
			}
		}
	}
	if !found {
		return domain.ApprovalRequest{}, fmt.Errorf("approval %s not found", approvalID)
	}
	return approval, nil
}

func (service *Service) event(
	missionID string,
	runID string,
	traceID string,
	eventType string,
	payload any,
) (domain.DomainEvent, error) {
	encodedPayload, err := domain.EncodePayload(payload)
	if err != nil {
		return domain.DomainEvent{}, err
	}
	event := domain.DomainEvent{
		EventID:       service.ids.NewID("evt"),
		SchemaVersion: domain.EventSchemaVersion,
		TraceID:       traceID,
		MissionID:     missionID,
		RunID:         runID,
		Actor:         "approval",
		Timestamp:     service.clock.Now(),
		Type:          eventType,
		Payload:       encodedPayload,
	}
	if err := event.Validate(); err != nil {
		return domain.DomainEvent{}, err
	}
	return event, nil
}

func decodeEventPayload(event domain.DomainEvent, target any) error {
	if err := event.Validate(); err != nil {
		return err
	}
	if err := json.Unmarshal(event.Payload, target); err != nil {
		return fmt.Errorf("decode %s payload: %w", event.Type, err)
	}
	return nil
}
