package codex

import (
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

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
	sessionLine := `{"rateLimits":{"primary":{"used_percent":25}}}` + "\n"
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
		pw := repWithAuth.PrimaryWindow()
		if pw == nil {
			t.Errorf("expected PrimaryWindow not to be nil")
		} else if pw.RemainingPct != 75 {
			t.Errorf("expected primary remaining 75%%, got %d%%", pw.RemainingPct)
		}
	}

	// Test real nested date-partitioned session rollout structure:
	// sessions/YYYY/MM/DD/rollout-*.jsonl with payload.rate_limits
	nestedDir := filepath.Join(sessionsDir, "2026", "09", "13")
	_ = os.MkdirAll(nestedDir, 0755)
	futureReset := 1789316216
	realCodexEvent := `{"type":"event_msg","payload":{"type":"token_count","rate_limits":{"limit_id":"codex_spark","limit_name":"GPT-5.3-Codex-Spark","primary":{"used_percent":10.5,"window_minutes":300,"resets_at":` +
		"1789316216" + `},"secondary":{"used_percent":5.0,"window_minutes":10080,"resets_at":` +
		"1789903016" + `},"credits":{"has_credits":true,"unlimited":false,"balance":"$15.00"}}}}` + "\n"
	_ = os.WriteFile(filepath.Join(nestedDir, "rollout-2026-09-13T13-05-43-test.jsonl"), []byte(realCodexEvent), 0644)

	repNested, err := a.GetUsage(context.Background(), "usage-test", profileDir)
	if err != nil {
		t.Fatalf("unexpected error with nested session: %v", err)
	}
	if len(repNested.Windows) != 2 {
		t.Fatalf("expected 2 windows (primary and secondary), got %d", len(repNested.Windows))
	}
	pw := repNested.PrimaryWindow()
	if pw == nil {
		t.Fatalf("expected primary window, got nil")
	}
	if pw.RemainingPct != 90 { // 100 - round(10.5) = 100 - 10 = 90
		t.Errorf("expected primary remaining 90%%, got %d%%", pw.RemainingPct)
	}
	ww := repNested.WeeklyWindow()
	if ww == nil {
		t.Fatalf("expected weekly window, got nil")
	}
	if ww.RemainingPct != 95 { // 100 - round(5.0) = 95
		t.Errorf("expected weekly remaining 95%%, got %d%%", ww.RemainingPct)
	}
	if repNested.Credits != "$15.00" {
		t.Errorf("expected credits '$15.00', got %q", repNested.Credits)
	}
	_ = futureReset
}
