package profile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
)

type mockAdapter struct {
	name      string
	tokenPath string
}

func (m *mockAdapter) Name() string                                         { return m.name }
func (m *mockAdapter) DisplayName() string                                  { return m.name }
func (m *mockAdapter) Aliases() []string                                    { return nil }
func (m *mockAdapter) BinaryName() string                                   { return m.name }
func (m *mockAdapter) HasCredentials(p string) bool {
	if m.tokenPath != "" {
		fi, err := os.Stat(filepath.Join(p, m.tokenPath))
		return err == nil && !fi.IsDir() && fi.Size() > 0
	}
	return false
}
func (m *mockAdapter) TokenPath(p string) string                            { return filepath.Join(p, m.tokenPath) }
func (m *mockAdapter) Login(ctx any, p, d string) error                     { return nil }
func (m *mockAdapter) PrepareEnv(p, d string) (agents.LaunchEnv, error)     { return agents.LaunchEnv{}, nil }
func (m *mockAdapter) Doctor(ctx any, p, d string) []agents.DiagnosticResult { return nil }
func (m *mockAdapter) GetUsage(ctx any, p, d string) (any, error)           { return nil, nil }

func TestMoveAgent_Validation(t *testing.T) {
	tempDir := t.TempDir()
	pm := NewProfileManager(tempDir)
	cfg := config.NewDefaultConfig()

	// Empty source
	if err := pm.MoveAgent("codex", "", "work", false, cfg, nil); err == nil {
		t.Error("expected error for empty source profile, got nil")
	}

	// Empty target
	if err := pm.MoveAgent("codex", "rs", "", false, cfg, nil); err == nil {
		t.Error("expected error for empty target profile, got nil")
	}

	// Same source and target
	if err := pm.MoveAgent("codex", "rs", "rs", false, cfg, nil); err == nil {
		t.Error("expected error when source == target, got nil")
	}

	// Non-existent source
	if err := pm.MoveAgent("codex", "non-existent", "work", false, cfg, nil); err == nil {
		t.Error("expected error for non-existent source, got nil")
	}

	// Source exists but agent not associated
	_, _ = pm.EnsureProfile("rs")
	if err := pm.MoveAgent("codex", "rs", "work", false, cfg, nil); err == nil {
		t.Error("expected error when agent not associated, got nil")
	}
}

func TestMoveAgent_MultiAgentPreservation(t *testing.T) {
	tempDir := t.TempDir()
	pm := NewProfileManager(tempDir)
	cfg := config.NewDefaultConfig()

	// Setup rs with agy and codex
	rsDir, err := pm.EnsureProfile("rs")
	if err != nil {
		t.Fatalf("EnsureProfile rs error: %v", err)
	}
	cfg.AddProfileAgent("rs", "agy")
	cfg.AddProfileAgent("rs", "codex")

	// Create codex data in rs
	rsCodexDir := filepath.Join(rsDir, ".codex")
	if err := os.MkdirAll(rsCodexDir, 0700); err != nil {
		t.Fatalf("mkdir rsCodexDir error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rsCodexDir, "auth.json"), []byte(`{"token":"codex-secret"}`), 0600); err != nil {
		t.Fatalf("write auth.json error: %v", err)
	}

	// Create agy data in rs
	rsAgyDir := filepath.Join(rsDir, ".gemini", "antigravity-cli")
	if err := os.MkdirAll(rsAgyDir, 0700); err != nil {
		t.Fatalf("mkdir rsAgyDir error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rsAgyDir, "antigravity-oauth-token"), []byte(`{"token":"agy-secret"}`), 0600); err != nil {
		t.Fatalf("write agy token error: %v", err)
	}

	// Move codex from rs to work
	if err := pm.MoveAgent("codex", "rs", "work", false, cfg, nil); err != nil {
		t.Fatalf("MoveAgent failed: %v", err)
	}

	// Verify rs still exists and has agy
	if !cfg.HasAgent("rs", "agy") {
		t.Error("expected rs to still have agy in config")
	}
	if cfg.HasAgent("rs", "codex") {
		t.Error("expected rs to no longer have codex in config")
	}
	if _, err := os.Stat(filepath.Join(rsDir, ".gemini", "antigravity-cli", "antigravity-oauth-token")); err != nil {
		t.Error("expected rs to keep agy token on disk")
	}
	if _, err := os.Stat(filepath.Join(rsDir, ".codex")); !os.IsNotExist(err) {
		t.Error("expected rs to no longer have .codex directory")
	}

	// Verify work has codex
	workDir := pm.ProfileDir("work")
	if !cfg.HasAgent("work", "codex") {
		t.Error("expected work to have codex in config")
	}
	authData, err := os.ReadFile(filepath.Join(workDir, ".codex", "auth.json"))
	if err != nil {
		t.Fatalf("expected work to have .codex/auth.json: %v", err)
	}
	if string(authData) != `{"token":"codex-secret"}` {
		t.Errorf("expected moved codex auth data, got %s", string(authData))
	}
}

func TestMoveAgent_CollisionAndForce(t *testing.T) {
	tempDir := t.TempDir()
	pm := NewProfileManager(tempDir)
	cfg := config.NewDefaultConfig()

	rsDir, _ := pm.EnsureProfile("rs")
	cfg.AddProfileAgent("rs", "codex")
	rsCodexDir := filepath.Join(rsDir, ".codex")
	_ = os.MkdirAll(rsCodexDir, 0700)
	_ = os.WriteFile(filepath.Join(rsCodexDir, "auth.json"), []byte("rs-token"), 0600)

	workDir, _ := pm.EnsureProfile("work")
	cfg.AddProfileAgent("work", "codex")
	workCodexDir := filepath.Join(workDir, ".codex")
	_ = os.MkdirAll(workCodexDir, 0700)
	_ = os.WriteFile(filepath.Join(workCodexDir, "auth.json"), []byte("work-token"), 0600)

	// Attempt move without force should fail
	err := pm.MoveAgent("codex", "rs", "work", false, cfg, nil)
	if err == nil {
		t.Fatal("expected collision error without force, got nil")
	}

	// Move with force should overwrite
	err = pm.MoveAgent("codex", "rs", "work", true, cfg, nil)
	if err != nil {
		t.Fatalf("expected move with force to succeed, got %v", err)
	}

	overwritten, _ := os.ReadFile(filepath.Join(workDir, ".codex", "auth.json"))
	if string(overwritten) != "rs-token" {
		t.Errorf("expected 'rs-token' in target, got %s", string(overwritten))
	}
}

func TestMoveAgent_SingleAgentSourceCleanup(t *testing.T) {
	tempDir := t.TempDir()
	pm := NewProfileManager(tempDir)
	cfg := config.NewDefaultConfig()

	srcDir, _ := pm.EnsureProfile("temp-src")
	cfg.AddProfileAgent("temp-src", "codex")
	_ = os.MkdirAll(filepath.Join(srcDir, ".codex"), 0700)
	_ = os.WriteFile(filepath.Join(srcDir, ".codex", "auth.json"), []byte("tok"), 0600)

	if err := pm.MoveAgent("codex", "temp-src", "dest-prof", false, cfg, nil); err != nil {
		t.Fatalf("MoveAgent failed: %v", err)
	}

	// temp-src had 0 remaining agents, should be cleaned up
	if cfg.Profiles != nil {
		if _, ok := cfg.Profiles["temp-src"]; ok {
			t.Error("expected temp-src to be removed from config")
		}
	}
	if _, err := os.Stat(srcDir); !os.IsNotExist(err) {
		t.Error("expected temp-src directory to be removed on disk")
	}

	// dest-prof should have codex
	if !cfg.HasAgent("dest-prof", "codex") {
		t.Error("expected dest-prof to have codex")
	}
}
