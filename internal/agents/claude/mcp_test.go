package claude

import (
	"path/filepath"
	"testing"

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
	if !a.IsBackground([]string{"--bg"}) || !a.IsBackground([]string{"agents"}) || a.IsBackground([]string{"--resume", "x"}) {
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
