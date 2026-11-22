package mcp

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestClientDrivers_Registry(t *testing.T) {
	drivers := AllDrivers()
	if len(drivers) != 4 {
		t.Fatalf("expected 4 drivers (claude, cursor, claude-desktop, codex-desktop), got %d", len(drivers))
	}

	for _, id := range []ClientID{ClientClaudeCode, ClientCursor, ClientClaudeDesktop, ClientCodexDesktop} {
		d, ok := GetDriver(id)
		if !ok || d == nil {
			t.Errorf("expected driver for %q to exist", id)
			continue
		}
		if d.ID() != id {
			t.Errorf("expected driver ID %q, got %q", id, d.ID())
		}
		if d.Name() == "" {
			t.Errorf("expected non-empty Name for driver %q", id)
		}
	}

	_, ok := GetDriver(ClientID("unknown-client"))
	if ok {
		t.Errorf("expected unknown client to return false")
	}
}

func TestClientDrivers_ConfigPaths(t *testing.T) {
	fakeHome := filepath.Join(string(filepath.Separator), "tmp", "fake-home")
	fakeProfile := filepath.Join(string(filepath.Separator), "tmp", "fake-profile")

	claude, _ := GetDriver(ClientClaudeCode)
	p, err := claude.ConfigPath(fakeHome, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p != filepath.Join(fakeHome, ".claude.json") {
		t.Errorf("unexpected global claude path: %s", p)
	}

	pProf, err := claude.ConfigPath(fakeHome, fakeProfile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pProf != filepath.Join(fakeProfile, ".claude", ".claude.json") {
		t.Errorf("unexpected profile claude path: %s", pProf)
	}

	cursor, _ := GetDriver(ClientCursor)
	pCur, err := cursor.ConfigPath(fakeHome, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pCur != filepath.Join(fakeHome, ".cursor", "mcp.json") {
		t.Errorf("unexpected global cursor path: %s", pCur)
	}

	pCurProf, err := cursor.ConfigPath(fakeHome, fakeProfile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pCurProf != filepath.Join(fakeProfile, ".cursor", "mcp.json") {
		t.Errorf("unexpected profile cursor path: %s", pCurProf)
	}
}

func TestClientDrivers_WriteAndReadServer(t *testing.T) {
	tempDir := t.TempDir()
	claude, _ := GetDriver(ClientClaudeCode)

	cfgPath, err := claude.ConfigPath(tempDir, "")
	if err != nil {
		t.Fatalf("failed to resolve claude path: %v", err)
	}

	// 1. Initial read on non-existent file returns empty map, no error
	servers, err := claude.ReadServers(cfgPath)
	if err != nil {
		t.Fatalf("expected nil error on missing config, got: %v", err)
	}
	if len(servers) != 0 {
		t.Fatalf("expected 0 servers initially, got %d", len(servers))
	}

	// 2. Write an aim server entry
	aimEntry := RawMCPEntry{
		Command: "aim",
		Args:    []string{"mcp", "serve"},
	}
	if err := claude.WriteServer(cfgPath, "aim", aimEntry); err != nil {
		t.Fatalf("failed to write aim server: %v", err)
	}

	// Verify file permissions 0600
	fi, err := os.Stat(cfgPath)
	if err != nil {
		t.Fatalf("failed to stat written config: %v", err)
	}
	if runtime.GOOS != "windows" {
		perm := fi.Mode().Perm()
		if perm != 0600 {
			t.Errorf("expected 0600 permissions, got %o", perm)
		}
	}

	// 3. Read back
	servers, err = claude.ReadServers(cfgPath)
	if err != nil {
		t.Fatalf("failed to read back servers: %v", err)
	}
	s, ok := servers["aim"]
	if !ok {
		t.Fatalf("expected 'aim' in servers map")
	}
	if s.Command != "aim" || len(s.Args) != 2 || s.Args[0] != "mcp" || s.Args[1] != "serve" {
		t.Errorf("unexpected aim entry: %+v", s)
	}

	// 4. Add a second server (e.g. github) and verify preservation
	ghEntry := RawMCPEntry{
		Command: "npx",
		Args:    []string{"-y", "@modelcontextprotocol/server-github"},
	}
	if err := claude.WriteServer(cfgPath, "github", ghEntry); err != nil {
		t.Fatalf("failed to write github server: %v", err)
	}

	servers, err = claude.ReadServers(cfgPath)
	if err != nil {
		t.Fatalf("failed to read back servers: %v", err)
	}
	if len(servers) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(servers))
	}
	if _, ok := servers["aim"]; !ok {
		t.Errorf("expected 'aim' to be preserved")
	}
	if _, ok := servers["github"]; !ok {
		t.Errorf("expected 'github' to be present")
	}

	// 5. Delete server
	if err := claude.DeleteServer(cfgPath, "aim"); err != nil {
		t.Fatalf("failed to delete aim server: %v", err)
	}

	servers, err = claude.ReadServers(cfgPath)
	if err != nil {
		t.Fatalf("failed to read back servers after deletion: %v", err)
	}
	if len(servers) != 1 {
		t.Fatalf("expected 1 server left, got %d", len(servers))
	}
	if _, ok := servers["aim"]; ok {
		t.Errorf("aim should have been removed")
	}
	if _, ok := servers["github"]; !ok {
		t.Errorf("github should remain")
	}

	// 6. Delete from empty file should not error
	emptyFile := filepath.Join(tempDir, "empty.json")
	if err := os.WriteFile(emptyFile, []byte(""), 0600); err != nil {
		t.Fatalf("failed to create empty file: %v", err)
	}
	if err := claude.DeleteServer(emptyFile, "nonexistent"); err != nil {
		t.Errorf("expected no error deleting from empty file, got: %v", err)
	}
}

func TestClientDrivers_Detect(t *testing.T) {
	tempDir := t.TempDir()
	// Isolate PATH so host executables do not trigger detection
	t.Setenv("PATH", tempDir)

	claude, _ := GetDriver(ClientClaudeCode)
	cursor, _ := GetDriver(ClientCursor)

	// Initially none detected in empty dir
	if claude.Detect(tempDir) {
		t.Errorf("claude should not be detected in empty temp dir")
	}

	// Create fake claude file
	claudeFile := filepath.Join(tempDir, ".claude.json")
	if err := os.WriteFile(claudeFile, []byte("{}"), 0600); err != nil {
		t.Fatalf("failed to create claude file: %v", err)
	}

	if !claude.Detect(tempDir) {
		t.Errorf("expected claude to be detected when ~/.claude.json exists")
	}

	detected := DetectInstalledClients(tempDir)
	if len(detected) != 1 {
		t.Errorf("expected 1 detected client, got %d", len(detected))
	}

	// Create fake cursor directory
	cursorDir := filepath.Join(tempDir, ".cursor")
	if err := os.MkdirAll(cursorDir, 0700); err != nil {
		t.Fatalf("failed to create cursor dir: %v", err)
	}

	if !cursor.Detect(tempDir) {
		t.Errorf("expected cursor to be detected when ~/.cursor exists")
	}

	detectedAfter := DetectInstalledClients(tempDir)
	if len(detectedAfter) < 2 {
		t.Errorf("expected at least 2 detected clients, got %d", len(detectedAfter))
	}
}

func TestDesktopClients_ConfigPaths(t *testing.T) {
	tempHome := t.TempDir()
	profileDir := filepath.Join(tempHome, ".aim", "profiles", "work")
	_ = os.MkdirAll(profileDir, 0700)

	claudeDriver, ok := GetDriver(ClientClaudeDesktop)
	if !ok {
		t.Fatal("expected claude-desktop driver to be registered")
	}
	cfgPath, err := claudeDriver.ConfigPath(tempHome, profileDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := filepath.Join(profileDir, ".claude-desktop", "claude_desktop_config.json")
	if cfgPath != expected {
		t.Errorf("expected %s, got %s", expected, cfgPath)
	}

	// Host path check
	hostClaudePath, err := claudeDriver.ConfigPath(tempHome, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hostClaudePath == "" {
		t.Error("expected non-empty host path for claude-desktop")
	}

	codexDriver, ok := GetDriver(ClientCodexDesktop)
	if !ok {
		t.Fatal("expected codex-desktop driver to be registered")
	}
	cxdPath, err := codexDriver.ConfigPath(tempHome, profileDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedCxd := filepath.Join(profileDir, ".codex-desktop", "config.json")
	if cxdPath != expectedCxd {
		t.Errorf("expected %s, got %s", expectedCxd, cxdPath)
	}

	hostCodexPath, err := codexDriver.ConfigPath(tempHome, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hostCodexPath == "" {
		t.Error("expected non-empty host path for codex-desktop")
	}
}
