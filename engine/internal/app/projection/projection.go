package projection

import (
	"context"
	"time"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/domain"
	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/ports"
)

type MissionProjection struct {
	MissionID     string
	EventCount    int
	LastEventID   string
	LastEventType string
	UpdatedAt     time.Time
}

type RunProjection struct {
	RunID         string
	EventCount    int
	LastEventID   string
	LastEventType string
	UpdatedAt     time.Time
}

func ReplayMission(ctx context.Context, store ports.EventStore, missionID string) (MissionProjection, error) {
	events, err := store.LoadMission(ctx, missionID)
	if err != nil {
		return MissionProjection{}, err
	}

	projection := MissionProjection{MissionID: missionID}
	for _, event := range events {
		applyMissionEvent(&projection, event)
	}
	return projection, nil
}

func ReplayRun(ctx context.Context, store ports.EventStore, runID string) (RunProjection, error) {
	events, err := store.LoadRun(ctx, runID)
	if err != nil {
		return RunProjection{}, err
	}

	projection := RunProjection{RunID: runID}
	for _, event := range events {
		applyRunEvent(&projection, event)
	}
	return projection, nil
}

func applyMissionEvent(projection *MissionProjection, event domain.DomainEvent) {
	projection.EventCount++
	projection.LastEventID = event.EventID
	projection.LastEventType = event.Type
	projection.UpdatedAt = event.Timestamp
}

func applyRunEvent(projection *RunProjection, event domain.DomainEvent) {
	projection.EventCount++
	projection.LastEventID = event.EventID
	projection.LastEventType = event.Type
	projection.UpdatedAt = event.Timestamp
}
