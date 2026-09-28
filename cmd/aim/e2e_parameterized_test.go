package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/agents/agy"
	"github.com/aim-cli/aim/internal/agents/claude"
	"github.com/aim-cli/aim/internal/agents/codex"
	"github.com/aim-cli/aim/internal/agents/gemini"
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
}

var allAdapters = []adapterTestCase{
	{agent: "agy", hasSessions: true, supportsFork: false},
	{agent: "codex", hasSessions: true, supportsFork: true},
	{agent: "claude", hasSessions: true, supportsFork: true},
	{agent: "gemini", hasSessions: false, supportsFork: false},
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
