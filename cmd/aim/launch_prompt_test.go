package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/logger"
	"github.com/aim-cli/aim/internal/merge"
	"github.com/aim-cli/aim/internal/profile"
)

// No test may reach the real controlling terminal.
func init() {
	openTTY = func() (io.ReadWriteCloser, error) { return nil, errors.New("no tty in tests") }
	ttyIsForeground = func(io.ReadWriteCloser) bool { return true }
}

type fakeTTY struct {
	io.Reader
	out    bytes.Buffer
	closed bool
}

func (f *fakeTTY) Write(p []byte) (int, error) { return f.out.Write(p) }
func (f *fakeTTY) Close() error                { f.closed = true; return nil }

// promptSetup stubs the prompt inputs and returns a run func that adds a
// server "foo" during the session.
func promptSetup(t *testing.T, terminal bool, in io.Reader, tty func() (io.ReadWriteCloser, error)) (*collectionMock, merge.Store, string, string, func() int) {
	t.Helper()
	d := isolate(t)
	ad, host, prof := mockWithFiles(t, d)
	oIn, oTerm, oTTY := sessionPromptIn, stdinIsTerminal, openTTY
	sessionPromptIn, stdinIsTerminal, openTTY = in, func() bool { return terminal }, tty
	t.Cleanup(func() { sessionPromptIn, stdinIsTerminal, openTTY = oIn, oTerm, oTTY })
	run := func() int {
		_ = os.WriteFile(prof, []byte(`{"mcpServers":{"jev":{"command":"npx"},"foo":{"command":"f"}}}`), 0o600)
		return 0
	}
	return ad, merge.Store{Dir: filepath.Join(d, "state")}, host, prof, run
}

func TestWithSessionMerge_PromptOnStdinWhenTerminal(t *testing.T) {
	ad, store, host, _, run := promptSetup(t, true, strings.NewReader("p\n"), func() (io.ReadWriteCloser, error) {
		t.Fatal("stdin is a terminal: /dev/tty must not be opened")
		return nil, nil
	})
	captureOutput(t, func() { withSessionMerge(ad, store, "work", "", nil, nil, run) })
	if b, _ := os.ReadFile(host); !strings.Contains(string(b), `"foo"`) {
		t.Fatalf("promote from stdin did not reach the host: %s", b)
	}
}

func TestWithSessionMerge_PipedStdinPromptsOnTTY(t *testing.T) {
	tty := &fakeTTY{Reader: strings.NewReader("p\n")}
	ad, store, host, _, run := promptSetup(t, false, strings.NewReader("k\n"), func() (io.ReadWriteCloser, error) { return tty, nil })
	captureOutput(t, func() { withSessionMerge(ad, store, "work", "", nil, nil, run) })
	if b, _ := os.ReadFile(host); !strings.Contains(string(b), `"foo"`) {
		t.Fatalf("answer on the tty was not used: %s", b)
	}
	if !strings.Contains(tty.out.String(), "Session in work (mock) changed:") || !tty.closed {
		t.Fatalf("prompt must be written to the tty and the tty closed: %q closed=%v", tty.out.String(), tty.closed)
	}
}

// A backgrounded run must not read the terminal (SIGTTIN would stop it with
// the profile still merged): it keeps every change without prompting.
func TestWithSessionMerge_BackgroundJobKeepsWithoutPrompt(t *testing.T) {
	tty := &fakeTTY{Reader: strings.NewReader("p\n")}
	ad, store, host, prof, run := promptSetup(t, false, strings.NewReader("p\n"), func() (io.ReadWriteCloser, error) { return tty, nil })
	orig := ttyIsForeground
	ttyIsForeground = func(io.ReadWriteCloser) bool { return false }
	t.Cleanup(func() { ttyIsForeground = orig })
	_, out := captureOutput(t, func() { withSessionMerge(ad, store, "work", "", nil, nil, run) })
	if tty.out.Len() != 0 || !tty.closed {
		t.Fatalf("background: nothing may be written to the tty and it must be closed: %q closed=%v", tty.out.String(), tty.closed)
	}
	if !strings.Contains(out, "1 change(s) kept in work (mock)") {
		t.Fatalf("output = %q", out)
	}
	if b, _ := os.ReadFile(host); strings.Contains(string(b), `"foo"`) {
		t.Fatalf("nothing may reach the host: %s", b)
	}
	if b, _ := os.ReadFile(prof); !strings.Contains(string(b), `"foo"`) || strings.Contains(string(b), `"jev"`) {
		t.Fatalf("at rest = %s", b)
	}
}

// A background job whose stdin is still the terminal (`aim run … &` with job
// control on) keeps every change: reading stdin would stop it with SIGTTIN.
func TestWithSessionMerge_BackgroundJobOnTerminalStdinKeeps(t *testing.T) {
	ad, store, host, prof, run := promptSetup(t, true, strings.NewReader("p\n"), func() (io.ReadWriteCloser, error) {
		return nil, errors.New("ENXIO")
	})
	orig := ttyIsForeground
	ttyIsForeground = func(io.ReadWriteCloser) bool { return false }
	t.Cleanup(func() { ttyIsForeground = orig })
	_, out := captureOutput(t, func() { withSessionMerge(ad, store, "work", "", nil, nil, run) })
	if strings.Contains(out, "[p] promote") || !strings.Contains(out, "1 change(s) kept in work (mock)") {
		t.Fatalf("output = %q", out)
	}
	if b, _ := os.ReadFile(host); strings.Contains(string(b), `"foo"`) {
		t.Fatalf("nothing may reach the host: %s", b)
	}
	if b, _ := os.ReadFile(prof); !strings.Contains(string(b), `"foo"`) || strings.Contains(string(b), `"jev"`) {
		t.Fatalf("at rest = %s", b)
	}
}

func TestWithSessionMerge_NoTerminalKeepsWithoutPrompt(t *testing.T) {
	ad, store, host, prof, run := promptSetup(t, false, strings.NewReader("p\n"), func() (io.ReadWriteCloser, error) {
		return nil, errors.New("ENXIO")
	})
	_, out := captureOutput(t, func() { withSessionMerge(ad, store, "work", "", nil, nil, run) })
	if strings.Contains(out, "[p] promote") || !strings.Contains(out, "1 change(s) kept in work (mock)") {
		t.Fatalf("output = %q", out)
	}
	if b, _ := os.ReadFile(host); strings.Contains(string(b), `"foo"`) {
		t.Fatalf("nothing may reach the host: %s", b)
	}
	if b, _ := os.ReadFile(prof); !strings.Contains(string(b), `"foo"`) {
		t.Fatalf("change not kept: %s", b)
	}
}

// blockingReader signals aim at the first read and never answers, like a
// user pressing Ctrl+C (or closing the terminal) at the prompt.
type blockingReader struct {
	done chan struct{}
	sig  syscall.Signal
}

func (b blockingReader) Read([]byte) (int, error) {
	_ = syscall.Kill(os.Getpid(), b.sig)
	<-b.done
	return 0, io.EOF
}

// A SIGINT while the agent runs must not kill aim before Finish, and one at
// the exit prompt keeps every change.
func TestWithSessionMerge_SIGINTDuringRunAndPromptKeeps(t *testing.T) {
	testSignalDuringRunAndPromptKeeps(t, syscall.SIGINT)
}

// Closing the terminal sends SIGHUP; at the prompt it must keep and strip too.
func TestWithSessionMerge_SIGHUPDuringRunAndPromptKeeps(t *testing.T) {
	testSignalDuringRunAndPromptKeeps(t, syscall.SIGHUP)
}

func testSignalDuringRunAndPromptKeeps(t *testing.T, sig syscall.Signal) {
	br := blockingReader{done: make(chan struct{}), sig: sig}
	t.Cleanup(func() { close(br.done) })
	ad, store, host, prof, run := promptSetup(t, true, br, nil)
	interrupted := func() int {
		_ = syscall.Kill(os.Getpid(), sig)
		time.Sleep(50 * time.Millisecond)
		return run()
	}
	_, out := captureOutput(t, func() { withSessionMerge(ad, store, "work", "", nil, nil, interrupted) })
	if !strings.Contains(out, "1 change(s) kept in work (mock)") {
		t.Fatalf("output = %q", out)
	}
	if b, _ := os.ReadFile(host); strings.Contains(string(b), `"foo"`) {
		t.Fatalf("interrupt must not promote: %s", b)
	}
	if b, _ := os.ReadFile(prof); !strings.Contains(string(b), `"foo"`) || strings.Contains(string(b), `"jev"`) {
		t.Fatalf("at rest = %s", b)
	}
}

// A SIGINT while Start merges (here: waiting on the merge lock) must not
// kill aim between a profile write and the state that records it.
func TestWithSessionMerge_SIGINTDuringStartIsHeld(t *testing.T) {
	ad, store, _, prof, run := promptSetup(t, false, strings.NewReader(""), func() (io.ReadWriteCloser, error) {
		return nil, errors.New("ENXIO")
	})
	if err := os.MkdirAll(store.Dir, 0o700); err != nil {
		t.Fatal(err)
	}
	ml, err := merge.LockExclusive(store.MergeLockPath("work"), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	// Under load the SIGINT below can land before withSessionMerge installs
	// its handler; catch it here so it cannot kill the test binary.
	guard := make(chan os.Signal, 1)
	signal.Notify(guard, syscall.SIGINT)
	defer signal.Stop(guard)
	go func() {
		time.Sleep(50 * time.Millisecond)
		_ = syscall.Kill(os.Getpid(), syscall.SIGINT)
		time.Sleep(50 * time.Millisecond)
		_ = ml.Unlock()
	}()
	_, out := captureOutput(t, func() { withSessionMerge(ad, store, "work", "", nil, nil, run) })
	if !strings.Contains(out, "1 change(s) kept in work (mock)") {
		t.Fatalf("output = %q", out)
	}
	if b, _ := os.ReadFile(prof); !strings.Contains(string(b), `"foo"`) || strings.Contains(string(b), `"jev"`) {
		t.Fatalf("at rest = %s", b)
	}
}

// A session launched from the TUI runs after the TUI has closed, so its
// warnings (logger.Warn) must reach the console.
func TestRunTUI_LaunchedSessionShowsWarnings(t *testing.T) {
	var buf bytes.Buffer
	logger.SetWarnOutput(&buf)
	t.Cleanup(func() { logger.SetWarnOutput(nil) })
	orig := tuiRunner
	tuiRunner = func(*agents.Registry, *profile.ProfileManager) int {
		logger.Warn("seeding refused")
		return 0
	}
	t.Cleanup(func() { tuiRunner = orig })
	runTUI(nil, nil)
	if !strings.Contains(buf.String(), "aim: seeding refused") {
		t.Fatalf("warning from the launched session was hidden: %q", buf.String())
	}
}
