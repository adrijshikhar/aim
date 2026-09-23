package main

import (
	"fmt"
	"os"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/tui"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

func newMvCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	var forceFlag bool

	cmd := &cobra.Command{
		Use:   "mv <agent> <src-profile> <target-profile> [flags]",
		Short: "Move an agent account and credentials from one profile to another",
		Long: `Move an agent account, tokens, credentials, and session state from a source profile to a target profile.

Arguments:
  <agent>           Agent adapter (agy, codex, claude, gemini)
  <src-profile>     Source profile currently holding the account
  <target-profile>  Destination profile to receive the account

Flags:
  -f, --force       Overwrite existing credentials in target profile if already present`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			exitCode := executeMv(reg, pm, args[0], args[1], args[2], forceFlag)
			if exitCode != 0 {
				return &ExitError{Code: exitCode}
			}
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return completeAgents(reg, toComplete), cobra.ShellCompDirectiveNoFileComp
			}
			if len(args) == 1 {
				return completeProfilesForAgent(pm, reg, args[0], toComplete), cobra.ShellCompDirectiveNoFileComp
			}
			if len(args) == 2 {
				return completeAllProfiles(pm, toComplete), cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Overwrite existing credentials in target profile if present")
	return cmd
}

func executeMv(reg *agents.Registry, pm *profile.ProfileManager, agentName, sourceProfile, targetProfile string, force bool) int {
	cfg, _ := config.LoadConfig()
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	err := pm.MoveAgent(agentName, sourceProfile, targetProfile, force, cfg, reg)
	if err != nil {
		errBadge := lipgloss.NewStyle().Foreground(tui.StatusRed).Bold(true).Render("✖")
		fmt.Fprintf(os.Stderr, "%s Error: %v\n", errBadge, err)
		return 1
	}

	okBadge := lipgloss.NewStyle().Foreground(tui.StatusGreen).Bold(true).Render("✔")
	cyan := lipgloss.NewStyle().Foreground(tui.AccentCyan)
	bold := lipgloss.NewStyle().Bold(true).Foreground(tui.TextBright)

	fmt.Printf("%s Moved agent %s from profile %s to %s.\n",
		okBadge,
		bold.Render(agentName),
		cyan.Render(sourceProfile),
		cyan.Render(targetProfile),
	)
	return 0
}
