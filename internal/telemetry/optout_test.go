package telemetry

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aim-cli/aim/internal/config"
)

func TestIsTelemetryEnabled_OptOutRules(t *testing.T) {
	// Clean env
	origDNT := os.Getenv("DO_NOT_TRACK")
	origAIM := os.Getenv("AIM_TELEMETRY_DISABLED")
	defer func() {
		_ = os.Setenv("DO_NOT_TRACK", origDNT)
		_ = os.Setenv("AIM_TELEMETRY_DISABLED", origAIM)
	}()

	_ = os.Unsetenv("DO_NOT_TRACK")
	_ = os.Unsetenv("AIM_TELEMETRY_DISABLED")

	// Default should be enabled
	cfg := config.NewDefaultConfig()
	if !IsTelemetryEnabled(cfg) {
		t.Errorf("expected telemetry enabled by default")
	}

	// 1. DO_NOT_TRACK=1
	_ = os.Setenv("DO_NOT_TRACK", "1")
	if IsTelemetryEnabled(cfg) {
		t.Errorf("expected DO_NOT_TRACK=1 to disable telemetry")
	}
	_ = os.Unsetenv("DO_NOT_TRACK")

	// 2. AIM_TELEMETRY_DISABLED=1
	_ = os.Setenv("AIM_TELEMETRY_DISABLED", "1")
	if IsTelemetryEnabled(cfg) {
		t.Errorf("expected AIM_TELEMETRY_DISABLED=1 to disable telemetry")
	}
	_ = os.Unsetenv("AIM_TELEMETRY_DISABLED")

	// 3. cfg.TelemetryDisabled = true
	cfg.TelemetryDisabled = true
	if IsTelemetryEnabled(cfg) {
		t.Errorf("expected cfg.TelemetryDisabled=true to disable telemetry")
	}
}

func TestDisplayFirstRunNoticeIfUnnoticed(t *testing.T) {
	tempDir := t.TempDir()

	// First run should return true (and write marker)
	shown1 := MaybeDisplayFirstRunNotice(tempDir)
	if !shown1 {
		t.Errorf("expected notice to be shown on first run")
	}

	markerPath := filepath.Join(tempDir, ".telemetry_noticed")
	if _, err := os.Stat(markerPath); os.IsNotExist(err) {
		t.Errorf("expected notice marker file to exist at %s", markerPath)
	}

	// Second run should return false (already noticed)
	shown2 := MaybeDisplayFirstRunNotice(tempDir)
	if shown2 {
		t.Errorf("expected notice NOT to be shown on subsequent runs")
	}
}
