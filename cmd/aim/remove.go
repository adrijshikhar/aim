package main

import (
	"fmt"
	"io"
	"os"
	"strings"

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
	if out == nil {
		out = os.Stdout
	}
	if errOut == nil {
		errOut = os.Stderr
	}
	if reg == nil {
		reg = defaultRegistry()
	}

	var agentName, profileName string
	switch len(args) {
	case 0:
		_, _ = fmt.Fprintln(errOut, "Usage: aim remove [agent] <profile>")
		return 1
	case 1:
		profileName = args[0]
	default:
		agentName = args[0]
		profileName = args[1]
	}

	errBadge := lipgloss.NewStyle().Foreground(tui.StatusRed).Bold(true).Render("✖")
	if pm == nil {
		_, _ = fmt.Fprintf(errOut, "%s Error: profile manager not initialized\n", errBadge)
		return 1
	}

	if agentName != "" {
		adapter, err := reg.Get(agentName)
		if err != nil {
			_, _ = fmt.Fprintf(errOut, "%s Error: %v\n", errBadge, err)
			return 1
		}
		agentName = adapter.Name()
	}

	cfg, err := loadOrRecoverConfig(errOut)
	if err != nil {
		_, _ = fmt.Fprintf(errOut, "%s Error loading config: %v\n", errBadge, err)
		return 1
	}

	okBadge := lipgloss.NewStyle().Foreground(tui.StatusGreen).Bold(true).Render("✔")
	cyan := lipgloss.NewStyle().Foreground(tui.AccentCyan)

	if agentName != "" {
		return executeRemoveAgent(out, errOut, reg, pm, cfg, agentName, profileName, okBadge, errBadge, cyan)
	}
	return executeRemoveProfile(out, errOut, reg, pm, cfg, profileName, okBadge, errBadge, cyan)
}

func executeRemoveAgent(out, errOut io.Writer, reg *agents.Registry, pm *profile.ProfileManager, cfg *config.Config, agentName, profileName, okBadge, errBadge string, cyan lipgloss.Style) int {
	bold := lipgloss.NewStyle().Bold(true).Foreground(tui.TextBright)
	account := profile.GetProfileAccountInfoForAgent(pm.ProfileDir(profileName), agentName)
	purgers := collectAgentPurgers(reg, agentName)
	cleanedUp, err := pm.RemoveAgent(profileName, agentName, cfg, purgers...)
	if err != nil {
		if strings.Contains(err.Error(), "not associated") {
			_, _ = fmt.Fprintf(errOut, "%s Agent %q is not associated with profile %q\n", errBadge, agentName, profileName)
		} else {
			_, _ = fmt.Fprintf(errOut, "%s Error: %v\n", errBadge, err)
		}
		return 1
	}
	clearRemovedStateStores(pm, profileName, agentName)
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

func executeRemoveProfile(out, errOut io.Writer, reg *agents.Registry, pm *profile.ProfileManager, cfg *config.Config, profileName, okBadge, errBadge string, cyan lipgloss.Style) int {
	purgers := collectAllPurgers(reg)
	if err := pm.DeleteProfile(profileName, cfg, purgers...); err != nil {
		_, _ = fmt.Fprintf(errOut, "%s Error: %v\n", errBadge, err)
		return 1
	}
	for _, ad := range reg.All() {
		clearRemovedStateStores(pm, profileName, ad.Name())
	}
	_, _ = fmt.Fprintf(out, "%s Profile %s removed successfully.\n", okBadge, cyan.Render(profileName))
	return 0
}

func collectAgentPurgers(reg *agents.Registry, agentName string) []agents.CredentialPurger {
	if reg == nil {
		return nil
	}
	ad, err := reg.Get(agentName)
	if err != nil {
		return nil
	}
	if purger, ok := ad.(agents.CredentialPurger); ok {
		return []agents.CredentialPurger{purger}
	}
	return nil
}

func collectAllPurgers(reg *agents.Registry) []agents.CredentialPurger {
	if reg == nil {
		return nil
	}
	var purgers []agents.CredentialPurger
	for _, ad := range reg.All() {
		if purger, ok := ad.(agents.CredentialPurger); ok {
			purgers = append(purgers, purger)
		}
	}
	return purgers
}

func clearRemovedStateStores(pm *profile.ProfileManager, profileName, agentName string) {
	baseDir := config.BaseDir()
	if pm != nil && pm.BaseDir != "" {
		baseDir = pm.BaseDir
	}
	_ = oauth.NewRefusalStore(baseDir).Clear(agentName, profileName)
	usage.NewCacheStore(baseDir, usage.DefaultTTL).Delete(agentName, profileName)
}

func runRemove(reg *agents.Registry, pm *profile.ProfileManager, args []string) error {
	if code := executeRemove(reg, pm, args); code != 0 {
		return fmt.Errorf("remove failed with exit code %d", code)
	}
	return nil
}
