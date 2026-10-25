package claude

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/merge"
)

func TestMCPCollections_Claude(t *testing.T) {
	a := &Adapter{}
	cols := a.MCPCollections("/p", "/h")
	if len(cols) != 3 {
		t.Fatalf("collections = %d", len(cols))
	}
	c := cols[0]
	if c.ID() != "claude/mcpServers" || c.Format != merge.JSON || c.Key != "mcpServers" || c.Group != "mcp" || c.Noun != "server" ||
		c.HostPath != filepath.Join("/h", ".claude.json") ||
		c.ProfilePath != filepath.Join("/p", ".claude", ".claude.json") {
		t.Fatalf("collection = %+v", c)
	}
	written := map[string]any{"command": "npx"}
	rewritten := map[string]any{"type": "stdio", "command": "npx", "env": map[string]any{}}
	if merge.Hash(c.Normalise, written) != merge.Hash(c.Normalise, rewritten) {
		t.Fatal("Claude's normalisation noise must not change the hash")
	}
	if !a.IsBackground(nil, []string{"--bg"}) || !a.IsBackground(nil, []string{"agents"}) || a.IsBackground(nil, []string{"--resume", "x"}) {
		t.Fatal("background forms")
	}
}

func TestMCPCollections_ClaudePlugins(t *testing.T) {
	cols := (&Adapter{}).MCPCollections("/p", "/h")
	want := []struct{ name, noun string }{{"enabledPlugins", "plugin"}, {"extraKnownMarketplaces", "marketplace"}}
	for i, w := range want {
		c := cols[i+1]
		if c.ID() != "claude/"+w.name || c.Key != w.name || c.Format != merge.JSON || c.Group != "plugins" || c.Noun != w.noun ||
			c.HostPath != filepath.Join("/h", ".claude", "settings.json") ||
			c.ProfilePath != filepath.Join("/p", ".claude", "settings.json") || c.Normalise != nil {
			t.Fatalf("collection %d = %+v", i+1, c)
		}
	}
	// the JSON codec wraps a bool as {"value": b}; the default normaliser keeps false
	if merge.Hash(nil, map[string]any{"value": true}) == merge.Hash(nil, map[string]any{"value": false}) {
		t.Fatal("enabled and disabled must hash differently")
	}
}

// TestIsSession_Claude: the spellings `claude --help` lists (-v, --version,
// -h, --help) plus -V, which 2.1.283 also accepts. Claude has no help or
// version subcommand, so those words are a prompt and start a session.
func TestIsSession_Claude(t *testing.T) {
	a := &Adapter{}
	cases := []struct {
		args []string
		want bool
	}{
		{nil, true},
		{[]string{"--resume", "x"}, true},
		{[]string{"mcp", "list"}, true}, // shows the merged set, so it merges
		{[]string{"-v"}, false},
		{[]string{"-V"}, false},
		{[]string{"--version"}, false},
		{[]string{"-h"}, false},
		{[]string{"--help"}, false},
		{[]string{"--model", "opus", "--version"}, false},
		{[]string{"version"}, true},
		{[]string{"help"}, true},
		{[]string{"fix", "--", "--help"}, true}, // after "--" belongs to the agent
	}
	for _, c := range cases {
		if got := a.IsSession(nil, c.args); got != c.want {
			t.Errorf("%v: got %v want %v", c.args, got, c.want)
		}
	}
}

func TestClaude_AddAndRemoveHostMCPServer(t *testing.T) {
	home := isolatedHome(t)
	a := &Adapter{}
	ctx := context.Background()

	cfg := agents.MCPServerConfig{
		Command: "npx",
		Args:    []string{"-y", "agentation-mcp", "server"},
	}

	if err := a.AddHostMCPServer(ctx, "agentation", cfg); err != nil {
		t.Fatalf("AddHostMCPServer failed: %v", err)
	}

	claudePath := filepath.Join(home, ".claude.json")
	data, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("failed to read claude config: %v", err)
	}
	if !strings.Contains(string(data), "agentation-mcp") {
		t.Fatalf("expected agentation in claude config, got: %s", string(data))
	}

	// Verify via ListMCPServers
	list, err := a.ListMCPServers(ctx, "default", filepath.Join(home, "profiles", "none"))
	if err != nil {
		t.Fatalf("ListMCPServers failed: %v", err)
	}
	var found bool
	for _, s := range list {
		if s.Name == "agentation" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected agentation in ListMCPServers")
	}

	// Remove server
	if err := a.RemoveHostMCPServer(ctx, "agentation"); err != nil {
		t.Fatalf("RemoveHostMCPServer failed: %v", err)
	}

	data, err = os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("failed to read claude config after remove: %v", err)
	}
	if strings.Contains(string(data), "agentation-mcp") {
		t.Fatalf("expected agentation to be removed, got: %s", string(data))
	}

	// Removing again must return os.ErrNotExist
	if err := a.RemoveHostMCPServer(ctx, "agentation"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected os.ErrNotExist when removing nonexistent server, got: %v", err)
	}
}

func TestClaude_AddProfileMCPServer_Isolation(t *testing.T) {
	home := isolatedHome(t)
	profDir := filepath.Join(home, "profiles", "work")
	_ = os.MkdirAll(profDir, 0o700)

	a := &Adapter{}
	ctx := context.Background()

	cfg := agents.MCPServerConfig{
		Command: "fetch",
		Args:    []string{"https://example.com"},
	}

	if err := a.AddProfileMCPServer(ctx, "work", profDir, "fetch", cfg); err != nil {
		t.Fatalf("AddProfileMCPServer failed: %v", err)
	}

	// Profile config must exist
	pConfig := filepath.Join(profDir, ".claude", ".claude.json")
	if _, err := os.Stat(pConfig); os.IsNotExist(err) {
		t.Fatalf("profile config was not created at %s", pConfig)
	}

	// Host config must NOT exist
	hConfig := filepath.Join(home, ".claude.json")
	if _, err := os.Stat(hConfig); !os.IsNotExist(err) {
		t.Fatalf("host config was mutated during profile operation!")
	}

	// Remove from profile
	if err := a.RemoveProfileMCPServer(ctx, "work", profDir, "fetch"); err != nil {
		t.Fatalf("RemoveProfileMCPServer failed: %v", err)
	}
	data, err := os.ReadFile(pConfig)
	if err != nil {
		t.Fatalf("failed to read profile config: %v", err)
	}
	if strings.Contains(string(data), "fetch") {
		t.Fatalf("expected fetch to be removed from profile config, got: %s", string(data))
	}
}

func TestClaude_DetectInstalled(t *testing.T) {
	home := isolatedHome(t)
	t.Setenv("PATH", t.TempDir()) // Ensure binary is not in PATH for isolated test
	a := &Adapter{}
	if a.DetectInstalled(home) {
		t.Errorf("expected not installed in empty home without binary in path")
	}

	_ = os.WriteFile(filepath.Join(home, ".claude.json"), []byte("{}"), 0o600)
	if !a.DetectInstalled(home) {
		t.Errorf("expected installed when ~/.claude.json exists")
	}
}

func TestClaude_ListMCPServers(t *testing.T) {
	home := isolatedHome(t)
	hostConfig := `{
  "mcpServers": {
    "host-tool": {
      "command": "npx",
      "args": ["-y", "host-tool@latest"]
    },
    "shared-api": {
      "url": "https://mcp.shared.com/v1",
      "headers": {"Authorization": "Bearer tok"}
    }
  }
}`
	_ = os.WriteFile(filepath.Join(home, ".claude.json"), []byte(hostConfig), 0o600)

	prof := filepath.Join(home, "profiles", "dev")
	profClaude := filepath.Join(prof, ".claude")
	_ = os.MkdirAll(profClaude, 0o700)

	profConfig := `{
  "mcpServers": {
    "prof-tool": {
      "command": "/opt/tools/bin/tool",
      "args": []
    },
    "shared-api": {
      "url": "https://mcp.shared.com/v2"
    }
  }
}`
	_ = os.WriteFile(filepath.Join(profClaude, ".claude.json"), []byte(profConfig), 0o600)

	a := &Adapter{}
	servers, err := a.ListMCPServers(context.Background(), "dev", prof)
	if err != nil {
		t.Fatalf("ListMCPServers failed: %v", err)
	}

	if len(servers) != 3 {
		t.Fatalf("expected 3 servers, got %d", len(servers))
	}

	// Sorted: host-tool, prof-tool, shared-api
	if servers[0].Name != "host-tool" || servers[0].Origin != "host" || servers[0].Target != "npx -y host-tool@latest" {
		t.Errorf("host-tool mismatch: %+v", servers[0])
	}

	if servers[1].Name != "prof-tool" || servers[1].Origin != "profile" || servers[1].Target != "tool" {
		t.Errorf("prof-tool mismatch: %+v", servers[1])
	}

	// shared-api was overridden by profile
	if servers[2].Name != "shared-api" || servers[2].Origin != "profile" || servers[2].Target != "https://mcp.shared.com/v2" {
		t.Errorf("shared-api mismatch: %+v", servers[2])
	}
}
