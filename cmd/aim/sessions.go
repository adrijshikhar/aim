package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/session"
	"github.com/aim-cli/aim/internal/session/providers/agy"
	claudesess "github.com/aim-cli/aim/internal/session/providers/claude"
	"github.com/aim-cli/aim/internal/session/providers/codex"
	"github.com/aim-cli/aim/internal/tui"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
)

var defaultSessionManager = func() *session.Manager {
	mgr := session.NewManager()
	mgr.RegisterProvider(agy.NewProvider())
	mgr.RegisterProvider(codex.NewProvider())
	mgr.RegisterProvider(claudesess.NewProvider())
	return mgr
}

// SessionsCmd encapsulates dependencies for executing the sessions command.
type SessionsCmd struct {
	reg        *agents.Registry
	pm         *profile.ProfileManager
	sessionMgr *session.Manager
	isTerminal func(cmd *cobra.Command) bool
	tuiRunner  func(reg *agents.Registry, pm *profile.ProfileManager, initialAgent, profileFilter string, activeOnly bool) int
}

// NewSessionsCmd creates a new SessionsCmd factory with optional dependency overrides.
func NewSessionsCmd(reg *agents.Registry, pm *profile.ProfileManager) *SessionsCmd {
	return &SessionsCmd{
		reg: reg,
		pm:  pm,
	}
}

// WithSessionManager injects an explicit session manager.
func (sc *SessionsCmd) WithSessionManager(mgr *session.Manager) *SessionsCmd {
	sc.sessionMgr = mgr
	return sc
}

// WithIsTerminal injects a custom terminal interactivity check.
func (sc *SessionsCmd) WithIsTerminal(fn func(cmd *cobra.Command) bool) *SessionsCmd {
	sc.isTerminal = fn
	return sc
}

// WithTUIRunner injects a custom sessions TUI runner.
func (sc *SessionsCmd) WithTUIRunner(fn func(reg *agents.Registry, pm *profile.ProfileManager, initialAgent, profileFilter string, activeOnly bool) int) *SessionsCmd {
	sc.tuiRunner = fn
	return sc
}

func (sc *SessionsCmd) getSessionManager(ctx context.Context) *session.Manager {
	if sc.sessionMgr != nil {
		return sc.sessionMgr
	}
	return getSessionManager(ctx)
}

func (sc *SessionsCmd) checkInteractiveTerminal(cmd *cobra.Command) bool {
	if sc.isTerminal != nil {
		return sc.isTerminal(cmd)
	}
	if fn := InteractiveTerminalFromContext(cmd.Context()); fn != nil {
		return fn(cmd)
	}
	return isInteractiveSessionsTerminal(cmd)
}

func (sc *SessionsCmd) runTUI(ctx context.Context, reg *agents.Registry, pm *profile.ProfileManager, targetAgent, profileFlag string, activeFlag bool) int {
	if sc.tuiRunner != nil {
		return sc.tuiRunner(reg, pm, targetAgent, profileFlag, activeFlag)
	}
	return runTUISessionsWithContext(ctx, reg, pm, targetAgent, profileFlag, activeFlag)
}

func newSessionsCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	return NewSessionsCmd(reg, pm).Command()
}

// Command constructs the cobra.Command for sessions.
func (sc *SessionsCmd) Command() *cobra.Command {
	var (
		profileFlag string
		agentFlag   string
		activeFlag  bool
		allFlag     bool
		jsonFlag    bool
		plainFlag   bool
	)

	cmd := &cobra.Command{
		Use:     "sessions [agent]",
		Aliases: []string{"chats", "list-sessions"},
		Short:   "List active and past conversation sessions across profiles and host",
		Long: `List active and past conversation sessions across profiles and host.

Aliases:
  aim chats
  aim list-sessions

Flags:
  -p, --profile <name>  Filter by profile name (use "host" for host-only)
  -a, --agent <name>    Filter by agent (agy, codex, claude, etc.)
      --active          Show only currently active sessions
      --all             Show full history (default limits to 20 most recent)
      --json            Output raw JSON for scripting and automation
      --plain           Output static text table instead of interactive TUI`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetAgent := agentFlag
			if len(args) > 0 {
				targetAgent = args[0]
			}
			if sc.reg != nil && targetAgent != "" {
				if ad, err := sc.reg.Get(targetAgent); err == nil {
					targetAgent = ad.Name()
				}
			}

			if jsonFlag {
				mgr := sc.getSessionManager(cmd.Context())
				sessions, err := mgr.ListSessions(cmd.Context(), targetAgent, profileFlag, activeFlag)
				if err != nil {
					return fmt.Errorf("failed to list sessions: %w", err)
				}
				if sessions == nil {
					sessions = []session.Session{}
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(sessions)
			}

			// If interactive terminal and not plain mode, launch interactive TUI directly!
			if sc.checkInteractiveTerminal(cmd) && !plainFlag {
				code := sc.runTUI(cmd.Context(), sc.reg, sc.pm, targetAgent, profileFlag, activeFlag)
				if code != 0 {
					return &ExitError{Code: code}
				}
				return nil
			}

			// Non-interactive fallback (piped output, test buffer, or --plain)
			mgr := sc.getSessionManager(cmd.Context())
			sessions, err := mgr.ListSessions(cmd.Context(), targetAgent, profileFlag, activeFlag)
			if err != nil {
				return fmt.Errorf("failed to list sessions: %w", err)
			}

			if len(sessions) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No sessions found matching criteria.")
				return nil
			}

			if !allFlag && len(sessions) > 20 {
				sessions = sessions[:20]
			}

			renderSessionsTable(cmd.OutOrStdout(), sessions, activeFlag)
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return completeAgents(sc.reg, toComplete), cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().StringVarP(&profileFlag, "profile", "p", "", "Filter by profile name (or 'host')")
	cmd.Flags().StringVarP(&agentFlag, "agent", "a", "", "Filter by agent (agy, codex, claude)")
	cmd.Flags().BoolVar(&activeFlag, "active", false, "Show only currently active sessions")
	cmd.Flags().BoolVar(&allFlag, "all", false, "Show full history (default limits to 20)")
	cmd.Flags().BoolVar(&jsonFlag, "json", false, "Output raw JSON for scripting")
	cmd.Flags().BoolVar(&plainFlag, "plain", false, "Output static text table instead of interactive TUI")

	return cmd
}

var isInteractiveSessionsTerminal = defaultIsInteractiveTerminal

func renderSessionsTable(w io.Writer, sessions []session.Session, activeOnly bool) {
	var activeSessions []session.Session
	var recentSessions []session.Session

	for _, s := range sessions {
		if s.Status == session.StatusActive {
			activeSessions = append(activeSessions, s)
		} else {
			recentSessions = append(recentSessions, s)
		}
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(tui.TextBright)
	activeStatusStyle := lipgloss.NewStyle().Bold(true).Foreground(tui.StatusGreen)
	mutedStyle := lipgloss.NewStyle().Foreground(tui.TextMuted)
	cyanStyle := lipgloss.NewStyle().Foreground(tui.AccentCyan)

	if len(activeSessions) > 0 {
		fmt.Fprintln(w, headerStyle.Render("ACTIVE SESSIONS"))
		t := table.New().
			Border(lipgloss.HiddenBorder()).
			Headers("PROFILE", "AGENT", "SESSION ID", "DIR", "TITLE", "STARTED", "STATUS")

		for _, s := range activeSessions {
			title := truncateString(s.Title, 40)
			if title == "" {
				title = "(untitled)"
			}
			started := formatTimeOrRelative(s.StartedAt, s.LastActiveAt)
			statusStr := "ACTIVE"
			if s.PID > 0 {
				statusStr = fmt.Sprintf("ACTIVE (PID %d)", s.PID)
			}
			dir := session.FormatDir(s.Cwd, 20)

			t.Row(
				s.Profile,
				s.Agent,
				s.ShortID,
				dir,
				cyanStyle.Render(title),
				started,
				activeStatusStyle.Render(statusStr),
			)
		}

		t.StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return lipgloss.NewStyle().Bold(true).Foreground(tui.AccentBlue)
			}
			return lipgloss.NewStyle().Foreground(tui.TextPrimary)
		})

		fmt.Fprintln(w, t.Render())
	}

	if !activeOnly && len(recentSessions) > 0 {
		if len(activeSessions) > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintln(w, headerStyle.Render("RECENT SESSIONS"))
		t := table.New().
			Border(lipgloss.HiddenBorder()).
			Headers("PROFILE", "AGENT", "SESSION ID", "DIR", "TITLE", "LAST ACTIVE")

		for _, s := range recentSessions {
			title := truncateString(s.Title, 46)
			if title == "" {
				title = "(untitled)"
			}
			lastActive := formatRelativeTime(s.LastActiveAt)
			dir := session.FormatDir(s.Cwd, 20)

			t.Row(
				s.Profile,
				s.Agent,
				s.ShortID,
				dir,
				title,
				mutedStyle.Render(lastActive),
			)
		}

		t.StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return lipgloss.NewStyle().Bold(true).Foreground(tui.AccentBlue)
			}
			return lipgloss.NewStyle().Foreground(tui.TextPrimary)
		})

		fmt.Fprintln(w, t.Render())
	}
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
