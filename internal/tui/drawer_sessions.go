package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aim-cli/aim/internal/session"
	"github.com/aim-cli/aim/internal/session/providers/agy"
	"github.com/aim-cli/aim/internal/session/providers/codex"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var SessionsDrawerStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(AccentBlue).
	Padding(1, 2)

type sessionsDrawerState struct {
	active       bool
	sessions     []session.Session
	cursor       int
	filterActive bool
	filterInput  textinput.Model
	agentFilter  string
	fork         bool
}

func (m Model) IsSessionsDrawerActive() bool {
	return m.sessionsDrawer.active
}

func (m Model) SessionsDrawerList() []session.Session {
	return m.sessionsDrawer.sessions
}

func (m *Model) SetSessionsForTest(sessions []session.Session) {
	m.sessionsDrawer.sessions = sessions
}

func (m Model) IsForkResume() bool {
	return m.sessionsDrawer.fork
}

func (m Model) openSessionsDrawer() (Model, tea.Cmd) {
	ti := textinput.New()
	ti.Placeholder = "Filter sessions by title, id, or profile..."
	ti.CharLimit = 64
	ti.Prompt = "Filter: "
	ti.PromptStyle = lipgloss.NewStyle().Bold(true).Foreground(AccentCyan)

	m.sessionsDrawer = sessionsDrawerState{
		active:      true,
		filterInput: ti,
		agentFilter: m.agent, // Default to currently selected agent tab
		cursor:      0,
	}
	m = m.fetchSessions()
	return m, nil
}

func (m Model) fetchSessions() Model {
	mgr := session.NewManager()
	mgr.RegisterProvider(agy.NewProvider())
	mgr.RegisterProvider(codex.NewProvider())

	sessions, err := mgr.ListSessions(context.Background(), m.sessionsDrawer.agentFilter, "", false)
	if err != nil {
		sessions = []session.Session{}
	}
	m.sessionsDrawer.sessions = sessions
	if m.sessionsDrawer.cursor >= len(sessions) {
		if len(sessions) > 0 {
			m.sessionsDrawer.cursor = len(sessions) - 1
		} else {
			m.sessionsDrawer.cursor = 0
		}
	}
	return m
}

func (m Model) filteredSessions() []session.Session {
	term := strings.ToLower(strings.TrimSpace(m.sessionsDrawer.filterInput.Value()))
	if term == "" {
		return m.sessionsDrawer.sessions
	}
	var res []session.Session
	for _, s := range m.sessionsDrawer.sessions {
		if strings.Contains(strings.ToLower(s.ID), term) ||
			strings.Contains(strings.ToLower(s.ShortID), term) ||
			strings.Contains(strings.ToLower(s.Title), term) ||
			strings.Contains(strings.ToLower(s.Profile), term) ||
			strings.Contains(strings.ToLower(s.Agent), term) {
			res = append(res, s)
		}
	}
	return res
}

func (m Model) updateSessionsDrawer(msg tea.KeyMsg) (Model, tea.Cmd) {
	if m.sessionsDrawer.filterActive {
		switch msg.Type {
		case tea.KeyEsc, tea.KeyEnter:
			m.sessionsDrawer.filterActive = false
			return m, nil
		case tea.KeyCtrlC:
			m.cancelStream()
			return m, tea.Quit
		}
		var cmd tea.Cmd
		m.sessionsDrawer.filterInput, cmd = m.sessionsDrawer.filterInput.Update(msg)
		m.sessionsDrawer.cursor = 0
		return m, cmd
	}

	switch msg.String() {
	case "ctrl+c":
		m.cancelStream()
		return m, tea.Quit
	case "s", "esc", "q":
		m.sessionsDrawer = sessionsDrawerState{}
		return m, nil
	case "/":
		m.sessionsDrawer.filterActive = true
		m.sessionsDrawer.filterInput.Focus()
		return m, textinput.Blink
	case "up", "k":
		if m.sessionsDrawer.cursor > 0 {
			m.sessionsDrawer.cursor--
		}
		return m, nil
	case "down", "j":
		filtered := m.filteredSessions()
		if m.sessionsDrawer.cursor < len(filtered)-1 {
			m.sessionsDrawer.cursor++
		}
		return m, nil
	case "enter":
		filtered := m.filteredSessions()
		if len(filtered) > 0 && m.sessionsDrawer.cursor >= 0 && m.sessionsDrawer.cursor < len(filtered) {
			target := filtered[m.sessionsDrawer.cursor]
			m.selectedSession = &target
			m.selected = target.Profile
			m.outcome = ActionResumeExact
			m.cancelStream()
			return m, tea.Quit
		}
	case "c":
		filtered := m.filteredSessions()
		if len(filtered) > 0 && m.sessionsDrawer.cursor >= 0 && m.sessionsDrawer.cursor < len(filtered) {
			target := filtered[m.sessionsDrawer.cursor]
			m.selectedSession = &target
			m.selected = target.Profile
			m.outcome = ActionResumeCatalyst
			m.cancelStream()
			return m, tea.Quit
		}
	case "b", "f":
		filtered := m.filteredSessions()
		if len(filtered) > 0 && m.sessionsDrawer.cursor >= 0 && m.sessionsDrawer.cursor < len(filtered) {
			target := filtered[m.sessionsDrawer.cursor]
			m.selectedSession = &target
			m.selected = target.Profile
			m.sessionsDrawer.fork = true
			m.outcome = ActionResumeExact
			m.cancelStream()
			return m, tea.Quit
		}
	case "tab":
		if m.sessionsDrawer.agentFilter == "agy" {
			m.sessionsDrawer.agentFilter = "codex"
		} else if m.sessionsDrawer.agentFilter == "codex" {
			m.sessionsDrawer.agentFilter = ""
		} else {
			m.sessionsDrawer.agentFilter = "agy"
		}
		m.sessionsDrawer.cursor = 0
		m = m.fetchSessions()
		return m, nil
	case "1":
		m.sessionsDrawer.agentFilter = "agy"
		m.sessionsDrawer.cursor = 0
		m = m.fetchSessions()
		return m, nil
	case "2":
		m.sessionsDrawer.agentFilter = "codex"
		m.sessionsDrawer.cursor = 0
		m = m.fetchSessions()
		return m, nil
	case "0":
		m.sessionsDrawer.agentFilter = ""
		m.sessionsDrawer.cursor = 0
		m = m.fetchSessions()
		return m, nil
	}

	return m, nil
}

func (m Model) renderSessionsDrawer() string {
	var b strings.Builder

	// Top Bar
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(AccentBlue)
	filterTabStyle := func(active bool) lipgloss.Style {
		if active {
			return lipgloss.NewStyle().Bold(true).Foreground(AccentCyan).Background(BgTabActive).Padding(0, 1)
		}
		return lipgloss.NewStyle().Foreground(TextMuted).Padding(0, 1)
	}

	allActive := m.sessionsDrawer.agentFilter == ""
	agyActive := m.sessionsDrawer.agentFilter == "agy"
	codexActive := m.sessionsDrawer.agentFilter == "codex"

	topBar := fmt.Sprintf("%s  %s %s %s  %s",
		titleStyle.Render("⚡ Sessions Explorer"),
		filterTabStyle(allActive).Render("[0] All"),
		filterTabStyle(agyActive).Render("[1] Antigravity"),
		filterTabStyle(codexActive).Render("[2] Codex"),
		lipgloss.NewStyle().Foreground(TextMuted).Render("| [Tab] Cycle | [/] Filter | [Esc] Close"),
	)
	b.WriteString(topBar + "\n\n")

	if m.sessionsDrawer.filterActive || m.sessionsDrawer.filterInput.Value() != "" {
		b.WriteString(m.sessionsDrawer.filterInput.View() + "\n\n")
	}

	// Columns header
	colProfile := lipgloss.NewStyle().Bold(true).Foreground(AccentBlue).Width(10).Render("PROFILE")
	colAgent := lipgloss.NewStyle().Bold(true).Foreground(AccentBlue).Width(8).Render("AGENT")
	colID := lipgloss.NewStyle().Bold(true).Foreground(AccentBlue).Width(12).Render("SESSION ID")
	colTitle := lipgloss.NewStyle().Bold(true).Foreground(AccentBlue).Width(32).Render("TITLE")
	colActive := lipgloss.NewStyle().Bold(true).Foreground(AccentBlue).Width(14).Render("LAST ACTIVE")
	colMode := lipgloss.NewStyle().Bold(true).Foreground(AccentBlue).Render("MODE")

	b.WriteString(fmt.Sprintf("  %s %s %s %s %s %s\n", colProfile, colAgent, colID, colTitle, colActive, colMode))

	filtered := m.filteredSessions()
	if len(filtered) == 0 {
		b.WriteString("\n  " + lipgloss.NewStyle().Foreground(TextMuted).Render("(no conversation sessions found matching filter)") + "\n\n")
	} else {
		for i, s := range filtered {
			cursorStr := "  "
			rowStyle := NormalRowStyle
			if i == m.sessionsDrawer.cursor {
				cursorStr = "> "
				rowStyle = SelectedRowStyle
			}

			titleStr := truncateString(s.Title, 30)
			if titleStr == "" {
				titleStr = "(untitled)"
			}

			lastActiveStr := formatRelativeTime(s.LastActiveAt)
			if s.Status == session.StatusActive {
				lastActiveStr = "ACTIVE"
			}

			modeStr := "Catalyst Ready"
			if s.Status == session.StatusActive {
				modeStr = "In-Progress"
			}

			cProfile := lipgloss.NewStyle().Width(10).Render(s.Profile)
			cAgent := lipgloss.NewStyle().Width(8).Render(s.Agent)
			cID := lipgloss.NewStyle().Width(12).Render(s.ShortID)
			cTitle := lipgloss.NewStyle().Width(32).Render(titleStr)
			cActive := lipgloss.NewStyle().Width(14).Render(lastActiveStr)
			cMode := lipgloss.NewStyle().Foreground(TextMuted).Render(modeStr)

			rowContent := fmt.Sprintf("%s%s %s %s %s %s %s", cursorStr, cProfile, cAgent, cID, cTitle, cActive, cMode)
			b.WriteString(rowStyle.Render(rowContent) + "\n")
		}
	}

	b.WriteString("\n" + lipgloss.NewStyle().Foreground(TextMuted).Render(
		"  [↑/↓] Navigate  •  [Enter] Exact Resume  •  [c] Catalyst Handoff  •  [b] Fork  •  [Esc] Close",
	))

	box := SessionsDrawerStyle.Render(b.String())
	return "\n" + box + "\n"
}

func truncateString(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}

func formatRelativeTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	diff := time.Since(t)
	if diff < 0 {
		diff = 0
	}
	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	} else if diff < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	} else if diff < 48*time.Hour {
		return "Yesterday"
	} else if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%d days ago", days)
	}
	return t.Format("Jan 02")
}
