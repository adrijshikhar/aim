package session

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/aim-cli/aim/internal/logger"
)

var (
	// Matches: aim run <agent> <profile> ... (--conversation[= ]<id> | --resume[= ]<id> | resume <id>)
	aimRunConvRegex = regexp.MustCompile(`aim\s+run\s+([a-zA-Z0-9_-]+)\s+([a-zA-Z0-9_-]+).*?(?:--conversation[=\s]|--resume[=\s]|resume\s+)([a-zA-Z0-9_-]+)`)
	// Matches: agy ... --conversation[= ]<id>
	agyConvRegex = regexp.MustCompile(`(?:^|/|\s)agy(?:\.exe)?\s+.*?--conversation[=\s]([a-zA-Z0-9_-]+)`)
	// Matches: codex ... resume <id>
	codexResumeRegex = regexp.MustCompile(`(?:^|/|\s)codex(?:\.exe)?\s+.*?resume\s+([a-zA-Z0-9_-]+)`)
	// Matches: claude ... --resume[= ]<id>
	claudeResumeRegex = regexp.MustCompile(`(?:^|/|\s)claude(?:\.exe)?\s+.*?--resume[=\s]([a-zA-Z0-9_-]+)`)
)

// DefaultProcessScanner scans local OS processes via `ps -eo pid,command`.
type DefaultProcessScanner struct{}

func NewDefaultProcessScanner() *DefaultProcessScanner {
	return &DefaultProcessScanner{}
}

func (s *DefaultProcessScanner) ScanActiveProcesses(ctx context.Context) (map[string]ActiveProcessInfo, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctxTimeout, "ps", "-eo", "pid,command")
	out, err := cmd.Output()
	if err != nil {
		return map[string]ActiveProcessInfo{}, nil
	}

	return ParseProcessOutput(bytes.NewReader(out)), nil
}

// ParseProcessOutput parses tabular `ps -eo pid,command` output.
func ParseProcessOutput(r io.Reader) map[string]ActiveProcessInfo {
	result := make(map[string]ActiveProcessInfo)
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "PID") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		pid, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}

		cmdline := strings.Join(fields[1:], " ")

		// 1. Check for `aim run <agent> <profile> ... --conversation=<id>`
		if matches := aimRunConvRegex.FindStringSubmatch(cmdline); len(matches) >= 4 {
			agent := matches[1]
			profile := matches[2]
			convID := matches[3]
			result[convID] = ActiveProcessInfo{
				PID:            pid,
				Agent:          agent,
				Profile:        profile,
				ConversationID: convID,
			}
			continue
		}

		// 2. Check for native `agy ... --conversation=<id>`
		if matches := agyConvRegex.FindStringSubmatch(cmdline); len(matches) >= 2 {
			convID := matches[1]
			// Only set if not already set by parent aim process
			if _, exists := result[convID]; !exists {
				result[convID] = ActiveProcessInfo{
					PID:            pid,
					Agent:          "agy",
					Profile:        "<host>",
					ConversationID: convID,
				}
			}
			continue
		}

		// 3. Check for `codex ... resume <id>`
		if matches := codexResumeRegex.FindStringSubmatch(cmdline); len(matches) >= 2 {
			convID := matches[1]
			if _, exists := result[convID]; !exists {
				result[convID] = ActiveProcessInfo{
					PID:            pid,
					Agent:          "codex",
					Profile:        "<host>",
					ConversationID: convID,
				}
			}
			continue
		}

		// 4. Check for `claude ... --resume <id>`
		if matches := claudeResumeRegex.FindStringSubmatch(cmdline); len(matches) >= 2 {
			convID := matches[1]
			if _, exists := result[convID]; !exists {
				result[convID] = ActiveProcessInfo{
					PID:            pid,
					Agent:          "claude",
					Profile:        "<host>",
					ConversationID: convID,
				}
			}
			continue
		}
	}
	if err := scanner.Err(); err != nil {
		logger.Debug("[session/process] scanner error parsing process output: %v", err)
	}

	return result
}

// GracefulTerminate terminates a process gracefully using SIGTERM, falling back
// to SIGKILL on its process group and process if it fails to exit within timeout.
func GracefulTerminate(pid int, timeout time.Duration) error {
	if pid <= 0 {
		return fmt.Errorf("invalid PID: %d", pid)
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	// Check if already dead
	if err := syscall.Kill(pid, 0); err != nil && errors.Is(err, syscall.ESRCH) {
		return nil
	}

	// Try graceful SIGTERM first
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		if errors.Is(err, os.ErrProcessDone) || errors.Is(err, syscall.ESRCH) {
			return nil
		}
		// If sending SIGTERM fails, attempt SIGKILL fallback
		_ = syscall.Kill(-pid, syscall.SIGKILL)
		if killErr := proc.Kill(); killErr != nil && !errors.Is(killErr, os.ErrProcessDone) && !errors.Is(killErr, syscall.ESRCH) {
			return killErr
		}
		return nil
	}

	done := make(chan struct{}, 1)
	stop := make(chan struct{})
	defer close(stop)

	go func() {
		_, waitErr := proc.Wait()
		if waitErr == nil || !errors.Is(waitErr, syscall.ECHILD) {
			// Child process exited (or wait succeeded)
			select {
			case done <- struct{}{}:
			default:
			}
			return
		}

		// Non-child process (proc.Wait returned syscall.ECHILD):
		// Monitor exit using syscall.Kill(pid, 0) until it returns ESRCH.
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()

		for {
			if err := syscall.Kill(pid, 0); err != nil {
				if errors.Is(err, syscall.ESRCH) {
					select {
					case done <- struct{}{}:
					default:
					}
					return
				}
			}

			select {
			case <-stop:
				return
			case <-ticker.C:
			}
		}
	}()

	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		// Force SIGKILL on process group if timeout expires
		_ = syscall.Kill(-pid, syscall.SIGKILL)
		if killErr := proc.Kill(); killErr != nil && !errors.Is(killErr, os.ErrProcessDone) && !errors.Is(killErr, syscall.ESRCH) {
			return killErr
		}
		return nil
	}
}
