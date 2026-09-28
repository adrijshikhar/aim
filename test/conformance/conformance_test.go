//go:build conformance

package conformance

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/agents/claude"
	"github.com/aim-cli/aim/internal/agents/codex"
	"github.com/aim-cli/aim/internal/merge"
)

type agentCase struct {
	name     string
	provider agents.MCPProvider
	bin      string
	envFor   func(home, profile string) []string
	list     []string // native argv listing MCP servers
	add      func(name string) []string
	seedHost func(t *testing.T, home string)
	seedProf func(t *testing.T, profile string)
	mcp      string // the MCP servers collection's Name
	// plugins checks a native switch-off of a host plugin inside a session;
	// nil for adapters without a plugins collection.
	plugins func(t *testing.T, c agentCase, eng *merge.Engine, home, prof string)
}

func cases() []agentCase {
	return []agentCase{
		{
			name: "claude", provider: &claude.Adapter{}, bin: "claude", mcp: "mcpServers", plugins: claudePlugins,
			envFor: func(home, p string) []string {
				return append(os.Environ(), "HOME="+p, "CLAUDE_CONFIG_DIR="+filepath.Join(p, ".claude"))
			},
			list: []string{"mcp", "list"},
			add:  func(n string) []string { return []string{"mcp", "add", "-s", "user", n, "--", "true"} },
			seedHost: func(t *testing.T, h string) {
				_ = os.WriteFile(filepath.Join(h, ".claude.json"), []byte(`{"mcpServers":{"hostsrv":{"command":"true"}}}`), 0o600)
			},
			seedProf: func(t *testing.T, p string) {
				_ = os.MkdirAll(filepath.Join(p, ".claude"), 0o700)
				_ = os.WriteFile(filepath.Join(p, ".claude", ".claude.json"), []byte(`{"mcpServers":{"own":{"command":"true"}}}`), 0o600)
			},
		},
		{
			name: "codex", provider: &codex.Adapter{}, bin: "codex", mcp: "mcp_servers", plugins: codexPlugins,
			envFor: func(home, p string) []string {
				return append(os.Environ(), "HOME="+p, "CODEX_HOME="+filepath.Join(p, ".codex"))
			},
			list: []string{"mcp", "list"},
			add:  func(n string) []string { return []string{"mcp", "add", n, "--", "true"} },
			seedHost: func(t *testing.T, h string) {
				_ = os.MkdirAll(filepath.Join(h, ".codex"), 0o700)
				_ = os.WriteFile(filepath.Join(h, ".codex", "config.toml"), []byte("[mcp_servers.hostsrv]\ncommand = \"true\"\n"), 0o600)
			},
			seedProf: func(t *testing.T, p string) {
				_ = os.MkdirAll(filepath.Join(p, ".codex"), 0o700)
				_ = os.WriteFile(filepath.Join(p, ".codex", "config.toml"), []byte("model = \"x\"\n\n[mcp_servers.own]\ncommand = \"true\"\n"), 0o600)
			},
		},
	}
}

func parse(t *testing.T, c merge.Collection, data []byte) merge.Entries {
	p := filepath.Join(t.TempDir(), filepath.Base(c.ProfilePath))
	_ = os.WriteFile(p, data, 0o600)
	e, _, err := c.Read(p)
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func named(t *testing.T, cols []merge.Collection, name string) merge.Collection {
	t.Helper()
	for _, c := range cols {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("no %s collection in %+v", name, cols)
	return merge.Collection{}
}

func keep(x []merge.Change) []merge.Decision { return make([]merge.Decision, len(x)) }

func run(t *testing.T, c agentCase, env []string, argv ...string) string {
	cmd := exec.Command(c.bin, argv...)
	cmd.Env = env
	out, _ := cmd.CombinedOutput()
	return string(out)
}

func TestConformance(t *testing.T) {
	for _, c := range cases() {
		c := c
		t.Run(c.name, func(t *testing.T) {
			if _, err := exec.LookPath(c.bin); err != nil {
				t.Skip(c.bin + " not installed")
			}
			t.Setenv("AIM_HOME", t.TempDir()) // never the real ~/.aim
			home, prof := t.TempDir(), t.TempDir()
			c.seedHost(t, home)
			c.seedProf(t, prof)
			cols := c.provider.MCPCollections(prof, home)
			mcp := named(t, cols, c.mcp)
			eng := &merge.Engine{Store: merge.Store{Dir: filepath.Join(t.TempDir(), "st")}, Out: os.Stderr, Now: time.Now}
			before, _ := os.ReadFile(mcp.ProfilePath)

			// 1. a session sees the host item; a no-change session restores the same items
			s, err := eng.Start("p", c.name, cols, merge.StartOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if out := run(t, c, c.envFor(home, prof), c.list...); !strings.Contains(out, "hostsrv") || !strings.Contains(out, "own") {
				t.Fatalf("session does not see host+own:\n%s", out)
			}
			ch, _ := s.Diff()
			if len(ch) != 0 {
				t.Fatalf("3. agent's own writes reported as changes: %+v", ch)
			}
			if err := s.Finish(ch, keep); err != nil {
				t.Fatal(err)
			}
			afterEntries, _, _ := mcp.Read(mcp.ProfilePath)
			beforeEntries := parse(t, mcp, before)
			if strings.Join(afterEntries.Order, ",") != strings.Join(beforeEntries.Order, ",") {
				t.Fatalf("1. items after no-op session = %v, want %v", afterEntries.Order, beforeEntries.Order)
			}

			// 2. a native add during a session is reported as added and kept at rest
			s, _ = eng.Start("p", c.name, cols, merge.StartOptions{})
			run(t, c, c.envFor(home, prof), c.add("newsrv")...)
			ch, _ = s.Diff()
			if len(ch) != 1 || ch[0].Name != "newsrv" || ch[0].Kind != merge.Added {
				t.Fatalf("2. diff = %+v", ch)
			}
			if err := s.Finish(ch, keep); err != nil {
				t.Fatal(err)
			}
			if e, _, _ := mcp.Read(mcp.ProfilePath); !e.Has("newsrv") || e.Has("hostsrv") {
				t.Fatalf("4. at rest = %v", e.Order)
			}

			// 5. non-session subcommands still work
			if out := run(t, c, c.envFor(home, prof), "--version"); out == "" {
				t.Fatal("5. --version produced no output")
			}

			// 6. switching a host plugin off inside a session is one edit of a
			// host item; kept, it is the profile's own from then on
			if c.plugins != nil {
				c.plugins(t, c, eng, home, prof)
			}
		})
	}
}

// claudePlugins installs an offline marketplace plugin on the host, shares the
// plugin registry with the profile as aim does, and drives the native
// `claude plugin disable` / `enable` inside sessions.
func claudePlugins(t *testing.T, c agentCase, eng *merge.Engine, home, prof string) {
	mkt := filepath.Join(t.TempDir(), "mkt")
	_ = os.MkdirAll(filepath.Join(mkt, ".claude-plugin"), 0o755)
	_ = os.MkdirAll(filepath.Join(mkt, "plugins", "probe", ".claude-plugin"), 0o755)
	_ = os.WriteFile(filepath.Join(mkt, ".claude-plugin", "marketplace.json"),
		[]byte(`{"name":"localmkt","owner":{"name":"x"},"plugins":[{"name":"probe","source":"./plugins/probe","description":"p"}]}`), 0o644)
	_ = os.WriteFile(filepath.Join(mkt, "plugins", "probe", ".claude-plugin", "plugin.json"),
		[]byte(`{"name":"probe","version":"0.0.1","description":"p"}`), 0o644)
	_ = os.MkdirAll(filepath.Join(home, ".claude"), 0o700)
	hostEnv := append(os.Environ(), "HOME="+home, "CLAUDE_CONFIG_DIR="+filepath.Join(home, ".claude"))
	run(t, c, hostEnv, "plugin", "marketplace", "add", mkt)
	run(t, c, hostEnv, "plugin", "install", "probe@localmkt")
	const id = "probe@localmkt"
	cols := c.provider.MCPCollections(prof, home)
	ep := named(t, cols, "enabledPlugins")
	if e, _, _ := ep.Read(ep.HostPath); e.Values[id]["value"] != true {
		t.Fatalf("6. host install did not enable %s: %+v", id, e.Values)
	}
	if err := os.Symlink(filepath.Join(home, ".claude", "plugins"), filepath.Join(prof, ".claude", "plugins")); err != nil {
		t.Fatal(err)
	}
	env := c.envFor(home, prof)

	s, err := eng.Start("p", c.name, cols, merge.StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if e, _, _ := ep.Read(ep.ProfilePath); e.Values[id]["value"] != true {
		t.Fatalf("6. host enablement not merged: %+v", e.Values)
	}
	out := run(t, c, env, "plugin", "disable", id)
	ch, _ := s.Diff()
	if len(ch) != 1 || ch[0].Collection.Name != "enabledPlugins" || ch[0].Name != id || ch[0].Kind != merge.Edited || ch[0].Value["value"] != false {
		t.Fatalf("6. diff after disable = %+v\n%s", ch, out)
	}
	run(t, c, env, "plugin", "enable", id)
	if ch, _ := s.Diff(); len(ch) != 0 {
		t.Fatalf("6. enabling back is no change: %+v", ch)
	}
	run(t, c, env, "plugin", "disable", id)
	ch, _ = s.Diff()
	if err := s.Finish(ch, keep); err != nil {
		t.Fatal(err)
	}
	if e, _, _ := ep.Read(ep.ProfilePath); e.Values[id]["value"] != false {
		t.Fatalf("6. a kept disable stays in the profile: %+v", e.Values)
	}

	s, err = eng.Start("p", c.name, cols, merge.StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if e, _, _ := ep.Read(ep.ProfilePath); e.Values[id]["value"] != false {
		t.Fatalf("6. the profile's false must shadow the host's true: %+v", e.Values)
	}
	run(t, c, env, "plugin", "enable", id)
	ch, _ = s.Diff()
	if len(ch) != 0 {
		t.Fatalf("6. the profile's own entry is not asked about: %+v", ch)
	}
	if err := s.Finish(ch, keep); err != nil {
		t.Fatal(err)
	}
}

// codexPlugins edits the TOML by hand: codex has no native enable/disable.
func codexPlugins(t *testing.T, c agentCase, eng *merge.Engine, home, prof string) {
	const header = `[plugins."probe@local"]`
	hostCfg := filepath.Join(home, ".codex", "config.toml")
	data, _ := os.ReadFile(hostCfg)
	_ = os.WriteFile(hostCfg, append(data, []byte("\n"+header+"\nenabled = true\n")...), 0o600)
	cols := c.provider.MCPCollections(prof, home)
	pl := named(t, cols, "plugins")

	s, err := eng.Start("p", c.name, cols, merge.StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(pl.ProfilePath)
	i := strings.Index(string(b), header)
	if i < 0 {
		t.Fatalf("6. host plugin not merged:\n%s", b)
	}
	j := i + strings.Index(string(b)[i:], "enabled = true")
	_ = os.WriteFile(pl.ProfilePath, []byte(string(b)[:j]+"enabled = false"+string(b)[j+len("enabled = true"):]), 0o600)
	if out := run(t, c, c.envFor(home, prof), "--version"); out == "" {
		t.Fatal("6. codex does not start with the edited config")
	}
	ch, _ := s.Diff()
	if len(ch) != 1 || ch[0].Collection.Name != "plugins" || ch[0].Kind != merge.Edited || ch[0].Value["enabled"] != false {
		t.Fatalf("6. diff after disable = %+v", ch)
	}
	if err := s.Finish(ch, keep); err != nil {
		t.Fatal(err)
	}
	s, err = eng.Start("p", c.name, cols, merge.StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if e, _, _ := pl.Read(pl.ProfilePath); e.Values["probe@local"]["enabled"] != false {
		t.Fatalf("6. the profile's false must shadow the host's true: %+v", e.Values)
	}
	ch, _ = s.Diff()
	_ = s.Finish(ch, keep)
}
