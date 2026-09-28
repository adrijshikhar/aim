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
	if !a.IsBackground(nil, []string{"app-server"}) || !a.IsBackground(nil, []string{"remote-control", "start"}) || a.IsBackground(nil, []string{"exec", "x"}) {
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
