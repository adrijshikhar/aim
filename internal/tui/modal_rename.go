package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type renameModalState struct {
	active        bool
	targetProfile string
	input         textinput.Model
	err           string
}

func (m Model) IsRenameModalActive() bool {
	return m.renameModal.active
}

func (m Model) RenameModalTarget() string {
	return m.renameModal.targetProfile
}

func (m Model) RenameModalInputValue() string {
	return m.renameModal.input.Value()
}

func (m Model) RenameModalError() string {
	return m.renameModal.err
}

func (m Model) openRenameModal() (Model, tea.Cmd) {
	if len(m.profiles) == 0 || m.cursor < 0 || m.cursor >= len(m.profiles) {
		return m, nil
	}
	target := m.profiles[m.cursor]
	ti := textinput.New()
	ti.Placeholder = target
	ti.CharLimit = 64
	ti.Width = 30
	cmd := ti.Focus()
	m.renameModal = renameModalState{
		active:        true,
		targetProfile: target,
		input:         ti,
	}
	return m, cmd
}

func (m Model) updateRenameModal(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.cancelStream()
			return m, tea.Quit
		case "esc":
			m.renameModal = renameModalState{}
			return m, nil
		case "enter":
			return m.executeRename()
		default:
			var cmd tea.Cmd
			m.renameModal.input, cmd = m.renameModal.input.Update(msg)
			return m, cmd
		}
	default:
		var cmd tea.Cmd
		m.renameModal.input, cmd = m.renameModal.input.Update(msg)
		return m, cmd
	}
}

func (m Model) executeRename() (Model, tea.Cmd) {
	newName := strings.TrimSpace(m.renameModal.input.Value())
	if newName == "" {
		m.renameModal.err = "Profile name cannot be empty"
		return m, nil
	}
	if newName == m.renameModal.targetProfile {
		m.renameModal.err = "New profile name must be different from current name"
		return m, nil
	}
	if strings.ContainsAny(newName, "/\\") || strings.Contains(newName, "..") {
		m.renameModal.err = "Profile name cannot contain slashes or '..'"
		return m, nil
	}
	if m.pm != nil {
		if err := m.pm.RenameProfile(m.renameModal.targetProfile, newName, m.cfg); err != nil {
			m.renameModal.err = err.Error()
			return m, nil
		}
	}
	if m.cache != nil {
		m.cache.Rename(m.renameModal.targetProfile, newName)
	}
	targetProfile := m.renameModal.targetProfile
	if m.reports != nil {
		for k, rep := range m.reports {
			parts := strings.SplitN(k, ":", 2)
			if len(parts) == 2 && parts[1] == targetProfile {
				delete(m.reports, k)
				rep.Profile = newName
				m.reports[fmt.Sprintf("%s:%s", parts[0], newName)] = rep
			} else if k == targetProfile {
				delete(m.reports, k)
				rep.Profile = newName
				m.reports[newName] = rep
			}
		}
	}
	m.renameModal = renameModalState{}
	m = m.refreshProfiles()
	for idx, p := range m.profiles {
		if p == newName {
			m.cursor = idx
			break
		}
	}
	return m, nil
}

func (m Model) renderRenameModal() string {
	var b strings.Builder
	pName := m.renameModal.targetProfile

	title := RenameModalTitleStyle.Render("✎ Rename Profile: " + pName)
	b.WriteString(title + "\n\n")

	b.WriteString(lipgloss.NewStyle().Foreground(TextSecondary).Render(
		"Enter new name for profile:",
	) + "\n\n")

	b.WriteString("  " + m.renameModal.input.View() + "\n\n")

	if m.renameModal.err != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(StatusRed).Bold(true).Render(
			"  ✕ "+m.renameModal.err,
		) + "\n\n")
	}

	b.WriteString(lipgloss.NewStyle().Foreground(TextMuted).Render(
		"  [Enter] Confirm  •  [Esc] Cancel",
	))

	box := RenameModalBoxStyle.Render(b.String())
	return "\n" + box + "\n"
}
