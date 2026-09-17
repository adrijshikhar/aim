package codex

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/usage"
)

func TestCodexAdapter_Metadata(t *testing.T) {
	a := NewAdapter()
	if a.Name() != "codex" {
		t.Errorf("expected name 'codex', got %q", a.Name())
	}
	if a.DisplayName() != "Codex CLI" {
		t.Errorf("expected displayName 'Codex CLI', got %q", a.DisplayName())
	}
	if a.BinaryName() != "codex" {
		t.Errorf("expected binaryName 'codex', got %q", a.BinaryName())
	}
	aliases := a.Aliases()
	if len(aliases) == 0 || aliases[0] != "codex-cli" {
		t.Errorf("expected aliases to include 'codex-cli', got %v", aliases)
	}

	aliasAdapter := NewCodexAdapter()
	if aliasAdapter.Name() != "codex" {
		t.Errorf("expected NewCodexAdapter to return codex adapter")
	}
}

func TestCodexAdapter_HasCredentials(t *testing.T) {
	a := NewAdapter()
	tmpDir := t.TempDir()

	// Missing credentials
	if a.HasCredentials(tmpDir) {
		t.Errorf("expected HasCredentials=false for empty dir")
	}

	// Valid auth.json
	authDir := filepath.Join(tmpDir, ".codex")
	_ = os.MkdirAll(authDir, 0755)
	authFile := filepath.Join(authDir, "auth.json")
	_ = os.WriteFile(authFile, []byte(`{"tokens":{"access_token":"mock-token"}}`), 0600)

	if !a.HasCredentials(tmpDir) {
		t.Errorf("expected HasCredentials=true for valid auth.json")
	}
}

func TestCodexAdapter_PrepareEnv(t *testing.T) {
	a := NewAdapter()
	tmpDir := t.TempDir()
	profileDir := filepath.Join(tmpDir, "profiles", "work")

	launchEnv, err := a.PrepareEnv("work", profileDir)
	if err != nil {
		t.Fatalf("PrepareEnv failed: %v", err)
	}

	codexDir := filepath.Join(profileDir, ".codex")
	if fi, err := os.Stat(codexDir); err != nil || !fi.IsDir() {
		t.Errorf("expected .codex directory to be created in profileDir")
	} else if fi.Mode().Perm() != 0700 {
		t.Errorf("expected .codex directory permissions 0700, got %v", fi.Mode().Perm())
	}

	if launchEnv.Env["HOME"] != profileDir {
		t.Errorf("expected HOME=%q, got %q", profileDir, launchEnv.Env["HOME"])
	}
	if launchEnv.Env["CODEX_HOME"] != codexDir {
		t.Errorf("expected CODEX_HOME=%q, got %q", codexDir, launchEnv.Env["CODEX_HOME"])
	}
	if launchEnv.Env["AIM_AGENT"] != "codex" {
		t.Errorf("expected AIM_AGENT='codex', got %q", launchEnv.Env["AIM_AGENT"])
	}
	if launchEnv.Env["AIM_PROFILE"] != "work" {
		t.Errorf("expected AIM_PROFILE='work', got %q", launchEnv.Env["AIM_PROFILE"])
	}
}

func TestCodexAdapter_Doctor(t *testing.T) {
	a := NewAdapter()
	tmpDir := t.TempDir()
	profileDir := filepath.Join(tmpDir, "profiles", "test")
	_ = os.MkdirAll(filepath.Join(profileDir, ".codex"), 0755)

	// Doctor with no credentials
	results := a.Doctor(context.Background(), "test", profileDir)
	if len(results) == 0 {
		t.Fatalf("expected diagnostic results, got none")
	}

	var foundAuthWarn bool
	for _, r := range results {
		if r.Category == "Auth" && r.Status == "WARN" {
			foundAuthWarn = true
		}
	}
	if !foundAuthWarn {
		t.Errorf("expected Auth warning when credentials missing")
	}

	// Add auth.json and re-run Doctor
	authFile := filepath.Join(profileDir, ".codex", "auth.json")
	_ = os.WriteFile(authFile, []byte(`{"tokens":{"access_token":"mock"}}`), 0600)

	resultsAuth := a.Doctor(context.Background(), "test", profileDir)
	var foundAuthOK bool
	for _, r := range resultsAuth {
		if r.Category == "Auth" && r.Status == "OK" {
			foundAuthOK = true
		}
	}
	if !foundAuthOK {
		t.Errorf("expected Auth OK when credentials exist")
	}
}

func TestCodexAdapter_GetUsage(t *testing.T) {
	a := NewAdapter()
	tmpDir := t.TempDir()
	profileDir := filepath.Join(tmpDir, "profiles", "usage-test")

	// No credentials
	rep, err := a.GetUsage(context.Background(), "usage-test", profileDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rep.Status != usage.StatusUnknown {
		t.Errorf("expected StatusUnknown, got %v", rep.Status)
	}

	// Create auth.json with JWT
	codexDir := filepath.Join(profileDir, ".codex")
	_ = os.MkdirAll(codexDir, 0755)

	claimsJSON := `{"email":"codex@example.com","name":"Codex Pro User","https://api.openai.com/auth":{"chatgpt_plan_type":"pro"}}`
	encodedPayload := base64.RawURLEncoding.EncodeToString([]byte(claimsJSON))
	mockJWT := "eyJhbGciOiJSUzI1NiJ9." + encodedPayload + ".sig"
	authJSON := `{"tokens":{"id_token":"` + mockJWT + `"}}`
	_ = os.WriteFile(filepath.Join(codexDir, "auth.json"), []byte(authJSON), 0600)

	// Create sessions dir with rate limit jsonl
	sessionsDir := filepath.Join(codexDir, "sessions")
	_ = os.MkdirAll(sessionsDir, 0755)
	sessionLine := `{"rateLimits":{"secondary":{"used_percent":25}}}` + "\n"
	_ = os.WriteFile(filepath.Join(sessionsDir, "session1.jsonl"), []byte(sessionLine), 0644)

	repWithAuth, err := a.GetUsage(context.Background(), "usage-test", profileDir)
	if err != nil {
		t.Fatalf("unexpected error with auth: %v", err)
	}
	if repWithAuth.Status != usage.StatusOK {
		t.Errorf("expected StatusOK, got %v", repWithAuth.Status)
	}
	if repWithAuth.AccountEmail != "codex@example.com" {
		t.Errorf("expected email 'codex@example.com', got %q", repWithAuth.AccountEmail)
	}
	if repWithAuth.AuthMethod != "ChatGPT Pro" {
		t.Errorf("expected AuthMethod 'ChatGPT Pro', got %q", repWithAuth.AuthMethod)
	}
	if len(repWithAuth.Windows) == 0 {
		t.Errorf("expected windows to be populated, got 0")
	} else {
		ww := repWithAuth.WeeklyWindow()
		if ww == nil {
			t.Errorf("expected WeeklyWindow not to be nil")
		} else if ww.RemainingPct != 75 {
			t.Errorf("expected weekly remaining 75%%, got %d%%", ww.RemainingPct)
		}
	}

	// Test real nested date-partitioned session rollout structure:
	// sessions/YYYY/MM/DD/rollout-*.jsonl with payload.rate_limits
	nestedDir := filepath.Join(sessionsDir, "2026", "09", "13")
	_ = os.MkdirAll(nestedDir, 0755)
	futureReset := time.Now().Add(2 * time.Hour).Unix()
	futureWeeklyReset := time.Now().Add(7 * 24 * time.Hour).Unix()
	codexDefaultEvent := fmt.Sprintf(`{"type":"event_msg","payload":{"type":"token_count","rate_limits":{"limit_id":"codex","limit_name":null,"primary":{"used_percent":21.0,"window_minutes":10080,"resets_at":%d}}}}`+"\n", futureWeeklyReset)
	sparkEvent := fmt.Sprintf(`{"type":"event_msg","payload":{"type":"token_count","rate_limits":{"limit_id":"codex_spark","limit_name":"GPT-5.3-Codex-Spark","primary":{"used_percent":10.5,"window_minutes":300,"resets_at":%d},"secondary":{"used_percent":5.0,"window_minutes":10080,"resets_at":%d},"credits":{"has_credits":true,"unlimited":false,"balance":"$15.00"}}}}`+"\n", futureReset, futureWeeklyReset)
	_ = os.WriteFile(filepath.Join(nestedDir, "rollout-2026-09-13T13-05-43-test.jsonl"), []byte(codexDefaultEvent+sparkEvent), 0644)

	repNested, err := a.GetUsage(context.Background(), "usage-test", profileDir)
	if err != nil {
		t.Fatalf("unexpected error with nested session: %v", err)
	}
	if len(repNested.Windows) != 3 {
		t.Fatalf("expected 3 windows (Codex weekly, Spark 5h, Spark weekly), got %d: %+v", len(repNested.Windows), repNested.Windows)
	}

	// Verify Codex default weekly window
	var codexWeekly, spark5h, sparkWeekly *usage.LimitWindow
	for i := range repNested.Windows {
		w := &repNested.Windows[i]
		if w.Category == "Codex" && w.Name == "Weekly Limit" {
			codexWeekly = w
		} else if w.Category == "GPT-5.3-Codex-Spark" && w.Name == "5h Limit" {
			spark5h = w
		} else if w.Category == "GPT-5.3-Codex-Spark" && w.Name == "Weekly Limit" {
			sparkWeekly = w
		}
	}

	if codexWeekly == nil || codexWeekly.RemainingPct != 79 {
		t.Errorf("expected Codex weekly remaining 79%%, got %v", codexWeekly)
	}
	if spark5h == nil || spark5h.RemainingPct != 90 {
		t.Errorf("expected Spark 5h remaining 90%%, got %v", spark5h)
	}
	if sparkWeekly == nil || sparkWeekly.RemainingPct != 95 {
		t.Errorf("expected Spark weekly remaining 95%%, got %v", sparkWeekly)
	}
	if repNested.Credits != "$15.00" {
		t.Errorf("expected credits '$15.00', got %q", repNested.Credits)
	}
	_ = futureReset
}

func TestCodexAdapter_BridgePluginsAndHooks(t *testing.T) {
	mockHome := t.TempDir()
	t.Setenv("HOME", mockHome)

	mockHostCodex := filepath.Join(mockHome, ".codex")
	mockPlugins := filepath.Join(mockHostCodex, "plugins")
	mockHooksDir := filepath.Join(mockHostCodex, "hooks")
	mockHooksJSON := filepath.Join(mockHostCodex, "hooks.json")

	_ = os.MkdirAll(mockPlugins, 0755)
	_ = os.MkdirAll(mockHooksDir, 0755)
	_ = os.WriteFile(mockHooksJSON, []byte(`{"hooks":{}}`), 0644)

	profileCodexDir := filepath.Join(t.TempDir(), "profile", ".codex")
	_ = os.MkdirAll(profileCodexDir, 0700)

	bridgePluginsAndHooks(mockHome, profileCodexDir)

	// Check plugins symlink
	targetPlugins := filepath.Join(profileCodexDir, "plugins")
	fi, err := os.Lstat(targetPlugins)
	if err != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Errorf("expected plugins to be symlink, err=%v", err)
	}

	// Check hooks dir symlink
	targetHooks := filepath.Join(profileCodexDir, "hooks")
	fi, err = os.Lstat(targetHooks)
	if err != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Errorf("expected hooks to be symlink, err=%v", err)
	}

	// Check hooks.json symlink
	targetJSON := filepath.Join(profileCodexDir, "hooks.json")
	fi, err = os.Lstat(targetJSON)
	if err != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Errorf("expected hooks.json to be symlink, err=%v", err)
	}
}

func TestCodexAdapter_BridgeCxStatusline(t *testing.T) {
	mockHome := t.TempDir()
	hostCxDir := filepath.Join(mockHome, ".config", "cxstatusline")
	if err := os.MkdirAll(hostCxDir, 0755); err != nil {
		t.Fatalf("mkdir host cxstatusline failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(hostCxDir, "settings.json"), []byte(`{"lines":[]}`), 0644); err != nil {
		t.Fatalf("write host settings failed: %v", err)
	}

	t.Run("fresh profile", func(t *testing.T) {
		profileDir := t.TempDir()
		bridgeCxStatusline(mockHome, profileDir)

		dest := filepath.Join(profileDir, ".config", "cxstatusline")
		fi, err := os.Lstat(dest)
		if err != nil {
			t.Fatalf("expected bridged cxstatusline to exist: %v", err)
		}
		if fi.Mode()&os.ModeSymlink == 0 {
			t.Errorf("expected bridged cxstatusline to be a symlink")
		}
	})

	t.Run("migrate auto-generated default settings directory", func(t *testing.T) {
		profileDir := t.TempDir()
		dest := filepath.Join(profileDir, ".config", "cxstatusline")
		if err := os.MkdirAll(dest, 0755); err != nil {
			t.Fatalf("mkdir dest failed: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dest, "settings.json"), []byte(`{"default":true}`), 0644); err != nil {
			t.Fatalf("write fallback settings failed: %v", err)
		}

		bridgeCxStatusline(mockHome, profileDir)

		fi, err := os.Lstat(dest)
		if err != nil {
			t.Fatalf("expected bridged cxstatusline to exist after migration: %v", err)
		}
		if fi.Mode()&os.ModeSymlink == 0 {
			t.Errorf("expected auto-generated directory to be replaced with symlink")
		}
	})

	t.Run("preserve custom profile directory with extra files", func(t *testing.T) {
		profileDir := t.TempDir()
		dest := filepath.Join(profileDir, ".config", "cxstatusline")
		if err := os.MkdirAll(dest, 0755); err != nil {
			t.Fatalf("mkdir dest failed: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dest, "settings.json"), []byte(`{}`), 0644); err != nil {
			t.Fatalf("write settings failed: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dest, "custom.theme"), []byte("theme"), 0644); err != nil {
			t.Fatalf("write custom theme failed: %v", err)
		}

		bridgeCxStatusline(mockHome, profileDir)

		fi, err := os.Lstat(dest)
		if err != nil {
			t.Fatalf("expected custom cxstatusline to exist: %v", err)
		}
		if fi.Mode()&os.ModeSymlink != 0 {
			t.Errorf("expected custom directory with extra files to NOT be replaced with symlink")
		}
	})

	t.Run("noop when host has no cxstatusline", func(t *testing.T) {
		emptyHome := t.TempDir()
		profileDir := t.TempDir()

		bridgeCxStatusline(emptyHome, profileDir)

		dest := filepath.Join(profileDir, ".config", "cxstatusline")
		if _, err := os.Lstat(dest); !os.IsNotExist(err) {
			t.Errorf("expected no cxstatusline bridged when host has none")
		}
	})
}

func TestCodexAdapter_CopyHostConfig_HookTrustPathRewriting(t *testing.T) {
	mockHome := t.TempDir()
	hostCodexDir := filepath.Join(mockHome, ".codex")
	if err := os.MkdirAll(hostCodexDir, 0755); err != nil {
		t.Fatalf("mkdir host codex failed: %v", err)
	}

	hostHooksJSON := filepath.Join(hostCodexDir, "hooks.json")
	hostConfigPath := filepath.Join(hostCodexDir, "config.toml")
	hostConfigContent := fmt.Sprintf(`model_provider = "caveman"
[hooks.state."%s:pre_tool_use:0:0"]
trusted_hash = "sha256:abc12345"
`, hostHooksJSON)

	if err := os.WriteFile(hostConfigPath, []byte(hostConfigContent), 0644); err != nil {
		t.Fatalf("write host config failed: %v", err)
	}

	t.Run("fresh copy rewrites hook trust paths", func(t *testing.T) {
		profileCodexDir := filepath.Join(t.TempDir(), ".codex")
		if err := os.MkdirAll(profileCodexDir, 0700); err != nil {
			t.Fatalf("mkdir profile codex failed: %v", err)
		}

		copyHostConfig(mockHome, profileCodexDir)

		destConfig := filepath.Join(profileCodexDir, "config.toml")
		data, err := os.ReadFile(destConfig)
		if err != nil {
			t.Fatalf("read dest config failed: %v", err)
		}

		expectedDestHooksJSON := filepath.Join(profileCodexDir, "hooks.json")
		if strings.Contains(string(data), hostHooksJSON) {
			t.Errorf("expected hostHooksJSON to be rewritten, found in config: %s", string(data))
		}
		if !strings.Contains(string(data), expectedDestHooksJSON) {
			t.Errorf("expected destHooksJSON %q in config: %s", expectedDestHooksJSON, string(data))
		}
	})

	t.Run("migrates existing config with old host hook paths", func(t *testing.T) {
		profileCodexDir := filepath.Join(t.TempDir(), ".codex")
		if err := os.MkdirAll(profileCodexDir, 0700); err != nil {
			t.Fatalf("mkdir profile codex failed: %v", err)
		}

		destConfig := filepath.Join(profileCodexDir, "config.toml")
		// Write unmigrated config directly
		if err := os.WriteFile(destConfig, []byte(hostConfigContent), 0644); err != nil {
			t.Fatalf("write dest config failed: %v", err)
		}

		copyHostConfig(mockHome, profileCodexDir)

		data, err := os.ReadFile(destConfig)
		if err != nil {
			t.Fatalf("read dest config failed: %v", err)
		}

		expectedDestHooksJSON := filepath.Join(profileCodexDir, "hooks.json")
		if strings.Contains(string(data), hostHooksJSON) {
			t.Errorf("expected existing hostHooksJSON to be migrated, still found: %s", string(data))
		}
		if !strings.Contains(string(data), expectedDestHooksJSON) {
			t.Errorf("expected destHooksJSON %q in migrated config: %s", expectedDestHooksJSON, string(data))
		}
	})
}

func TestCodexAdapter_EnsureSidecarDaemons(t *testing.T) {
	mockHome := t.TempDir()
	profileCodexDir := filepath.Join(t.TempDir(), ".codex")
	_ = os.MkdirAll(profileCodexDir, 0700)

	t.Run("noop when no caveman proxy configured", func(t *testing.T) {
		cfgPath := filepath.Join(profileCodexDir, "config.toml")
		_ = os.WriteFile(cfgPath, []byte("model = \"gpt-4\"\n"), 0644)

		// Should not error or panic
		ensureSidecarDaemons(mockHome, profileCodexDir)
	})

	t.Run("handles missing binary gracefully when caveman is configured", func(t *testing.T) {
		cfgPath := filepath.Join(profileCodexDir, "config.toml")
		_ = os.WriteFile(cfgPath, []byte("model_provider = \"caveman\"\nbase_url = \"http://127.0.0.1:8787/chatgpt\"\n"), 0644)

		// mockHome has no caveman-proxy binary; should log debug and not panic
		ensureSidecarDaemons(mockHome, profileCodexDir)
	})
}

func TestCodexAdapter_Doctor_SidecarAndHooks(t *testing.T) {
	adapter := NewAdapter()
	profileDir := t.TempDir()
	profileCodexDir := filepath.Join(profileDir, ".codex")
	_ = os.MkdirAll(profileCodexDir, 0700)

	cfgPath := filepath.Join(profileCodexDir, "config.toml")
	cfgContent := `model_provider = "caveman"
[hooks.state."/fake/host/.codex/hooks.json:pre_tool_use:0:0"]
trusted_hash = "sha256:123"
`
	_ = os.WriteFile(cfgPath, []byte(cfgContent), 0644)

	results := adapter.Doctor(context.Background(), "testprof", profileDir)

	hasSidecar := false
	for _, r := range results {
		if r.Category == "Sidecar" {
			hasSidecar = true
			if r.Status != "OK" && r.Status != "WARN" {
				t.Errorf("unexpected sidecar status: %s", r.Status)
			}
		}
	}
	if !hasSidecar {
		t.Errorf("expected Sidecar check in Doctor results")
	}
}

func TestDeduplicateTomlTables(t *testing.T) {
	input := `model_provider = "caveman"
base_url = "http://127.0.0.1:8787"

[hooks.state."/path/to/hooks.json:pre_tool_use:0:0"]
trusted_hash = "sha256:111"

[desktop]
followUpQueueMode = "steer"

[hooks.state."/path/to/hooks.json:pre_tool_use:0:0"]
trusted_hash = "sha256:222"

[[array_table]]
key = "v1"

[[array_table]]
key = "v2"
`

	cleaned, dupes := deduplicateTomlTables(input)
	if dupes != 1 {
		t.Fatalf("expected 1 duplicate removed, got %d", dupes)
	}

	if strings.Count(cleaned, `[hooks.state."/path/to/hooks.json:pre_tool_use:0:0"]`) != 1 {
		t.Errorf("expected exactly 1 instance of the table header, got:\n%s", cleaned)
	}

	if !strings.Contains(cleaned, `trusted_hash = "sha256:111"`) {
		t.Errorf("expected first instance to be preserved, got:\n%s", cleaned)
	}

	if strings.Contains(cleaned, `trusted_hash = "sha256:222"`) {
		t.Errorf("expected second duplicate instance to be removed, got:\n%s", cleaned)
	}

	if strings.Count(cleaned, "[[array_table]]") != 2 {
		t.Errorf("expected array of tables to be preserved, got:\n%s", cleaned)
	}
}

func TestCodexAdapter_Doctor_DuplicateTomlRepair(t *testing.T) {
	adapter := NewAdapter()
	profileDir := t.TempDir()
	profileCodexDir := filepath.Join(profileDir, ".codex")
	_ = os.MkdirAll(profileCodexDir, 0700)

	cfgPath := filepath.Join(profileCodexDir, "config.toml")
	cfgContent := `[hooks.state."/path/to/hooks.json:pre_tool_use:0:0"]
trusted_hash = "sha256:111"

[hooks.state."/path/to/hooks.json:pre_tool_use:0:0"]
trusted_hash = "sha256:222"
`
	_ = os.WriteFile(cfgPath, []byte(cfgContent), 0644)

	results := adapter.Doctor(context.Background(), "testprof", profileDir)

	hasConfigRepair := false
	for _, r := range results {
		if r.Category == "Config" && strings.Contains(r.Message, "duplicate table key") {
			hasConfigRepair = true
			if r.Status != "OK" {
				t.Errorf("expected Config repair status OK, got %s", r.Status)
			}
		}
	}
	if !hasConfigRepair {
		t.Errorf("expected Config duplicate repair in Doctor results")
	}

	// Verify file was repaired
	repairedData, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("failed to read repaired config: %v", err)
	}
	if strings.Count(string(repairedData), `[hooks.state."/path/to/hooks.json:pre_tool_use:0:0"]`) != 1 {
		t.Errorf("expected file to be deduplicated on disk, got:\n%s", string(repairedData))
	}
}
