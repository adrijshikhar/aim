package main

import (
	"bytes"
	"context"
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
	if portFlag == nil || portFlag.DefValue != "8080" {
		t.Fatalf("expected default port 8080, got %v", portFlag)
	}

	devFlag := cmd.Flag("dev")
	if devFlag == nil || devFlag.DefValue != "false" {
		t.Fatalf("expected default dev false, got %v", devFlag)
	}

	noOpenFlag := cmd.Flag("no-open")
	if noOpenFlag == nil || noOpenFlag.DefValue != "false" {
		t.Fatalf("expected default no-open false, got %v", noOpenFlag)
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
	if !strings.Contains(out, "--port") || !strings.Contains(out, "--dev") || !strings.Contains(out, "--no-open") {
		t.Fatalf("expected flags in help output, got:\n%s", out)
	}
}

func TestWebCmd_Execution(t *testing.T) {
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

	// Use an ephemeral port or cancelled context to test clean startup and shutdown
	ctx, cancel := context.WithCancel(context.Background())
	cmd.SetArgs([]string{"--port", "0", "--no-open"})

	done := make(chan error, 1)
	go func() {
		done <- cmd.ExecuteContext(ctx)
	}()

	// Allow server to start listening
	time.Sleep(100 * time.Millisecond)

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
}
