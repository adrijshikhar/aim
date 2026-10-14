package diagnostics

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/oauth"
	"github.com/aim-cli/aim/internal/profile"
)

type mockReportAdapter struct {
	name string
}

func (m *mockReportAdapter) Name() string                          { return m.name }
func (m *mockReportAdapter) DisplayName() string                   { return m.name }
func (m *mockReportAdapter) Aliases() []string                     { return nil }
func (m *mockReportAdapter) BinaryName() string                    { return m.name }
func (m *mockReportAdapter) HasCredentials(profileDir string) bool { return true }
func (m *mockReportAdapter) PrepareEnv(ctx context.Context, p, d string) (agents.LaunchEnv, error) {
	return agents.LaunchEnv{}, nil
}

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

func TestGenerateReport_IncludesOAuthRefusal(t *testing.T) {
	tempBase := t.TempDir()
	pm := profile.NewProfileManager(tempBase)
	_, _ = pm.EnsureProfile("work")

	refStore := oauth.NewRefusalStore(tempBase)
	if err := refStore.Record("claude", "work", oauth.RefusalRefreshTokenReused, "token consumed"); err != nil {
		t.Fatalf("failed to record refusal: %v", err)
	}

	reg := agents.NewRegistry()
	reg.Register(&mockReportAdapter{name: "claude"})

	report := GenerateReport(reg, pm, nil, "claude", "v1.0.0", "abcdef1")

	expectedRow := "| claude | Auth | FAIL | Re-login Required (refresh_token_reused) |"
	if !strings.Contains(report, expectedRow) {
		t.Errorf("expected report to contain %q, but got:\n%s", expectedRow, report)
	}
}
