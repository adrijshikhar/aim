package runner

import (
	"context"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/charmbracelet/x/term"
)

func TestSetupSignalForwarding_GoroutineLeak(t *testing.T) {
	// Settle goroutines before taking baseline
	time.Sleep(50 * time.Millisecond)
	baseline := runtime.NumGoroutine()

	// Spin up and tear down forwarding multiple times
	for i := 0; i < 20; i++ {
		isTerm := (i%2 == 0)
		cleanup := setupSignalForwarding(nil, isTerm)
		cleanup()
	}

	// Wait up to 1 second for background goroutines to terminate
	deadline := time.Now().Add(1 * time.Second)
	leaked := true
	for time.Now().Before(deadline) {
		runtime.Gosched()
		time.Sleep(10 * time.Millisecond)
		if runtime.NumGoroutine() <= baseline {
			leaked = false
			break
		}
	}

	if leaked {
		t.Errorf("goroutine leak detected: baseline %d, current %d", baseline, runtime.NumGoroutine())
	}
}

func TestSetupSignalForwarding_IdempotentCleanup(t *testing.T) {
	cleanup := setupSignalForwarding(nil, true)
	// Calling cleanup multiple times must not panic
	cleanup()
	cleanup()
	cleanup()
}

func TestSetupSignalForwarding_NilProcess(t *testing.T) {
	testSig := make(chan os.Signal, 1)
	signal.Notify(testSig, syscall.SIGHUP)
	defer func() {
		signal.Stop(testSig)
		signal.Reset(syscall.SIGHUP)
	}()

	cleanup := setupSignalForwarding(nil, true)
	defer cleanup()

	// Sending SIGHUP when proc is nil must not panic
	_ = syscall.Kill(os.Getpid(), syscall.SIGHUP)
	select {
	case <-testSig:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for SIGHUP")
	}
}

func TestSetupSignalForwarding_NonTerminal_ForwardsSIGINT(t *testing.T) {
	cmd := exec.Command("sleep", "10")
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start sleep process: %v", err)
	}
	defer func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	}()

	// Register test signal channel to prevent SIGINT from killing test runner
	testSig := make(chan os.Signal, 1)
	signal.Notify(testSig, os.Interrupt)
	defer func() {
		signal.Stop(testSig)
		signal.Reset(os.Interrupt)
	}()

	cleanup := setupSignalForwarding(cmd.Process, false)
	defer cleanup()

	// Send SIGINT to self
	if err := syscall.Kill(os.Getpid(), syscall.SIGINT); err != nil {
		t.Fatalf("failed to send SIGINT: %v", err)
	}

	select {
	case <-testSig:
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for self SIGINT")
	}

	// Verify child process terminates from forwarded SIGINT
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected child process to exit with signal error, got nil")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("child process did not terminate after forwarded SIGINT")
	}
}

func TestSetupSignalForwarding_Terminal_DoesNotForwardSIGINT(t *testing.T) {
	cmd := exec.Command("sleep", "10")
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start sleep process: %v", err)
	}
	defer func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	}()

	// Register test signal channel to prevent SIGINT from killing test runner
	testSig := make(chan os.Signal, 1)
	signal.Notify(testSig, os.Interrupt)
	defer func() {
		signal.Stop(testSig)
		signal.Reset(os.Interrupt)
	}()

	cleanup := setupSignalForwarding(cmd.Process, true)
	defer cleanup()

	// Send SIGINT to self
	if err := syscall.Kill(os.Getpid(), syscall.SIGINT); err != nil {
		t.Fatalf("failed to send SIGINT: %v", err)
	}

	select {
	case <-testSig:
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for self SIGINT")
	}

	// Give time for any unexpected forwarding
	time.Sleep(100 * time.Millisecond)

	// Child process should STILL be running because SIGINT was NOT forwarded
	if err := cmd.Process.Signal(syscall.Signal(0)); err != nil {
		t.Fatalf("child process died unexpectedly (SIGINT was incorrectly forwarded): %v", err)
	}
}

func TestSetupSignalForwarding_Terminal_ForwardsSIGTERM(t *testing.T) {
	cmd := exec.Command("sleep", "10")
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start sleep process: %v", err)
	}
	defer func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	}()

	testSig := make(chan os.Signal, 1)
	signal.Notify(testSig, syscall.SIGTERM)
	defer func() {
		signal.Stop(testSig)
		signal.Reset(syscall.SIGTERM)
	}()

	cleanup := setupSignalForwarding(cmd.Process, true)
	defer cleanup()

	// Send SIGTERM to self
	if err := syscall.Kill(os.Getpid(), syscall.SIGTERM); err != nil {
		t.Fatalf("failed to send SIGTERM: %v", err)
	}

	select {
	case <-testSig:
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for self SIGTERM")
	}

	// Verify child process terminates from forwarded SIGTERM
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected child process to exit with error, got nil")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("child process did not terminate after forwarded SIGTERM")
	}
}

func TestSetupSignalForwarding_Terminal_ForwardsSIGHUP(t *testing.T) {
	cmd := exec.Command("sleep", "10")
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start sleep process: %v", err)
	}
	defer func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	}()

	testSig := make(chan os.Signal, 1)
	signal.Notify(testSig, syscall.SIGHUP)
	defer func() {
		signal.Stop(testSig)
		signal.Reset(syscall.SIGHUP)
	}()

	cleanup := setupSignalForwarding(cmd.Process, true)
	defer cleanup()

	// Send SIGHUP to self
	if err := syscall.Kill(os.Getpid(), syscall.SIGHUP); err != nil {
		t.Fatalf("failed to send SIGHUP: %v", err)
	}

	select {
	case <-testSig:
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for self SIGHUP")
	}

	// Verify child process terminates from forwarded SIGHUP
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected child process to exit with error, got nil")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("child process did not terminate after forwarded SIGHUP")
	}
}

func TestTerminalStateRestoration(t *testing.T) {
	// Test that Run executes cleanly and defer restore does not panic
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("sh not available")
	}

	r := NewRunner()
	env := agents.LaunchEnv{
		BinaryPath: sh,
	}

	code, err := r.Run(context.Background(), env, []string{"-c", "exit 0"})
	if err != nil || code != 0 {
		t.Fatalf("expected code 0, got %d, err: %v", code, err)
	}

	// Verify pipe is recognized as non-terminal
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	defer pr.Close()
	defer pw.Close()

	if term.IsTerminal(pr.Fd()) {
		t.Error("expected pipe read end to not be a terminal")
	}

	// term.GetState on a non-terminal should fail gracefully
	if _, err := term.GetState(pr.Fd()); err == nil {
		t.Error("expected GetState on pipe to return an error")
	}

	// If Stdin is a terminal, verify GetState and Restore work properly
	if term.IsTerminal(os.Stdin.Fd()) {
		state, err := term.GetState(os.Stdin.Fd())
		if err != nil {
			t.Fatalf("failed to get state of terminal: %v", err)
		}
		if state == nil {
			t.Fatal("expected non-nil terminal state")
		}
		if err := term.Restore(os.Stdin.Fd(), state); err != nil {
			t.Fatalf("failed to restore terminal state: %v", err)
		}
	}
}
