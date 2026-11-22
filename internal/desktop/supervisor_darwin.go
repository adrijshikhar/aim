//go:build darwin

package desktop

import (
	"context"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

type darwinScanner struct{}

func (d *darwinScanner) listProcesses(ctx context.Context) ([]rawProcess, error) {
	cmd := exec.CommandContext(ctx, "ps", "-eo", "pid,command")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var results []rawProcess
	lines := strings.Split(string(out), "\n")
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}
		pid, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		cmdline := parts[1]
		cmdParts := strings.Fields(cmdline)
		if len(cmdParts) == 0 {
			continue
		}
		results = append(results, rawProcess{
			pid:     pid,
			binPath: cmdParts[0],
			args:    cmdParts[1:],
		})
	}
	return results, nil
}

func (d *darwinScanner) isAlive(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}

func (d *darwinScanner) sendSignal(pid int, sig syscall.Signal) error {
	return syscall.Kill(pid, sig)
}

func (d *darwinScanner) activateApp(ctx context.Context, appName string) error {
	cmd := exec.CommandContext(ctx, "osascript", "-e", "on run argv", "-e", "tell application (item 1 of argv) to activate", "-e", "end run", appName)
	return cmd.Run()
}

func (d *darwinScanner) quitApp(ctx context.Context, appName string) error {
	cmd := exec.CommandContext(ctx, "osascript", "-e", "on run argv", "-e", "tell application (item 1 of argv) to quit", "-e", "end run", appName)
	return cmd.Run()
}

func newPlatformSupervisor() Supervisor {
	return newSupervisorWithScanner(&darwinScanner{})
}

func NewSupervisor() Supervisor {
	return newPlatformSupervisor()
}
