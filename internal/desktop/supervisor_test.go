package desktop

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"syscall"
	"testing"
)

type mockScanner struct {
	processes []rawProcess
	alivePIDs map[int]bool
	killed    []int
}

func (m *mockScanner) listProcesses(ctx context.Context) ([]rawProcess, error) {
	return m.processes, nil
}

func (m *mockScanner) isAlive(pid int) bool {
	return m.alivePIDs[pid]
}

func (m *mockScanner) sendSignal(pid int, sig syscall.Signal) error {
	m.killed = append(m.killed, pid)
	if sig == syscall.SIGTERM {
		delete(m.alivePIDs, pid)
	}
	return nil
}

func (m *mockScanner) activateApp(ctx context.Context, appName string) error {
	return nil
}

func (m *mockScanner) quitApp(ctx context.Context, appName string) error {
	for pid := range m.alivePIDs {
		delete(m.alivePIDs, pid)
	}
	return nil
}

func TestSupervisor_FindInstance(t *testing.T) {
	scanner := &mockScanner{
		processes: []rawProcess{
			{
				pid:     1234,
				binPath: "/Applications/Claude.app/Contents/MacOS/Claude",
				args:    []string{"--user-data-dir=/Users/test/.aim/profiles/work/.claude-desktop/user-data"},
			},
		},
		alivePIDs: map[int]bool{1234: true},
	}

	sup := newSupervisorWithScanner(scanner)
	ctx := context.Background()

	proc, err := sup.FindInstance(ctx, "Claude", []string{"claude", "Claude"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if proc == nil {
		t.Fatal("expected process to be found")
	}
	if proc.PID != 1234 {
		t.Errorf("expected PID 1234, got %d", proc.PID)
	}
	if proc.ProfileName != "work" {
		t.Errorf("expected profile 'work', got '%s'", proc.ProfileName)
	}
}

func TestSupervisor_FocusInstance(t *testing.T) {
	scanner := &mockScanner{
		alivePIDs: map[int]bool{1234: true},
	}
	sup := newSupervisorWithScanner(scanner)
	ctx := context.Background()

	if err := sup.FocusInstance(ctx, nil); !errors.Is(err, ErrProcessNotFound) {
		t.Fatalf("expected ErrProcessNotFound for nil proc, got %v", err)
	}

	proc := &ProcessInfo{PID: 1234, BinaryPath: "/Applications/Claude.app/Contents/MacOS/Claude"}
	if err := sup.FocusInstance(ctx, proc); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSupervisor_GracefulQuit_Success(t *testing.T) {
	scanner := &mockScanner{
		alivePIDs: map[int]bool{1234: true},
	}
	sup := newSupervisorWithScanner(scanner)
	ctx := context.Background()

	proc := &ProcessInfo{PID: 1234, BinaryPath: "/Applications/Claude.app/Contents/MacOS/Claude"}
	err := sup.GracefulQuit(ctx, proc)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if scanner.alivePIDs[1234] {
		t.Errorf("expected PID 1234 to be removed from alivePIDs")
	}
}

func TestSupervisor_GracefulQuit_ContextCanceled(t *testing.T) {
	// scanner that does not remove PID on quit
	scanner := &stubHangingScanner{alive: true}
	sup := newSupervisorWithScanner(scanner)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	proc := &ProcessInfo{PID: 9999, BinaryPath: "test"}
	err := sup.GracefulQuit(ctx, proc)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
}

type stubHangingScanner struct {
	alive bool
}

func (s *stubHangingScanner) listProcesses(ctx context.Context) ([]rawProcess, error) {
	return nil, nil
}
func (s *stubHangingScanner) isAlive(pid int) bool                                  { return s.alive }
func (s *stubHangingScanner) sendSignal(pid int, sig syscall.Signal) error          { return nil }
func (s *stubHangingScanner) activateApp(ctx context.Context, appName string) error { return nil }
func (s *stubHangingScanner) quitApp(ctx context.Context, appName string) error     { return nil }

func TestSupervisor_EnsureSingleInstance_SameProfile(t *testing.T) {
	scanner := &mockScanner{
		processes: []rawProcess{
			{
				pid:     1234,
				binPath: "/Applications/Claude.app/Contents/MacOS/Claude",
				args:    []string{"--user-data-dir=/Users/test/.aim/profiles/work/.claude-desktop/user-data"},
			},
		},
		alivePIDs: map[int]bool{1234: true},
	}

	sup := newSupervisorWithScanner(scanner)
	ctx := context.Background()

	var in bytes.Buffer
	var out bytes.Buffer

	err := sup.EnsureSingleInstance(ctx, &in, &out, "Claude", "work", "/Users/test/.aim/profiles/work/.claude-desktop/user-data", []string{"claude"}, false)
	if err != nil {
		t.Fatalf("expected nil error on same-profile activation, got: %v", err)
	}
	if !strings.Contains(out.String(), "already running") {
		t.Errorf("expected output to mention already running, got: %s", out.String())
	}
}

func TestSupervisor_EnsureSingleInstance_DifferentProfile_Refusal(t *testing.T) {
	scanner := &mockScanner{
		processes: []rawProcess{
			{
				pid:     1234,
				binPath: "/Applications/Claude.app/Contents/MacOS/Claude",
				args:    []string{"--user-data-dir=/Users/test/.aim/profiles/personal/.claude-desktop/user-data"},
			},
		},
		alivePIDs: map[int]bool{1234: true},
	}

	sup := newSupervisorWithScanner(scanner)
	ctx := context.Background()

	in := bytes.NewBufferString("n\n")
	var out bytes.Buffer

	err := sup.EnsureSingleInstance(ctx, in, &out, "Claude", "work", "/Users/test/.aim/profiles/work/.claude-desktop/user-data", []string{"claude"}, false)
	if !errors.Is(err, ErrInstanceRunning) {
		t.Fatalf("expected ErrInstanceRunning on user refusal, got: %v", err)
	}
}

func TestSupervisor_EnsureSingleInstance_DifferentProfile_Force(t *testing.T) {
	scanner := &mockScanner{
		processes: []rawProcess{
			{
				pid:     1234,
				binPath: "/Applications/Claude.app/Contents/MacOS/Claude",
				args:    []string{"--user-data-dir=/Users/test/.aim/profiles/personal/.claude-desktop/user-data"},
			},
		},
		alivePIDs: map[int]bool{1234: true},
	}

	sup := newSupervisorWithScanner(scanner)
	ctx := context.Background()

	var in bytes.Buffer
	var out bytes.Buffer

	err := sup.EnsureSingleInstance(ctx, &in, &out, "Claude", "work", "/Users/test/.aim/profiles/work/.claude-desktop/user-data", []string{"claude"}, true)
	if err != nil {
		t.Fatalf("expected successful switch with force=true, got: %v", err)
	}
	if scanner.alivePIDs[1234] {
		t.Errorf("expected PID 1234 to be terminated")
	}
}
