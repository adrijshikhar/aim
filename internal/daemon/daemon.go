package daemon

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/usage"
)

// LaunchAgentPath returns the destination plist path for macOS launchd.
func LaunchAgentPath(homeDir string) string {
	return filepath.Join(homeDir, "Library", "LaunchAgents", DefaultServiceLabel+".plist")
}

// SystemdServicePath returns the destination service unit path on Linux.
func SystemdServicePath(homeDir string) string {
	return filepath.Join(homeDir, ".config", "systemd", "user", DefaultSystemdService+".service")
}

// SystemdTimerPath returns the destination timer unit path on Linux.
func SystemdTimerPath(homeDir string) string {
	return filepath.Join(homeDir, ".config", "systemd", "user", DefaultSystemdService+".timer")
}

// GeneratePlist produces a launchd-compliant XML plist document.
func GeneratePlist(cfg PlistConfig) ([]byte, error) {
	if cfg.Label == "" {
		cfg.Label = DefaultServiceLabel
	}
	if cfg.IntervalSec <= 0 {
		cfg.IntervalSec = DefaultIntervalSec
	}

	var buf bytes.Buffer
	buf.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	buf.WriteString("<!DOCTYPE plist PUBLIC \"-//Apple//DTD PLIST 1.0//EN\" \"http://www.apple.com/DTDs/PropertyList-1.0.dtd\">\n")
	buf.WriteString("<plist version=\"1.0\">\n<dict>\n")

	// Label
	buf.WriteString(fmt.Sprintf("\t<key>Label</key>\n\t<string>%s</string>\n", cfg.Label))

	// ProgramArguments
	buf.WriteString("\t<key>ProgramArguments</key>\n\t<array>\n")
	buf.WriteString(fmt.Sprintf("\t\t<string>%s</string>\n", cfg.BinaryPath))
	for _, arg := range cfg.Args {
		buf.WriteString(fmt.Sprintf("\t\t<string>%s</string>\n", arg))
	}
	buf.WriteString("\t</array>\n")

	// StartInterval
	buf.WriteString(fmt.Sprintf("\t<key>StartInterval</key>\n\t<integer>%d</integer>\n", cfg.IntervalSec))

	// RunAtLoad
	buf.WriteString("\t<key>RunAtLoad</key>\n\t<true/>\n")

	// StandardOutPath & StandardErrorPath
	if cfg.LogPath != "" {
		buf.WriteString(fmt.Sprintf("\t<key>StandardOutPath</key>\n\t<string>%s</string>\n", cfg.LogPath))
		buf.WriteString(fmt.Sprintf("\t<key>StandardErrorPath</key>\n\t<string>%s</string>\n", cfg.LogPath))
	}

	// EnvironmentVariables
	if len(cfg.EnvVars) > 0 {
		buf.WriteString("\t<key>EnvironmentVariables</key>\n\t<dict>\n")
		for k, v := range cfg.EnvVars {
			buf.WriteString(fmt.Sprintf("\t\t<key>%s</key>\n\t\t<string>%s</string>\n", k, v))
		}
		buf.WriteString("\t</dict>\n")
	}

	buf.WriteString("</dict>\n</plist>\n")
	return buf.Bytes(), nil
}

// GenerateSystemdService generates a systemd user service unit.
func GenerateSystemdService(cfg SystemdConfig) ([]byte, error) {
	if cfg.ServiceName == "" {
		cfg.ServiceName = DefaultSystemdService
	}

	var buf bytes.Buffer
	buf.WriteString("[Unit]\n")
	buf.WriteString("Description=AIM CLI Background Daemon\n")
	buf.WriteString("After=network.target\n\n")

	buf.WriteString("[Service]\n")
	buf.WriteString("Type=oneshot\n")

	cmdParts := []string{cfg.BinaryPath}
	cmdParts = append(cmdParts, cfg.Args...)
	buf.WriteString(fmt.Sprintf("ExecStart=%s\n", strings.Join(cmdParts, " ")))

	if cfg.LogPath != "" {
		buf.WriteString(fmt.Sprintf("StandardOutput=append:%s\n", cfg.LogPath))
		buf.WriteString(fmt.Sprintf("StandardError=append:%s\n", cfg.LogPath))
	}

	return buf.Bytes(), nil
}

// GenerateSystemdTimer generates a systemd user timer unit.
func GenerateSystemdTimer(cfg SystemdConfig) ([]byte, error) {
	if cfg.ServiceName == "" {
		cfg.ServiceName = DefaultSystemdService
	}
	intervalStr := "15m"
	if cfg.Interval > 0 {
		intervalStr = fmt.Sprintf("%dm", int(cfg.Interval.Minutes()))
	}

	var buf bytes.Buffer
	buf.WriteString("[Unit]\n")
	buf.WriteString(fmt.Sprintf("Description=AIM CLI Periodic Background Tasks (%s)\n\n", intervalStr))

	buf.WriteString("[Timer]\n")
	buf.WriteString("OnBootSec=1m\n")
	buf.WriteString(fmt.Sprintf("OnUnitActiveSec=%s\n", intervalStr))
	buf.WriteString(fmt.Sprintf("Unit=%s.service\n\n", cfg.ServiceName))

	buf.WriteString("[Install]\n")
	buf.WriteString("WantedBy=timers.target\n")

	return buf.Bytes(), nil
}

// AppendDaemonLog writes a timestamped record to the daemon log.
func AppendDaemonLog(logPath, message string) error {
	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05 UTC")
	_, err = fmt.Fprintf(f, "[%s] %s\n", timestamp, strings.TrimSpace(message))
	return err
}

// ReadLastDaemonRun parses the most recent execution timestamp and message from the daemon log.
func ReadLastDaemonRun(logPath string) (time.Time, string) {
	data, err := os.ReadFile(logPath)
	if err != nil || len(data) == 0 {
		return time.Time{}, ""
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "[") {
			idx := strings.Index(line, "]")
			if idx > 1 {
				timeStr := line[1:idx]
				msg := strings.TrimSpace(line[idx+1:])
				if t, err := time.Parse("2006-01-02 15:04:05 MST", timeStr); err == nil {
					return t, msg
				}
			}
		}
	}
	return time.Time{}, ""
}

// Status inspects whether the daemon is installed and active on the host OS.
func Status(baseDir string, homeOverride ...string) (*ServiceInfo, error) {
	home := config.RealHomeDir()
	if len(homeOverride) > 0 && homeOverride[0] != "" {
		home = homeOverride[0]
	}

	logPath := filepath.Join(baseDir, "daemon.log")
	lastRun, lastMsg := ReadLastDaemonRun(logPath)

	info := &ServiceInfo{
		Label:          DefaultServiceLabel,
		Interval:       DefaultInterval,
		LogPath:        logPath,
		LastRun:        lastRun,
		LastRunMessage: lastMsg,
	}

	if runtime.GOOS == "darwin" {
		plistPath := LaunchAgentPath(home)
		info.ConfigPath = plistPath

		if _, err := os.Stat(plistPath); err == nil {
			info.Installed = true
			// Check if loaded in launchctl
			out, err := exec.Command("launchctl", "list").Output()
			if err == nil && strings.Contains(string(out), DefaultServiceLabel) {
				info.Active = true
			}
		}
		return info, nil
	}

	if runtime.GOOS == "linux" {
		timerPath := SystemdTimerPath(home)
		info.ConfigPath = timerPath

		if _, err := os.Stat(timerPath); err == nil {
			info.Installed = true
			out, err := exec.Command("systemctl", "--user", "is-active", DefaultSystemdService+".timer").Output()
			if err == nil && strings.TrimSpace(string(out)) == "active" {
				info.Active = true
			}
		}
		return info, nil
	}

	return info, nil
}

// Install configures and starts the background daemon using the native OS service manager.
func Install(binaryPath, baseDir string, homeOverride ...string) (*ServiceInfo, error) {
	home := config.RealHomeDir()
	if len(homeOverride) > 0 && homeOverride[0] != "" {
		home = homeOverride[0]
	}

	if binaryPath == "" {
		self, err := os.Executable()
		if err != nil {
			return nil, fmt.Errorf("could not determine executable path: %w", err)
		}
		binaryPath = self
	}

	logPath := filepath.Join(baseDir, "daemon.log")

	if runtime.GOOS == "darwin" {
		plistPath := LaunchAgentPath(home)
		if err := os.MkdirAll(filepath.Dir(plistPath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create LaunchAgents directory: %w", err)
		}

		pathEnv := os.Getenv("PATH")
		if pathEnv == "" {
			pathEnv = "/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin"
		}

		plistData, err := GeneratePlist(PlistConfig{
			Label:       DefaultServiceLabel,
			BinaryPath:  binaryPath,
			Args:        []string{"daemon", "run"},
			IntervalSec: DefaultIntervalSec,
			LogPath:     logPath,
			EnvVars: map[string]string{
				"PATH": pathEnv,
			},
		})
		if err != nil {
			return nil, err
		}

		// Unload existing service if already registered
		_ = exec.Command("launchctl", "unload", plistPath).Run()

		if err := os.WriteFile(plistPath, plistData, 0644); err != nil {
			return nil, fmt.Errorf("failed to write plist file: %w", err)
		}

		// Load service with launchctl
		cmd := exec.Command("launchctl", "load", plistPath)
		if out, err := cmd.CombinedOutput(); err != nil {
			// Try bootstrap as fallback on modern macOS
			uid := os.Getuid()
			bootCmd := exec.Command("launchctl", "bootstrap", fmt.Sprintf("gui/%d", uid), plistPath)
			if bOut, bErr := bootCmd.CombinedOutput(); bErr != nil {
				return nil, fmt.Errorf("launchctl load failed: %s (%w)", strings.TrimSpace(string(out)+" "+string(bOut)), err)
			}
		}

		_ = AppendDaemonLog(logPath, fmt.Sprintf("daemon service installed (binary: %s, interval: 15m)", binaryPath))
		return Status(baseDir)
	}

	if runtime.GOOS == "linux" {
		servicePath := SystemdServicePath(home)
		timerPath := SystemdTimerPath(home)

		if err := os.MkdirAll(filepath.Dir(servicePath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create systemd user directory: %w", err)
		}

		cfg := SystemdConfig{
			ServiceName: DefaultSystemdService,
			BinaryPath:  binaryPath,
			Args:        []string{"daemon", "run"},
			Interval:    DefaultInterval,
			LogPath:     logPath,
		}

		svcData, err := GenerateSystemdService(cfg)
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(servicePath, svcData, 0644); err != nil {
			return nil, err
		}

		timerData, err := GenerateSystemdTimer(cfg)
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(timerPath, timerData, 0644); err != nil {
			return nil, err
		}

		_ = exec.Command("systemctl", "--user", "daemon-reload").Run()
		if out, err := exec.Command("systemctl", "--user", "enable", "--now", DefaultSystemdService+".timer").CombinedOutput(); err != nil {
			return nil, fmt.Errorf("failed to enable systemd timer: %s (%w)", strings.TrimSpace(string(out)), err)
		}

		_ = AppendDaemonLog(logPath, fmt.Sprintf("daemon service installed (binary: %s, interval: 15m)", binaryPath))
		return Status(baseDir)
	}

	return nil, fmt.Errorf("unsupported operating system for background daemon: %s", runtime.GOOS)
}

// Uninstall stops and removes the background daemon from the host OS service manager.
func Uninstall(baseDir string, homeOverride ...string) error {
	home := config.RealHomeDir()
	if len(homeOverride) > 0 && homeOverride[0] != "" {
		home = homeOverride[0]
	}

	logPath := filepath.Join(baseDir, "daemon.log")

	if runtime.GOOS == "darwin" {
		plistPath := LaunchAgentPath(home)
		// Try unloading via launchctl
		_ = exec.Command("launchctl", "unload", plistPath).Run()
		uid := os.Getuid()
		_ = exec.Command("launchctl", "bootout", fmt.Sprintf("gui/%d", uid), plistPath).Run()

		if err := os.Remove(plistPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove plist file: %w", err)
		}

		_ = AppendDaemonLog(logPath, "daemon service uninstalled")
		return nil
	}

	if runtime.GOOS == "linux" {
		_ = exec.Command("systemctl", "--user", "disable", "--now", DefaultSystemdService+".timer").Run()
		_ = os.Remove(SystemdTimerPath(home))
		_ = os.Remove(SystemdServicePath(home))
		_ = exec.Command("systemctl", "--user", "daemon-reload").Run()

		_ = AppendDaemonLog(logPath, "daemon service uninstalled")
		return nil
	}

	return nil
}

// RunOnce executes all scheduled periodic tasks (e.g. refreshing quota caches across all profiles).
func RunOnce(ctx context.Context, reg *agents.Registry, pm *profile.ProfileManager, baseDir string) error {
	if reg == nil || pm == nil {
		return fmt.Errorf("registry or profile manager is nil")
	}

	logPath := filepath.Join(baseDir, "daemon.log")
	cfg, _ := config.LoadConfig()

	var targets []usage.TargetProfile
	adapters := reg.All()
	for _, ad := range adapters {
		profs, _ := pm.ListProfilesForAgent(ad.Name(), cfg, reg)
		var usageFn func(context.Context, string, string) (*usage.Report, error)
		if up, ok := ad.(agents.UsageProvider); ok {
			usageFn = up.GetUsage
		}
		for _, p := range profs {
			targets = append(targets, usage.TargetProfile{
				Agent:      ad.Name(),
				Profile:    p,
				ProfileDir: pm.ProfileDir(p),
				GetUsageFn: usageFn,
			})
		}
	}

	if len(targets) == 0 {
		_ = AppendDaemonLog(logPath, "daemon run: 0 configured profiles found")
		return nil
	}

	cache := usage.NewCacheStore(baseDir, usage.DefaultTTL)
	reportsChan := usage.RefreshAsync(ctx, targets, cache, true)

	successCount := 0
	errorCount := 0
	for r := range reportsChan {
		if r.Error != "" {
			errorCount++
		} else {
			successCount++
		}
	}

	msg := fmt.Sprintf("daemon run: quota refresh completed (%d updated, %d errors out of %d targets)",
		successCount, errorCount, len(targets))
	_ = AppendDaemonLog(logPath, msg)
	return nil
}
