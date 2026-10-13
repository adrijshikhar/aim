package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
)

func TestWebCmd_Flags(t *testing.T) {
	reg := defaultRegistry()
	pm := profile.NewProfileManager(t.TempDir())
	cmd := newWebCmd(reg, pm)

	// Verify flag defaults
	portFlag := cmd.Flag("port")
	if portFlag == nil || portFlag.DefValue != "18480" {
		t.Fatalf("expected default port 18480, got %v", portFlag)
	}

	devFlag := cmd.Flag("dev")
	if devFlag == nil || devFlag.DefValue != "false" {
		t.Fatalf("expected default dev false, got %v", devFlag)
	}

	noOpenFlag := cmd.Flag("no-open")
	if noOpenFlag == nil || noOpenFlag.DefValue != "false" {
		t.Fatalf("expected default no-open false, got %v", noOpenFlag)
	}

	fgFlag := cmd.Flag("foreground")
	if fgFlag == nil || fgFlag.DefValue != "false" {
		t.Fatalf("expected default foreground false, got %v", fgFlag)
	}
}

func TestWebCmd_Help(t *testing.T) {
	buf := new(bytes.Buffer)
	reg := defaultRegistry()
	pm := profile.NewProfileManager(t.TempDir())
	root := newRootCmd(reg, pm)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"web", "--help"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error executing web --help: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Start the AIM web dashboard") {
		t.Fatalf("expected help description in output, got:\n%s", out)
	}
	if !strings.Contains(out, "--port") || !strings.Contains(out, "--dev") || !strings.Contains(out, "--no-open") || !strings.Contains(out, "--foreground") {
		t.Fatalf("expected flags in help output, got:\n%s", out)
	}
	if !strings.Contains(out, "18480") {
		t.Fatalf("expected port 18480 in help, got:\n%s", out)
	}
}

func TestWebCmd_ExecutionForeground(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("AIM_HOME", tmpDir)

	reg := defaultRegistry()
	pm := profile.NewProfileManager(config.BaseDir())
	cmd := newWebCmd(reg, pm)

	// Mock browser open to avoid opening a real browser in tests
	browserOpened := false
	openBrowserURL = func(url string) error {
		browserOpened = true
		return nil
	}
	defer func() {
		openBrowserURL = func(url string) error { return nil }
	}()

	// Use an ephemeral port in foreground mode
	ctx, cancel := context.WithCancel(context.Background())
	cmd.SetArgs([]string{"--foreground", "--port", "0", "--no-open"})

	done := make(chan error, 1)
	go func() {
		done <- cmd.ExecuteContext(ctx)
	}()

	// Allow server to start listening and write PID
	time.Sleep(150 * time.Millisecond)

	pidPath := filepath.Join(config.BaseDir(), "web.pid")
	if _, err := os.Stat(pidPath); err != nil {
		t.Errorf("expected PID file at %s during foreground execution, got: %v", pidPath, err)
	}

	// Trigger shutdown
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("webCmd failed: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("webCmd did not shut down within timeout")
	}

	if browserOpened {
		t.Errorf("expected browser not to open with --no-open")
	}

	// PID file should be cleaned up
	if _, err := os.Stat(pidPath); !os.IsNotExist(err) {
		t.Errorf("expected PID file to be removed after shutdown")
	}
}

func TestWebCmd_StatusAndStop(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("AIM_HOME", tmpDir)

	// 1. Status when stopped
	buf := new(bytes.Buffer)
	cmdStatus := newWebStatusCmd()
	cmdStatus.SetOut(buf)
	if err := cmdStatus.Execute(); err != nil {
		t.Fatalf("status cmd failed: %v", err)
	}
	// Note: cmd prints with fmt.Println, verify no error returned

	// 2. Stop when stopped
	cmdStop := newWebStopCmd()
	if err := cmdStop.Execute(); err != nil {
		t.Fatalf("stop cmd failed: %v", err)
	}

	// 3. Fake PID writing and reading
	if err := writeWebPID(999999, 18480); err != nil {
		t.Fatalf("failed to write web PID: %v", err)
	}
	info, err := readWebPID()
	if err != nil || info == nil || info.Port != 18480 || info.PID != 999999 {
		t.Fatalf("unexpected PID info: %+v, err: %v", info, err)
	}

	// 4. Stop cleans up PID
	if err := cmdStop.Execute(); err != nil {
		t.Fatalf("stop cmd on dead pid failed: %v", err)
	}
	if _, err := readWebPID(); err == nil {
		t.Fatalf("expected PID file to be removed after stop on dead pid")
	}
}
