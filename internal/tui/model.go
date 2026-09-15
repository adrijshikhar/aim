package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/session"
	"github.com/aim-cli/aim/internal/usage"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type ActionOutcome int

const (
	ActionNone ActionOutcome = iota
	ActionRun
	ActionShell
	ActionLogin
	ActionResumeExact
	ActionResumeCatalyst
)

type usageReportMsg usage.Report
type usageStreamClosedMsg struct{}
type usageBatchMsg []usage.Report

type usageStream struct {
	ch     <-chan usage.Report
	cancel context.CancelFunc
}

// Version is the package-level version string shown in the TUI header.
var Version = "0.4.0"

type Model struct {
	reg      *agents.Registry
	pm       *profile.ProfileManager
	cfg      *config.Config
	profiles []string
	cursor   int
	agent    string
	outcome  ActionOutcome
	selected string

	cache       *usage.CacheStore
	reports     map[string]usage.Report
	width       int
	usageChan   <-chan usage.Report
	usageStream *usageStream

	loading    bool
	spinner    spinner.Model
	spinnerIdx int
	version    string

	deleteModal    deleteModalState
	renameModal    renameModalState
	moveModal      moveModalState
	doctorDrawer   doctorDrawerState
	sessionsDrawer sessionsDrawerState
	resumeModal    resumeModalState
	helpModal      helpModalState
	filter         filterState
	keys           KeyMap

	selectedSession *session.Session
}

func NewModel(reg *agents.Registry, pm *profile.ProfileManager, cfg *config.Config) Model {
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}
	baseDir := config.BaseDir()
	if pm != nil && pm.BaseDir != "" {
		baseDir = pm.BaseDir
	}
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = CursorStyle

	m := Model{
		reg:         reg,
		pm:          pm,
		cfg:         cfg,
		agent:       "agy",
		version:     Version,
		reports:     make(map[string]usage.Report),
		cache:       usage.NewCacheStore(baseDir, usage.DefaultTTL),
		usageStream: &usageStream{},
		loading:     false,
		spinner:     s,
		keys:        DefaultKeyMap(),
		filter:      newFilterState(),
	}
	m = m.refreshProfiles()
	m = m.loadCachedReports()
	if len(m.profiles) > 0 && len(m.reports) < len(m.profiles) {
		m.loading = true
	}
	return m
}

func (m Model) loadCachedReports() Model {
	if m.cache == nil {
		return m
	}
	for _, p := range m.profiles {
		if rep, found := m.cache.Get(m.agent, p); found {
			if m.reports == nil {
				m.reports = make(map[string]usage.Report)
			}
			m.reports[fmt.Sprintf("%s:%s", m.agent, p)] = rep
		}
	}
	return m
}

func (m Model) refreshProfiles() Model {
	if m.pm == nil {
		m.profiles = nil
		return m
	}
	profs, _ := m.pm.ListProfilesForAgent(m.agent, m.cfg, m.reg)
	m.profiles = profs
	if m.cursor >= len(m.profiles) {
		m.cursor = 0
	}
	return m
}

// RunTUI runs the interactive TUI program and returns any execution error.
func RunTUI(reg *agents.Registry, pm *profile.ProfileManager, cfg *config.Config) error {
	m := NewModel(reg, pm, cfg)
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}

func (m Model) Outcome() ActionOutcome {
	return m.outcome
}

func (m Model) SelectedProfile() string {
	return m.selected
}

func (m Model) SelectedSession() *session.Session {
	return m.selectedSession
}

func (m Model) SelectedAgent() string {
	return m.agent
}

func (m Model) KeyMap() KeyMap {
	return m.keys
}

func (m Model) Profiles() []string {
	return m.profiles
}

// IsHelpActive reports whether the help cheatsheet overlay is currently active.
func (m Model) IsHelpActive() bool {
	return m.helpModal.active
}

// IsFilterActive reports whether the profile filter bar is currently active.
func (m Model) IsFilterActive() bool {
	return m.filter.active
}

// WithVersion sets the version string displayed in the header and returns the updated model.
func (m Model) WithVersion(v string) Model {
	m.version = v
	return m
}

// SetVersion sets the version string displayed in the header on the model pointer.
func (m *Model) SetVersion(v string) {
	m.version = v
}

// Version returns the version displayed in the header, falling back to package Version.
func (m Model) Version() string {
	if m.version != "" {
		return m.version
	}
	return Version
}

func (m Model) getRegisteredAgentNames() []string {
	preferred := []string{"agy", "gemini", "codex"}
	if m.reg == nil {
		return preferred
	}
	var res []string
	seen := make(map[string]bool)
	for _, name := range preferred {
		if _, err := m.reg.Get(name); err == nil {
			res = append(res, name)
			seen[name] = true
		}
	}
	for _, a := range m.reg.All() {
		if !seen[a.Name()] {
			res = append(res, a.Name())
			seen[a.Name()] = true
		}
	}
	if len(res) == 0 {
		return preferred
	}
	return res
}

func (m Model) switchAgent(targetAgent string) (Model, tea.Cmd) {
	m.agent = targetAgent
	m.cursor = 0
	m = m.refreshProfiles()
	m = m.loadCachedReports()
	if m.doctorDrawer.active {
		m = m.fetchDoctorDiagnostics()
	}
	m.loading = true
	return m, tea.Batch(m.triggerRefreshCmd(), m.spinTickCmd())
}

func (m Model) cycleAgent(forward bool) (Model, tea.Cmd) {
	agents := m.getRegisteredAgentNames()
	if len(agents) == 0 {
		return m, nil
	}
	curIdx := -1
	for i, a := range agents {
		if a == m.agent {
			curIdx = i
			break
		}
	}
	var nextIdx int
	if forward {
		if curIdx == -1 || curIdx+1 >= len(agents) {
			nextIdx = 0
		} else {
			nextIdx = curIdx + 1
		}
	} else {
		if curIdx <= 0 {
			nextIdx = len(agents) - 1
		} else {
			nextIdx = curIdx - 1
		}
	}
	return m.switchAgent(agents[nextIdx])
}

func (m Model) selectAgentByIndex(index int) (Model, tea.Cmd) {
	agents := m.getRegisteredAgentNames()
	if index >= 0 && index < len(agents) {
		return m.switchAgent(agents[index])
	}
	return m, nil
}

func (m Model) getReport(prof string) (usage.Report, bool) {
	if m.reports == nil {
		return usage.Report{}, false
	}
	if rep, ok := m.reports[fmt.Sprintf("%s:%s", m.agent, prof)]; ok {
		return rep, true
	}
	rep, ok := m.reports[prof]
	return rep, ok
}

func (m Model) getReportForAgent(agent, prof string) (usage.Report, bool) {
	if m.reports != nil {
		if agent != "" {
			if rep, ok := m.reports[fmt.Sprintf("%s:%s", agent, prof)]; ok {
				return rep, true
			}
		}
		if rep, ok := m.reports[prof]; ok {
			return rep, true
		}
	}
	if m.cache != nil && agent != "" {
		if rep, ok := m.cache.Get(agent, prof); ok {
			return rep, true
		}
	}
	return m.getReport(prof)
}

func waitForUsageReport(ch <-chan usage.Report) tea.Cmd {
	if ch == nil {
		return nil
	}
	return func() tea.Msg {
		rep, ok := <-ch
		if !ok {
			return usageStreamClosedMsg{}
		}
		return usageReportMsg(rep)
	}
}

func (m Model) triggerRefreshCmd(force ...bool) tea.Cmd {
	reg := m.reg
	if reg == nil {
		reg = agents.DefaultRegistry()
	}

	var targets []usage.TargetProfile
	if reg != nil {
		if ad, err := reg.Get(m.agent); err == nil && ad != nil {
			for _, p := range m.profiles {
				pDir := ""
				if m.pm != nil {
					pDir = m.pm.ProfileDir(p)
				}
				targets = append(targets, usage.TargetProfile{
					Agent:      ad.Name(),
					Profile:    p,
					ProfileDir: pDir,
					GetUsageFn: ad.GetUsage,
				})
			}
		}
	}

	if m.usageStream == nil {
		m.usageStream = &usageStream{}
	}
	if m.usageStream.cancel != nil {
		m.usageStream.cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	m.usageStream.cancel = cancel

	ch := usage.RefreshAsync(ctx, targets, m.cache, force...)
	m.usageStream.ch = ch
	return waitForUsageReport(ch)
}

var spinnerFrames = spinner.MiniDot.Frames

type spinnerTickMsg time.Time

func (m Model) spinTickCmd() tea.Cmd {
	return m.spinner.Tick
}

func spinTickCmd() tea.Cmd {
	return spinner.Tick
}

type tickMsg time.Time

func tickEvery(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) cancelStream() {
	if m.usageStream != nil && m.usageStream.cancel != nil {
		m.usageStream.cancel()
		m.usageStream.cancel = nil
	}
}

func formatBadge(rep usage.Report, isNarrow bool) string {
	if rep.Error != "" || rep.Status == usage.StatusUnknown {
		errLower := strings.ToLower(rep.Error)
		summaryLower := strings.ToLower(rep.Summary)
		if strings.Contains(errLower, "credential") || strings.Contains(summaryLower, "credential") {
			return "[no credentials]"
		}
		if strings.Contains(errLower, "offline") || strings.Contains(summaryLower, "offline") ||
			strings.Contains(errLower, "connect") || strings.Contains(errLower, "network") ||
			strings.Contains(errLower, "timeout") {
			return "[offline]"
		}
		if rep.Error != "" {
			return fmt.Sprintf("[%s]", strings.ToLower(rep.Error))
		}
		if rep.Status != "" {
			return fmt.Sprintf("[%s]", strings.ToLower(string(rep.Status)))
		}
		return ""
	}

	if len(rep.Windows) == 0 {
		return ""
	}

	// Always report the bottleneck / most constrained limit percentage so the
	// displayed percentage is strictly consistent with the badge color/status.
	return fmt.Sprintf("[%d%%]", rep.BottleneckPct())
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.triggerRefreshCmd(), m.spinTickCmd(), tickEvery(5*time.Minute))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			m.spinnerIdx = (m.spinnerIdx + 1) % len(spinnerFrames)
			return m, cmd
		}
		return m, nil

	case spinnerTickMsg:
		if m.loading {
			m.spinnerIdx = (m.spinnerIdx + 1) % len(spinnerFrames)
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(spinner.TickMsg{})
			return m, tea.Batch(cmd, m.spinTickCmd())
		}
		return m, nil

	case tickMsg:
		m.loading = true
		return m, tea.Batch(m.triggerRefreshCmd(), m.spinTickCmd(), tickEvery(5*time.Minute))

	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil

	case usageReportMsg:
		if m.reports == nil {
			m.reports = make(map[string]usage.Report)
		}
		rep := usage.Report(msg)
		m.reports[fmt.Sprintf("%s:%s", rep.Agent, rep.Profile)] = rep
		var ch <-chan usage.Report
		if m.usageStream != nil {
			ch = m.usageStream.ch
		}
		m.usageChan = ch
		return m, waitForUsageReport(ch)

	case usageStreamClosedMsg:
		m.loading = false
		m.usageChan = nil
		if m.usageStream != nil {
			m.usageStream.ch = nil
		}
		return m, nil

	case usageBatchMsg:
		m.loading = false
		if m.reports == nil {
			m.reports = make(map[string]usage.Report)
		}
		for _, rep := range msg {
			m.reports[fmt.Sprintf("%s:%s", rep.Agent, rep.Profile)] = rep
		}
		return m, nil

	case tea.KeyMsg:
		if m.deleteModal.active {
			return m.updateDeleteModal(msg)
		}

		if m.renameModal.active {
			return m.updateRenameModal(msg)
		}

		if m.moveModal.active {
			return m.updateMoveModal(msg)
		}

		if m.doctorDrawer.active {
			return m.updateDoctorDrawer(msg)
		}

		if m.resumeModal.active {
			return m.updateResumeModal(msg)
		}

		if m.sessionsDrawer.active {
			return m.updateSessionsDrawer(msg)
		}

		if m.helpModal.active {
			return m.updateHelpOverlay(msg)
		}

		if m.filter.active {
			return m.updateFilter(msg)
		}

		keys := m.keys
		if len(keys.Quit.Keys()) == 0 {
			keys = DefaultKeyMap()
		}

		switch {
		case key.Matches(msg, keys.Quit):
			if m.filter.input.Value() != "" && msg.String() == "esc" {
				prevSelected := ""
				filtered := m.filteredProfiles()
				if m.cursor >= 0 && m.cursor < len(filtered) {
					prevSelected = filtered[m.cursor]
				}
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
			}
			m.cancelStream()
			return m, tea.Quit
		case key.Matches(msg, keys.Up):
			filtered := m.filteredProfiles()
			if m.cursor >= len(filtered) && len(filtered) > 0 {
				m.cursor = len(filtered) - 1
			}
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, keys.Down):
			filtered := m.filteredProfiles()
			if m.cursor < len(filtered)-1 {
				m.cursor++
			}
		case msg.String() == "1":
			return m.selectAgentByIndex(0)
		case msg.String() == "2":
			return m.selectAgentByIndex(1)
		case msg.String() == "3":
			return m.selectAgentByIndex(2)
		case key.Matches(msg, keys.Tab):
			return m.cycleAgent(true)
		case msg.String() == "shift+tab":
			return m.cycleAgent(false)
		case key.Matches(msg, keys.Refresh):
			m.loading = true
			return m, tea.Batch(m.triggerRefreshCmd(true), m.spinTickCmd())
		case key.Matches(msg, keys.Run):
			filtered := m.filteredProfiles()
			if len(filtered) > 0 && m.cursor >= 0 && m.cursor < len(filtered) {
				m.selected = filtered[m.cursor]
				m.outcome = ActionRun
				m.cancelStream()
				return m, tea.Quit
			}
		case key.Matches(msg, keys.Sessions):
			return m.openSessionsDrawer()
		case key.Matches(msg, keys.Shell):
			filtered := m.filteredProfiles()
			if len(filtered) > 0 && m.cursor >= 0 && m.cursor < len(filtered) {
				m.selected = filtered[m.cursor]
				m.outcome = ActionShell
				m.cancelStream()
				return m, tea.Quit
			}
		case key.Matches(msg, keys.Login):
			m.outcome = ActionLogin
			m.cancelStream()
			return m, tea.Quit
		case key.Matches(msg, keys.Doctor):
			return m.openDoctorDrawer()
		case key.Matches(msg, keys.Rename):
			return m.openRenameModal()
		case key.Matches(msg, keys.Move):
			return m.openMoveModal()
		case key.Matches(msg, keys.Delete):
			return m.openDeleteModal()
		case key.Matches(msg, keys.Filter), msg.String() == "/":
			return m.openFilter()
		case key.Matches(msg, keys.Help), msg.String() == "?":
			return m.openHelpOverlay()
		}

	default:
		if m.renameModal.active {
			return m.updateRenameModal(msg)
		}
		if m.moveModal.active {
			return m.updateMoveModal(msg)
		}
		if m.filter.active {
			return m.updateFilter(msg)
		}
	}
	return m, nil
}

func (m Model) View() string {
	var s strings.Builder
	s.WriteString(m.renderHeader())
	s.WriteString(m.renderTabBar())

	if m.filter.active || m.filter.input.Value() != "" {
		s.WriteString(m.renderFilterBar())
	}

	isNarrow := m.width > 0 && m.width < 85

	s.WriteString("  PROFILES:\n")
	filtered := m.filteredProfiles()
	if len(filtered) == 0 {
		if m.filter.input.Value() != "" {
			s.WriteString(fmt.Sprintf("    (no profiles matching %q)\n", m.filter.input.Value()))
		} else {
			s.WriteString(fmt.Sprintf("    (no profiles configured for %s - press 'l' to log in)\n", m.agent))
		}
	} else {
		for i, p := range filtered {
			prefix := "    "
			style := NormalRowStyle
			if i == m.cursor {
				prefix = CursorStyle.Render("  > ")
				style = SelectedRowStyle
			}
			label := p
			if m.cfg != nil {
				agentsList := m.cfg.GetProfileAgents(p)
				if len(agentsList) > 1 {
					label = fmt.Sprintf("%s [%s]", p, strings.Join(agentsList, ", "))
				}
			}

			badgeStr := ""
			if rep, ok := m.getReport(p); ok {
				badge := formatBadge(rep, isNarrow)
				if badge != "" {
					gaugeStyle := GaugeStyleForStatus(rep.Status)
					badgeStr = "  " + gaugeStyle.Render(badge)
				}
			}

			s.WriteString(fmt.Sprintf("%s%s%s\n", prefix, style.Render(label), badgeStr))
		}
	}

	if m.deleteModal.active {
		s.WriteString(m.renderDeleteModal())
		return s.String()
	}

	if m.renameModal.active {
		s.WriteString(m.renderRenameModal())
		return s.String()
	}

	if m.moveModal.active {
		s.WriteString(m.renderMoveModal())
		return s.String()
	}

	if m.doctorDrawer.active {
		s.WriteString(m.renderDoctorDrawer())
		return s.String()
	}

	if m.resumeModal.active {
		s.WriteString(m.renderResumeModal())
		return s.String()
	}

	if m.sessionsDrawer.active {
		s.WriteString(m.renderSessionsDrawer())
		return s.String()
	}

	if m.helpModal.active {
		s.WriteString(m.renderHelpOverlay())
		return s.String()
	}

	// Bottom inspector section when a profile is highlighted
	if len(filtered) > 0 && m.cursor >= 0 && m.cursor < len(filtered) {
		curProfile := filtered[m.cursor]
		s.WriteString(m.renderInspector(curProfile))
	}

	refreshHint := HintKeyStyle.Render("[r]") + " " + HintLabelStyle.Render("Refresh Quota  ")
	if m.loading {
		spinnerChar := m.spinner.View()
		refreshHint = HintKeyStyle.Render("[r]") + " " + spinnerChar + " " + HintLabelStyle.Render("Refresh Quota  ")
	}

	s.WriteString("\n  " +
		HintKeyStyle.Render("[Enter]") + " " + HintLabelStyle.Render("Run  ") +
		HintKeyStyle.Render("[s]") + " " + HintLabelStyle.Render("Sessions  ") +
		HintKeyStyle.Render("[S]") + " " + HintLabelStyle.Render("Shell  ") +
		HintKeyStyle.Render("[l]") + " " + HintLabelStyle.Render("Login  ") +
		HintKeyStyle.Render("[Tab]") + " " + HintLabelStyle.Render("Switch Agent  ") +
		HintKeyStyle.Render("[d]") + " " + HintLabelStyle.Render("Doctor  ") +
		HintKeyStyle.Render("[m]") + " " + HintLabelStyle.Render("Rename  ") +
		HintKeyStyle.Render("[x]") + " " + HintLabelStyle.Render("Delete  ") +
		HintKeyStyle.Render("[/]") + " " + HintLabelStyle.Render("Filter  ") +
		refreshHint +
		HintKeyStyle.Render("[?]") + " " + HintLabelStyle.Render("Help  ") +
		HintKeyStyle.Render("[q]") + " " + HintLabelStyle.Render("Quit") + "\n")

	return s.String()
}
