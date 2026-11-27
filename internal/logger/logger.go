package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/aim-cli/aim/internal/config"
	"github.com/charmbracelet/lipgloss"
)

const maxLogFileBytes int64 = 5 * 1024 * 1024

var (
	bearerRedactPattern    = regexp.MustCompile(`(?i)\b(Bearer|Basic)\s+[A-Za-z0-9._~+/=-]{8,}`)
	jsonSecretFieldPattern = regexp.MustCompile(`(?i)("(?:access_token|refresh_token|id_token|client_secret|api_key|token|secret)"\s*:\s*")([^"]+)(")`)
	urlParamSecretPattern  = regexp.MustCompile(`(?i)((?:access_token|refresh_token|code|token|api_key|secret)=)([^&\s]+)`)
)

var (
	mu             sync.RWMutex
	debugExplicit  *bool
	consoleOutput  = true
	logFile        *os.File
	logFilePath    string
	warnOut        io.Writer = os.Stderr
	debugTagStyle            = lipgloss.NewStyle().Foreground(lipgloss.Color("#c678dd")).Faint(true)
	debugTimeStyle           = lipgloss.NewStyle().Faint(true)
)

// LogFilePath returns the current or default path for aim-debug.log.
func LogFilePath() string {
	mu.RLock()
	defer mu.RUnlock()
	if logFilePath != "" {
		return logFilePath
	}
	return filepath.Join(config.StateDir(), "aim-debug.log")
}

// Init initializes the logger with the base directory where aim-debug.log should be stored.
// If dir is empty or matches config.BaseDir(), it defaults to config.StateDir() so logs adhere to the XDG state standard.
// If an explicit custom dir is provided, it is respected.
func Init(dir string) {
	mu.Lock()
	defer mu.Unlock()

	if dir == "" || dir == config.BaseDir() {
		dir = config.StateDir()
	}
	targetPath := filepath.Join(dir, "aim-debug.log")
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
	openLogFileLocked()
}

func openLogFileLocked() {
	if logFilePath == "" {
		logFilePath = filepath.Join(config.StateDir(), "aim-debug.log")
	}
	dir := filepath.Dir(logFilePath)
	base := filepath.Base(logFilePath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return
	}
	defer func() { _ = root.Close() }()

	if fi, statErr := root.Stat(base); statErr == nil && fi.Size() >= maxLogFileBytes {
		_ = root.Remove(base + ".1")
		_ = os.Rename(filepath.Join(dir, base), filepath.Join(dir, base+".1"))
	}
	if f, openErr := root.OpenFile(base, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600); openErr == nil {
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
	warnOut = os.Stderr
}

// SetWarnOutput redirects Warn's console output (tests); nil means os.Stderr.
func SetWarnOutput(w io.Writer) {
	mu.Lock()
	defer mu.Unlock()
	if w == nil {
		w = os.Stderr
	}
	warnOut = w
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

func sanitizeLogMessage(msg string) string {
	s := strings.NewReplacer("\r", `\r`, "\n", `\n`).Replace(msg)
	s = bearerRedactPattern.ReplaceAllString(s, "$1 [REDACTED]")
	s = jsonSecretFieldPattern.ReplaceAllString(s, `$1[REDACTED]$3`)
	s = urlParamSecretPattern.ReplaceAllString(s, `$1[REDACTED]`)
	return s
}

// Debug logs a debug-level message if debug mode is active.
func Debug(format string, args ...any) {
	mu.Lock()
	defer mu.Unlock()
	if !isDebugLocked() {
		return
	}
	msg := sanitizeLogMessage(fmt.Sprintf(format, args...))
	now := time.Now()

	// Console output to stderr
	if consoleOutput {
		timeStr := now.Format("15:04:05.000")
		fmt.Fprintf(os.Stderr, "%s %s %s\n", debugTagStyle.Render("[AIM DEBUG]"), debugTimeStyle.Render(timeStr), msg)
	}

	writeLogLocked(now, "DEBUG", msg)
}

// Warn prints a message the user needs to see whether or not debug is on, as
// "aim: <msg>" on stderr (unless console output is off, as in the TUI), and
// records it in the debug log when debug is on.
func Warn(format string, args ...any) {
	msg := sanitizeLogMessage(fmt.Sprintf(format, args...))
	mu.Lock()
	defer mu.Unlock()
	if consoleOutput {
		fmt.Fprintf(warnOut, "aim: %s\n", msg)
	}
	if isDebugLocked() {
		writeLogLocked(time.Now(), "WARN", msg)
	}
}

// Error records an error-level message unconditionally in the debug log file.
func Error(format string, args ...any) {
	msg := sanitizeLogMessage(fmt.Sprintf(format, args...))
	mu.Lock()
	defer mu.Unlock()
	writeLogLocked(time.Now(), "ERROR", msg)
}

// writeLogLocked appends one line to aim-debug.log; mu must be held.
func writeLogLocked(now time.Time, level, msg string) {
	if logFile != nil {
		if fi, err := logFile.Stat(); err == nil && fi.Size() >= maxLogFileBytes {
			_ = logFile.Close()
			logFile = nil
		}
	}
	if logFile == nil {
		openLogFileLocked()
	}
	if logFile != nil {
		timeStr := now.Format("2006-01-02 15:04:05.000")
		fmt.Fprintf(logFile, "%s [%s] %s\n", timeStr, level, msg)
	}
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
