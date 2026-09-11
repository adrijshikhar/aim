package gemini

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aim-cli/aim/internal/usage"
)

func TestGeminiMetadata(t *testing.T) {
	adapter := NewAdapter()
	if adapter.Name() != "gemini" {
		t.Errorf("expected Name to be 'gemini', got '%s'", adapter.Name())
	}
	if adapter.DisplayName() != "Gemini CLI" {
		t.Errorf("expected DisplayName to be 'Gemini CLI', got '%s'", adapter.DisplayName())
	}
	if len(adapter.Aliases()) != 1 || adapter.Aliases()[0] != "gemini-cli" {
		t.Errorf("expected Aliases to contain 'gemini-cli', got %v", adapter.Aliases())
	}
	if adapter.BinaryName() != "gemini" {
		t.Errorf("expected BinaryName to be 'gemini', got '%s'", adapter.BinaryName())
	}
}

func TestGeminiLogin(t *testing.T) {
	adapter := NewAdapter()
	err := adapter.Login(context.Background(), "work", "/fake/dir")
	if err == nil {
		t.Fatalf("expected login error for unready gemini login, got nil")
	}
}

func TestGeminiPrepareEnv(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aim-gemini-test-*")
	if err != nil {
		t.Fatalf("temp dir error: %v", err)
	}
	defer os.RemoveAll(tempDir)

	adapter := NewAdapter()
	launchEnv, err := adapter.PrepareEnv("testprof", tempDir)
	if err != nil {
		t.Fatalf("PrepareEnv failed: %v", err)
	}

	if launchEnv.Env["HOME"] != tempDir {
		t.Errorf("expected HOME to be %s, got %s", tempDir, launchEnv.Env["HOME"])
	}
	expectedGeminiHome := filepath.Join(tempDir, ".gemini")
	if launchEnv.Env["GEMINI_CLI_HOME"] != expectedGeminiHome {
		t.Errorf("expected GEMINI_CLI_HOME to be %s, got %s", expectedGeminiHome, launchEnv.Env["GEMINI_CLI_HOME"])
	}
	if launchEnv.Env["AIM_AGENT"] != "gemini" {
		t.Errorf("expected AIM_AGENT to be 'gemini', got '%s'", launchEnv.Env["AIM_AGENT"])
	}
	if launchEnv.Env["AIM_PROFILE"] != "testprof" {
		t.Errorf("expected AIM_PROFILE to be 'testprof', got '%s'", launchEnv.Env["AIM_PROFILE"])
	}
	if launchEnv.BinaryPath == "" {
		t.Errorf("expected BinaryPath to be non-empty")
	}
}

func TestGeminiPrepareEnv_BinaryFallback(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("PATH", "")

	adapter := NewAdapter()
	launchEnv, err := adapter.PrepareEnv("testprof", tempDir)
	if err != nil {
		t.Fatalf("PrepareEnv failed: %v", err)
	}

	if launchEnv.BinaryPath != "gemini" {
		t.Errorf("expected BinaryPath fallback to be 'gemini', got '%s'", launchEnv.BinaryPath)
	}
}

func TestGeminiDoctor(t *testing.T) {
	adapter := NewAdapter()
	results := adapter.Doctor(context.Background(), "testprof", "/fake/dir")
	if len(results) == 0 {
		t.Fatalf("expected doctor results, got empty")
	}
	if results[0].Status != "OK" {
		t.Errorf("expected status OK, got %s", results[0].Status)
	}
}

func TestGeminiTokenPath(t *testing.T) {
	adapter := NewGeminiAdapter()
	expected := filepath.Join("/fake/profile", ".gemini", "gemini-oauth-token")
	if got := adapter.TokenPath("/fake/profile"); got != expected {
		t.Errorf("expected TokenPath %s, got %s", expected, got)
	}
}

func TestGeminiAdapter_HasCredentials(t *testing.T) {
	adapter := NewGeminiAdapter()
	tmpDir := t.TempDir()

	// 1. Not present
	if adapter.HasCredentials(tmpDir) {
		t.Errorf("expected HasCredentials to be false on empty profile")
	}

	// 2. Present but empty file
	tokenPath := adapter.TokenPath(tmpDir)
	_ = os.MkdirAll(filepath.Dir(tokenPath), 0755)
	_ = os.WriteFile(tokenPath, []byte(""), 0600)
	if adapter.HasCredentials(tmpDir) {
		t.Errorf("expected HasCredentials to be false on empty token file")
	}

	// 3. Valid token file
	_ = os.WriteFile(tokenPath, []byte(`{"access_token":"valid"}`), 0600)
	if !adapter.HasCredentials(tmpDir) {
		t.Errorf("expected HasCredentials to be true on valid token file")
	}

	// 4. Directory instead of file
	dirProfile := t.TempDir()
	dirTokenPath := adapter.TokenPath(dirProfile)
	_ = os.MkdirAll(dirTokenPath, 0755)
	if adapter.HasCredentials(dirProfile) {
		t.Errorf("expected HasCredentials to be false when token path is a directory")
	}
}

func TestGeminiAdapterGetUsage(t *testing.T) {
	adapter := NewAdapter()
	tmpDir, err := os.MkdirTemp("", "aim-gemini-usage-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	rep, err := adapter.GetUsage(context.Background(), "test_prof", tmpDir)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if rep.Status != usage.StatusUnknown {
		t.Errorf("expected StatusUnknown for profile without token, got %s", rep.Status)
	}
}

func TestGeminiAdapterGetUsageWithCredentials(t *testing.T) {
	adapter := NewAdapter()
	tmpDir := t.TempDir()

	tokenPath := adapter.TokenPath(tmpDir)
	_ = os.MkdirAll(filepath.Dir(tokenPath), 0700)
	_ = os.WriteFile(tokenPath, []byte("valid-token"), 0600)

	rep, err := adapter.GetUsage(context.Background(), "work", tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rep.Status != usage.StatusOK {
		t.Errorf("expected StatusOK, got %s", rep.Status)
	}
	if rep.Summary != "Active credentials" {
		t.Errorf("expected 'Active credentials', got '%s'", rep.Summary)
	}
}
