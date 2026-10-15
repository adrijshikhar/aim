package mcp

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/profile"
)

func TestMCPDoctor_SelfHandshakePasses(t *testing.T) {
	tempDir := t.TempDir()
	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	report := RunMCPDoctor(ctx, reg, pm, tempDir)

	if report.OverallStatus == StatusFail {
		t.Fatalf("expected overall status not to be FAIL, got %s. Results: %+v", report.OverallStatus, report.Results)
	}

	foundProtocolPass := false
	for _, res := range report.Results {
		if res.Category == "Protocol" && res.Status == StatusPass {
			foundProtocolPass = true
			break
		}
	}
	if !foundProtocolPass {
		t.Errorf("expected passing Protocol check in results: %+v", report.Results)
	}
}

func TestMCPDoctor_ClientRegistrationAudit(t *testing.T) {
	tempDir := t.TempDir()
	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)

	// Create Cursor config directory and file without aim
	cursor, _ := GetDriver(ClientCursor)
	cfgPath, _ := cursor.ConfigPath(tempDir, "")
	_ = cursor.WriteServer(cfgPath, "existing-tool", RawMCPEntry{
		Command: "ls", // valid binary in PATH
		Args:    []string{"-la"},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	report := RunMCPDoctor(ctx, reg, pm, tempDir)

	// Since Cursor is detected but 'aim' is missing, it should report a WARN
	foundCursorWarn := false
	for _, res := range report.Results {
		if res.Category == "Client" && res.Status == StatusWarn {
			foundCursorWarn = true
			if res.Remedy == "" {
				t.Errorf("expected remedy instructions for missing aim registration")
			}
			break
		}
	}
	if !foundCursorWarn {
		t.Errorf("expected WARN for unregistered Cursor client in results: %+v", report.Results)
	}

	// Now register aim in Cursor and re-run
	_ = cursor.WriteServer(cfgPath, "aim", RawMCPEntry{
		Command: "aim",
		Args:    []string{"mcp", "serve"},
	})

	report2 := RunMCPDoctor(ctx, reg, pm, tempDir)
	foundCursorPass := false
	for _, res := range report2.Results {
		if res.Category == "Client" && res.Status == StatusPass {
			foundCursorPass = true
			break
		}
	}
	if !foundCursorPass {
		t.Errorf("expected OK for registered Cursor client in results: %+v", report2.Results)
	}
}

func TestMCPDoctor_ThirdPartyBinaryMissing(t *testing.T) {
	tempDir := t.TempDir()
	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)

	cursor, _ := GetDriver(ClientCursor)
	cfgPath, _ := cursor.ConfigPath(tempDir, "")

	// Write an entry pointing to a completely non-existent binary
	_ = cursor.WriteServer(cfgPath, "broken-server", RawMCPEntry{
		Command: "totally-non-existent-binary-12345",
	})
	_ = cursor.WriteServer(cfgPath, "aim", RawMCPEntry{
		Command: "aim",
		Args:    []string{"mcp", "serve"},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	report := RunMCPDoctor(ctx, reg, pm, tempDir)

	if report.OverallStatus != StatusFail {
		t.Errorf("expected overall status FAIL when third-party binary is missing, got %s", report.OverallStatus)
	}

	foundBinaryFail := false
	for _, res := range report.Results {
		if res.Category == "Third-Party" && res.Status == StatusFail {
			foundBinaryFail = true
			if res.Remedy == "" {
				t.Errorf("expected remedy instructions for missing binary")
			}
			break
		}
	}
	if !foundBinaryFail {
		t.Errorf("expected FAIL for missing binary in third-party server: %+v", report.Results)
	}
}

func TestMCPDoctor_MalformedJSONConfig(t *testing.T) {
	tempDir := t.TempDir()
	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)

	cursorDir := filepath.Join(tempDir, ".cursor")
	_ = os.MkdirAll(cursorDir, 0700)
	_ = os.WriteFile(filepath.Join(cursorDir, "mcp.json"), []byte("{ broken-json: "), 0600)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	report := RunMCPDoctor(ctx, reg, pm, tempDir)

	foundConfigFail := false
	for _, res := range report.Results {
		if res.Category == "Config" && res.Status == StatusFail {
			foundConfigFail = true
			break
		}
	}
	if !foundConfigFail {
		t.Errorf("expected FAIL for malformed JSON config: %+v", report.Results)
	}
}
