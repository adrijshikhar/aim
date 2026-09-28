package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/agents/agy"
	"github.com/aim-cli/aim/internal/agents/claude"
	"github.com/aim-cli/aim/internal/agents/codex"
	"github.com/aim-cli/aim/internal/agents/gemini"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/merge"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/session"
	agysess "github.com/aim-cli/aim/internal/session/providers/agy"
	claudesess "github.com/aim-cli/aim/internal/session/providers/claude"
	codexsess "github.com/aim-cli/aim/internal/session/providers/codex"
)

type adapterTestCase struct {
	agent        string
	hasSessions  bool
	supportsFork bool
	// hasMCP: the adapter implements agents.MCPProvider, so host MCP servers
	// are merged into the profile per session; bgArgs is a native invocation
	// that starts a background session (merge, run, no exit step).
	hasMCP bool
	bgArgs []string
}

var allAdapters = []adapterTestCase{
	{agent: "agy", hasSessions: true, supportsFork: false, hasMCP: true, bgArgs: []string{"remote-control"}},
	{agent: "codex", hasSessions: true, supportsFork: true, hasMCP: true, bgArgs: []string{"app-server"}},
	{agent: "claude", hasSessions: true, supportsFork: true, hasMCP: true, bgArgs: []string{"--bg"}},
	{agent: "gemini", hasSessions: false, supportsFork: false, hasMCP: false},
}

func setupParameterizedRegistry() *agents.Registry {
	reg := agents.NewRegistry()
	reg.Register(agy.NewAdapter())
	reg.Register(codex.NewAdapter())
	reg.Register(claude.NewAdapter())
	reg.Register(gemini.NewAdapter())
	return reg
}

func setupParameterizedSessionManager() *session.Manager {
	mgr := session.NewManager()
	mgr.RegisterProvider(agysess.NewProvider())
	mgr.RegisterProvider(codexsess.NewProvider())
	mgr.RegisterProvider(claudesess.NewProvider())
	return mgr
}

func TestParameterized_AllAdapters_Registration(t *testing.T) {
	reg := setupParameterizedRegistry()
	for _, tc := range allAdapters {
		t.Run(tc.agent, func(t *testing.T) {
			adapter, err := reg.Get(tc.agent)
			if err != nil {
				t.Fatalf("expected adapter %q to be registered in registry, got error: %v", tc.agent, err)
			}
			if adapter.Name() != tc.agent {
				t.Errorf("expected adapter name %q, got %q", tc.agent, adapter.Name())
			}
			if adapter.BinaryName() == "" {
				t.Errorf("expected non-empty binary name for adapter %q", tc.agent)
			}
		})
	}
}

func TestParameterized_AllAdapters_TypoGuard(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("AIM_HOME", t.TempDir())
	t.Setenv("AIM_REAL_HOME", t.TempDir())
	t.Setenv("AIM_AUTO_CREATE", "1")

	reg := setupParameterizedRegistry()
	pm := profile.NewProfileManager(t.TempDir())

	for _, tc := range allAdapters {
		t.Run(tc.agent, func(t *testing.T) {
			cmd := newRunCmd(reg, pm)
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			// Pass a profile name with missing space before flag
			cmd.SetArgs([]string{tc.agent, "work--dangerously-skip-permissions"})

			err := cmd.Execute()
			if err == nil {
				t.Fatalf("expected typo profile 'work--dangerously-skip-permissions' to be rejected for %s, got nil error", tc.agent)
			}
			errMsg := err.Error()
			if !strings.Contains(errMsg, "containing \"--\"") {
				t.Errorf("expected error to mention containing '--', got: %s", errMsg)
			}
			if !strings.Contains(errMsg, `did you mean "work" with flag "--dangerously-skip-permissions"`) {
				t.Errorf("expected error to suggest 'work' with flag, got: %s", errMsg)
			}
		})
	}
}

func TestParameterized_AllAdapters_Doctor(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("AIM_HOME", t.TempDir())
	t.Setenv("AIM_REAL_HOME", t.TempDir())

	reg := setupParameterizedRegistry()
	pm := profile.NewProfileManager(t.TempDir())

	for _, tc := range allAdapters {
		t.Run(tc.agent, func(t *testing.T) {
			adapter, err := reg.Get(tc.agent)
			if err != nil {
				t.Fatalf("failed to get adapter: %v", err)
			}
			profDir := filepath.Join(t.TempDir(), "prof-"+tc.agent)
			_ = os.MkdirAll(profDir, 0755)

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			results := adapter.Doctor(ctx, "prof-"+tc.agent, profDir)
			if len(results) == 0 {
				t.Errorf("expected doctor diagnostic results for adapter %q", tc.agent)
			}
			_ = pm
		})
	}
}

func TestParameterized_NonSessionAdapters_ResumeRejection(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("AIM_HOME", t.TempDir())
	t.Setenv("AIM_REAL_HOME", t.TempDir())

	reg := setupParameterizedRegistry()
	pm := profile.NewProfileManager(t.TempDir())
	_, _ = pm.EnsureProfile("office")

	oldMgr := defaultSessionManager
	defer func() { defaultSessionManager = oldMgr }()
	defaultSessionManager = setupParameterizedSessionManager

	for _, tc := range allAdapters {
		if tc.hasSessions {
			continue
		}
		t.Run(tc.agent, func(t *testing.T) {
			cmd := newResumeCmd(reg, pm)
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs([]string{tc.agent, "office", "any-session-id"})

			err := cmd.Execute()
			if err == nil {
				t.Fatalf("expected resume to fail for non-session adapter %q, but got nil", tc.agent)
			}
		})
	}
}

func TestParameterized_SessionAdapters_ListingAndHydration(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)

	reg := setupParameterizedRegistry()
	_ = reg
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("work")
	_, _ = pm.EnsureProfile("office")

	mgr := setupParameterizedSessionManager()
	oldMgr := defaultSessionManager
	defer func() { defaultSessionManager = oldMgr }()
	defaultSessionManager = func() *session.Manager { return mgr }

	ctx := context.Background()

	for _, tc := range allAdapters {
		if !tc.hasSessions {
			continue
		}
		t.Run(tc.agent, func(t *testing.T) {
			prov := mgr.Provider(tc.agent)
			if prov == nil {
				t.Fatalf("expected session provider for %q", tc.agent)
			}
			if prov.Agent() != tc.agent {
				t.Errorf("expected provider agent %q, got %q", tc.agent, prov.Agent())
			}

			// Test listing empty profile
			workDir := filepath.Join(tempDir, "profiles", "work")
			sessions, err := prov.ListSessions(ctx, workDir, false)
			if err != nil {
				t.Fatalf("ListSessions failed for %s: %v", tc.agent, err)
			}
			_ = sessions

			// Test Hydrate interface with nil session
			_, err = prov.Hydrate(ctx, nil, filepath.Join(tempDir, "profiles", "office"), false)
			if err == nil {
				t.Errorf("expected Hydrate to fail with nil session for %s", tc.agent)
			}
		})
	}
}

// TestParameterized_AllAdapters_MCPSessionMerge drives the launch wrapper the
// way `aim run` does: adapters with an MCPProvider see the host's servers for
// the session only, keep their own additions at rest, honour mcp_global:false,
// and recover a background launch (from CLI or profile args) on the next one;
// the rest launch untouched.
func TestParameterized_AllAdapters_MCPSessionMerge(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("AIM_HOME", t.TempDir())
	oldTTY := stdinIsTerminal
	stdinIsTerminal = func() bool { return false } // no terminal: the exit prompt keeps
	defer func() { stdinIsTerminal = oldTTY }()

	reg := setupParameterizedRegistry()
	server := map[string]any{"command": "true"}

	for _, tc := range allAdapters {
		t.Run(tc.agent, func(t *testing.T) {
			adapter, err := reg.Get(tc.agent)
			if err != nil {
				t.Fatalf("failed to get adapter: %v", err)
			}
			mp, ok := adapter.(agents.MCPProvider)
			if ok != tc.hasMCP {
				t.Fatalf("adapter %q implements MCPProvider=%v, table says %v", tc.agent, ok, tc.hasMCP)
			}

			realHome := t.TempDir()
			t.Setenv("AIM_REAL_HOME", realHome)
			profDir := filepath.Join(t.TempDir(), "profiles", "work")
			store := merge.Store{Dir: filepath.Join(t.TempDir(), "profile-merge")}
			cfg := &config.Config{Profiles: map[string]config.ProfileConfig{}}
			launch := func(args []string, run func() int) int {
				return withSessionMerge(adapter, store, "work", profDir, cfg, args, run)
			}

			if !tc.hasMCP {
				ran := false
				if code := launch(nil, func() int { ran = true; return 7 }); code != 7 || !ran {
					t.Fatalf("launch without MCPProvider: ran=%v code=%d, want ran=true code=7", ran, code)
				}
				return
			}

			// agy's host layer is the real ~/.gemini/config when it exists
			_ = os.MkdirAll(filepath.Join(realHome, ".gemini", "config"), 0o700)
			c := mp.MCPCollections(profDir, realHome)[0]
			write := func(path string, names ...string) {
				e := merge.NewEntries()
				for _, n := range names {
					e.Set(n, server, nil)
				}
				_ = os.MkdirAll(filepath.Dir(path), 0o700)
				if _, err := os.Stat(path); err != nil {
					skeleton := []byte("{}\n")
					if c.Format == merge.TOML {
						skeleton = []byte("")
					}
					_ = os.WriteFile(path, skeleton, 0o600)
				}
				if _, err := c.Write(path, e); err != nil {
					t.Fatalf("write %s: %v", path, err)
				}
			}
			names := func(path string) string {
				e, _, err := c.Read(path)
				if err != nil {
					t.Fatalf("read %s: %v", path, err)
				}
				order := append([]string(nil), e.Order...)
				sort.Strings(order)
				return strings.Join(order, ",")
			}
			expect := func(what, path, want string) {
				t.Helper()
				if got := names(path); got != want {
					t.Errorf("%s: servers = %q, want %q", what, got, want)
				}
			}
			write(c.HostPath, "hostsrv")
			write(c.ProfilePath, "own")

			// 0. non-session (--version): runs with no merge, no strip, no merge state
			before, _ := os.ReadFile(c.ProfilePath)
			ran := false
			if code := launch([]string{"--version"}, func() int { ran = true; return 0 }); code != 0 || !ran {
				t.Fatalf("--version launch: ran=%v code=%d, want ran=true code=0", ran, code)
			}
			if after, _ := os.ReadFile(c.ProfilePath); string(after) != string(before) {
				t.Errorf("--version launch rewrote the profile file:\n%s\n%s", before, after)
			}
			if _, err := os.Stat(filepath.Join(store.Dir, "work.json")); !os.IsNotExist(err) {
				t.Errorf("--version launch created merge state (stat err = %v)", err)
			}

			// 1. foreground: host merged for the session; a native add is kept at rest
			var during string
			launch(nil, func() int {
				during = names(c.ProfilePath)
				e, _, _ := c.Read(c.ProfilePath)
				e.Set("newsrv", server, nil)
				if _, err := c.Write(c.ProfilePath, e); err != nil {
					t.Errorf("native add during session: %v", err)
				}
				return 0
			})
			if during != "hostsrv,own" {
				t.Errorf("during session: servers = %q, want %q", during, "hostsrv,own")
			}
			expect("at rest after keep", c.ProfilePath, "newsrv,own")
			expect("host after keep", c.HostPath, "hostsrv")

			// 2. mcp_global: false — the profile runs on its own servers only
			off := false
			cfg.Profiles["work"] = config.ProfileConfig{MCPGlobal: &off}
			launch(nil, func() int { during = names(c.ProfilePath); return 0 })
			if during != "newsrv,own" {
				t.Errorf("mcp_global=false: servers = %q, want %q", during, "newsrv,own")
			}
			delete(cfg.Profiles, "work")

			// 3. background: merged with no exit step; the next launch recovers and strips
			launch(tc.bgArgs, func() int { return 0 })
			expect("after background launch", c.ProfilePath, "hostsrv,newsrv,own")
			launch(nil, func() int { during = names(c.ProfilePath); return 0 })
			if during != "hostsrv,newsrv,own" {
				t.Errorf("recovered session: servers = %q, want %q", during, "hostsrv,newsrv,own")
			}
			expect("at rest after recovery", c.ProfilePath, "newsrv,own")
			expect("host after recovery", c.HostPath, "hostsrv")

			// 4. background via profiles.work.args, no CLI args: still no exit step
			cfg.Profiles["work"] = config.ProfileConfig{Args: tc.bgArgs}
			launch(nil, func() int { return 0 })
			expect("after background launch from profile args", c.ProfilePath, "hostsrv,newsrv,own")
			delete(cfg.Profiles, "work")
			launch(nil, func() int { during = names(c.ProfilePath); return 0 })
			if during != "hostsrv,newsrv,own" {
				t.Errorf("recovered session (profile args): servers = %q, want %q", during, "hostsrv,newsrv,own")
			}
			expect("at rest after profile-args recovery", c.ProfilePath, "newsrv,own")
		})
	}
}
