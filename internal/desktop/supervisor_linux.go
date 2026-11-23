//go:build linux

package desktop

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

type linuxScanner struct{}

func (l *linuxScanner) listProcesses(ctx context.Context) ([]rawProcess, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}

	var results []rawProcess
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		cmdlinePath := filepath.Join("/proc", entry.Name(), "cmdline")
		data, err := os.ReadFile(cmdlinePath)
		if err != nil || len(data) == 0 {
			continue
		}

		parts := bytes.Split(data, []byte{0})
		var strParts []string
		for _, p := range parts {
			if len(p) > 0 {
				strParts = append(strParts, string(p))
			}
		}
		if len(strParts) == 0 {
			continue
		}

		results = append(results, rawProcess{
			pid:     pid,
			binPath: strParts[0],
			args:    strParts[1:],
		})
	}
	return results, nil
}

func (l *linuxScanner) isAlive(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}

func (l *linuxScanner) sendSignal(pid int, sig syscall.Signal) error {
	return syscall.Kill(pid, sig)
}

func (l *linuxScanner) activateApp(ctx context.Context, appName string) error {
	if _, err := exec.LookPath("wmctrl"); err == nil {
		return exec.CommandContext(ctx, "wmctrl", "-x", "-a", appName).Run()
	}
	return nil
}

func (l *linuxScanner) quitApp(ctx context.Context, appName string) error {
	return nil
}

func newPlatformSupervisor() Supervisor {
	return newSupervisorWithScanner(&linuxScanner{})
}

func NewSupervisor() Supervisor {
	return newPlatformSupervisor()
}
