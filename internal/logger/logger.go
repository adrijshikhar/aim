package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	mu             sync.RWMutex
	debugExplicit  *bool
	consoleOutput  = true
	logFile        *os.File
	logFilePath    string
	debugTagStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#c678dd")).Faint(true)
	debugTimeStyle = lipgloss.NewStyle().Faint(true)
)

// Init initializes the logger with the base directory where aim-debug.log should be stored.
func Init(baseDir string) {
	mu.Lock()
	defer mu.Unlock()

	if baseDir == "" {
		return
	}
	targetPath := filepath.Join(baseDir, "aim-debug.log")
	if logFilePath == targetPath && logFile != nil {
		return
	}
	if logFile != nil {
		_ = logFile.Close()
		logFile = nil
	}
	logFilePath = targetPath
	if !isDebugLocked() {
		return
	}
	_ = os.MkdirAll(baseDir, 0755)
	f, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err == nil {
		logFile = f
	}
}

// SetDebug explicitly enables or disables debug logging.
func SetDebug(enabled bool) {
	mu.Lock()
	defer mu.Unlock()
	debugExplicit = &enabled
}

// Reset resets explicit debug settings and file state (primarily for tests).
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	debugExplicit = nil
	consoleOutput = true
	if logFile != nil {
		_ = logFile.Close()
		logFile = nil
	}
	logFilePath = ""
}

// SetConsoleOutput controls whether debug messages are printed to os.Stderr.
// Set to false during interactive TUI sessions to prevent screen corruption.
func SetConsoleOutput(enabled bool) {
	mu.Lock()
	defer mu.Unlock()
	consoleOutput = enabled
}

func isDebugLocked() bool {
	if debugExplicit != nil {
		return *debugExplicit
	}
	env := strings.TrimSpace(strings.ToLower(os.Getenv("AIM_DEBUG")))
	switch env {
	case "1", "true", "yes", "on", "debug":
		return true
	case "0", "false", "no", "off", "":
		return false
	default:
		return len(env) > 0
	}
}

// IsDebug returns whether debug logging is currently enabled.
// Priority:
// 1. Explicitly set via SetDebug(true/false) (e.g. CLI --debug flag or config)
// 2. Environment variable AIM_DEBUG (e.g. "1", "true", "yes", "on", "debug")
func IsDebug() bool {
	mu.RLock()
	defer mu.RUnlock()
	return isDebugLocked()
}

// Debug logs a debug-level message if debug mode is active.
func Debug(format string, args ...any) {
	if !IsDebug() {
		return
	}
	msg := fmt.Sprintf(format, args...)
	now := time.Now()

	mu.Lock()
	defer mu.Unlock()

	// Console output to stderr
	if consoleOutput {
		timeStr := now.Format("15:04:05.000")
		fmt.Fprintf(os.Stderr, "%s %s %s\n", debugTagStyle.Render("[AIM DEBUG]"), debugTimeStyle.Render(timeStr), msg)
	}

	// Persistent file output
	if logFile == nil && logFilePath != "" {
		_ = os.MkdirAll(filepath.Dir(logFilePath), 0755)
		f, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err == nil {
			logFile = f
		}
	}
	if logFile != nil {
		timeStr := now.Format("2006-01-02 15:04:05.000")
		fmt.Fprintf(logFile, "%s [DEBUG] %s\n", timeStr, msg)
	}
}

// Debugf is an alias for Debug.
func Debugf(format string, args ...any) {
	Debug(format, args...)
}

// Close flushes and closes the underlying log file if open.
func Close() {
	mu.Lock()
	defer mu.Unlock()
	if logFile != nil {
		_ = logFile.Close()
		logFile = nil
	}
}
