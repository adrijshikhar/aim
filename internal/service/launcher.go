package service

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

var (
	// ErrNoTerminalAvailable indicates that no supported terminal emulator was found on the system.
	ErrNoTerminalAvailable = errors.New("no supported terminal emulator found")
	// ErrEmptyCommand indicates that an empty command string was provided to LaunchTerminal.
	ErrEmptyCommand = errors.New("command string cannot be empty")
)

// FormatGhosttyArgs formats arguments for launching Ghostty via `open -a Ghostty --args -e ...`.
func FormatGhosttyArgs(cmdStr string) []string {
	return append([]string{"-a", "Ghostty", "--args", "-e"}, strings.Fields(cmdStr)...)
}

// FormatITermScript generates an AppleScript command to launch a command in iTerm2.
func FormatITermScript(appName, cmdStr string) string {
	escaped := strings.ReplaceAll(cmdStr, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	return fmt.Sprintf(`tell application "%s" to create window with default profile command "%s"`, appName, escaped)
}

// FormatTerminalScript generates an AppleScript command to launch a command in macOS Terminal.app.
func FormatTerminalScript(cmdStr string) string {
	escaped := strings.ReplaceAll(cmdStr, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	return fmt.Sprintf(`tell application "Terminal" to do script "%s"`, escaped)
}

// FormatLinuxArgs formats execution arguments for known Linux terminal emulators.
func FormatLinuxArgs(term, cmdStr string) []string {
	base := filepath.Base(term)
	switch base {
	case "gnome-terminal":
		return []string{term, "--", "sh", "-c", cmdStr}
	case "kitty":
		return []string{term, "-e", "sh", "-c", cmdStr}
	case "alacritty":
		return []string{term, "-e", "sh", "-c", cmdStr}
	case "wezterm":
		return []string{term, "start", "--", "sh", "-c", cmdStr}
	default:
		// x-terminal-emulator, $TERMINAL, konsole, xfce4-terminal, xterm
		return []string{term, "-e", "sh", "-c", cmdStr}
	}
}

// LinuxCandidateTerminals returns the default priority order of terminal emulators for Linux.
var LinuxCandidateTerminals = []string{
	"x-terminal-emulator",
	"gnome-terminal",
	"kitty",
	"alacritty",
	"wezterm",
	"konsole",
	"xfce4-terminal",
	"xterm",
}

// NewLauncherService returns the default platform-specific LauncherService implementation.
func NewLauncherService() LauncherService {
	return newPlatformLauncher()
}
