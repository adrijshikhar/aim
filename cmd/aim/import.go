package main

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/session"
	"github.com/aim-cli/aim/internal/tui"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

func newSessionsImportCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	var (
		forkFlag        bool
		allFlag         bool
		fromProfileFlag string
	)

	cmd := &cobra.Command{
		Use:   "import <agent> <target-profile> [session-id] [flags]",
		Short: "Import or hydrate a session into a target profile",
		Long: `Explicitly copy or hydrate a session from host or another profile into a target profile without immediately launching it.

Arguments:
  <agent>           Agent adapter (agy, codex, claude)
  <target-profile>  Target profile name to receive the session
  [session-id]      Specific session ID, prefix, or session name to import

Flags:
      --all         Import all sessions for this agent
      --from        Source profile or 'host' to import from (default: auto-detect latest)
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

			if len(args) >= 3 && sessionID != "" {
				if proflist, err := pm.ListProfiles(); err == nil {
					isArg1Profile := false
					isArg2Profile := false
					for _, p := range proflist {
						if p == args[1] {
							isArg1Profile = true
						}
						if p == args[2] {
							isArg2Profile = true
						}
					}
					if isArg1Profile && isArg2Profile {
						return fmt.Errorf("syntax is: aim sessions import <agent> <target-profile> <session-id> [--from <source-profile>]; both %q and %q are profile names", args[1], args[2])
					}
				}
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
				sourceProfile := "host"
				if fromProfileFlag != "" {
					sourceProfile = fromProfileFlag
				}
				sourceSessions, err := mgr.ListSessions(ctx, agentName, sourceProfile, false)
				if err != nil {
					return fmt.Errorf("failed to list sessions from %s: %w", sourceProfile, err)
				}
				if len(sourceSessions) == 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "No sessions found to import from %s.\n", sourceProfile)
					return nil
				}

				count := 0
				var failed []string
				for _, s := range sourceSessions {
					sessCopy := s
					if _, err := prov.Hydrate(ctx, &sessCopy, targetProfileDir, forkFlag); err == nil {
						count++
					} else {
						failed = append(failed, fmt.Sprintf("%s (%v)", s.ShortID, err))
					}
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s Successfully imported %d session(s) from %s into profile %s.\n",
					greenStyle.Render("✔"),
					count,
					cyanStyle.Render(sourceProfile),
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
			var sess *session.Session
			if fromProfileFlag != "" {
				isHost := fromProfileFlag == "host" || fromProfileFlag == "<host>"
				var pDir string
				if isHost {
					pDir = config.RealHomeDir()
				} else {
					pDir = filepath.Join(config.BaseDir(), "profiles", fromProfileFlag)
				}
				var getErr error
				sess, getErr = prov.GetSession(ctx, sessionID, pDir, isHost)
				if getErr != nil {
					return fmt.Errorf("failed to query session %q from profile %q: %w", sessionID, fromProfileFlag, getErr)
				}
				if sess == nil {
					// Fallback to title/name match within that profile
					proflist, err := mgr.ListSessions(ctx, agentName, fromProfileFlag, false)
					if err == nil {
						lowerTarget := strings.ToLower(strings.TrimSpace(sessionID))
						for _, s := range proflist {
							if strings.ToLower(strings.TrimSpace(s.Title)) == lowerTarget || strings.HasPrefix(strings.ToLower(strings.TrimSpace(s.Title)), lowerTarget) {
								match := s
								sess = &match
								break
							}
						}
					}
				}
				if sess == nil {
					return fmt.Errorf("session %q not found in profile %q", sessionID, fromProfileFlag)
				}
			} else {
				matches, err := mgr.FindAllSessionsByID(ctx, agentName, sessionID)
				if err != nil {
					return err
				}
				if len(matches) == 0 {
					return fmt.Errorf("%w: %s", session.ErrSessionNotFound, sessionID)
				}
				var candidates []session.Session
				for _, m := range matches {
					if m.Profile != targetProfile {
						candidates = append(candidates, m)
					}
				}
				if len(candidates) == 1 {
					sess = &candidates[0]
				} else if len(candidates) > 1 {
					sort.Slice(candidates, func(i, j int) bool {
						return candidates[i].LastActiveAt.After(candidates[j].LastActiveAt)
					})
					sess = &candidates[0]
				} else {
					if !forkFlag {
						return fmt.Errorf("session %s already belongs to target profile %q (use --fork to duplicate it, or specify --from <source-profile>)", session.ComputeShortID(matches[0].ID), targetProfile)
					}
					sess = &matches[0]
				}
			}

			hydratedID, err := prov.Hydrate(ctx, sess, targetProfileDir, forkFlag)
			if err != nil {
				return fmt.Errorf("failed to hydrate session: %w", err)
			}

			sourceName := sess.Profile
			if sourceName == "" {
				sourceName = "<host>"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s Imported session %s (%q) from profile %q into profile %q (new ID: %s).\n",
				greenStyle.Render("✔"),
				cyanStyle.Render(sess.ShortID),
				sess.Title,
				sourceName,
				targetProfile,
				cyanStyle.Render(session.ComputeShortID(hydratedID)),
			)
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completeAgentProfileAndSession(reg, pm, args, toComplete)
		},
	}

	cmd.Flags().BoolVarP(&forkFlag, "fork", "b", false, "Fork into new session IDs during import")
	cmd.Flags().BoolVar(&allFlag, "all", false, "Import all sessions from source profile (or host) into target profile")
	cmd.Flags().StringVar(&fromProfileFlag, "from", "", "Source profile or 'host' to import from (default: auto-detect latest across profiles)")

	return cmd
}
