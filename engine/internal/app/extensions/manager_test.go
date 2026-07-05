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

func TestSetSecretRejectsNonASCIIToken(t *testing.T) {
	manager := NewNodeExtensionManager(WithBaseDir(t.TempDir()), WithPersistence(true))

	if err := manager.SetSecret(NineRouterExtensionID, "sk-4b5鈥⑩€⑩€"); err == nil || err.Code != "EXTENSION_SECRET_INVALID" {
		t.Fatalf("expected invalid secret error, got %#v", err)
	}
	if manager.Secret(NineRouterExtensionID) != "" {
		t.Fatalf("expected invalid secret not to be stored")
	}
}

func TestSettingsNormalizeKiroModelAlias(t *testing.T) {
	manager := NewNodeExtensionManager(WithBaseDir(t.TempDir()))
	model := "kiro/claude-sonnet-4.5"
	settings, err := manager.UpdateSettings(UpdateSettingsInput{NineRouterDefaultModel: &model})
	if err != nil {
		t.Fatal(err)
	}
	if settings.NineRouterDefaultModel != "kr/claude-sonnet-4.5" {
		t.Fatalf("expected kr alias, got %q", settings.NineRouterDefaultModel)
	}
}

func TestPersistentStateSkipsCorruptedSecret(t *testing.T) {
	baseDir := t.TempDir()
	data := []byte(`{"settings":{"nineRouterDefaultModel":"kiro/claude-sonnet-4.5"},"secrets":{"extension_9router":"sk-4b5鈥⑩€⑩€"}}`)
	if err := os.WriteFile(filepath.Join(baseDir, "extensions.config.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}

	manager := NewNodeExtensionManager(WithBaseDir(baseDir), WithPersistence(true))
	if manager.Secret(NineRouterExtensionID) != "" {
		t.Fatalf("expected corrupted secret to be ignored")
	}
	if manager.GetSettings().NineRouterDefaultModel != "kr/claude-sonnet-4.5" {
		t.Fatalf("expected persisted kiro alias to normalize, got %q", manager.GetSettings().NineRouterDefaultModel)
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

func TestLocalNineRouterHTTPSSettingsNormalizeToHTTP(t *testing.T) {
	manager := NewNodeExtensionManager(WithBaseDir(t.TempDir()))
	baseURL := "https://127.0.0.1:20128/v1"
	dashboardURL := "https://localhost:20128/dashboard"
	settings, err := manager.UpdateSettings(UpdateSettingsInput{
		NineRouterBaseURL:      &baseURL,
		NineRouterDashboardURL: &dashboardURL,
	})
	if err != nil {
		t.Fatal(err)
	}
	if settings.NineRouterBaseURL != "http://127.0.0.1:20128/v1" {
		t.Fatalf("expected local base url to normalize to http, got %q", settings.NineRouterBaseURL)
	}
	if settings.NineRouterDashboardURL != "http://localhost:20128/dashboard" {
		t.Fatalf("expected local dashboard url to normalize to http, got %q", settings.NineRouterDashboardURL)
	}
}

func TestStartExtensionExternalModeExplainsManualStartup(t *testing.T) {
	manager := NewNodeExtensionManager(WithBaseDir(t.TempDir()))
	baseURL := "http://127.0.0.1:1/v1"
	dashboardURL := "http://127.0.0.1:1"
	if _, err := manager.UpdateSettings(UpdateSettingsInput{
		NineRouterBaseURL:      &baseURL,
		NineRouterDashboardURL: &dashboardURL,
	}); err != nil {
		t.Fatal(err)
	}

	result, appErr := manager.StartExtension(context.Background(), NineRouterExtensionID)
	if appErr != nil {
		t.Fatal(appErr)
	}
	if result.OK {
		t.Fatalf("expected external startup to report unreachable service")
	}
	if result.Status.ProcessState != "stopped" || result.Status.HealthStatus != "error" || result.Status.LastErrorCode != "EXTENSION_EXTERNAL_SERVICE_UNREACHABLE" {
		t.Fatalf("expected external service error status, got %#v", result.Status)
	}
	if !strings.Contains(result.Message, "外部服务模式") || !strings.Contains(result.Message, "不会启动受管 9Router") {
		t.Fatalf("expected external mode guidance, got %q", result.Message)
	}
	lines, err := manager.TailLogs(TailLogsInput{ExtensionID: NineRouterExtensionID, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) == 0 || !strings.Contains(lines[0].Line, "外部服务模式") {
		t.Fatalf("expected external mode log line, got %#v", lines)
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
		case "/v1/models/image":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"object":"list","data":[{"id":"cx/gpt-5.5-image"}]}`))
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
	if len(result.Models) != 2 || result.Models[0] != "cx/gpt-5.5-image" || result.Models[1] != "kr/claude-sonnet-4.5" {
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
