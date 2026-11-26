package agents_test

import (
	"context"
	"strings"
	"testing"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/usage"
)

type fullMockAdapter struct {
	agents.BaseAdapter
}

func (m *fullMockAdapter) HasCredentials(profileDir string) bool { return true }
func (m *fullMockAdapter) PrepareEnv(ctx context.Context, profileName, profileDir string) (agents.LaunchEnv, error) {
	return agents.LaunchEnv{}, nil
}

func (m *fullMockAdapter) GetCachedUsage(ctx context.Context, profileName, profileDir string) (*usage.Report, error) {
	return &usage.Report{Profile: profileName}, nil
}

func (m *fullMockAdapter) GetLiveUsage(ctx context.Context, profileName, profileDir string) (*usage.Report, error) {
	return &usage.Report{Profile: profileName}, nil
}

func (m *fullMockAdapter) HarvestCredentials(ctx context.Context, profileName, profileDir string) error {
	return nil
}

func (m *fullMockAdapter) PurgeCredentials(ctx context.Context, profileName, profileDir string) error {
	return nil
}

func (m *fullMockAdapter) SeedCredentials(_ context.Context, _, _ string) error {
	return nil
}

func (m *fullMockAdapter) ResumeArgs(sessionID string, extraArgs []string) []string {
	return append([]string{"--resume", sessionID}, extraArgs...)
}

type unextendedMockAdapter struct {
	agents.BaseAdapter
}

func (m *unextendedMockAdapter) HasCredentials(profileDir string) bool { return false }
func (m *unextendedMockAdapter) PrepareEnv(ctx context.Context, profileName, profileDir string) (agents.LaunchEnv, error) {
	return agents.LaunchEnv{}, nil
}

func TestSegregatedInterfaces_Assertions(t *testing.T) {
	full := &fullMockAdapter{
		BaseAdapter: agents.NewBaseAdapter("full", "Full Mock", "mock", nil),
	}

	// Verify full implements AgentAdapter
	var _ agents.AgentAdapter = full

	// Verify type assertions for newly segregated interfaces
	cachedProvider, ok := any(full).(agents.CachedUsageProvider)
	if !ok {
		t.Fatal("expected full mock to implement CachedUsageProvider")
	}
	report, err := cachedProvider.GetCachedUsage(context.Background(), "test", "/tmp")
	if err != nil || report == nil || report.Profile != "test" {
		t.Fatalf("unexpected GetCachedUsage result: report=%v, err=%v", report, err)
	}

	liveProvider, ok := any(full).(agents.LiveUsageProvider)
	if !ok {
		t.Fatal("expected full mock to implement LiveUsageProvider")
	}
	report, err = liveProvider.GetLiveUsage(context.Background(), "test", "/tmp")
	if err != nil || report == nil || report.Profile != "test" {
		t.Fatalf("unexpected GetLiveUsage result: report=%v, err=%v", report, err)
	}

	harvester, ok := any(full).(agents.CredentialHarvester)
	if !ok {
		t.Fatal("expected full mock to implement CredentialHarvester")
	}
	if err := harvester.HarvestCredentials(context.Background(), "test", "/tmp"); err != nil {
		t.Fatalf("unexpected HarvestCredentials error: %v", err)
	}

	purger, ok := any(full).(agents.CredentialPurger)
	if !ok {
		t.Fatal("expected full mock to implement CredentialPurger")
	}
	if err := purger.PurgeCredentials(context.Background(), "test", "/tmp"); err != nil {
		t.Fatalf("unexpected PurgeCredentials error: %v", err)
	}

	seeder, ok := any(full).(agents.CredentialSeeder)
	if !ok {
		t.Fatal("expected full mock to implement CredentialSeeder")
	}
	if err := seeder.SeedCredentials(context.Background(), "test", "/tmp"); err != nil {
		t.Fatalf("unexpected SeedCredentials error: %v", err)
	}

	resumer, ok := any(full).(agents.SessionResumer)
	if !ok {
		t.Fatal("expected full mock to implement SessionResumer")
	}
	args := resumer.ResumeArgs("sess-123", []string{"--debug"})
	if len(args) != 3 || args[0] != "--resume" || args[1] != "sess-123" || args[2] != "--debug" {
		t.Fatalf("unexpected ResumeArgs output: %v", args)
	}

	// Verify unextended mock does NOT implement any of them
	unextended := &unextendedMockAdapter{
		BaseAdapter: agents.NewBaseAdapter("unextended", "Unextended Mock", "mock", nil),
	}
	if _, ok := any(unextended).(agents.CachedUsageProvider); ok {
		t.Error("unextended mock should not implement CachedUsageProvider")
	}
	if _, ok := any(unextended).(agents.LiveUsageProvider); ok {
		t.Error("unextended mock should not implement LiveUsageProvider")
	}
	if _, ok := any(unextended).(agents.CredentialHarvester); ok {
		t.Error("unextended mock should not implement CredentialHarvester")
	}
	if _, ok := any(unextended).(agents.CredentialPurger); ok {
		t.Error("unextended mock should not implement CredentialPurger")
	}
	if _, ok := any(unextended).(agents.CredentialSeeder); ok {
		t.Error("unextended mock should not implement CredentialSeeder")
	}
	if _, ok := any(unextended).(agents.SessionResumer); ok {
		t.Error("unextended mock should not implement SessionResumer")
	}
}

func TestBuildEnv_StripsAmbientProviderKeys(t *testing.T) {
	ambient := []string{
		"PATH=/usr/bin",
		"KEEP_ME=ok",
		"HOME=/host/home",
		"SSH_CONNECTION=1.2.3.4 22",
		"ANTHROPIC_API_KEY=host-ant",
		"CLAUDE_CODE_OAUTH_TOKEN=host-oauth",
		"CLAUDE_CODE_OAUTH_REFRESH_TOKEN=host-refresh",
		"OPENAI_API_KEY=host-openai",
		"CODEX_API_KEY=host-codex",
		"GEMINI_API_KEY=host-gemini",
		"GOOGLE_API_KEY=host-google",
		"CURSOR_API_KEY=host-cursor",
		"CURSOR_AUTH_TOKEN=host-cursor-auth",
		"CODEX_HOME=/host/.codex",
		"CLAUDE_CONFIG_DIR=/host/.claude",
		"GEMINI_CLI_HOME=/host/.gemini",
	}

	got := agents.BuildEnv(ambient, map[string]string{
		"HOME":           "/profiles/work",
		"AIM_PROFILE":    "work",
		"OPENAI_API_KEY": "profile-explicit-key",
	})

	envMap := make(map[string]string, len(got))
	for _, kv := range got {
		k, v, ok := strings.Cut(kv, "=")
		if ok {
			envMap[k] = v
		}
	}

	if envMap["KEEP_ME"] != "ok" || envMap["PATH"] != "/usr/bin" {
		t.Fatalf("expected non-sensitive vars kept, got %v", envMap)
	}
	if envMap["HOME"] != "/profiles/work" || envMap["AIM_PROFILE"] != "work" {
		t.Fatalf("expected profile overrides applied, got %v", envMap)
	}
	if envMap["OPENAI_API_KEY"] != "profile-explicit-key" {
		t.Fatalf("expected explicit profile override for OPENAI_API_KEY, got %q", envMap["OPENAI_API_KEY"])
	}

	stripped := []string{
		"SSH_CONNECTION", "ANTHROPIC_API_KEY", "CLAUDE_CODE_OAUTH_TOKEN",
		"CLAUDE_CODE_OAUTH_REFRESH_TOKEN", "CODEX_API_KEY", "GEMINI_API_KEY",
		"GOOGLE_API_KEY", "CURSOR_API_KEY", "CURSOR_AUTH_TOKEN",
		"CODEX_HOME", "CLAUDE_CONFIG_DIR", "GEMINI_CLI_HOME",
	}
	for _, k := range stripped {
		if _, exists := envMap[k]; exists {
			t.Errorf("expected ambient key %s to be stripped", k)
		}
	}
}
