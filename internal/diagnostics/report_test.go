package diagnostics

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/profile"
)

func TestGenerateReport_AnonymizesProfileNamesAndHomePaths(t *testing.T) {
	tempBase := t.TempDir()
	pm := profile.NewProfileManager(tempBase)

	// Create test profiles with sensitive names
	p1Dir := filepath.Join(tempBase, "profiles", "sensitive-client-alpha")
	p2Dir := filepath.Join(tempBase, "profiles", "sensitive-client-beta")
	_ = os.MkdirAll(p1Dir, 0755)
	_ = os.MkdirAll(p2Dir, 0755)

	reg := agents.NewRegistry()

	report := GenerateReport(reg, pm, nil, "", "v1.0.0", "abcdef1")

	// Verify report header and metadata
	if !strings.Contains(report, "# AIM Diagnostic Report") {
		t.Fatal("expected report title")
	}
	if !strings.Contains(report, "AIM Version: v1.0.0 (commit: abcdef1)") {
		t.Errorf("expected version and commit in report")
	}

	// Verify total profile count is present
	if !strings.Contains(report, "Total Configured Profiles: 2") {
		t.Errorf("expected total configured profiles to be 2, got:\n%s", report)
	}

	// Sensitive profile names must be redacted from diagnostics checks
	if strings.Contains(report, "sensitive-client-alpha") {
		t.Errorf("diagnostic report leaked profile name 'sensitive-client-alpha':\n%s", report)
	}
	if strings.Contains(report, "sensitive-client-beta") {
		t.Errorf("diagnostic report leaked profile name 'sensitive-client-beta':\n%s", report)
	}

	// Home dir / base dir paths must be sanitized
	home, _ := os.UserHomeDir()
	if home != "" && strings.Contains(report, home) {
		t.Errorf("report contains unredacted home directory %q", home)
	}
}
