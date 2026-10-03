package main

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
)

type panickingDoctorAdapter struct {
	mockAdapter
}

func (p *panickingDoctorAdapter) Doctor(ctx context.Context, profileName, profileDir string) []agents.DiagnosticResult {
	panic("simulated fatal crash in doctor diagnostics")
}

func TestDispatch_PanicRecovery(t *testing.T) {
	tempBase := t.TempDir()
	origHome := os.Getenv("AIM_HOME")
	defer func() { _ = os.Setenv("AIM_HOME", origHome) }()
	_ = os.Setenv("AIM_HOME", tempBase)

	pm := profile.NewProfileManager(tempBase)
	_, _ = pm.EnsureProfile("testprof")
	cfg := config.NewDefaultConfig()
	cfg.AddProfileAgent("testprof", "panic-doc")
	_ = config.SaveConfig(cfg)

	reg := agents.NewRegistry()
	reg.Register(&panickingDoctorAdapter{mockAdapter: mockAdapter{name: "panic-doc", binaryPath: "/bin/sh"}})

	var code int
	_, errOut := captureOutput(t, func() {
		code = dispatch([]string{"doctor", "panic-doc"}, reg, pm)
	})

	if code != 1 {
		t.Errorf("expected exit code 1 on panic, got %d", code)
	}
	if !strings.Contains(errOut, "AIM encountered an unexpected crash") {
		t.Errorf("expected crash banner in stderr, got:\n%s", errOut)
	}
	if !strings.Contains(errOut, "github.com/adrijshikhar/aim/issues/new") {
		t.Errorf("expected GitHub issue link in stderr, got:\n%s", errOut)
	}
	if !strings.Contains(errOut, "simulated fatal crash in doctor diagnostics") {
		t.Errorf("expected panic message in output, got:\n%s", errOut)
	}
}
