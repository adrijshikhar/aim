package main

import (
	"fmt"
	"io"
	"os"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/oauth"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/telemetry"
	"github.com/aim-cli/aim/internal/tui"
	"github.com/aim-cli/aim/internal/usage"
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
			exitCode := executeMvTo(cmd.OutOrStdout(), cmd.ErrOrStderr(), reg, pm, args[0], args[1], args[2], forceFlag)
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
	return executeMvTo(os.Stdout, os.Stderr, reg, pm, agentName, sourceProfile, targetProfile, force)
}

func executeMvTo(out, errOut io.Writer, reg *agents.Registry, pm *profile.ProfileManager, agentName, sourceProfile, targetProfile string, force bool) int {
	if out == nil {
		out = os.Stdout
	}
	if errOut == nil {
		errOut = os.Stderr
	}
	errBadge := lipgloss.NewStyle().Foreground(tui.StatusRed).Bold(true).Render("✖")
	if pm == nil {
		_, _ = fmt.Fprintf(errOut, "%s Error: profile manager not initialized\n", errBadge)
		return 1
	}
	if reg == nil {
		reg = defaultRegistry()
	}

	cfg, err := loadOrRecoverConfig(errOut)
	if err != nil {
		_, _ = fmt.Fprintf(errOut, "%s Error loading config: %v\n", errBadge, err)
		return 1
	}

	adapter, err := reg.Get(agentName)
	if err != nil {
		_, _ = fmt.Fprintf(errOut, "%s Error: %v\n", errBadge, err)
		return 1
	}
	agentName = adapter.Name()

	account := profile.GetProfileAccountInfoForAgent(pm.ProfileDir(sourceProfile), agentName)
	if err := pm.MoveAgent(agentName, sourceProfile, targetProfile, force, cfg, reg); err != nil {
		_, _ = fmt.Fprintf(errOut, "%s Error: %v\n", errBadge, err)
		return 1
	}

	baseDir := config.BaseDir()
	if pm.BaseDir != "" {
		baseDir = pm.BaseDir
	}
	refStore := oauth.NewRefusalStore(baseDir)
	_ = refStore.Clear(agentName, sourceProfile)
	_ = refStore.Clear(agentName, targetProfile)
	cacheStore := usage.NewCacheStore(baseDir, usage.DefaultTTL)
	cacheStore.Delete(agentName, sourceProfile)
	cacheStore.Delete(agentName, targetProfile)

	okBadge := lipgloss.NewStyle().Foreground(tui.StatusGreen).Bold(true).Render("✔")
	cyan := lipgloss.NewStyle().Foreground(tui.AccentCyan)
	bold := lipgloss.NewStyle().Bold(true).Foreground(tui.TextBright)

	telemetry.CaptureAccountEvent(agentName, account, "account_moved", map[string]any{
		"overwrote_target": force,
	})

	_, _ = fmt.Fprintf(out, "%s Moved agent %s from profile %s to %s.\n",
		okBadge,
		bold.Render(agentName),
		cyan.Render(sourceProfile),
		cyan.Render(targetProfile),
	)
	return 0
}
