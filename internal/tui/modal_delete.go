package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type deleteModalState struct {
	active        bool
	targetProfile string
	isShared      bool
	agents        []string
	focusedIndex  int
}

func (m Model) IsDeleteModalActive() bool {
	return m.deleteModal.active
}

func (m Model) DeleteModalTarget() string {
	return m.deleteModal.targetProfile
}

func (m Model) DeleteModalFocusedIndex() int {
	return m.deleteModal.focusedIndex
}

func (m Model) openDeleteModal() (Model, tea.Cmd) {
	if len(m.profiles) == 0 || m.cursor < 0 || m.cursor >= len(m.profiles) {
		return m, nil
	}
	target := m.profiles[m.cursor]
	var agentsList []string
	if m.cfg != nil {
		agentsList = m.cfg.GetProfileAgents(target)
	}
	isShared := len(agentsList) > 1
	defaultFocus := 1
	if isShared {
		defaultFocus = 2
	}
	m.deleteModal = deleteModalState{
		active:        true,
		targetProfile: target,
		isShared:      isShared,
		agents:        agentsList,
		focusedIndex:  defaultFocus,
	}
	return m, nil
}

func (m Model) updateDeleteModal(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.cancelStream()
		return m, tea.Quit
	case "esc", "q", "n":
		m.deleteModal = deleteModalState{}
		return m, nil
	case "left", "h":
		numOpts := 2
		if m.deleteModal.isShared {
			numOpts = 3
		}
		m.deleteModal.focusedIndex = (m.deleteModal.focusedIndex - 1 + numOpts) % numOpts
		return m, nil
	case "right", "l", "tab":
		numOpts := 2
		if m.deleteModal.isShared {
			numOpts = 3
		}
		m.deleteModal.focusedIndex = (m.deleteModal.focusedIndex + 1) % numOpts
		return m, nil
	case "shift+tab":
		numOpts := 2
		if m.deleteModal.isShared {
			numOpts = 3
		}
		m.deleteModal.focusedIndex = (m.deleteModal.focusedIndex - 1 + numOpts) % numOpts
		return m, nil
	case "1":
		if m.deleteModal.isShared {
			return m.executeDeleteChoice(0)
		}
	case "2":
		if m.deleteModal.isShared {
			return m.executeDeleteChoice(1)
		}
	case "y":
		if !m.deleteModal.isShared {
			return m.executeDeleteChoice(0)
		}
	case "enter":
		return m.executeDeleteChoice(m.deleteModal.focusedIndex)
	}
	return m, nil
}

func (m Model) executeDeleteChoice(idx int) (Model, tea.Cmd) {
	target := m.deleteModal.targetProfile
	isShared := m.deleteModal.isShared
	m.deleteModal = deleteModalState{}

	if !isShared {
		if idx == 1 {
			return m, nil
		}
		if m.pm != nil {
			_ = m.pm.DeleteProfile(target, m.cfg)
		}
	} else {
		if idx == 2 {
			return m, nil
		}
		if idx == 0 {
			if m.pm != nil {
				_, _ = m.pm.RemoveAgent(target, m.agent, m.cfg)
			}
		} else if idx == 1 {
			if m.pm != nil {
				_ = m.pm.DeleteProfile(target, m.cfg)
			}
		}
	}

	if m.cache != nil {
		m.cache.Delete(m.agent, target)
	}
	delete(m.reports, fmt.Sprintf("%s:%s", m.agent, target))
	delete(m.reports, target)

	m = m.refreshProfiles()
	filtered := m.filteredProfiles()
	if m.cursor >= len(filtered) {
		if len(filtered) > 0 {
			m.cursor = len(filtered) - 1
		} else {
			m.cursor = 0
		}
	}

	if len(m.profiles) > 0 {
		return m, m.triggerRefreshCmd()
	}
	return m, nil
}

func (m Model) renderDeleteModal() string {
	var b strings.Builder
	pName := m.deleteModal.targetProfile

	title := ModalTitleStyle.Render("[!] Confirm Deletion: " + pName)
	b.WriteString(title + "\n\n")

	if m.deleteModal.isShared {
		agentsStr := strings.Join(m.deleteModal.agents, ", ")
		b.WriteString(lipgloss.NewStyle().Foreground(TextSecondary).Render(
			fmt.Sprintf("Profile %q is shared across: %s", pName, agentsStr),
		) + "\n")
		b.WriteString(lipgloss.NewStyle().Foreground(TextMuted).Render(
			fmt.Sprintf("Do you want to unlink '%s' or delete the entire profile?", m.agent),
		) + "\n\n")

		btn0Style := ModalBtnInactiveStyle
		btn1Style := ModalBtnInactiveStyle
		btn2Style := ModalBtnInactiveStyle

		if m.deleteModal.focusedIndex == 0 {
			btn0Style = ModalBtnActiveStyle
		} else if m.deleteModal.focusedIndex == 1 {
			btn1Style = ModalBtnActiveStyle
		} else if m.deleteModal.focusedIndex == 2 {
			btn2Style = ModalBtnCancelActiveStyle
		}

		btn0 := btn0Style.Render(fmt.Sprintf("[1] Remove '%s' Only", m.agent))
		btn1 := btn1Style.Render("[2] Delete Entire Profile")
		btn2 := btn2Style.Render("[Cancel]")

		b.WriteString(fmt.Sprintf("  %s    %s    %s\n\n", btn0, btn1, btn2))
		b.WriteString(lipgloss.NewStyle().Foreground(TextMuted).Render(
			"  [←/→/Tab] Select  •  [Enter] Confirm  •  [Esc] Cancel",
		))
	} else {
		b.WriteString(lipgloss.NewStyle().Foreground(TextSecondary).Render(
			fmt.Sprintf("Are you sure you want to permanently delete profile %q?", pName),
		) + "\n")
		b.WriteString(lipgloss.NewStyle().Foreground(TextMuted).Render(
			"This will permanently delete all stored credentials and isolated state.",
		) + "\n\n")

		btn0Style := ModalBtnInactiveStyle
		btn1Style := ModalBtnInactiveStyle

		if m.deleteModal.focusedIndex == 0 {
			btn0Style = ModalBtnActiveStyle
		} else if m.deleteModal.focusedIndex == 1 {
			btn1Style = ModalBtnCancelActiveStyle
		}

		btn0 := btn0Style.Render("[ Delete Profile ]")
		btn1 := btn1Style.Render("[ Cancel ]")

		b.WriteString(fmt.Sprintf("      %s      %s\n\n", btn0, btn1))
		b.WriteString(lipgloss.NewStyle().Foreground(TextMuted).Render(
			"  [←/→/Tab] Select  •  [Enter/y] Confirm  •  [Esc] Cancel",
		))
	}

	box := ModalBoxStyle.Render(b.String())
	return "\n" + box + "\n"
}
