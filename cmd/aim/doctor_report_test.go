package main

import (
	"os"
	"strings"
	"testing"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
)

func TestDoctorReport_FlagGeneratesSanitizedMarkdown(t *testing.T) {
	tempBase := t.TempDir()
	origHome := os.Getenv("AIM_HOME")
	defer func() { _ = os.Setenv("AIM_HOME", origHome) }()
	_ = os.Setenv("AIM_HOME", tempBase)

	pm := profile.NewProfileManager(tempBase)
	_, _ = pm.EnsureProfile("work")
	_, _ = pm.EnsureProfile("staging")

	cfg := config.NewDefaultConfig()
	cfg.AddProfileAgent("work", "mock")
	cfg.AddProfileAgent("staging", "mock")
	_ = config.SaveConfig(cfg)

	reg := agents.NewRegistry()
	reg.Register(&mockAdapter{name: "mock", binaryPath: "/bin/sh"})

	// Execute doctor with --report flag
	out, errOut := captureOutput(t, func() {
		code := dispatch([]string{"doctor", "--report"}, reg, pm)
		if code != 0 {
			t.Fatalf("expected doctor --report to exit with code 0, got %d", code)
		}
	})

	if errOut != "" {
		t.Errorf("expected clean stderr, got:\n%s", errOut)
	}

	// Assert Markdown structure
	if !strings.Contains(out, "# AIM Diagnostic Report") {
		t.Errorf("expected markdown header in doctor --report, got:\n%s", out)
	}
	if !strings.Contains(out, "AIM Version:") {
		t.Errorf("expected AIM Version section, got:\n%s", out)
	}
	if !strings.Contains(out, "Environment:") {
		t.Errorf("expected Environment section, got:\n%s", out)
	}
	if !strings.Contains(out, "Agents & Tooling") {
		t.Errorf("expected Agents & Tooling section, got:\n%s", out)
	}
	if !strings.Contains(out, "| mock |") {
		t.Errorf("expected mock agent in tooling table, got:\n%s", out)
	}
	if !strings.Contains(out, "Total Configured Profiles: 2") {
		t.Errorf("expected profile count in report, got:\n%s", out)
	}

	// Verify home directory is sanitized to ~/
	realHome, _ := os.UserHomeDir()
	if realHome != "" && realHome != "/" && strings.Contains(out, realHome) {
		t.Errorf("expected real home dir %q to be sanitized with ~/, got:\n%s", realHome, out)
	}
}
