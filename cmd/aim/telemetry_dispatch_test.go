package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/telemetry"
)

func TestTelemetry_DispatchesCommandEvent(t *testing.T) {
	tempBase := t.TempDir()
	origHome := os.Getenv("AIM_HOME")
	origTestTelemetry := os.Getenv("AIM_ENABLE_TEST_TELEMETRY")
	defer func() {
		_ = os.Setenv("AIM_HOME", origHome)
		_ = os.Setenv("AIM_ENABLE_TEST_TELEMETRY", origTestTelemetry)
	}()
	_ = os.Setenv("AIM_HOME", tempBase)
	_ = os.Setenv("AIM_ENABLE_TEST_TELEMETRY", "1")

	pm := profile.NewProfileManager(tempBase)
	reg := agents.NewRegistry()

	// Ensure DO_NOT_TRACK is not set
	_ = os.Unsetenv("DO_NOT_TRACK")
	_ = os.Unsetenv("AIM_TELEMETRY_DISABLED")

	captureOutput(t, func() {
		code := dispatch([]string{"version"}, reg, pm)
		if code != 0 {
			t.Fatalf("expected version command to exit 0, got %d", code)
		}
	})

	// Check that spool file was written with command_executed event
	cacheDir := filepath.Join(tempBase, "cache")
	spoolFile := filepath.Join(cacheDir, "telemetry_spool.json")
	spooler := telemetry.NewSpooler(spoolFile, "", "", nil)

	events, err := spooler.Read()
	if err != nil {
		t.Fatalf("failed to read spool: %v", err)
	}
	if len(events) == 0 {
		t.Fatal("expected at least 1 telemetry event in spool, got 0")
	}

	found := false
	for _, ev := range events {
		if ev.EventName == telemetry.EventCommandExecuted {
			found = true
			if ev.Properties["command"] != "version" {
				t.Errorf("expected command 'version', got %v", ev.Properties["command"])
			}
		}
	}
	if !found {
		t.Errorf("expected EventCommandExecuted in events, got: %+v", events)
	}
}

func TestTelemetry_RespectsDoNotTrackInCLI(t *testing.T) {
	tempBase := t.TempDir()
	origHome := os.Getenv("AIM_HOME")
	origDNT := os.Getenv("DO_NOT_TRACK")
	defer func() {
		_ = os.Setenv("AIM_HOME", origHome)
		_ = os.Setenv("DO_NOT_TRACK", origDNT)
	}()
	_ = os.Setenv("AIM_HOME", tempBase)
	_ = os.Setenv("DO_NOT_TRACK", "1")

	pm := profile.NewProfileManager(tempBase)
	reg := agents.NewRegistry()

	captureOutput(t, func() {
		_ = dispatch([]string{"version"}, reg, pm)
	})

	cacheDir := filepath.Join(tempBase, "cache")
	spoolFile := filepath.Join(cacheDir, "telemetry_spool.json")
	if _, err := os.Stat(spoolFile); err == nil {
		t.Errorf("expected spool file NOT to exist when DO_NOT_TRACK=1")
	}
}
