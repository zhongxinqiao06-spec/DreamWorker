package mineru

import "testing"

func TestResolveCommandReportsMissingCLIWithoutError(t *testing.T) {
	t.Setenv("PATH", "")
	t.Setenv("MINERU_COMMAND", "")

	command, ok, err := resolveCommand()
	if err != nil {
		t.Fatalf("resolve command: %v", err)
	}
	if ok {
		t.Fatalf("expected missing cli, got %q", command)
	}
}

func TestNewMinerUAPIClientUsesFlashWithoutToken(t *testing.T) {
	t.Setenv("MINERU_TOKEN", "")

	client, mode, err := newMinerUAPIClient()
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	if client == nil {
		t.Fatal("expected built-in MinerU Open API client")
	}
	if mode != "flash" {
		t.Fatalf("expected flash mode, got %q", mode)
	}
}

func TestNewMinerUAPIClientUsesPrecisionWithToken(t *testing.T) {
	t.Setenv("MINERU_TOKEN", "test-token")

	client, mode, err := newMinerUAPIClient()
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	if client == nil {
		t.Fatal("expected built-in MinerU Open API client")
	}
	if mode != "precision" {
		t.Fatalf("expected precision mode, got %q", mode)
	}
}
