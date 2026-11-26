package cursor

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCursorAdapter_Identity(t *testing.T) {
	a := NewAdapter()
	if a.Name() != "cursor" {
		t.Errorf("expected name 'cursor', got %q", a.Name())
	}
	if a.DisplayName() != "Cursor" {
		t.Errorf("expected displayName 'Cursor', got %q", a.DisplayName())
	}
	if a.BinaryName() != "cursor" {
		t.Errorf("expected binaryName 'cursor', got %q", a.BinaryName())
	}
	if len(a.Aliases()) == 0 || a.Aliases()[0] != "cur" {
		t.Errorf("expected alias 'cur', got %v", a.Aliases())
	}
}

func TestCursorAdapter_HasCredentials(t *testing.T) {
	a := NewAdapter()
	tempDir := t.TempDir()

	// 1. Without credentials
	t.Setenv("CURSOR_API_KEY", "")
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")
	if a.HasCredentials(tempDir) {
		t.Errorf("expected HasCredentials false for empty directory")
	}

	// 2. Ambient CURSOR_API_KEY env does NOT establish profile-scoped credentials
	t.Setenv("CURSOR_API_KEY", "test-key-123")
	if a.HasCredentials(tempDir) {
		t.Errorf("expected HasCredentials false when only ambient CURSOR_API_KEY is set")
	}
	t.Setenv("CURSOR_API_KEY", "")

	// 3. With profile state.vscdb
	dbDir := filepath.Join(tempDir, ".cursor", "user-data", "User", "globalStorage")
	_ = os.MkdirAll(dbDir, 0700)
	_ = os.WriteFile(filepath.Join(dbDir, "state.vscdb"), []byte("sqlite data"), 0600)
	if !a.HasCredentials(tempDir) {
		t.Errorf("expected HasCredentials true with state.vscdb")
	}
}

func TestCursorAdapter_PrepareEnv(t *testing.T) {
	a := NewAdapter()
	tempDir := t.TempDir()

	// Mock binary in a temp directory
	mockBin := filepath.Join(tempDir, "cursor")
	_ = os.WriteFile(mockBin, []byte("#!/bin/sh\nexit 0\n"), 0755)

	origPath := os.Getenv("PATH")
	t.Setenv("PATH", tempDir+":"+origPath)

	ctx := context.Background()
	env, err := a.PrepareEnv(ctx, "dev", tempDir)
	if err != nil {
		t.Fatalf("PrepareEnv failed: %v", err)
	}

	if env.BinaryPath == "" {
		t.Errorf("expected non-empty BinaryPath")
	}

	// Verify --user-data-dir and --extensions-dir args
	hasUserData := false
	hasExtensions := false
	for i, arg := range env.Args {
		if arg == "--user-data-dir" && i+1 < len(env.Args) {
			if strings.Contains(env.Args[i+1], filepath.Join(tempDir, ".cursor", "user-data")) {
				hasUserData = true
			}
		}
		if arg == "--extensions-dir" && i+1 < len(env.Args) {
			if strings.Contains(env.Args[i+1], filepath.Join(tempDir, ".cursor", "extensions")) {
				hasExtensions = true
			}
		}
	}

	if !hasUserData {
		t.Errorf("expected --user-data-dir pointing to profile in args: %v", env.Args)
	}
	if !hasExtensions {
		t.Errorf("expected --extensions-dir pointing to profile in args: %v", env.Args)
	}

	// Verify dirs were created
	if fi, err := os.Stat(filepath.Join(tempDir, ".cursor", "user-data")); err != nil || !fi.IsDir() {
		t.Errorf("expected user-data dir created")
	}
	if fi, err := os.Stat(filepath.Join(tempDir, ".cursor", "extensions")); err != nil || !fi.IsDir() {
		t.Errorf("expected extensions dir created")
	}
}

func TestCursorAdapter_Doctor(t *testing.T) {
	a := NewAdapter()
	tempDir := t.TempDir()

	// Create a profile mcp.json
	cursorDir := filepath.Join(tempDir, ".cursor")
	_ = os.MkdirAll(cursorDir, 0700)
	_ = os.WriteFile(filepath.Join(cursorDir, "mcp.json"), []byte(`{"mcpServers":{}}`), 0600)

	results := a.Doctor(context.Background(), "dev", tempDir)
	foundConfig := false
	for _, res := range results {
		if res.Category == "Config" && res.Status == "OK" {
			foundConfig = true
			break
		}
	}
	if !foundConfig {
		t.Errorf("expected OK for Config in Doctor results: %+v", results)
	}
}
