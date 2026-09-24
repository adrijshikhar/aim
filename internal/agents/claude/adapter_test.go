package claude

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/usage"
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

	aliasAdapter := NewClaudeAdapter()
	if aliasAdapter.Name() != "claude" {
		t.Errorf("expected NewClaudeAdapter to return claude adapter")
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
	profileDir := filepath.Join(tempDir, "profiles", "work")
	_ = os.MkdirAll(profileDir, 0755)

	a := NewAdapter()
	launchEnv, err := a.PrepareEnv("work", profileDir)
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
	if launchEnv.Env["AIM_HOME"] != config.BaseDir() {
		t.Errorf("expected AIM_HOME=%s, got %s", config.BaseDir(), launchEnv.Env["AIM_HOME"])
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
	launch, err := a.PrepareEnv("work", profileDir)
	if err != nil {
		t.Fatalf("PrepareEnv returned error: %v", err)
	}

	for _, k := range []string{"CLAUDE_CODE_OAUTH_TOKEN", "ANTHROPIC_API_KEY", "CLAUDE_CODE_OAUTH_REFRESH_TOKEN"} {
		if val, exists := launch.Env[k]; exists && val != "" {
			t.Errorf("expected %s to be stripped from LaunchEnv, got %q", k, val)
		}
	}
}

