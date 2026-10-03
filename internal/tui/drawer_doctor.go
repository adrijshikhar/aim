package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type doctorDrawerState struct {
	active        bool
	loading       bool
	targetProfile string
	targetAgent   string
	results       []agents.DiagnosticResult
}

type doctorDiagnosticsLoadedMsg struct {
	targetAgent   string
	targetProfile string
	results       []agents.DiagnosticResult
}

func (m Model) IsDoctorDrawerActive() bool {
	return m.doctorDrawer.active
}

func (m Model) IsDoctorLoading() bool {
	return m.doctorDrawer.loading
}

func (m Model) DoctorDrawerResults() []agents.DiagnosticResult {
	return m.doctorDrawer.results
}

func (m Model) DoctorDrawerTargetProfile() string {
	return m.doctorDrawer.targetProfile
}

func (m Model) fetchDoctorDiagnosticsCmd() tea.Cmd {
	targetAgent := m.agent
	if len(m.profiles) == 0 || m.cursor < 0 || m.cursor >= len(m.profiles) {
		return nil
	}
	targetProfile := m.profiles[m.cursor]
	var pDir string
	if m.pm != nil {
		pDir = m.pm.ProfileDir(targetProfile)
	}
	reg := m.reg
	cfg := m.cfg

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var results []agents.DiagnosticResult
		if reg != nil {
			if ad, err := reg.Get(targetAgent); err == nil && ad != nil {
				results = ad.Doctor(ctx, targetProfile, pDir)
			}
		}
		if len(results) == 0 {
			results = []agents.DiagnosticResult{
				{Category: "Status", Status: "OK", Message: "All checks passed"},
			}
		}
		if pDir != "" {
			var extraPaths []string
			if cfg != nil {
				extraPaths = cfg.CustomBridgedPaths
			}
			results = append(results, profile.BridgeDiagnostics(targetProfile, config.RealHomeDir(), pDir, extraPaths...)...)
		}
		if cfg != nil {
			if env := cfg.GetProfileEnv(targetProfile); len(env) > 0 {
				results = append(results, agents.DiagnosticResult{
					Category: "Config",
					Status:   "OK",
					Message:  fmt.Sprintf("%d custom env var(s) configured", len(env)),
				})
			}
			if args := cfg.GetProfileArgs(targetProfile); len(args) > 0 {
				results = append(results, agents.DiagnosticResult{
					Category: "Config",
					Status:   "OK",
					Message:  fmt.Sprintf("%d custom launch arg(s) configured", len(args)),
				})
			}
		}

		return doctorDiagnosticsLoadedMsg{
			targetAgent:   targetAgent,
			targetProfile: targetProfile,
			results:       results,
		}
	}
}

func (m Model) openDoctorDrawer() (Model, tea.Cmd) {
	m = m.fetchDoctorDiagnostics()
	m.doctorDrawer.loading = true
	return m, m.fetchDoctorDiagnosticsCmd()
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
		// No profile: do not run Doctor with an empty profile dir — adapters
		// resolve files relative to it and would write into the working dir.
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
	if pDir != "" {
		var extraPaths []string
		if m.cfg != nil {
			extraPaths = m.cfg.CustomBridgedPaths
		}
		results = append(results, profile.BridgeDiagnostics(p, config.RealHomeDir(), pDir, extraPaths...)...)
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
	km := m.keys.DoctorDrawer

	switch {
	case key.Matches(msg, km.Quit):
		m.cancelStream()
		return m, tea.Quit
	case key.Matches(msg, km.Close):
		m.doctorDrawer = doctorDrawerState{}
		return m, nil
	case key.Matches(msg, km.Up):
		if m.cursor > 0 {
			m.cursor--
			m = m.fetchDoctorDiagnostics()
			m.doctorDrawer.loading = true
			return m, m.fetchDoctorDiagnosticsCmd()
		}
		return m, nil
	case key.Matches(msg, km.Down):
		if m.cursor < len(m.profiles)-1 {
			m.cursor++
			m = m.fetchDoctorDiagnostics()
			m.doctorDrawer.loading = true
			return m, m.fetchDoctorDiagnosticsCmd()
		}
		return m, nil
	case key.Matches(msg, km.Agent1):
		return m.selectAgentByIndex(0)
	case key.Matches(msg, km.Agent2):
		return m.selectAgentByIndex(1)
	case key.Matches(msg, km.Agent3):
		return m.selectAgentByIndex(2)
	case key.Matches(msg, km.NextAgent):
		return m.cycleAgent(true)
	case key.Matches(msg, km.PrevAgent):
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

	km := m.keys.DoctorDrawer
	b.WriteString("\n  " + m.help.ShortHelpView(km.ShortHelp()))

	box := DoctorDrawerStyle.Render(b.String())
	return "\n" + box + "\n"
}
