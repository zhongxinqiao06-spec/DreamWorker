package capability

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/domain"
	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/ports"
)

var _ ports.CapabilityInvoker = (*Invoker)(nil)

type Invoker struct {
	registry  ports.CapabilityRegistry
	policy    ports.PolicyEngine
	approvals ports.ApprovalStore
	events    ports.EventStore
	clock     ports.Clock
	ids       ports.IdGenerator
	handlers  map[string]domain.CapabilityHandler
}

func NewInvoker(
	registry ports.CapabilityRegistry,
	policy ports.PolicyEngine,
	approvals ports.ApprovalStore,
	events ports.EventStore,
	clock ports.Clock,
	ids ports.IdGenerator,
	handlers map[string]domain.CapabilityHandler,
) *Invoker {
	return &Invoker{
		registry:  registry,
		policy:    policy,
		approvals: approvals,
		events:    events,
		clock:     clock,
		ids:       ids,
		handlers:  handlers,
	}
}

func (invoker *Invoker) Invoke(
	ctx context.Context,
	request domain.CapabilityInvocationRequest,
) (domain.CapabilityInvocationResult, error) {
	record, err := invoker.registry.Get(ctx, request.CapabilityID)
	if err != nil {
		result := failedResult(request, "CAPABILITY_NOT_FOUND", "capability is not registered")
		_ = invoker.appendFinished(ctx, request, domain.EventCapabilityInvocationFailed, result)
		return result, err
	}

	requestedEvent, err := invoker.event(
		request,
		domain.EventCapabilityInvocationRequested,
		domain.CapabilityInvocationRequestedPayload{
			CapabilityID:   request.CapabilityID,
			LifecycleState: record.Lifecycle,
			TrustLevel:     record.TrustLevel,
			RiskActions:    record.RiskActions,
			Input:          request.Input,
			ApprovalID:     request.ApprovalID,
		},
	)
	if err != nil {
		return domain.CapabilityInvocationResult{}, err
	}

	if lifecycleErr := record.CanInvoke(); lifecycleErr != nil {
		result := failedResult(request, errorCode(lifecycleErr), lifecycleErr.Error())
		failedEvent, eventErr := invoker.finishedEvent(request, domain.EventCapabilityInvocationFailed, result)
		if eventErr != nil {
			return result, eventErr
		}
		_ = invoker.events.Append(ctx, []domain.DomainEvent{requestedEvent, failedEvent})
		return result, lifecycleErr
	}

	decision, err := invoker.policy.Evaluate(ctx, domain.PolicyRequest{
		PolicyID:     invoker.ids.NewID("pol"),
		TraceID:      request.TraceID,
		Action:       "invoke_capability",
		Actor:        request.Actor,
		CapabilityID: request.CapabilityID,
		Record:       record,
		RiskActions:  record.RiskActions,
	})
	if err != nil {
		return domain.CapabilityInvocationResult{}, err
	}
	policyEvent, err := invoker.event(
		request,
		domain.EventCapabilityPolicyEvaluated,
		domain.CapabilityPolicyEvaluatedPayload{CapabilityID: request.CapabilityID, Decision: decision},
	)
	if err != nil {
		return domain.CapabilityInvocationResult{}, err
	}

	switch decision.Result {
	case domain.PolicyDeny:
		result := failedResult(request, "POLICY_DENIED", decision.Reason)
		failedEvent, eventErr := invoker.finishedEvent(request, domain.EventCapabilityInvocationFailed, result)
		if eventErr != nil {
			return result, eventErr
		}
		_ = invoker.events.Append(ctx, []domain.DomainEvent{requestedEvent, policyEvent, failedEvent})
		return result, errors.New(decision.Reason)
	case domain.PolicyRequiresApproval:
		if request.ApprovalID == "" {
			approval := domain.ApprovalRequest{
				SchemaVersion: domain.ContractSchemaVersion,
				ApprovalID:    invoker.ids.NewID("apr"),
				TraceID:       request.TraceID,
				CapabilityID:  request.CapabilityID,
				Risk:          decision.Risk,
				Status:        domain.ApprovalPending,
				Reason:        decision.Reason,
				DiffSummary:   "Capability invocation requires approval.",
			}
			approvalEvent, eventErr := invoker.event(request, domain.EventApprovalRequested, approval)
			if eventErr != nil {
				return domain.CapabilityInvocationResult{}, eventErr
			}
			if err := invoker.events.Append(ctx, []domain.DomainEvent{requestedEvent, policyEvent, approvalEvent}); err != nil {
				return domain.CapabilityInvocationResult{}, err
			}
			return domain.CapabilityInvocationResult{
				CapabilityID: request.CapabilityID,
				TraceID:      request.TraceID,
				OK:           false,
				ErrorCode:    "APPROVAL_REQUIRED",
				ErrorMessage: domain.ErrApprovalRequired.Error(),
				Approval:     &approval,
			}, domain.ErrApprovalRequired
		}
		if err := invoker.ensureApproved(ctx, request); err != nil {
			result := failedResult(request, "APPROVAL_REJECTED", err.Error())
			failedEvent, eventErr := invoker.finishedEvent(request, domain.EventCapabilityInvocationFailed, result)
			if eventErr != nil {
				return result, eventErr
			}
			_ = invoker.events.Append(ctx, []domain.DomainEvent{requestedEvent, policyEvent, failedEvent})
			return result, err
		}
	}

	handler, ok := invoker.handlers[request.CapabilityID]
	if !ok {
		result := failedResult(request, "CAPABILITY_HANDLER_MISSING", "capability handler is not registered")
		failedEvent, eventErr := invoker.finishedEvent(request, domain.EventCapabilityInvocationFailed, result)
		if eventErr != nil {
			return result, eventErr
		}
		_ = invoker.events.Append(ctx, []domain.DomainEvent{requestedEvent, policyEvent, failedEvent})
		return result, errors.New("capability handler is not registered")
	}

	result, err := handler(request)
	result.CapabilityID = request.CapabilityID
	result.TraceID = request.TraceID
	eventType := domain.EventCapabilityInvocationSucceeded
	if err != nil || !result.OK {
		if result.ErrorCode == "" {
			result.ErrorCode = "CAPABILITY_FAILED"
		}
		if result.ErrorMessage == "" && err != nil {
			result.ErrorMessage = err.Error()
		}
		eventType = domain.EventCapabilityInvocationFailed
	}
	finishedEvent, eventErr := invoker.finishedEvent(request, eventType, result)
	if eventErr != nil {
		return result, eventErr
	}
	if err := invoker.events.Append(ctx, []domain.DomainEvent{requestedEvent, policyEvent, finishedEvent}); err != nil {
		return result, err
	}
	return result, err
}

func (invoker *Invoker) ensureApproved(
	ctx context.Context,
	request domain.CapabilityInvocationRequest,
) error {
	approval, err := invoker.approvals.GetApproval(ctx, request.MissionID, request.ApprovalID)
	if err != nil {
		return err
	}
	if approval.Status != domain.ApprovalApproved {
		return fmt.Errorf("%w: %s", domain.ErrApprovalRejected, approval.Status)
	}
	return nil
}

func (invoker *Invoker) appendFinished(
	ctx context.Context,
	request domain.CapabilityInvocationRequest,
	eventType string,
	result domain.CapabilityInvocationResult,
) error {
	event, err := invoker.finishedEvent(request, eventType, result)
	if err != nil {
		return err
	}
	return invoker.events.Append(ctx, []domain.DomainEvent{event})
}

func (invoker *Invoker) finishedEvent(
	request domain.CapabilityInvocationRequest,
	eventType string,
	result domain.CapabilityInvocationResult,
) (domain.DomainEvent, error) {
	return invoker.event(
		request,
		eventType,
		domain.CapabilityInvocationFinishedPayload{
			CapabilityID: result.CapabilityID,
			OK:           result.OK,
			Output:       result.Output,
			ErrorCode:    result.ErrorCode,
			ErrorMessage: result.ErrorMessage,
		},
	)
}

func (invoker *Invoker) event(
	request domain.CapabilityInvocationRequest,
	eventType string,
	payload any,
) (domain.DomainEvent, error) {
	encoded, err := domain.EncodePayload(payload)
	if err != nil {
		return domain.DomainEvent{}, err
	}
	event := domain.DomainEvent{
		EventID:       invoker.ids.NewID("evt"),
		SchemaVersion: domain.EventSchemaVersion,
		TraceID:       request.TraceID,
		MissionID:     request.MissionID,
		RunID:         request.RunID,
		Actor:         request.Actor,
		Timestamp:     invoker.clock.Now(),
		Type:          eventType,
		Payload:       encoded,
	}
	if err := event.Validate(); err != nil {
		return domain.DomainEvent{}, err
	}
	return event, nil
}

func failedResult(
	request domain.CapabilityInvocationRequest,
	code string,
	message string,
) domain.CapabilityInvocationResult {
	return domain.CapabilityInvocationResult{
		CapabilityID: request.CapabilityID,
		TraceID:      request.TraceID,
		OK:           false,
		ErrorCode:    code,
		ErrorMessage: message,
	}
}

func errorCode(err error) string {
	switch {
	case errors.Is(err, domain.ErrCapabilityRevoked):
		return "CAPABILITY_REVOKED"
	case errors.Is(err, domain.ErrCapabilityNotEnabled):
		return "CAPABILITY_NOT_ENABLED"
	default:
		return "CAPABILITY_ERROR"
	}
}

func JSONOutput(value any) json.RawMessage {
	data, _ := json.Marshal(value)
	return data
}
