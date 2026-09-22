package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/session"
	"github.com/aim-cli/aim/internal/session/catalyst"
	"github.com/aim-cli/aim/internal/tui"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

func newResumeCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	var (
		exactFlag    bool
		catalystFlag bool
		forkFlag     bool
		forceFlag    bool
	)

	cmd := &cobra.Command{
		Use:   "resume <agent> <profile> [session-id] [flags] [-- args...]",
		Short: "Resume an existing session under a specified profile",
		Long: `Resume an existing session under a specified profile.

Arguments:
  <agent>       Agent adapter (agy, codex)
  <profile>     Destination profile name
  [session-id]  Full UUID or short prefix (e.g. 8 chars)

Flags:
      --exact     Resume native verbatim conversation thread (default)
  -c, --catalyst  Resume via Catalyst summary handoff with fresh context window
  -s, --summary   Alias for --catalyst
  -b, --fork      Fork session into a new conversation ID
  -f, --force     Force resume even if session is currently active in another process`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("resume requires <agent> and <profile>")
			}
			agentName := args[0]
			profileName := args[1]

			if reg != nil {
				if ad, err := reg.Get(agentName); err == nil {
					agentName = ad.Name()
				} else {
					return err
				}
			}

			var sessionID string
			var extraArgs []string

			// Parse sessionID and extraArgs
			if len(args) > 2 {
				if args[2] == "--" {
					extraArgs = args[3:]
				} else {
					sessionID = args[2]
					if len(args) > 3 {
						if args[3] == "--" {
							extraArgs = args[4:]
						} else {
							extraArgs = args[3:]
						}
					}
				}
			}

			if sessionID == "" {
				return fmt.Errorf("missing session ID to resume; run 'aim sessions %s' to list available sessions", agentName)
			}

			mgr := defaultSessionManager()
			ctx := cmd.Context()
			sess, err := mgr.ResolveSession(ctx, agentName, sessionID)
			if err != nil {
				return err
			}

			// Active process check
			if sess.Status == session.StatusActive && !forceFlag {
				if isInteractive(cmd.InOrStdin()) {
					fmt.Fprintf(cmd.OutOrStdout(), "Session %s (%q) is currently ACTIVE in PID %d.\nResume anyway? [y/N]: ", sess.ShortID, sess.Title, sess.PID)
					reader := bufio.NewReader(cmd.InOrStdin())
					ans, _ := reader.ReadString('\n')
					ans = strings.TrimSpace(strings.ToLower(ans))
					if ans != "y" && ans != "yes" {
						fmt.Fprintln(cmd.OutOrStdout(), "Aborted resume.")
						return nil
					}
				} else {
					return fmt.Errorf("session %s is currently ACTIVE in PID %d; use --force to resume anyway", sess.ShortID, sess.PID)
				}
			}

			ok, err := confirmProfileExists(cmd, pm, profileName, fmt.Sprintf("resume %s", agentName), false)
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}

			pDir, err := pm.EnsureProfile(profileName)
			if err != nil {
				return fmt.Errorf("failed to ensure destination profile %q: %w", profileName, err)
			}

			if catalystFlag {
				return executeCatalystResume(cmd, reg, pm, agentName, profileName, sess, extraArgs)
			}

			return executeExactResume(cmd, reg, pm, mgr, agentName, profileName, pDir, sess, forkFlag, extraArgs)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completeAgentAndProfile(reg, pm, args, toComplete)
		},
	}

	cmd.Flags().BoolVar(&exactFlag, "exact", false, "Resume native verbatim conversation thread")
	cmd.Flags().BoolVarP(&catalystFlag, "catalyst", "c", false, "Resume via Catalyst summary handoff")
	cmd.Flags().BoolVarP(&catalystFlag, "summary", "s", false, "Alias for --catalyst")
	cmd.Flags().BoolVarP(&forkFlag, "fork", "b", false, "Fork session into a new conversation ID")
	cmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Force resume even if session is currently active")

	return cmd
}

func executeCatalystResume(cmd *cobra.Command, reg *agents.Registry, pm *profile.ProfileManager, agent, profile string, sess *session.Session, extraArgs []string) error {
	repoRoot := getGitRepoRoot()
	branch := getGitBranch(repoRoot)

	bridge := catalyst.NewBridge()
	handoffPath, err := bridge.WriteHandoffBrief(repoRoot, branch, sess)
	if err != nil {
		return fmt.Errorf("failed to write Catalyst handoff brief: %w", err)
	}

	cyanStyle := lipgloss.NewStyle().Foreground(tui.AccentCyan)
	greenStyle := lipgloss.NewStyle().Foreground(tui.StatusGreen)
	fmt.Fprintf(cmd.OutOrStdout(), "%s Catalyst handoff brief written to %s\n",
		greenStyle.Render("✔"),
		cyanStyle.Render(handoffPath),
	)

	// Launch agent primed with Catalyst handoff resume
	launchArgs := append([]string{"handoff resume"}, extraArgs...)
	code := executeRun(reg, pm, agent, profile, launchArgs)
	if code != 0 {
		return &ExitError{Code: code}
	}
	return nil
}

func executeExactResume(cmd *cobra.Command, reg *agents.Registry, pm *profile.ProfileManager, mgr *session.Manager, agent, profile, pDir string, sess *session.Session, fork bool, extraArgs []string) error {
	resumeID := sess.ID

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Check if another profile has a newer version of this session
	if mgr != nil {
		if latest, err := findLatestSessionAcrossProfiles(ctx, mgr, agent, sess.ID); err == nil && latest != nil {
			if latest.LastActiveAt.After(sess.LastActiveAt) {
				sess = latest
			}
		}
	}

	// If session belongs to host or a different profile, or fork requested, hydrate into dest profile
	if sess.IsHost || sess.Profile != profile || fork {
		prov := mgr.Provider(agent)
		if prov != nil {
			hydratedID, err := prov.Hydrate(ctx, sess, pDir, fork)
			if err != nil {
				return fmt.Errorf("failed to hydrate session into profile %q: %w", profile, err)
			}
			resumeID = hydratedID
		}
	}

	var resumeArgs []string
	switch agent {
	case "agy":
		resumeArgs = append([]string{fmt.Sprintf("--conversation=%s", resumeID)}, extraArgs...)
	case "codex":
		resumeArgs = append([]string{"resume", resumeID}, extraArgs...)
	default:
		resumeArgs = append([]string{resumeID}, extraArgs...)
	}

	arrowStyle := lipgloss.NewStyle().Foreground(tui.AccentBlue).Bold(true)
	cyanStyle := lipgloss.NewStyle().Foreground(tui.AccentCyan)
	boldStyle := lipgloss.NewStyle().Bold(true).Foreground(tui.TextBright)
	fmt.Fprintf(cmd.OutOrStdout(), "%s Resuming %s session %s under profile %q...\n",
		arrowStyle.Render("➜"),
		boldStyle.Render(agent),
		cyanStyle.Render(sess.ShortID),
		profile,
	)

	code := executeRun(reg, pm, agent, profile, resumeArgs)
	if code != 0 {
		return &ExitError{Code: code}
	}
	return nil
}

func getGitRepoRoot() string {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err == nil && len(out) > 0 {
		return strings.TrimSpace(string(out))
	}
	cwd, err := os.Getwd()
	if err == nil {
		return cwd
	}
	return "."
}

func getGitBranch(repoRoot string) string {
	cmd := exec.Command("git", "branch", "--show-current")
	if repoRoot != "" {
		cmd.Dir = repoRoot
	}
	out, err := cmd.Output()
	if err == nil && len(strings.TrimSpace(string(out))) > 0 {
		return strings.TrimSpace(string(out))
	}
	return "main"
}

func isInteractive(r io.Reader) bool {
	return isInteractiveFunc(r)
}

func findLatestSessionAcrossProfiles(ctx context.Context, mgr *session.Manager, agent, sessionID string) (*session.Session, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if mgr == nil {
		return nil, fmt.Errorf("session manager cannot be nil")
	}
	matches, err := mgr.FindAllSessionsByID(ctx, agent, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions across profiles: %w", err)
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("%w: %s", session.ErrSessionNotFound, sessionID)
	}
	var latest *session.Session
	for i := range matches {
		s := &matches[i]
		if latest == nil || s.LastActiveAt.After(latest.LastActiveAt) {
			latest = s
		}
	}
	return latest, nil
}

