package claude

import (
	"path/filepath"
	"testing"

	"github.com/aim-cli/aim/internal/merge"
)

func TestMCPCollections_Claude(t *testing.T) {
	a := &Adapter{}
	cols := a.MCPCollections("/p", "/h")
	if len(cols) != 1 {
		t.Fatalf("collections = %d", len(cols))
	}
	c := cols[0]
	if c.ID() != "claude/mcpServers" || c.Format != merge.JSON || c.Key != "mcpServers" ||
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
