package service

import (
	"context"
	"io"
	"testing"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/agents/claudedesktop"
	"github.com/aim-cli/aim/internal/desktop"
)

type mockSupervisor struct {
	findFunc  func(ctx context.Context, appName string, binaries []string) (*desktop.ProcessInfo, error)
	quitFunc  func(ctx context.Context, proc *desktop.ProcessInfo) error
	focusFunc func(ctx context.Context, proc *desktop.ProcessInfo) error
}

func (m *mockSupervisor) FindInstance(ctx context.Context, appName string, binaries []string) (*desktop.ProcessInfo, error) {
	if m.findFunc != nil {
		return m.findFunc(ctx, appName, binaries)
	}
	return nil, nil
}

func (m *mockSupervisor) FocusInstance(ctx context.Context, proc *desktop.ProcessInfo) error {
	if m.focusFunc != nil {
		return m.focusFunc(ctx, proc)
	}
	return nil
}

func (m *mockSupervisor) GracefulQuit(ctx context.Context, proc *desktop.ProcessInfo) error {
	if m.quitFunc != nil {
		return m.quitFunc(ctx, proc)
	}
	return nil
}

func (m *mockSupervisor) EnsureSingleInstance(
	ctx context.Context,
	in io.Reader,
	out io.Writer,
	appName string,
	targetProfile, targetUserDataDir string,
	binaries []string,
	force bool,
) (bool, error) {
	return false, nil
}

func TestDesktopService_ListStatus_RunningAndStopped(t *testing.T) {
	reg := agents.NewRegistry()
	reg.Register(claudedesktop.NewAdapter())

	// Test case 1: Running
	mockRunning := &mockSupervisor{
		findFunc: func(ctx context.Context, appName string, binaries []string) (*desktop.ProcessInfo, error) {
			return &desktop.ProcessInfo{
				PID:         5432,
				ProfileName: "work",
			}, nil
		},
	}
	svcRunning := NewDesktopService(reg, mockRunning, nil)
	statuses, err := svcRunning.ListStatus(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(statuses) != 1 {
		t.Fatalf("expected 1 status, got %d", len(statuses))
	}
	if !statuses[0].Running || statuses[0].PID != 5432 || statuses[0].Profile != "work" {
		t.Errorf("unexpected running status: %+v", statuses[0])
	}

	// Test case 2: Stopped
	mockStopped := &mockSupervisor{
		findFunc: func(ctx context.Context, appName string, binaries []string) (*desktop.ProcessInfo, error) {
			return nil, nil
		},
	}
	svcStopped := NewDesktopService(reg, mockStopped, nil)
	statusesStopped, err := svcStopped.ListStatus(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if statusesStopped[0].Running || statusesStopped[0].PID != 0 {
		t.Errorf("unexpected stopped status: %+v", statusesStopped[0])
	}
}

func TestDesktopService_QuitAndFocus_NotFound(t *testing.T) {
	reg := agents.NewRegistry()
	reg.Register(claudedesktop.NewAdapter())

	mock := &mockSupervisor{
		findFunc: func(ctx context.Context, appName string, binaries []string) (*desktop.ProcessInfo, error) {
			return nil, nil
		},
	}

	svc := NewDesktopService(reg, mock, nil)
	ctx := context.Background()

	if err := svc.Quit(ctx, "claude-desktop"); err == nil {
		t.Error("expected error when quitting non-running app")
	}

	if err := svc.Focus(ctx, "claude-desktop"); err == nil {
		t.Error("expected error when focusing non-running app")
	}
}

func TestDesktopService_QuitAndFocus_Success(t *testing.T) {
	reg := agents.NewRegistry()
	reg.Register(claudedesktop.NewAdapter())

	quitCalled := false
	focusCalled := false
	mock := &mockSupervisor{
		findFunc: func(ctx context.Context, appName string, binaries []string) (*desktop.ProcessInfo, error) {
			return &desktop.ProcessInfo{PID: 1234, ProfileName: "work"}, nil
		},
		quitFunc: func(ctx context.Context, proc *desktop.ProcessInfo) error {
			quitCalled = true
			return nil
		},
		focusFunc: func(ctx context.Context, proc *desktop.ProcessInfo) error {
			focusCalled = true
			return nil
		},
	}

	svc := NewDesktopService(reg, mock, nil)
	ctx := context.Background()

	if err := svc.Quit(ctx, "claude-desktop"); err != nil {
		t.Errorf("unexpected error on quit: %v", err)
	}
	if !quitCalled {
		t.Error("expected quit to be called on supervisor")
	}

	if err := svc.Focus(ctx, "claude-desktop"); err != nil {
		t.Errorf("unexpected error on focus: %v", err)
	}
	if !focusCalled {
		t.Error("expected focus to be called on supervisor")
	}
}
