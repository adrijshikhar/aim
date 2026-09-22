package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/agents/agy"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/session"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

func sampleTestSessions() []session.Session {
	return []session.Session{
		{
			ID:           "session-1-uuid-1111",
			ShortID:      "11111111",
			Title:        "First Session Design",
			Agent:        "agy",
			Profile:      "office",
			IsHost:       false,
			LastActiveAt: time.Now().Add(-10 * time.Minute),
			Status:       session.StatusIdle,
		},
		{
			ID:           "session-2-uuid-2222",
			ShortID:      "22222222",
			Title:        "Second Session Bugfix",
			Agent:        "agy",
			Profile:      "work",
			IsHost:       false,
			LastActiveAt: time.Now().Add(-2 * time.Hour),
			Status:       session.StatusIdle,
		},
		{
			ID:           "session-3-uuid-3333",
			ShortID:      "33333333",
			Title:        "Third Session Feature",
			Agent:        "agy",
			Profile:      "",
			IsHost:       true,
			LastActiveAt: time.Now().Add(-48 * time.Hour),
			Status:       session.StatusIdle,
		},
	}
}

func newTestCmdWithIO(input string) (*cobra.Command, *bytes.Buffer) {
	cmd := &cobra.Command{}
	var out bytes.Buffer
	cmd.SetIn(strings.NewReader(input))
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	return cmd, &out
}

func TestResumePicker_EmptySessions(t *testing.T) {
	cmd, _ := newTestCmdWithIO("\n")
	selected, err := promptSelectSession(cmd, []session.Session{})
	if err != nil {
		t.Fatalf("unexpected error on empty sessions: %v", err)
	}
	if selected != nil {
		t.Errorf("expected nil selected session on empty list, got: %+v", selected)
	}
}

func TestResumePicker_SelectFirstSession(t *testing.T) {
	cmd, _ := newTestCmdWithIO("\n")
	sessions := sampleTestSessions()

	selected, err := promptSelectSession(cmd, sessions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if selected == nil {
		t.Fatalf("expected a selected session, got nil")
	}
	if selected.ID != "session-1-uuid-1111" {
		t.Errorf("expected session-1-uuid-1111, got: %s", selected.ID)
	}
}

func TestResumePicker_NavigateAndSelect(t *testing.T) {
	// Navigate down with 'j' and press Enter
	cmd, _ := newTestCmdWithIO("j\n")
	sessions := sampleTestSessions()

	selected, err := promptSelectSession(cmd, sessions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if selected == nil {
		t.Fatalf("expected a selected session, got nil")
	}
	if selected.ID != "session-2-uuid-2222" {
		t.Errorf("expected session-2-uuid-2222, got: %s", selected.ID)
	}
}

func TestResumePicker_NavigateArrowsAndSelect(t *testing.T) {
	// Down arrow is \x1b[B in ANSI, then Enter
	cmd, _ := newTestCmdWithIO("\x1b[B\n")
	sessions := sampleTestSessions()

	selected, err := promptSelectSession(cmd, sessions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if selected == nil {
		t.Fatalf("expected a selected session, got nil")
	}
	if selected.ID != "session-2-uuid-2222" {
		t.Errorf("expected session-2-uuid-2222, got: %s", selected.ID)
	}
}

func TestResumePicker_FilterAndSelect(t *testing.T) {
	// Filter with '/', type 'Bugfix', press Enter
	cmd, _ := newTestCmdWithIO("/Bugfix\n")
	sessions := sampleTestSessions()

	selected, err := promptSelectSession(cmd, sessions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if selected == nil {
		t.Fatalf("expected a selected session, got nil")
	}
	if selected.ID != "session-2-uuid-2222" {
		t.Errorf("expected session-2-uuid-2222, got: %s", selected.ID)
	}
}

func TestResumePicker_CancelEsc(t *testing.T) {
	// Press Esc
	cmd, _ := newTestCmdWithIO("\x1b")
	sessions := sampleTestSessions()

	selected, err := promptSelectSession(cmd, sessions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if selected != nil {
		t.Errorf("expected nil selected session on Esc cancel, got: %+v", selected)
	}
}

func TestResumePicker_CancelCtrlC(t *testing.T) {
	// Press Ctrl+C (\x03)
	cmd, _ := newTestCmdWithIO("\x03")
	sessions := sampleTestSessions()

	selected, err := promptSelectSession(cmd, sessions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if selected != nil {
		t.Errorf("expected nil selected session on Ctrl+C cancel, got: %+v", selected)
	}
}

func TestResumePicker_CancelQ(t *testing.T) {
	// Press 'q'
	cmd, _ := newTestCmdWithIO("q")
	sessions := sampleTestSessions()

	selected, err := promptSelectSession(cmd, sessions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if selected != nil {
		t.Errorf("expected nil selected session on 'q' cancel, got: %+v", selected)
	}
}

func TestResumePicker_FilterByID(t *testing.T) {
	// Filter with '/', type '3333', press Enter
	cmd, _ := newTestCmdWithIO("/3333\n")
	sessions := sampleTestSessions()

	selected, err := promptSelectSession(cmd, sessions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if selected == nil {
		t.Fatalf("expected a selected session, got nil")
	}
	if selected.ID != "session-3-uuid-3333" {
		t.Errorf("expected session-3-uuid-3333, got: %s", selected.ID)
	}
}

func TestResumePicker_ViewFormatting(t *testing.T) {
	sessions := sampleTestSessions()
	m := newSessionPickerModel(sessions)
	view := m.View()

	// Verify ShortID
	if !strings.Contains(view, "11111111") {
		t.Errorf("expected view to contain ShortID '11111111', got:\n%s", view)
	}
	// Verify Title
	if !strings.Contains(view, "First Session Design") {
		t.Errorf("expected view to contain title 'First Session Design', got:\n%s", view)
	}
	// Verify Profile Badges
	if !strings.Contains(view, "[office]") {
		t.Errorf("expected view to contain profile badge '[office]', got:\n%s", view)
	}
	if !strings.Contains(view, "[work]") {
		t.Errorf("expected view to contain profile badge '[work]', got:\n%s", view)
	}
	if !strings.Contains(view, "[host]") {
		t.Errorf("expected view to contain profile badge '[host]', got:\n%s", view)
	}
	// Verify relative time
	if !strings.Contains(view, "10m ago") && !strings.Contains(view, "ago") {
		t.Errorf("expected view to contain relative time, got:\n%s", view)
	}
}

func TestResumePicker_NavigateBounds(t *testing.T) {
	// Move up from top ('k'), should stay at top (0), then press Enter
	cmd, _ := newTestCmdWithIO("k\n")
	sessions := sampleTestSessions()

	selected, err := promptSelectSession(cmd, sessions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if selected == nil || selected.ID != "session-1-uuid-1111" {
		t.Errorf("expected session-1-uuid-1111, got: %+v", selected)
	}

	// Move down multiple times past bottom, should clamp to last item (session-3), then press Enter
	cmd2, _ := newTestCmdWithIO("jjjjjj\n")
	selected2, err := promptSelectSession(cmd2, sessions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if selected2 == nil || selected2.ID != "session-3-uuid-3333" {
		t.Errorf("expected session-3-uuid-3333, got: %+v", selected2)
	}
}

func TestResumeCmd_InteractivePicker_Select(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	fakeBinDir := filepath.Join(tempDir, "bin")
	_ = os.MkdirAll(fakeBinDir, 0755)
	fakeAgy := filepath.Join(fakeBinDir, "agy")
	_ = os.WriteFile(fakeAgy, []byte("#!/bin/sh\nexit 0\n"), 0755)
	t.Setenv("PATH", fakeBinDir+":"+os.Getenv("PATH"))

	reg := agents.NewRegistry()
	reg.Register(agy.NewAdapter())
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("work")

	oldMgr := defaultSessionManager
	defer func() { defaultSessionManager = oldMgr }()
	mockMgr := session.NewManager()
	mockMgr.RegisterProvider(&testSessionProvider{
		agent:    "agy",
		sessions: sampleTestSessions(),
	})
	defaultSessionManager = func() *session.Manager { return mockMgr }

	oldInteractive := isInteractiveFunc
	defer func() { isInteractiveFunc = oldInteractive }()
	isInteractiveFunc = func(r io.Reader) bool { return true }

	var buf bytes.Buffer
	cmd := newRootCmd(reg, pm)
	cmd.SetIn(strings.NewReader("\n"))
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"resume", "agy", "work"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error running interactive resume: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Resuming agy session 11111111") {
		t.Errorf("expected output to contain resuming message for session 11111111, got:\n%s", out)
	}
}

func TestResumeCmd_InteractivePicker_Abort(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := agents.NewRegistry()
	reg.Register(agy.NewAdapter())
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("work")

	oldMgr := defaultSessionManager
	defer func() { defaultSessionManager = oldMgr }()
	mockMgr := session.NewManager()
	mockMgr.RegisterProvider(&testSessionProvider{
		agent:    "agy",
		sessions: sampleTestSessions(),
	})
	defaultSessionManager = func() *session.Manager { return mockMgr }

	oldInteractive := isInteractiveFunc
	defer func() { isInteractiveFunc = oldInteractive }()
	isInteractiveFunc = func(r io.Reader) bool { return true }

	var buf bytes.Buffer
	cmd := newRootCmd(reg, pm)
	cmd.SetIn(strings.NewReader("q"))
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"resume", "agy", "work"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error running interactive abort: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Resume aborted.") {
		t.Errorf("expected 'Resume aborted.' in output, got:\n%s", out)
	}
}

func TestResumeCmd_InteractivePicker_NoSessions(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := agents.NewRegistry()
	reg.Register(agy.NewAdapter())
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("work")

	oldMgr := defaultSessionManager
	defer func() { defaultSessionManager = oldMgr }()
	mockMgr := session.NewManager()
	mockMgr.RegisterProvider(&testSessionProvider{
		agent:    "agy",
		sessions: nil,
	})
	defaultSessionManager = func() *session.Manager { return mockMgr }

	oldInteractive := isInteractiveFunc
	defer func() { isInteractiveFunc = oldInteractive }()
	isInteractiveFunc = func(r io.Reader) bool { return true }

	var buf bytes.Buffer
	cmd := newRootCmd(reg, pm)
	cmd.SetIn(strings.NewReader("\n"))
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"resume", "agy", "work"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error on empty sessions: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "No sessions found for agy") {
		t.Errorf("expected 'No sessions found for agy' in output, got:\n%s", out)
	}
}

func TestResumeCmd_NonInteractive_MissingSessionIDError(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := agents.NewRegistry()
	reg.Register(agy.NewAdapter())
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("work")

	oldInteractive := isInteractiveFunc
	defer func() { isInteractiveFunc = oldInteractive }()
	isInteractiveFunc = func(r io.Reader) bool { return false }

	var buf bytes.Buffer
	cmd := newRootCmd(reg, pm)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"resume", "agy", "work"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error for non-interactive resume without session ID, got nil")
	}
	expected := "missing session ID to resume (use 'aim sessions agy' to browse or run interactively)"
	if !strings.Contains(err.Error(), expected) {
		t.Errorf("expected error %q, got: %q", expected, err.Error())
	}
}

func TestResumePicker_ReverseScrollingOffset(t *testing.T) {
	var sessions []session.Session
	for i := 0; i < 25; i++ {
		sessions = append(sessions, session.Session{
			ID:           fmt.Sprintf("session-%02d-uuid", i),
			ShortID:      fmt.Sprintf("%08d", i),
			Title:        fmt.Sprintf("Session Number %02d", i),
			Agent:        "agy",
			Profile:      "work",
			LastActiveAt: time.Now().Add(-time.Duration(i) * time.Minute),
		})
	}

	m := newSessionPickerModel(sessions)
	if m.offset != 0 || m.cursor != 0 {
		t.Fatalf("expected initial offset=0, cursor=0; got offset=%d, cursor=%d", m.offset, m.cursor)
	}

	// 1. Move down past maxVisible (10)
	// Pressing 'j' 12 times brings cursor to 12. offset should become 12 - 10 + 1 = 3.
	for i := 0; i < 12; i++ {
		newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		m = newM.(sessionPickerModel)
	}
	if m.cursor != 12 {
		t.Errorf("expected cursor 12, got %d", m.cursor)
	}
	if m.offset != 3 {
		t.Errorf("expected offset 3 after scrolling down to 12, got %d", m.offset)
	}

	viewDown := m.View()
	if !strings.Contains(viewDown, "(showing 4-13 of 25 sessions)") {
		t.Errorf("expected view to show 'showing 4-13 of 25 sessions', got:\n%s", viewDown)
	}
	if strings.Contains(viewDown, "Session Number 00") {
		t.Errorf("expected Session 00 to be scrolled out of view, got:\n%s", viewDown)
	}
	if !strings.Contains(viewDown, "Session Number 12") {
		t.Errorf("expected Session 12 to be visible, got:\n%s", viewDown)
	}

	// 2. Reverse scroll up within the visible window: offset should remain unchanged
	for i := 0; i < 5; i++ {
		newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		m = newM.(sessionPickerModel)
	}
	if m.cursor != 7 {
		t.Errorf("expected cursor 7 after moving up 5 times, got %d", m.cursor)
	}
	if m.offset != 3 {
		t.Errorf("expected offset to stay 3 while cursor (7) >= offset (3), got %d", m.offset)
	}

	// 3. Reverse scroll up past the top of the visible window: offset should follow cursor
	// Moving up 6 more times brings cursor to 1, offset should become 1.
	for i := 0; i < 6; i++ {
		newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		m = newM.(sessionPickerModel)
	}
	if m.cursor != 1 {
		t.Errorf("expected cursor 1, got %d", m.cursor)
	}
	if m.offset != 1 {
		t.Errorf("expected offset 1 when cursor < previous offset, got %d", m.offset)
	}

	// Move up once more to reach index 0
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = newM.(sessionPickerModel)
	if m.cursor != 0 || m.offset != 0 {
		t.Errorf("expected cursor=0, offset=0, got cursor=%d, offset=%d", m.cursor, m.offset)
	}

	viewTop := m.View()
	if !strings.Contains(viewTop, "(showing 1-10 of 25 sessions)") {
		t.Errorf("expected view to show 'showing 1-10 of 25 sessions', got:\n%s", viewTop)
	}
	if !strings.Contains(viewTop, "Session Number 00") {
		t.Errorf("expected Session 00 to be visible at top, got:\n%s", viewTop)
	}
}
