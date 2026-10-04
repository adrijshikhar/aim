package daemon

import (
	"time"
)

const (
	// DefaultServiceLabel identifies the AIM background daemon in launchd and systemd.
	DefaultServiceLabel = "dev.aim-cli.daemon"

	// DefaultInterval is the scheduled period between background job runs.
	DefaultInterval = 15 * time.Minute

	// DefaultIntervalSec is the interval in seconds for launchd StartInterval.
	DefaultIntervalSec = 900

	// DefaultSystemdService is the unit name on Linux platforms.
	DefaultSystemdService = "aim-daemon"
)

// PlistConfig holds parameters for generating macOS launchd plists.
type PlistConfig struct {
	Label       string
	BinaryPath  string
	Args        []string
	IntervalSec int
	LogPath     string
	EnvVars     map[string]string
}

// SystemdConfig holds parameters for generating Linux systemd unit and timer files.
type SystemdConfig struct {
	ServiceName string
	BinaryPath  string
	Args        []string
	Interval    time.Duration
	LogPath     string
}

// ServiceInfo describes the operational state of the background daemon.
type ServiceInfo struct {
	Installed      bool          `json:"installed"`
	Active         bool          `json:"active"`
	Label          string        `json:"label"`
	ConfigPath     string        `json:"config_path"`
	Interval       time.Duration `json:"interval"`
	BinaryPath     string        `json:"binary_path"`
	LogPath        string        `json:"log_path"`
	LastRun        time.Time     `json:"last_run,omitempty"`
	LastRunMessage string        `json:"last_run_message,omitempty"`
}
