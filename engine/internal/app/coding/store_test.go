package coding

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/resources"
)

func newCodingStoreForTest(t *testing.T, root string) *Store {
	t.Helper()
	state := resources.NewStore(
		resources.WithClock(func() string { return "2026-07-05T00:00:00Z" }),
		resources.WithTraceID(func() string { return "tr_coding" }),
	)
	state.Mu.Lock()
	state.Projects["project_001"] = resources.Project{
		ProjectID:             "project_001",
		Title:                 "Coding Project",
		Description:           "Project with local root.",
		Status:                "active",
		LocalRootPath:         &root,
		LocalDirectoryStatus:  "valid",
		DefaultModelProfileID: "profile_fast",
		CreatedAt:             "2026-07-05T00:00:00Z",
		UpdatedAt:             "2026-07-05T00:00:00Z",
	}
	state.Providers["provider_deepseek"] = resources.ModelProviderRecord{
		SafeModelProvider: resources.SafeModelProvider{
			ProviderID:      "provider_deepseek",
			ProviderType:    resources.ProviderDeepSeek,
			DisplayName:     "DeepSeek",
			BaseURL:         "https://api.deepseek.com",
			DefaultModel:    "deepseek-chat",
			AvailableModels: []string{"deepseek-chat"},
			Enabled:         true,
		},
		APIKey: "sk-test",
	}
	state.Mu.Unlock()
	return NewStore(state)
}

func TestListEnginesReportsRuntimeAndDescriptors(t *testing.T) {
	runtimeDir := t.TempDir()
	adapterPath := filepath.Join(runtimeDir, "adapter", "dist")
	if err := os.MkdirAll(adapterPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(adapterPath, "index.js"), []byte("console.log('ok')"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DREAMWORKER_CODING_AGENT_RUNTIME_DIR", runtimeDir)
	t.Setenv("DREAMWORKER_CODING_AGENT_NODE_BIN", "node")

	store := newCodingStoreForTest(t, t.TempDir())
	status := store.ListEngines()

	if !status.Available {
		t.Fatalf("runtime available = false, message %q", status.Message)
	}
	if status.AdapterPath == "" {
		t.Fatal("expected adapter path")
	}
	if len(status.Engines) != 3 {
		t.Fatalf("engines = %d, want 3", len(status.Engines))
	}
}

func TestCreateSessionUsesProjectRootProviderAndModelFallback(t *testing.T) {
	root := t.TempDir()
	store := newCodingStoreForTest(t, root)

	session, appErr := store.CreateSession(CreateSessionInput{
		ProjectID:  "project_001",
		EngineID:   EngineCodex,
		ProviderID: "provider_deepseek",
	})
	if appErr != nil {
		t.Fatalf("CreateSession error = %#v", appErr)
	}
	if session.LocalRootPath != root {
		t.Fatalf("local root = %q, want %q", session.LocalRootPath, root)
	}
	if session.Model != "deepseek-chat" {
		t.Fatalf("model = %q, want provider default", session.Model)
	}
	if session.Status != "ready" {
		t.Fatalf("status = %q, want ready", session.Status)
	}
}

func TestListAndReadFilesStayInsideProjectRoot(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	store := newCodingStoreForTest(t, root)

	files, appErr := store.ListFiles(ListFilesInput{ProjectID: "project_001", Query: "main"})
	if appErr != nil {
		t.Fatalf("ListFiles error = %#v", appErr)
	}
	if len(files) != 1 || files[0].Path != "src/main.go" {
		t.Fatalf("files = %#v, want src/main.go", files)
	}
	result, appErr := store.ReadFile(ReadFileInput{ProjectID: "project_001", Path: "src/main.go"})
	if appErr != nil {
		t.Fatalf("ReadFile error = %#v", appErr)
	}
	if result.Content != "package main\n" || result.MimeType != "text/plain" {
		t.Fatalf("read result = %#v", result)
	}
	if _, appErr := store.ReadFile(ReadFileInput{ProjectID: "project_001", Path: "../outside.txt"}); appErr == nil || appErr.Code != "PATH_OUTSIDE_PROJECT" {
		t.Fatalf("escape error = %#v, want PATH_OUTSIDE_PROJECT", appErr)
	}
}

func TestCreateSessionRequiresConfiguredLocalRoot(t *testing.T) {
	state := resources.NewStore()
	state.Mu.Lock()
	state.Projects["project_001"] = resources.Project{ProjectID: "project_001", Title: "No root", Status: "active"}
	state.Providers["provider_deepseek"] = resources.ModelProviderRecord{
		SafeModelProvider: resources.SafeModelProvider{
			ProviderID:   "provider_deepseek",
			ProviderType: resources.ProviderDeepSeek,
			DisplayName:  "DeepSeek",
			BaseURL:      "https://api.deepseek.com",
			DefaultModel: "deepseek-chat",
			Enabled:      true,
		},
	}
	state.Mu.Unlock()
	store := NewStore(state)

	_, appErr := store.CreateSession(CreateSessionInput{
		ProjectID:  "project_001",
		ProviderID: "provider_deepseek",
	})
	if appErr == nil || appErr.Code != "LOCAL_DIRECTORY_NOT_SET" {
		t.Fatalf("CreateSession error = %#v, want LOCAL_DIRECTORY_NOT_SET", appErr)
	}
}
