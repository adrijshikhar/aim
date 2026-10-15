package mcp

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestClientDrivers_Registry(t *testing.T) {
	drivers := AllDrivers()
	if len(drivers) < 4 {
		t.Fatalf("expected at least 4 drivers, got %d", len(drivers))
	}

	expectedIDs := []ClientID{ClientClaudeCode, ClientClaudeDesktop, ClientCursor, ClientWindsurf}
	for _, id := range expectedIDs {
		d, ok := GetDriver(id)
		if !ok || d == nil {
			t.Errorf("expected driver for %q to exist", id)
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
	pCursor, err := cursor.ConfigPath(fakeHome, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pCursor != filepath.Join(fakeHome, ".cursor", "mcp.json") {
		t.Errorf("unexpected cursor path: %s", pCursor)
	}

	windsurf, _ := GetDriver(ClientWindsurf)
	pWindsurf, err := windsurf.ConfigPath(fakeHome, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pWindsurf != filepath.Join(fakeHome, ".codeium", "windsurf", "mcp_config.json") {
		t.Errorf("unexpected windsurf path: %s", pWindsurf)
	}

	desktop, _ := GetDriver(ClientClaudeDesktop)
	pDesktop, err := desktop.ConfigPath(fakeHome, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	switch runtime.GOOS {
	case "darwin":
		expected := filepath.Join(fakeHome, "Library", "Application Support", "Claude", "claude_desktop_config.json")
		if pDesktop != expected {
			t.Errorf("expected desktop path %q, got %q", expected, pDesktop)
		}
	case "windows":
		// On Windows, expected to be in Roaming/Claude
		if !filepath.IsAbs(pDesktop) {
			t.Errorf("expected absolute path on Windows: %s", pDesktop)
		}
	default:
		expected := filepath.Join(fakeHome, ".config", "Claude", "claude_desktop_config.json")
		if pDesktop != expected {
			t.Errorf("expected linux desktop path %q, got %q", expected, pDesktop)
		}
	}
}

func TestClientDrivers_WriteAndReadServer(t *testing.T) {
	tempDir := t.TempDir()
	cursor, _ := GetDriver(ClientCursor)

	cfgPath, err := cursor.ConfigPath(tempDir, "")
	if err != nil {
		t.Fatalf("failed to resolve cursor path: %v", err)
	}

	// 1. Initial read on non-existent file returns empty map, no error
	servers, err := cursor.ReadServers(cfgPath)
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
	if err := cursor.WriteServer(cfgPath, "aim", aimEntry); err != nil {
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
	servers, err = cursor.ReadServers(cfgPath)
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
	if err := cursor.WriteServer(cfgPath, "github", ghEntry); err != nil {
		t.Fatalf("failed to write github server: %v", err)
	}

	servers, err = cursor.ReadServers(cfgPath)
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
	if err := cursor.DeleteServer(cfgPath, "aim"); err != nil {
		t.Fatalf("failed to delete aim server: %v", err)
	}

	servers, err = cursor.ReadServers(cfgPath)
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
}

func TestClientDrivers_Detect(t *testing.T) {
	tempDir := t.TempDir()

	// Initially, none are detected in empty tempDir (unless client is on PATH)
	claude, _ := GetDriver(ClientClaudeCode)
	cursor, _ := GetDriver(ClientCursor)

	// Create fake cursor directory
	cursorDir := filepath.Join(tempDir, ".cursor")
	if err := os.MkdirAll(cursorDir, 0700); err != nil {
		t.Fatalf("failed to create cursor dir: %v", err)
	}

	if !cursor.Detect(tempDir) {
		t.Errorf("expected cursor to be detected when ~/.cursor exists")
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
	if len(detected) < 2 {
		t.Errorf("expected at least 2 detected clients, got %d", len(detected))
	}
}
