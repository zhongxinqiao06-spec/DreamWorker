package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/domain"
	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/ports"
)

var _ ports.CapabilityRegistry = (*CapabilityRegistry)(nil)

type CapabilityRegistry struct {
	db *sql.DB
}

func NewCapabilityRegistry(db *sql.DB) *CapabilityRegistry {
	return &CapabilityRegistry{db: db}
}

func (registry *CapabilityRegistry) Discover(
	ctx context.Context,
	manifest domain.CapabilityManifest,
	trustLevel domain.TrustLevel,
) (domain.CapabilityRecord, error) {
	if err := manifest.Validate(); err != nil {
		return domain.CapabilityRecord{}, err
	}
	now := time.Now().UTC()
	record := domain.CapabilityRecord{
		Manifest:       manifest,
		Lifecycle:      domain.CapabilityDiscovered,
		TrustLevel:     trustLevel,
		RiskLevel:      manifest.Risk.Level,
		RiskActions:    riskActionsFromReasons(manifest.Risk.Reasons),
		CreatedAt:      now,
		UpdatedAt:      now,
		LastTransition: "discover",
	}
	if err := registry.insert(ctx, record); err != nil {
		return domain.CapabilityRecord{}, err
	}
	return record, nil
}

func (registry *CapabilityRegistry) Transition(
	ctx context.Context,
	capabilityID string,
	nextState domain.LifecycleState,
) (domain.CapabilityRecord, error) {
	record, err := registry.Get(ctx, capabilityID)
	if err != nil {
		return domain.CapabilityRecord{}, err
	}
	if !record.Lifecycle.CanTransitionTo(nextState) {
		return domain.CapabilityRecord{}, fmt.Errorf(
			"%w: %s -> %s",
			domain.ErrInvalidCapabilityTransition,
			record.Lifecycle,
			nextState,
		)
	}

	record.Lifecycle = nextState
	record.UpdatedAt = time.Now().UTC()
	record.LastTransition = string(nextState)
	if err := registry.update(ctx, record); err != nil {
		return domain.CapabilityRecord{}, err
	}
	return record, nil
}

func (registry *CapabilityRegistry) Get(
	ctx context.Context,
	capabilityID string,
) (domain.CapabilityRecord, error) {
	row := registry.db.QueryRowContext(ctx, `
SELECT manifest, lifecycle, trust_level, risk_level, risk_actions, created_at, updated_at, last_transition
FROM capabilities
WHERE capability_id = ?`, capabilityID)
	return scanCapabilityRecord(row)
}

func (registry *CapabilityRegistry) ListEnabled(ctx context.Context) ([]domain.CapabilityRecord, error) {
	rows, err := registry.db.QueryContext(ctx, `
SELECT manifest, lifecycle, trust_level, risk_level, risk_actions, created_at, updated_at, last_transition
FROM capabilities
WHERE lifecycle = ?
ORDER BY capability_id`, domain.CapabilityEnabled)
	if err != nil {
		return nil, fmt.Errorf("list enabled capabilities: %w", err)
	}
	defer rows.Close()

	var records []domain.CapabilityRecord
	for rows.Next() {
		record, err := scanCapabilityRecord(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan enabled capabilities: %w", err)
	}
	return records, nil
}

func (registry *CapabilityRegistry) insert(ctx context.Context, record domain.CapabilityRecord) error {
	manifestJSON, riskActionsJSON, err := encodeCapabilityRecord(record)
	if err != nil {
		return err
	}
	_, err = registry.db.ExecContext(ctx, `
INSERT INTO capabilities (
  capability_id,
  manifest,
  lifecycle,
  trust_level,
  risk_level,
  risk_actions,
  created_at,
  updated_at,
  last_transition
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		record.Manifest.Metadata.ID,
		string(manifestJSON),
		record.Lifecycle,
		record.TrustLevel,
		record.RiskLevel,
		string(riskActionsJSON),
		record.CreatedAt.Format(time.RFC3339Nano),
		record.UpdatedAt.Format(time.RFC3339Nano),
		record.LastTransition,
	)
	if err != nil {
		return fmt.Errorf("insert capability %s: %w", record.Manifest.Metadata.ID, err)
	}
	return nil
}

func (registry *CapabilityRegistry) update(ctx context.Context, record domain.CapabilityRecord) error {
	manifestJSON, riskActionsJSON, err := encodeCapabilityRecord(record)
	if err != nil {
		return err
	}
	_, err = registry.db.ExecContext(ctx, `
UPDATE capabilities
SET manifest = ?, lifecycle = ?, trust_level = ?, risk_level = ?, risk_actions = ?, updated_at = ?, last_transition = ?
WHERE capability_id = ?`,
		string(manifestJSON),
		record.Lifecycle,
		record.TrustLevel,
		record.RiskLevel,
		string(riskActionsJSON),
		record.UpdatedAt.Format(time.RFC3339Nano),
		record.LastTransition,
		record.Manifest.Metadata.ID,
	)
	if err != nil {
		return fmt.Errorf("update capability %s: %w", record.Manifest.Metadata.ID, err)
	}
	return nil
}

func scanCapabilityRecord(scanner interface {
	Scan(dest ...any) error
}) (domain.CapabilityRecord, error) {
	var manifestJSON string
	var lifecycle string
	var trustLevel string
	var riskLevel string
	var riskActionsJSON string
	var createdAt string
	var updatedAt string
	var lastTransition string

	if err := scanner.Scan(
		&manifestJSON,
		&lifecycle,
		&trustLevel,
		&riskLevel,
		&riskActionsJSON,
		&createdAt,
		&updatedAt,
		&lastTransition,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.CapabilityRecord{}, domain.ErrCapabilityNotFound
		}
		return domain.CapabilityRecord{}, fmt.Errorf("scan capability: %w", err)
	}

	var manifest domain.CapabilityManifest
	if err := json.Unmarshal([]byte(manifestJSON), &manifest); err != nil {
		return domain.CapabilityRecord{}, fmt.Errorf("decode capability manifest: %w", err)
	}
	var riskActions []domain.RiskAction
	if err := json.Unmarshal([]byte(riskActionsJSON), &riskActions); err != nil {
		return domain.CapabilityRecord{}, fmt.Errorf("decode capability risk actions: %w", err)
	}
	created, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return domain.CapabilityRecord{}, fmt.Errorf("parse capability created_at: %w", err)
	}
	updated, err := time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return domain.CapabilityRecord{}, fmt.Errorf("parse capability updated_at: %w", err)
	}

	return domain.CapabilityRecord{
		Manifest:       manifest,
		Lifecycle:      domain.LifecycleState(lifecycle),
		TrustLevel:     domain.TrustLevel(trustLevel),
		RiskLevel:      domain.RiskLevel(riskLevel),
		RiskActions:    riskActions,
		CreatedAt:      created,
		UpdatedAt:      updated,
		LastTransition: lastTransition,
	}, nil
}

func encodeCapabilityRecord(record domain.CapabilityRecord) ([]byte, []byte, error) {
	manifestJSON, err := json.Marshal(record.Manifest)
	if err != nil {
		return nil, nil, fmt.Errorf("encode capability manifest: %w", err)
	}
	riskActionsJSON, err := json.Marshal(record.RiskActions)
	if err != nil {
		return nil, nil, fmt.Errorf("encode capability risk actions: %w", err)
	}
	return manifestJSON, riskActionsJSON, nil
}

func riskActionsFromReasons(reasons []string) []domain.RiskAction {
	actions := make([]domain.RiskAction, 0, len(reasons))
	for _, reason := range reasons {
		switch domain.RiskAction(reason) {
		case domain.RiskExternalWrite,
			domain.RiskFileWriteOutsideProject,
			domain.RiskSecretAccess,
			domain.RiskNetworkUntrusted,
			domain.RiskPaidCall,
			domain.RiskCodeExecution,
			domain.RiskBrowserAction,
			domain.RiskPublishContent,
			domain.RiskSendEmail,
			domain.RiskInstallSkill,
			domain.RiskConnectRemoteMCP:
			actions = append(actions, domain.RiskAction(reason))
		}
	}
	return actions
}
