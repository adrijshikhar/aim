package tui

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/session"
	"github.com/aim-cli/aim/internal/session/providers/agy"
	claudesess "github.com/aim-cli/aim/internal/session/providers/claude"
	"github.com/aim-cli/aim/internal/session/providers/codex"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	terminateProcessFunc = terminateProcess
	copyToClipboardFunc  = copyToClipboard
	openDirectoryFunc    = openDirectory
)

func terminateProcess(pid int) error {
	return session.GracefulTerminate(pid, 2*time.Second)
}

func copyToClipboard(text string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux":
		if _, err := exec.LookPath("wl-copy"); err == nil {
			cmd = exec.Command("wl-copy")
		} else if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--clipboard", "--input")
		} else {
			return fmt.Errorf("no clipboard utility found")
		}
	case "windows":
		cmd = exec.Command("clip")
	default:
		return fmt.Errorf("unsupported platform for clipboard")
	}
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

func openDirectory(dir string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", dir)
	case "linux":
		cmd = exec.Command("xdg-open", dir)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", dir)
	default:
		return fmt.Errorf("unsupported platform for opening directory")
	}
	return cmd.Start()
}

type SessionsDrawer struct {
	active         bool
	loading        bool
	standalone     bool
	sessions       []session.Session
	index          *SessionsIndex
	cursor         int
	filterActive   bool
	filterInput    textinput.Model
	agentFilter    string
	profileFilter  string
	activeOnly     bool
	hidePreview    bool
	statusMessage  string
	killConfirmPID int
	fork           bool

	keys          SessionsDrawerKeyMap
	help          help.Model
	height        int
	profilesCount int
	pendingAction tea.Msg
}

type sessionsDrawerState = SessionsDrawer

var _ SubModel = SessionsDrawer{}

type SessionsCloseMsg struct{}
type SessionsQuitMsg struct{}
type SessionsResumeMsg struct {
	Session   *session.Session
	Action    ActionOutcome
	Fork      bool
	WithFlags bool
}

type sessionsLoadedMsg struct {
	sessions []session.Session
	err      error
}

func (d SessionsDrawer) Init() tea.Cmd {
	if d.loading {
		return d.fetchSessionsCmd()
	}
	return nil
}

func (d *SessionsDrawer) setSessions(sessions []session.Session) {
	d.sessions = sessions
	d.index = NewSessionsIndex(sessions)
	if d.cursor >= len(sessions) {
		if len(sessions) > 0 {
			d.cursor = len(sessions) - 1
		} else {
			d.cursor = 0
		}
	}
}

func (d SessionsDrawer) filteredSessions() []session.Session {
	idx := d.index
	if idx == nil {
		idx = NewSessionsIndex(d.sessions)
	}
	term := d.filterInput.Value()
	return idx.SearchWithProfile(term, d.activeOnly, d.profileFilter)
}

func (d *SessionsDrawer) fetchSessions() {
	mgr := session.NewManager()
	mgr.RegisterProvider(agy.NewProvider())
	mgr.RegisterProvider(codex.NewProvider())
	mgr.RegisterProvider(claudesess.NewProvider())

	sessions, err := mgr.ListSessions(context.Background(), d.agentFilter, d.profileFilter, false)
	if err != nil {
		sessions = []session.Session{}
	}
	d.setSessions(sessions)
}

func (d SessionsDrawer) fetchSessionsCmd() tea.Cmd {
	agentFilter := d.agentFilter
	profileFilter := d.profileFilter
	return func() tea.Msg {
		mgr := session.NewManager()
		mgr.RegisterProvider(agy.NewProvider())
		mgr.RegisterProvider(codex.NewProvider())
		mgr.RegisterProvider(claudesess.NewProvider())

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		sessions, err := mgr.ListSessions(ctx, agentFilter, profileFilter, false)
		return sessionsLoadedMsg{sessions: sessions, err: err}
	}
}

func (d SessionsDrawer) Update(msg tea.Msg) (SubModel, tea.Cmd) {
	switch msg := msg.(type) {
	case sessionsLoadedMsg:
		d.loading = false
		if msg.err == nil {
			d.setSessions(msg.sessions)
		}
		return d, nil

	case tea.WindowSizeMsg:
		d.height = msg.Height
		if msg.Width > 8 {
			d.help.Width = msg.Width - 8
		} else {
			d.help.Width = 100
		}
		return d, nil

	case tea.KeyMsg:
		km := d.keys
		if len(km.Quit.Keys()) == 0 {
			km = DefaultSessionsDrawerKeyMap()
		}

		if d.filterActive {
			filterKm := km.ForFilterMode()
			switch {
			case key.Matches(msg, filterKm.Quit):
				d.pendingAction = SessionsQuitMsg{}
				return d, func() tea.Msg { return SessionsQuitMsg{} }
			case key.Matches(msg, filterKm.TabFocus):
				d.filterActive = false
				d.filterInput.Blur()
				return d, nil
			case key.Matches(msg, filterKm.Up):
				if d.cursor > 0 {
					d.cursor--
				}
				d.statusMessage = ""
				d.killConfirmPID = 0
				return d, nil
			case key.Matches(msg, filterKm.Down):
				filtered := d.filteredSessions()
				if d.cursor < len(filtered)-1 {
					d.cursor++
				}
				d.statusMessage = ""
				d.killConfirmPID = 0
				return d, nil
			case key.Matches(msg, filterKm.PageUp):
				step := 8
				if d.cursor >= step {
					d.cursor -= step
				} else {
					d.cursor = 0
				}
				d.statusMessage = ""
				d.killConfirmPID = 0
				return d, nil
			case key.Matches(msg, filterKm.PageDown):
				filtered := d.filteredSessions()
				step := 8
				if d.cursor+step < len(filtered) {
					d.cursor += step
				} else if len(filtered) > 0 {
					d.cursor = len(filtered) - 1
				}
				d.statusMessage = ""
				d.killConfirmPID = 0
				return d, nil
			case key.Matches(msg, filterKm.Flags):
				filtered := d.filteredSessions()
				if len(filtered) > 0 && d.cursor >= 0 && d.cursor < len(filtered) {
					d.filterActive = false
					d.filterInput.Blur()
					target := filtered[d.cursor]
					d.pendingAction = SessionsResumeMsg{Session: &target, Action: ActionResumeExact, WithFlags: true}
					return d, func() tea.Msg { return d.pendingAction }
				}
				return d, nil
			case key.Matches(msg, filterKm.Enter):
				filtered := d.filteredSessions()
				if len(filtered) > 0 && d.cursor >= 0 && d.cursor < len(filtered) {
					d.filterActive = false
					d.filterInput.Blur()
					target := filtered[d.cursor]
					d.pendingAction = SessionsResumeMsg{Session: &target, Action: ActionResumeExact}
					return d, func() tea.Msg { return d.pendingAction }
				}
				d.filterActive = false
				d.filterInput.Blur()
				return d, nil
			default:
				oldVal := d.filterInput.Value()
				var cmd tea.Cmd
				d.filterInput, cmd = d.filterInput.Update(msg)
				if d.filterInput.Value() != oldVal {
					d.cursor = 0
				}
				return d, cmd
			}
		}

		switch {
		case key.Matches(msg, km.Quit):
			d.pendingAction = SessionsQuitMsg{}
			return d, func() tea.Msg { return SessionsQuitMsg{} }
		case key.Matches(msg, km.Close):
			if key.Matches(msg, km.ClearFilter) && d.filterInput.Value() != "" {
				d.filterInput.SetValue("")
				d.cursor = 0
				d.statusMessage = ""
				d.killConfirmPID = 0
				return d, nil
			}
			d.pendingAction = SessionsCloseMsg{}
			return d, func() tea.Msg { return SessionsCloseMsg{} }
		case key.Matches(msg, km.Filter):
			d.filterActive = true
			d.statusMessage = ""
			d.killConfirmPID = 0
			cmd := d.filterInput.Focus()
			return d, cmd
		case key.Matches(msg, km.TabFocus):
			if d.filterInput.Value() != "" {
				d.filterActive = true
				cmd := d.filterInput.Focus()
				return d, cmd
			}
			switch d.agentFilter {
			case "agy":
				d.agentFilter = "codex"
			case "codex":
				d.agentFilter = "claude"
			case "claude":
				d.agentFilter = ""
			default:
				d.agentFilter = "agy"
			}
			d.cursor = 0
			d.statusMessage = ""
			d.killConfirmPID = 0
			d.loading = true
			d.fetchSessions()
			return d, d.fetchSessionsCmd()
		case key.Matches(msg, km.Up):
			if d.cursor > 0 {
				d.cursor--
			}
			d.statusMessage = ""
			d.killConfirmPID = 0
			return d, nil
		case key.Matches(msg, km.Down):
			filtered := d.filteredSessions()
			if d.cursor < len(filtered)-1 {
				d.cursor++
			}
			d.statusMessage = ""
			d.killConfirmPID = 0
			return d, nil
		case key.Matches(msg, km.PageUp):
			step := 8
			if d.cursor >= step {
				d.cursor -= step
			} else {
				d.cursor = 0
			}
			d.statusMessage = ""
			d.killConfirmPID = 0
			return d, nil
		case key.Matches(msg, km.PageDown):
			filtered := d.filteredSessions()
			step := 8
			if d.cursor+step < len(filtered) {
				d.cursor += step
			} else if len(filtered) > 0 {
				d.cursor = len(filtered) - 1
			}
			d.statusMessage = ""
			d.killConfirmPID = 0
			return d, nil
		case key.Matches(msg, km.Enter):
			filtered := d.filteredSessions()
			if len(filtered) > 0 && d.cursor >= 0 && d.cursor < len(filtered) {
				target := filtered[d.cursor]
				d.pendingAction = SessionsResumeMsg{Session: &target, Action: ActionResumeExact}
				return d, func() tea.Msg { return d.pendingAction }
			}
		case key.Matches(msg, km.Flags):
			filtered := d.filteredSessions()
			if len(filtered) > 0 && d.cursor >= 0 && d.cursor < len(filtered) {
				target := filtered[d.cursor]
				d.pendingAction = SessionsResumeMsg{Session: &target, Action: ActionResumeExact, WithFlags: true}
				return d, func() tea.Msg { return d.pendingAction }
			}
		case key.Matches(msg, km.Catalyst):
			filtered := d.filteredSessions()
			if len(filtered) > 0 && d.cursor >= 0 && d.cursor < len(filtered) {
				target := filtered[d.cursor]
				d.pendingAction = SessionsResumeMsg{Session: &target, Action: ActionResumeCatalyst}
				return d, func() tea.Msg { return d.pendingAction }
			}
		case key.Matches(msg, km.Fork):
			filtered := d.filteredSessions()
			if len(filtered) > 0 && d.cursor >= 0 && d.cursor < len(filtered) {
				target := filtered[d.cursor]
				d.pendingAction = SessionsResumeMsg{Session: &target, Action: ActionResumeExact, Fork: true}
				return d, func() tea.Msg { return d.pendingAction }
			}
		case key.Matches(msg, km.AgentAll):
			d.agentFilter = ""
			d.cursor = 0
			d.statusMessage = ""
			d.killConfirmPID = 0
			d.loading = true
			d.fetchSessions()
			return d, d.fetchSessionsCmd()
		case key.Matches(msg, km.AgentAgy):
			d.agentFilter = "agy"
			d.cursor = 0
			d.statusMessage = ""
			d.killConfirmPID = 0
			d.loading = true
			d.fetchSessions()
			return d, d.fetchSessionsCmd()
		case key.Matches(msg, km.AgentCodex):
			d.agentFilter = "codex"
			d.cursor = 0
			d.statusMessage = ""
			d.killConfirmPID = 0
			d.loading = true
			d.fetchSessions()
			return d, d.fetchSessionsCmd()
		case key.Matches(msg, km.AgentClaude):
			d.agentFilter = "claude"
			d.cursor = 0
			d.statusMessage = ""
			d.killConfirmPID = 0
			d.loading = true
			d.fetchSessions()
			return d, d.fetchSessionsCmd()
		case key.Matches(msg, km.ToggleActive):
			d.activeOnly = !d.activeOnly
			d.cursor = 0
			d.statusMessage = ""
			d.killConfirmPID = 0
			return d, nil
		case key.Matches(msg, km.TogglePreview):
			d.hidePreview = !d.hidePreview
			return d, nil
		case key.Matches(msg, km.Kill):
			filtered := d.filteredSessions()
			if len(filtered) > 0 && d.cursor >= 0 && d.cursor < len(filtered) {
				target := filtered[d.cursor]
				if target.Status != session.StatusActive || target.PID <= 0 {
					d.statusMessage = fmt.Sprintf("Session %s is idle (no active process to terminate)", target.ShortID)
					d.killConfirmPID = 0
					return d, nil
				}
				if d.killConfirmPID != target.PID {
					d.killConfirmPID = target.PID
					d.statusMessage = fmt.Sprintf("Press 'x' again to terminate PID %d (%s)", target.PID, target.ShortID)
					return d, nil
				}
				// Confirmed kill
				err := terminateProcessFunc(target.PID)
				if err != nil {
					d.statusMessage = fmt.Sprintf("Failed to terminate PID %d: %v", target.PID, err)
				} else {
					d.statusMessage = fmt.Sprintf("Terminated PID %d (%s)", target.PID, target.ShortID)
				}
				d.killConfirmPID = 0
				d.loading = true
				d.fetchSessions()
				return d, d.fetchSessionsCmd()
			}
		case key.Matches(msg, km.Copy):
			filtered := d.filteredSessions()
			if len(filtered) > 0 && d.cursor >= 0 && d.cursor < len(filtered) {
				target := filtered[d.cursor]
				toCopy := target.Cwd
				if toCopy == "" {
					toCopy = target.ID
				}
				err := copyToClipboardFunc(toCopy)
				if err == nil {
					d.statusMessage = fmt.Sprintf("Copied to clipboard: %s", toCopy)
				} else {
					d.statusMessage = fmt.Sprintf("Copy failed: %v", err)
				}
				return d, nil
			}
		case key.Matches(msg, km.OpenDir):
			filtered := d.filteredSessions()
			if len(filtered) > 0 && d.cursor >= 0 && d.cursor < len(filtered) {
				target := filtered[d.cursor]
				if target.Cwd == "" {
					d.statusMessage = "No workspace directory recorded for this session"
					return d, nil
				}
				err := openDirectoryFunc(target.Cwd)
				if err == nil {
					d.statusMessage = fmt.Sprintf("Opened workspace: %s", target.Cwd)
				} else {
					d.statusMessage = fmt.Sprintf("Failed to open directory: %v", err)
				}
				return d, nil
			}
		}
	}

	return d, nil
}

func (d SessionsDrawer) View() string {
	var b strings.Builder

	// Top Bar
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(AccentBlue)
	filterTabStyle := func(active bool) lipgloss.Style {
		if active {
			return lipgloss.NewStyle().Bold(true).Foreground(AccentCyan).Background(BgTabActive).Padding(0, 1)
		}
		return lipgloss.NewStyle().Foreground(TextMuted).Padding(0, 1)
	}

	allActive := d.agentFilter == ""
	agyActive := d.agentFilter == "agy"
	codexActive := d.agentFilter == "codex"
	claudeActive := d.agentFilter == "claude"

	activeOnlyBadge := ""
	if d.activeOnly {
		activeOnlyBadge = " " + lipgloss.NewStyle().Bold(true).Foreground(StatusGreen).Background(BgTabActive).Padding(0, 1).Render("[a] Active: ON")
	}

	previewBadge := ""
	if d.hidePreview {
		previewBadge = " " + lipgloss.NewStyle().Foreground(TextMuted).Padding(0, 1).Render("[p] Preview: OFF")
	}

	topBar := fmt.Sprintf("%s  %s %s %s %s%s%s",
		titleStyle.Render("Sessions Explorer"),
		filterTabStyle(allActive).Render("[0] All"),
		filterTabStyle(agyActive).Render("[1] Antigravity"),
		filterTabStyle(codexActive).Render("[2] Codex"),
		filterTabStyle(claudeActive).Render("[3] Claude"),
		activeOnlyBadge,
		previewBadge,
	)
	cmdBar := lipgloss.NewStyle().Foreground(TextMuted).Render(
		"  [0-3/Tab] Switch Agent • [a] Active Only • [p] Toggle Preview • [/] Filter • [Esc] Close",
	)
	b.WriteString(topBar + "\n")
	b.WriteString(cmdBar + "\n\n")

	// Status Message Notification Banner
	if d.statusMessage != "" {
		msgStyle := lipgloss.NewStyle().Bold(true).Foreground(StatusYellow)
		if strings.HasPrefix(d.statusMessage, "Terminated") ||
			strings.HasPrefix(d.statusMessage, "Copied") ||
			strings.HasPrefix(d.statusMessage, "Opened") {
			msgStyle = lipgloss.NewStyle().Bold(true).Foreground(StatusGreen)
		}
		b.WriteString("  " + msgStyle.Render("▶ "+d.statusMessage) + "\n\n")
	}

	if d.filterActive || d.filterInput.Value() != "" {
		b.WriteString(d.filterInput.View() + "\n\n")
	}

	// Columns header
	colProfile := lipgloss.NewStyle().Bold(true).Foreground(AccentBlue).Width(8).Render("PROFILE")
	colAgent := lipgloss.NewStyle().Bold(true).Foreground(AccentBlue).Width(7).Render("AGENT")
	colID := lipgloss.NewStyle().Bold(true).Foreground(AccentBlue).Width(10).Render("SESSION ID")
	colDir := lipgloss.NewStyle().Bold(true).Foreground(AccentBlue).Width(18).Render("DIR")
	colTitle := lipgloss.NewStyle().Bold(true).Foreground(AccentBlue).Width(34).Render("TITLE")
	colActive := lipgloss.NewStyle().Bold(true).Foreground(AccentBlue).Width(14).Render("STATUS / TIME")

	b.WriteString(fmt.Sprintf("  %s %s %s %s %s %s\n", colProfile, colAgent, colID, colDir, colTitle, colActive))

	filtered := d.filteredSessions()
	if len(filtered) == 0 {
		b.WriteString("\n  " + lipgloss.NewStyle().Foreground(TextMuted).Render("(no conversation sessions found matching filter)") + "\n\n")
	} else {
		overhead := 18
		if d.profilesCount > 0 {
			overhead += d.profilesCount
		}
		maxVisible := 8
		if d.hidePreview {
			maxVisible = 14
			if d.height > 0 && d.height-overhead < maxVisible {
				maxVisible = d.height - overhead
			}
			if maxVisible < 5 {
				maxVisible = 5
			}
			if maxVisible > 16 {
				maxVisible = 16
			}
		} else {
			if d.height > 0 && d.height-overhead < maxVisible {
				maxVisible = d.height - overhead
			}
			if maxVisible < 4 {
				maxVisible = 4
			}
			if maxVisible > 8 {
				maxVisible = 8
			}
		}

		start := 0
		if d.cursor >= maxVisible {
			start = d.cursor - maxVisible + 1
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
			cursorStr := "  "
			rowStyle := NormalRowStyle
			if i == d.cursor {
				cursorStr = "> "
				rowStyle = SelectedRowStyle
			}

			dirStr := session.FormatDir(s.Cwd, 18)
			titleStr := truncateString(s.Title, 34)
			if titleStr == "" {
				titleStr = "(untitled)"
			}

			lastActiveStr := formatRelativeTime(s.LastActiveAt)
			if s.Status == session.StatusActive {
				if s.PID > 0 {
					lastActiveStr = lipgloss.NewStyle().Bold(true).Foreground(StatusGreen).Render(fmt.Sprintf("● PID %d", s.PID))
				} else {
					lastActiveStr = lipgloss.NewStyle().Bold(true).Foreground(StatusGreen).Render("● ACTIVE")
				}
			} else {
				lastActiveStr = lipgloss.NewStyle().Foreground(TextMuted).Render(lastActiveStr)
			}

			cProfile := lipgloss.NewStyle().Width(8).Render(s.Profile)
			cAgent := lipgloss.NewStyle().Width(7).Render(s.Agent)
			cID := lipgloss.NewStyle().Width(10).Render(s.ShortID)
			cDir := lipgloss.NewStyle().Width(18).Render(dirStr)
			cTitle := lipgloss.NewStyle().Width(34).Render(titleStr)
			cActive := lipgloss.NewStyle().Width(14).Render(lastActiveStr)

			rowContent := fmt.Sprintf("%s%s %s %s %s %s %s", cursorStr, cProfile, cAgent, cID, cDir, cTitle, cActive)
			b.WriteString(rowStyle.Render(rowContent) + "\n")
		}

		if len(filtered) > maxVisible {
			scrollInfo := fmt.Sprintf("  (showing %d-%d of %d sessions)", start+1, end, len(filtered))
			b.WriteString(lipgloss.NewStyle().Foreground(TextMuted).Render(scrollInfo) + "\n")
		}

		// Dedicated Live Preview Box for the currently highlighted session (when preview not toggled off)
		if !d.hidePreview && d.cursor >= 0 && d.cursor < len(filtered) {
			sel := filtered[d.cursor]
			previewText := sel.Summary
			if previewText == "" {
				previewText = sel.Title
			}
			previewText = strings.TrimSpace(previewText)
			if previewText == "" {
				previewText = "(no summary recorded for this session)"
			}

			displayLines := wrapText(previewText, 92, 4)
			if len(displayLines) == 0 {
				displayLines = []string{"(no summary recorded for this session)"}
			}

			previewCard := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(AccentCyan).
				Padding(0, 1).
				MarginTop(1).
				Width(96)

			var pb strings.Builder
			statusLabel := lipgloss.NewStyle().Foreground(TextMuted).Render("IDLE")
			if sel.Status == session.StatusActive {
				if sel.PID > 0 {
					statusLabel = lipgloss.NewStyle().Bold(true).Foreground(StatusGreen).Render(fmt.Sprintf("ACTIVE (PID %d)", sel.PID))
				} else {
					statusLabel = lipgloss.NewStyle().Bold(true).Foreground(StatusGreen).Render("ACTIVE")
				}
			}

			previewHeader := fmt.Sprintf("%s %s  %s %s  %s %s  %s %s",
				lipgloss.NewStyle().Bold(true).Foreground(AccentCyan).Render("Summary:"),
				lipgloss.NewStyle().Bold(true).Foreground(TextBright).Render(sel.ShortID),
				lipgloss.NewStyle().Foreground(TextMuted).Render("Agent:"),
				lipgloss.NewStyle().Foreground(AccentBlue).Render(sel.Agent),
				lipgloss.NewStyle().Foreground(TextMuted).Render("Profile:"),
				lipgloss.NewStyle().Foreground(StatusYellow).Render(sel.Profile),
				lipgloss.NewStyle().Foreground(TextMuted).Render("Status:"),
				statusLabel,
			)
			pb.WriteString(previewHeader + "\n")

			timeInfo := fmt.Sprintf("%s %s  %s %s",
				lipgloss.NewStyle().Foreground(TextMuted).Render("Started:"),
				lipgloss.NewStyle().Foreground(TextBright).Render(formatTimeOrRelative(sel.StartedAt, sel.LastActiveAt)),
				lipgloss.NewStyle().Foreground(TextMuted).Render("Last Active:"),
				lipgloss.NewStyle().Foreground(TextBright).Render(formatRelativeTime(sel.LastActiveAt)),
			)
			pb.WriteString(timeInfo + "\n")

			if sel.Cwd != "" {
				displayCwd := sel.Cwd
				if home := config.RealHomeDir(); strings.HasPrefix(displayCwd, home) {
					displayCwd = "~" + displayCwd[len(home):]
				}
				pb.WriteString(fmt.Sprintf("%s %s\n",
					lipgloss.NewStyle().Bold(true).Foreground(TextMuted).Render("Workspace:"),
					lipgloss.NewStyle().Foreground(AccentCyan).Render(truncateString(displayCwd, 80)),
				))
			}

			hasStructured := false
			if sel.Goal != "" {
				hasStructured = true
				goalLines := wrapText(sel.Goal, 80, 2)
				for idx, gl := range goalLines {
					label := "Goal:     "
					if idx > 0 {
						label = "          "
					}
					pb.WriteString(fmt.Sprintf("%s %s\n",
						lipgloss.NewStyle().Bold(true).Foreground(TextMuted).Render(label),
						lipgloss.NewStyle().Foreground(TextBright).Render(gl),
					))
				}
			}
			if sel.Progress != "" {
				hasStructured = true
				pb.WriteString(fmt.Sprintf("%s %s\n",
					lipgloss.NewStyle().Bold(true).Foreground(TextMuted).Render("Progress: "),
					lipgloss.NewStyle().Foreground(StatusGreen).Render(truncateString(sel.Progress, 80)),
				))
			}
			if sel.Recent != "" && sel.Recent != sel.Goal {
				hasStructured = true
				recentLines := wrapText(sel.Recent, 80, 2)
				for idx, rl := range recentLines {
					label := "Latest:   "
					if idx > 0 {
						label = "          "
					}
					pb.WriteString(fmt.Sprintf("%s %s\n",
						lipgloss.NewStyle().Bold(true).Foreground(TextMuted).Render(label),
						lipgloss.NewStyle().Foreground(AccentCyan).Render(rl),
					))
				}
			}

			if !hasStructured {
				for _, dl := range displayLines {
					pb.WriteString(lipgloss.NewStyle().Foreground(TextPrimary).Render(dl) + "\n")
				}
			}

			b.WriteString(previewCard.Render(pb.String()) + "\n")
		}
	}

	km := d.keys
	if len(km.Quit.Keys()) == 0 {
		km = DefaultSessionsDrawerKeyMap()
	}
	h := d.help
	if h.Width == 0 {
		h = NewThemedHelp()
	}
	var footerHelp string
	if d.filterActive {
		footerHelp = h.ShortHelpView(km.ShortHelpFilter())
	} else if d.filterInput.Value() != "" {
		footerHelp = h.ShortHelpView(km.ShortHelpQuery())
	} else {
		footerHelp = h.ShortHelpView(km.ShortHelp())
	}
	b.WriteString("\n  " + footerHelp)

	box := SessionsDrawerStyle.Render(b.String())
	return "\n" + box + "\n"
}

func (m Model) IsSessionsDrawerActive() bool {
	return m.sessionsDrawer.active
}

func (m Model) IsSessionsLoading() bool {
	return m.sessionsDrawer.loading
}

func (m Model) SessionsDrawerList() []session.Session {
	return m.sessionsDrawer.sessions
}

func (m *Model) setSessions(sessions []session.Session) {
	m.sessionsDrawer.setSessions(sessions)
}

func (m *Model) SetSessionsForTest(sessions []session.Session) {
	m.setSessions(sessions)
}

func (m Model) IsForkResume() bool {
	return m.sessionsDrawer.fork
}

func (m Model) fetchSessionsCmd() tea.Cmd {
	return m.sessionsDrawer.fetchSessionsCmd()
}

func (m Model) openSessionsDrawer() (Model, tea.Cmd) {
	m = m.openSessionsDrawerConfig(m.agent, "", false, false)
	m.sessionsDrawer.loading = true
	return m, m.fetchSessionsCmd()
}

func (m Model) openSessionsDrawerConfig(initialAgent, profileFilter string, activeOnly, standalone bool) Model {
	ti := textinput.New()
	ti.Placeholder = "Filter sessions by title, id, profile, or workspace..."
	ti.CharLimit = 64
	ti.Prompt = "Filter: "
	ti.PromptStyle = lipgloss.NewStyle().Bold(true).Foreground(AccentCyan)

	m.sessionsDrawer = SessionsDrawer{
		active:        true,
		standalone:    standalone,
		filterInput:   ti,
		agentFilter:   initialAgent,
		profileFilter: profileFilter,
		activeOnly:    activeOnly,
		cursor:        0,
		keys:          m.keys.SessionsDrawer,
		help:          m.help,
		height:        m.height,
		profilesCount: len(m.filteredProfiles()),
	}
	m.sessionsDrawer.fetchSessions()
	return m
}

func (m Model) fetchSessions() Model {
	m.sessionsDrawer.fetchSessions()
	return m
}

func (m Model) filteredSessions() []session.Session {
	return m.sessionsDrawer.filteredSessions()
}

func (m Model) updateSessionsDrawer(msg tea.KeyMsg) (Model, tea.Cmd) {
	m.sessionsDrawer.keys = m.keys.SessionsDrawer
	m.sessionsDrawer.help = m.help
	m.sessionsDrawer.height = m.height
	m.sessionsDrawer.profilesCount = len(m.filteredProfiles())
	sub, cmd := m.sessionsDrawer.Update(msg)
	d := sub.(SessionsDrawer)
	action := d.pendingAction
	d.pendingAction = nil
	m.sessionsDrawer = d

	if action != nil {
		switch act := action.(type) {
		case SessionsQuitMsg:
			m.cancelStream()
			return m, tea.Quit
		case SessionsCloseMsg:
			if m.sessionsDrawer.standalone {
				m.cancelStream()
				return m, tea.Quit
			}
			m.sessionsDrawer = SessionsDrawer{}
			return m, nil
		case SessionsResumeMsg:
			if act.WithFlags {
				return m.openResumeModalWithFlags(act.Session, act.Action, act.Fork, true)
			}
			return m.openResumeModal(act.Session, act.Action, act.Fork)
		}
	}

	return m, cmd
}

func (m Model) renderSessionsDrawer() string {
	d := m.sessionsDrawer
	d.height = m.height
	d.keys = m.keys.SessionsDrawer
	d.help = m.help
	d.profilesCount = len(m.filteredProfiles())
	return d.View()
}

func truncateString(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	r := []rune(s)
	if len(r) > maxLen {
		if maxLen > 3 {
			return string(r[:maxLen-3]) + "..."
		}
		return string(r[:maxLen])
	}
	return s
}

func formatTimeOrRelative(started, lastActive time.Time) string {
	t := started
	if t.IsZero() {
		t = lastActive
	}
	if t.IsZero() {
		return "-"
	}
	if time.Since(t) < 24*time.Hour {
		return t.Format("3:04PM")
	}
	return t.Format("Jan 02 3:04PM")
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

func wrapText(text string, maxWidth int, maxLines int) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	paragraphs := strings.Split(text, "\n")
	var lines []string

	for _, p := range paragraphs {
		words := strings.Fields(p)
		if len(words) == 0 {
			continue
		}

		var current strings.Builder
		for i, w := range words {
			if len(w) > maxWidth {
				w = w[:maxWidth-3] + "..."
			}

			if len(lines) == maxLines-1 {
				// We are on the final line allowed.
				// Fit as many remaining words as possible.
				if current.Len() == 0 {
					current.WriteString(w)
				} else if current.Len()+1+len(w) <= maxWidth-4 {
					current.WriteString(" " + w)
				} else {
					// Word doesn't fit or there are more words remaining
					current.WriteString("...")
					lines = append(lines, current.String())
					return lines
				}
				// If this was the last word of all paragraphs, append and return
				if i == len(words)-1 && p == paragraphs[len(paragraphs)-1] {
					lines = append(lines, current.String())
					return lines
				}
				continue
			}

			if current.Len() == 0 {
				current.WriteString(w)
			} else if current.Len()+1+len(w) <= maxWidth {
				current.WriteString(" " + w)
			} else {
				lines = append(lines, current.String())
				current.Reset()
				current.WriteString(w)
			}
		}

		if current.Len() > 0 && len(lines) < maxLines {
			lines = append(lines, current.String())
			current.Reset()
		}

		if len(lines) >= maxLines {
			break
		}
	}

	return lines
}
