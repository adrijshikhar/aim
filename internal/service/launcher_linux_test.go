//go:build linux

package service_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/aim-cli/aim/internal/service"
)

func TestLinuxLauncher_Unit(t *testing.T) {
	ctx := context.Background()

	// 1. Available terminals with $TERMINAL set
	mockEnv := map[string]string{
		"TERMINAL": "custom-term",
	}
	mockLookPath := func(file string) (string, error) {
		if file == "custom-term" || file == "alacritty" || file == "kitty" {
			return "/usr/bin/" + file, nil
		}
		return "", errors.New("not found")
	}

	var startedCmds [][]string
	mockStartCmd := func(ctx context.Context, name string, args ...string) error {
		startedCmds = append(startedCmds, append([]string{name}, args...))
		return nil
	}

	launcher := service.NewLinuxLauncher(
		mockLookPath,
		func(k string) string { return mockEnv[k] },
		mockStartCmd,
	)

	avail := launcher.AvailableTerminals()
	expectedAvail := []string{"custom-term", "kitty", "alacritty"}
	if !reflect.DeepEqual(avail, expectedAvail) {
		t.Fatalf("expected %v, got %v", expectedAvail, avail)
	}

	// 2. LaunchTerminal
	err := launcher.LaunchTerminal(ctx, "aim resume claude 1234")
	if err != nil {
		t.Fatalf("unexpected error launching terminal: %v", err)
	}
	if len(startedCmds) != 1 {
		t.Fatalf("expected 1 command started, got %d", len(startedCmds))
	}
	expectedArgs := []string{"custom-term", "-e", "sh", "-c", "aim resume claude 1234"}
	if !reflect.DeepEqual(startedCmds[0], expectedArgs) {
		t.Errorf("expected command %v, got %v", expectedArgs, startedCmds[0])
	}

	// 3. Fallback when first terminal start fails
	startedCmds = nil
	mockStartCmdFallback := func(ctx context.Context, name string, args ...string) error {
		startedCmds = append(startedCmds, append([]string{name}, args...))
		if name == "custom-term" {
			return errors.New("cannot start custom-term")
		}
		return nil
	}
	launcherFallback := service.NewLinuxLauncher(
		mockLookPath,
		func(k string) string { return mockEnv[k] },
		mockStartCmdFallback,
	)
	err = launcherFallback.LaunchTerminal(ctx, "aim resume claude 1234")
	if err != nil {
		t.Fatalf("unexpected error on fallback: %v", err)
	}
	if len(startedCmds) != 2 {
		t.Fatalf("expected 2 commands attempted, got %d", len(startedCmds))
	}
	if startedCmds[1][0] != "kitty" {
		t.Errorf("expected kitty fallback, got %v", startedCmds[1])
	}

	// 4. Empty command string
	if err := launcher.LaunchTerminal(ctx, ""); !errors.Is(err, service.ErrEmptyCommand) {
		t.Errorf("expected ErrEmptyCommand, got %v", err)
	}

	// 5. No terminals available
	emptyLauncher := service.NewLinuxLauncher(
		func(file string) (string, error) { return "", errors.New("not found") },
		func(k string) string { return "" },
		mockStartCmd,
	)
	if err := emptyLauncher.LaunchTerminal(ctx, "echo 1"); !errors.Is(err, service.ErrNoTerminalAvailable) {
		t.Errorf("expected ErrNoTerminalAvailable, got %v", err)
	}
}
