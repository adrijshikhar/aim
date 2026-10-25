package agy

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/merge"
)

func TestMCPCollections_Agy(t *testing.T) {
	home := t.TempDir()
	_ = os.MkdirAll(filepath.Join(home, ".gemini", "config"), 0o755)
	a := &Adapter{}
	cols := a.MCPCollections("/p", home)
	if len(cols) != 1 || cols[0].ID() != "agy/mcpServers" || cols[0].Group != "mcp" || cols[0].Noun != "server" {
		t.Fatalf("collections = %+v", cols)
	}
	c := cols[0]
	if c.ProfilePath != filepath.Join("/p", ".gemini", "config", "mcp_config.json") ||
		c.HostPath != filepath.Join(home, ".gemini", "config", "mcp_config.json") {
		t.Fatalf("paths = %+v", c)
	}
	if merge.Hash(c.Normalise, map[string]any{"command": "x", "disabled": false}) != merge.Hash(c.Normalise, map[string]any{"command": "x"}) {
		t.Fatal(`"disabled": false is agy noise`)
	}
	if !a.IsBackground(nil, []string{"remote-control"}) || a.IsBackground(nil, []string{"chat"}) {
		t.Fatal("background forms")
	}
}

func TestSharedConfigDir_FallsBackUnderAimHome(t *testing.T) {
	aimHome := t.TempDir()
	t.Setenv("AIM_HOME", aimHome)
	if got := sharedConfigDir(t.TempDir()); got != filepath.Join(aimHome, "shared", "gemini-config") {
		t.Fatalf("fallback = %s", got)
	}
}

// TestIsSession_Agy: agy (1.2.12) parses Go-style flags, so -version and
// -help work beside the double-dash forms; -v is its log-verbosity flag and
// -V is undefined. `help` is a subcommand, `version` is not.
func TestIsSession_Agy(t *testing.T) {
	a := &Adapter{}
	cases := []struct {
		args []string
		want bool
	}{
		{nil, true},
		{[]string{"-c"}, true},
		{[]string{"mcp", "list"}, true},
		{[]string{"--version"}, false},
		{[]string{"-version"}, false},
		{[]string{"-h"}, false},
		{[]string{"--help"}, false},
		{[]string{"-help"}, false},
		{[]string{"help"}, false},
		{[]string{"-v", "2"}, true},
		{[]string{"-V"}, true},
		{[]string{"version"}, true},
		{[]string{"-p", "x", "--", "--help"}, true},
	}
	for _, c := range cases {
		if got := a.IsSession(nil, c.args); got != c.want {
			t.Errorf("%v: got %v want %v", c.args, got, c.want)
		}
	}
}

func TestAgy_ListMCPServers(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("AIM_REAL_HOME", home)
	t.Setenv("AIM_HOME", filepath.Join(home, ".aim"))
	sharedDir := filepath.Join(home, ".gemini", "config")
	_ = os.MkdirAll(sharedDir, 0o755)

	hostConfig := `{
  "mcpServers": {
    "host-tool": {
      "command": "npx",
      "args": ["-y", "host-tool@latest"]
    },
    "shared-api": {
      "serverUrl": "https://mcp.shared.com/v1"
    }
  }
}`
	_ = os.WriteFile(filepath.Join(sharedDir, "mcp_config.json"), []byte(hostConfig), 0o600)

	prof := filepath.Join(home, "profiles", "dev")
	profCfgDir := filepath.Join(prof, ".gemini", "config")
	_ = os.MkdirAll(profCfgDir, 0o700)

	profConfig := `{
  "mcpServers": {
    "prof-tool": {
      "command": "python",
      "args": ["-m", "tool"],
      "disabled": true
    },
    "shared-api": {
      "serverUrl": "https://mcp.shared.com/v2"
    }
  }
}`
	_ = os.WriteFile(filepath.Join(profCfgDir, "mcp_config.json"), []byte(profConfig), 0o600)

	a := &Adapter{}
	servers, err := a.ListMCPServers(context.Background(), "dev", prof)
	if err != nil {
		t.Fatalf("ListMCPServers failed: %v", err)
	}

	if len(servers) != 3 {
		t.Fatalf("expected 3 servers, got %d", len(servers))
	}

	// Sorted: host-tool, prof-tool, shared-api
	if servers[0].Name != "host-tool" || servers[0].Origin != "host" || servers[0].Status != "enabled" {
		t.Errorf("host-tool mismatch: %+v", servers[0])
	}

	if servers[1].Name != "prof-tool" || servers[1].Origin != "profile" || servers[1].Status != "disabled" {
		t.Errorf("prof-tool mismatch: %+v", servers[1])
	}

	if servers[2].Name != "shared-api" || servers[2].Origin != "profile" || servers[2].Target != "https://mcp.shared.com/v2" {
		t.Errorf("shared-api mismatch: %+v", servers[2])
	}
}

func TestAgy_AddAndRemoveHostMCPServer(t *testing.T) {
	home := t.TempDir()
	_ = os.MkdirAll(filepath.Join(home, ".gemini", "config"), 0o755)
	t.Setenv("HOME", home)
	t.Setenv("AIM_REAL_HOME", home)

	a := &Adapter{}
	ctx := context.Background()

	cfg := agents.MCPServerConfig{
		Command: "npx",
		Args:    []string{"-y", "agentation-mcp", "server"},
	}

	if err := a.AddHostMCPServer(ctx, "agentation", cfg); err != nil {
		t.Fatalf("AddHostMCPServer failed: %v", err)
	}

	agyPath := filepath.Join(home, ".gemini", "config", "mcp_config.json")
	data, err := os.ReadFile(agyPath)
	if err != nil {
		t.Fatalf("failed to read agy config: %v", err)
	}
	if !strings.Contains(string(data), "agentation-mcp") {
		t.Fatalf("expected agentation in agy config, got:\n%s", string(data))
	}

	// Remove server
	if err := a.RemoveHostMCPServer(ctx, "agentation"); err != nil {
		t.Fatalf("RemoveHostMCPServer failed: %v", err)
	}

	data, err = os.ReadFile(agyPath)
	if err != nil {
		t.Fatalf("failed to read agy config after remove: %v", err)
	}
	if strings.Contains(string(data), "agentation-mcp") {
		t.Fatalf("expected agentation to be removed, got:\n%s", string(data))
	}
}

func TestAgy_AddProfileMCPServer_Isolation(t *testing.T) {
	home := t.TempDir()
	_ = os.MkdirAll(filepath.Join(home, ".gemini", "config"), 0o755)
	profDir := filepath.Join(home, "profiles", "work")
	_ = os.MkdirAll(profDir, 0o700)
	t.Setenv("HOME", home)
	t.Setenv("AIM_REAL_HOME", home)

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
	pConfig := filepath.Join(profDir, ".gemini", "config", "mcp_config.json")
	if _, err := os.Stat(pConfig); os.IsNotExist(err) {
		t.Fatalf("profile config was not created at %s", pConfig)
	}

	// Host config must NOT have been created or modified with fetch
	hConfig := filepath.Join(home, ".gemini", "config", "mcp_config.json")
	if data, err := os.ReadFile(hConfig); err == nil && strings.Contains(string(data), "fetch") {
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
		t.Fatalf("expected fetch to be removed from profile config, got:\n%s", string(data))
	}
}

func TestAgy_HonorsNormalization(t *testing.T) {
	home := t.TempDir()
	_ = os.MkdirAll(filepath.Join(home, ".gemini", "config"), 0o755)
	t.Setenv("HOME", home)
	t.Setenv("AIM_REAL_HOME", home)

	a := &Adapter{}
	ctx := context.Background()

	cfg := agents.MCPServerConfig{
		Command:  "echo",
		Disabled: false, // Should be normalized away
	}

	if err := a.AddHostMCPServer(ctx, "norm-test", cfg); err != nil {
		t.Fatalf("AddHostMCPServer failed: %v", err)
	}

	agyPath := filepath.Join(home, ".gemini", "config", "mcp_config.json")
	data, _ := os.ReadFile(agyPath)
	if strings.Contains(string(data), `"disabled": false`) {
		t.Errorf("expected disabled: false to be dropped by normalization, got:\n%s", string(data))
	}
}

func TestAgy_DetectInstalled(t *testing.T) {
	home := t.TempDir()
	t.Setenv("PATH", t.TempDir()) // Ensure binary not in PATH
	a := &Adapter{}
	if a.DetectInstalled(home) {
		t.Errorf("expected not installed in empty home without binary in path")
	}

	_ = os.MkdirAll(filepath.Join(home, ".gemini"), 0o755)
	if !a.DetectInstalled(home) {
		t.Errorf("expected installed when ~/.gemini directory exists")
	}
}
