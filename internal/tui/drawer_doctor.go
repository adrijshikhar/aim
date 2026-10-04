package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/daemon"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DoctorDrawer struct {
	active        bool
	loading       bool
	targetProfile string
	targetAgent   string
	results       []agents.DiagnosticResult

	keys          DoctorDrawerKeyMap
	help          help.Model
	pendingAction tea.Msg
}

var _ SubModel = DoctorDrawer{}

type DoctorCloseMsg struct{}
type DoctorQuitMsg struct{}
type DoctorNavigateProfileMsg struct {
	Delta int
}
type DoctorSelectAgentMsg struct {
	Index int
}
type DoctorCycleAgentMsg struct {
	Forward bool
}

type doctorDiagnosticsLoadedMsg struct {
	targetAgent   string
	targetProfile string
	results       []agents.DiagnosticResult
}

func (d DoctorDrawer) Init() tea.Cmd {
	return nil
}

func (d DoctorDrawer) Update(msg tea.Msg) (SubModel, tea.Cmd) {
	switch msg := msg.(type) {
	case doctorDiagnosticsLoadedMsg:
		d.loading = false
		if msg.targetAgent == d.targetAgent && (msg.targetProfile == d.targetProfile || d.targetProfile == "") {
			d.results = msg.results
		}
		return d, nil

	case tea.KeyMsg:
		km := d.keys
		if len(km.Quit.Keys()) == 0 {
			km = DefaultDoctorDrawerKeyMap()
		}

		switch {
		case key.Matches(msg, km.Quit):
			d.pendingAction = DoctorQuitMsg{}
			return d, func() tea.Msg { return DoctorQuitMsg{} }
		case key.Matches(msg, km.Close):
			d.pendingAction = DoctorCloseMsg{}
			return d, func() tea.Msg { return DoctorCloseMsg{} }
		case key.Matches(msg, km.Up):
			d.pendingAction = DoctorNavigateProfileMsg{Delta: -1}
			return d, func() tea.Msg { return DoctorNavigateProfileMsg{Delta: -1} }
		case key.Matches(msg, km.Down):
			d.pendingAction = DoctorNavigateProfileMsg{Delta: 1}
			return d, func() tea.Msg { return DoctorNavigateProfileMsg{Delta: 1} }
		case key.Matches(msg, km.Agent1):
			d.pendingAction = DoctorSelectAgentMsg{Index: 0}
			return d, func() tea.Msg { return DoctorSelectAgentMsg{Index: 0} }
		case key.Matches(msg, km.Agent2):
			d.pendingAction = DoctorSelectAgentMsg{Index: 1}
			return d, func() tea.Msg { return DoctorSelectAgentMsg{Index: 1} }
		case key.Matches(msg, km.Agent3):
			d.pendingAction = DoctorSelectAgentMsg{Index: 2}
			return d, func() tea.Msg { return DoctorSelectAgentMsg{Index: 2} }
		case key.Matches(msg, km.NextAgent):
			d.pendingAction = DoctorCycleAgentMsg{Forward: true}
			return d, func() tea.Msg { return DoctorCycleAgentMsg{Forward: true} }
		case key.Matches(msg, km.PrevAgent):
			d.pendingAction = DoctorCycleAgentMsg{Forward: false}
			return d, func() tea.Msg { return DoctorCycleAgentMsg{Forward: false} }
		}
	}
	return d, nil
}

func (d DoctorDrawer) View() string {
	var b strings.Builder
	target := d.targetProfile
	if target == "" {
		target = "(none)"
	}
	title := lipgloss.NewStyle().Bold(true).Foreground(AccentBlue).Render(
		fmt.Sprintf("🩺  Diagnostics: %s / %s", d.targetAgent, target),
	)
	b.WriteString(title + "\n\n")

	for _, r := range d.results {
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

	km := d.keys
	if len(km.Quit.Keys()) == 0 {
		km = DefaultDoctorDrawerKeyMap()
	}
	h := d.help
	if h.Width == 0 {
		h = NewThemedHelp()
	}
	b.WriteString("\n  " + h.ShortHelpView(km.ShortHelp()))

	box := DoctorDrawerStyle.Render(b.String())
	return "\n" + box + "\n"
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
				if diag, ok := ad.(agents.Diagnostician); ok {
					results = diag.Doctor(ctx, targetProfile, pDir)
				}
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

		if info, err := daemon.Status(config.BaseDir()); err == nil && info != nil {
			if info.Active {
				results = append(results, agents.DiagnosticResult{
					Category: "Daemon",
					Status:   "OK",
					Message:  "Active (15m interval)",
				})
			} else if info.Installed {
				results = append(results, agents.DiagnosticResult{
					Category: "Daemon",
					Status:   "WARN",
					Message:  "Installed but inactive",
				})
			} else {
				results = append(results, agents.DiagnosticResult{
					Category: "Daemon",
					Status:   "INFO",
					Message:  "Not installed (run 'aim daemon install')",
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
	m.doctorDrawer.keys = m.keys.DoctorDrawer
	m.doctorDrawer.help = m.help
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
		m.doctorDrawer = DoctorDrawer{
			active:        true,
			targetAgent:   m.agent,
			targetProfile: "(none)",
			results:       results,
			keys:          m.keys.DoctorDrawer,
			help:          m.help,
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
			if diag, ok := ad.(agents.Diagnostician); ok {
				results = diag.Doctor(context.Background(), p, pDir)
			}
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

	if info, err := daemon.Status(config.BaseDir()); err == nil && info != nil {
		if info.Active {
			results = append(results, agents.DiagnosticResult{
				Category: "Daemon",
				Status:   "OK",
				Message:  "Active (15m interval)",
			})
		} else if info.Installed {
			results = append(results, agents.DiagnosticResult{
				Category: "Daemon",
				Status:   "WARN",
				Message:  "Installed but inactive",
			})
		} else {
			results = append(results, agents.DiagnosticResult{
				Category: "Daemon",
				Status:   "INFO",
				Message:  "Not installed (run 'aim daemon install')",
			})
		}
	}

	m.doctorDrawer = DoctorDrawer{
		active:        true,
		targetProfile: p,
		targetAgent:   m.agent,
		results:       results,
		keys:          m.keys.DoctorDrawer,
		help:          m.help,
	}
	return m
}

func (m Model) updateDoctorDrawer(msg tea.KeyMsg) (Model, tea.Cmd) {
	m.doctorDrawer.keys = m.keys.DoctorDrawer
	m.doctorDrawer.help = m.help
	sub, cmd := m.doctorDrawer.Update(msg)
	d := sub.(DoctorDrawer)
	action := d.pendingAction
	d.pendingAction = nil
	m.doctorDrawer = d

	if action != nil {
		switch act := action.(type) {
		case DoctorQuitMsg:
			m.cancelStream()
			return m, tea.Quit
		case DoctorCloseMsg:
			m.doctorDrawer = DoctorDrawer{}
			return m, nil
		case DoctorNavigateProfileMsg:
			if act.Delta < 0 && m.cursor > 0 {
				m.cursor--
				m = m.fetchDoctorDiagnostics()
				m.doctorDrawer.loading = true
				return m, m.fetchDoctorDiagnosticsCmd()
			} else if act.Delta > 0 && m.cursor < len(m.profiles)-1 {
				m.cursor++
				m = m.fetchDoctorDiagnostics()
				m.doctorDrawer.loading = true
				return m, m.fetchDoctorDiagnosticsCmd()
			}
			return m, nil
		case DoctorSelectAgentMsg:
			return m.selectAgentByIndex(act.Index)
		case DoctorCycleAgentMsg:
			return m.cycleAgent(act.Forward)
		}
	}
	return m, cmd
}

func (m Model) renderDoctorDrawer() string {
	d := m.doctorDrawer
	d.keys = m.keys.DoctorDrawer
	d.help = m.help
	return d.View()
}
