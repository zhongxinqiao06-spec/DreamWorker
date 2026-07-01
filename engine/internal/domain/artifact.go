package domain

import "time"

type ArtifactMeta struct {
	SchemaVersion string    `json:"schema_version"`
	ArtifactID    string    `json:"artifact_id"`
	MissionID     string    `json:"mission_id"`
	RunID         string    `json:"run_id,omitempty"`
	Kind          string    `json:"kind"`
	Title         string    `json:"title"`
	Version       int       `json:"version"`
	URI           string    `json:"uri"`
	ContentType   string    `json:"content_type,omitempty"`
	Path          string    `json:"path"`
	TraceID       string    `json:"trace_id"`
	CreatedAt     time.Time `json:"created_at"`
}

type ArtifactWrite struct {
	ArtifactID  string
	MissionID   string
	RunID       string
	Kind        string
	Title       string
	Version     int
	ContentType string
	TraceID     string
	FileName    string
	Content     []byte
}

type Artifact struct {
	Meta    ArtifactMeta
	Content []byte
}
