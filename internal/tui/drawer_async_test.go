package tui

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/session"
	tea "github.com/charmbracelet/bubbletea"
)

func TestSessionsDrawer_AsyncFetchAndMessageHandling(t *testing.T) {
	m := newTestModel(t, "alpha", "beta")

	// Open sessions drawer via key 's'
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = updated.(Model)

	if !m.IsSessionsDrawerActive() {
		t.Fatal("expected sessions drawer to be active")
	}
	if !m.IsSessionsLoading() {
		t.Fatal("expected IsSessionsLoading() to be true upon opening drawer")
	}
	if cmd == nil {
		t.Fatal("expected non-nil tea.Cmd for async session fetching")
	}

	// Execute cmd
	msg := cmd()
	_, ok := msg.(sessionsLoadedMsg)
	if !ok {
		t.Fatalf("expected sessionsLoadedMsg, got %T", msg)
	}

	// Deliver custom sessions via sessionsLoadedMsg
	s1 := session.NewSession("11111111-2222-3333-4444-555566667777", "Async Session 1", "agy", "alpha", false, time.Now())
	s2 := session.NewSession("88888888-9999-aaaa-bbbb-ccccddddeeee", "Async Session 2", "agy", "beta", false, time.Now())

	updated, _ = m.Update(sessionsLoadedMsg{
		sessions: []session.Session{s1, s2},
		err:      nil,
	})
	m = updated.(Model)

	if m.IsSessionsLoading() {
		t.Fatal("expected loading to be false after sessionsLoadedMsg")
	}
	if len(m.SessionsDrawerList()) != 2 {
		t.Fatalf("expected 2 sessions loaded, got %d", len(m.SessionsDrawerList()))
	}
	if m.SessionsDrawerList()[0].Title != "Async Session 1" {
		t.Errorf("expected first session title 'Async Session 1', got %q", m.SessionsDrawerList()[0].Title)
	}

	// Error handling: error should clear loading and preserve sessions
	updated, _ = m.Update(sessionsLoadedMsg{
		sessions: nil,
		err:      errors.New("disk read timeout"),
	})
	m = updated.(Model)

	if m.IsSessionsLoading() {
		t.Fatal("expected loading to be false even on error")
	}
	if len(m.SessionsDrawerList()) != 2 {
		t.Fatalf("expected sessions to be preserved on error, got %d", len(m.SessionsDrawerList()))
	}
}

func TestDoctorDrawer_AsyncFetchAndMessageHandling(t *testing.T) {
	base := t.TempDir()
	host := t.TempDir()
	t.Setenv("AIM_HOME", base)
	t.Setenv("AIM_REAL_HOME", host)
	if err := os.WriteFile(filepath.Join(host, ".gitconfig"), []byte("[user]"), 0644); err != nil {
		t.Fatal(err)
	}
	pm := profile.NewProfileManager(base)
	if _, err := pm.EnsureProfile("prof1"); err != nil {
		t.Fatal(err)
	}
	if _, err := pm.EnsureProfile("prof2"); err != nil {
		t.Fatal(err)
	}
	cfg := config.NewDefaultConfig()
	cfg.Profiles = map[string]config.ProfileConfig{
		"prof1": {Agents: []string{"agy"}},
		"prof2": {Agents: []string{"agy"}},
	}
	reg := agents.NewRegistry()
	reg.Register(&doctorRecorder{})

	m := NewModel(reg, pm, cfg)

	// Open doctor drawer via 'd'
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = updated.(Model)

	if !m.IsDoctorDrawerActive() {
		t.Fatal("expected doctor drawer to be active")
	}
	if !m.IsDoctorLoading() {
		t.Fatal("expected IsDoctorLoading() to be true upon opening drawer")
	}
	if cmd == nil {
		t.Fatal("expected non-nil tea.Cmd for async doctor diagnostics")
	}

	// Execute cmd
	msg := cmd()
	docMsg, ok := msg.(doctorDiagnosticsLoadedMsg)
	if !ok {
		t.Fatalf("expected doctorDiagnosticsLoadedMsg, got %T", msg)
	}
	if docMsg.targetAgent != "agy" || docMsg.targetProfile != "prof1" {
		t.Fatalf("expected agy / prof1, got %s / %s", docMsg.targetAgent, docMsg.targetProfile)
	}

	// Update with new diagnostics
	newResults := []agents.DiagnosticResult{
		{Category: "Custom", Status: "OK", Message: "Async check passed"},
	}
	updated, _ = m.Update(doctorDiagnosticsLoadedMsg{
		targetAgent:   "agy",
		targetProfile: "prof1",
		results:       newResults,
	})
	m = updated.(Model)

	if m.IsDoctorLoading() {
		t.Fatal("expected loading to be false after doctorDiagnosticsLoadedMsg")
	}
	results := m.DoctorDrawerResults()
	if len(results) != 1 || results[0].Message != "Async check passed" {
		t.Fatalf("expected async results applied, got %+v", results)
	}

	// Mismatched agent/profile message should be ignored
	m.doctorDrawer.loading = true
	updated, _ = m.Update(doctorDiagnosticsLoadedMsg{
		targetAgent:   "codex", // mismatched agent
		targetProfile: "prof1",
		results: []agents.DiagnosticResult{
			{Category: "Stale", Status: "FAIL", Message: "Should not apply"},
		},
	})
	m = updated.(Model)
	if m.DoctorDrawerResults()[0].Message != "Async check passed" {
		t.Fatal("mismatched agent results should not overwrite current drawer results")
	}

	// Down navigation in drawer triggers async fetch
	updated, downCmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)
	if !m.IsDoctorLoading() {
		t.Fatal("expected IsDoctorLoading() true after navigating down")
	}
	if downCmd == nil {
		t.Fatal("expected non-nil cmd after navigating down")
	}
}
