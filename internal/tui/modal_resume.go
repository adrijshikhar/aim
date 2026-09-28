package tui

import (
	"fmt"
	"strings"

	"github.com/aim-cli/aim/internal/session"
	"github.com/aim-cli/aim/internal/usage"
	"github.com/charmbracelet/bubbles/key"
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
	flagsMode  bool
	flagsInput textinput.Model
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

func (m Model) ResumeModalIsFlagsMode() bool {
	return m.resumeModal.flagsMode
}

func (m Model) ResumeModalFlagsInput() string {
	return m.resumeModal.flagsInput.Value()
}

func (m Model) openResumeModal(target *session.Session, outcome ActionOutcome, fork bool) (Model, tea.Cmd) {
	return m.openResumeModalWithFlags(target, outcome, fork, false)
}

func (m Model) openResumeModalWithFlags(target *session.Session, outcome ActionOutcome, fork bool, initialFlagsMode bool) (Model, tea.Cmd) {
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

	fi := textinput.New()
	fi.Placeholder = "e.g. --yolo -m o3 --search"
	fi.CharLimit = 128
	fi.Width = 45
	fi.Prompt = "Flags: "
	fi.PromptStyle = lipgloss.NewStyle().Bold(true).Foreground(AccentCyan)

	var cmd tea.Cmd
	if initialFlagsMode {
		cmd = fi.Focus()
	}

	m.resumeModal = resumeModalState{
		active:     true,
		session:    target,
		outcome:    outcome,
		fork:       fork,
		profiles:   ordered,
		cursor:     0,
		customMode: false,
		input:      ti,
		flagsMode:  initialFlagsMode,
		flagsInput: fi,
	}

	return m, cmd
}

func (m Model) updateResumeModal(msg tea.Msg) (Model, tea.Cmd) {
	km := m.keys.ResumeModal
	if len(km.Quit.Keys()) == 0 {
		km = DefaultResumeModalKeyMap()
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.resumeModal.customMode {
			switch {
			case key.Matches(msg, km.Quit):
				m.cancelStream()
				return m, tea.Quit
			case msg.String() == "esc":
				m.resumeModal.customMode = false
				return m, nil
			case key.Matches(msg, km.Enter):
				val := strings.TrimSpace(m.resumeModal.input.Value())
				if val == "" {
					if m.resumeModal.session.Profile != "" && m.resumeModal.session.Profile != "<host>" {
						val = m.resumeModal.session.Profile
					} else {
						val = "default"
					}
				}
				return m.finishResumeSelection(val)
			default:
				var cmd tea.Cmd
				m.resumeModal.input, cmd = m.resumeModal.input.Update(msg)
				return m, cmd
			}
		}

		if m.resumeModal.flagsMode {
			switch {
			case key.Matches(msg, km.Quit):
				m.cancelStream()
				return m, tea.Quit
			case key.Matches(msg, km.DoneFlags):
				m.resumeModal.flagsMode = false
				m.resumeModal.flagsInput.Blur()
				return m, nil
			case key.Matches(msg, km.Enter):
				// Confirm with current selected profile and flags
				profile := m.resumeModal.profiles[m.resumeModal.cursor]
				return m.finishResumeSelection(profile)
			default:
				var cmd tea.Cmd
				m.resumeModal.flagsInput, cmd = m.resumeModal.flagsInput.Update(msg)
				return m, cmd
			}
		}

		totalOptions := len(m.resumeModal.profiles) + 1 // +1 for "Enter custom profile name..."
		switch {
		case key.Matches(msg, km.Quit):
			m.cancelStream()
			return m, tea.Quit
		case key.Matches(msg, km.Cancel):
			m.resumeModal.active = false
			return m, nil
		case key.Matches(msg, km.Up):
			if m.resumeModal.cursor > 0 {
				m.resumeModal.cursor--
			}
			return m, nil
		case key.Matches(msg, km.Down):
			if m.resumeModal.cursor < totalOptions-1 {
				m.resumeModal.cursor++
			}
			return m, nil
		case key.Matches(msg, km.ToggleFlags):
			m.resumeModal.flagsMode = true
			cmd := m.resumeModal.flagsInput.Focus()
			return m, cmd
		case key.Matches(msg, km.Enter):
			if m.resumeModal.cursor == len(m.resumeModal.profiles) {
				// User selected "Enter custom profile name..."
				m.resumeModal.customMode = true
				cmd := m.resumeModal.input.Focus()
				return m, cmd
			}

			// User selected an existing profile
			return m.finishResumeSelection(m.resumeModal.profiles[m.resumeModal.cursor])
		}
	}

	return m, nil
}

func (m Model) finishResumeSelection(profile string) (Model, tea.Cmd) {
	m.selected = profile
	m.selectedSession = m.resumeModal.session
	m.outcome = m.resumeModal.outcome
	m.sessionsDrawer.fork = m.resumeModal.fork
	rawFlags := strings.TrimSpace(m.resumeModal.flagsInput.Value())
	if rawFlags != "" {
		m.selectedArgs = splitArgs(rawFlags)
	} else {
		m.selectedArgs = nil
	}
	m.resumeModal.active = false
	m.cancelStream()
	return m, tea.Quit
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

	title := lipgloss.NewStyle().Bold(true).Foreground(AccentCyan).Render(fmt.Sprintf("Resume Session: %s", sess.ShortID))
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

			b.WriteString(fmt.Sprintf("  %s%s%s%s%s%s\n", cursor, itemStyle.Render(p), origBadge, padding, quotaBadge, activeBadge))
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

		flagsHeader := lipgloss.NewStyle().Foreground(TextSecondary).Render("Extra CLI Flags (optional):")
		b.WriteString("  " + flagsHeader + "\n")
		km := m.keys.ResumeModal
		if len(km.Quit.Keys()) == 0 {
			km = DefaultResumeModalKeyMap()
		}
		if m.resumeModal.flagsMode {
			b.WriteString("  " + m.resumeModal.flagsInput.View() + "\n\n")
			b.WriteString("  " + m.help.ShortHelpView(km.ShortHelpFlags()))
		} else {
			val := m.resumeModal.flagsInput.Value()
			valDisplay := val
			flagsBoxStyle := lipgloss.NewStyle().Bold(true).Foreground(AccentCyan)
			if valDisplay == "" {
				valDisplay = "(none — press [f] or [Tab] to add e.g. --yolo)"
				flagsBoxStyle = lipgloss.NewStyle().Foreground(TextMuted)
			}
			b.WriteString(fmt.Sprintf("  Flags: %s\n\n", flagsBoxStyle.Render(valDisplay)))
			b.WriteString("  " + m.help.ShortHelpView(km.ShortHelpProfile()))
		}
	}

	box := ResumeModalBoxStyle.Render(b.String())
	return "\n" + box + "\n"
}

func splitArgs(s string) []string {
	var args []string
	var current strings.Builder
	inSingle := false
	inDouble := false
	escaped := false

	for _, r := range s {
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' && !inSingle {
			escaped = true
			continue
		}
		if r == '\'' && !inDouble {
			inSingle = !inSingle
			continue
		}
		if r == '"' && !inSingle {
			inDouble = !inDouble
			continue
		}
		if (r == ' ' || r == '\t') && !inSingle && !inDouble {
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
			continue
		}
		current.WriteRune(r)
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args
}
