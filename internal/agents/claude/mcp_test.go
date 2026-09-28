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
	if !a.IsBackground([]string{"--bg"}) || !a.IsBackground([]string{"agents"}) || a.IsBackground([]string{"--resume", "x"}) {
		t.Fatal("background forms")
	}
}
