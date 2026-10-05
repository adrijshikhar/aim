//go:build darwin

package service_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/aim-cli/aim/internal/service"
)

func TestDarwinLauncher_Unit(t *testing.T) {
	ctx := context.Background()

	// 1. Ghostty available and succeeds
	var executedCmds [][]string
	mockRun := func(ctx context.Context, name string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{name}, args...))
		return nil
	}
	appMap := map[string]bool{
		"Ghostty":  true,
		"Terminal": true,
	}
	mockAppExists := func(name string) bool {
		return appMap[name]
	}

	launcher := service.NewDarwinLauncher(nil, mockAppExists, mockRun)
	avail := launcher.AvailableTerminals()
	if !reflect.DeepEqual(avail, []string{"Ghostty", "Terminal"}) {
		t.Fatalf("expected [Ghostty, Terminal], got %v", avail)
	}

	err := launcher.LaunchTerminal(ctx, "aim resume claude 12345678")
	if err != nil {
		t.Fatalf("unexpected error launching Ghostty: %v", err)
	}
	if len(executedCmds) != 1 {
		t.Fatalf("expected 1 command executed, got %d", len(executedCmds))
	}
	if executedCmds[0][0] != "open" || executedCmds[0][2] != "Ghostty" {
		t.Errorf("unexpected command executed: %v", executedCmds[0])
	}

	// 2. Ghostty fails, falls back to Terminal.app
	executedCmds = nil
	mockRunFallback := func(ctx context.Context, name string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{name}, args...))
		if len(args) > 1 && args[1] == "Ghostty" {
			return errors.New("ghostty launch error")
		}
		return nil
	}
	launcherFallback := service.NewDarwinLauncher(nil, mockAppExists, mockRunFallback)
	err = launcherFallback.LaunchTerminal(ctx, "aim resume claude 12345678")
	if err != nil {
		t.Fatalf("unexpected error on fallback launch: %v", err)
	}
	if len(executedCmds) != 2 {
		t.Fatalf("expected 2 commands executed (Ghostty attempt then Terminal fallback), got %d", len(executedCmds))
	}
	if executedCmds[1][0] != "osascript" {
		t.Errorf("expected osascript fallback, got %v", executedCmds[1])
	}

	// 3. iTerm2 only
	executedCmds = nil
	itermOnlyMap := map[string]bool{
		"iTerm2": true,
	}
	launcherITerm := service.NewDarwinLauncher(nil, func(name string) bool { return itermOnlyMap[name] }, mockRun)
	err = launcherITerm.LaunchTerminal(ctx, "aim resume claude 12345678")
	if err != nil {
		t.Fatalf("unexpected error launching iTerm2: %v", err)
	}
	if len(executedCmds) != 1 || executedCmds[0][0] != "osascript" {
		t.Fatalf("expected osascript for iTerm2, got %v", executedCmds)
	}

	// 4. Empty command string
	if err := launcher.LaunchTerminal(ctx, ""); !errors.Is(err, service.ErrEmptyCommand) {
		t.Errorf("expected ErrEmptyCommand, got %v", err)
	}

	// 5. No terminal available
	noneLauncher := service.NewDarwinLauncher(nil, func(name string) bool { return false }, mockRun)
	if err := noneLauncher.LaunchTerminal(ctx, "echo 1"); !errors.Is(err, service.ErrNoTerminalAvailable) {
		t.Errorf("expected ErrNoTerminalAvailable, got %v", err)
	}
}
