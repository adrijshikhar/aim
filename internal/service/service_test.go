package service_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/service"
	"github.com/aim-cli/aim/internal/session"
	"github.com/aim-cli/aim/internal/usage"
)

// mockAdapter implements agents.AgentAdapter for testing.
type mockAdapter struct {
	name        string
	aliases     []string
	hasCredsMap map[string]bool
}

func (m *mockAdapter) Name() string        { return m.name }
func (m *mockAdapter) DisplayName() string { return strings.ToUpper(m.name) }
func (m *mockAdapter) Aliases() []string   { return m.aliases }
func (m *mockAdapter) BinaryName() string  { return m.name }
func (m *mockAdapter) HasCredentials(profileDir string) bool {
	if m.hasCredsMap != nil {
		profName := filepath.Base(profileDir)
		return m.hasCredsMap[profName]
	}
	return false
}
func (m *mockAdapter) PrepareEnv(ctx context.Context, profileName, profileDir string) (agents.LaunchEnv, error) {
	return agents.LaunchEnv{BinaryPath: m.name, WorkingDir: profileDir}, nil
}

// mockSessionProvider implements session.SessionProvider for testing.
type mockSessionProvider struct {
	agent    string
	sessions []session.Session
}

func (m *mockSessionProvider) Agent() string { return m.agent }
func (m *mockSessionProvider) ListSessions(ctx context.Context, profileDir string, isHost bool) ([]session.Session, error) {
	profName := filepath.Base(profileDir)
	var res []session.Session
	for _, s := range m.sessions {
		if isHost && s.IsHost {
			res = append(res, s)
		} else if !isHost && !s.IsHost && (s.Profile == "" || s.Profile == profName) {
			res = append(res, s)
		}
	}
	return res, nil
}
func (m *mockSessionProvider) GetSession(ctx context.Context, idOrPrefix, profileDir string, isHost bool) (*session.Session, error) {
	profName := filepath.Base(profileDir)
	for _, s := range m.sessions {
		if isHost && !s.IsHost {
			continue
		}
		if !isHost && (s.IsHost || (s.Profile != "" && s.Profile != profName)) {
			continue
		}
		if s.ID == idOrPrefix || strings.HasPrefix(s.ID, idOrPrefix) {
			copyS := s
			return &copyS, nil
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
	return session.SessionSummary{Goal: s.Goal, RecentActivity: s.Summary}, nil
}

// mockLauncher implements service.LauncherService for testing.
type mockLauncher struct {
	launchedCmds []string
	available    []string
	err          error
}

func (m *mockLauncher) LaunchTerminal(ctx context.Context, cmdStr string) error {
	if m.err != nil {
		return m.err
	}
	m.launchedCmds = append(m.launchedCmds, cmdStr)
	return nil
}
func (m *mockLauncher) AvailableTerminals() []string {
	return m.available
}

func setupTestEnv(t *testing.T) (string, *profile.ProfileManager, *agents.Registry, *usage.CacheStore) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "aim-service-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tempDir) })

	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_CONFIG_DIR", tempDir)
	t.Setenv("AIM_DATA_DIR", tempDir)

	pm := profile.NewProfileManager(tempDir)
	reg := agents.NewRegistry()

	claudeAd := &mockAdapter{
		name:    "claude",
		aliases: []string{"claude-code"},
		hasCredsMap: map[string]bool{
			"alpha": true,
		},
	}
	agyAd := &mockAdapter{
		name:    "agy",
		aliases: []string{"antigravity"},
		hasCredsMap: map[string]bool{
			"beta": true,
		},
	}
	reg.Register(claudeAd)
	reg.Register(agyAd)

	cache := usage.NewCacheStore(tempDir, usage.DefaultTTL)

	return tempDir, pm, reg, cache
}

func TestProfileService_ListProfiles(t *testing.T) {
	ctx := context.Background()
	_, pm, reg, cache := setupTestEnv(t)
	svc := service.NewProfileService(pm, reg, cache)

	// Initially empty
	profs, err := svc.ListProfiles(ctx, "claude")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(profs) != 0 {
		t.Fatalf("expected 0 profiles, got %d", len(profs))
	}

	// Create profiles
	_, err = pm.EnsureProfile("alpha")
	if err != nil {
		t.Fatalf("failed to ensure alpha: %v", err)
	}
	_, err = pm.EnsureProfile("beta")
	if err != nil {
		t.Fatalf("failed to ensure beta: %v", err)
	}

	cfg, _ := config.LoadConfig()
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}
	cfg.AddProfileAgent("alpha", "claude")
	cfg.AddProfileAgent("beta", "agy")
	_ = config.SaveConfig(cfg)

	// Populate cache for alpha
	_ = cache.Put(usage.Report{
		Agent:   "claude",
		Profile: "alpha",
		Status:  usage.StatusOK,
		Summary: "5h: 80%",
		Windows: []usage.LimitWindow{
			{Category: "claude", Name: "5-hour limit", RemainingPct: 80},
		},
	})

	// Query for claude
	claudeProfs, err := svc.ListProfiles(ctx, "claude")
	if err != nil {
		t.Fatalf("failed to list profiles: %v", err)
	}
	if len(claudeProfs) != 1 {
		t.Fatalf("expected 1 profile for claude, got %d", len(claudeProfs))
	}
	if claudeProfs[0].Name != "alpha" {
		t.Errorf("expected profile name alpha, got %s", claudeProfs[0].Name)
	}
	if claudeProfs[0].Agent != "claude" {
		t.Errorf("expected agent claude, got %s", claudeProfs[0].Agent)
	}
	if !claudeProfs[0].HasCredentials {
		t.Errorf("expected HasCredentials to be true for alpha")
	}
	if claudeProfs[0].Quota == nil {
		t.Fatalf("expected quota info for alpha, got nil")
	}
	if claudeProfs[0].Quota.BottleneckPct != 80 {
		t.Errorf("expected BottleneckPct 80, got %d", claudeProfs[0].Quota.BottleneckPct)
	}
	if claudeProfs[0].Quota.IsExhausted {
		t.Errorf("expected IsExhausted to be false")
	}

	// Query with alias
	agyProfs, err := svc.ListProfiles(ctx, "antigravity")
	if err != nil {
		t.Fatalf("failed to list profiles with alias: %v", err)
	}
	if len(agyProfs) != 1 || agyProfs[0].Name != "beta" {
		t.Fatalf("expected beta profile for agy alias, got %v", agyProfs)
	}

	// Query all
	allProfs, err := svc.ListProfiles(ctx, "")
	if err != nil {
		t.Fatalf("failed to list all profiles: %v", err)
	}
	if len(allProfs) != 2 {
		t.Fatalf("expected 2 profiles in total, got %d", len(allProfs))
	}
}

func TestProfileService_CreateProfile(t *testing.T) {
	ctx := context.Background()
	_, pm, reg, cache := setupTestEnv(t)
	svc := service.NewProfileService(pm, reg, cache)

	// Validation errors
	if _, err := svc.CreateProfile(ctx, service.CreateProfileRequest{Name: ""}); err == nil {
		t.Errorf("expected error for empty name, got nil")
	}
	if _, err := svc.CreateProfile(ctx, service.CreateProfileRequest{Name: "../bad"}); err == nil {
		t.Errorf("expected error for path traversal, got nil")
	}

	// Create valid profile
	dto, err := svc.CreateProfile(ctx, service.CreateProfileRequest{
		Agent: "claude",
		Name:  "work",
		Email: "dev@example.com",
	})
	if err != nil {
		t.Fatalf("failed to create profile: %v", err)
	}
	if dto.Name != "work" || dto.Agent != "claude" {
		t.Errorf("unexpected profile DTO: %+v", dto)
	}
	if dto.Account == nil || dto.Account.Email != "dev@example.com" {
		t.Errorf("expected email dev@example.com, got %+v", dto.Account)
	}

	// Duplicate error
	if _, err := svc.CreateProfile(ctx, service.CreateProfileRequest{Name: "work"}); err == nil {
		t.Errorf("expected error creating existing profile, got nil")
	}

	// Clone from existing
	cloned, err := svc.CreateProfile(ctx, service.CreateProfileRequest{
		Agent:     "claude",
		Name:      "work-clone",
		CloneFrom: "work",
	})
	if err != nil {
		t.Fatalf("failed to clone profile: %v", err)
	}
	if cloned.Name != "work-clone" {
		t.Errorf("expected name work-clone, got %s", cloned.Name)
	}

	// Clone from non-existent
	if _, err := svc.CreateProfile(ctx, service.CreateProfileRequest{
		Agent:     "claude",
		Name:      "bad-clone",
		CloneFrom: "does-not-exist",
	}); err == nil {
		t.Errorf("expected error cloning from non-existent profile, got nil")
	}
}

func TestProfileService_RemoveProfile(t *testing.T) {
	ctx := context.Background()
	_, pm, reg, cache := setupTestEnv(t)
	svc := service.NewProfileService(pm, reg, cache)

	// Empty name error
	if err := svc.RemoveProfile(ctx, "", ""); err == nil {
		t.Errorf("expected error for empty name, got nil")
	}

	// Non-existent profile error
	if err := svc.RemoveProfile(ctx, "claude", "ghost"); err == nil {
		t.Errorf("expected error for non-existent profile, got nil")
	}

	// Create a profile with multiple agents
	_, _ = pm.EnsureProfile("multi")
	cfg, _ := config.LoadConfig()
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}
	cfg.AddProfileAgent("multi", "claude")
	cfg.AddProfileAgent("multi", "agy")
	_ = config.SaveConfig(cfg)

	// Remove claude from multi
	err := svc.RemoveProfile(ctx, "claude", "multi")
	if err != nil {
		t.Fatalf("failed to remove agent: %v", err)
	}
	// Profile dir should still exist because agy remains
	if !pm.ProfileExists("multi") {
		t.Errorf("profile multi should still exist")
	}

	// Remove agy from multi -> should clean up directory
	err = svc.RemoveProfile(ctx, "agy", "multi")
	if err != nil {
		t.Fatalf("failed to remove last agent: %v", err)
	}
	if pm.ProfileExists("multi") {
		t.Errorf("profile multi should have been deleted")
	}

	// Create single profile and delete without agent filter
	_, _ = pm.EnsureProfile("solo")
	if err := svc.RemoveProfile(ctx, "", "solo"); err != nil {
		t.Fatalf("failed to delete solo profile: %v", err)
	}
	if pm.ProfileExists("solo") {
		t.Errorf("profile solo should have been deleted")
	}
}

func TestProfileService_RenameProfile(t *testing.T) {
	ctx := context.Background()
	_, pm, reg, cache := setupTestEnv(t)
	svc := service.NewProfileService(pm, reg, cache)

	// Empty and invalid name checks
	if err := svc.RenameProfile(ctx, "", "", "new"); err == nil {
		t.Errorf("expected error for empty oldName")
	}
	if err := svc.RenameProfile(ctx, "", "old", ""); err == nil {
		t.Errorf("expected error for empty newName")
	}

	// Create profile
	_, _ = pm.EnsureProfile("orig")
	cfg, _ := config.LoadConfig()
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}
	cfg.AddProfileAgent("orig", "claude")
	_ = config.SaveConfig(cfg)

	// Put cache entry
	_ = cache.Put(usage.Report{
		Agent:   "claude",
		Profile: "orig",
		Summary: "cached",
	})

	// Renaming with unassociated agent should error
	if err := svc.RenameProfile(ctx, "agy", "orig", "renamed"); err == nil {
		t.Errorf("expected error renaming with non-associated agent")
	}

	// Valid rename
	if err := svc.RenameProfile(ctx, "claude", "orig", "renamed"); err != nil {
		t.Fatalf("failed to rename profile: %v", err)
	}

	if pm.ProfileExists("orig") {
		t.Errorf("old profile orig should no longer exist")
	}
	if !pm.ProfileExists("renamed") {
		t.Errorf("new profile renamed should exist")
	}

	// Cache entry should be renamed
	if _, ok := cache.Get("claude", "renamed"); !ok {
		t.Errorf("expected cache entry to be migrated to renamed profile")
	}
}

type mockProcessScanner struct {
	active map[string]session.ActiveProcessInfo
}

func (s *mockProcessScanner) ScanActiveProcesses(ctx context.Context) (map[string]session.ActiveProcessInfo, error) {
	return s.active, nil
}

func TestSessionService_ListSessions(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	tmpDir := t.TempDir()
	t.Setenv("AIM_HOME", tmpDir)
	for _, prof := range []string{"work", "personal"} {
		p := filepath.Join(tmpDir, "profiles", prof)
		if err := os.MkdirAll(p, 0755); err != nil {
			t.Fatalf("failed to create profile dir %s: %v", p, err)
		}
	}

	scanner := &mockProcessScanner{
		active: map[string]session.ActiveProcessInfo{
			"11112222-3333-4444-5555-666677778888": {PID: 4242, Agent: "claude", Profile: "work"},
		},
	}
	mgr := session.NewManagerWithScanner(scanner)
	prov := &mockSessionProvider{
		agent: "claude",
		sessions: []session.Session{
			{
				ID:           "11112222-3333-4444-5555-666677778888",
				ShortID:      "11112222",
				Agent:        "claude",
				Profile:      "work",
				Title:        "Implement feature A",
				Goal:         "Feature A goal",
				Cwd:          "/projects/feature-a",
				MessageCount: 10,
				Status:       session.StatusActive,
				LastActiveAt: now.Add(-5 * time.Minute),
			},
			{
				ID:           "99998888-7777-6666-5555-444433332222",
				ShortID:      "99998888",
				Agent:        "claude",
				Profile:      "personal",
				Title:        "Debug crash",
				Goal:         "Fix null pointer",
				Cwd:          "/home/dev/app",
				MessageCount: 4,
				Status:       session.StatusIdle,
				LastActiveAt: now.Add(-10 * time.Minute),
			},
		},
	}
	mgr.RegisterProvider(prov)

	svc := service.NewSessionService(mgr, &mockLauncher{})

	// List without filter
	all, err := svc.ListSessions(ctx, service.SessionFilter{Agent: "claude"})
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(all))
	}
	if !all[0].IsActive {
		t.Errorf("expected first session to be active")
	}
	if all[0].Turns != 10 {
		t.Errorf("expected turns 10, got %d", all[0].Turns)
	}

	// Filter by profile
	workOnly, err := svc.ListSessions(ctx, service.SessionFilter{Agent: "claude", Profile: "work"})
	if err != nil {
		t.Fatalf("ListSessions with profile failed: %v", err)
	}
	if len(workOnly) != 1 || workOnly[0].Profile != "work" {
		t.Fatalf("expected 1 work session, got %v", workOnly)
	}

	// Filter by query
	crashOnly, err := svc.ListSessions(ctx, service.SessionFilter{Agent: "claude", Query: "crash"})
	if err != nil {
		t.Fatalf("ListSessions with query failed: %v", err)
	}
	if len(crashOnly) != 1 || crashOnly[0].ID != "99998888-7777-6666-5555-444433332222" {
		t.Fatalf("expected debug crash session, got %v", crashOnly)
	}

	// Limit
	limited, err := svc.ListSessions(ctx, service.SessionFilter{Agent: "claude", Limit: 1})
	if err != nil {
		t.Fatalf("ListSessions with limit failed: %v", err)
	}
	if len(limited) != 1 {
		t.Fatalf("expected 1 session with limit=1, got %d", len(limited))
	}
}

func TestSessionService_ResumeSessionInTerminal(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	t.Setenv("AIM_HOME", tmpDir)
	p := filepath.Join(tmpDir, "profiles", "work")
	_ = os.MkdirAll(p, 0755)

	mgr := session.NewManager()
	launcher := &mockLauncher{}

	prov := &mockSessionProvider{
		agent: "claude",
		sessions: []session.Session{
			{
				ID:           "abcdef12-3456-7890-abcd-ef1234567890",
				ShortID:      "abcdef12",
				Agent:        "claude",
				Profile:      "work",
				LastActiveAt: time.Now(),
			},
			{
				ID:           "12345678-abcd-ef12-3456-7890abcdef12",
				ShortID:      "12345678",
				Agent:        "claude",
				Profile:      "<host>",
				IsHost:       true,
				LastActiveAt: time.Now(),
			},
		},
	}
	mgr.RegisterProvider(prov)

	svc := service.NewSessionService(mgr, launcher)

	// Validation: missing agent
	if err := svc.ResumeSessionInTerminal(ctx, service.ResumeRequest{Agent: "", SessionID: "123"}); err == nil {
		t.Errorf("expected error for empty agent")
	}

	// Validation: missing session ID
	if err := svc.ResumeSessionInTerminal(ctx, service.ResumeRequest{Agent: "claude", SessionID: ""}); err == nil {
		t.Errorf("expected error for empty session ID")
	}

	// Validation: non-existent session
	if err := svc.ResumeSessionInTerminal(ctx, service.ResumeRequest{Agent: "claude", SessionID: "non-existent"}); err == nil {
		t.Errorf("expected error for non-existent session ID")
	}

	// Valid profile session resolution
	err := svc.ResumeSessionInTerminal(ctx, service.ResumeRequest{
		Agent:     "claude",
		SessionID: "abcdef12",
	})
	if err != nil {
		t.Fatalf("unexpected error resuming session: %v", err)
	}
	if len(launcher.launchedCmds) != 1 {
		t.Fatalf("expected 1 launched command, got %d", len(launcher.launchedCmds))
	}
	expectedCmd := "aim resume claude work abcdef12-3456-7890-abcd-ef1234567890"
	if launcher.launchedCmds[0] != expectedCmd {
		t.Errorf("expected cmd %q, got %q", expectedCmd, launcher.launchedCmds[0])
	}

	// Host session resolution (no profile in aim resume)
	launcher.launchedCmds = nil
	err = svc.ResumeSessionInTerminal(ctx, service.ResumeRequest{
		Agent:     "claude",
		SessionID: "12345678",
	})
	if err != nil {
		t.Fatalf("unexpected error resuming host session: %v", err)
	}
	expectedHostCmd := "aim resume claude 12345678-abcd-ef12-3456-7890abcdef12"
	if launcher.launchedCmds[0] != expectedHostCmd {
		t.Errorf("expected host cmd %q, got %q", expectedHostCmd, launcher.launchedCmds[0])
	}
}

func TestMCPService_ListServers(t *testing.T) {
	ctx := context.Background()
	_, pm, _, _ := setupTestEnv(t)

	// Create profile "isolated"
	_, _ = pm.EnsureProfile("isolated")

	mcpSvc := service.NewMCPService(pm)

	// Query with empty profile
	servers, err := mcpSvc.ListServers(ctx, "")
	if err != nil {
		t.Fatalf("unexpected error listing servers: %v", err)
	}
	if servers == nil {
		t.Fatalf("expected non-nil slice")
	}

	// Create profile-scoped mcp_config.json in isolated profile
	profDir := pm.ProfileDir("isolated")
	geminiConfig := filepath.Join(profDir, ".gemini", "config")
	_ = os.MkdirAll(geminiConfig, 0755)
	_ = os.WriteFile(filepath.Join(geminiConfig, "mcp_config.json"), []byte(`{
		"mcpServers": {
			"custom-tool": {
				"command": "node",
				"args": ["tool.js"]
			}
		}
	}`), 0644)

	isoServers, err := mcpSvc.ListServers(ctx, "isolated")
	if err != nil {
		t.Fatalf("unexpected error listing servers for isolated: %v", err)
	}
	found := false
	for _, s := range isoServers {
		if s.Name == "custom-tool" {
			found = true
			if s.Scope != "profile" {
				t.Errorf("expected scope profile, got %s", s.Scope)
			}
			if s.Command != "node" {
				t.Errorf("expected command node, got %s", s.Command)
			}
		}
	}
	if !found {
		t.Errorf("expected to find custom-tool in isolated profile servers")
	}
}
