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

type sessionsDrawerState struct {
	active         bool
	standalone     bool
	sessions       []session.Session
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
	return m.openSessionsDrawerConfig(m.agent, "", false, false), nil
}

func (m Model) openSessionsDrawerConfig(initialAgent, profileFilter string, activeOnly, standalone bool) Model {
	ti := textinput.New()
	ti.Placeholder = "Filter sessions by title, id, profile, or workspace..."
	ti.CharLimit = 64
	ti.Prompt = "Filter: "
	ti.PromptStyle = lipgloss.NewStyle().Bold(true).Foreground(AccentCyan)

	m.sessionsDrawer = sessionsDrawerState{
		active:        true,
		standalone:    standalone,
		filterInput:   ti,
		agentFilter:   initialAgent,
		profileFilter: profileFilter,
		activeOnly:    activeOnly,
		cursor:        0,
	}
	m = m.fetchSessions()
	return m
}

func (m Model) fetchSessions() Model {
	mgr := session.NewManager()
	mgr.RegisterProvider(agy.NewProvider())
	mgr.RegisterProvider(codex.NewProvider())
	mgr.RegisterProvider(claudesess.NewProvider())

	sessions, err := mgr.ListSessions(context.Background(), m.sessionsDrawer.agentFilter, m.sessionsDrawer.profileFilter, false)
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
	var res []session.Session
	for _, s := range m.sessionsDrawer.sessions {
		if m.sessionsDrawer.activeOnly && s.Status != session.StatusActive {
			continue
		}
		if m.sessionsDrawer.profileFilter != "" && !strings.EqualFold(s.Profile, m.sessionsDrawer.profileFilter) {
			continue
		}
		if term != "" {
			if !strings.Contains(strings.ToLower(s.ID), term) &&
				!strings.Contains(strings.ToLower(s.ShortID), term) &&
				!strings.Contains(strings.ToLower(s.Title), term) &&
				!strings.Contains(strings.ToLower(s.Summary), term) &&
				!strings.Contains(strings.ToLower(s.Goal), term) &&
				!strings.Contains(strings.ToLower(s.Progress), term) &&
				!strings.Contains(strings.ToLower(s.Recent), term) &&
				!strings.Contains(strings.ToLower(s.Cwd), term) &&
				!strings.Contains(strings.ToLower(s.Profile), term) &&
				!strings.Contains(strings.ToLower(s.Agent), term) {
				continue
			}
		}
		res = append(res, s)
	}
	return res
}

func (m Model) updateSessionsDrawer(msg tea.KeyMsg) (Model, tea.Cmd) {
	km := m.keys.SessionsDrawer

	if m.sessionsDrawer.filterActive {
		filterKm := km.ForFilterMode()
		switch {
		case key.Matches(msg, filterKm.Quit):
			m.cancelStream()
			return m, tea.Quit
		case key.Matches(msg, filterKm.TabFocus):
			m.sessionsDrawer.filterActive = false
			m.sessionsDrawer.filterInput.Blur()
			return m, nil
		case key.Matches(msg, filterKm.Up):
			if m.sessionsDrawer.cursor > 0 {
				m.sessionsDrawer.cursor--
			}
			m.sessionsDrawer.statusMessage = ""
			m.sessionsDrawer.killConfirmPID = 0
			return m, nil
		case key.Matches(msg, filterKm.Down):
			filtered := m.filteredSessions()
			if m.sessionsDrawer.cursor < len(filtered)-1 {
				m.sessionsDrawer.cursor++
			}
			m.sessionsDrawer.statusMessage = ""
			m.sessionsDrawer.killConfirmPID = 0
			return m, nil
		case key.Matches(msg, filterKm.PageUp):
			step := 8
			if m.sessionsDrawer.cursor >= step {
				m.sessionsDrawer.cursor -= step
			} else {
				m.sessionsDrawer.cursor = 0
			}
			m.sessionsDrawer.statusMessage = ""
			m.sessionsDrawer.killConfirmPID = 0
			return m, nil
		case key.Matches(msg, filterKm.PageDown):
			filtered := m.filteredSessions()
			step := 8
			if m.sessionsDrawer.cursor+step < len(filtered) {
				m.sessionsDrawer.cursor += step
			} else if len(filtered) > 0 {
				m.sessionsDrawer.cursor = len(filtered) - 1
			}
			m.sessionsDrawer.statusMessage = ""
			m.sessionsDrawer.killConfirmPID = 0
			return m, nil
		case key.Matches(msg, filterKm.Flags):
			filtered := m.filteredSessions()
			if len(filtered) > 0 && m.sessionsDrawer.cursor >= 0 && m.sessionsDrawer.cursor < len(filtered) {
				m.sessionsDrawer.filterActive = false
				m.sessionsDrawer.filterInput.Blur()
				target := filtered[m.sessionsDrawer.cursor]
				return m.openResumeModalWithFlags(&target, ActionResumeExact, false, true)
			}
			return m, nil
		case key.Matches(msg, filterKm.Enter):
			filtered := m.filteredSessions()
			if len(filtered) > 0 && m.sessionsDrawer.cursor >= 0 && m.sessionsDrawer.cursor < len(filtered) {
				m.sessionsDrawer.filterActive = false
				m.sessionsDrawer.filterInput.Blur()
				target := filtered[m.sessionsDrawer.cursor]
				return m.openResumeModal(&target, ActionResumeExact, false)
			}
			m.sessionsDrawer.filterActive = false
			m.sessionsDrawer.filterInput.Blur()
			return m, nil
		default:
			oldVal := m.sessionsDrawer.filterInput.Value()
			var cmd tea.Cmd
			m.sessionsDrawer.filterInput, cmd = m.sessionsDrawer.filterInput.Update(msg)
			if m.sessionsDrawer.filterInput.Value() != oldVal {
				m.sessionsDrawer.cursor = 0
			}
			return m, cmd
		}
	}

	switch {
	case key.Matches(msg, km.Quit):
		m.cancelStream()
		return m, tea.Quit
	case key.Matches(msg, km.Close):
		if key.Matches(msg, km.ClearFilter) && m.sessionsDrawer.filterInput.Value() != "" {
			m.sessionsDrawer.filterInput.SetValue("")
			m.sessionsDrawer.cursor = 0
			m.sessionsDrawer.statusMessage = ""
			m.sessionsDrawer.killConfirmPID = 0
			return m, nil
		}
		if m.sessionsDrawer.standalone {
			m.cancelStream()
			return m, tea.Quit
		}
		m.sessionsDrawer = sessionsDrawerState{}
		return m, nil
	case key.Matches(msg, km.Filter):
		m.sessionsDrawer.filterActive = true
		m.sessionsDrawer.statusMessage = ""
		m.sessionsDrawer.killConfirmPID = 0
		cmd := m.sessionsDrawer.filterInput.Focus()
		return m, cmd
	case key.Matches(msg, km.TabFocus):
		if m.sessionsDrawer.filterInput.Value() != "" {
			m.sessionsDrawer.filterActive = true
			cmd := m.sessionsDrawer.filterInput.Focus()
			return m, cmd
		}
		switch m.sessionsDrawer.agentFilter {
		case "agy":
			m.sessionsDrawer.agentFilter = "codex"
		case "codex":
			m.sessionsDrawer.agentFilter = "claude"
		case "claude":
			m.sessionsDrawer.agentFilter = ""
		default:
			m.sessionsDrawer.agentFilter = "agy"
		}
		m.sessionsDrawer.cursor = 0
		m.sessionsDrawer.statusMessage = ""
		m.sessionsDrawer.killConfirmPID = 0
		m = m.fetchSessions()
		return m, nil
	case key.Matches(msg, km.Up):
		if m.sessionsDrawer.cursor > 0 {
			m.sessionsDrawer.cursor--
		}
		m.sessionsDrawer.statusMessage = ""
		m.sessionsDrawer.killConfirmPID = 0
		return m, nil
	case key.Matches(msg, km.Down):
		filtered := m.filteredSessions()
		if m.sessionsDrawer.cursor < len(filtered)-1 {
			m.sessionsDrawer.cursor++
		}
		m.sessionsDrawer.statusMessage = ""
		m.sessionsDrawer.killConfirmPID = 0
		return m, nil
	case key.Matches(msg, km.PageUp):
		step := 8
		if m.sessionsDrawer.cursor >= step {
			m.sessionsDrawer.cursor -= step
		} else {
			m.sessionsDrawer.cursor = 0
		}
		m.sessionsDrawer.statusMessage = ""
		m.sessionsDrawer.killConfirmPID = 0
		return m, nil
	case key.Matches(msg, km.PageDown):
		filtered := m.filteredSessions()
		step := 8
		if m.sessionsDrawer.cursor+step < len(filtered) {
			m.sessionsDrawer.cursor += step
		} else if len(filtered) > 0 {
			m.sessionsDrawer.cursor = len(filtered) - 1
		}
		m.sessionsDrawer.statusMessage = ""
		m.sessionsDrawer.killConfirmPID = 0
		return m, nil
	case key.Matches(msg, km.Enter):
		filtered := m.filteredSessions()
		if len(filtered) > 0 && m.sessionsDrawer.cursor >= 0 && m.sessionsDrawer.cursor < len(filtered) {
			target := filtered[m.sessionsDrawer.cursor]
			return m.openResumeModal(&target, ActionResumeExact, false)
		}
	case key.Matches(msg, km.Flags):
		filtered := m.filteredSessions()
		if len(filtered) > 0 && m.sessionsDrawer.cursor >= 0 && m.sessionsDrawer.cursor < len(filtered) {
			target := filtered[m.sessionsDrawer.cursor]
			return m.openResumeModalWithFlags(&target, ActionResumeExact, false, true)
		}
	case key.Matches(msg, km.Catalyst):
		filtered := m.filteredSessions()
		if len(filtered) > 0 && m.sessionsDrawer.cursor >= 0 && m.sessionsDrawer.cursor < len(filtered) {
			target := filtered[m.sessionsDrawer.cursor]
			return m.openResumeModal(&target, ActionResumeCatalyst, false)
		}
	case key.Matches(msg, km.Fork):
		filtered := m.filteredSessions()
		if len(filtered) > 0 && m.sessionsDrawer.cursor >= 0 && m.sessionsDrawer.cursor < len(filtered) {
			target := filtered[m.sessionsDrawer.cursor]
			return m.openResumeModal(&target, ActionResumeExact, true)
		}
	case key.Matches(msg, km.AgentAll):
		m.sessionsDrawer.agentFilter = ""
		m.sessionsDrawer.cursor = 0
		m.sessionsDrawer.statusMessage = ""
		m.sessionsDrawer.killConfirmPID = 0
		m = m.fetchSessions()
		return m, nil
	case key.Matches(msg, km.AgentAgy):
		m.sessionsDrawer.agentFilter = "agy"
		m.sessionsDrawer.cursor = 0
		m.sessionsDrawer.statusMessage = ""
		m.sessionsDrawer.killConfirmPID = 0
		m = m.fetchSessions()
		return m, nil
	case key.Matches(msg, km.AgentCodex):
		m.sessionsDrawer.agentFilter = "codex"
		m.sessionsDrawer.cursor = 0
		m.sessionsDrawer.statusMessage = ""
		m.sessionsDrawer.killConfirmPID = 0
		m = m.fetchSessions()
		return m, nil
	case key.Matches(msg, km.AgentClaude):
		m.sessionsDrawer.agentFilter = "claude"
		m.sessionsDrawer.cursor = 0
		m.sessionsDrawer.statusMessage = ""
		m.sessionsDrawer.killConfirmPID = 0
		m = m.fetchSessions()
		return m, nil
	case key.Matches(msg, km.ToggleActive):
		m.sessionsDrawer.activeOnly = !m.sessionsDrawer.activeOnly
		m.sessionsDrawer.cursor = 0
		m.sessionsDrawer.statusMessage = ""
		m.sessionsDrawer.killConfirmPID = 0
		return m, nil
	case key.Matches(msg, km.TogglePreview):
		m.sessionsDrawer.hidePreview = !m.sessionsDrawer.hidePreview
		return m, nil
	case key.Matches(msg, km.Kill):
		filtered := m.filteredSessions()
		if len(filtered) > 0 && m.sessionsDrawer.cursor >= 0 && m.sessionsDrawer.cursor < len(filtered) {
			target := filtered[m.sessionsDrawer.cursor]
			if target.Status != session.StatusActive || target.PID <= 0 {
				m.sessionsDrawer.statusMessage = fmt.Sprintf("Session %s is idle (no active process to terminate)", target.ShortID)
				m.sessionsDrawer.killConfirmPID = 0
				return m, nil
			}
			if m.sessionsDrawer.killConfirmPID != target.PID {
				m.sessionsDrawer.killConfirmPID = target.PID
				m.sessionsDrawer.statusMessage = fmt.Sprintf("Press 'x' again to terminate PID %d (%s)", target.PID, target.ShortID)
				return m, nil
			}
			// Confirmed kill
			err := terminateProcessFunc(target.PID)
			if err != nil {
				m.sessionsDrawer.statusMessage = fmt.Sprintf("Failed to terminate PID %d: %v", target.PID, err)
			} else {
				m.sessionsDrawer.statusMessage = fmt.Sprintf("Terminated PID %d (%s)", target.PID, target.ShortID)
			}
			m.sessionsDrawer.killConfirmPID = 0
			m = m.fetchSessions()
			return m, nil
		}
	case key.Matches(msg, km.Copy):
		filtered := m.filteredSessions()
		if len(filtered) > 0 && m.sessionsDrawer.cursor >= 0 && m.sessionsDrawer.cursor < len(filtered) {
			target := filtered[m.sessionsDrawer.cursor]
			toCopy := target.Cwd
			if toCopy == "" {
				toCopy = target.ID
			}
			err := copyToClipboardFunc(toCopy)
			if err == nil {
				m.sessionsDrawer.statusMessage = fmt.Sprintf("Copied to clipboard: %s", toCopy)
			} else {
				m.sessionsDrawer.statusMessage = fmt.Sprintf("Copy failed: %v", err)
			}
			return m, nil
		}
	case key.Matches(msg, km.OpenDir):
		filtered := m.filteredSessions()
		if len(filtered) > 0 && m.sessionsDrawer.cursor >= 0 && m.sessionsDrawer.cursor < len(filtered) {
			target := filtered[m.sessionsDrawer.cursor]
			if target.Cwd == "" {
				m.sessionsDrawer.statusMessage = "No workspace directory recorded for this session"
				return m, nil
			}
			err := openDirectoryFunc(target.Cwd)
			if err == nil {
				m.sessionsDrawer.statusMessage = fmt.Sprintf("Opened workspace: %s", target.Cwd)
			} else {
				m.sessionsDrawer.statusMessage = fmt.Sprintf("Failed to open directory: %v", err)
			}
			return m, nil
		}
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
	claudeActive := m.sessionsDrawer.agentFilter == "claude"

	activeOnlyBadge := ""
	if m.sessionsDrawer.activeOnly {
		activeOnlyBadge = " " + lipgloss.NewStyle().Bold(true).Foreground(StatusGreen).Background(BgTabActive).Padding(0, 1).Render("[a] Active: ON")
	}

	previewBadge := ""
	if m.sessionsDrawer.hidePreview {
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
	if m.sessionsDrawer.statusMessage != "" {
		msgStyle := lipgloss.NewStyle().Bold(true).Foreground(StatusYellow)
		if strings.HasPrefix(m.sessionsDrawer.statusMessage, "Terminated") ||
			strings.HasPrefix(m.sessionsDrawer.statusMessage, "Copied") ||
			strings.HasPrefix(m.sessionsDrawer.statusMessage, "Opened") {
			msgStyle = lipgloss.NewStyle().Bold(true).Foreground(StatusGreen)
		}
		b.WriteString("  " + msgStyle.Render("▶ "+m.sessionsDrawer.statusMessage) + "\n\n")
	}

	if m.sessionsDrawer.filterActive || m.sessionsDrawer.filterInput.Value() != "" {
		b.WriteString(m.sessionsDrawer.filterInput.View() + "\n\n")
	}

	// Columns header
	colProfile := lipgloss.NewStyle().Bold(true).Foreground(AccentBlue).Width(8).Render("PROFILE")
	colAgent := lipgloss.NewStyle().Bold(true).Foreground(AccentBlue).Width(7).Render("AGENT")
	colID := lipgloss.NewStyle().Bold(true).Foreground(AccentBlue).Width(10).Render("SESSION ID")
	colDir := lipgloss.NewStyle().Bold(true).Foreground(AccentBlue).Width(18).Render("DIR")
	colTitle := lipgloss.NewStyle().Bold(true).Foreground(AccentBlue).Width(34).Render("TITLE")
	colActive := lipgloss.NewStyle().Bold(true).Foreground(AccentBlue).Width(14).Render("STATUS / TIME")

	b.WriteString(fmt.Sprintf("  %s %s %s %s %s %s\n", colProfile, colAgent, colID, colDir, colTitle, colActive))

	filtered := m.filteredSessions()
	if len(filtered) == 0 {
		b.WriteString("\n  " + lipgloss.NewStyle().Foreground(TextMuted).Render("(no conversation sessions found matching filter)") + "\n\n")
	} else {
		// Available vertical space for the sessions table.
		// Keep the table compact (default 8 rows) and account for the profiles list above
		// so that the PROFILES section remains fully visible on screen without scrolling off.
		overhead := 18
		if len(m.profiles) > 0 {
			overhead += len(m.filteredProfiles())
		}
		maxVisible := 8
		if m.sessionsDrawer.hidePreview {
			maxVisible = 14
			if m.height > 0 && m.height-overhead < maxVisible {
				maxVisible = m.height - overhead
			}
			if maxVisible < 5 {
				maxVisible = 5
			}
			if maxVisible > 16 {
				maxVisible = 16
			}
		} else {
			if m.height > 0 && m.height-overhead < maxVisible {
				maxVisible = m.height - overhead
			}
			if maxVisible < 4 {
				maxVisible = 4
			}
			if maxVisible > 8 {
				maxVisible = 8
			}
		}

		start := 0
		if m.sessionsDrawer.cursor >= maxVisible {
			start = m.sessionsDrawer.cursor - maxVisible + 1
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
			if i == m.sessionsDrawer.cursor {
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
		if !m.sessionsDrawer.hidePreview && m.sessionsDrawer.cursor >= 0 && m.sessionsDrawer.cursor < len(filtered) {
			sel := filtered[m.sessionsDrawer.cursor]
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

	km := m.keys.SessionsDrawer
	var footerHelp string
	if m.sessionsDrawer.filterActive {
		footerHelp = m.help.ShortHelpView(km.ShortHelpFilter())
	} else if m.sessionsDrawer.filterInput.Value() != "" {
		footerHelp = m.help.ShortHelpView(km.ShortHelpQuery())
	} else {
		footerHelp = m.help.ShortHelpView(km.ShortHelp())
	}
	b.WriteString("\n  " + footerHelp)

	box := SessionsDrawerStyle.Render(b.String())
	return "\n" + box + "\n"
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
