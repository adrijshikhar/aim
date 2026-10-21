package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
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
	"github.com/aim-cli/aim/internal/usage"
)

type adapterTestCase struct {
	agent        string
	hasSessions  bool
	supportsFork bool
	// hasMCP: the adapter implements agents.MCPProvider, so host MCP servers
	// are merged into the profile per session; bgArgs is a native invocation
	// that starts a background session (merge, run, no exit step), and
	// versionArgs the CLI's own version flag (no session, no merge).
	hasMCP      bool
	bgArgs      []string
	versionArgs []string
	// hasPlugins: the adapter also declares a "plugins" group collection named
	// pluginsCol, whose entries are enablement flags (Spec B).
	hasPlugins bool
	pluginsCol string
}

var allAdapters = []adapterTestCase{
	{agent: "agy", hasSessions: true, supportsFork: false, hasMCP: true, bgArgs: []string{"remote-control"}, versionArgs: []string{"--version"}},
	{agent: "codex", hasSessions: true, supportsFork: true, hasMCP: true, bgArgs: []string{"app-server"}, versionArgs: []string{"-V"}, hasPlugins: true, pluginsCol: "plugins"},
	{agent: "claude", hasSessions: true, supportsFork: true, hasMCP: true, bgArgs: []string{"--bg"}, versionArgs: []string{"-v"}, hasPlugins: true, pluginsCol: "enabledPlugins"},
	{agent: "gemini", hasSessions: false, supportsFork: false, hasMCP: false},
}

// collectionNamed returns the collection in group with the given name, or
// fails the test.
func collectionNamed(t *testing.T, cols []merge.Collection, group, name string) merge.Collection {
	t.Helper()
	for _, c := range cols {
		if c.Group == group && c.Name == name {
			return c
		}
	}
	t.Fatalf("no %s collection %q in %+v", group, name, cols)
	return merge.Collection{}
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

			var results []agents.DiagnosticResult
			if diag, ok := adapter.(agents.Diagnostician); ok {
				results = diag.Doctor(ctx, "prof-"+tc.agent, profDir)
			}
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

	mgr := setupParameterizedSessionManager()
	ctx := WithSessionManager(context.Background(), mgr)

	for _, tc := range allAdapters {
		if tc.hasSessions {
			continue
		}
		t.Run(tc.agent, func(t *testing.T) {
			cmd := newResumeCmd(reg, pm)
			cmd.SetContext(ctx)
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
			cols := mp.MCPCollections(profDir, realHome)
			hasPlugins := false
			for _, c := range cols {
				hasPlugins = hasPlugins || c.Group == "plugins"
			}
			if hasPlugins != tc.hasPlugins {
				t.Fatalf("adapter %q declares a plugins collection=%v, table says %v", tc.agent, hasPlugins, tc.hasPlugins)
			}
			mcpName := "mcpServers"
			if tc.agent == "codex" {
				mcpName = "mcp_servers"
			}
			t.Run("mcp", func(t *testing.T) {
				runCollectionMerge(t, launch, cfg, store, tc.bgArgs, tc.versionArgs, collectionNamed(t, cols, "mcp", mcpName),
					map[string]any{"command": "true"}, func(p *config.ProfileConfig, v *bool) { p.MCPGlobal = v })
			})
			if tc.hasPlugins {
				value := map[string]any{"value": true} // a JSON bool, as the codec wraps it
				if tc.agent == "codex" {
					value = map[string]any{"enabled": true}
				}
				t.Run("plugins", func(t *testing.T) {
					runCollectionMerge(t, launch, cfg, store, tc.bgArgs, tc.versionArgs, collectionNamed(t, cols, "plugins", tc.pluginsCol),
						value, func(p *config.ProfileConfig, v *bool) { p.PluginsGlobal = v })
				})
			}
		})
	}
}

// runCollectionMerge drives one collection through foreground keep, its
// *_global switch, background launches and recovery; setFlag sets the switch
// for the collection's group.
func runCollectionMerge(t *testing.T, launch func([]string, func() int) int, cfg *config.Config, store merge.Store, bgArgs, versionArgs []string,
	c merge.Collection, value map[string]any, setFlag func(*config.ProfileConfig, *bool)) {
	write := func(path string, names ...string) {
		e, _, _ := c.Read(path)
		for _, n := range names {
			e.Set(n, value, nil)
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
			t.Errorf("%s: %s = %q, want %q", what, c.Name, got, want)
		}
	}
	switchOff := func() {
		off := false
		p := config.ProfileConfig{}
		setFlag(&p, &off)
		cfg.Profiles["work"] = p
	}
	// ids carry "@" like plugin ids, so a TOML collection needs quoted headers
	write(c.HostPath, "host@m")
	write(c.ProfilePath, "own@m")

	// 0. non-session (the CLI's own version flag): runs with no merge, no strip
	// and no merge state for this collection
	before, _ := os.ReadFile(c.ProfilePath)
	ran := false
	if code := launch(versionArgs, func() int { ran = true; return 0 }); code != 0 || !ran {
		t.Fatalf("%v launch: ran=%v code=%d, want ran=true code=0", versionArgs, ran, code)
	}
	if after, _ := os.ReadFile(c.ProfilePath); string(after) != string(before) {
		t.Errorf("%v launch rewrote the profile file:\n%s\n%s", versionArgs, before, after)
	}
	if st, err := store.Load("work"); err != nil {
		t.Fatal(err)
	} else if _, ok := st.Active[c.ID()]; ok {
		t.Errorf("%v launch merged %s", versionArgs, c.ID())
	}

	// 1. foreground: host merged for the session; a native add is kept at rest
	var during string
	launch(nil, func() int {
		during = names(c.ProfilePath)
		write(c.ProfilePath, "new@m")
		return 0
	})
	if during != "host@m,own@m" {
		t.Errorf("during session: %s = %q, want %q", c.Name, during, "host@m,own@m")
	}
	expect("at rest after keep", c.ProfilePath, "new@m,own@m")
	expect("host after keep", c.HostPath, "host@m")

	// 2. the group's *_global: false — the profile runs on its own entries only
	switchOff()
	launch(nil, func() int { during = names(c.ProfilePath); return 0 })
	if during != "new@m,own@m" {
		t.Errorf("switched off: %s = %q, want %q", c.Name, during, "new@m,own@m")
	}
	delete(cfg.Profiles, "work")

	// 3. background: merged with no exit step; the next launch recovers and strips
	launch(bgArgs, func() int { return 0 })
	expect("after background launch", c.ProfilePath, "host@m,new@m,own@m")
	launch(nil, func() int { during = names(c.ProfilePath); return 0 })
	if during != "host@m,new@m,own@m" {
		t.Errorf("recovered session: %s = %q, want %q", c.Name, during, "host@m,new@m,own@m")
	}
	expect("at rest after recovery", c.ProfilePath, "new@m,own@m")
	expect("host after recovery", c.HostPath, "host@m")

	// 4. background via profiles.work.args, no CLI args: still no exit step
	cfg.Profiles["work"] = config.ProfileConfig{Args: bgArgs}
	launch(nil, func() int { return 0 })
	expect("after background launch from profile args", c.ProfilePath, "host@m,new@m,own@m")
	delete(cfg.Profiles, "work")
	launch(nil, func() int { during = names(c.ProfilePath); return 0 })
	if during != "host@m,new@m,own@m" {
		t.Errorf("recovered session (profile args): %s = %q, want %q", c.Name, during, "host@m,new@m,own@m")
	}
	expect("at rest after profile-args recovery", c.ProfilePath, "new@m,own@m")

	// 5. background, then the group switched off: the next launch still
	// recovers, so nothing merged is left behind (G6)
	launch(bgArgs, func() int { return 0 })
	expect("after background launch", c.ProfilePath, "host@m,new@m,own@m")
	switchOff()
	launch(nil, func() int { during = names(c.ProfilePath); return 0 })
	if during != "new@m,own@m" {
		t.Errorf("switched off after background: %s = %q, want %q", c.Name, during, "new@m,own@m")
	}
	expect("at rest after switched-off recovery", c.ProfilePath, "new@m,own@m")
	delete(cfg.Profiles, "work")
}

// findHostBinary dynamically resolves an agent binary on the host using PATH
// and standard system/user directories without hardcoding username paths.
func findHostBinary(name string) string {
	if bin, err := exec.LookPath(name); err == nil {
		return bin
	}
	realHome := config.RealHomeDir()
	fallbacks := []string{
		filepath.Join(realHome, ".local", "bin", name),
		filepath.Join("/opt/homebrew", "bin", name),
		filepath.Join("/usr/local/bin", name),
		filepath.Join("/usr/bin", name),
		filepath.Join("/bin", name),
	}
	for _, fb := range fallbacks {
		if fi, err := os.Stat(fb); err == nil && !fi.IsDir() {
			return fb
		}
	}
	return ""
}

func TestParameterized_RealTools_Claude(t *testing.T) {
	bin := findHostBinary("claude")
	if bin == "" {
		t.Skip("claude binary not found on PATH")
	}

	binDir := filepath.Dir(bin)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	t.Run("VersionInvocation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, bin, "--version")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("claude --version failed: %v, output: %s", err, string(out))
		}
		outStr := string(out)
		if !strings.Contains(outStr, "Claude") && !strings.Contains(outStr, "2.") {
			t.Errorf("unexpected output from claude --version: %s", outStr)
		}
	})

	t.Run("HelpFlag", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, bin, "--help")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("claude --help failed: %v, output: %s", err, string(out))
		}
		if !strings.Contains(string(out), "--resume") {
			t.Errorf("expected claude --help to list --resume flag, got: %s", string(out))
		}
	})

	t.Run("DoctorOK", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		tempDir := t.TempDir()
		t.Setenv("AIM_HOME", tempDir)
		t.Setenv("AIM_REAL_HOME", tempDir)
		t.Setenv("HOME", tempDir)

		reg := agents.NewRegistry()
		reg.Register(claude.NewAdapter())

		pm := profile.NewProfileManager(tempDir)
		_, err := pm.EnsureProfile("work")
		if err != nil {
			t.Fatalf("failed to ensure profile: %v", err)
		}

		cmd := newDoctorCmd(reg, pm)
		cmd.SetContext(ctx)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{"claude"})

		err = cmd.Execute()
		if err != nil {
			t.Fatalf("aim doctor claude failed: %v", err)
		}

		outStr := buf.String()
		if !strings.Contains(outStr, "[OK]") || (!strings.Contains(outStr, "Claude installed") && !strings.Contains(outStr, "installed")) {
			t.Errorf("expected aim doctor claude output to detect installed binary and return [OK], got:\n%s", outStr)
		}
	})

	t.Run("ResumeArgumentValidation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Invoke claude with --resume and a non-existent session ID in print mode (-p).
		// Claude Code CLI strictly validates --resume and reports that the provided value
		// is not a UUID / does not match any session, proving the CLI parser validates --resume.
		cmd := exec.CommandContext(ctx, bin, "--resume", "dummy-session-uuid-1234", "-p", "echo")
		out, _ := cmd.CombinedOutput()
		outStr := string(out)
		if !strings.Contains(outStr, "--resume") {
			t.Errorf("expected Claude CLI parser to accept and validate --resume flag, output was: %s", outStr)
		}
	})
}

func TestParameterized_RealTools_AllAvailable(t *testing.T) {
	tools := []struct {
		agent       string
		versionArgs []string
		identText   string
	}{
		{agent: "claude", versionArgs: []string{"--version"}, identText: "Claude"},
		{agent: "codex", versionArgs: []string{"--version"}, identText: "codex"},
		{agent: "agy", versionArgs: []string{"--version"}, identText: "1."},
	}

	for _, tc := range tools {
		t.Run(tc.agent, func(t *testing.T) {
			bin := findHostBinary(tc.agent)
			if bin == "" {
				t.Skipf("%s binary not found on PATH, skipping real invocation", tc.agent)
			}

			binDir := filepath.Dir(bin)
			t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, bin, tc.versionArgs...)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("%s %v failed: %v, output: %s", tc.agent, tc.versionArgs, err, string(out))
			}
			if !strings.Contains(strings.ToLower(string(out)), strings.ToLower(tc.identText)) {
				t.Errorf("expected version output of %s to contain %q, got: %s", tc.agent, tc.identText, string(out))
			}

			// Run aim doctor against the real binary
			tempDir := t.TempDir()
			t.Setenv("AIM_HOME", tempDir)
			t.Setenv("AIM_REAL_HOME", tempDir)
			t.Setenv("HOME", tempDir)

			reg := setupParameterizedRegistry()
			pm := profile.NewProfileManager(tempDir)
			_, _ = pm.EnsureProfile("work")

			docCmd := newDoctorCmd(reg, pm)
			docCmd.SetContext(ctx)
			var buf bytes.Buffer
			docCmd.SetOut(&buf)
			docCmd.SetErr(&buf)
			docCmd.SetArgs([]string{tc.agent})

			if err := docCmd.Execute(); err != nil {
				t.Fatalf("aim doctor %s failed: %v", tc.agent, err)
			}
			outStr := buf.String()
			if !strings.Contains(outStr, "[OK]") {
				t.Errorf("expected aim doctor %s to return [OK], got:\n%s", tc.agent, outStr)
			}
		})
	}
}

func TestParameterized_MultiAgent_SessionResumer(t *testing.T) {
	reg := setupParameterizedRegistry()

	tests := []struct {
		agent        string
		sessionID    string
		extraArgs    []string
		expectedArgs []string
	}{
		{
			agent:        "claude",
			sessionID:    "sess-claude-uuid-123",
			extraArgs:    nil,
			expectedArgs: []string{"--resume", "sess-claude-uuid-123"},
		},
		{
			agent:        "claude",
			sessionID:    "sess-claude-uuid-456",
			extraArgs:    []string{"--dangerously-skip-permissions", "--verbose"},
			expectedArgs: []string{"--resume", "sess-claude-uuid-456", "--dangerously-skip-permissions", "--verbose"},
		},
		{
			agent:        "codex",
			sessionID:    "sess-codex-id-789",
			extraArgs:    nil,
			expectedArgs: []string{"resume", "sess-codex-id-789"},
		},
		{
			agent:        "codex",
			sessionID:    "sess-codex-id-012",
			extraArgs:    []string{"--model", "gpt-4o"},
			expectedArgs: []string{"resume", "sess-codex-id-012", "--model", "gpt-4o"},
		},
		{
			agent:        "agy",
			sessionID:    "sess-agy-id-345",
			extraArgs:    nil,
			expectedArgs: []string{"--conversation=sess-agy-id-345"},
		},
		{
			agent:        "agy",
			sessionID:    "sess-agy-id-678",
			extraArgs:    []string{"--telemetry=off"},
			expectedArgs: []string{"--conversation=sess-agy-id-678", "--telemetry=off"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.agent+"_"+tc.sessionID, func(t *testing.T) {
			adapter, err := reg.Get(tc.agent)
			if err != nil {
				t.Fatalf("failed to get adapter %q: %v", tc.agent, err)
			}
			resumer, ok := adapter.(agents.SessionResumer)
			if !ok {
				t.Fatalf("adapter %q does not implement agents.SessionResumer", tc.agent)
			}
			args := resumer.ResumeArgs(tc.sessionID, tc.extraArgs)
			if len(args) != len(tc.expectedArgs) {
				t.Fatalf("expected args len %d, got %d: %v", len(tc.expectedArgs), len(args), args)
			}
			for i := range args {
				if args[i] != tc.expectedArgs[i] {
					t.Errorf("arg[%d] mismatch: expected %q, got %q (all args: %v)", i, tc.expectedArgs[i], args[i], args)
				}
			}
		})
	}
}

func TestParameterized_MultiAgent_ResolveCwd(t *testing.T) {
	mgr := setupParameterizedSessionManager()
	ctx := context.Background()

	testPaths := []string{
		"path with spaces/nested workspace",
		"deeply/nested/sub/directory/project",
		"unicode_🚀_项目_ünicode/workspace",
		"spaces and unicode 🎯/dir with spaces/子目录",
	}

	sessionAgents := []string{"claude", "codex", "agy"}

	for _, agent := range sessionAgents {
		t.Run(agent, func(t *testing.T) {
			for _, relPath := range testPaths {
				t.Run(relPath, func(t *testing.T) {
					tempDir := t.TempDir()
					absPath := filepath.Join(tempDir, relPath)
					if err := os.MkdirAll(absPath, 0755); err != nil {
						t.Fatalf("failed to create test directory %q: %v", absPath, err)
					}

					sess := &session.Session{
						ID:      "test-sess-" + agent,
						Agent:   agent,
						Profile: "work",
						Cwd:     absPath,
					}

					resolved, err := mgr.ResolveCwd(ctx, sess)
					if err != nil {
						t.Fatalf("ResolveCwd failed for agent %s: %v", agent, err)
					}
					if canonicalDir(t, resolved) != canonicalDir(t, absPath) {
						t.Errorf("ResolveCwd mismatch for agent %s:\nexpected: %s\ngot:      %s",
							agent, canonicalDir(t, absPath), canonicalDir(t, resolved))
					}
				})
			}
		})
	}

	t.Run("Claude_SlugResolution_SpacesAndUnicode", func(t *testing.T) {
		tempDir := t.TempDir()
		projectDir := filepath.Join(tempDir, "claude-nested-spaces 🚀", "sub-project")
		if err := os.MkdirAll(projectDir, 0755); err != nil {
			t.Fatalf("failed to create directory: %v", err)
		}

		slug := claudesess.PathToSlug(projectDir)
		recovered := claudesess.SlugToPath(slug)
		if canonicalDir(t, recovered) != canonicalDir(t, projectDir) {
			t.Errorf("SlugToPath failed for path with spaces & unicode:\nexpected: %s\ngot:      %s",
				canonicalDir(t, projectDir), canonicalDir(t, recovered))
		}

		fakeProjectsDir := filepath.Join(tempDir, "profiles", "work", ".claude", "projects", slug)
		if err := os.MkdirAll(fakeProjectsDir, 0755); err != nil {
			t.Fatalf("failed to create fake projects dir: %v", err)
		}
		fakeSessionFile := filepath.Join(fakeProjectsDir, "test-sess.jsonl")
		_ = os.WriteFile(fakeSessionFile, []byte(`{"type":"user","text":"hello"}`), 0600)

		sess := &session.Session{
			ID:          "test-sess",
			Agent:       "claude",
			Profile:     "work",
			StoragePath: fakeSessionFile,
			Cwd:         "",
		}
		resolved, err := mgr.ResolveCwd(ctx, sess)
		if err != nil {
			t.Fatalf("ResolveCwd via StoragePath failed: %v", err)
		}
		if canonicalDir(t, resolved) != canonicalDir(t, projectDir) {
			t.Errorf("expected resolved Cwd from StoragePath to match %s, got %s",
				canonicalDir(t, projectDir), canonicalDir(t, resolved))
		}
	})
}

func TestParameterized_MultiAgent_WorkingDirPropagation(t *testing.T) {
	sessionAgents := []string{"claude", "codex", "agy"}

	for _, agent := range sessionAgents {
		t.Run(agent, func(t *testing.T) {
			tempDir := t.TempDir()
			t.Setenv("AIM_HOME", tempDir)
			t.Setenv("AIM_REAL_HOME", tempDir)
			t.Setenv("HOME", tempDir)

			targetWorkDir := filepath.Join(tempDir, "workspace with spaces and 🚀 unicode", agent+"_dir")
			if err := os.MkdirAll(targetWorkDir, 0755); err != nil {
				t.Fatalf("failed to create target workspace: %v", err)
			}

			dumpFile := filepath.Join(tempDir, "cwd_dump_"+agent+".txt")
			fakeBinDir := filepath.Join(tempDir, "bin_"+agent)
			_ = os.MkdirAll(fakeBinDir, 0755)

			reg := setupParameterizedRegistry()
			adapter, err := reg.Get(agent)
			if err != nil {
				t.Fatalf("failed to get adapter %s: %v", agent, err)
			}

			binName := adapter.BinaryName()
			fakeBin := filepath.Join(fakeBinDir, binName)
			fakeScript := fmt.Sprintf("#!/bin/sh\npwd > %q\nexit 0\n", dumpFile)
			if err := os.WriteFile(fakeBin, []byte(fakeScript), 0755); err != nil {
				t.Fatalf("failed to write fake binary: %v", err)
			}
			t.Setenv("PATH", fakeBinDir+":"+os.Getenv("PATH"))

			pm := profile.NewProfileManager(tempDir)
			pDir, err := pm.EnsureProfile("default")
			if err != nil {
				t.Fatalf("failed to ensure profile: %v", err)
			}

			var buf bytes.Buffer
			cmd := newRootCmd(reg, pm)
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			sess := &session.Session{
				ID:           "test-cwd-session-" + agent,
				ShortID:      "test-cwd-" + agent,
				Title:        "CWD Test Session " + agent,
				Agent:        agent,
				Profile:      "default",
				Cwd:          targetWorkDir,
				LastActiveAt: time.Now(),
				Status:       session.StatusIdle,
			}

			mgr := setupParameterizedSessionManager()
			ctx := WithSessionManager(context.Background(), mgr)
			cmd.SetContext(ctx)

			err = executeExactResume(cmd, reg, pm, mgr, agent, "default", pDir, sess, false, nil)
			if err != nil {
				t.Fatalf("executeExactResume failed for %s: %v", agent, err)
			}

			content, err := os.ReadFile(dumpFile)
			if err != nil {
				t.Fatalf("failed to read cwd dump file for %s: %v", agent, err)
			}
			gotCwd := strings.TrimSpace(string(content))
			if canonicalDir(t, gotCwd) != canonicalDir(t, targetWorkDir) {
				t.Errorf("working dir mismatch for agent %s:\nexpected: %s\ngot:      %s",
					agent, canonicalDir(t, targetWorkDir), canonicalDir(t, gotCwd))
			}
		})
	}
}

func TestParameterized_MultiAgent_CachedUsageProvider(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := setupParameterizedRegistry()
	usageCache := usage.NewCacheStore(tempDir, usage.DefaultTTL)
	ctx := context.Background()

	t.Run("claude_CachedUsageProvider", func(t *testing.T) {
		adapter, err := reg.Get("claude")
		if err != nil {
			t.Fatalf("failed to get claude adapter: %v", err)
		}
		cachedProv, ok := adapter.(agents.CachedUsageProvider)
		if !ok {
			t.Fatalf("claude adapter must implement agents.CachedUsageProvider")
		}

		profileDir := filepath.Join(tempDir, "profiles", "claude-work")
		claudeSubdir := filepath.Join(profileDir, ".claude")
		if err := os.MkdirAll(claudeSubdir, 0700); err != nil {
			t.Fatalf("failed to create claude profile dir: %v", err)
		}

		_ = os.WriteFile(filepath.Join(claudeSubdir, ".credentials.json"), []byte(`{"sessionKey":"sk-ant-test"}`), 0600)

		fixturePath := filepath.Join(getGitRepoRoot(), "internal", "agents", "claude", "testdata", "claude_cached_usage.json")
		fixtureData, err := os.ReadFile(fixturePath)
		if err != nil {
			t.Fatalf("failed to read golden fixture: %v", err)
		}
		if err := os.WriteFile(filepath.Join(profileDir, ".claude.json"), fixtureData, 0600); err != nil {
			t.Fatalf("failed to write .claude.json: %v", err)
		}

		target := usage.TargetProfile{
			Agent:            "claude",
			Profile:          "claude-work",
			ProfileDir:       profileDir,
			GetCachedUsageFn: cachedProv.GetCachedUsage,
		}

		rep := usage.InspectProfile(ctx, target, usageCache)
		if !rep.FromCache {
			t.Errorf("expected rep.FromCache to be true for claude cached usage")
		}
		if rep.Status != usage.StatusExhausted && rep.Status != usage.StatusOK {
			t.Errorf("expected StatusExhausted or StatusOK for valid claude cached usage, got: %s", rep.Status)
		}
		if len(rep.Windows) == 0 {
			t.Fatalf("expected quota windows to be parsed from claude_cached_usage.json")
		}

		var fiveHourWindow *usage.LimitWindow
		for i := range rep.Windows {
			if strings.Contains(rep.Windows[i].Name, "5-Hour") || strings.Contains(rep.Windows[i].Name, "Session") {
				fiveHourWindow = &rep.Windows[i]
				break
			}
		}
		if fiveHourWindow == nil {
			t.Errorf("expected 5-Hour Session window in parsed report, got: %+v", rep.Windows)
		} else if fiveHourWindow.RemainingPct != 97 {
			t.Errorf("expected 5-Hour Session RemainingPct 97, got %v", fiveHourWindow.RemainingPct)
		}
	})

	sessionAgents := []struct {
		agent      string
		credRelDir string
		credFile   string
		credData   string
	}{
		{
			agent:      "claude",
			credRelDir: ".claude",
			credFile:   ".credentials.json",
			credData:   `{"sessionKey":"sk-ant-test"}`,
		},
		{
			agent:      "codex",
			credRelDir: ".codex",
			credFile:   "auth.json",
			credData:   `{"tokens":{"access_token":"valid"}}`,
		},
		{
			agent:      "agy",
			credRelDir: filepath.Join(".gemini", "antigravity-cli"),
			credFile:   "antigravity-oauth-token",
			credData:   `{"access_token":"valid"}`,
		},
	}

	for _, tc := range sessionAgents {
		t.Run(tc.agent+"_InspectProfile_Lifecycle", func(t *testing.T) {
			adapter, err := reg.Get(tc.agent)
			if err != nil {
				t.Fatalf("failed to get adapter %s: %v", tc.agent, err)
			}

			pDir := filepath.Join(tempDir, "profiles", tc.agent+"-idle")
			credDir := filepath.Join(pDir, tc.credRelDir)
			if err := os.MkdirAll(credDir, 0700); err != nil {
				t.Fatalf("failed to create cred dir: %v", err)
			}
			if err := os.WriteFile(filepath.Join(credDir, tc.credFile), []byte(tc.credData), 0600); err != nil {
				t.Fatalf("failed to write credentials: %v", err)
			}

			target := usage.TargetProfile{
				Agent:      tc.agent,
				Profile:    tc.agent + "-idle",
				ProfileDir: pDir,
			}
			if cp, ok := adapter.(agents.CachedUsageProvider); ok {
				target.GetCachedUsageFn = cp.GetCachedUsage
			}

			rep := usage.InspectProfile(ctx, target, usageCache)
			if rep.Status != usage.StatusIdle && rep.Status != usage.StatusOK {
				t.Errorf("expected StatusIdle or StatusOK for %s, got: %s (error: %s)", tc.agent, rep.Status, rep.Error)
			}
		})
	}
}
