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
		if got := a.IsSession(c.args); got != c.want {
			t.Errorf("%v: got %v want %v", c.args, got, c.want)
		}
	}
}
