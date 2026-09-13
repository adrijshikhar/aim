package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aim-cli/aim/internal/config"
)

func TestIsDebug_EnvVar(t *testing.T) {
	defer Reset()

	tests := []struct {
		val      string
		expected bool
	}{
		{"1", true},
		{"true", true},
		{"TRUE", true},
		{"yes", true},
		{"on", true},
		{"debug", true},
		{"0", false},
		{"false", false},
		{"no", false},
		{"off", false},
		{"", false},
	}

	for _, tt := range tests {
		Reset()
		t.Setenv("AIM_DEBUG", tt.val)
		if got := IsDebug(); got != tt.expected {
			t.Errorf("AIM_DEBUG=%q: expected %v, got %v", tt.val, tt.expected, got)
		}
	}
}

func TestIsDebug_ExplicitOverride(t *testing.T) {
	defer Reset()

	t.Setenv("AIM_DEBUG", "0")
	if IsDebug() {
		t.Fatalf("expected debug false initially")
	}

	SetDebug(true)
	if !IsDebug() {
		t.Fatalf("expected debug true after SetDebug(true)")
	}

	SetDebug(false)
	if IsDebug() {
		t.Fatalf("expected debug false after SetDebug(false)")
	}
}

func TestDebug_FileLogging(t *testing.T) {
	defer Reset()

	dir := t.TempDir()
	Init(dir)
	SetDebug(true)
	SetConsoleOutput(false)

	Debug("testing message with arg: %s", "foo")
	Debugf("another message %d", 42)

	Close()

	logPath := filepath.Join(dir, "aim-debug.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "[DEBUG] testing message with arg: foo") {
		t.Errorf("expected log file to contain test message, got: %s", content)
	}
	if !strings.Contains(content, "[DEBUG] another message 42") {
		t.Errorf("expected log file to contain second test message, got: %s", content)
	}
}

func TestDebug_DisabledDoesNotWrite(t *testing.T) {
	defer Reset()

	dir := t.TempDir()
	Init(dir)
	SetDebug(false)
	t.Setenv("AIM_DEBUG", "")

	Debug("should not be logged")
	Close()

	logPath := filepath.Join(dir, "aim-debug.log")
	if fi, err := os.Stat(logPath); err == nil && fi.Size() > 0 {
		t.Errorf("expected empty log file when debug disabled, got size %d", fi.Size())
	}
}

func TestInit_DefaultStateDir(t *testing.T) {
	defer Reset()

	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)

	Init("")
	expectedPath := filepath.Join(tempDir, "aim-debug.log")
	if got := LogFilePath(); got != expectedPath {
		t.Errorf("LogFilePath() = %q, want %q", got, expectedPath)
	}

	// Calling with config.BaseDir() should also map cleanly to StateDir
	Init(config.BaseDir())
	if got := LogFilePath(); got != expectedPath {
		t.Errorf("LogFilePath() with BaseDir = %q, want %q", got, expectedPath)
	}
}

func TestDebug_UninitializedDefaultStateDir(t *testing.T) {
	defer Reset()

	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)

	SetDebug(true)
	SetConsoleOutput(false)

	Debug("uninitialized debug message: %s", "hello")
	Close()

	expectedPath := filepath.Join(tempDir, "aim-debug.log")
	data, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("failed to read expected log file at %s: %v", expectedPath, err)
	}
	if !strings.Contains(string(data), "[DEBUG] uninitialized debug message: hello") {
		t.Errorf("log file did not contain expected message: %s", string(data))
	}
}

func TestInit_XDGStateDir(t *testing.T) {
	defer Reset()

	cleanHome := t.TempDir()
	t.Setenv("AIM_HOME", "")
	t.Setenv("AIM_REAL_HOME", cleanHome)

	stateDir := filepath.Join(cleanHome, "xdg-state")
	t.Setenv("XDG_STATE_HOME", stateDir)
	config.ReloadXDG()

	Init("")
	expectedPath := filepath.Join(stateDir, "aim", "aim-debug.log")
	if got := LogFilePath(); got != expectedPath {
		t.Errorf("LogFilePath() = %q, want %q", got, expectedPath)
	}
}
