package tui

import (
	"testing"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
	tea "github.com/charmbracelet/bubbletea"
)

func stepModel(m Model, msg tea.Msg) (Model, tea.Cmd) {
	updated, cmd := m.Update(msg)
	return updated.(Model), cmd
}

func TestFeedbackModal_OpenAndCancel(t *testing.T) {
	tempBase := t.TempDir()
	pm := profile.NewProfileManager(tempBase)
	_, _ = pm.EnsureProfile("work")
	reg := agents.NewRegistry()
	cfg := config.NewDefaultConfig()

	m := NewModel(reg, pm, cfg)
	if m.IsFeedbackModalActive() {
		t.Fatal("expected feedback modal to be initially inactive")
	}

	// Press 'f' to open modal
	var cmd tea.Cmd
	m, cmd = stepModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	if !m.IsFeedbackModalActive() {
		t.Fatal("expected feedback modal to be active after pressing 'f'")
	}
	_ = cmd

	// Press 'esc' to cancel modal
	m, _ = stepModel(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.IsFeedbackModalActive() {
		t.Fatal("expected feedback modal to be inactive after pressing 'esc'")
	}
}

func TestFeedbackModal_CycleCategoryAndToggleDoctor(t *testing.T) {
	tempBase := t.TempDir()
	pm := profile.NewProfileManager(tempBase)
	reg := agents.NewRegistry()
	cfg := config.NewDefaultConfig()

	m := NewModel(reg, pm, cfg)
	m, _ = stepModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	if !m.IsFeedbackModalActive() {
		t.Fatal("expected feedback modal active")
	}

	initCat := m.feedbackModal.categoryIndex

	// Press Tab to cycle category
	m, _ = stepModel(m, tea.KeyMsg{Type: tea.KeyTab})
	if m.feedbackModal.categoryIndex == initCat {
		t.Errorf("expected categoryIndex to change after Tab, got %d", m.feedbackModal.categoryIndex)
	}

	// Press Ctrl+D to toggle diagnostic report inclusion
	initDoctor := m.feedbackModal.includeDoctor
	m, _ = stepModel(m, tea.KeyMsg{Type: tea.KeyCtrlD})
	if m.feedbackModal.includeDoctor == initDoctor {
		t.Errorf("expected includeDoctor to toggle after Ctrl+D")
	}
}

func TestFeedbackModal_EmptyMessageShowsError(t *testing.T) {
	tempBase := t.TempDir()
	pm := profile.NewProfileManager(tempBase)
	reg := agents.NewRegistry()
	cfg := config.NewDefaultConfig()

	m := NewModel(reg, pm, cfg)
	m, _ = stepModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})

	// Press Enter without message
	m, _ = stepModel(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.feedbackModal.err == "" {
		t.Fatal("expected error on empty feedback message submission")
	}
	if !m.IsFeedbackModalActive() {
		t.Fatal("expected modal to stay active on error")
	}
}
