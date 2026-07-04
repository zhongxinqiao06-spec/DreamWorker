package runtimeapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/workspace"
)

func TestProjectDirectoryRoutesPersistAndInitializeLocalFiles(t *testing.T) {
	store := workspace.NewStore(
		workspace.WithClock(func() string { return "2026-07-04T00:00:00Z" }),
		workspace.WithTraceID(func() string { return "tr_routes" }),
	)
	mux := NewMuxWithStore("secret-token", store)
	root := filepath.Join(t.TempDir(), "DreamWorkerProject")

	project := requestJSON[workspace.Project](t, mux, http.MethodPost, "/projects/create", map[string]any{
		"title":       "Route Project",
		"description": "created through route",
	})

	updated := requestJSON[workspace.Project](t, mux, http.MethodPost, "/projects/update", map[string]any{
		"projectId":     project.ProjectID,
		"title":         "Route Project Saved",
		"localRootPath": root,
	})
	if updated.LocalRootPath == nil || *updated.LocalRootPath != root {
		t.Fatalf("update route localRootPath = %#v, want %q", updated.LocalRootPath, root)
	}
	if updated.LocalDirectoryStatus != "invalid" {
		t.Fatalf("update route status = %q, want invalid", updated.LocalDirectoryStatus)
	}
	if _, err := os.Stat(filepath.Join(root, ".dreamworker")); !os.IsNotExist(err) {
		t.Fatalf("save before initialization should not create .dreamworker, err=%v", err)
	}

	check := requestJSON[workspace.ProjectDirectoryCheck](t, mux, http.MethodPost, "/projects/local-directory/initialize", map[string]any{
		"projectId": project.ProjectID,
	})
	if check.Status != "valid" || !check.Exists || !check.DreamworkerInitialized {
		t.Fatalf("initialize route check = %#v, want valid initialized directory", check)
	}
	if _, err := os.Stat(filepath.Join(root, ".dreamworker", "project.json")); err != nil {
		t.Fatalf("project.json missing after initialize route: %v", err)
	}

	updated = requestJSON[workspace.Project](t, mux, http.MethodPost, "/projects/update", map[string]any{
		"projectId":   project.ProjectID,
		"description": "saved after initialization",
	})
	payload, err := os.ReadFile(filepath.Join(root, ".dreamworker", "project.json"))
	if err != nil {
		t.Fatalf("read project.json: %v", err)
	}
	var localProject map[string]any
	if err := json.Unmarshal(payload, &localProject); err != nil {
		t.Fatalf("decode project.json: %v", err)
	}
	if localProject["description"] != updated.Description {
		t.Fatalf("local project description = %#v, want %q", localProject["description"], updated.Description)
	}
}

func requestJSON[T any](
	t *testing.T,
	handler http.Handler,
	method string,
	path string,
	body any,
) T {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	request := httptest.NewRequest(method, path, bytes.NewReader(payload))
	request.Header.Set("Authorization", "Bearer secret-token")
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("%s %s status = %d body=%s", method, path, recorder.Code, recorder.Body.String())
	}
	var response T
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return response
}
