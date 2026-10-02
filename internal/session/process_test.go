package session_test

import (
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/session"
)

func TestParseProcessLines(t *testing.T) {
	sampleOutput := `
  PID COMMAND
21484 aim run agy bby --dangerously-skip-permissions --conversation=775e6ada-1595-4e7e-84fa-ce0ea71e3007
21485 /Users/nemesis/.local/bin/agy --dangerously-skip-permissions --conversation=775e6ada-1595-4e7e-84fa-ce0ea71e3007
92974 aim run agy work --dangerously-skip-permissions --conversation=fcdbc2e0-2dc8-4ffa-9ee2-eb5aaa3e556f
33523 agy --dangerously-skip-permissions --conversation=c49a97be-0342-43cc-a90f-85a359959aaf
74382 codex --yolo resume 01a09eb7-2f6c-7c52-895f-218f9ac9eecd
55746 /Applications/ChatGPT.app/Contents/Resources/codex app-server --listen stdio://
21081 aim run codex office --yolo resume 01a0b351-2f2f-7d22-8176-49e45bde8f9b
66778 aim run claude work --resume 55555555-6666-7777-8888-999999999999
`
	active := session.ParseProcessOutput(strings.NewReader(sampleOutput))
	if len(active) == 0 {
		t.Fatal("expected active processes to be found, got none")
	}

	// Verify aim run agy bby
	info1, ok := active["775e6ada-1595-4e7e-84fa-ce0ea71e3007"]
	if !ok {
		t.Errorf("expected 775e6ada-1595-4e7e-84fa-ce0ea71e3007 to be active")
	} else {
		if info1.Agent != "agy" {
			t.Errorf("expected agent 'agy', got %q", info1.Agent)
		}
		if info1.Profile != "bby" {
			t.Errorf("expected profile 'bby', got %q", info1.Profile)
		}
		if info1.PID != 21484 && info1.PID != 21485 {
			t.Errorf("unexpected PID %d", info1.PID)
		}
	}

	// Verify host native agy run
	info2, ok := active["c49a97be-0342-43cc-a90f-85a359959aaf"]
	if !ok {
		t.Errorf("expected c49a97be-0342-43cc-a90f-85a359959aaf to be active")
	} else {
		if info2.Agent != "agy" {
			t.Errorf("expected agent 'agy', got %q", info2.Agent)
		}
		if info2.Profile != "<host>" {
			t.Errorf("expected profile '<host>', got %q", info2.Profile)
		}
		if info2.PID != 33523 {
			t.Errorf("expected PID 33523, got %d", info2.PID)
		}
	}

	// Verify codex resume
	info3, ok := active["01a09eb7-2f6c-7c52-895f-218f9ac9eecd"]
	if !ok {
		t.Errorf("expected 01a09eb7-2f6c-7c52-895f-218f9ac9eecd to be active")
	} else {
		if info3.Agent != "codex" {
			t.Errorf("expected agent 'codex', got %q", info3.Agent)
		}
		if info3.PID != 74382 {
			t.Errorf("expected PID 74382, got %d", info3.PID)
		}
	}

	// Verify aim run codex office resume
	info4, ok := active["01a0b351-2f2f-7d22-8176-49e45bde8f9b"]
	if !ok {
		t.Errorf("expected 01a0b351-2f2f-7d22-8176-49e45bde8f9b to be active")
	} else {
		if info4.Agent != "codex" {
			t.Errorf("expected agent 'codex', got %q", info4.Agent)
		}
		if info4.Profile != "office" {
			t.Errorf("expected profile 'office', got %q", info4.Profile)
		}
		if info4.PID != 21081 {
			t.Errorf("expected PID 21081, got %d", info4.PID)
		}
	}

	// Verify aim run claude work --resume
	info5, ok := active["55555555-6666-7777-8888-999999999999"]
	if !ok {
		t.Errorf("expected 55555555-6666-7777-8888-999999999999 to be active")
	} else {
		if info5.Agent != "claude" {
			t.Errorf("expected agent 'claude', got %q", info5.Agent)
		}
		if info5.Profile != "work" {
			t.Errorf("expected profile 'work', got %q", info5.Profile)
		}
		if info5.PID != 66778 {
			t.Errorf("expected PID 66778, got %d", info5.PID)
		}
	}
}

func TestGracefulTerminate_ValidProcess(t *testing.T) {
	cmd := exec.Command("sleep", "10")
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start dummy process: %v", err)
	}
	pid := cmd.Process.Pid

	err := session.GracefulTerminate(pid, 2*time.Second)
	if err != nil {
		t.Fatalf("expected graceful termination, got error: %v", err)
	}

	// Verify process is no longer running
	if err := syscall.Kill(pid, 0); err == nil {
		t.Fatalf("process %d is still running after GracefulTerminate", pid)
	}
}

func TestGracefulTerminate_InvalidPID(t *testing.T) {
	testCases := []int{0, -1, -100}
	for _, pid := range testCases {
		err := session.GracefulTerminate(pid, 100*time.Millisecond)
		if err == nil {
			t.Errorf("expected error for invalid PID %d, got nil", pid)
		} else if !strings.Contains(err.Error(), "invalid PID") {
			t.Errorf("expected 'invalid PID' in error, got %v", err)
		}
	}
}

func TestGracefulTerminate_AlreadyDeadPID(t *testing.T) {
	cmd := exec.Command("sleep", "0.01")
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start dummy process: %v", err)
	}
	pid := cmd.Process.Pid
	_ = cmd.Wait()

	// Terminating an already dead process should succeed cleanly (return nil)
	err := session.GracefulTerminate(pid, 500*time.Millisecond)
	if err != nil {
		t.Fatalf("expected nil for already dead process, got %v", err)
	}

	// Completely non-existent PID should also return nil
	nonExistentPID := 9999999
	err = session.GracefulTerminate(nonExistentPID, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("expected nil for non-existent PID %d, got %v", nonExistentPID, err)
	}
}

func TestGracefulTerminate_TimeoutFallbackSIGKILL(t *testing.T) {
	// Process ignores SIGTERM
	cmd := exec.Command("sh", "-c", "trap '' TERM; while true; do sleep 1; done")
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start process: %v", err)
	}
	pid := cmd.Process.Pid
	defer func() {
		_ = syscall.Kill(pid, syscall.SIGKILL)
		_ = cmd.Wait()
	}()

	// Timeout should expire, sending SIGKILL
	err := session.GracefulTerminate(pid, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("expected success on SIGKILL fallback, got error: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	if err := syscall.Kill(pid, 0); err == nil {
		t.Fatalf("process %d is still running after SIGKILL fallback", pid)
	}
}

func TestGracefulTerminate_NonChildProcess(t *testing.T) {
	// Spawn an orphaned process reparented away from the current test process
	cmd := exec.Command("sh", "-c", "sleep 10 >/dev/null 2>&1 & echo $!")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("failed to spawn orphaned process: %v", err)
	}

	pidStr := strings.TrimSpace(string(out))
	pid, err := strconv.Atoi(pidStr)
	if err != nil || pid <= 0 {
		t.Fatalf("invalid spawned PID: %q", pidStr)
	}

	defer func() {
		_ = syscall.Kill(pid, syscall.SIGKILL)
	}()

	// Verify process is initially running
	if err := syscall.Kill(pid, 0); err != nil {
		t.Fatalf("spawned background process %d is not running: %v", pid, err)
	}

	// GracefulTerminate on non-child process
	err = session.GracefulTerminate(pid, 2*time.Second)
	if err != nil {
		t.Fatalf("expected graceful termination of non-child process, got: %v", err)
	}

	// Verify process is dead
	if err := syscall.Kill(pid, 0); err == nil {
		t.Fatalf("non-child process %d is still running after GracefulTerminate", pid)
	}
}
