package main

import (
	"fmt"
	"strings"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/session"
	"github.com/aim-cli/aim/internal/tui"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

func newSessionsImportCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	var (
		forkFlag bool
		allFlag  bool
	)

	cmd := &cobra.Command{
		Use:   "import <agent> <target-profile> [session-id] [flags]",
		Short: "Import or hydrate a session from host into a target profile",
		Long: `Explicitly copy or hydrate a session from host into a target profile without immediately launching it.

Arguments:
  <agent>           Agent adapter (agy, codex)
  <target-profile>  Target profile name to receive the session
  [session-id]      Specific session ID or prefix to import

Flags:
      --all         Import all sessions for this agent from host
  -b, --fork        Fork into new session IDs during import`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("import requires <agent> and <target-profile>")
			}
			agentName := args[0]
			targetProfile := args[1]

			if reg != nil {
				if ad, err := reg.Get(agentName); err == nil {
					agentName = ad.Name()
				} else {
					return err
				}
			}

			var sessionID string
			if len(args) > 2 {
				sessionID = args[2]
			}

			if sessionID == "" && !allFlag {
				return fmt.Errorf("must specify session-id or --all")
			}

			targetProfileDir, err := pm.EnsureProfile(targetProfile)
			if err != nil {
				return fmt.Errorf("failed to ensure profile %q: %w", targetProfile, err)
			}

			mgr := defaultSessionManager()
			ctx := cmd.Context()
			prov := mgr.Provider(agentName)
			if prov == nil {
				return fmt.Errorf("no session provider registered for agent %q", agentName)
			}

			cyanStyle := lipgloss.NewStyle().Foreground(tui.AccentCyan)
			greenStyle := lipgloss.NewStyle().Foreground(tui.StatusGreen)

			if allFlag {
				hostSessions, err := mgr.ListSessions(ctx, agentName, "host", false)
				if err != nil {
					return fmt.Errorf("failed to list host sessions: %w", err)
				}
				if len(hostSessions) == 0 {
					fmt.Fprintln(cmd.OutOrStdout(), lipgloss.NewStyle().Foreground(tui.TextMuted).Render("No host sessions found to import."))
					return nil
				}

				count := 0
				var failed []string
				for _, s := range hostSessions {
					sessCopy := s
					if _, err := prov.Hydrate(ctx, &sessCopy, targetProfileDir, forkFlag); err == nil {
						count++
					} else {
						failed = append(failed, fmt.Sprintf("%s (%v)", s.ShortID, err))
					}
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s Successfully imported %d host session(s) into profile %s.\n",
					greenStyle.Render("✔"),
					count,
					cyanStyle.Render(targetProfile),
				)
				if len(failed) > 0 {
					yellowStyle := lipgloss.NewStyle().Foreground(tui.StatusYellow)
					fmt.Fprintf(cmd.OutOrStdout(), "%s Warning: %d session(s) failed to import: %s\n",
						yellowStyle.Render("!"),
						len(failed),
						strings.Join(failed, ", "),
					)
				}
				return nil
			}

			// Single session import
			sess, err := mgr.ResolveSession(ctx, agentName, sessionID)
			if err != nil {
				return err
			}

			hydratedID, err := prov.Hydrate(ctx, sess, targetProfileDir, forkFlag)
			if err != nil {
				return fmt.Errorf("failed to hydrate session: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%s Imported session %s (%q) into profile %q (new ID: %s).\n",
				greenStyle.Render("✔"),
				cyanStyle.Render(sess.ShortID),
				sess.Title,
				targetProfile,
				cyanStyle.Render(session.ComputeShortID(hydratedID)),
			)
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completeAgentAndProfile(reg, pm, args, toComplete)
		},
	}

	cmd.Flags().BoolVarP(&forkFlag, "fork", "b", false, "Fork into new session IDs during import")
	cmd.Flags().BoolVar(&allFlag, "all", false, "Import all sessions from host into the target profile")

	return cmd
}
