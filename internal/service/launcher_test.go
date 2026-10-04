package service_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/aim-cli/aim/internal/service"
)

func TestFormatGhosttyArgs(t *testing.T) {
	cmdStr := "aim resume claude work 12345678"
	args := service.FormatGhosttyArgs(cmdStr)
	expected := []string{"-a", "Ghostty", "--args", "-e", "aim", "resume", "claude", "work", "12345678"}
	if !reflect.DeepEqual(args, expected) {
		t.Errorf("FormatGhosttyArgs:\nexpected: %v\ngot:      %v", expected, args)
	}
}

func TestFormatITermScript(t *testing.T) {
	cmdStr := `aim resume claude "my session" --flag="value"`
	script := service.FormatITermScript("iTerm2", cmdStr)

	if !strings.HasPrefix(script, `tell application "iTerm2" to create window with default profile command `) {
		t.Errorf("unexpected iTerm script prefix: %s", script)
	}
	if !strings.Contains(script, `\"my session\"`) {
		t.Errorf("expected escaped quotes in script: %s", script)
	}
}

func TestFormatTerminalScript(t *testing.T) {
	cmdStr := `aim resume agy work 87654321`
	script := service.FormatTerminalScript(cmdStr)

	expected := `tell application "Terminal" to do script "aim resume agy work 87654321"`
	if script != expected {
		t.Errorf("FormatTerminalScript:\nexpected: %s\ngot:      %s", expected, script)
	}

	// Test backslash and quote escaping
	complexCmd := `echo "hello \world"`
	escapedScript := service.FormatTerminalScript(complexCmd)
	if !strings.Contains(escapedScript, `\\world`) || !strings.Contains(escapedScript, `\"hello`) {
		t.Errorf("expected properly escaped backslashes and quotes: %s", escapedScript)
	}
}

func TestFormatLinuxArgs(t *testing.T) {
	tests := []struct {
		term     string
		cmdStr   string
		expected []string
	}{
		{
			term:     "gnome-terminal",
			cmdStr:   "aim resume claude 1234",
			expected: []string{"gnome-terminal", "--", "sh", "-c", "aim resume claude 1234"},
		},
		{
			term:     "/usr/bin/gnome-terminal",
			cmdStr:   "aim resume claude 1234",
			expected: []string{"/usr/bin/gnome-terminal", "--", "sh", "-c", "aim resume claude 1234"},
		},
		{
			term:     "kitty",
			cmdStr:   "aim resume agy 5678",
			expected: []string{"kitty", "-e", "sh", "-c", "aim resume agy 5678"},
		},
		{
			term:     "alacritty",
			cmdStr:   "aim resume codex 9999",
			expected: []string{"alacritty", "-e", "sh", "-c", "aim resume codex 9999"},
		},
		{
			term:     "wezterm",
			cmdStr:   "aim resume claude 4321",
			expected: []string{"wezterm", "start", "--", "sh", "-c", "aim resume claude 4321"},
		},
		{
			term:     "x-terminal-emulator",
			cmdStr:   "aim resume claude 1111",
			expected: []string{"x-terminal-emulator", "-e", "sh", "-c", "aim resume claude 1111"},
		},
	}

	for _, tc := range tests {
		actual := service.FormatLinuxArgs(tc.term, tc.cmdStr)
		if !reflect.DeepEqual(actual, tc.expected) {
			t.Errorf("FormatLinuxArgs(%s):\nexpected: %v\ngot:      %v", tc.term, tc.expected, actual)
		}
	}
}
