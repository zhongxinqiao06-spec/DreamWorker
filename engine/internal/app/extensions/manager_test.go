package extensions

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNineRouterSpecUsesGenericExtensionShape(t *testing.T) {
	manager := NewNodeExtensionManager(WithBaseDir(t.TempDir()))

	specs := manager.ListExtensions()
	if len(specs) != 1 {
		t.Fatalf("expected one system extension, got %d", len(specs))
	}
	spec := specs[0]
	if spec.ExtensionID != NineRouterExtensionID {
		t.Fatalf("expected 9router extension id, got %q", spec.ExtensionID)
	}
	if spec.RuntimeKind != "node" || spec.Kind != "node_managed_provider" {
		t.Fatalf("expected generic node managed provider spec, got kind=%q runtime=%q", spec.Kind, spec.RuntimeKind)
	}
	if spec.ProviderBridge == nil || spec.ProviderBridge.ProviderID != NineRouterProviderID {
		t.Fatalf("expected provider bridge for %s", NineRouterProviderID)
	}
	if spec.Install.PackageName != "9router" || spec.Process.DefaultCommand != "9router" {
		t.Fatalf("expected 9router package and command, got %#v %#v", spec.Install, spec.Process)
	}
}

func TestSetSecretOnlyReturnsMaskedStatus(t *testing.T) {
	manager := NewNodeExtensionManager(WithBaseDir(t.TempDir()))

	if err := manager.SetSecret(NineRouterExtensionID, "sk-test-secret-value"); err != nil {
		t.Fatal(err)
	}
	status, err := manager.GetExtensionStatus(NineRouterExtensionID)
	if err != nil {
		t.Fatal(err)
	}
	if !status.HasAPIKey {
		t.Fatalf("expected status to report configured key")
	}
	if status.MaskedKey == "" || strings.Contains(status.MaskedKey, "secret") {
		t.Fatalf("expected masked key, got %q", status.MaskedKey)
	}
}

func TestPersistentStateRestoresSettingsAndSecret(t *testing.T) {
	baseDir := t.TempDir()
	manager := NewNodeExtensionManager(WithBaseDir(baseDir), WithPersistence(true))
	baseURL := "http://127.0.0.1:20128/v1"
	if _, err := manager.UpdateSettings(UpdateSettingsInput{NineRouterBaseURL: &baseURL}); err != nil {
		t.Fatal(err)
	}
	if err := manager.SetSecret(NineRouterExtensionID, "sk-test-secret-value"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(baseDir, "extensions.config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "sk-test-secret-value") {
		t.Fatalf("expected persisted config to contain endpoint key")
	}

	reloaded := NewNodeExtensionManager(WithBaseDir(baseDir), WithPersistence(true))
	if reloaded.Secret(NineRouterExtensionID) != "sk-test-secret-value" {
		t.Fatalf("expected reloaded secret")
	}
	if reloaded.GetSettings().NineRouterBaseURL != baseURL {
		t.Fatalf("expected reloaded base url")
	}
}

func TestExtensionLogsAreRedacted(t *testing.T) {
	manager := NewNodeExtensionManager(WithBaseDir(t.TempDir()))

	manager.appendLog(NineRouterExtensionID, "stderr", "Authorization: Bearer sk-test-secret-value")
	lines, err := manager.TailLogs(TailLogsInput{ExtensionID: NineRouterExtensionID, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 1 {
		t.Fatalf("expected one log line, got %d", len(lines))
	}
	if strings.Contains(lines[0].Line, "sk-test-secret-value") || strings.Contains(lines[0].Line, "Bearer") {
		t.Fatalf("expected redacted log line, got %q", lines[0].Line)
	}
}

func TestHealthCheckDiscoversModelsWithoutEndpointKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/", "/v1/models":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"object":"list","data":[{"id":"kr/claude-sonnet-4.5"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	manager := NewNodeExtensionManager(WithBaseDir(t.TempDir()))
	baseURL := server.URL + "/v1"
	dashboardURL := server.URL
	_, err := manager.UpdateSettings(UpdateSettingsInput{
		NineRouterBaseURL:      &baseURL,
		NineRouterDashboardURL: &dashboardURL,
	})
	if err != nil {
		t.Fatal(err)
	}

	result, appErr := manager.RefreshModels(context.Background(), NineRouterExtensionID)
	if appErr != nil {
		t.Fatal(appErr)
	}
	if !result.OK || result.Status.HealthStatus != "connected" {
		t.Fatalf("expected connected model discovery, got ok=%v status=%q error=%q", result.OK, result.Status.HealthStatus, result.Status.LastErrorMessage)
	}
	if result.Status.HasAPIKey {
		t.Fatalf("expected endpoint key to remain optional")
	}
	if len(result.Models) != 1 || result.Models[0] != "kr/claude-sonnet-4.5" {
		t.Fatalf("unexpected models: %#v", result.Models)
	}
}

func TestVerifyStreamingReportsNineRouterCredentialError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/", "/v1/models":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"object":"list","data":[{"id":"kr/claude-sonnet-4.5"}]}`))
		case "/v1/chat/completions":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":{"message":"No active credentials for provider: kiro","code":"model_not_found"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	manager := NewNodeExtensionManager(WithBaseDir(t.TempDir()))
	baseURL := server.URL + "/v1"
	dashboardURL := server.URL
	_, err := manager.UpdateSettings(UpdateSettingsInput{
		NineRouterBaseURL:      &baseURL,
		NineRouterDashboardURL: &dashboardURL,
	})
	if err != nil {
		t.Fatal(err)
	}

	result, appErr := manager.VerifyStreaming(context.Background(), NineRouterExtensionID)
	if appErr != nil {
		t.Fatal(appErr)
	}
	if result.OK {
		t.Fatalf("expected streaming verification to fail without upstream credentials")
	}
	if !strings.Contains(result.Message, "No active credentials") {
		t.Fatalf("expected upstream credential detail, got %q", result.Message)
	}
}
