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
	if !a.IsBackground(nil, []string{"app-server"}) || !a.IsBackground(nil, []string{"remote-control", "start"}) || a.IsBackground(nil, []string{"exec", "x"}) {
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
		if got := a.IsSession(nil, c.args); got != c.want {
			t.Errorf("%v: got %v want %v", c.args, got, c.want)
		}
	}
}

// Profile args such as `--model opus` precede the CLI's in the launched
// command; they must not hide the CLI's first word from either check.
func TestArgs_Codex_ProfileArgsFirst(t *testing.T) {
	a := &Adapter{}
	opus := []string{"--model", "opus"}
	if a.IsSession(opus, []string{"help"}) {
		t.Error("codex --model opus help prints help: not a session")
	}
	if a.IsSession(opus, []string{"-V"}) {
		t.Error("codex --model opus -V prints the version: not a session")
	}
	if !a.IsSession(opus, []string{"exec", "help"}) {
		t.Error("codex --model opus exec help is a session")
	}
	if !a.IsBackground(opus, []string{"app-server"}) {
		t.Error("codex --model opus app-server is a background session")
	}
	if !a.IsBackground([]string{"app-server"}, nil) {
		t.Error("profiles.<p>.args [app-server] is a background session")
	}
	if a.IsBackground(opus, []string{"exec", "app-server"}) {
		t.Error("codex --model opus exec app-server is a prompt, not the app server")
	}
}
