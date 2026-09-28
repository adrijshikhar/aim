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
}

func cases() []agentCase {
	return []agentCase{
		{
			name: "claude", provider: &claude.Adapter{}, bin: "claude",
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
			name: "codex", provider: &codex.Adapter{}, bin: "codex",
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
			eng := &merge.Engine{Store: merge.Store{Dir: filepath.Join(t.TempDir(), "st")}, Out: os.Stderr, Now: time.Now}
			before, _ := os.ReadFile(cols[0].ProfilePath)
			keep := func(x []merge.Change) []merge.Decision { return make([]merge.Decision, len(x)) }

			// 1. a session sees the host item; a no-change session restores the same items
			s, err := eng.Start("p", c.name, cols, merge.StartOptions{Enabled: true})
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
			afterEntries, _, _ := cols[0].Read(cols[0].ProfilePath)
			beforeEntries := parse(t, cols[0], before)
			if strings.Join(afterEntries.Order, ",") != strings.Join(beforeEntries.Order, ",") {
				t.Fatalf("1. items after no-op session = %v, want %v", afterEntries.Order, beforeEntries.Order)
			}

			// 2. a native add during a session is reported as added and kept at rest
			s, _ = eng.Start("p", c.name, cols, merge.StartOptions{Enabled: true})
			run(t, c, c.envFor(home, prof), c.add("newsrv")...)
			ch, _ = s.Diff()
			if len(ch) != 1 || ch[0].Name != "newsrv" || ch[0].Kind != merge.Added {
				t.Fatalf("2. diff = %+v", ch)
			}
			if err := s.Finish(ch, keep); err != nil {
				t.Fatal(err)
			}
			if e, _, _ := cols[0].Read(cols[0].ProfilePath); !e.Has("newsrv") || e.Has("hostsrv") {
				t.Fatalf("4. at rest = %v", e.Order)
			}

			// 5. non-session subcommands still work
			if out := run(t, c, c.envFor(home, prof), "--version"); out == "" {
				t.Fatal("5. --version produced no output")
			}
		})
	}
}
