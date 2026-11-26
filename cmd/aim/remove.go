package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/telemetry"
	"github.com/aim-cli/aim/internal/tui"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

func newRemoveCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	return &cobra.Command{
		Use:   "remove [agent] <profile>",
		Short: "Delete profile credentials and state",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			exitCode := executeRemoveTo(cmd.OutOrStdout(), cmd.ErrOrStderr(), reg, pm, args)
			if exitCode != 0 {
				return &ExitError{Code: exitCode}
			}
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				agents := completeAgents(reg, toComplete)
				profiles := completeAllProfiles(pm, toComplete)
				return append(agents, profiles...), cobra.ShellCompDirectiveNoFileComp
			}
			if len(args) == 1 {
				return completeProfilesForAgent(pm, reg, args[0], toComplete), cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}
}

func executeRemove(reg *agents.Registry, pm *profile.ProfileManager, args []string) int {
	return executeRemoveTo(os.Stdout, os.Stderr, reg, pm, args)
}

func executeRemoveTo(out, errOut io.Writer, reg *agents.Registry, pm *profile.ProfileManager, args []string) int {
	var agentName, profileName string
	switch len(args) {
	case 0:
		_, _ = fmt.Fprintln(out, "Usage: aim remove [agent] <profile>")
		return 1
	case 1:
		profileName = args[0]
	default:
		agentName = args[0]
		profileName = args[1]
	}

	if agentName != "" && reg != nil {
		if adapter, err := reg.Get(agentName); err == nil {
			agentName = adapter.Name()
		}
	}

	cfg, _ := config.LoadConfig()
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	okBadge := lipgloss.NewStyle().Foreground(tui.StatusGreen).Bold(true).Render("✔")
	cyan := lipgloss.NewStyle().Foreground(tui.AccentCyan)

	if agentName != "" {
		return executeRemoveAgent(out, errOut, reg, pm, cfg, agentName, profileName, okBadge, cyan)
	}
	return executeRemoveProfile(out, errOut, reg, pm, cfg, profileName, okBadge, cyan)
}

func executeRemoveAgent(out, errOut io.Writer, reg *agents.Registry, pm *profile.ProfileManager, cfg *config.Config, agentName, profileName, okBadge string, cyan lipgloss.Style) int {
	bold := lipgloss.NewStyle().Bold(true).Foreground(tui.TextBright)
	account := profile.GetProfileAccountInfoForAgent(pm.ProfileDir(profileName), agentName)
	cleanedUp, err := pm.RemoveAgent(profileName, agentName, cfg)
	if err != nil {
		errBadge := lipgloss.NewStyle().Foreground(tui.StatusRed).Bold(true).Render("✖")
		if strings.Contains(err.Error(), "not associated") {
			_, _ = fmt.Fprintf(errOut, "%s Agent %q is not associated with profile %q\n", errBadge, agentName, profileName)
		} else {
			_, _ = fmt.Fprintf(errOut, "%s Error: %v\n", errBadge, err)
		}
		return 1
	}
	purgeRemovedAgent(reg, agentName, profileName, pm.ProfileDir(profileName))
	telemetry.CaptureAccountEvent(agentName, account, "agent_removed", map[string]any{
		"profile_cleaned_up": cleanedUp,
	})
	if cleanedUp {
		_, _ = fmt.Fprintf(out, "%s Removed agent %s and cleaned up profile %s\n", okBadge, bold.Render(agentName), cyan.Render(profileName))
	} else {
		_, _ = fmt.Fprintf(out, "%s Removed agent %s from profile %s\n", okBadge, bold.Render(agentName), cyan.Render(profileName))
	}
	return 0
}

func executeRemoveProfile(out, errOut io.Writer, reg *agents.Registry, pm *profile.ProfileManager, cfg *config.Config, profileName, okBadge string, cyan lipgloss.Style) int {
	pDir := pm.ProfileDir(profileName)
	if err := pm.DeleteProfile(profileName, cfg); err != nil {
		errBadge := lipgloss.NewStyle().Foreground(tui.StatusRed).Bold(true).Render("✖")
		_, _ = fmt.Fprintf(errOut, "%s Error: %v\n", errBadge, err)
		return 1
	}
	purgeRemovedProfile(reg, profileName, pDir)
	_, _ = fmt.Fprintf(out, "%s Profile %s removed successfully.\n", okBadge, cyan.Render(profileName))
	return 0
}

func purgeRemovedAgent(reg *agents.Registry, agentName, profileName, pDir string) {
	if reg == nil {
		return
	}
	ad, err := reg.Get(agentName)
	if err != nil {
		return
	}
	if purger, ok := ad.(agents.CredentialPurger); ok {
		_ = purger.PurgeCredentials(context.Background(), profileName, pDir)
	}
}

func purgeRemovedProfile(reg *agents.Registry, profileName, pDir string) {
	if reg == nil {
		return
	}
	for _, ad := range reg.All() {
		if purger, ok := ad.(agents.CredentialPurger); ok {
			_ = purger.PurgeCredentials(context.Background(), profileName, pDir)
		}
	}
}

func runRemove(reg *agents.Registry, pm *profile.ProfileManager, args []string) error {
	if code := executeRemove(reg, pm, args); code != 0 {
		return fmt.Errorf("remove failed with exit code %d", code)
	}
	return nil
}
