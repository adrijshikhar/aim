package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type filterState struct {
	active bool
	input  textinput.Model
}

func newFilterState() filterState {
	ti := textinput.New()
	ti.Prompt = "/"
	ti.CharLimit = 64
	ti.Width = 30
	return filterState{
		active: false,
		input:  ti,
	}
}

// FilterInputValue returns the current text inside the filter input.
func (m Model) FilterInputValue() string {
	return m.filter.input.Value()
}

// FilteredProfiles returns the list of profiles matching the current filter query.
func (m Model) FilteredProfiles() []string {
	return m.filteredProfiles()
}

func (m Model) filteredProfiles() []string {
	query := strings.TrimSpace(m.filter.input.Value())
	query = strings.TrimPrefix(query, "/")
	if query == "" {
		return m.profiles
	}
	query = strings.ToLower(query)

	var res []string
	for _, p := range m.profiles {
		// 1. Check profile name
		if strings.Contains(strings.ToLower(p), query) {
			res = append(res, p)
			continue
		}

		// 2. Check attached agent names
		if m.cfg != nil {
			matchedAgent := false
			for _, ag := range m.cfg.GetProfileAgents(p) {
				if strings.Contains(strings.ToLower(ag), query) {
					matchedAgent = true
					break
				}
			}
			if matchedAgent {
				res = append(res, p)
			}
		}
	}
	return res
}

func (m Model) renderFilterBar() string {
	return fmt.Sprintf("  FILTER: %s  (Enter to apply, Esc to clear)\n", m.filter.input.View())
}

func (m Model) openFilter() (Model, tea.Cmd) {
	if m.filter.input.Prompt == "" && m.filter.input.CharLimit == 0 {
		m.filter = newFilterState()
	}
	m.filter.active = true
	cmd := m.filter.input.Focus()
	return m, cmd
}

func (m Model) updateFilter(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.cancelStream()
			return m, tea.Quit
		case "enter":
			m.filter.active = false
			m.filter.input.Blur()
			filtered := m.filteredProfiles()
			if m.cursor >= len(filtered) {
				if len(filtered) > 0 {
					m.cursor = len(filtered) - 1
				} else {
					m.cursor = 0
				}
			}
			return m, nil
		case "esc":
			prevSelected := ""
			filtered := m.filteredProfiles()
			if m.cursor >= 0 && m.cursor < len(filtered) {
				prevSelected = filtered[m.cursor]
			}
			m.filter.active = false
			m.filter.input.Blur()
			m.filter.input.SetValue("")
			m.cursor = 0
			if prevSelected != "" {
				for idx, p := range m.profiles {
					if p == prevSelected {
						m.cursor = idx
						break
					}
				}
			} else if m.cursor >= len(m.profiles) {
				if len(m.profiles) > 0 {
					m.cursor = len(m.profiles) - 1
				}
			}
			return m, nil
		default:
			var cmd tea.Cmd
			m.filter.input, cmd = m.filter.input.Update(msg)
			filtered := m.filteredProfiles()
			if m.cursor >= len(filtered) {
				if len(filtered) > 0 {
					m.cursor = len(filtered) - 1
				} else {
					m.cursor = 0
				}
			}
			return m, cmd
		}
	default:
		var cmd tea.Cmd
		m.filter.input, cmd = m.filter.input.Update(msg)
		return m, cmd
	}
}
