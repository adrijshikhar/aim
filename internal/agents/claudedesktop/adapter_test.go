package claudedesktop_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/agents/claudedesktop"
)

func TestClaudeDesktopAdapter(t *testing.T) {
	adapter := claudedesktop.NewAdapter()

	if adapter.Name() != "claude-desktop" {
		t.Errorf("expected name 'claude-desktop', got '%s'", adapter.Name())
	}
	if adapter.AppBundleName() != "Claude" {
		t.Errorf("expected bundle name 'Claude', got '%s'", adapter.AppBundleName())
	}
	if !adapter.IsDesktop() {
		t.Errorf("expected IsDesktop() to return true")
	}

	tempDir := t.TempDir()
	userDataDir := adapter.UserDataDir(tempDir)
	expectedUserData := filepath.Join(tempDir, ".claude-desktop", "user-data")
	if userDataDir != expectedUserData {
		t.Errorf("expected %s, got %s", expectedUserData, userDataDir)
	}

	ctx := context.Background()
	env, err := adapter.PrepareEnv(ctx, "testprof", tempDir)
	if err != nil {
		t.Fatalf("unexpected error preparing env: %v", err)
	}
	if len(env.Args) == 0 || env.Args[0] != "--user-data-dir="+userDataDir {
		t.Errorf("expected arg --user-data-dir=%s, got %v", userDataDir, env.Args)
	}

	// MCP Config testing
	mcpCfg := agents.MCPServerConfig{
		Command: "npx",
		Args:    []string{"-y", "@test/mcp"},
	}
	if err := adapter.AddProfileMCPServer(ctx, "testprof", tempDir, "test-server", mcpCfg); err != nil {
		t.Fatalf("failed to add profile MCP server: %v", err)
	}

	servers, err := adapter.ListMCPServers(ctx, "testprof", tempDir)
	if err != nil {
		t.Fatalf("failed to list MCP servers: %v", err)
	}
	if len(servers) != 1 || servers[0].Name != "test-server" {
		t.Errorf("expected 1 server named 'test-server', got %v", servers)
	}

	if err := adapter.SetProfileMCPServerState(ctx, "testprof", tempDir, "test-server", false); err != nil {
		t.Fatalf("failed to toggle server state: %v", err)
	}

	servers, _ = adapter.ListMCPServers(ctx, "testprof", tempDir)
	if servers[0].Status != "disabled" {
		t.Errorf("expected server to be disabled, got status %q", servers[0].Status)
	}

	if err := adapter.RemoveProfileMCPServer(ctx, "testprof", tempDir, "test-server"); err != nil {
		t.Fatalf("failed to remove MCP server: %v", err)
	}

	servers, _ = adapter.ListMCPServers(ctx, "testprof", tempDir)
	if len(servers) != 0 {
		t.Errorf("expected 0 servers after removal, got %d", len(servers))
	}
}

func TestClaudeDesktopAdapter_HasCredentials(t *testing.T) {
	adapter := claudedesktop.NewAdapter()
	tempDir := t.TempDir()

	if adapter.HasCredentials(tempDir) {
		t.Errorf("expected false for empty profile dir")
	}

	// Add dummy config
	mcpDir := filepath.Join(tempDir, ".claude-desktop")
	_ = os.MkdirAll(mcpDir, 0700)
	_ = os.WriteFile(filepath.Join(mcpDir, "claude_desktop_config.json"), []byte(`{}`), 0600)

	if !adapter.HasCredentials(tempDir) {
		t.Errorf("expected true after creating mcp config")
	}
}
