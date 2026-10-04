//go:build darwin

package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type darwinLauncher struct {
	lookPath  func(file string) (string, error)
	appExists func(appName string) bool
	runCmd    func(ctx context.Context, name string, args ...string) error
}

func newPlatformLauncher() LauncherService {
	return NewDarwinLauncher(nil, nil, nil)
}

// NewDarwinLauncher creates a LauncherService configured for macOS.
func NewDarwinLauncher(
	lookPath func(file string) (string, error),
	appExists func(appName string) bool,
	runCmd func(ctx context.Context, name string, args ...string) error,
) LauncherService {
	if lookPath == nil {
		lookPath = exec.LookPath
	}
	if appExists == nil {
		appExists = defaultDarwinAppExists
	}
	if runCmd == nil {
		runCmd = func(ctx context.Context, name string, args ...string) error {
			cmd := exec.CommandContext(ctx, name, args...)
			return cmd.Run()
		}
	}
	return &darwinLauncher{
		lookPath:  lookPath,
		appExists: appExists,
		runCmd:    runCmd,
	}
}

func defaultDarwinAppExists(appName string) bool {
	// 1. Check standard application directories
	dirs := []string{
		"/Applications",
		"/System/Applications/Utilities",
		"/System/Applications",
	}
	if home := os.Getenv("HOME"); home != "" {
		dirs = append(dirs, filepath.Join(home, "Applications"))
	}
	for _, dir := range dirs {
		appBundle := filepath.Join(dir, appName+".app")
		if info, err := os.Stat(appBundle); err == nil && info.IsDir() {
			return true
		}
	}
	// 2. Fall back to open -Ra query
	if err := exec.Command("open", "-Ra", appName).Run(); err == nil {
		return true
	}
	return false
}

// AvailableTerminals probes and returns installed macOS terminal emulators in order of preference.
func (d *darwinLauncher) AvailableTerminals() []string {
	var available []string
	if d.appExists("Ghostty") || d.hasBinary("ghostty") {
		available = append(available, "Ghostty")
	}
	if d.appExists("iTerm2") || d.appExists("iTerm") {
		available = append(available, "iTerm2")
	}
	if d.appExists("Terminal") {
		available = append(available, "Terminal")
	}
	return available
}

func (d *darwinLauncher) hasBinary(bin string) bool {
	if d.lookPath == nil {
		return false
	}
	_, err := d.lookPath(bin)
	return err == nil
}

// LaunchTerminal launches the interactive command in the highest-priority available terminal.
func (d *darwinLauncher) LaunchTerminal(ctx context.Context, cmdStr string) error {
	cmdStr = strings.TrimSpace(cmdStr)
	if cmdStr == "" {
		return ErrEmptyCommand
	}

	terms := d.AvailableTerminals()
	if len(terms) == 0 {
		return ErrNoTerminalAvailable
	}

	var lastErr error
	for _, term := range terms {
		switch term {
		case "Ghostty":
			args := FormatGhosttyArgs(cmdStr)
			if err := d.runCmd(ctx, "open", args...); err == nil {
				return nil
			} else {
				lastErr = fmt.Errorf("ghostty launch failed: %w", err)
			}
		case "iTerm2", "iTerm":
			appName := "iTerm2"
			if !d.appExists("iTerm2") && d.appExists("iTerm") {
				appName = "iTerm"
			}
			script := FormatITermScript(appName, cmdStr)
			if err := d.runCmd(ctx, "osascript", "-e", script); err == nil {
				return nil
			} else {
				lastErr = fmt.Errorf("iterm launch failed: %w", err)
			}
		case "Terminal":
			script := FormatTerminalScript(cmdStr)
			if err := d.runCmd(ctx, "osascript", "-e", "tell application \"Terminal\" to activate", "-e", script); err == nil {
				return nil
			} else {
				lastErr = fmt.Errorf("terminal.app launch failed: %w", err)
			}
		}
	}

	if lastErr != nil {
		return lastErr
	}
	return ErrNoTerminalAvailable
}
