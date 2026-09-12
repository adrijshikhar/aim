package runner

import (
	"context"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/agents"
)

func TestRunnerExecute(t *testing.T) {
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("sh not available")
	}

	r := NewRunner()
	env := agents.LaunchEnv{
		BinaryPath: sh,
		Env: map[string]string{
			"TEST_VAR": "aim_test_val",
		},
	}

	code, err := r.Run(context.Background(), env, []string{"-c", "exit 0"})
	if err != nil || code != 0 {
		t.Fatalf("expected exit code 0, got %d, err: %v", code, err)
	}

	codeFail, _ := r.Run(context.Background(), env, []string{"-c", "exit 42"})
	if codeFail != 42 {
		t.Errorf("expected exit code 42, got %d", codeFail)
	}
}

func TestRunnerExecute_SignalExitCode(t *testing.T) {
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("sh not available")
	}

	r := NewRunner()
	env := agents.LaunchEnv{
		BinaryPath: sh,
	}

	// Terminating via SIGTERM should yield exit code 128 + 15 = 143
	code, err := r.Run(context.Background(), env, []string{"-c", "kill -TERM $$"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := 128 + int(syscall.SIGTERM)
	if code != expected {
		t.Errorf("expected exit code %d (128+SIGTERM), got %d", expected, code)
	}
}

func TestRunnerExecute_NonExistentBinary(t *testing.T) {
	r := NewRunner()
	env := agents.LaunchEnv{
		BinaryPath: "/path/to/nonexistent_binary_xyz_12345",
	}

	code, err := r.Run(context.Background(), env, nil)
	if err == nil {
		t.Fatalf("expected error for nonexistent binary, got nil")
	}
	if code != 1 {
		t.Errorf("expected exit code 1 for non-ExitError failure, got %d", code)
	}
}

func TestRunnerExecute_EnvAndArgs(t *testing.T) {
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("sh not available")
	}

	r := NewRunner()
	env := agents.LaunchEnv{
		BinaryPath: sh,
		Args:       []string{"-c"},
		Env: map[string]string{
			"TEST_FOO": "bar123",
		},
	}

	// Verify extraArgs are appended and env is available in child process
	code, err := r.Run(context.Background(), env, []string{`[ "$TEST_FOO" = "bar123" ] && exit 0 || exit 1`})
	if err != nil || code != 0 {
		t.Fatalf("expected exit code 0 with env propagated, got %d, err: %v", code, err)
	}
}

func TestRunnerRunShell(t *testing.T) {
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("sh not available")
	}

	// Override SHELL for this test
	origShell := os.Getenv("SHELL")
	defer os.Setenv("SHELL", origShell)
	os.Setenv("SHELL", sh)

	r := NewRunner()
	launch := agents.LaunchEnv{
		Env: map[string]string{
			"AIM_AGENT":   "testagent",
			"AIM_PROFILE": "testprofile",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// With stdin connected (closed/EOF in test runner), sh exits immediately with 0
	code, err := r.RunShell(ctx, launch)
	if err != nil {
		t.Fatalf("RunShell failed: %v", err)
	}
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	if launch.Env["PS1"] != "[aim:testagent:testprofile] $ " {
		t.Errorf("expected PS1 to be set, got %q", launch.Env["PS1"])
	}
}

func TestSetupWindowSizeListener(t *testing.T) {
	resized := false
	cleanup := SetupWindowSizeListener(func() {
		resized = true
	})
	if cleanup == nil {
		t.Fatal("expected non-nil cleanup function")
	}
	// Calling cleanup should stop the listener cleanly
	cleanup()
	if resized {
		t.Log("resize handler was invoked")
	}
}

func TestSetupSignalForwarding(t *testing.T) {
	cmd := exec.Command("sleep", "1")
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start sleep: %v", err)
	}
	cleanup := setupSignalForwarding(cmd.Process)
	if cleanup == nil {
		_ = cmd.Process.Kill()
		t.Fatal("expected non-nil cleanup function")
	}
	_ = cmd.Process.Kill()
	_ = cmd.Wait()
	cleanup()
}

func TestRunnerExecute_SSHEnvFiltering(t *testing.T) {
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("sh not available")
	}

	// Simulate host having SSH connection environment variables
	t.Setenv("SSH_CONNECTION", "192.168.1.100 54321 192.168.1.50 22")
	t.Setenv("SSH_CLIENT", "192.168.1.100 54321 22")
	t.Setenv("SSH_TTY", "/dev/ttys001")
	t.Setenv("GEMINI_CLI_HOME", "/some/unwanted/dir")

	r := NewRunner()

	// Scenario 1: Unauthenticated session where SSH_CONNECTION is omitted to enable browser auto-open.
	// Verify that host SSH variables are filtered out and not inherited by child process.
	envUnauth := agents.LaunchEnv{
		BinaryPath: sh,
		Args:       []string{"-c"},
		Env: map[string]string{
			"HOME": t.TempDir(),
		},
	}
	checkUnauthScript := `
		if [ -n "$SSH_CONNECTION" ] || [ -n "$SSH_CLIENT" ] || [ -n "$SSH_TTY" ] || [ -n "$GEMINI_CLI_HOME" ]; then
			exit 1
		fi
		exit 0
	`
	code, err := r.Run(context.Background(), envUnauth, []string{checkUnauthScript})
	if err != nil || code != 0 {
		t.Fatalf("expected exit code 0 (SSH variables suppressed), got %d, err: %v", code, err)
	}

	// Scenario 2: Authenticated session where SSH_CONNECTION is explicitly provided in launch.Env for keyring bypass.
	// Verify child process receives SSH_CONNECTION, but still suppresses SSH_CLIENT and SSH_TTY.
	envAuth := agents.LaunchEnv{
		BinaryPath: sh,
		Args:       []string{"-c"},
		Env: map[string]string{
			"HOME":           t.TempDir(),
			"SSH_CONNECTION": "127.0.0.1 50000 127.0.0.1 22",
		},
	}
	checkAuthScript := `
		if [ "$SSH_CONNECTION" != "127.0.0.1 50000 127.0.0.1 22" ]; then
			exit 2
		fi
		if [ -n "$SSH_CLIENT" ] || [ -n "$SSH_TTY" ] || [ -n "$GEMINI_CLI_HOME" ]; then
			exit 3
		fi
		exit 0
	`
	codeAuth, err := r.Run(context.Background(), envAuth, []string{checkAuthScript})
	if err != nil || codeAuth != 0 {
		t.Fatalf("expected exit code 0 (SSH_CONNECTION preserved, others filtered), got %d, err: %v", codeAuth, err)
	}
}
