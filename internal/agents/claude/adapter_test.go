package claude

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aim-cli/aim/internal/logger"
	"github.com/aim-cli/aim/internal/usage"
	"time"
)

func TestAdapter_Metadata(t *testing.T) {
	a := NewAdapter()
	if a.Name() != "claude" {
		t.Errorf("expected Name 'claude', got %q", a.Name())
	}
	if a.DisplayName() != "Claude Code" {
		t.Errorf("expected DisplayName 'Claude Code', got %q", a.DisplayName())
	}
	if a.BinaryName() != "claude" {
		t.Errorf("expected BinaryName 'claude', got %q", a.BinaryName())
	}

	aliases := a.Aliases()
	expectedAliases := map[string]bool{"cc": true, "claude-code": true}
	for _, alias := range aliases {
		if !expectedAliases[alias] {
			t.Errorf("unexpected alias %q", alias)
		}
		delete(expectedAliases, alias)
	}
	if len(expectedAliases) > 0 {
		t.Errorf("missing aliases: %v", expectedAliases)
	}

}

func TestAdapter_ResolveBinary(t *testing.T) {
	a := NewAdapter()
	bin := a.ResolveBinary()
	if bin == "" {
		t.Errorf("expected non-empty binary path")
	}
}

func TestAdapter_PrepareEnv(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", "")
	t.Setenv("AIM_REAL_HOME", tempDir)
	for _, kind := range []string{"CONFIG", "DATA", "CACHE", "STATE"} {
		t.Setenv("AIM_"+kind+"_DIR", "")
		t.Setenv("XDG_"+kind+"_HOME", filepath.Join(tempDir, kind))
	}
	bin := filepath.Join(tempDir, "bin")
	if err := os.Mkdir(bin, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bin, "security"), []byte("#!/bin/sh\nexit 44\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	profileDir := filepath.Join(tempDir, "profiles", "work")
	_ = os.MkdirAll(profileDir, 0755)

	a := NewAdapter()
	launchEnv, err := a.PrepareEnv(t.Context(), "work", profileDir)
	if err != nil {
		t.Fatalf("PrepareEnv failed: %v", err)
	}

	if launchEnv.Env["HOME"] != profileDir {
		t.Errorf("expected HOME=%s, got %s", profileDir, launchEnv.Env["HOME"])
	}
	expectedClaudeDir := filepath.Join(profileDir, ".claude")
	if launchEnv.Env["CLAUDE_CONFIG_DIR"] != expectedClaudeDir {
		t.Errorf("expected CLAUDE_CONFIG_DIR=%s, got %s", expectedClaudeDir, launchEnv.Env["CLAUDE_CONFIG_DIR"])
	}
	if launchEnv.Env["AIM_AGENT"] != "claude" {
		t.Errorf("expected AIM_AGENT=claude, got %s", launchEnv.Env["AIM_AGENT"])
	}
	if launchEnv.Env["AIM_PROFILE"] != "work" {
		t.Errorf("expected AIM_PROFILE=work, got %s", launchEnv.Env["AIM_PROFILE"])
	}
	if launchEnv.Env["AIM_HOME"] != "" {
		t.Errorf("XDG launch must not select legacy AIM_HOME: %q", launchEnv.Env["AIM_HOME"])
	}
	for _, kind := range []string{"CONFIG", "DATA", "CACHE", "STATE"} {
		key := "AIM_" + kind + "_DIR"
		if got, want := launchEnv.Env[key], filepath.Join(tempDir, kind, "aim"); got != want {
			t.Errorf("%s = %q; want %q", key, got, want)
		}
	}
	if _, ok := launchEnv.Env["AIM_SESSION_ID"]; ok {
		t.Errorf("AIM_SESSION_ID should be stripped from LaunchEnv")
	}

	fi, err := os.Stat(expectedClaudeDir)
	if err != nil {
		t.Fatalf("expected .claude directory to exist: %v", err)
	}
	if fi.Mode().Perm() != 0700 {
		t.Errorf("expected .claude permissions 0700, got %o", fi.Mode().Perm())
	}
}

func TestAdapter_HasCredentials(t *testing.T) {
	a := NewAdapter()

	// 1. Empty profile with no env var
	t.Run("empty profile", func(t *testing.T) {
		tempDir := t.TempDir()
		profileDir := filepath.Join(tempDir, "profiles", "work")
		_ = os.MkdirAll(filepath.Join(profileDir, ".claude"), 0755)

		t.Setenv("ANTHROPIC_API_KEY", "")
		t.Setenv("CLAUDE_CODE_OAUTH_TOKEN", "")
		if a.HasCredentials(profileDir) {
			t.Errorf("expected HasCredentials=false on empty profile")
		}
	})

	// 2. With ANTHROPIC_API_KEY environment variable
	t.Run("with ANTHROPIC_API_KEY env", func(t *testing.T) {
		tempDir := t.TempDir()
		profileDir := filepath.Join(tempDir, "profiles", "work")
		_ = os.MkdirAll(filepath.Join(profileDir, ".claude"), 0755)

		t.Setenv("ANTHROPIC_API_KEY", "sk-ant-test-key-123")
		t.Setenv("CLAUDE_CODE_OAUTH_TOKEN", "")
		if !a.HasCredentials(profileDir) {
			t.Errorf("expected HasCredentials=true when ANTHROPIC_API_KEY is set")
		}
	})

	// 3. With auth.json
	t.Run("with auth.json", func(t *testing.T) {
		tempDir := t.TempDir()
		profileDir := filepath.Join(tempDir, "profiles", "work")
		_ = os.MkdirAll(filepath.Join(profileDir, ".claude"), 0755)
		t.Setenv("ANTHROPIC_API_KEY", "")
		t.Setenv("CLAUDE_CODE_OAUTH_TOKEN", "")

		// Zero byte file should return false
		authFile := filepath.Join(profileDir, ".claude", "auth.json")
		_ = os.WriteFile(authFile, []byte(""), 0600)
		if a.HasCredentials(profileDir) {
			t.Errorf("expected HasCredentials=false for 0-byte auth.json")
		}

		// Non-empty file should return true
		_ = os.WriteFile(authFile, []byte(`{"token": "test-token"}`), 0600)
		if !a.HasCredentials(profileDir) {
			t.Errorf("expected HasCredentials=true with valid auth.json")
		}
	})

	// 4. .claude.json containing oauthAccount does NOT establish credentials
	t.Run("with .claude.json oauthAccount does not establish credentials", func(t *testing.T) {
		tempDir := t.TempDir()
		profileDir := filepath.Join(tempDir, "profiles", "work")
		_ = os.MkdirAll(profileDir, 0755)
		t.Setenv("ANTHROPIC_API_KEY", "")
		t.Setenv("CLAUDE_CODE_OAUTH_TOKEN", "")

		credsFile := filepath.Join(profileDir, ".claude.json")
		_ = os.WriteFile(credsFile, []byte(`{"oauthAccount": {"email": "user@example.com", "accountUuid": "acc-123"}}`), 0600)
		if a.HasCredentials(profileDir) {
			t.Errorf("expected HasCredentials=false when only .claude.json oauthAccount is present")
		}
	})

	// 5. With .credentials.json: mcpOAuth alone fails, claudeAiOauth succeeds
	t.Run("with .credentials.json validation", func(t *testing.T) {
		tempDir := t.TempDir()
		profileDir := filepath.Join(tempDir, "profiles", "work")
		_ = os.MkdirAll(filepath.Join(profileDir, ".claude"), 0755)
		t.Setenv("ANTHROPIC_API_KEY", "")
		t.Setenv("CLAUDE_CODE_OAUTH_TOKEN", "")

		credsFile := filepath.Join(profileDir, ".claude", ".credentials.json")

		// mcpOAuth only
		_ = os.WriteFile(credsFile, []byte(`{"mcpOAuth":{"test":"token"}}`), 0600)
		if a.HasCredentials(profileDir) {
			t.Errorf("expected HasCredentials=false when .credentials.json contains only mcpOAuth")
		}

		// valid claudeAiOauth
		_ = os.WriteFile(credsFile, []byte(`{"claudeAiOauth":{"accessToken":"sk-ant-access-123","refreshToken":"ref"}}`), 0600)
		if !a.HasCredentials(profileDir) {
			t.Errorf("expected HasCredentials=true when .credentials.json contains claudeAiOauth")
		}
	})
}

func TestAdapter_RewriteHooks(t *testing.T) {
	tempDir := t.TempDir()
	hostHome := filepath.Join(tempDir, "host")
	profileDir := filepath.Join(tempDir, "profile")
	_ = os.MkdirAll(filepath.Join(hostHome, ".claude"), 0755)
	_ = os.MkdirAll(filepath.Join(profileDir, ".claude"), 0755)

	hostSettings := filepath.Join(hostHome, ".claude", "settings.json")
	hostContent := `{"hooks": {"PreToolUse": "` + hostHome + `/.claude/hooks/pre.sh"}}`
	_ = os.WriteFile(hostSettings, []byte(hostContent), 0644)

	rewriteSettingsHooks(hostHome, profileDir)

	profileSettings := filepath.Join(profileDir, ".claude", "settings.json")
	data, err := os.ReadFile(profileSettings)
	if err != nil {
		t.Fatalf("failed to read profile settings: %v", err)
	}
	expected := `{"hooks": {"PreToolUse": "` + profileDir + `/.claude/hooks/pre.sh"}}`
	if string(data) != expected {
		t.Errorf("expected %s, got %s", expected, string(data))
	}

	// Test updating existing profile settings with host paths
	existingProfileContent := `{"hooks": {"PostToolUse": "` + hostHome + `/.claude/hooks/post.sh"}, "other": 123}`
	_ = os.WriteFile(profileSettings, []byte(existingProfileContent), 0644)

	rewriteSettingsHooks(hostHome, profileDir)

	dataAfter, err := os.ReadFile(profileSettings)
	if err != nil {
		t.Fatalf("failed to re-read profile settings: %v", err)
	}
	expectedAfter := `{"hooks": {"PostToolUse": "` + profileDir + `/.claude/hooks/post.sh"}, "other": 123}`
	if string(dataAfter) != expectedAfter {
		t.Errorf("expected %s, got %s", expectedAfter, string(dataAfter))
	}
}

func TestAdapter_Doctor(t *testing.T) {
	tempDir := t.TempDir()
	profileDir := filepath.Join(tempDir, "profiles", "work")
	claudeDir := filepath.Join(profileDir, ".claude")
	_ = os.MkdirAll(claudeDir, 0700)

	a := NewAdapter()
	ctx := context.Background()

	results := a.Doctor(ctx, "work", profileDir)
	if len(results) < 3 {
		t.Fatalf("expected at least 3 diagnostic results, got %d", len(results))
	}

	categories := make(map[string]bool)
	for _, r := range results {
		categories[r.Category] = true
	}
	if !categories["Binary"] {
		t.Errorf("missing Binary diagnostic result")
	}
	if !categories["Auth"] {
		t.Errorf("missing Auth diagnostic result")
	}
	if !categories["Storage"] {
		t.Errorf("missing Storage diagnostic result")
	}
}

func TestAdapter_GetUsage(t *testing.T) {
	a := NewAdapter()
	ctx := context.Background()

	report, err := a.GetUsage(ctx, "work", "/tmp/nonexistent")
	if err != nil {
		t.Fatalf("GetUsage returned unexpected error: %v", err)
	}
	if report == nil {
		t.Fatalf("GetUsage returned nil report")
	}
	if report.Agent != "claude" {
		t.Errorf("expected Agent 'claude', got %q", report.Agent)
	}
	if report.Profile != "work" {
		t.Errorf("expected Profile 'work', got %q", report.Profile)
	}
	if report.Status != usage.StatusUnknown {
		t.Errorf("expected StatusUnknown, got %v", report.Status)
	}
}

func TestAdapter_PrepareEnv_StripsHostTokens(t *testing.T) {
	tempDir := t.TempDir()
	profileDir := filepath.Join(tempDir, "profiles", "work")
	_ = os.MkdirAll(filepath.Join(profileDir, ".claude"), 0755)

	t.Setenv("CLAUDE_CODE_OAUTH_TOKEN", "sk-ant-oat01-ambient-token")
	t.Setenv("ANTHROPIC_API_KEY", "sk-ant-ambient-api-key")
	t.Setenv("CLAUDE_CODE_OAUTH_REFRESH_TOKEN", "ambient-refresh")

	a := NewAdapter()
	launch, err := a.PrepareEnv(t.Context(), "work", profileDir)
	if err != nil {
		t.Fatalf("PrepareEnv returned error: %v", err)
	}

	for _, k := range []string{"CLAUDE_CODE_OAUTH_TOKEN", "ANTHROPIC_API_KEY", "CLAUDE_CODE_OAUTH_REFRESH_TOKEN"} {
		if val, exists := launch.Env[k]; exists && val != "" {
			t.Errorf("expected %s to be stripped from LaunchEnv, got %q", k, val)
		}
	}
}

// settingsFixture is a host settings.json with the host .claude path in hooks,
// a marketplace source and statusLine.
func settingsFixture(hostHome string) string {
	h := filepath.Join(hostHome, ".claude")
	return `{
  "permissions": {"allow": ["Bash"]},
  "hooks": {"Stop": [{"command": "` + h + `/hooks/stop.sh"}]},
  "enabledPlugins": {"x@m": true, "y@m": false},
  "extraKnownMarketplaces": {"m": {"source": {"source": "directory", "path": "` + h + `/mkt"}}},
  "statusLine": {"command": "` + h + `/status.sh"}
}`
}

func TestRewriteSettingsHooks_FirstCopyHasNoPluginEnablement(t *testing.T) {
	tempDir := t.TempDir()
	hostHome := filepath.Join(tempDir, "host")
	profileDir := filepath.Join(tempDir, "profile")
	_ = os.MkdirAll(filepath.Join(hostHome, ".claude"), 0o755)
	_ = os.WriteFile(filepath.Join(hostHome, ".claude", "settings.json"), []byte(settingsFixture(hostHome)), 0o644)

	rewriteSettingsHooks(hostHome, profileDir)

	dest := filepath.Join(profileDir, ".claude", "settings.json")
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	h, p := filepath.Join(hostHome, ".claude"), filepath.Join(profileDir, ".claude")
	want := `{
  "permissions": {"allow": ["Bash"]},
  "hooks": {"Stop": [{"command": "` + p + `/hooks/stop.sh"}]},
  "statusLine": {"command": "` + h + `/status.sh"}
}`
	if string(data) != want {
		t.Fatalf("first copy =\n%s\nwant\n%s", data, want)
	}
	if fi, _ := os.Stat(dest); fi.Mode().Perm() != 0o600 {
		t.Fatalf("first copy mode = %v, want 0600", fi.Mode().Perm())
	}
}

func TestRewriteSettingsHooks_OnlyHooksAreRewritten(t *testing.T) {
	tempDir := t.TempDir()
	hostHome := filepath.Join(tempDir, "host")
	profileDir := filepath.Join(tempDir, "profile")
	_ = os.MkdirAll(filepath.Join(profileDir, ".claude"), 0o700)
	dest := filepath.Join(profileDir, ".claude", "settings.json")
	src := settingsFixture(hostHome)
	_ = os.WriteFile(dest, []byte(src), 0o644)

	rewriteSettingsHooks(hostHome, profileDir)

	data, _ := os.ReadFile(dest)
	h, p := filepath.Join(hostHome, ".claude"), filepath.Join(profileDir, ".claude")
	want := strings.Replace(src, h+"/hooks", p+"/hooks", 1)
	if string(data) != want {
		t.Fatalf("a marketplace path must not be rewritten:\n%s", data)
	}
}

func TestAdapter_DoctorHooksCheckIgnoresOtherMembers(t *testing.T) {
	tempDir := t.TempDir()
	hostHome := filepath.Join(tempDir, "host")
	profileDir := filepath.Join(tempDir, "profiles", "work")
	t.Setenv("AIM_REAL_HOME", hostHome)
	_ = os.MkdirAll(filepath.Join(profileDir, ".claude"), 0o700)
	dest := filepath.Join(profileDir, ".claude", "settings.json")
	h := filepath.Join(hostHome, ".claude")
	src := `{"hooks": {}, "extraKnownMarketplaces": {"m": {"source": {"source": "directory", "path": "` + h + `/mkt"}}}}`
	_ = os.WriteFile(dest, []byte(src), 0o644)

	hooksReported := func() bool {
		for _, r := range NewAdapter().Doctor(context.Background(), "work", profileDir) {
			if r.Category == "Hooks" {
				return true
			}
		}
		return false
	}
	if hooksReported() {
		t.Fatal("a host path outside hooks is not a hook to migrate")
	}
	if data, _ := os.ReadFile(dest); string(data) != src {
		t.Fatalf("Doctor must leave the file alone:\n%s", data)
	}
	_ = os.WriteFile(dest, []byte(`{"hooks": {"Stop": "`+h+`/hooks/stop.sh"}}`), 0o644)
	if !hooksReported() {
		t.Fatal("a host path in hooks must be migrated and reported")
	}
	if data, _ := os.ReadFile(dest); strings.Contains(string(data), h) {
		t.Fatalf("hooks not rewritten:\n%s", data)
	}
}

func TestRewriteSettingsHooks_FirstCopyIsOneWrite(t *testing.T) {
	tempDir := t.TempDir()
	hostHome := filepath.Join(tempDir, "host")
	profileDir := filepath.Join(tempDir, "profile")
	_ = os.MkdirAll(filepath.Join(hostHome, ".claude"), 0o755)
	_ = os.WriteFile(filepath.Join(hostHome, ".claude", "settings.json"), []byte(settingsFixture(hostHome)), 0o644)
	var writes int
	orig := writeSettings
	writeSettings = func(path string, data []byte, mode os.FileMode) error {
		writes++
		return orig(path, data, mode)
	}
	defer func() { writeSettings = orig }()

	rewriteSettingsHooks(hostHome, profileDir)

	if writes != 1 {
		t.Fatalf("first copy took %d writes, want 1", writes)
	}
	dest := filepath.Join(profileDir, ".claude", "settings.json")
	data, _ := os.ReadFile(dest)
	h, p := filepath.Join(hostHome, ".claude"), filepath.Join(profileDir, ".claude")
	want := `{
  "permissions": {"allow": ["Bash"]},
  "hooks": {"Stop": [{"command": "` + p + `/hooks/stop.sh"}]},
  "statusLine": {"command": "` + h + `/status.sh"}
}`
	if string(data) != want {
		t.Fatalf("first copy =\n%s\nwant\n%s", data, want)
	}
	if fi, _ := os.Stat(dest); fi.Mode().Perm() != 0o600 {
		t.Fatalf("first copy mode = %v, want 0600", fi.Mode().Perm())
	}
}

// A host settings.json the splicer cannot strip safely (a duplicate top-level
// key) is not copied at all: seeding it would carry the host's enablement.
func TestRewriteSettingsHooks_UnstrippableHostIsNotCopied(t *testing.T) {
	var warned strings.Builder
	logger.SetWarnOutput(&warned)
	defer logger.Reset()
	tempDir := t.TempDir()
	hostHome := filepath.Join(tempDir, "host")
	profileDir := filepath.Join(tempDir, "profile")
	_ = os.MkdirAll(filepath.Join(hostHome, ".claude"), 0o755)
	_ = os.WriteFile(filepath.Join(hostHome, ".claude", "settings.json"),
		[]byte(`{"enabledPlugins": {"x@m": true}, "hooks": {}, "enabledPlugins": {"y@m": true}}`), 0o644)

	rewriteSettingsHooks(hostHome, profileDir)

	if _, err := os.Stat(filepath.Join(profileDir, ".claude", "settings.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("an unstrippable host copy must not be written (stat err = %v)", err)
	}
	if !strings.Contains(warned.String(), "settings.json") {
		t.Fatalf("want a visible warning, got %q", warned.String())
	}
}

// Only a missing settings.json is a first copy; any other stat error (here a
// symlink loop) must not be treated as absent and overwritten.
func TestRewriteSettingsHooks_StatErrorIsNotAbsent(t *testing.T) {
	var warned strings.Builder
	logger.SetWarnOutput(&warned)
	defer logger.Reset()
	tempDir := t.TempDir()
	hostHome := filepath.Join(tempDir, "host")
	profileDir := filepath.Join(tempDir, "profile")
	_ = os.MkdirAll(filepath.Join(hostHome, ".claude"), 0o755)
	_ = os.MkdirAll(filepath.Join(profileDir, ".claude"), 0o700)
	_ = os.WriteFile(filepath.Join(hostHome, ".claude", "settings.json"), []byte(settingsFixture(hostHome)), 0o644)
	dest := filepath.Join(profileDir, ".claude", "settings.json")
	if err := os.Symlink(dest, dest); err != nil {
		t.Fatal(err)
	}

	rewriteSettingsHooks(hostHome, profileDir)

	if fi, err := os.Lstat(dest); err != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("settings.json was overwritten (lstat = %v, %v)", fi, err)
	}
	if !strings.Contains(warned.String(), dest) {
		t.Fatalf("want a visible warning naming %s, got %q", dest, warned.String())
	}
}

func TestParseClaudeUsage_Standard(t *testing.T) {
	output := `You are currently using your subscription to power your Claude Code usage

Current session: 4% used · resets Oct 2 at 3:19pm (Asia/Calcutta)
Current week (all models): 37% used · resets Oct 7 at 7:29am (Asia/Calcutta)
Current week (Fable): 11% used · resets Oct 7 at 7:29am (Asia/Calcutta)

What's contributing to your limits usage?
Approximate, based on local sessions on this machine — does not include other devices or claude.ai. Behaviors are independent characteristics, not a breakdown.

Last 24h · 79 requests · 79 sessions
  19% of your usage was while 4+ sessions ran in parallel

Last 7d · 85 requests · 85 sessions
  18% of your usage was while 4+ sessions ran in parallel`

	now := time.Date(2026, time.October, 2, 8, 0, 0, 0, time.UTC)
	windows := ParseClaudeUsage(output, now)
	if len(windows) != 3 {
		t.Fatalf("expected 3 windows, got %d", len(windows))
	}

	// Window 0: Current session
	w0 := windows[0]
	if w0.Name != "5h Limit" {
		t.Errorf("w0: expected name '5h Limit', got %q", w0.Name)
	}
	if w0.Category != "All Models" {
		t.Errorf("w0: expected category 'All Models', got %q", w0.Category)
	}
	if w0.RemainingPct != 96 {
		t.Errorf("w0: expected RemainingPct 96, got %d", w0.RemainingPct)
	}
	if !w0.IsHourly() {
		t.Errorf("w0: expected IsHourly() to be true")
	}

	// Window 1: Current week (all models)
	w1 := windows[1]
	if w1.Name != "Weekly Limit" {
		t.Errorf("w1: expected name 'Weekly Limit', got %q", w1.Name)
	}
	if w1.Category != "All Models" {
		t.Errorf("w1: expected category 'All Models', got %q", w1.Category)
	}
	if w1.RemainingPct != 63 {
		t.Errorf("w1: expected RemainingPct 63, got %d", w1.RemainingPct)
	}
	if !w1.IsWeekly() {
		t.Errorf("w1: expected IsWeekly() to be true")
	}

	// Window 2: Current week (Fable)
	w2 := windows[2]
	if w2.Name != "Weekly Limit" {
		t.Errorf("w2: expected name 'Weekly Limit', got %q", w2.Name)
	}
	if w2.Category != "Fable" {
		t.Errorf("w2: expected category 'Fable', got %q", w2.Category)
	}
	if w2.RemainingPct != 89 {
		t.Errorf("w2: expected RemainingPct 89, got %d", w2.RemainingPct)
	}

	status := usage.CalculateStatus(windows)
	if status != usage.StatusOK {
		t.Errorf("expected status OK, got %v", status)
	}
}

func TestParseClaudeUsage_EdgeCases(t *testing.T) {
	output := `Current session: 100% used
Current week (all models): 0% used · resets in 5 hours`

	now := time.Now()
	windows := ParseClaudeUsage(output, now)
	if len(windows) != 2 {
		t.Fatalf("expected 2 windows, got %d", len(windows))
	}
	if windows[0].RemainingPct != 0 {
		t.Errorf("expected remaining 0, got %d", windows[0].RemainingPct)
	}
	if windows[1].RemainingPct != 100 {
		t.Errorf("expected remaining 100, got %d", windows[1].RemainingPct)
	}

	status := usage.CalculateStatus(windows)
	if status != usage.StatusExhausted {
		t.Errorf("expected StatusExhausted, got %v", status)
	}
}

func TestClaudeAdapter_GetUsage_NoCredentials(t *testing.T) {
	tempDir := t.TempDir()
	a := NewAdapter()
	rep, err := a.GetUsage(context.Background(), "test", tempDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rep.Status != usage.StatusUnknown {
		t.Errorf("expected StatusUnknown, got %v", rep.Status)
	}
	if rep.Error != "no credentials" {
		t.Errorf("expected 'no credentials', got %q", rep.Error)
	}
}

func TestParseClaudeUsage_PassedResetTime(t *testing.T) {
	output := `Current session: 50% used · resets Oct 2 at 3:19pm (Asia/Calcutta)`
	// Now is after 3:19pm IST (15:19 IST = 09:49 UTC)
	now := time.Date(2026, time.October, 2, 16, 0, 0, 0, time.UTC)
	windows := ParseClaudeUsage(output, now)
	if len(windows) != 1 {
		t.Fatalf("expected 1 window, got %d", len(windows))
	}
	if windows[0].RemainingPct != 100 {
		t.Errorf("expected remaining 100%% after reset passed, got %d", windows[0].RemainingPct)
	}
	if windows[0].ResetsIn != 0 {
		t.Errorf("expected resetsIn 0 after reset passed, got %v", windows[0].ResetsIn)
	}
}

func TestClaudeAdapter_GetUsage_Success(t *testing.T) {
	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	_ = os.MkdirAll(binDir, 0755)

	mockScript := filepath.Join(binDir, "claude")
	content := "#!/bin/sh\n" +
		"echo 'Current session: 5% used · resets in 2 hours'\n" +
		"echo 'Current week (all models): 15% used · resets in 5 days'\n"
	if err := os.WriteFile(mockScript, []byte(content), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	profileDir := filepath.Join(tempDir, "profile")
	_ = os.MkdirAll(filepath.Join(profileDir, ".claude"), 0700)
	_ = os.WriteFile(filepath.Join(profileDir, ".claude", "auth.json"), []byte(`{"apiKey":"test-key"}`), 0600)

	a := NewAdapter()
	rep, err := a.GetUsage(context.Background(), "test", profileDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rep.Status != usage.StatusOK {
		t.Errorf("expected StatusOK, got %v (error: %q)", rep.Status, rep.Error)
	}
	if len(rep.Windows) != 2 {
		t.Fatalf("expected 2 windows, got %d", len(rep.Windows))
	}
	if rep.Windows[0].RemainingPct != 95 {
		t.Errorf("expected remaining 95, got %d", rep.Windows[0].RemainingPct)
	}
	if rep.Windows[1].RemainingPct != 85 {
		t.Errorf("expected remaining 85, got %d", rep.Windows[1].RemainingPct)
	}
	if !strings.Contains(rep.Summary, "5h: 95%") {
		t.Errorf("expected summary to contain '5h: 95%%', got %q", rep.Summary)
	}
}
