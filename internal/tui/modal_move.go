package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type moveModalState struct {
	active        bool
	agent         string
	sourceProfile string
	input         textinput.Model
	err           string
	forceRequired bool
}

func (m Model) IsMoveModalActive() bool {
	return m.moveModal.active
}

func (m Model) MoveModalSource() string {
	return m.moveModal.sourceProfile
}

func (m Model) MoveModalAgent() string {
	return m.moveModal.agent
}

func (m Model) MoveModalInputValue() string {
	return m.moveModal.input.Value()
}

func (m Model) MoveModalError() string {
	return m.moveModal.err
}

func (m Model) openMoveModal() (Model, tea.Cmd) {
	filtered := m.filteredProfiles()
	if len(filtered) == 0 || m.cursor < 0 || m.cursor >= len(filtered) {
		return m, nil
	}
	target := filtered[m.cursor]
	currentAgent := m.agent
	if currentAgent == "" || currentAgent == "all" {
		// If on "all" tab, find first agent for this profile
		if m.cfg != nil {
			agents := m.cfg.GetProfileAgents(target)
			if len(agents) > 0 {
				currentAgent = agents[0]
			}
		}
	}
	if currentAgent == "" || currentAgent == "all" {
		currentAgent = "agy"
	}

	ti := textinput.New()
	ti.Placeholder = "destination profile name"
	ti.CharLimit = 64
	ti.Width = 30
	cmd := ti.Focus()
	m.moveModal = moveModalState{
		active:        true,
		agent:         currentAgent,
		sourceProfile: target,
		input:         ti,
	}
	return m, cmd
}

func (m Model) updateMoveModal(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.cancelStream()
			return m, tea.Quit
		case "esc":
			m.moveModal = moveModalState{}
			return m, nil
		case "enter":
			return m.executeMove()
		default:
			// If typing after collision warning, reset forceRequired
			m.moveModal.forceRequired = false
			var cmd tea.Cmd
			m.moveModal.input, cmd = m.moveModal.input.Update(msg)
			return m, cmd
		}
	default:
		var cmd tea.Cmd
		m.moveModal.input, cmd = m.moveModal.input.Update(msg)
		return m, cmd
	}
}

func (m Model) executeMove() (Model, tea.Cmd) {
	dstProfile := strings.TrimSpace(m.moveModal.input.Value())
	if dstProfile == "" {
		m.moveModal.err = "Destination profile name cannot be empty"
		return m, nil
	}
	if dstProfile == m.moveModal.sourceProfile {
		m.moveModal.err = "Destination profile must be different from source profile"
		return m, nil
	}
	if strings.ContainsAny(dstProfile, "/\\") || strings.Contains(dstProfile, "..") {
		m.moveModal.err = "Profile name cannot contain slashes or '..'"
		return m, nil
	}

	force := m.moveModal.forceRequired
	if m.pm != nil {
		err := m.pm.MoveAgent(m.moveModal.agent, m.moveModal.sourceProfile, dstProfile, force, m.cfg, m.reg)
		if err != nil {
			if strings.Contains(err.Error(), "--force") && !force {
				m.moveModal.forceRequired = true
				m.moveModal.err = fmt.Sprintf("Target profile %q already has credentials for %s.\nPress [Enter] again to overwrite, or [Esc] to cancel.", dstProfile, m.moveModal.agent)
				return m, nil
			}
			m.moveModal.err = err.Error()
			return m, nil
		}
	}

	srcProfile := m.moveModal.sourceProfile
	agent := m.moveModal.agent

	// Update cache & reports
	if m.cache != nil {
		if rep, ok := m.cache.Get(agent, srcProfile); ok {
			rep.Profile = dstProfile
			_ = m.cache.Put(rep)
		}
	}
	if m.reports != nil {
		srcKey := fmt.Sprintf("%s:%s", agent, srcProfile)
		if rep, ok := m.reports[srcKey]; ok {
			delete(m.reports, srcKey)
			rep.Profile = dstProfile
			m.reports[fmt.Sprintf("%s:%s", agent, dstProfile)] = rep
		}
	}

	m.moveModal = moveModalState{}
	m = m.refreshProfiles()
	filtered := m.filteredProfiles()
	m.cursor = 0
	for idx, p := range filtered {
		if p == dstProfile {
			m.cursor = idx
			break
		}
	}
	return m, nil
}

func (m Model) renderMoveModal() string {
	var b strings.Builder
	pName := m.moveModal.sourceProfile
	ag := m.moveModal.agent

	title := MoveModalTitleStyle.Render(fmt.Sprintf("⇄ Move Account: %s (%s)", ag, pName))
	b.WriteString(title + "\n\n")

	b.WriteString(lipgloss.NewStyle().Foreground(TextSecondary).Render(
		fmt.Sprintf("Move agent '%s' from profile '%s' to destination:", ag, pName),
	) + "\n\n")

	b.WriteString("  " + m.moveModal.input.View() + "\n\n")

	if m.moveModal.err != "" {
		errColor := StatusRed
		if m.moveModal.forceRequired {
			errColor = StatusYellow
		}
		b.WriteString(lipgloss.NewStyle().Foreground(errColor).Bold(true).Render(
			"  ✕ "+m.moveModal.err,
		) + "\n\n")
	}

	b.WriteString(lipgloss.NewStyle().Foreground(TextMuted).Render(
		"  [Enter] Confirm  •  [Esc] Cancel",
	))

	box := MoveModalBoxStyle.Render(b.String())
	return "\n" + box + "\n"
}
