package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type helpModalState struct {
	active bool
}

type shortcutEntry struct {
	key  string
	desc string
}

func (m Model) openHelpOverlay() (Model, tea.Cmd) {
	m.helpModal = helpModalState{active: true}
	return m, nil
}

func (m Model) closeHelpOverlay() (Model, tea.Cmd) {
	m.helpModal = helpModalState{active: false}
	return m, nil
}

func (m Model) updateHelpOverlay(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.cancelStream()
		return m, tea.Quit
	case "?", "esc", "q":
		m.helpModal = helpModalState{active: false}
		return m, nil
	}
	return m, nil
}

func (m Model) renderHelpOverlay() string {
	var b strings.Builder

	title := TitleStyle.Render("⌨  AIM Keyboard Shortcuts")
	b.WriteString(title + "\n\n")

	navShortcuts := []shortcutEntry{
		{"↑ / k", "Previous profile"},
		{"↓ / j", "Next profile"},
		{"Tab", "Next agent tab"},
		{"Shift+Tab", "Prev agent tab"},
		{"1 - 9", "Jump to agent tab"},
		{"r", "Refresh quota"},
		{"Ctrl+C", "Force quit"},
	}

	actionShortcuts := []shortcutEntry{
		{"Enter", "Run profile"},
		{"s", "Open subshell"},
		{"l", "Login profile"},
		{"d", "Doctor diagnostics"},
		{"m / R", "Rename profile"},
		{"x / Del", "Delete profile"},
		{"?", "Toggle help"},
		{"q / Esc", "Close / Exit"},
	}

	col1Header := lipgloss.NewStyle().Bold(true).Foreground(TextBright).Render("Navigation & Tabs")
	col2Header := lipgloss.NewStyle().Bold(true).Foreground(TextBright).Render("Actions & Commands")

	var col1Lines []string
	col1Lines = append(col1Lines, col1Header, "")
	for _, s := range navShortcuts {
		keyStr := HintKeyStyle.Width(11).Render(s.key)
		descStr := lipgloss.NewStyle().Foreground(TextPrimary).Render(s.desc)
		col1Lines = append(col1Lines, fmt.Sprintf("%s %s", keyStr, descStr))
	}

	var col2Lines []string
	col2Lines = append(col2Lines, col2Header, "")
	for _, s := range actionShortcuts {
		keyStr := HintKeyStyle.Width(11).Render(s.key)
		descStr := lipgloss.NewStyle().Foreground(TextPrimary).Render(s.desc)
		col2Lines = append(col2Lines, fmt.Sprintf("%s %s", keyStr, descStr))
	}

	isNarrow := m.width > 0 && m.width < 70
	if isNarrow {
		allLines := append(col1Lines, "")
		allLines = append(allLines, col2Lines...)
		b.WriteString(strings.Join(allLines, "\n"))
	} else {
		col1Block := strings.Join(col1Lines, "\n")
		col2Block := strings.Join(col2Lines, "\n")
		twoCols := lipgloss.JoinHorizontal(lipgloss.Top, col1Block, "    ", col2Block)
		b.WriteString(twoCols)
	}

	footer := lipgloss.NewStyle().Foreground(TextMuted).Render(
		"Press [?] or [Esc] or [q] to close help",
	)
	b.WriteString("\n\n  " + footer)

	box := DoctorDrawerStyle.Render(b.String())
	return "\n" + box + "\n"
}
