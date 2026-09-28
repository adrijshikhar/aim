package codex

import (
	"path/filepath"
	"testing"

	"github.com/aim-cli/aim/internal/merge"
)

func TestMCPCollections_Codex(t *testing.T) {
	a := &Adapter{}
	cols := a.MCPCollections("/p", "/h")
	if len(cols) != 1 {
		t.Fatalf("collections = %d", len(cols))
	}
	c := cols[0]
	if c.ID() != "codex/mcp_servers" || c.Format != merge.TOML ||
		c.HostPath != filepath.Join("/h", ".codex", "config.toml") ||
		c.ProfilePath != filepath.Join("/p", ".codex", "config.toml") {
		t.Fatalf("collection = %+v", c)
	}
	if !a.IsBackground([]string{"app-server"}) || !a.IsBackground([]string{"remote-control", "start"}) || a.IsBackground([]string{"exec", "x"}) {
		t.Fatal("background forms")
	}
}

// TestIsSession_Codex: codex spells version -V/--version (`-v` is rejected
// with "unexpected argument", which aim leaves to codex) and has a help
// subcommand but no version one.
func TestIsSession_Codex(t *testing.T) {
	a := &Adapter{}
	cases := []struct {
		args []string
		want bool
	}{
		{nil, true},
		{[]string{"exec", "x"}, true},
		{[]string{"mcp", "list"}, true},
		{[]string{"-V"}, false},
		{[]string{"--version"}, false},
		{[]string{"-h"}, false},
		{[]string{"--help"}, false},
		{[]string{"help"}, false},
		{[]string{"help", "exec"}, false},
		{[]string{"-v"}, true}, // not a codex flag: codex errors, aim does not special-case it
		{[]string{"version"}, true},
		{[]string{"exec", "--", "-V"}, true},
	}
	for _, c := range cases {
		if got := a.IsSession(c.args); got != c.want {
			t.Errorf("%v: got %v want %v", c.args, got, c.want)
		}
	}
}
