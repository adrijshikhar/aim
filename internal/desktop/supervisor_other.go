//go:build !darwin && !linux

package desktop

import (
	"context"
	"io"
)

type unsupportedSupervisor struct{}

func (u *unsupportedSupervisor) FindInstance(ctx context.Context, appName string, binaryNames []string) (*ProcessInfo, error) {
	return nil, ErrDesktopUnsupported
}

func (u *unsupportedSupervisor) FocusInstance(ctx context.Context, proc *ProcessInfo) error {
	return ErrDesktopUnsupported
}

func (u *unsupportedSupervisor) GracefulQuit(ctx context.Context, proc *ProcessInfo) error {
	return ErrDesktopUnsupported
}

func (u *unsupportedSupervisor) EnsureSingleInstance(ctx context.Context, in io.Reader, out io.Writer, appName string, targetProfile, targetUserDataDir string, binaries []string, force bool) error {
	return ErrDesktopUnsupported
}

func newPlatformSupervisor() Supervisor {
	return &unsupportedSupervisor{}
}

func NewSupervisor() Supervisor {
	return newPlatformSupervisor()
}
