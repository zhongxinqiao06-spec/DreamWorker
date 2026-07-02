package workspace

import (
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
