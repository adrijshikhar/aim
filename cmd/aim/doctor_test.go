package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/oauth"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/usage"
)

type doctorMockAdapter struct {
	mockAdapter
	failDr bool
}

func (m *doctorMockAdapter) Doctor(ctx context.Context, profileName, profileDir string) []agents.DiagnosticResult {
	if m.failDr {
		return []agents.DiagnosticResult{
			{Category: "Binary", Status: "FAIL", Message: "binary missing or broken"},
		}
	}
	return []agents.DiagnosticResult{
		{Category: "Binary", Status: "OK", Message: "binary ready"},
	}
}

func TestDoctorVPNDetection(t *testing.T) {
	var buf bytes.Buffer
	diagnoseVPNTo(&buf)
	out := buf.String()
	if !strings.Contains(out, "P2P Sharing & VPN Tunnel") {
		t.Errorf("expected doctor output to contain P2P Sharing & VPN Tunnel section, got: %s", out)
	}
}

func TestDoctorCmd_CheckFlag(t *testing.T) {
	tempDir := t.TempDir()
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("testprof")

	reg := agents.NewRegistry()
	reg.Register(&doctorMockAdapter{mockAdapter: mockAdapter{name: "healthy", hasCreds: true}, failDr: false})
	reg.Register(&doctorMockAdapter{mockAdapter: mockAdapter{name: "broken", hasCreds: true}, failDr: true})

	cmdHealthy := newDoctorCmd(reg, pm)
	cmdHealthy.SetArgs([]string{"healthy", "--check"})
	if err := cmdHealthy.Execute(); err != nil {
		t.Fatalf("expected healthy doctor --check to succeed, got: %v", err)
	}

	cmdBroken := newDoctorCmd(reg, pm)
	cmdBroken.SetArgs([]string{"broken", "--check"})
	if err := cmdBroken.Execute(); err == nil {
		t.Fatalf("expected broken doctor --check to fail with error")
	}

	cmdAllBroken := newDoctorCmd(reg, pm)
	cmdAllBroken.SetArgs([]string{"--check"})
	if err := cmdAllBroken.Execute(); err == nil {
		t.Fatalf("expected doctor --check across all agents to fail if broken agent present")
	}

	// Without --check, command exits cleanly (exit 0)
	cmdNoCheck := newDoctorCmd(reg, pm)
	cmdNoCheck.SetArgs([]string{"broken"})
	if err := cmdNoCheck.Execute(); err != nil {
		t.Fatalf("expected doctor without --check to exit cleanly, got: %v", err)
	}
}

// The Bridge block describes the profile, not the agent, so a profile with two
// agents shows it once, under the first agent's block.
func TestDoctor_BridgeBlockOncePerProfile(t *testing.T) {
	host := t.TempDir()
	t.Setenv("AIM_REAL_HOME", host)
	if err := os.WriteFile(filepath.Join(host, ".gitconfig"), []byte("[user]"), 0644); err != nil {
		t.Fatal(err)
	}
	pm := profile.NewProfileManager(t.TempDir())
	if _, err := pm.EnsureProfile("shared"); err != nil {
		t.Fatal(err)
	}
	reg := agents.NewRegistry()
	reg.Register(&doctorMockAdapter{mockAdapter: mockAdapter{name: "one", hasCreds: true}})
	reg.Register(&doctorMockAdapter{mockAdapter: mockAdapter{name: "two", hasCreds: true}})

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	runDoctor(reg, pm, "")
	_ = w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	out := buf.String()

	if n := strings.Count(out, "Profile: "); n != 2 {
		t.Fatalf("expected the profile under both agents, got %d blocks:\n%s", n, out)
	}
	if n := strings.Count(out, "Bridge:"); n != 1 {
		t.Errorf("expected one Bridge line for the profile, got %d:\n%s", n, out)
	}
}

func TestDoctor_UnclaimedTokens(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)

	store := oauth.NewStashStore(tempDir)
	_ = store.Save("spent-tok", "claude", "work", "acc", "ref", time.Now().Add(time.Hour))

	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)

	cmd := newRootCmd(reg, pm)
	cmd.SetArgs([]string{"doctor"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	_ = cmd.Execute()
	output := buf.String()

	if !strings.Contains(output, "Unclaimed OAuth tokens found") {
		t.Errorf("expected doctor to report unclaimed OAuth tokens, got: %s", output)
	}
}

func TestDoctor_UnclaimedTokens_FilterByAgent(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)

	store := oauth.NewStashStore(tempDir)
	_ = store.Save("spent-claude", "claude", "work", "acc", "ref", time.Now().Add(time.Hour))
	_ = store.Save("spent-agy", "agy", "personal", "acc", "ref", time.Now().Add(time.Hour))

	reg := agents.NewRegistry()
	reg.Register(&doctorMockAdapter{mockAdapter: mockAdapter{name: "claude", hasCreds: true}})
	reg.Register(&doctorMockAdapter{mockAdapter: mockAdapter{name: "agy", hasCreds: true}})
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("work")
	_, _ = pm.EnsureProfile("personal")

	// Running doctor for "claude" should report claude's unclaimed token, not agy's
	cmd := newRootCmd(reg, pm)
	cmd.SetArgs([]string{"doctor", "claude"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	_ = cmd.Execute()
	output := buf.String()

	if !strings.Contains(output, "Unclaimed OAuth tokens found (1 uncommitted") {
		t.Errorf("expected doctor to report 1 unclaimed token for claude, got:\n%s", output)
	}
	if strings.Contains(output, "Agent: agy") {
		t.Errorf("expected doctor claude not to report agy token, got:\n%s", output)
	}
}

func TestDoctor_UnclaimedTokens_ReportsPruning(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)

	store := oauth.NewStashStore(tempDir)
	spent := "spent-old-to-prune"
	_ = store.Save(spent, "claude", "work", "acc", "ref", time.Now())
	oldPath := store.StashPath(spent)
	oldTime := time.Now().Add(-10 * 24 * time.Hour)
	_ = os.Chtimes(oldPath, oldTime, oldTime)

	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)

	cmd := newRootCmd(reg, pm)
	cmd.SetArgs([]string{"doctor"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	_ = cmd.Execute()
	output := buf.String()

	if !strings.Contains(output, "Pruned 1 expired unclaimed token record(s)") {
		t.Errorf("expected doctor to report pruned records, got:\n%s", output)
	}
}

func TestDoctor_RefusalRendering(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)

	refStore := oauth.NewRefusalStore(tempDir)
	if err := refStore.Record("claude", "work", oauth.RefusalRefreshTokenReused, "refresh token reused"); err != nil {
		t.Fatalf("failed to record refusal: %v", err)
	}

	reg := agents.NewRegistry()
	reg.Register(&doctorMockAdapter{mockAdapter: mockAdapter{name: "claude", hasCreds: true}, failDr: false})

	pm := profile.NewProfileManager(tempDir)
	if _, err := pm.EnsureProfile("work"); err != nil {
		t.Fatalf("failed to ensure profile: %v", err)
	}

	cmd := newRootCmd(reg, pm)
	cmd.SetArgs([]string{"doctor", "claude"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	_ = cmd.Execute()
	output := buf.String()

	expectedSubstrings := []string{
		"[FAIL]",
		"Auth:",
		"Re-login Required (refresh_token_reused)",
		"Remedy: Run 'aim login claude work'",
	}

	for _, sub := range expectedSubstrings {
		if !strings.Contains(output, sub) {
			t.Errorf("expected doctor output to contain %q, but got:\n%s", sub, output)
		}
	}

	// Also verify that doctor --check fails when refusal is present
	cmdCheck := newDoctorCmd(reg, pm)
	cmdCheck.SetArgs([]string{"claude", "--check"})
	if err := cmdCheck.Execute(); err == nil {
		t.Errorf("expected doctor --check to return an error when refusal is active")
	}
}

func TestLogin_ClearsRefusalStore(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)

	refStore := oauth.NewRefusalStore(tempDir)
	if err := refStore.Record("claude", "work", oauth.RefusalRefreshTokenReused, "token consumed"); err != nil {
		t.Fatalf("failed to record refusal: %v", err)
	}

	if _, ok := refStore.IsRefused("claude", "work"); !ok {
		t.Fatalf("expected refusal before login")
	}

	reg := agents.NewRegistry()
	reg.Register(&mockAdapter{name: "claude"})
	pm := profile.NewProfileManager(tempDir)

	cmd := newRootCmd(reg, pm)
	cmd.SetArgs([]string{"login", "claude", "work"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected login to succeed, got: %v", err)
	}

	// Verify that the refusal was cleared from the store
	if rec, ok := refStore.IsRefused("claude", "work"); ok || rec != nil {
		t.Fatalf("expected refusal to be cleared after successful login, got: ok=%v, rec=%v", ok, rec)
	}
}

type aliasMockAdapter struct {
	mockAdapter
	customAliases []string
}

func (a *aliasMockAdapter) Aliases() []string {
	if a.customAliases != nil {
		return a.customAliases
	}
	return a.mockAdapter.Aliases()
}

func TestLogin_ClearsRefusalStore_Alias(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)

	refStore := oauth.NewRefusalStore(tempDir)
	// Record refusal under canonical name "claude"
	if err := refStore.Record("claude", "work", oauth.RefusalRefreshTokenReused, "token consumed"); err != nil {
		t.Fatalf("failed to record refusal: %v", err)
	}
	// Also record refusal under alias "anthropic" defensively
	if err := refStore.Record("anthropic", "work", oauth.RefusalInvalidGrant, "invalid grant"); err != nil {
		t.Fatalf("failed to record alias refusal: %v", err)
	}

	if _, ok := refStore.IsRefused("claude", "work"); !ok {
		t.Fatalf("expected canonical refusal before login")
	}
	if _, ok := refStore.IsRefused("anthropic", "work"); !ok {
		t.Fatalf("expected alias refusal before login")
	}

	// Also seed a cached failure report to verify it is invalidated upon login
	cacheStore := usage.NewCacheStore(tempDir, usage.DefaultTTL)
	_ = cacheStore.Put(usage.Report{
		Agent:   "claude",
		Profile: "work",
		Status:  usage.StatusUnknown,
		Error:   "re-login required (invalid_grant)",
	})

	reg := agents.NewRegistry()
	reg.Register(&aliasMockAdapter{
		mockAdapter:   mockAdapter{name: "claude"},
		customAliases: []string{"anthropic"},
	})
	pm := profile.NewProfileManager(tempDir)

	// Log in via alias "anthropic"
	cmd := newRootCmd(reg, pm)
	cmd.SetArgs([]string{"login", "anthropic", "work"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected login via alias to succeed, got: %v", err)
	}

	// Verify that both canonical and alias refusals were cleared
	if rec, ok := refStore.IsRefused("claude", "work"); ok || rec != nil {
		t.Fatalf("expected canonical refusal to be cleared after login via alias, got: ok=%v, rec=%v", ok, rec)
	}
	if rec, ok := refStore.IsRefused("anthropic", "work"); ok || rec != nil {
		t.Fatalf("expected alias refusal to be cleared after login via alias, got: ok=%v, rec=%v", ok, rec)
	}

	// Verify that the stale usage failure report was deleted from cache
	freshCache := usage.NewCacheStore(tempDir, usage.DefaultTTL)
	if rep, found := freshCache.Get("claude", "work"); found {
		t.Fatalf("expected cached usage report to be invalidated after login, got: %+v", rep)
	}
}

func TestDoctor_IdleProfileAuthMessage(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)

	pm := profile.NewProfileManager(tempDir)
	if _, err := pm.EnsureProfile("idle-prof"); err != nil {
		t.Fatalf("failed to ensure profile: %v", err)
	}

	cfg, _ := config.LoadConfig()
	cfg.AddProfileAgent("idle-prof", "claude")
	_ = config.SaveConfig(cfg)

	reg := agents.NewRegistry()
	reg.Register(&mockAdapter{name: "claude", hasCreds: true})

	cmd := newRootCmd(reg, pm)
	cmd.SetArgs([]string{"doctor", "claude"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected doctor to succeed, got: %v", err)
	}
	output := buf.String()

	cleanOut := regexp.MustCompile(`\x1b\[[0-9;]*m`).ReplaceAllString(output, "")
	expectedStr := "[OK]     Auth: Credentials configured (Idle - will renew on next turn)"
	if !strings.Contains(cleanOut, expectedStr) {
		t.Errorf("expected doctor output to contain %q, but got:\n%s", expectedStr, output)
	}
}

type doctorCachedUsageMockAdapter struct {
	doctorMockAdapter
	cachedUsageCalled bool
}

func (m *doctorCachedUsageMockAdapter) GetCachedUsage(ctx context.Context, profileName, profileDir string) (*usage.Report, error) {
	m.cachedUsageCalled = true
	return &usage.Report{
		Agent:   m.name,
		Profile: profileName,
		Status:  usage.StatusOK,
		Windows: []usage.LimitWindow{
			{Name: "Session (5-hour)", RemainingPct: 90},
		},
	}, nil
}

func TestDoctorCmd_CachedUsageProvider(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("cached-prof")

	pDir := pm.ProfileDir("cached-prof")
	_ = os.WriteFile(filepath.Join(pDir, "auth.json"), []byte(`{"tokens":{"access_token":"valid"}}`), 0600)

	cfg, _ := config.LoadConfig()
	cfg.AddProfileAgent("cached-prof", "claude")
	_ = config.SaveConfig(cfg)

	cachedAdapter := &doctorCachedUsageMockAdapter{
		doctorMockAdapter: doctorMockAdapter{
			mockAdapter: mockAdapter{name: "claude", hasCreds: true},
		},
	}
	reg := agents.NewRegistry()
	reg.Register(cachedAdapter)

	cmd := newRootCmd(reg, pm)
	cmd.SetArgs([]string{"doctor", "claude"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected doctor to succeed, got: %v", err)
	}

	if !cachedAdapter.cachedUsageCalled {
		t.Errorf("expected GetCachedUsage to be called during doctor run")
	}

	outStr := buf.String()
	if !strings.Contains(outStr, "AIM Doctor Diagnostics") {
		t.Errorf("expected doctor output to contain 'AIM Doctor Diagnostics', got:\n%s", outStr)
	}
}
