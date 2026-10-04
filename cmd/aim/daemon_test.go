package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/daemon"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/usage"
	"github.com/spf13/cobra"
)

func captureStdout(fn func() error) (string, error) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := fn()

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String(), err
}

func executeCmd(cmd *cobra.Command, args ...string) (string, error) {
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

func TestDaemonCmd_Help(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)

	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)

	cmd := newRootCmd(reg, pm)
	output, err := executeCmd(cmd, "daemon", "--help")

	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	for _, sub := range []string{"install", "uninstall", "status", "run"} {
		if !strings.Contains(output, sub) {
			t.Errorf("expected help output to mention subcommand %q, got: %s", sub, output)
		}
	}
}

func TestDaemonCmd_Status_Plain(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)

	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)

	cmd := newRootCmd(reg, pm)
	output, err := executeCmd(cmd, "daemon", "status")

	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	if !strings.Contains(output, "AIM Background Daemon Status") {
		t.Errorf("expected header, got: %s", output)
	}
	if !strings.Contains(output, "Installed:    No") {
		t.Errorf("expected Installed: No, got: %s", output)
	}
	if !strings.Contains(output, "dev.aim-cli.daemon") {
		t.Errorf("expected service label in output, got: %s", output)
	}
}

func TestDaemonCmd_Status_JSON(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)

	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)

	cmd := newRootCmd(reg, pm)
	output, err := executeCmd(cmd, "daemon", "status", "--json")

	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	var info daemon.ServiceInfo
	if err := json.Unmarshal([]byte(output), &info); err != nil {
		t.Fatalf("failed to parse JSON status: %v (raw: %s)", err, output)
	}

	if info.Installed {
		t.Errorf("expected Installed to be false")
	}
	if info.Label != daemon.DefaultServiceLabel {
		t.Errorf("expected label %q, got %q", daemon.DefaultServiceLabel, info.Label)
	}
}

type testUsageAdapter struct {
	agents.BaseAdapter
}

func (m *testUsageAdapter) HasCredentials(profileDir string) bool { return true }
func (m *testUsageAdapter) PrepareEnv(ctx context.Context, profileName, profileDir string) (agents.LaunchEnv, error) {
	return agents.LaunchEnv{}, nil
}
func (m *testUsageAdapter) GetUsage(ctx context.Context, profile, profileDir string) (*usage.Report, error) {
	return &usage.Report{
		Agent:     m.Name(),
		Profile:   profile,
		FetchedAt: time.Now(),
		Status:    usage.StatusOK,
		Windows: []usage.LimitWindow{
			{Name: "5-hour limit", RemainingPct: 80},
		},
	}, nil
}

func TestDaemonCmd_Run(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)

	reg := agents.NewRegistry()
	ad := &testUsageAdapter{
		BaseAdapter: agents.NewBaseAdapter("mock", "Mock Agent", "mock", nil),
	}
	reg.Register(ad)

	pm := profile.NewProfileManager(tempDir)
	if _, err := pm.EnsureProfile("work"); err != nil {
		t.Fatalf("EnsureProfile failed: %v", err)
	}

	cmd := newRootCmd(reg, pm)
	_, err := executeCmd(cmd, "daemon", "run")
	if err != nil {
		t.Fatalf("daemon run failed: %v", err)
	}

	// Verify daemon.log was populated
	logPath := filepath.Join(tempDir, "daemon.log")
	lastRun, msg := daemon.ReadLastDaemonRun(logPath)
	if lastRun.IsZero() {
		t.Errorf("expected valid last run timestamp in daemon.log")
	}
	if !strings.Contains(msg, "quota refresh completed") {
		t.Errorf("expected log message to contain 'quota refresh completed', got: %s", msg)
	}

	// Verify cache was pre-warmed
	cache := usage.NewCacheStore(tempDir, usage.DefaultTTL)
	rep, ok := cache.Get("mock", "work")
	if !ok || rep.Status != usage.StatusOK {
		t.Errorf("expected cached usage report for mock:work, got: %+v (ok=%v)", rep, ok)
	}
}

func TestDoctorCmd_DaemonSection(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)

	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)

	output, err := captureStdout(func() error {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{"doctor"})
		return cmd.Execute()
	})

	if err != nil {
		t.Fatalf("doctor failed: %v", err)
	}

	if !strings.Contains(output, "[Background Daemon]") {
		t.Errorf("expected doctor output to contain [Background Daemon], got: %s", output)
	}
	if !strings.Contains(output, "Daemon:") {
		t.Errorf("expected doctor output to contain Daemon: line, got: %s", output)
	}
	if !strings.Contains(output, "aim daemon install") {
		t.Errorf("expected doctor output to suggest 'aim daemon install', got: %s", output)
	}
}
