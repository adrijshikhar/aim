package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
