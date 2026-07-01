package sqlite

import (
	"bytes"
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/domain"
)

func TestArtifactStorePutGetAndRead(t *testing.T) {
	ctx := context.Background()
	db := openBootstrappedTestDB(t, ctx)
	store := newTestArtifactStore(t, db)

	meta, err := store.Put(ctx, sampleArtifactWrite("art_dream_brief", 1, "dream_brief.md"))
	if err != nil {
		t.Fatalf("put artifact: %v", err)
	}

	if meta.SchemaVersion != domain.ContractSchemaVersion {
		t.Fatalf("expected schema version %q, got %q", domain.ContractSchemaVersion, meta.SchemaVersion)
	}
	if meta.URI != "artifact://msn_001/art_dream_brief/v1/dream_brief.md" {
		t.Fatalf("unexpected artifact uri %q", meta.URI)
	}

	loadedMeta, err := store.GetMeta(ctx, "art_dream_brief", 1)
	if err != nil {
		t.Fatalf("get artifact metadata: %v", err)
	}
	if loadedMeta.ArtifactID != meta.ArtifactID {
		t.Fatalf("expected artifact id %q, got %q", meta.ArtifactID, loadedMeta.ArtifactID)
	}

	artifact, err := store.Read(ctx, "art_dream_brief", 1)
	if err != nil {
		t.Fatalf("read artifact: %v", err)
	}
	if !bytes.Equal(artifact.Content, []byte("# Dream Brief\n")) {
		t.Fatalf("unexpected artifact content %q", string(artifact.Content))
	}
}

func TestArtifactStoreRejectsPathTraversal(t *testing.T) {
	ctx := context.Background()
	db := openBootstrappedTestDB(t, ctx)
	projectDir := t.TempDir()
	store, err := NewArtifactStore(db, projectDir)
	if err != nil {
		t.Fatalf("new artifact store: %v", err)
	}

	_, err = store.Put(ctx, sampleArtifactWrite("art_dream_brief", 1, "../evil.md"))
	if err == nil {
		t.Fatal("expected path traversal error")
	}

	if _, statErr := os.Stat(filepath.Join(projectDir, "evil.md")); !os.IsNotExist(statErr) {
		t.Fatalf("expected no file outside artifacts dir, stat err=%v", statErr)
	}
}

func TestArtifactStoreRejectsVersionConflict(t *testing.T) {
	ctx := context.Background()
	db := openBootstrappedTestDB(t, ctx)
	store := newTestArtifactStore(t, db)

	write := sampleArtifactWrite("art_dream_brief", 1, "dream_brief.md")
	if _, err := store.Put(ctx, write); err != nil {
		t.Fatalf("put first artifact: %v", err)
	}
	if _, err := store.Put(ctx, write); err == nil {
		t.Fatal("expected artifact version conflict")
	}
}

func newTestArtifactStore(t *testing.T, db *sql.DB) *ArtifactStore {
	t.Helper()
	store, err := NewArtifactStore(db, t.TempDir())
	if err != nil {
		t.Fatalf("new artifact store: %v", err)
	}
	return store
}

func sampleArtifactWrite(artifactID string, version int, fileName string) domain.ArtifactWrite {
	return domain.ArtifactWrite{
		ArtifactID:  artifactID,
		MissionID:   "msn_001",
		RunID:       "run_001",
		Kind:        "dream_brief",
		Title:       "Dream Brief",
		Version:     version,
		ContentType: "text/markdown",
		TraceID:     "tr_artifact",
		FileName:    fileName,
		Content:     []byte("# Dream Brief\n"),
	}
}
