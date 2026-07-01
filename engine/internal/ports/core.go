package ports

import (
	"context"
	"time"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/domain"
)

type EventStore interface {
	Append(ctx context.Context, events []domain.DomainEvent) error
	LoadMission(ctx context.Context, missionID string) ([]domain.DomainEvent, error)
	LoadRun(ctx context.Context, runID string) ([]domain.DomainEvent, error)
}

type ArtifactStore interface {
	Put(ctx context.Context, write domain.ArtifactWrite) (domain.ArtifactMeta, error)
	GetMeta(ctx context.Context, artifactID string, version int) (domain.ArtifactMeta, error)
	Read(ctx context.Context, artifactID string, version int) (domain.Artifact, error)
}

type Clock interface {
	Now() time.Time
}

type IdGenerator interface {
	NewID(prefix string) string
}

type CapabilityRegistry interface {
	Discover(
		ctx context.Context,
		manifest domain.CapabilityManifest,
		trustLevel domain.TrustLevel,
	) (domain.CapabilityRecord, error)
	Transition(
		ctx context.Context,
		capabilityID string,
		nextState domain.LifecycleState,
	) (domain.CapabilityRecord, error)
	Get(ctx context.Context, capabilityID string) (domain.CapabilityRecord, error)
	ListEnabled(ctx context.Context) ([]domain.CapabilityRecord, error)
}

type CapabilityInvoker interface {
	Invoke(
		ctx context.Context,
		request domain.CapabilityInvocationRequest,
	) (domain.CapabilityInvocationResult, error)
}

type PolicyEngine interface {
	Evaluate(ctx context.Context, request domain.PolicyRequest) (domain.PolicyDecision, error)
}

type ApprovalStore interface {
	GetApproval(ctx context.Context, missionID string, approvalID string) (domain.ApprovalRequest, error)
}

type ModelGateway interface {
	Generate(ctx context.Context, request domain.ModelRequest) (domain.ModelResponse, error)
	Stream(ctx context.Context, request domain.ModelRequest) (<-chan domain.ModelStreamEvent, error)
}
type SecretStore interface{}
type SearchIndex interface{}
