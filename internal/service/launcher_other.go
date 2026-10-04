//go:build !darwin && !linux

package service

import (
	"context"
	"strings"
)

type fallbackLauncher struct{}

func newPlatformLauncher() LauncherService {
	return &fallbackLauncher{}
}

func (f *fallbackLauncher) LaunchTerminal(ctx context.Context, cmdStr string) error {
	if strings.TrimSpace(cmdStr) == "" {
		return ErrEmptyCommand
	}
	return ErrNoTerminalAvailable
}

func (f *fallbackLauncher) AvailableTerminals() []string {
	return nil
}
