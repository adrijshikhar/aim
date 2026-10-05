//go:build linux

package service

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

type linuxLauncher struct {
	lookPath func(file string) (string, error)
	getenv   func(key string) string
	startCmd func(ctx context.Context, name string, args ...string) error
}

func newPlatformLauncher() LauncherService {
	return NewLinuxLauncher(nil, nil, nil)
}

// NewLinuxLauncher creates a LauncherService configured for Linux.
func NewLinuxLauncher(
	lookPath func(file string) (string, error),
	getenv func(key string) string,
	startCmd func(ctx context.Context, name string, args ...string) error,
) LauncherService {
	if lookPath == nil {
		lookPath = exec.LookPath
	}
	if getenv == nil {
		getenv = os.Getenv
	}
	if startCmd == nil {
		startCmd = func(ctx context.Context, name string, args ...string) error {
			cmd := exec.Command(name, args...)
			cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			return cmd.Start()
		}
	}
	return &linuxLauncher{
		lookPath: lookPath,
		getenv:   getenv,
		startCmd: startCmd,
	}
}

// AvailableTerminals probes and returns installed Linux terminal emulators in order of preference.
func (l *linuxLauncher) AvailableTerminals() []string {
	seen := make(map[string]bool)
	var available []string

	// 1. Check $TERMINAL
	if termEnv := strings.TrimSpace(l.getenv("TERMINAL")); termEnv != "" {
		if _, err := l.lookPath(termEnv); err == nil {
			available = append(available, termEnv)
			seen[termEnv] = true
		}
	}

	// 2. Check known candidates
	for _, term := range LinuxCandidateTerminals {
		if seen[term] {
			continue
		}
		if _, err := l.lookPath(term); err == nil {
			available = append(available, term)
			seen[term] = true
		}
	}

	return available
}

// LaunchTerminal launches the interactive command in the highest-priority available terminal.
func (l *linuxLauncher) LaunchTerminal(ctx context.Context, cmdStr string) error {
	cmdStr = strings.TrimSpace(cmdStr)
	if cmdStr == "" {
		return ErrEmptyCommand
	}

	terms := l.AvailableTerminals()
	if len(terms) == 0 {
		return ErrNoTerminalAvailable
	}

	var lastErr error
	for _, term := range terms {
		args := FormatLinuxArgs(term, cmdStr)
		if err := l.startCmd(ctx, args[0], args[1:]...); err == nil {
			return nil
		} else {
			lastErr = err
		}
	}

	if lastErr != nil {
		return lastErr
	}
	return ErrNoTerminalAvailable
}
