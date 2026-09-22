package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aim-cli/aim/internal/session"
	"github.com/aim-cli/aim/internal/tui"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

type sessionPickerModel struct {
	sessions     []session.Session
	cursor       int
	filterActive bool
	filterInput  textinput.Model
	selected     *session.Session
	cancelled    bool
}

func newSessionPickerModel(sessions []session.Session) sessionPickerModel {
	ti := textinput.New()
	ti.Placeholder = "type to filter..."
	ti.Prompt = "/ "
	ti.PromptStyle = lipgloss.NewStyle().Bold(true).Foreground(tui.AccentCyan)
	ti.CharLimit = 64

	return sessionPickerModel{
		sessions:    sessions,
		filterInput: ti,
		cursor:      0,
	}
}

func (m sessionPickerModel) filteredSessions() []session.Session {
	term := strings.ToLower(strings.TrimSpace(m.filterInput.Value()))
	if term == "" {
		return m.sessions
	}
	var res []session.Session
	for _, s := range m.sessions {
		if strings.Contains(strings.ToLower(s.ID), term) ||
			strings.Contains(strings.ToLower(s.ShortID), term) ||
			strings.Contains(strings.ToLower(s.Title), term) ||
			strings.Contains(strings.ToLower(s.Summary), term) ||
			strings.Contains(strings.ToLower(s.Profile), term) {
			res = append(res, s)
		}
	}
	return res
}

func (m sessionPickerModel) Init() tea.Cmd {
	return nil
}

func isEnterKey(msg tea.KeyMsg) bool {
	return msg.Type == tea.KeyEnter || msg.Type == tea.KeyCtrlJ || msg.String() == "enter" || msg.String() == "ctrl+j"
}

func (m sessionPickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			m.cancelled = true
			return m, tea.Quit
		}

		if m.filterActive {
			if isEnterKey(msg) {
				filtered := m.filteredSessions()
				if len(filtered) > 0 && m.cursor >= 0 && m.cursor < len(filtered) {
					sel := filtered[m.cursor]
					m.selected = &sel
					return m, tea.Quit
				}
				return m, nil
			}

			switch msg.Type {
			case tea.KeyEsc:
				if m.filterInput.Value() == "" {
					m.cancelled = true
					return m, tea.Quit
				}
				m.filterActive = false
				m.filterInput.Blur()
				m.filterInput.SetValue("")
				m.cursor = 0
				return m, nil
			case tea.KeyUp:
				if m.cursor > 0 {
					m.cursor--
				}
				return m, nil
			case tea.KeyDown:
				filtered := m.filteredSessions()
				if m.cursor < len(filtered)-1 {
					m.cursor++
				}
				return m, nil
			default:
				var cmd tea.Cmd
				m.filterInput, cmd = m.filterInput.Update(msg)
				filtered := m.filteredSessions()
				if m.cursor >= len(filtered) {
					if len(filtered) > 0 {
						m.cursor = len(filtered) - 1
					} else {
						m.cursor = 0
					}
				}
				return m, cmd
			}
		}

		if isEnterKey(msg) {
			filtered := m.filteredSessions()
			if len(filtered) > 0 && m.cursor >= 0 && m.cursor < len(filtered) {
				sel := filtered[m.cursor]
				m.selected = &sel
			}
			return m, tea.Quit
		}

		if len(msg.Runes) > 0 && msg.Runes[0] == '/' {
			m.filterActive = true
			cmd := m.filterInput.Focus()
			if len(msg.Runes) > 1 {
				m.filterInput.SetValue(string(msg.Runes[1:]))
				m.cursor = 0
			}
			return m, cmd
		}

		// Handle navigation runes (e.g. 'j', 'k', 'jjj', 'kkk')
		if len(msg.Runes) > 0 {
			allJ := true
			allK := true
			for _, r := range msg.Runes {
				if r != 'j' {
					allJ = false
				}
				if r != 'k' {
					allK = false
				}
			}
			if allJ {
				filtered := m.filteredSessions()
				m.cursor += len(msg.Runes)
				if m.cursor >= len(filtered) {
					m.cursor = len(filtered) - 1
				}
				return m, nil
			}
			if allK {
				m.cursor -= len(msg.Runes)
				if m.cursor < 0 {
					m.cursor = 0
				}
				return m, nil
			}
		}

		switch msg.String() {
		case "esc", "q":
			m.cancelled = true
			return m, tea.Quit
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "down":
			filtered := m.filteredSessions()
			if m.cursor < len(filtered)-1 {
				m.cursor++
			}
			return m, nil
		}
	}

	return m, nil
}

func (m sessionPickerModel) View() string {
	var b strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(tui.AccentBlue)
	helpStyle := lipgloss.NewStyle().Foreground(tui.TextMuted)

	b.WriteString(headerStyle.Render("Select session to resume:") + " ")
	b.WriteString(helpStyle.Render("(↑/k up, ↓/j down, / filter, enter select, esc/q cancel)") + "\n")

	if m.filterActive || m.filterInput.Value() != "" {
		b.WriteString(m.filterInput.View() + "\n")
	}

	filtered := m.filteredSessions()
	if len(filtered) == 0 {
		b.WriteString("  " + lipgloss.NewStyle().Foreground(tui.TextMuted).Render("(no sessions found matching filter)") + "\n")
		return b.String()
	}

	const maxVisible = 10
	start := 0
	if m.cursor >= maxVisible {
		start = m.cursor - maxVisible + 1
	}
	end := start + maxVisible
	if end > len(filtered) {
		end = len(filtered)
		start = end - maxVisible
		if start < 0 {
			start = 0
		}
	}

	for i := start; i < end; i++ {
		s := filtered[i]

		var profileBadge string
		if s.IsHost || s.Profile == "" || s.Profile == "host" {
			profileBadge = "[host]"
		} else {
			profileBadge = fmt.Sprintf("[%s]", s.Profile)
		}

		shortID := s.ShortID
		if shortID == "" && len(s.ID) >= 8 {
			shortID = s.ID[:8]
		}

		title := s.Title
		if title == "" {
			title = s.Summary
		}
		if title == "" {
			title = "(untitled)"
		}
		title = truncateString(title, 40)

		relTime := formatRelativeTime(s.LastActiveAt)
		if s.Status == session.StatusActive {
			relTime = "ACTIVE"
		}

		badgeCol := lipgloss.NewStyle().Width(12).Render(profileBadge)
		idCol := lipgloss.NewStyle().Width(10).Render(shortID)
		titleCol := lipgloss.NewStyle().Width(42).Render(title)
		timeCol := lipgloss.NewStyle().Width(12).Render(relTime)

		cursorStr := "  "
		if i == m.cursor {
			cursorStr = "❯ "
		}

		rowContent := fmt.Sprintf("%s%s %s %s %s", cursorStr, badgeCol, idCol, titleCol, timeCol)
		if i == m.cursor {
			b.WriteString(tui.SelectedRowStyle.Render(rowContent) + "\n")
		} else {
			b.WriteString(tui.NormalRowStyle.Render(rowContent) + "\n")
		}
	}

	if len(filtered) > maxVisible {
		scrollInfo := fmt.Sprintf("  (showing %d-%d of %d sessions)", start+1, end, len(filtered))
		b.WriteString(lipgloss.NewStyle().Foreground(tui.TextMuted).Render(scrollInfo) + "\n")
	}

	return b.String()
}

var sessionPickerRunner = func(m sessionPickerModel, in io.Reader, out io.Writer) (*session.Session, error) {
	p := tea.NewProgram(
		m,
		tea.WithInput(in),
		tea.WithOutput(out),
	)
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}
	res, ok := finalModel.(sessionPickerModel)
	if !ok || res.cancelled {
		return nil, nil
	}
	return res.selected, nil
}

func promptSelectSession(cmd *cobra.Command, sessions []session.Session) (*session.Session, error) {
	if len(sessions) == 0 {
		return nil, nil
	}

	var in io.Reader = os.Stdin
	var out io.Writer = os.Stdout
	if cmd != nil {
		if cmd.InOrStdin() != nil {
			in = cmd.InOrStdin()
		}
		if cmd.OutOrStdout() != nil {
			out = cmd.OutOrStdout()
		}
	}

	m := newSessionPickerModel(sessions)
	return sessionPickerRunner(m, in, out)
}
