package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/domain"
	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/ports"
)

var _ ports.ArtifactStore = (*ArtifactStore)(nil)

type ArtifactStore struct {
	db          *sql.DB
	artifactDir string
}

func NewArtifactStore(db *sql.DB, projectDir string) (*ArtifactStore, error) {
	artifactDir, err := filepath.Abs(filepath.Join(projectDir, "artifacts"))
	if err != nil {
		return nil, fmt.Errorf("resolve artifact directory: %w", err)
	}
	return &ArtifactStore{db: db, artifactDir: artifactDir}, nil
}

func (store *ArtifactStore) Put(
	ctx context.Context,
	write domain.ArtifactWrite,
) (domain.ArtifactMeta, error) {
	if err := validateArtifactWrite(write); err != nil {
		return domain.ArtifactMeta{}, err
	}

	fileName, err := safePathPart(write.FileName)
	if err != nil {
		return domain.ArtifactMeta{}, fmt.Errorf("invalid artifact file name: %w", err)
	}
	artifactID, err := safePathPart(write.ArtifactID)
	if err != nil {
		return domain.ArtifactMeta{}, fmt.Errorf("invalid artifact id: %w", err)
	}

	finalPath, err := store.artifactPath(artifactID, write.Version, fileName)
	if err != nil {
		return domain.ArtifactMeta{}, err
	}
	if err := os.MkdirAll(filepath.Dir(finalPath), 0o700); err != nil {
		return domain.ArtifactMeta{}, fmt.Errorf("create artifact directory: %w", err)
	}

	tempPath := finalPath + ".tmp"
	if err := os.WriteFile(tempPath, write.Content, 0o600); err != nil {
		return domain.ArtifactMeta{}, fmt.Errorf("write artifact temp file: %w", err)
	}
	cleanupTemp := true
	defer func() {
		if cleanupTemp {
			_ = os.Remove(tempPath)
		}
	}()

	meta := domain.ArtifactMeta{
		SchemaVersion: domain.ContractSchemaVersion,
		ArtifactID:    write.ArtifactID,
		MissionID:     write.MissionID,
		RunID:         write.RunID,
		Kind:          write.Kind,
		Title:         write.Title,
		Version:       write.Version,
		URI:           fmt.Sprintf("artifact://%s/%s/v%d/%s", write.MissionID, write.ArtifactID, write.Version, fileName),
		ContentType:   write.ContentType,
		Path:          finalPath,
		TraceID:       write.TraceID,
		CreatedAt:     time.Now().UTC(),
	}

	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.ArtifactMeta{}, fmt.Errorf("begin artifact write: %w", err)
	}
	defer rollbackUnlessCommitted(tx)

	if _, err := tx.ExecContext(ctx, `
INSERT INTO artifacts (
  artifact_id,
  version,
  schema_version,
  mission_id,
  run_id,
  kind,
  title,
  uri,
  content_type,
  path,
  trace_id,
  created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		meta.ArtifactID,
		meta.Version,
		meta.SchemaVersion,
		meta.MissionID,
		nullableString(meta.RunID),
		meta.Kind,
		meta.Title,
		meta.URI,
		nullableString(meta.ContentType),
		meta.Path,
		meta.TraceID,
		meta.CreatedAt.Format(time.RFC3339Nano),
	); err != nil {
		return domain.ArtifactMeta{}, fmt.Errorf("insert artifact metadata: %w", err)
	}

	if err := os.Rename(tempPath, finalPath); err != nil {
		return domain.ArtifactMeta{}, fmt.Errorf("commit artifact file: %w", err)
	}
	cleanupTemp = false

	if err := tx.Commit(); err != nil {
		_ = os.Remove(finalPath)
		return domain.ArtifactMeta{}, fmt.Errorf("commit artifact metadata: %w", err)
	}
	return meta, nil
}

func (store *ArtifactStore) GetMeta(
	ctx context.Context,
	artifactID string,
	version int,
) (domain.ArtifactMeta, error) {
	row := store.db.QueryRowContext(ctx, `
SELECT artifact_id, version, schema_version, mission_id, run_id, kind, title, uri, content_type, path, trace_id, created_at
FROM artifacts
WHERE artifact_id = ? AND version = ?`, artifactID, version)
	return scanArtifactMeta(row)
}

func (store *ArtifactStore) Read(
	ctx context.Context,
	artifactID string,
	version int,
) (domain.Artifact, error) {
	meta, err := store.GetMeta(ctx, artifactID, version)
	if err != nil {
		return domain.Artifact{}, err
	}
	content, err := os.ReadFile(meta.Path)
	if err != nil {
		return domain.Artifact{}, fmt.Errorf("read artifact file: %w", err)
	}
	return domain.Artifact{Meta: meta, Content: content}, nil
}

func (store *ArtifactStore) artifactPath(
	artifactID string,
	version int,
	fileName string,
) (string, error) {
	target := filepath.Join(store.artifactDir, artifactID, fmt.Sprintf("v%d", version), fileName)
	artifactRoot, err := filepath.Abs(store.artifactDir)
	if err != nil {
		return "", fmt.Errorf("resolve artifact root: %w", err)
	}
	targetPath, err := filepath.Abs(target)
	if err != nil {
		return "", fmt.Errorf("resolve artifact path: %w", err)
	}
	if !isWithinDirectory(artifactRoot, targetPath) {
		return "", errors.New("artifact path escapes project directory")
	}
	return targetPath, nil
}

func scanArtifactMeta(scanner interface {
	Scan(dest ...any) error
}) (domain.ArtifactMeta, error) {
	var meta domain.ArtifactMeta
	var runID sql.NullString
	var contentType sql.NullString
	var createdAt string

	if err := scanner.Scan(
		&meta.ArtifactID,
		&meta.Version,
		&meta.SchemaVersion,
		&meta.MissionID,
		&runID,
		&meta.Kind,
		&meta.Title,
		&meta.URI,
		&contentType,
		&meta.Path,
		&meta.TraceID,
		&createdAt,
	); err != nil {
		return domain.ArtifactMeta{}, fmt.Errorf("scan artifact metadata: %w", err)
	}

	meta.RunID = runID.String
	meta.ContentType = contentType.String
	parsedCreatedAt, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return domain.ArtifactMeta{}, fmt.Errorf("parse artifact created_at: %w", err)
	}
	meta.CreatedAt = parsedCreatedAt
	return meta, nil
}

func validateArtifactWrite(write domain.ArtifactWrite) error {
	switch {
	case !strings.HasPrefix(write.ArtifactID, "art_"):
		return errors.New("artifact_id must start with art_")
	case !strings.HasPrefix(write.MissionID, "msn_"):
		return errors.New("mission_id must start with msn_")
	case write.RunID != "" && !strings.HasPrefix(write.RunID, "run_"):
		return errors.New("run_id must start with run_")
	case strings.TrimSpace(write.Kind) == "":
		return errors.New("kind is required")
	case strings.TrimSpace(write.Title) == "":
		return errors.New("title is required")
	case write.Version < 1:
		return errors.New("version must be at least 1")
	case !strings.HasPrefix(write.TraceID, "tr_"):
		return errors.New("trace_id must start with tr_")
	case len(write.Content) == 0:
		return errors.New("content is required")
	default:
		return nil
	}
}

func safePathPart(value string) (string, error) {
	if value == "" {
		return "", errors.New("path part is required")
	}
	if value != filepath.Base(value) || strings.Contains(value, "..") {
		return "", errors.New("path traversal is not allowed")
	}
	return value, nil
}

func isWithinDirectory(root string, target string) bool {
	relative, err := filepath.Rel(root, target)
	if err != nil {
		return false
	}
	return relative == "." || (!strings.HasPrefix(relative, "..") && !filepath.IsAbs(relative))
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
