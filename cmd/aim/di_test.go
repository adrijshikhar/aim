package main

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/session"
	"github.com/spf13/cobra"
)

type mockSessionProvider struct {
	agentName string
	sessions  []session.Session
}

func (m *mockSessionProvider) Agent() string {
	return m.agentName
}

func (m *mockSessionProvider) ListSessions(ctx context.Context, profileDir string, isHost bool) ([]session.Session, error) {
	return m.sessions, nil
}

func (m *mockSessionProvider) GetSession(ctx context.Context, idOrPrefix string, profileDir string, isHost bool) (*session.Session, error) {
	for _, s := range m.sessions {
		if s.ID == idOrPrefix || s.ShortID == idOrPrefix {
			return &s, nil
		}
	}
	return nil, nil
}

func (m *mockSessionProvider) Hydrate(ctx context.Context, srcSession *session.Session, destProfileDir string, fork bool) (string, error) {
	return srcSession.ID, nil
}

func (m *mockSessionProvider) ResolveCwd(ctx context.Context, s *session.Session) (string, error) {
	return s.Cwd, nil
}

func (m *mockSessionProvider) ResolveSummary(ctx context.Context, s *session.Session) (session.SessionSummary, error) {
	return session.SessionSummary{Raw: s.Summary}, nil
}

func TestDI_SessionsCmd_StructInjection(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	pm := profile.NewProfileManager(tmpDir)
	reg := agents.NewRegistry()

	mockMgr := session.NewManager()
	mockMgr.RegisterProvider(&mockSessionProvider{
		agentName: "custom-agent",
		sessions: []session.Session{
			{
				ID:           "sess-123456",
				ShortID:      "sess-1",
				Agent:        "custom-agent",
				Profile:      "p1",
				Title:        "Parallel DI Test Session",
				LastActiveAt: time.Now(),
			},
		},
	})

	var out bytes.Buffer
	sessionsCmd := NewSessionsCmd(reg, pm).
		WithSessionManager(mockMgr).
		WithIsTerminal(func(cmd *cobra.Command) bool {
			return false
		})

	cmd := sessionsCmd.Command()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--plain"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "sess-1") || !strings.Contains(result, "Parallel DI Test Session") {
		t.Errorf("expected session output, got: %q", result)
	}
}

func TestDI_SessionsCmd_TUIRunnerInjection(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	pm := profile.NewProfileManager(tmpDir)
	reg := agents.NewRegistry()

	var tuiCalledWithAgent string
	var tuiCalledWithProfile string

	sessionsCmd := NewSessionsCmd(reg, pm).
		WithIsTerminal(func(cmd *cobra.Command) bool {
			return true
		}).
		WithTUIRunner(func(r *agents.Registry, p *profile.ProfileManager, initialAgent, profileFilter string, activeOnly bool) int {
			tuiCalledWithAgent = initialAgent
			tuiCalledWithProfile = profileFilter
			return 0
		})

	cmd := sessionsCmd.Command()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"agy", "-p", "dev"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tuiCalledWithAgent != "agy" {
		t.Errorf("expected agent agy, got %q", tuiCalledWithAgent)
	}
	if tuiCalledWithProfile != "dev" {
		t.Errorf("expected profile dev, got %q", tuiCalledWithProfile)
	}
}

func TestDI_Context_SessionManager(t *testing.T) {
	t.Parallel()

	mockMgr := session.NewManager()
	mockMgr.RegisterProvider(&mockSessionProvider{
		agentName: "ctx-agent",
		sessions: []session.Session{
			{
				ID:      "ctx-sess-001",
				ShortID: "ctx-001",
				Agent:   "ctx-agent",
				Profile: "default",
				Title:   "Context Injected Session",
			},
		},
	})

	ctx := WithSessionManager(context.Background(), mockMgr)
	retrieved := getSessionManager(ctx)
	if retrieved != mockMgr {
		t.Fatalf("expected retrieved manager to match mockMgr")
	}

	sessions, err := retrieved.ListSessions(ctx, "ctx-agent", "", false)
	if err != nil {
		t.Fatalf("unexpected list error: %v", err)
	}
	if len(sessions) != 1 || sessions[0].ShortID != "ctx-001" {
		t.Fatalf("expected 1 session with shortID ctx-001, got: %v", sessions)
	}
}

func TestDI_Context_InteractiveCheck(t *testing.T) {
	t.Parallel()

	called := false
	customCheck := func(r io.Reader) bool {
		called = true
		return true
	}

	ctx := WithInteractiveCheck(context.Background(), customCheck)
	cmd := &cobra.Command{}
	cmd.SetContext(ctx)

	var in bytes.Buffer
	res := checkInteractive(cmd, &in)
	if !res {
		t.Errorf("expected checkInteractive to return true")
	}
	if !called {
		t.Errorf("expected customCheck to have been called")
	}
}

func TestDI_PromptHelper_StructInjection(t *testing.T) {
	t.Parallel()

	helper := NewPromptHelper(func(r io.Reader) bool {
		return false
	})

	tmpDir := t.TempDir()
	pm := profile.NewProfileManager(tmpDir)

	cmd := &cobra.Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	// Non-interactive should return error when profile does not exist
	ok, err := helper.ConfirmProfileExists(cmd, pm, "nonexistent-profile", "test action", false)
	if ok || err == nil {
		t.Errorf("expected non-interactive creation to fail, got ok=%v, err=%v", ok, err)
	}
}

func TestDI_Context_TUIRunners(t *testing.T) {
	t.Parallel()

	tuiCalled := false
	sessionsTuiCalled := false

	ctx := context.Background()
	ctx = WithTUIRunner(ctx, func(reg *agents.Registry, pm *profile.ProfileManager) int {
		tuiCalled = true
		return 42
	})
	ctx = WithTUISessionsRunner(ctx, func(reg *agents.Registry, pm *profile.ProfileManager, agent, profile string, active bool) int {
		sessionsTuiCalled = true
		return 84
	})

	code1 := runTUIWithContext(ctx, nil, nil)
	if code1 != 42 || !tuiCalled {
		t.Errorf("expected code 42 and tuiCalled true, got code=%d, called=%v", code1, tuiCalled)
	}

	code2 := runTUISessionsWithContext(ctx, nil, nil, "agy", "work", true)
	if code2 != 84 || !sessionsTuiCalled {
		t.Errorf("expected code 84 and sessionsTuiCalled true, got code=%d, called=%v", code2, sessionsTuiCalled)
	}
}
