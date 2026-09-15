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

func newCloneCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	return &cobra.Command{
		Use:   "clone [agent] <src> <dst>",
		Short: "Duplicate profile settings without copying tokens",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			exitCode := executeClone(reg, pm, args)
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

func executeClone(reg *agents.Registry, pm *profile.ProfileManager, args []string) int {
	var agentName, sourceProfile, newProfile string
	if len(args) == 2 {
		sourceProfile = args[0]
		newProfile = args[1]
	} else if len(args) == 3 {
		agentName = args[0]
		sourceProfile = args[1]
		newProfile = args[2]
	} else {
		fmt.Println("Usage: aim clone [agent] <source-profile> <new-profile>")
		return 1
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

	err := pm.CloneProfile(sourceProfile, newProfile, agentName, cfg)
	if err != nil {
		errBadge := lipgloss.NewStyle().Foreground(tui.StatusRed).Bold(true).Render("✖")
		if strings.Contains(err.Error(), "not associated") {
			fmt.Fprintf(os.Stderr, "%s Agent %q is not associated with profile %q\n", errBadge, agentName, sourceProfile)
		} else {
			fmt.Fprintf(os.Stderr, "%s Error: %v\n", errBadge, err)
		}
		return 1
	}

	okBadge := lipgloss.NewStyle().Foreground(tui.StatusGreen).Bold(true).Render("✔")
	cyan := lipgloss.NewStyle().Foreground(tui.AccentCyan)
	bold := lipgloss.NewStyle().Bold(true).Foreground(tui.TextBright)

	if agentName != "" {
		fmt.Printf("%s Cloned profile %s to %s for agent %s.\n", okBadge, cyan.Render(sourceProfile), cyan.Render(newProfile), bold.Render(agentName))
	} else {
		fmt.Printf("%s Cloned profile %s to %s successfully.\n", okBadge, cyan.Render(sourceProfile), cyan.Render(newProfile))
	}
	return 0
}

func runClone(reg *agents.Registry, pm *profile.ProfileManager, args []string) error {
	if code := executeClone(reg, pm, args); code != 0 {
		return fmt.Errorf("clone failed with exit code %d", code)
	}
	return nil
}
