package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/usage"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ActionOutcome int

const (
	ActionNone ActionOutcome = iota
	ActionRun
	ActionShell
	ActionLogin
)

type usageReportMsg usage.Report
type usageStreamClosedMsg struct{}
type usageBatchMsg []usage.Report

type usageStream struct {
	ch     <-chan usage.Report
	cancel context.CancelFunc
}

type deleteModalState struct {
	active        bool
	targetProfile string
	isShared      bool
	agents        []string
	focusedIndex  int
}

type renameModalState struct {
	active        bool
	targetProfile string
	input         textinput.Model
	err           string
}

type doctorDrawerState struct {
	active        bool
	targetProfile string
	targetAgent   string
	results       []agents.DiagnosticResult
}

// Version is the package-level version string shown in the TUI header.
var Version = "0.1.0"

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

	deleteModal  deleteModalState
	renameModal  renameModalState
	doctorDrawer doctorDrawerState
	keys         KeyMap
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

func (m Model) SelectedAgent() string {
	return m.agent
}

func (m Model) KeyMap() KeyMap {
	return m.keys
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

func (m Model) Profiles() []string {
	return m.profiles
}

func (m Model) IsDeleteModalActive() bool {
	return m.deleteModal.active
}

func (m Model) DeleteModalTarget() string {
	return m.deleteModal.targetProfile
}

func (m Model) DeleteModalFocusedIndex() int {
	return m.deleteModal.focusedIndex
}

func (m Model) IsRenameModalActive() bool {
	return m.renameModal.active
}

func (m Model) RenameModalTarget() string {
	return m.renameModal.targetProfile
}

func (m Model) RenameModalInputValue() string {
	return m.renameModal.input.Value()
}

func (m Model) RenameModalError() string {
	return m.renameModal.err
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

func (m Model) fetchDoctorDiagnostics() Model {
	reg := m.reg
	if reg == nil {
		reg = agents.DefaultRegistry()
	}

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
			switch msg.String() {
			case "ctrl+c":
				m.cancelStream()
				return m, tea.Quit
			case "esc", "q", "n":
				m.deleteModal = deleteModalState{}
				return m, nil
			case "left", "h":
				numOpts := 2
				if m.deleteModal.isShared {
					numOpts = 3
				}
				m.deleteModal.focusedIndex = (m.deleteModal.focusedIndex - 1 + numOpts) % numOpts
				return m, nil
			case "right", "l", "tab":
				numOpts := 2
				if m.deleteModal.isShared {
					numOpts = 3
				}
				m.deleteModal.focusedIndex = (m.deleteModal.focusedIndex + 1) % numOpts
				return m, nil
			case "shift+tab":
				numOpts := 2
				if m.deleteModal.isShared {
					numOpts = 3
				}
				m.deleteModal.focusedIndex = (m.deleteModal.focusedIndex - 1 + numOpts) % numOpts
				return m, nil
			case "1":
				if m.deleteModal.isShared {
					return m.executeDeleteChoice(0)
				}
			case "2":
				if m.deleteModal.isShared {
					return m.executeDeleteChoice(1)
				}
			case "y":
				if !m.deleteModal.isShared {
					return m.executeDeleteChoice(0)
				}
			case "enter":
				return m.executeDeleteChoice(m.deleteModal.focusedIndex)
			}
			return m, nil
		}

		if m.renameModal.active {
			switch msg.String() {
			case "ctrl+c":
				m.cancelStream()
				return m, tea.Quit
			case "esc":
				m.renameModal = renameModalState{}
				return m, nil
			case "enter":
				newName := strings.TrimSpace(m.renameModal.input.Value())
				if newName == "" {
					m.renameModal.err = "Profile name cannot be empty"
					return m, nil
				}
				if newName == m.renameModal.targetProfile {
					m.renameModal.err = "New profile name must be different from current name"
					return m, nil
				}
				if strings.ContainsAny(newName, "/\\") || strings.Contains(newName, "..") {
					m.renameModal.err = "Profile name cannot contain slashes or '..'"
					return m, nil
				}
				if m.pm != nil {
					if err := m.pm.RenameProfile(m.renameModal.targetProfile, newName, m.cfg); err != nil {
						m.renameModal.err = err.Error()
						return m, nil
					}
				}
				if m.cache != nil {
					m.cache.Rename(m.renameModal.targetProfile, newName)
				}
				targetProfile := m.renameModal.targetProfile
				if m.reports != nil {
					for k, rep := range m.reports {
						parts := strings.SplitN(k, ":", 2)
						if len(parts) == 2 && parts[1] == targetProfile {
							delete(m.reports, k)
							rep.Profile = newName
							m.reports[fmt.Sprintf("%s:%s", parts[0], newName)] = rep
						} else if k == targetProfile {
							delete(m.reports, k)
							rep.Profile = newName
							m.reports[newName] = rep
						}
					}
				}
				m.renameModal = renameModalState{}
				m = m.refreshProfiles()
				for idx, p := range m.profiles {
					if p == newName {
						m.cursor = idx
						break
					}
				}
				return m, nil
			default:
				var cmd tea.Cmd
				m.renameModal.input, cmd = m.renameModal.input.Update(msg)
				return m, cmd
			}
		}

		if m.doctorDrawer.active {
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

		keys := m.keys
		if len(keys.Quit.Keys()) == 0 {
			keys = DefaultKeyMap()
		}

		switch {
		case key.Matches(msg, keys.Quit):
			m.cancelStream()
			return m, tea.Quit
		case key.Matches(msg, keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, keys.Down):
			if m.cursor < len(m.profiles)-1 {
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
			if len(m.profiles) > 0 {
				m.selected = m.profiles[m.cursor]
				m.outcome = ActionRun
				m.cancelStream()
				return m, tea.Quit
			}
		case key.Matches(msg, keys.Shell):
			if len(m.profiles) > 0 {
				m.selected = m.profiles[m.cursor]
				m.outcome = ActionShell
				m.cancelStream()
				return m, tea.Quit
			}
		case key.Matches(msg, keys.Login):
			m.outcome = ActionLogin
			m.cancelStream()
			return m, tea.Quit
		case key.Matches(msg, keys.Doctor):
			m = m.fetchDoctorDiagnostics()
			return m, nil
		case key.Matches(msg, keys.Rename):
			if len(m.profiles) > 0 && m.cursor >= 0 && m.cursor < len(m.profiles) {
				target := m.profiles[m.cursor]
				ti := textinput.New()
				ti.Placeholder = target
				ti.CharLimit = 64
				ti.Width = 30
				cmd := ti.Focus()
				m.renameModal = renameModalState{
					active:        true,
					targetProfile: target,
					input:         ti,
				}
				return m, cmd
			}
		case key.Matches(msg, keys.Delete):
			if len(m.profiles) > 0 && m.cursor >= 0 && m.cursor < len(m.profiles) {
				target := m.profiles[m.cursor]
				var agentsList []string
				if m.cfg != nil {
					agentsList = m.cfg.GetProfileAgents(target)
				}
				isShared := len(agentsList) > 1
				defaultFocus := 1
				if isShared {
					defaultFocus = 2
				}
				m.deleteModal = deleteModalState{
					active:        true,
					targetProfile: target,
					isShared:      isShared,
					agents:        agentsList,
					focusedIndex:  defaultFocus,
				}
				return m, nil
			}
		}
	default:
		if m.renameModal.active {
			var cmd tea.Cmd
			m.renameModal.input, cmd = m.renameModal.input.Update(msg)
			return m, cmd
		}
	}
	return m, nil
}

func (m Model) executeDeleteChoice(idx int) (tea.Model, tea.Cmd) {
	target := m.deleteModal.targetProfile
	isShared := m.deleteModal.isShared
	m.deleteModal = deleteModalState{}

	if !isShared {
		if idx == 1 {
			return m, nil
		}
		if m.pm != nil {
			_ = m.pm.DeleteProfile(target, m.cfg)
		}
	} else {
		if idx == 2 {
			return m, nil
		}
		if idx == 0 {
			if m.pm != nil {
				_, _ = m.pm.RemoveAgent(target, m.agent, m.cfg)
			}
		} else if idx == 1 {
			if m.pm != nil {
				_ = m.pm.DeleteProfile(target, m.cfg)
			}
		}
	}

	if m.cache != nil {
		m.cache.Delete(m.agent, target)
	}
	delete(m.reports, fmt.Sprintf("%s:%s", m.agent, target))
	delete(m.reports, target)

	m = m.refreshProfiles()
	if m.cursor >= len(m.profiles) {
		if len(m.profiles) > 0 {
			m.cursor = len(m.profiles) - 1
		} else {
			m.cursor = 0
		}
	}

	if len(m.profiles) > 0 {
		return m, m.triggerRefreshCmd()
	}
	return m, nil
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

func formatWindowsBadge(windows []usage.LimitWindow) string {
	if len(windows) == 0 {
		return ""
	}
	formatWindow := func(w *usage.LimitWindow) string {
		if w == nil {
			return ""
		}
		label := "limit"
		if w.IsHourly() {
			label = "5h"
		} else if w.IsWeekly() {
			label = "wk"
		} else if w.Name != "" {
			label = w.Name
		}
		if w.RemainingPct < 100 && w.ResetsIn > 0 {
			return fmt.Sprintf("%s: %d%% (%s)", label, w.RemainingPct, usage.FormatDuration(w.ResetsIn))
		}
		return fmt.Sprintf("%s: %d%%", label, w.RemainingPct)
	}

	var pw, ww *usage.LimitWindow
	for i := range windows {
		if pw == nil && windows[i].IsHourly() {
			pw = &windows[i]
		}
		if ww == nil && windows[i].IsWeekly() {
			ww = &windows[i]
		}
	}

	if pw != nil && ww != nil && (pw == ww || pw.Name == ww.Name) {
		if !pw.IsHourly() {
			pw = nil
		} else {
			ww = nil
		}
	}

	var parts []string
	if pw != nil {
		parts = append(parts, formatWindow(pw))
	}
	if ww != nil {
		parts = append(parts, formatWindow(ww))
	}
	if len(parts) == 0 {
		for i := range windows {
			parts = append(parts, formatWindow(&windows[i]))
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return "[" + strings.Join(parts, " | ") + "]"
}

func (m Model) View() string {
	var s strings.Builder
	s.WriteString(m.renderHeader())
	s.WriteString(m.renderTabBar())

	isNarrow := m.width > 0 && m.width < 85

	s.WriteString("  PROFILES:\n")
	if len(m.profiles) == 0 {
		s.WriteString(fmt.Sprintf("    (no profiles configured for %s - press 'l' to log in)\n", m.agent))
	} else {
		for i, p := range m.profiles {
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

	if m.doctorDrawer.active {
		s.WriteString(m.renderDoctorDrawer())
		return s.String()
	}

	// Bottom inspector section when a profile is highlighted
	if len(m.profiles) > 0 && m.cursor >= 0 && m.cursor < len(m.profiles) {
		curProfile := m.profiles[m.cursor]
		s.WriteString(m.renderInspector(curProfile))
	}

	refreshHint := HintKeyStyle.Render("[r]") + " " + HintLabelStyle.Render("Refresh Quota  ")
	if m.loading {
		spinnerChar := m.spinner.View()
		refreshHint = HintKeyStyle.Render("[r]") + " " + spinnerChar + " " + HintLabelStyle.Render("Refresh Quota  ")
	}

	s.WriteString("\n  " +
		HintKeyStyle.Render("[Enter]") + " " + HintLabelStyle.Render("Run  ") +
		HintKeyStyle.Render("[s]") + " " + HintLabelStyle.Render("Shell  ") +
		HintKeyStyle.Render("[l]") + " " + HintLabelStyle.Render("Login  ") +
		HintKeyStyle.Render("[Tab]") + " " + HintLabelStyle.Render("Switch Agent  ") +
		HintKeyStyle.Render("[d]") + " " + HintLabelStyle.Render("Doctor  ") +
		HintKeyStyle.Render("[m]") + " " + HintLabelStyle.Render("Rename  ") +
		HintKeyStyle.Render("[x]") + " " + HintLabelStyle.Render("Delete  ") +
		refreshHint +
		HintKeyStyle.Render("[q]") + " " + HintLabelStyle.Render("Quit") + "\n")

	return s.String()
}

func (m Model) renderDeleteModal() string {
	var b strings.Builder
	pName := m.deleteModal.targetProfile

	title := ModalTitleStyle.Render("[!] Confirm Deletion: " + pName)
	b.WriteString(title + "\n\n")

	if m.deleteModal.isShared {
		agentsStr := strings.Join(m.deleteModal.agents, ", ")
		b.WriteString(lipgloss.NewStyle().Foreground(TextSecondary).Render(
			fmt.Sprintf("Profile %q is shared across: %s", pName, agentsStr),
		) + "\n")
		b.WriteString(lipgloss.NewStyle().Foreground(TextMuted).Render(
			fmt.Sprintf("Do you want to unlink '%s' or delete the entire profile?", m.agent),
		) + "\n\n")

		btn0Style := ModalBtnInactiveStyle
		btn1Style := ModalBtnInactiveStyle
		btn2Style := ModalBtnInactiveStyle

		if m.deleteModal.focusedIndex == 0 {
			btn0Style = ModalBtnActiveStyle
		} else if m.deleteModal.focusedIndex == 1 {
			btn1Style = ModalBtnActiveStyle
		} else if m.deleteModal.focusedIndex == 2 {
			btn2Style = ModalBtnCancelActiveStyle
		}

		btn0 := btn0Style.Render(fmt.Sprintf("[1] Remove '%s' Only", m.agent))
		btn1 := btn1Style.Render("[2] Delete Entire Profile")
		btn2 := btn2Style.Render("[Cancel]")

		b.WriteString(fmt.Sprintf("  %s    %s    %s\n\n", btn0, btn1, btn2))
		b.WriteString(lipgloss.NewStyle().Foreground(TextMuted).Render(
			"  [←/→/Tab] Select  •  [Enter] Confirm  •  [Esc] Cancel",
		))
	} else {
		b.WriteString(lipgloss.NewStyle().Foreground(TextSecondary).Render(
			fmt.Sprintf("Are you sure you want to permanently delete profile %q?", pName),
		) + "\n")
		b.WriteString(lipgloss.NewStyle().Foreground(TextMuted).Render(
			"This will permanently delete all stored credentials and isolated state.",
		) + "\n\n")

		btn0Style := ModalBtnInactiveStyle
		btn1Style := ModalBtnInactiveStyle

		if m.deleteModal.focusedIndex == 0 {
			btn0Style = ModalBtnActiveStyle
		} else if m.deleteModal.focusedIndex == 1 {
			btn1Style = ModalBtnCancelActiveStyle
		}

		btn0 := btn0Style.Render("[ Delete Profile ]")
		btn1 := btn1Style.Render("[ Cancel ]")

		b.WriteString(fmt.Sprintf("      %s      %s\n\n", btn0, btn1))
		b.WriteString(lipgloss.NewStyle().Foreground(TextMuted).Render(
			"  [←/→/Tab] Select  •  [Enter/y] Confirm  •  [Esc] Cancel",
		))
	}

	box := ModalBoxStyle.Render(b.String())
	return "\n" + box + "\n"
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

func (m Model) renderRenameModal() string {
	var b strings.Builder
	pName := m.renameModal.targetProfile

	title := RenameModalTitleStyle.Render("✎ Rename Profile: " + pName)
	b.WriteString(title + "\n\n")

	b.WriteString(lipgloss.NewStyle().Foreground(TextSecondary).Render(
		"Enter new name for profile:",
	) + "\n\n")

	b.WriteString("  " + m.renameModal.input.View() + "\n\n")

	if m.renameModal.err != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(StatusRed).Bold(true).Render(
			"  ✕ "+m.renameModal.err,
		) + "\n\n")
	}

	b.WriteString(lipgloss.NewStyle().Foreground(TextMuted).Render(
		"  [Enter] Confirm  •  [Esc] Cancel",
	))

	box := RenameModalBoxStyle.Render(b.String())
	return "\n" + box + "\n"
}
