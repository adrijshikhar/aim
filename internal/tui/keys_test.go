package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/session"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func TestTUI_Keymaps_InterfaceCompliance(t *testing.T) {
	// Verify that each keymap implements help.KeyMap interface
	var _ help.KeyMap = DefaultKeyMap()
	var _ help.KeyMap = DefaultSessionsDrawerKeyMap()
	var _ help.KeyMap = DefaultDoctorDrawerKeyMap()
	var _ help.KeyMap = DefaultResumeModalKeyMap()
	var _ help.KeyMap = DefaultDeleteModalKeyMap()
	var _ help.KeyMap = DefaultMoveModalKeyMap()
	var _ help.KeyMap = DefaultRenameModalKeyMap()

	km := DefaultKeyMap()
	if len(km.ShortHelp()) == 0 || len(km.FullHelp()) == 0 {
		t.Fatalf("expected non-empty ShortHelp and FullHelp in root KeyMap")
	}

	skm := DefaultSessionsDrawerKeyMap()
	if len(skm.ShortHelp()) == 0 || len(skm.FullHelp()) == 0 {
		t.Fatalf("expected non-empty ShortHelp and FullHelp in SessionsDrawerKeyMap")
	}
	if len(skm.ShortHelpFilter()) == 0 || len(skm.ShortHelpQuery()) == 0 {
		t.Fatalf("expected non-empty filter and query ShortHelp in SessionsDrawerKeyMap")
	}

	dkm := DefaultDoctorDrawerKeyMap()
	if len(dkm.ShortHelp()) == 0 || len(dkm.FullHelp()) == 0 {
		t.Fatalf("expected non-empty ShortHelp and FullHelp in DoctorDrawerKeyMap")
	}

	rkm := DefaultResumeModalKeyMap()
	if len(rkm.ShortHelp()) == 0 || len(rkm.FullHelp()) == 0 {
		t.Fatalf("expected non-empty ShortHelp and FullHelp in ResumeModalKeyMap")
	}
	if len(rkm.ShortHelpFlags()) == 0 || len(rkm.ShortHelpProfile()) == 0 {
		t.Fatalf("expected non-empty flags and profile ShortHelp in ResumeModalKeyMap")
	}

	delKm := DefaultDeleteModalKeyMap()
	if len(delKm.ShortHelp()) == 0 || len(delKm.FullHelp()) == 0 {
		t.Fatalf("expected non-empty ShortHelp and FullHelp in DeleteModalKeyMap")
	}
	if len(delKm.ShortHelpShared()) == 0 || len(delKm.ShortHelpSingle()) == 0 {
		t.Fatalf("expected non-empty shared and single ShortHelp in DeleteModalKeyMap")
	}

	mkm := DefaultMoveModalKeyMap()
	if len(mkm.ShortHelp()) == 0 || len(mkm.FullHelp()) == 0 {
		t.Fatalf("expected non-empty ShortHelp and FullHelp in MoveModalKeyMap")
	}

	renKm := DefaultRenameModalKeyMap()
	if len(renKm.ShortHelp()) == 0 || len(renKm.FullHelp()) == 0 {
		t.Fatalf("expected non-empty ShortHelp and FullHelp in RenameModalKeyMap")
	}
}

func TestTUI_Keymaps_ForFilterMode_VimKeyIsolation(t *testing.T) {
	skm := DefaultSessionsDrawerKeyMap()
	filterKm := skm.ForFilterMode()

	// In filter mode, 'k', 'j', 'f', 'c', 'b' must NOT match navigation
	kMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	jMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	fMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}}

	if key.Matches(kMsg, filterKm.Up) {
		t.Errorf("expected 'k' to NOT match filterKm.Up in filter mode")
	}
	if key.Matches(jMsg, filterKm.Down) {
		t.Errorf("expected 'j' to NOT match filterKm.Down in filter mode")
	}
	if key.Matches(fMsg, filterKm.Flags) {
		t.Errorf("expected 'f' to NOT match filterKm.Flags in filter mode")
	}
	if filterKm.Catalyst.Enabled() {
		t.Errorf("expected filterKm.Catalyst to be disabled in filter mode")
	}
	if filterKm.Fork.Enabled() {
		t.Errorf("expected filterKm.Fork to be disabled in filter mode")
	}
	if filterKm.Filter.Enabled() {
		t.Errorf("expected filterKm.Filter to be disabled in filter mode")
	}

	// But arrow keys and ctrl sequences MUST match
	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	ctrlPMsg := tea.KeyMsg{Type: tea.KeyCtrlP}
	ctrlNMsg := tea.KeyMsg{Type: tea.KeyCtrlN}
	ctrlFMsg := tea.KeyMsg{Type: tea.KeyCtrlF}
	tabMsg := tea.KeyMsg{Type: tea.KeyTab}
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}

	if !key.Matches(upMsg, filterKm.Up) {
		t.Errorf("expected Up arrow to match filterKm.Up")
	}
	if !key.Matches(ctrlPMsg, filterKm.Up) {
		t.Errorf("expected Ctrl+P to match filterKm.Up")
	}
	if !key.Matches(downMsg, filterKm.Down) {
		t.Errorf("expected Down arrow to match filterKm.Down")
	}
	if !key.Matches(ctrlNMsg, filterKm.Down) {
		t.Errorf("expected Ctrl+N to match filterKm.Down")
	}
	if !key.Matches(ctrlFMsg, filterKm.Flags) {
		t.Errorf("expected Ctrl+F to match filterKm.Flags")
	}
	if !key.Matches(tabMsg, filterKm.TabFocus) {
		t.Errorf("expected Tab to match filterKm.TabFocus")
	}
	if !key.Matches(escMsg, filterKm.TabFocus) {
		t.Errorf("expected Esc to match filterKm.TabFocus")
	}
}

func TestTUI_ThemedHelp_RenderingAndStyles(t *testing.T) {
	h := NewThemedHelp()
	if h.ShortSeparator != " • " {
		t.Errorf("expected ShortSeparator to be ' • ', got %q", h.ShortSeparator)
	}
	if h.Width != 100 {
		t.Errorf("expected clamped Width to be 100, got %d", h.Width)
	}

	// Verify rendered short help width does not exceed 100 columns for any component
	sessionsHelp := h.ShortHelpView(DefaultSessionsDrawerKeyMap().ShortHelp())
	if w := lipgloss.Width(sessionsHelp); w > 100 {
		t.Errorf("sessions drawer short help width %d exceeds 100: %q", w, sessionsHelp)
	}

	queryHelp := h.ShortHelpView(DefaultSessionsDrawerKeyMap().ShortHelpQuery())
	if w := lipgloss.Width(queryHelp); w > 100 {
		t.Errorf("sessions drawer query short help width %d exceeds 100: %q", w, queryHelp)
	}

	filterHelp := h.ShortHelpView(DefaultSessionsDrawerKeyMap().ShortHelpFilter())
	if w := lipgloss.Width(filterHelp); w > 100 {
		t.Errorf("sessions drawer filter short help width %d exceeds 100: %q", w, filterHelp)
	}

	resumeHelp := h.ShortHelpView(DefaultResumeModalKeyMap().ShortHelpProfile())
	if w := lipgloss.Width(resumeHelp); w > 100 {
		t.Errorf("resume modal profile short help width %d exceeds 100: %q", w, resumeHelp)
	}

	resumeFlagsHelp := h.ShortHelpView(DefaultResumeModalKeyMap().ShortHelpFlags())
	if w := lipgloss.Width(resumeFlagsHelp); w > 100 {
		t.Errorf("resume modal flags short help width %d exceeds 100: %q", w, resumeFlagsHelp)
	}

	doctorHelp := h.ShortHelpView(DefaultDoctorDrawerKeyMap().ShortHelp())
	if w := lipgloss.Width(doctorHelp); w > 100 {
		t.Errorf("doctor drawer short help width %d exceeds 100: %q", w, doctorHelp)
	}
	if !strings.Contains(doctorHelp, "Close Drawer") {
		t.Errorf("expected doctor drawer help to contain 'Close Drawer', got %q", doctorHelp)
	}
}

func TestTUI_ResumeModal_TwoPhaseEscDismissal(t *testing.T) {
	m := newTestModel(t, "alpha")
	sess := session.NewSession("11111111-2222-3333-4444-555566667777", "Session 1", "agy", "alpha", false, time.Now())

	// 1. Test flagsMode two-phase Esc
	mOpen, _ := m.openResumeModalWithFlags(&sess, ActionResumeExact, false, true)
	if !mOpen.IsResumeModalActive() || !mOpen.ResumeModalIsFlagsMode() {
		t.Fatalf("expected resume modal in flagsMode")
	}
	// Phase 1: Esc exits flagsMode, modal remains active
	updated, _ := mOpen.Update(tea.KeyMsg{Type: tea.KeyEsc})
	mPhase1 := updated.(Model)
	if !mPhase1.IsResumeModalActive() {
		t.Errorf("expected resume modal to remain active after first Esc from flagsMode")
	}
	if mPhase1.ResumeModalIsFlagsMode() {
		t.Errorf("expected flagsMode to be false after first Esc")
	}
	// Phase 2: Esc in profile select mode dismisses modal
	closed, _ := mPhase1.Update(tea.KeyMsg{Type: tea.KeyEsc})
	mClosed := closed.(Model)
	if mClosed.IsResumeModalActive() {
		t.Errorf("expected resume modal to be closed after second Esc")
	}

	// 2. Test customMode two-phase Esc
	mOpen2, _ := m.openResumeModal(&sess, ActionResumeExact, false)
	// Navigate to "Enter custom profile name..."
	customIdx := len(mOpen2.ResumeModalProfiles())
	mOpen2.resumeModal.cursor = customIdx
	mCustom, _ := mOpen2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mCustomModel := mCustom.(Model)
	if !mCustomModel.IsResumeModalActive() || !mCustomModel.ResumeModalIsCustomMode() {
		t.Fatalf("expected resume modal in customMode")
	}
	// Phase 1: Esc exits customMode, modal remains active
	updatedCustom, _ := mCustomModel.Update(tea.KeyMsg{Type: tea.KeyEsc})
	mCustomPhase1 := updatedCustom.(Model)
	if !mCustomPhase1.IsResumeModalActive() {
		t.Errorf("expected resume modal to remain active after first Esc from customMode")
	}
	if mCustomPhase1.ResumeModalIsCustomMode() {
		t.Errorf("expected customMode to be false after first Esc")
	}
	// Phase 2: Esc in profile select mode dismisses modal
	closedCustom, _ := mCustomPhase1.Update(tea.KeyMsg{Type: tea.KeyEsc})
	mClosedCustom := closedCustom.(Model)
	if mClosedCustom.IsResumeModalActive() {
		t.Errorf("expected resume modal to be closed after second Esc")
	}
}
