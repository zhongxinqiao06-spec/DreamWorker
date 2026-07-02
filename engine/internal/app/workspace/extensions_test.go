package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/extensions"
)

func TestNineRouterProviderBridgeIsSeededLast(t *testing.T) {
	store := NewStore()

	providers := store.ListProviders()
	if len(providers) == 0 {
		t.Fatalf("expected seeded providers")
	}
	last := providers[len(providers)-1]
	if last.ProviderID != extensions.NineRouterProviderID {
		t.Fatalf("expected 9router provider last, got %q", last.ProviderID)
	}
	if last.BaseURL != "http://localhost:20128/v1" {
		t.Fatalf("expected 9router base url, got %q", last.BaseURL)
	}
	if !last.Enabled {
		t.Fatalf("expected 9router bridge to be selectable when integration is enabled")
	}
	if last.HasAPIKey {
		t.Fatalf("expected 9router endpoint key to be optional by default")
	}
}

func TestNineRouterProviderCannotBeDeleted(t *testing.T) {
	store := NewStore()

	if _, err := store.DeleteProvider(extensions.NineRouterProviderID); err == nil {
		t.Fatalf("expected deleting system provider to fail")
	}
}

func TestNineRouterEndpointKeyPersistsAcrossStoreRestart(t *testing.T) {
	configDir := t.TempDir()
	store := NewStore(
		WithClock(func() string { return "2026-07-01T00:00:00Z" }),
		WithConfigDir(configDir),
	)

	provider, appErr := store.SaveProvider(SaveModelProviderInput{
		ProviderID:      extensions.NineRouterProviderID,
		ProviderType:    ProviderOpenAICompatible,
		DisplayName:     "9Router 免费模型路由",
		BaseURL:         "http://localhost:20128/v1",
		DefaultModel:    "kr/claude-sonnet-4.5",
		AvailableModels: []string{"kr/claude-sonnet-4.5"},
		Enabled:         true,
		Capabilities:    []string{"chat", "tools", "json_schema"},
		APIKey:          "sk-9router-endpoint-secret",
	})
	if appErr != nil {
		t.Fatalf("save 9router provider: %v", appErr)
	}
	if !provider.HasAPIKey || provider.MaskedKey == nil {
		t.Fatalf("expected saved 9router provider to report masked key, got %#v", provider)
	}

	configPath := filepath.Join(configDir, "extensions", "extensions.config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read extension config: %v", err)
	}
	if !strings.Contains(string(data), "sk-9router-endpoint-secret") {
		t.Fatalf("expected 9router endpoint key to persist in extension config")
	}

	reloaded := NewStore(
		WithClock(func() string { return "2026-07-01T00:00:00Z" }),
		WithConfigDir(configDir),
	)
	var found *SafeModelProvider
	for _, item := range reloaded.ListProviders() {
		if item.ProviderID == extensions.NineRouterProviderID {
			copied := item
			found = &copied
			break
		}
	}
	if found == nil || !found.HasAPIKey || found.MaskedKey == nil || *found.MaskedKey != "sk-9...cret" {
		t.Fatalf("expected reloaded 9router provider with masked key, got %#v", found)
	}
}

func TestNineRouterProviderNormalizesKiroModelAlias(t *testing.T) {
	store := NewStore(
		WithClock(func() string { return "2026-07-01T00:00:00Z" }),
		WithConfigDir(t.TempDir()),
	)

	provider, appErr := store.SaveProvider(SaveModelProviderInput{
		ProviderID:      extensions.NineRouterProviderID,
		ProviderType:    ProviderOpenAICompatible,
		DisplayName:     "9Router 免费模型路由",
		BaseURL:         "http://localhost:20128/v1",
		DefaultModel:    "kiro/claude-sonnet-4.5",
		AvailableModels: []string{"kiro/claude-sonnet-4.5", "kr/glm-5"},
		Enabled:         true,
		Capabilities:    []string{"chat", "tools", "json_schema"},
		APIKey:          "sk-9router-endpoint-secret",
	})
	if appErr != nil {
		t.Fatalf("save 9router provider: %v", appErr)
	}
	if provider.DefaultModel != "kr/claude-sonnet-4.5" {
		t.Fatalf("expected normalized default model, got %q", provider.DefaultModel)
	}
	if len(provider.AvailableModels) == 0 || provider.AvailableModels[0] != "kr/claude-sonnet-4.5" {
		t.Fatalf("expected normalized model list, got %#v", provider.AvailableModels)
	}
}

func TestNineRouterChatBindingNormalizesKiroModelAlias(t *testing.T) {
	store := NewStore(
		WithClock(func() string { return "2026-07-01T00:00:00Z" }),
		WithConfigDir(t.TempDir()),
	)

	session, appErr := store.CreateChatSession(CreateChatSessionInput{
		Title:      "Kiro Alias",
		ProviderID: extensions.NineRouterProviderID,
		Model:      "kiro/claude-sonnet-4.5",
	})
	if appErr != nil {
		t.Fatalf("create chat session: %v", appErr)
	}
	if session.Model != "kr/claude-sonnet-4.5" {
		t.Fatalf("expected normalized session model, got %q", session.Model)
	}
}

func TestNineRouterProviderRejectsCorruptedEndpointKey(t *testing.T) {
	store := NewStore(
		WithClock(func() string { return "2026-07-01T00:00:00Z" }),
		WithConfigDir(t.TempDir()),
	)

	_, appErr := store.SaveProvider(SaveModelProviderInput{
		ProviderID:      extensions.NineRouterProviderID,
		ProviderType:    ProviderOpenAICompatible,
		DisplayName:     "9Router 免费模型路由",
		BaseURL:         "http://localhost:20128/v1",
		DefaultModel:    "kr/claude-sonnet-4.5",
		AvailableModels: []string{"kr/claude-sonnet-4.5"},
		Enabled:         true,
		Capabilities:    []string{"chat", "tools", "json_schema"},
		APIKey:          "sk-4b5鈥⑩€⑩€",
	})
	if appErr == nil || appErr.Code != "EXTENSION_SECRET_INVALID" {
		t.Fatalf("expected invalid endpoint key error, got %#v", appErr)
	}
}
