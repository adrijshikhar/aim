package tui

import (
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/session"
	tea "github.com/charmbracelet/bubbletea"
)

func TestDoctorDrawer_SubModelInterface(t *testing.T) {
	var sm SubModel = DoctorDrawer{}
	if sm.Init() != nil {
		t.Fatal("expected DoctorDrawer.Init() to return nil")
	}

	d := DoctorDrawer{
		active:        true,
		targetAgent:   "agy",
		targetProfile: "default",
	}

	// Test Diagnostics Loaded message
	results := []agents.DiagnosticResult{
		{Category: "TestCat", Status: "OK", Message: "TestMsg"},
	}
	updated, _ := d.Update(doctorDiagnosticsLoadedMsg{
		targetAgent:   "agy",
		targetProfile: "default",
		results:       results,
	})
	d = updated.(DoctorDrawer)
	if len(d.results) != 1 || d.results[0].Message != "TestMsg" {
		t.Fatalf("expected results updated, got %+v", d.results)
	}

	// Test View
	view := d.View()
	if view == "" {
		t.Fatal("expected non-empty DoctorDrawer.View()")
	}

	// Test Navigate Up / Down typed messages
	_, upCmd := d.Update(tea.KeyMsg{Type: tea.KeyUp})
	if upCmd == nil {
		t.Fatal("expected non-nil tea.Cmd for Up key")
	}
	navMsg, ok := upCmd().(DoctorNavigateProfileMsg)
	if !ok || navMsg.Delta != -1 {
		t.Fatalf("expected DoctorNavigateProfileMsg{Delta: -1}, got %+v", navMsg)
	}

	_, downCmd := d.Update(tea.KeyMsg{Type: tea.KeyDown})
	if downCmd == nil {
		t.Fatal("expected non-nil tea.Cmd for Down key")
	}
	navMsg, ok = downCmd().(DoctorNavigateProfileMsg)
	if !ok || navMsg.Delta != 1 {
		t.Fatalf("expected DoctorNavigateProfileMsg{Delta: 1}, got %+v", navMsg)
	}

	// Test Agent Selection typed messages
	_, a1Cmd := d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	if a1Cmd == nil {
		t.Fatal("expected non-nil tea.Cmd for key 1")
	}
	selMsg, ok := a1Cmd().(DoctorSelectAgentMsg)
	if !ok || selMsg.Index != 0 {
		t.Fatalf("expected DoctorSelectAgentMsg{Index: 0}, got %+v", selMsg)
	}

	// Test Agent Cycling typed messages
	_, cycCmd := d.Update(tea.KeyMsg{Type: tea.KeyTab})
	if cycCmd == nil {
		t.Fatal("expected non-nil tea.Cmd for Tab key")
	}
	cycMsg, ok := cycCmd().(DoctorCycleAgentMsg)
	if !ok || !cycMsg.Forward {
		t.Fatalf("expected DoctorCycleAgentMsg{Forward: true}, got %+v", cycMsg)
	}

	// Test Close and Quit typed messages
	_, closeCmd := d.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if closeCmd == nil {
		t.Fatal("expected non-nil tea.Cmd for Esc key")
	}
	if _, ok := closeCmd().(DoctorCloseMsg); !ok {
		t.Fatalf("expected DoctorCloseMsg, got %T", closeCmd())
	}

	_, quitCmd := d.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if quitCmd == nil {
		t.Fatal("expected non-nil tea.Cmd for Quit key (ctrl+c)")
	}
	if _, ok := quitCmd().(DoctorQuitMsg); !ok {
		t.Fatalf("expected DoctorQuitMsg, got %T", quitCmd())
	}
}

func TestSessionsDrawer_SubModelInterface(t *testing.T) {
	var sm SubModel = SessionsDrawer{}
	if sm.Init() != nil {
		t.Fatal("expected non-loading SessionsDrawer.Init() to return nil")
	}

	smLoading := SessionsDrawer{loading: true}
	if smLoading.Init() == nil {
		t.Fatal("expected loading SessionsDrawer.Init() to return non-nil fetch cmd")
	}

	s1 := session.NewSession("11111111-2222-3333-4444-555566667777", "SubModel Session 1", "agy", "default", false, time.Now())
	d := SessionsDrawer{
		active:   true,
		sessions: []session.Session{s1},
		index:    NewSessionsIndex([]session.Session{s1}),
	}

	// Test sessionsLoadedMsg
	s2 := session.NewSession("22222222-3333-4444-5555-666677778888", "SubModel Session 2", "agy", "default", false, time.Now())
	updated, _ := d.Update(sessionsLoadedMsg{
		sessions: []session.Session{s1, s2},
		err:      nil,
	})
	d = updated.(SessionsDrawer)
	if len(d.sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(d.sessions))
	}

	// Test WindowSizeMsg
	updated, _ = d.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	d = updated.(SessionsDrawer)
	if d.height != 40 {
		t.Fatalf("expected height 40, got %d", d.height)
	}

	// Test View
	view := d.View()
	if view == "" {
		t.Fatal("expected non-empty SessionsDrawer.View()")
	}

	// Test Resume action typed messages
	_, enterCmd := d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if enterCmd == nil {
		t.Fatal("expected non-nil tea.Cmd for Enter key")
	}
	resumeMsg, ok := enterCmd().(SessionsResumeMsg)
	if !ok || resumeMsg.Session == nil || resumeMsg.Session.ID != s1.ID || resumeMsg.Action != ActionResumeExact {
		t.Fatalf("expected SessionsResumeMsg with exact resume, got %+v", resumeMsg)
	}

	// Test Flags resume typed message
	_, flagsCmd := d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	if flagsCmd == nil {
		t.Fatal("expected non-nil tea.Cmd for f key")
	}
	resumeMsg, ok = flagsCmd().(SessionsResumeMsg)
	if !ok || !resumeMsg.WithFlags {
		t.Fatalf("expected SessionsResumeMsg with WithFlags=true, got %+v", resumeMsg)
	}

	// Test Catalyst resume typed message
	_, catCmd := d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	if catCmd == nil {
		t.Fatal("expected non-nil tea.Cmd for c key")
	}
	resumeMsg, ok = catCmd().(SessionsResumeMsg)
	if !ok || resumeMsg.Action != ActionResumeCatalyst {
		t.Fatalf("expected SessionsResumeMsg with ActionResumeCatalyst, got %+v", resumeMsg)
	}

	// Test Fork resume typed message
	_, forkCmd := d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	if forkCmd == nil {
		t.Fatal("expected non-nil tea.Cmd for b key")
	}
	resumeMsg, ok = forkCmd().(SessionsResumeMsg)
	if !ok || !resumeMsg.Fork {
		t.Fatalf("expected SessionsResumeMsg with Fork=true, got %+v", resumeMsg)
	}

	// Test Close and Quit typed messages
	_, closeCmd := d.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if closeCmd == nil {
		t.Fatal("expected non-nil tea.Cmd for Esc key")
	}
	if _, ok := closeCmd().(SessionsCloseMsg); !ok {
		t.Fatalf("expected SessionsCloseMsg, got %T", closeCmd())
	}

	_, quitCmd := d.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if quitCmd == nil {
		t.Fatal("expected non-nil tea.Cmd for Quit key (ctrl+c)")
	}
	if _, ok := quitCmd().(SessionsQuitMsg); !ok {
		t.Fatalf("expected SessionsQuitMsg, got %T", quitCmd())
	}
}
