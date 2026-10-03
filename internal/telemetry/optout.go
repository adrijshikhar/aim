package telemetry

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aim-cli/aim/internal/config"
	"github.com/mattn/go-isatty"
)

// IsTelemetryEnabled checks the three opt-out tiers in order of precedence:
// 1. Universal DO_NOT_TRACK standard
// 2. AIM_TELEMETRY_DISABLED environment variable
// 3. User configuration file setting
func IsTelemetryEnabled(cfg *config.Config) bool {
	if os.Getenv("DO_NOT_TRACK") == "1" {
		return false
	}
	if os.Getenv("AIM_TELEMETRY_DISABLED") == "1" {
		return false
	}
	if cfg != nil && cfg.TelemetryDisabled {
		return false
	}
	return true
}

// MaybeDisplayFirstRunNotice prints a one-time transparent notice on stderr if the user
// has not previously run AIM, creating a marker file ~/.aim/.telemetry_noticed.
func MaybeDisplayFirstRunNotice(baseDir string) bool {
	if baseDir == "" {
		baseDir = config.BaseDir()
	}
	marker := filepath.Join(baseDir, ".telemetry_noticed")
	if _, err := os.Stat(marker); err == nil {
		return false
	}

	_ = os.MkdirAll(baseDir, 0755)
	_ = os.WriteFile(marker, []byte("1"), 0644)

	if !IsTelemetryEnabled(nil) {
		return true
	}

	if isatty.IsTerminal(os.Stderr.Fd()) || isatty.IsCygwinTerminal(os.Stderr.Fd()) {
		fmt.Fprintln(os.Stderr, "AIM collects anonymous usage metrics to improve developer experience.")
		fmt.Fprintln(os.Stderr, "To disable, set DO_NOT_TRACK=1, AIM_TELEMETRY_DISABLED=1, or run 'aim config set telemetry off'.")
	}

	return true
}
