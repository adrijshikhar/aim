package main

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/agents/agy"
	"github.com/aim-cli/aim/internal/agents/claude"
	"github.com/aim-cli/aim/internal/agents/codex"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/session"
	"github.com/spf13/cobra"
)

type testCompProvider struct {
	agent    string
	sessions []session.Session
}

func (p *testCompProvider) Agent() string { return p.agent }

func (p *testCompProvider) ListSessions(ctx context.Context, profileDir string, isHost bool) ([]session.Session, error) {
	if isHost {
		var res []session.Session
		for _, s := range p.sessions {
			if s.IsHost || s.Profile == "<host>" {
				res = append(res, s)
			}
		}
		return res, nil
	}
	profName := filepath.Base(profileDir)
	var res []session.Session
	for _, s := range p.sessions {
		if s.Profile == profName {
			res = append(res, s)
		}
	}
	return res, nil
}

func (p *testCompProvider) GetSession(ctx context.Context, idOrPrefix string, profileDir string, isHost bool) (*session.Session, error) {
	return nil, nil
}

func (p *testCompProvider) Hydrate(ctx context.Context, srcSession *session.Session, destProfileDir string, fork bool) (string, error) {
	return srcSession.ID, nil
}

func setupCompletionTestEnv(t *testing.T) (*agents.Registry, *profile.ProfileManager) {
	t.Helper()
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := agents.NewRegistry()
	reg.Register(codex.NewAdapter())
	reg.Register(agy.NewAdapter())
	reg.Register(claude.NewAdapter())

	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("work")
	_, _ = pm.EnsureProfile("office")
	_, _ = pm.EnsureProfile("default")

	cfg := &config.Config{
		Profiles: map[string]config.ProfileConfig{
			"work":    {Agents: []string{"codex", "agy"}},
			"office":  {Agents: []string{"codex"}},
			"default": {Agents: []string{"codex"}},
		},
	}
	_ = config.SaveConfig(cfg)

	now := time.Now()
	codexSessions := []session.Session{
		{
			ID:           "01a09eb7-1111-2222-3333-444444444444",
			ShortID:      "01a09eb7",
			Title:        "Fix login bug",
			Agent:        "codex",
			Profile:      "work",
			LastActiveAt: now.Add(-10 * time.Minute),
			Status:       session.StatusIdle,
		},
		{
			ID:           "02b18fc6-2222-3333-4444-555555555555",
			ShortID:      "02b18fc6",
			Title:        "Database migration bug",
			Agent:        "codex",
			Profile:      "office",
			LastActiveAt: now.Add(-20 * time.Minute),
			Status:       session.StatusIdle,
		},
		{
			ID:           "03c27eb5-3333-4444-5555-666666666666",
			ShortID:      "03c27eb5",
			Title:        "", // empty title to test format "[Profile]"
			Agent:        "codex",
			Profile:      "default",
			LastActiveAt: now.Add(-30 * time.Minute),
			Status:       session.StatusIdle,
		},
	}

	agySessions := []session.Session{
		{
			ID:           "99f99f99-9999-9999-9999-999999999999",
			ShortID:      "99f99f99",
			Title:        "Agy session",
			Agent:        "agy",
			Profile:      "work",
			LastActiveAt: now.Add(-5 * time.Minute),
			Status:       session.StatusIdle,
		},
	}

	oldMgr := defaultSessionManager
	t.Cleanup(func() {
		defaultSessionManager = oldMgr
	})

	mockMgr := session.NewManager()
	mockMgr.RegisterProvider(&testCompProvider{
		agent:    "codex",
		sessions: codexSessions,
	})
	mockMgr.RegisterProvider(&testCompProvider{
		agent:    "agy",
		sessions: agySessions,
	})
	defaultSessionManager = func() *session.Manager {
		return mockMgr
	}

	return reg, pm
}

func TestCompleteSessionIDs(t *testing.T) {
	expectedCodexCompletions := []string{
		"01a09eb7\tFix login bug [work]",
		"02b18fc6\tDatabase migration bug [office]",
		"03c27eb5\t[default]",
	}

	t.Run("ResumeCmd_ValidArgsFunction", func(t *testing.T) {
		reg, pm := setupCompletionTestEnv(t)
		cmd := newResumeCmd(reg, pm)

		comps, directive := cmd.ValidArgsFunction(cmd, []string{"codex", "work"}, "")
		if directive != cobra.ShellCompDirectiveNoFileComp {
			t.Errorf("expected directive ShellCompDirectiveNoFileComp, got %v", directive)
		}
		if len(comps) != len(expectedCodexCompletions) {
			t.Fatalf("expected %d completions, got %d: %v", len(expectedCodexCompletions), len(comps), comps)
		}
		for i, expected := range expectedCodexCompletions {
			if comps[i] != expected {
				t.Errorf("completion[%d] = %q, want %q", i, comps[i], expected)
			}
		}
	})

	t.Run("ResumeCmd_PrefixFiltering", func(t *testing.T) {
		reg, pm := setupCompletionTestEnv(t)
		cmd := newResumeCmd(reg, pm)

		// Prefix "01" should only match 01a09eb7
		comps, directive := cmd.ValidArgsFunction(cmd, []string{"codex", "work"}, "01")
		if directive != cobra.ShellCompDirectiveNoFileComp {
			t.Errorf("expected directive ShellCompDirectiveNoFileComp, got %v", directive)
		}
		if len(comps) != 1 || comps[0] != "01a09eb7\tFix login bug [work]" {
			t.Fatalf("expected only 01a09eb7 completion for prefix '01', got: %v", comps)
		}

		// Prefix "02b" should only match 02b18fc6
		comps, _ = cmd.ValidArgsFunction(cmd, []string{"codex", "work"}, "02b")
		if len(comps) != 1 || comps[0] != "02b18fc6\tDatabase migration bug [office]" {
			t.Fatalf("expected only 02b18fc6 completion for prefix '02b', got: %v", comps)
		}

		// Non-matching prefix
		comps, _ = cmd.ValidArgsFunction(cmd, []string{"codex", "work"}, "xyz")
		if len(comps) != 0 {
			t.Fatalf("expected 0 completions for non-matching prefix 'xyz', got: %v", comps)
		}
	})

	t.Run("ImportCmd_ValidArgsFunction", func(t *testing.T) {
		reg, pm := setupCompletionTestEnv(t)
		cmd := newSessionsImportCmd(reg, pm)

		comps, directive := cmd.ValidArgsFunction(cmd, []string{"codex", "work"}, "")
		if directive != cobra.ShellCompDirectiveNoFileComp {
			t.Errorf("expected directive ShellCompDirectiveNoFileComp, got %v", directive)
		}
		if len(comps) != len(expectedCodexCompletions) {
			t.Fatalf("expected %d completions, got %d: %v", len(expectedCodexCompletions), len(comps), comps)
		}
		for i, expected := range expectedCodexCompletions {
			if comps[i] != expected {
				t.Errorf("completion[%d] = %q, want %q", i, comps[i], expected)
			}
		}
	})

	t.Run("CompleteAgentProfileAndSession_Arity", func(t *testing.T) {
		reg, pm := setupCompletionTestEnv(t)

		// 0 args -> completes agents
		comps, directive := completeAgentProfileAndSession(reg, pm, []string{}, "")
		if directive != cobra.ShellCompDirectiveNoFileComp {
			t.Errorf("expected NoFileComp directive, got %v", directive)
		}
		foundCodex := false
		foundClaude := false
		for _, c := range comps {
			if strings.HasPrefix(c, "codex") {
				foundCodex = true
			}
			if strings.HasPrefix(c, "claude") {
				foundClaude = true
			}
		}
		if !foundCodex {
			t.Errorf("expected agent completions to include codex, got: %v", comps)
		}
		if !foundClaude {
			t.Errorf("expected agent completions to include claude, got: %v", comps)
		}

		// 1 arg -> completes profiles
		comps, directive = completeAgentProfileAndSession(reg, pm, []string{"codex"}, "")
		if directive != cobra.ShellCompDirectiveNoFileComp {
			t.Errorf("expected NoFileComp directive, got %v", directive)
		}
		if len(comps) == 0 {
			t.Errorf("expected profile completions for codex, got empty")
		}

		// 2 args -> completes sessions
		comps, directive = completeAgentProfileAndSession(reg, pm, []string{"codex", "work"}, "")
		if directive != cobra.ShellCompDirectiveNoFileComp {
			t.Errorf("expected NoFileComp directive, got %v", directive)
		}
		if len(comps) != 3 {
			t.Errorf("expected 3 session completions for codex, got %d: %v", len(comps), comps)
		}

		// 3 args -> returns nil
		comps, directive = completeAgentProfileAndSession(reg, pm, []string{"codex", "work", "01a09eb7"}, "")
		if directive != cobra.ShellCompDirectiveNoFileComp {
			t.Errorf("expected NoFileComp directive, got %v", directive)
		}
		if len(comps) != 0 {
			t.Errorf("expected 0 completions for 3+ args, got: %v", comps)
		}
	})

	t.Run("AgentIsolation", func(t *testing.T) {
		reg, pm := setupCompletionTestEnv(t)

		// Completing for "agy" should return only agy's session
		comps, _ := completeAgentProfileAndSession(reg, pm, []string{"agy", "work"}, "")
		if len(comps) != 1 || comps[0] != "99f99f99\tAgy session [work]" {
			t.Fatalf("expected only agy session '99f99f99\tAgy session [work]', got: %v", comps)
		}
	})

	t.Run("EdgeCases", func(t *testing.T) {
		reg, pm := setupCompletionTestEnv(t)

		// Unknown agent
		comps, directive := completeAgentProfileAndSession(reg, pm, []string{"nonexistent-agent", "work"}, "")
		if directive != cobra.ShellCompDirectiveNoFileComp {
			t.Errorf("expected NoFileComp directive, got %v", directive)
		}
		if len(comps) != 0 {
			t.Errorf("expected 0 completions for nonexistent agent, got: %v", comps)
		}

		// Nil registry
		comps, directive = completeAgentProfileAndSession(nil, pm, []string{"codex", "work"}, "")
		if directive != cobra.ShellCompDirectiveNoFileComp {
			t.Errorf("expected NoFileComp directive, got %v", directive)
		}
		if len(comps) != 3 {
			t.Errorf("expected 3 completions even with nil registry (codex provider still registered), got: %v", comps)
		}

		// Nil profile manager
		comps, directive = completeAgentProfileAndSession(reg, nil, []string{"codex", "work"}, "")
		if directive != cobra.ShellCompDirectiveNoFileComp {
			t.Errorf("expected NoFileComp directive, got %v", directive)
		}
		if len(comps) != 3 {
			t.Errorf("expected 3 completions even with nil pm, got: %v", comps)
		}
	})
}

func TestCompleteAgents_IncludesClaude(t *testing.T) {
	comps := completeAgents(nil, "")
	foundClaude := false
	for _, c := range comps {
		if strings.HasPrefix(c, "claude") {
			foundClaude = true
			break
		}
	}
	if !foundClaude {
		t.Errorf("expected completeAgents to include claude, got: %v", comps)
	}
}
