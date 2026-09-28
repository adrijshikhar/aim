package agy

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aim-cli/aim/internal/merge"
)

func TestMCPCollections_Agy(t *testing.T) {
	home := t.TempDir()
	_ = os.MkdirAll(filepath.Join(home, ".gemini", "config"), 0o755)
	a := &Adapter{}
	cols := a.MCPCollections("/p", home)
	if len(cols) != 1 || cols[0].ID() != "agy/mcpServers" {
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
	if !a.IsBackground([]string{"remote-control"}) || a.IsBackground([]string{"chat"}) {
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
