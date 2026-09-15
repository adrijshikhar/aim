package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
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
			exitCode := executeRemove(reg, pm, args)
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
	var agentName, profileName string
	if len(args) == 1 {
		profileName = args[0]
	} else if len(args) >= 2 {
		agentName = args[0]
		profileName = args[1]
	} else {
		fmt.Println("Usage: aim remove [agent] <profile>")
		return 1
	}

	if agentName != "" {
		if reg != nil {
			if adapter, err := reg.Get(agentName); err == nil {
				agentName = adapter.Name()
			}
		}
	}

	cfg, _ := config.LoadConfig()
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	okBadge := lipgloss.NewStyle().Foreground(tui.StatusGreen).Bold(true).Render("✔")
	cyan := lipgloss.NewStyle().Foreground(tui.AccentCyan)
	bold := lipgloss.NewStyle().Bold(true).Foreground(tui.TextBright)

	if agentName != "" {
		cleanedUp, err := pm.RemoveAgent(profileName, agentName, cfg)
		if err != nil {
			errBadge := lipgloss.NewStyle().Foreground(tui.StatusRed).Bold(true).Render("✖")
			if strings.Contains(err.Error(), "not associated") {
				fmt.Fprintf(os.Stderr, "%s Agent %q is not associated with profile %q\n", errBadge, agentName, profileName)
			} else {
				fmt.Fprintf(os.Stderr, "%s Error: %v\n", errBadge, err)
			}
			return 1
		}
		if cleanedUp {
			fmt.Printf("%s Removed agent %s and cleaned up profile %s\n", okBadge, bold.Render(agentName), cyan.Render(profileName))
		} else {
			fmt.Printf("%s Removed agent %s from profile %s\n", okBadge, bold.Render(agentName), cyan.Render(profileName))
		}
		return 0
	}

	if err := pm.DeleteProfile(profileName, cfg); err != nil {
		errBadge := lipgloss.NewStyle().Foreground(tui.StatusRed).Bold(true).Render("✖")
		fmt.Fprintf(os.Stderr, "%s Error: %v\n", errBadge, err)
		return 1
	}
	fmt.Printf("%s Profile %s removed successfully.\n", okBadge, cyan.Render(profileName))
	return 0
}

func runRemove(reg *agents.Registry, pm *profile.ProfileManager, args []string) error {
	if code := executeRemove(reg, pm, args); code != 0 {
		return fmt.Errorf("remove failed with exit code %d", code)
	}
	return nil
}
