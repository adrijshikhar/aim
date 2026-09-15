package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
	tea "github.com/charmbracelet/bubbletea"
)

func TestTUI_MoveModal_TriggerAndCancel(t *testing.T) {
	tempDir := t.TempDir()
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("alpha")
	_, _ = pm.EnsureProfile("beta")

	cfg := config.NewDefaultConfig()
	cfg.AddProfileAgent("alpha", "codex")
	cfg.AddProfileAgent("beta", "codex")

	reg := agents.NewRegistry()
	m := NewModel(reg, pm, cfg)
	m.agent = "codex"
	m = m.refreshProfiles()

	// Press 'v' to trigger move modal
	mOpen, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	m2 := mOpen.(Model)
	if !m2.IsMoveModalActive() {
		t.Fatalf("expected move modal to be active after pressing 'v'")
	}
	if m2.MoveModalSource() != "alpha" {
		t.Errorf("expected source 'alpha', got %q", m2.MoveModalSource())
	}
	if m2.MoveModalAgent() != "codex" {
		t.Errorf("expected agent 'codex', got %q", m2.MoveModalAgent())
	}

	// Press 'esc' to cancel
	mCancel, _ := m2.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m3 := mCancel.(Model)
	if m3.IsMoveModalActive() {
		t.Fatalf("expected move modal to be closed after Esc")
	}
}

func TestTUI_MoveModal_ValidationErrors(t *testing.T) {
	tempDir := t.TempDir()
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("alpha")

	cfg := config.NewDefaultConfig()
	cfg.AddProfileAgent("alpha", "codex")

	reg := agents.NewRegistry()
	m := NewModel(reg, pm, cfg)
	m.agent = "codex"
	m = m.refreshProfiles()

	mOpen, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	mActive := mOpen.(Model)

	// 1. Empty destination
	mEmptyRes, _ := mActive.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mEmpty := mEmptyRes.(Model)
	if !mEmpty.IsMoveModalActive() {
		t.Fatalf("expected modal to remain open on empty destination")
	}
	if mEmpty.MoveModalError() != "Destination profile name cannot be empty" {
		t.Errorf("expected empty destination error, got %q", mEmpty.MoveModalError())
	}

	// 2. Same source and destination
	mActive.moveModal.input.SetValue("alpha")
	mSameRes, _ := mActive.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mSame := mSameRes.(Model)
	if !mSame.IsMoveModalActive() {
		t.Fatalf("expected modal to remain open on same source")
	}
	if mSame.MoveModalError() != "Destination profile must be different from source profile" {
		t.Errorf("expected same profile error, got %q", mSame.MoveModalError())
	}

	// 3. Slashes in name
	mActive.moveModal.input.SetValue("invalid/name")
	mSlashRes, _ := mActive.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mSlash := mSlashRes.(Model)
	if !mSlash.IsMoveModalActive() {
		t.Fatalf("expected modal to remain open on invalid name")
	}
	if !strings.Contains(mSlash.MoveModalError(), "cannot contain slashes") {
		t.Errorf("expected slash error, got %q", mSlash.MoveModalError())
	}
}

func TestTUI_MoveModal_Success(t *testing.T) {
	tempDir := t.TempDir()
	pm := profile.NewProfileManager(tempDir)
	srcDir, _ := pm.EnsureProfile("rs")
	_ = os.MkdirAll(filepath.Join(srcDir, ".codex"), 0700)
	_ = os.WriteFile(filepath.Join(srcDir, ".codex", "auth.json"), []byte("tok-123"), 0600)

	cfg := config.NewDefaultConfig()
	cfg.AddProfileAgent("rs", "agy")
	cfg.AddProfileAgent("rs", "codex")

	reg := agents.NewRegistry()
	m := NewModel(reg, pm, cfg)
	m.agent = "codex"
	m = m.refreshProfiles()

	mOpen, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	mActive := mOpen.(Model)
	mActive.moveModal.input.SetValue("work")

	mDoneRes, _ := mActive.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mDone := mDoneRes.(Model)
	if mDone.IsMoveModalActive() {
		t.Fatalf("expected move modal to close on success, err: %s", mDone.MoveModalError())
	}

	// Verify rs still has agy but not codex
	if !cfg.HasAgent("rs", "agy") {
		t.Error("expected rs to keep agy")
	}
	if cfg.HasAgent("rs", "codex") {
		t.Error("expected rs to lose codex")
	}

	// Verify work has codex
	if !cfg.HasAgent("work", "codex") {
		t.Error("expected work to have codex")
	}
	destAuth, err := os.ReadFile(filepath.Join(pm.ProfileDir("work"), ".codex", "auth.json"))
	if err != nil || string(destAuth) != "tok-123" {
		t.Errorf("expected moved codex auth in work profile, got %s (err: %v)", string(destAuth), err)
	}
}

func TestTUI_MoveModal_ViewRendering(t *testing.T) {
	tempDir := t.TempDir()
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("alpha")

	cfg := config.NewDefaultConfig()
	cfg.AddProfileAgent("alpha", "codex")

	reg := agents.NewRegistry()
	m := NewModel(reg, pm, cfg)
	m.agent = "codex"
	m = m.refreshProfiles()

	mOpen, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	mActive := mOpen.(Model)
	view := mActive.View()

	if !strings.Contains(view, "Move Account: codex (alpha)") {
		t.Errorf("expected modal title in View, got:\n%s", view)
	}
	if !strings.Contains(view, "[Enter] Confirm") || !strings.Contains(view, "[Esc] Cancel") {
		t.Errorf("expected help hints in View, got:\n%s", view)
	}
}
