package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/aim-cli/aim/internal/agents"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type doctorDrawerState struct {
	active        bool
	targetProfile string
	targetAgent   string
	results       []agents.DiagnosticResult
}

func (m Model) IsDoctorDrawerActive() bool {
	return m.doctorDrawer.active
}

func (m Model) DoctorDrawerResults() []agents.DiagnosticResult {
	return m.doctorDrawer.results
}

func (m Model) DoctorDrawerTargetProfile() string {
	return m.doctorDrawer.targetProfile
}

func (m Model) openDoctorDrawer() (Model, tea.Cmd) {
	m = m.fetchDoctorDiagnostics()
	return m, nil
}

func (m Model) fetchDoctorDiagnostics() Model {
	reg := m.reg

	if len(m.profiles) == 0 || m.cursor < 0 || m.cursor >= len(m.profiles) {
		var results []agents.DiagnosticResult
		results = append(results, agents.DiagnosticResult{
			Category: "Profile",
			Status:   "WARN",
			Message:  fmt.Sprintf("No profiles configured for %s", m.agent),
		})
		if reg != nil {
			if ad, err := reg.Get(m.agent); err == nil && ad != nil {
				results = append(results, ad.Doctor(context.Background(), "", "")...)
			}
		}
		m.doctorDrawer = doctorDrawerState{
			active:        true,
			targetAgent:   m.agent,
			targetProfile: "(none)",
			results:       results,
		}
		return m
	}

	p := m.profiles[m.cursor]
	pDir := ""
	if m.pm != nil {
		pDir = m.pm.ProfileDir(p)
	}

	var results []agents.DiagnosticResult
	if reg != nil {
		if ad, err := reg.Get(m.agent); err == nil && ad != nil {
			results = ad.Doctor(context.Background(), p, pDir)
		}
	}
	if len(results) == 0 {
		results = []agents.DiagnosticResult{
			{Category: "Status", Status: "OK", Message: "All checks passed"},
		}
	}
	if m.cfg != nil {
		if env := m.cfg.GetProfileEnv(p); len(env) > 0 {
			results = append(results, agents.DiagnosticResult{
				Category: "Config",
				Status:   "OK",
				Message:  fmt.Sprintf("%d custom env var(s) configured", len(env)),
			})
		}
		if args := m.cfg.GetProfileArgs(p); len(args) > 0 {
			results = append(results, agents.DiagnosticResult{
				Category: "Config",
				Status:   "OK",
				Message:  fmt.Sprintf("%d custom launch arg(s) configured", len(args)),
			})
		}
	}

	m.doctorDrawer = doctorDrawerState{
		active:        true,
		targetProfile: p,
		targetAgent:   m.agent,
		results:       results,
	}
	return m
}

func (m Model) updateDoctorDrawer(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.cancelStream()
		return m, tea.Quit
	case "d", "esc", "q":
		m.doctorDrawer = doctorDrawerState{}
		return m, nil
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
			m = m.fetchDoctorDiagnostics()
		}
		return m, nil
	case "down", "j":
		if m.cursor < len(m.profiles)-1 {
			m.cursor++
			m = m.fetchDoctorDiagnostics()
		}
		return m, nil
	case "1":
		return m.selectAgentByIndex(0)
	case "2":
		return m.selectAgentByIndex(1)
	case "3":
		return m.selectAgentByIndex(2)
	case "tab":
		return m.cycleAgent(true)
	case "shift+tab":
		return m.cycleAgent(false)
	}
	return m, nil
}

func (m Model) renderDoctorDrawer() string {
	var b strings.Builder
	target := m.doctorDrawer.targetProfile
	if target == "" {
		target = "(none)"
	}
	title := lipgloss.NewStyle().Bold(true).Foreground(AccentBlue).Render(
		fmt.Sprintf("🩺  Diagnostics: %s / %s", m.doctorDrawer.targetAgent, target),
	)
	b.WriteString(title + "\n\n")

	for _, r := range m.doctorDrawer.results {
		var badgeStyle lipgloss.Style
		switch r.Status {
		case "OK":
			badgeStyle = GaugeGreenStyle
		case "WARN":
			badgeStyle = GaugeYellowStyle
		case "FAIL":
			badgeStyle = GaugeRedStyle
		default:
			badgeStyle = GaugeDimStyle
		}

		badge := badgeStyle.Width(8).Render(fmt.Sprintf("[%s]", r.Status))
		cat := lipgloss.NewStyle().Bold(true).Foreground(TextPrimary).Width(14).Render(r.Category + ":")
		msg := lipgloss.NewStyle().Foreground(TextSecondary).Render(r.Message)

		b.WriteString(fmt.Sprintf("  %s %s %s\n", badge, cat, msg))
	}

	b.WriteString("\n" + lipgloss.NewStyle().Foreground(TextMuted).Render(
		"  [↑/↓] Inspect Profile  •  [Tab] Switch Agent  •  [d/Esc/q] Close Drawer",
	))

	box := DoctorDrawerStyle.Render(b.String())
	return "\n" + box + "\n"
}
