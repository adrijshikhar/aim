package desktop

import (
	"context"
	"errors"
	"io"
	"syscall"
)

var (
	ErrDesktopUnsupported = errors.New("desktop operations are not supported on this operating system")
	ErrNoDisplayServer    = errors.New("no graphical display server found ($DISPLAY and $WAYLAND_DISPLAY are unset)")
	ErrInstanceRunning    = errors.New("another profile instance of this desktop app is currently running")
	ErrQuitTimeout        = errors.New("timed out waiting for running instance to exit")
	ErrProcessNotFound    = errors.New("desktop application process not found")
)

type ProcessInfo struct {
	PID         int    `json:"pid"`
	BinaryPath  string `json:"binary_path"`
	UserDataDir string `json:"user_data_dir"`
	ProfileName string `json:"profile_name"`
}

type Supervisor interface {
	FindInstance(ctx context.Context, appName string, binaryNames []string) (*ProcessInfo, error)
	FocusInstance(ctx context.Context, proc *ProcessInfo) error
	GracefulQuit(ctx context.Context, proc *ProcessInfo) error
	EnsureSingleInstance(ctx context.Context, in io.Reader, out io.Writer, appName string, targetProfile, targetUserDataDir string, binaries []string, force bool) error
}

type rawProcess struct {
	pid     int
	binPath string
	args    []string
}

type processScanner interface {
	listProcesses(ctx context.Context) ([]rawProcess, error)
	isAlive(pid int) bool
	sendSignal(pid int, sig syscall.Signal) error
	activateApp(ctx context.Context, appName string) error
	quitApp(ctx context.Context, appName string) error
}
