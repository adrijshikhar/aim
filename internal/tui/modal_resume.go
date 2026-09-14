package tui

import (
	"fmt"
	"strings"

	"github.com/aim-cli/aim/internal/session"
	"github.com/aim-cli/aim/internal/usage"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type resumeModalState struct {
	active     bool
	session    *session.Session
	outcome    ActionOutcome
	fork       bool
	profiles   []string
	cursor     int
	customMode bool
	input      textinput.Model
}

func (m Model) IsResumeModalActive() bool {
	return m.resumeModal.active
}

func (m Model) ResumeModalSession() *session.Session {
	return m.resumeModal.session
}

func (m Model) ResumeModalProfiles() []string {
	return m.resumeModal.profiles
}

func (m Model) ResumeModalCursor() int {
	return m.resumeModal.cursor
}

func (m Model) ResumeModalIsCustomMode() bool {
	return m.resumeModal.customMode
}

func (m Model) openResumeModal(target *session.Session, outcome ActionOutcome, fork bool) (Model, tea.Cmd) {
	if target == nil {
		return m, nil
	}

	// Collect configured profiles
	var allProfiles []string
	if m.pm != nil {
		if proflist, err := m.pm.ListProfiles(); err == nil {
			allProfiles = proflist
		}
	}
	if len(allProfiles) == 0 {
		allProfiles = m.profiles
	}

	// Ensure the session's original profile is first in the list
	origProfile := target.Profile
	if origProfile == "" || origProfile == "<host>" {
		origProfile = "default"
	}

	var ordered []string
	ordered = append(ordered, origProfile)
	for _, p := range allProfiles {
		if p != origProfile && p != "" && p != "<host>" {
			ordered = append(ordered, p)
		}
	}

	ti := textinput.New()
	ti.Placeholder = "custom-profile-name"
	ti.CharLimit = 64
	ti.Width = 32

	m.resumeModal = resumeModalState{
		active:     true,
		session:    target,
		outcome:    outcome,
		fork:       fork,
		profiles:   ordered,
		cursor:     0,
		customMode: false,
		input:      ti,
	}

	return m, nil
}

func (m Model) updateResumeModal(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.resumeModal.customMode {
			switch msg.String() {
			case "ctrl+c":
				m.cancelStream()
				return m, tea.Quit
			case "esc":
				m.resumeModal.customMode = false
				return m, nil
			case "enter":
				val := strings.TrimSpace(m.resumeModal.input.Value())
				if val == "" {
					if m.resumeModal.session.Profile != "" && m.resumeModal.session.Profile != "<host>" {
						val = m.resumeModal.session.Profile
					} else {
						val = "default"
					}
				}
				m.selected = val
				m.selectedSession = m.resumeModal.session
				m.outcome = m.resumeModal.outcome
				m.sessionsDrawer.fork = m.resumeModal.fork
				m.resumeModal.active = false
				m.cancelStream()
				return m, tea.Quit
			default:
				var cmd tea.Cmd
				m.resumeModal.input, cmd = m.resumeModal.input.Update(msg)
				return m, cmd
			}
		}

		totalOptions := len(m.resumeModal.profiles) + 1 // +1 for "Enter custom profile name..."
		switch msg.String() {
		case "ctrl+c":
			m.cancelStream()
			return m, tea.Quit
		case "esc", "q":
			m.resumeModal.active = false
			return m, nil
		case "up", "k":
			if m.resumeModal.cursor > 0 {
				m.resumeModal.cursor--
			}
			return m, nil
		case "down", "j":
			if m.resumeModal.cursor < totalOptions-1 {
				m.resumeModal.cursor++
			}
			return m, nil
		case "enter":
			if m.resumeModal.cursor == len(m.resumeModal.profiles) {
				// User selected "Enter custom profile name..."
				m.resumeModal.customMode = true
				cmd := m.resumeModal.input.Focus()
				return m, cmd
			}

			// User selected an existing profile
			chosen := m.resumeModal.profiles[m.resumeModal.cursor]
			m.selected = chosen
			m.selectedSession = m.resumeModal.session
			m.outcome = m.resumeModal.outcome
			m.sessionsDrawer.fork = m.resumeModal.fork
			m.resumeModal.active = false
			m.cancelStream()
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m Model) renderResumeModal() string {
	var b strings.Builder
	sess := m.resumeModal.session
	if sess == nil {
		return ""
	}

	modeStr := "Exact Resume (Verbatim)"
	if m.resumeModal.outcome == ActionResumeCatalyst {
		modeStr = "Catalyst Summary Handoff"
	} else if m.resumeModal.fork {
		modeStr = "Forked Session (Branch)"
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(AccentCyan).Render(fmt.Sprintf("🚀 Resume Session: %s", sess.ShortID))
	b.WriteString(title + "\n\n")

	infoStyle := lipgloss.NewStyle().Foreground(TextMuted)
	valStyle := lipgloss.NewStyle().Foreground(TextPrimary)
	b.WriteString(fmt.Sprintf("  %s %s   %s %s   %s %s\n\n",
		infoStyle.Render("Agent:"),
		lipgloss.NewStyle().Bold(true).Foreground(AccentBlue).Render(sess.Agent),
		infoStyle.Render("Original:"),
		lipgloss.NewStyle().Bold(true).Foreground(StatusYellow).Render(sess.Profile),
		infoStyle.Render("Mode:"),
		valStyle.Render(modeStr),
	))

	if m.resumeModal.customMode {
		b.WriteString(lipgloss.NewStyle().Foreground(TextSecondary).Render("Enter new target profile name:") + "\n\n")
		b.WriteString("  " + m.resumeModal.input.View() + "\n\n")
		b.WriteString(lipgloss.NewStyle().Foreground(TextMuted).Render(
			"  [Enter] Confirm & Resume  •  [Esc] Back to Profile List",
		))
	} else {
		b.WriteString(lipgloss.NewStyle().Foreground(TextSecondary).Render("Choose target profile for resumption:") + "\n\n")

		// Calculate max label length for clean column alignment
		maxLabelLen := 0
		for _, p := range m.resumeModal.profiles {
			l := len(p)
			if p == sess.Profile {
				l += len(" (original)")
			}
			if l > maxLabelLen {
				maxLabelLen = l
			}
		}

		for i, p := range m.resumeModal.profiles {
			isOriginal := p == sess.Profile
			cursor := "  "
			itemStyle := lipgloss.NewStyle().Foreground(TextPrimary)
			origBadge := ""
			if isOriginal {
				origBadge = lipgloss.NewStyle().Foreground(TextMuted).Render(" (original)")
			}

			if i == m.resumeModal.cursor {
				cursor = lipgloss.NewStyle().Bold(true).Foreground(AccentCyan).Render("> ")
				itemStyle = lipgloss.NewStyle().Bold(true).Foreground(AccentCyan)
			}

			bullet := "○ "
			if isOriginal {
				bullet = "● "
			}

			// Calculate padding for aligned columns
			curLen := len(p)
			if isOriginal {
				curLen += len(" (original)")
			}
			padding := ""
			if maxLabelLen > curLen {
				padding = strings.Repeat(" ", maxLabelLen-curLen)
			}

			// Usage quota / limit remaining badge
			quotaBadge := ""
			if rep, ok := m.getReportForAgent(sess.Agent, p); ok {
				badge := formatBadge(rep, false)
				if badge != "" {
					gaugeStyle := GaugeStyleForStatus(rep.Status)
					quotaBadge = "  " + gaugeStyle.Render(badge)

					// Show short reset duration if available and quota is not 100%
					if rep.PrimaryWindow() != nil && rep.PrimaryWindow().ResetsIn > 0 && rep.BottleneckPct() < 100 {
						resetStr := lipgloss.NewStyle().Foreground(TextDim).Render(fmt.Sprintf("(%s)", usage.FormatDuration(rep.PrimaryWindow().ResetsIn)))
						quotaBadge += " " + resetStr
					}
				}
			}

			// Active session count for profile
			activeCount := 0
			for _, s := range m.sessionsDrawer.sessions {
				if s.Profile == p && s.Status == session.StatusActive {
					activeCount++
				}
			}
			activeBadge := ""
			if activeCount > 0 {
				activeBadge = " " + lipgloss.NewStyle().Foreground(StatusYellow).Render(fmt.Sprintf("(%d active)", activeCount))
			}

			b.WriteString(fmt.Sprintf("  %s%s%s%s%s%s%s\n", cursor, bullet, itemStyle.Render(p), origBadge, padding, quotaBadge, activeBadge))
		}

		// Custom option
		customIdx := len(m.resumeModal.profiles)
		cursor := "  "
		itemStyle := lipgloss.NewStyle().Foreground(TextMuted)
		if m.resumeModal.cursor == customIdx {
			cursor = lipgloss.NewStyle().Bold(true).Foreground(AccentCyan).Render("> ")
			itemStyle = lipgloss.NewStyle().Bold(true).Foreground(AccentCyan)
		}
		b.WriteString(fmt.Sprintf("  %s+ %s\n\n", cursor, itemStyle.Render("Enter custom profile name...")))

		// Show contextual status hint if selected profile has warnings
		if m.resumeModal.cursor >= 0 && m.resumeModal.cursor < len(m.resumeModal.profiles) {
			selProf := m.resumeModal.profiles[m.resumeModal.cursor]
			activeCount := 0
			for _, s := range m.sessionsDrawer.sessions {
				if s.Profile == selProf && s.Status == session.StatusActive {
					activeCount++
				}
			}

			if rep, ok := m.getReportForAgent(sess.Agent, selProf); ok {
				if rep.Status == usage.StatusCritical || rep.Status == usage.StatusExhausted {
					b.WriteString(lipgloss.NewStyle().Foreground(StatusRed).Render(
						fmt.Sprintf("  ⚠️  Notice: Profile '%s' has low remaining limit (%d%%)\n\n", selProf, rep.BottleneckPct()),
					))
				} else if activeCount > 0 {
					b.WriteString(lipgloss.NewStyle().Foreground(StatusYellow).Render(
						fmt.Sprintf("  ℹ️  Notice: Profile '%s' currently has %d active session running\n\n", selProf, activeCount),
					))
				}
			} else if activeCount > 0 {
				b.WriteString(lipgloss.NewStyle().Foreground(StatusYellow).Render(
					fmt.Sprintf("  ℹ️  Notice: Profile '%s' currently has %d active session running\n\n", selProf, activeCount),
				))
			}
		}

		b.WriteString(lipgloss.NewStyle().Foreground(TextMuted).Render(
			"  [↑/↓] Select Profile  •  [Enter] Confirm & Resume  •  [Esc] Back",
		))
	}

	box := ResumeModalBoxStyle.Render(b.String())
	return "\n" + box + "\n"
}
