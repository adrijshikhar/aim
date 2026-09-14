package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/session"
	"github.com/aim-cli/aim/internal/session/providers/agy"
	"github.com/aim-cli/aim/internal/session/providers/codex"
	"github.com/aim-cli/aim/internal/tui"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

func defaultSessionManager() *session.Manager {
	mgr := session.NewManager()
	mgr.RegisterProvider(agy.NewProvider())
	mgr.RegisterProvider(codex.NewProvider())
	return mgr
}

func newSessionsCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	var (
		profileFlag string
		agentFlag   string
		activeFlag  bool
		allFlag     bool
		jsonFlag    bool
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
  -a, --agent <name>    Filter by agent (agy, codex, etc.)
      --active          Show only currently active sessions
      --all             Show full history (default limits to 20 most recent)
      --json            Output raw JSON for scripting and automation`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetAgent := agentFlag
			if len(args) > 0 {
				targetAgent = args[0]
			}
			if reg != nil && targetAgent != "" {
				if ad, err := reg.Get(targetAgent); err == nil {
					targetAgent = ad.Name()
				}
			}

			mgr := defaultSessionManager()
			sessions, err := mgr.ListSessions(cmd.Context(), targetAgent, profileFlag, activeFlag)
			if err != nil {
				return fmt.Errorf("failed to list sessions: %w", err)
			}

			if jsonFlag {
				if sessions == nil {
					sessions = []session.Session{}
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(sessions)
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
				return completeAgents(reg, toComplete), cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().StringVarP(&profileFlag, "profile", "p", "", "Filter by profile name (or 'host')")
	cmd.Flags().StringVarP(&agentFlag, "agent", "a", "", "Filter by agent (agy, codex)")
	cmd.Flags().BoolVar(&activeFlag, "active", false, "Show only currently active sessions")
	cmd.Flags().BoolVar(&allFlag, "all", false, "Show full history (default limits to 20)")
	cmd.Flags().BoolVar(&jsonFlag, "json", false, "Output raw JSON for scripting")

	cmd.AddCommand(newSessionsImportCmd(reg, pm))
	cmd.AddCommand(newSessionsShowCmd(reg, pm))

	return cmd
}

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
	colHeaderStyle := lipgloss.NewStyle().Bold(true).Foreground(tui.AccentBlue)
	activeStatusStyle := lipgloss.NewStyle().Bold(true).Foreground(tui.StatusGreen)
	mutedStyle := lipgloss.NewStyle().Foreground(tui.TextMuted)
	cyanStyle := lipgloss.NewStyle().Foreground(tui.AccentCyan)

	if len(activeSessions) > 0 {
		fmt.Fprintln(w, headerStyle.Render("ACTIVE SESSIONS"))
		fmt.Fprintf(w, "%-10s %-8s %-12s %-32s %-12s %s\n",
			colHeaderStyle.Render("PROFILE"),
			colHeaderStyle.Render("AGENT"),
			colHeaderStyle.Render("SESSION ID"),
			colHeaderStyle.Render("TITLE"),
			colHeaderStyle.Render("STARTED"),
			colHeaderStyle.Render("STATUS"),
		)
		for _, s := range activeSessions {
			title := truncateString(s.Title, 30)
			if title == "" {
				title = "(untitled)"
			}
			started := formatTimeOrRelative(s.StartedAt, s.LastActiveAt)
			statusStr := "ACTIVE"
			if s.PID > 0 {
				statusStr = fmt.Sprintf("ACTIVE (PID %d)", s.PID)
			}

			fmt.Fprintf(w, "%-10s %-8s %-12s %-32s %-12s %s\n",
				s.Profile,
				s.Agent,
				s.ShortID,
				cyanStyle.Render(title),
				started,
				activeStatusStyle.Render(statusStr),
			)
		}
	}

	if !activeOnly && len(recentSessions) > 0 {
		if len(activeSessions) > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintln(w, headerStyle.Render("RECENT SESSIONS"))
		fmt.Fprintf(w, "%-10s %-8s %-12s %-46s %s\n",
			colHeaderStyle.Render("PROFILE"),
			colHeaderStyle.Render("AGENT"),
			colHeaderStyle.Render("SESSION ID"),
			colHeaderStyle.Render("TITLE"),
			colHeaderStyle.Render("LAST ACTIVE"),
		)
		for _, s := range recentSessions {
			title := truncateString(s.Title, 44)
			if title == "" {
				title = "(untitled)"
			}
			lastActive := formatRelativeTime(s.LastActiveAt)

			fmt.Fprintf(w, "%-10s %-8s %-12s %-46s %s\n",
				s.Profile,
				s.Agent,
				s.ShortID,
				title,
				mutedStyle.Render(lastActive),
			)
		}
	}
}

func truncateString(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
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

func newSessionsShowCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	var (
		agentFlag string
		jsonFlag  bool
	)

	cmd := &cobra.Command{
		Use:     "show [agent] <session-id>",
		Aliases: []string{"preview", "info", "inspect"},
		Short:   "Show detailed preview and metadata for a conversation session",
		Long: `Show detailed preview, goal summary, storage path, and status for a specific conversation session.

Arguments:
  [agent]       Optional agent filter (agy, codex)
  <session-id>  Full UUID or short prefix (e.g. 8 chars)

Aliases:
  aim sessions preview <session-id>
  aim sessions info <session-id>
  aim sessions inspect <session-id>`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var agentName string
			var sessionID string

			if len(args) == 1 {
				sessionID = args[0]
				agentName = agentFlag
			} else {
				agentName = args[0]
				sessionID = args[1]
			}

			if reg != nil && agentName != "" {
				if ad, err := reg.Get(agentName); err == nil {
					agentName = ad.Name()
				}
			}

			mgr := defaultSessionManager()
			ctx := cmd.Context()
			sess, err := mgr.ResolveSession(ctx, agentName, sessionID)
			if err != nil {
				return err
			}

			if jsonFlag {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(sess)
			}

			renderSessionPreviewCard(cmd.OutOrStdout(), sess)
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return completeAgents(reg, toComplete), cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().StringVarP(&agentFlag, "agent", "a", "", "Filter by agent (agy, codex)")
	cmd.Flags().BoolVar(&jsonFlag, "json", false, "Output raw JSON for scripting")

	return cmd
}

func renderSessionPreviewCard(w io.Writer, s *session.Session) {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(tui.AccentBlue)
	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(tui.TextBright).Width(16)
	valStyle := lipgloss.NewStyle().Foreground(tui.TextPrimary)
	cyanStyle := lipgloss.NewStyle().Foreground(tui.AccentCyan)
	greenStyle := lipgloss.NewStyle().Bold(true).Foreground(tui.StatusGreen)
	yellowStyle := lipgloss.NewStyle().Foreground(tui.StatusYellow)
	mutedStyle := lipgloss.NewStyle().Foreground(tui.TextMuted)

	statusStr := mutedStyle.Render("IDLE")
	if s.Status == session.StatusActive {
		if s.PID > 0 {
			statusStr = greenStyle.Render(fmt.Sprintf("ACTIVE (PID %d)", s.PID))
		} else {
			statusStr = greenStyle.Render("ACTIVE")
		}
	}

	previewText := s.Summary
	if previewText == "" {
		previewText = s.Title
	}
	previewText = strings.TrimSpace(previewText)
	if previewText == "" {
		previewText = "(no preview text recorded)"
	}

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.AccentBlue).
		Padding(1, 2)

	var b strings.Builder
	b.WriteString(headerStyle.Render("⚡ Conversation Session Preview") + "\n\n")

	b.WriteString(fmt.Sprintf("%s %s (%s)\n", labelStyle.Render("Session ID:"), cyanStyle.Render(s.ShortID), mutedStyle.Render(s.ID)))
	b.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Agent:"), valStyle.Render(s.Agent)))
	b.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Profile:"), yellowStyle.Render(s.Profile)))
	b.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Status:"), statusStr))
	if !s.LastActiveAt.IsZero() {
		b.WriteString(fmt.Sprintf("%s %s (%s)\n", labelStyle.Render("Last Active:"), valStyle.Render(formatRelativeTime(s.LastActiveAt)), mutedStyle.Render(s.LastActiveAt.Format(time.RFC3339))))
	}
	if s.StoragePath != "" {
		b.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Storage:"), mutedStyle.Render(s.StoragePath)))
	}

	b.WriteString("\n" + labelStyle.Render("Title:") + "\n")
	b.WriteString("  " + lipgloss.NewStyle().Bold(true).Render(s.Title) + "\n\n")

	b.WriteString(labelStyle.Render("Summary / Goal:") + "\n")
	lines := strings.Split(previewText, "\n")
	for _, l := range lines {
		b.WriteString("  " + valStyle.Render(l) + "\n")
	}

	b.WriteString("\n" + mutedStyle.Render(fmt.Sprintf("Quick resume: aim resume %s %s %s", s.Agent, s.Profile, s.ShortID)) + "\n")

	fmt.Fprint(w, cardStyle.Render(b.String())+"\n")
}
