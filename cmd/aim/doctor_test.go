package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/profile"
)

type doctorMockAdapter struct {
	mockAdapter
	failDr bool
}

func (m *doctorMockAdapter) Doctor(ctx context.Context, profileName, profileDir string) []agents.DiagnosticResult {
	if m.failDr {
		return []agents.DiagnosticResult{
			{Category: "Binary", Status: "FAIL", Message: "binary missing or broken"},
		}
	}
	return []agents.DiagnosticResult{
		{Category: "Binary", Status: "OK", Message: "binary ready"},
	}
}

func TestDoctorCmd_CheckFlag(t *testing.T) {
	tempDir := t.TempDir()
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("testprof")

	reg := agents.NewRegistry()
	reg.Register(&doctorMockAdapter{mockAdapter: mockAdapter{name: "healthy", hasCreds: true}, failDr: false})
	reg.Register(&doctorMockAdapter{mockAdapter: mockAdapter{name: "broken", hasCreds: true}, failDr: true})

	cmdHealthy := newDoctorCmd(reg, pm)
	cmdHealthy.SetArgs([]string{"healthy", "--check"})
	if err := cmdHealthy.Execute(); err != nil {
		t.Fatalf("expected healthy doctor --check to succeed, got: %v", err)
	}

	cmdBroken := newDoctorCmd(reg, pm)
	cmdBroken.SetArgs([]string{"broken", "--check"})
	if err := cmdBroken.Execute(); err == nil {
		t.Fatalf("expected broken doctor --check to fail with error")
	}

	cmdAllBroken := newDoctorCmd(reg, pm)
	cmdAllBroken.SetArgs([]string{"--check"})
	if err := cmdAllBroken.Execute(); err == nil {
		t.Fatalf("expected doctor --check across all agents to fail if broken agent present")
	}

	// Without --check, command exits cleanly (exit 0)
	cmdNoCheck := newDoctorCmd(reg, pm)
	cmdNoCheck.SetArgs([]string{"broken"})
	if err := cmdNoCheck.Execute(); err != nil {
		t.Fatalf("expected doctor without --check to exit cleanly, got: %v", err)
	}
}

// The Bridge block describes the profile, not the agent, so a profile with two
// agents shows it once, under the first agent's block.
func TestDoctor_BridgeBlockOncePerProfile(t *testing.T) {
	host := t.TempDir()
	t.Setenv("AIM_REAL_HOME", host)
	if err := os.WriteFile(filepath.Join(host, ".gitconfig"), []byte("[user]"), 0644); err != nil {
		t.Fatal(err)
	}
	pm := profile.NewProfileManager(t.TempDir())
	if _, err := pm.EnsureProfile("shared"); err != nil {
		t.Fatal(err)
	}
	reg := agents.NewRegistry()
	reg.Register(&doctorMockAdapter{mockAdapter: mockAdapter{name: "one", hasCreds: true}})
	reg.Register(&doctorMockAdapter{mockAdapter: mockAdapter{name: "two", hasCreds: true}})

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	runDoctor(reg, pm, "")
	_ = w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	out := buf.String()

	if n := strings.Count(out, "Profile: "); n != 2 {
		t.Fatalf("expected the profile under both agents, got %d blocks:\n%s", n, out)
	}
	if n := strings.Count(out, "Bridge:"); n != 1 {
		t.Errorf("expected one Bridge line for the profile, got %d:\n%s", n, out)
	}
}
