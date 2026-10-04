package daemon

import (
	"context"
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/usage"
)

func TestGeneratePlist(t *testing.T) {
	cfg := PlistConfig{
		Label:       DefaultServiceLabel,
		BinaryPath:  "/usr/local/bin/aim",
		Args:        []string{"daemon", "run"},
		IntervalSec: 900,
		LogPath:     "/Users/test/.aim/daemon.log",
		EnvVars: map[string]string{
			"PATH": "/opt/homebrew/bin:/usr/bin:/bin",
		},
	}

	data, err := GeneratePlist(cfg)
	if err != nil {
		t.Fatalf("GeneratePlist failed: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "<string>dev.aim-cli.daemon</string>") {
		t.Errorf("expected plist to contain service label, got:\n%s", content)
	}
	if !strings.Contains(content, "<string>/usr/local/bin/aim</string>") {
		t.Errorf("expected plist to contain binary path, got:\n%s", content)
	}
	if !strings.Contains(content, "<string>daemon</string>") || !strings.Contains(content, "<string>run</string>") {
		t.Errorf("expected plist to contain daemon run args, got:\n%s", content)
	}
	if !strings.Contains(content, "<key>StartInterval</key>") || !strings.Contains(content, "<integer>900</integer>") {
		t.Errorf("expected plist to specify StartInterval 900, got:\n%s", content)
	}
	if !strings.Contains(content, "<key>RunAtLoad</key>\n\t<true/>") && !strings.Contains(content, "<key>RunAtLoad</key><true/>") && !strings.Contains(content, "<key>RunAtLoad</key>") {
		t.Errorf("expected plist to specify RunAtLoad, got:\n%s", content)
	}
	if !strings.Contains(content, "<string>/Users/test/.aim/daemon.log</string>") {
		t.Errorf("expected plist to specify log path, got:\n%s", content)
	}

	// Verify XML well-formedness
	var parsed struct {
		XMLName xml.Name `xml:"plist"`
	}
	if err := xml.Unmarshal(data, &parsed); err != nil {
		t.Errorf("generated plist is not valid XML: %v", err)
	}
}

func TestGenerateSystemd(t *testing.T) {
	cfg := SystemdConfig{
		ServiceName: "aim-daemon",
		BinaryPath:  "/usr/local/bin/aim",
		Args:        []string{"daemon", "run"},
		Interval:    15 * time.Minute,
	}

	serviceData, err := GenerateSystemdService(cfg)
	if err != nil {
		t.Fatalf("GenerateSystemdService failed: %v", err)
	}
	serviceStr := string(serviceData)
	if !strings.Contains(serviceStr, "ExecStart=/usr/local/bin/aim daemon run") {
		t.Errorf("expected ExecStart with daemon run, got:\n%s", serviceStr)
	}

	timerData, err := GenerateSystemdTimer(cfg)
	if err != nil {
		t.Fatalf("GenerateSystemdTimer failed: %v", err)
	}
	timerStr := string(timerData)
	if !strings.Contains(timerStr, "OnUnitActiveSec=15m") {
		t.Errorf("expected timer to have OnUnitActiveSec=15m, got:\n%s", timerStr)
	}
	if !strings.Contains(timerStr, "Unit=aim-daemon.service") {
		t.Errorf("expected timer to specify Unit=aim-daemon.service, got:\n%s", timerStr)
	}
}

func TestStatus_NotInstalled(t *testing.T) {
	tmpDir := t.TempDir()
	info, err := Status(tmpDir, tmpDir)
	if err != nil {
		t.Fatalf("Status returned unexpected error: %v", err)
	}
	if info.Installed {
		t.Errorf("expected Installed to be false for empty directory")
	}
	if info.Active {
		t.Errorf("expected Active to be false for empty directory")
	}
}

func TestLogExecution(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "daemon.log")

	if err := AppendDaemonLog(logPath, "quota refresh completed for 2 profiles"); err != nil {
		t.Fatalf("AppendDaemonLog failed: %v", err)
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read daemon log: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "quota refresh completed for 2 profiles") {
		t.Errorf("expected log to contain message, got: %s", content)
	}
	// Verify timestamp exists
	if !strings.Contains(content, "[") || !strings.Contains(content, "]") {
		t.Errorf("expected timestamp bracket in log, got: %s", content)
	}
}

func TestReadLastRunFromLog(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "daemon.log")

	// Empty / nonexistent log
	lastRun, msg := ReadLastDaemonRun(logPath)
	if !lastRun.IsZero() {
		t.Errorf("expected zero time for nonexistent log")
	}
	if msg != "" {
		t.Errorf("expected empty message for nonexistent log")
	}

	// Add entry
	_ = AppendDaemonLog(logPath, "first run")
	time.Sleep(10 * time.Millisecond)
	_ = AppendDaemonLog(logPath, "second run completed successfully")

	lastRun, msg = ReadLastDaemonRun(logPath)
	if lastRun.IsZero() {
		t.Errorf("expected non-zero time for populated log")
	}
	if !strings.Contains(msg, "second run completed successfully") {
		t.Errorf("expected last line message, got: %q", msg)
	}
}

type mockUsageAdapter struct {
	agents.BaseAdapter
}

func (m *mockUsageAdapter) HasCredentials(profileDir string) bool { return true }
func (m *mockUsageAdapter) PrepareEnv(ctx context.Context, profileName, profileDir string) (agents.LaunchEnv, error) {
	return agents.LaunchEnv{}, nil
}
func (m *mockUsageAdapter) GetUsage(ctx context.Context, profile, profileDir string) (*usage.Report, error) {
	return &usage.Report{
		Agent:     m.Name(),
		Profile:   profile,
		FetchedAt: time.Now(),
		Status:    usage.StatusOK,
		Windows: []usage.LimitWindow{
			{Name: "5-hour limit", RemainingPct: 75},
		},
	}, nil
}

func TestRunOnce_NilSafety(t *testing.T) {
	if err := RunOnce(context.Background(), nil, nil, t.TempDir()); err == nil {
		t.Errorf("expected error for nil registry and pm")
	}
}

func TestRunOnce_Success(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("AIM_HOME", tmpDir)
	reg := agents.NewRegistry()
	ad := &mockUsageAdapter{
		BaseAdapter: agents.NewBaseAdapter("mock", "Mock Agent", "mock", nil),
	}
	reg.Register(ad)

	pm := profile.NewProfileManager(tmpDir)
	if _, err := pm.EnsureProfile("work"); err != nil {
		t.Fatalf("EnsureProfile failed: %v", err)
	}

	err := RunOnce(context.Background(), reg, pm, tmpDir)
	if err != nil {
		t.Fatalf("RunOnce failed: %v", err)
	}

	// Verify log was appended
	logPath := filepath.Join(tmpDir, "daemon.log")
	lastRun, msg := ReadLastDaemonRun(logPath)
	if lastRun.IsZero() {
		t.Errorf("expected last run timestamp in daemon.log")
	}
	if !strings.Contains(msg, "quota refresh completed") {
		t.Errorf("expected log message to contain 'quota refresh completed', got: %s", msg)
	}

	// Verify cache store contains report
	cache := usage.NewCacheStore(tmpDir, usage.DefaultTTL)
	rep, ok := cache.Get("mock", "work")
	if !ok {
		t.Fatalf("expected cached report for mock:work, but not found")
	}
	if rep.Status != usage.StatusOK || len(rep.Windows) == 0 {
		t.Errorf("cached report does not match expected: %+v", rep)
	}
}
