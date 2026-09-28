package codex

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aim-cli/aim/internal/merge"
)

func TestMCPCollections_Codex(t *testing.T) {
	a := &Adapter{}
	cols := a.MCPCollections("/p", "/h")
	if len(cols) != 2 {
		t.Fatalf("collections = %d", len(cols))
	}
	c := cols[0]
	if c.ID() != "codex/mcp_servers" || c.Format != merge.TOML || c.Group != "mcp" || c.Noun != "server" ||
		c.HostPath != filepath.Join("/h", ".codex", "config.toml") ||
		c.ProfilePath != filepath.Join("/p", ".codex", "config.toml") {
		t.Fatalf("collection = %+v", c)
	}
	if p := cols[1]; p.ID() != "codex/plugins" || p.Key != "plugins" || p.Format != merge.TOML || p.Group != "plugins" || p.Noun != "plugin" ||
		p.HostPath != c.HostPath || p.ProfilePath != c.ProfilePath {
		t.Fatalf("plugins collection = %+v", p)
	}
	if !a.IsBackground([]string{"app-server"}) || !a.IsBackground([]string{"remote-control", "start"}) || a.IsBackground([]string{"exec", "x"}) {
		t.Fatal("background forms")
	}
}

// A plugin's own MCP server sub-table moves with the plugin, not with the
// mcp_servers collection, and the strip gives the profile file back byte for byte.
func TestPluginsCollection_MergeAndStripWithSubTable(t *testing.T) {
	home, prof := t.TempDir(), t.TempDir()
	_ = os.MkdirAll(filepath.Join(home, ".codex"), 0o700)
	_ = os.MkdirAll(filepath.Join(prof, ".codex"), 0o700)
	host := "model = \"x\"\n\n[mcp_servers.jev]\ncommand = \"npx\"\n\n" +
		"[plugins.\"hevo@hevo\"]\nenabled = true\n\n[plugins.\"hevo@hevo\".mcp_servers.atlassian]\nenabled = false\n\n" +
		"[plugins.\"claude-mem@thedotmack\"]\nenabled = false\n"
	src := "model = \"y\"\n\n[plugins.\"mine@local\"]\nenabled = true\n\n[hooks.state.\"mine@local:stop\"]\nenabled = false\n"
	_ = os.WriteFile(filepath.Join(home, ".codex", "config.toml"), []byte(host), 0o600)
	pf := filepath.Join(prof, ".codex", "config.toml")
	_ = os.WriteFile(pf, []byte(src), 0o600)
	cols := (&Adapter{}).MCPCollections(prof, home)
	eng := &merge.Engine{Store: merge.Store{Dir: filepath.Join(t.TempDir(), "state")}, Out: &bytes.Buffer{}}
	s, err := eng.Start("work", "codex", cols, merge.StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	during, _ := os.ReadFile(pf)
	for _, h := range []string{`[mcp_servers.jev]`, `[plugins."hevo@hevo"]`, `[plugins."hevo@hevo".mcp_servers.atlassian]`, `[plugins."claude-mem@thedotmack"]`, `[plugins."mine@local"]`} {
		if !strings.Contains(string(during), h) {
			t.Fatalf("session file lacks %s:\n%s", h, during)
		}
	}
	servers, _, _ := cols[0].Read(pf)
	plugins, _, _ := cols[1].Read(pf)
	if strings.Join(servers.Order, ",") != "jev" || len(plugins.Order) != 3 {
		t.Fatalf("servers = %v, plugins = %v", servers.Order, plugins.Order)
	}
	ch, err := s.Diff()
	if err != nil || len(ch) != 0 {
		t.Fatalf("changes = %+v, err = %v", ch, err)
	}
	if err := s.Finish(ch, func(c []merge.Change) []merge.Decision { return make([]merge.Decision, len(c)) }); err != nil {
		t.Fatal(err)
	}
	if after, _ := os.ReadFile(pf); string(after) != src {
		t.Fatalf("at rest =\n%s\nwant\n%s", after, src)
	}
}
