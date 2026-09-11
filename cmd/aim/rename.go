package main

import (
	"fmt"
	"os"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/usage"
	"github.com/spf13/cobra"
)

func newRenameCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	return &cobra.Command{
		Use:     "rename [agent] <old-profile> <new-profile>",
		Aliases: []string{"mv"},
		Short:   "Rename a profile preserving credentials, tokens, and state",
		Args:    cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			exitCode := executeRename(reg, pm, args)
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
				if reg != nil {
					if _, err := reg.Get(args[0]); err == nil {
						return completeProfilesForAgent(pm, reg, args[0], toComplete), cobra.ShellCompDirectiveNoFileComp
					}
				}
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}
}

func executeRename(reg *agents.Registry, pm *profile.ProfileManager, args []string) int {
	var agentName, oldProfile, newProfile string
	if len(args) == 2 {
		oldProfile = args[0]
		newProfile = args[1]
	} else if len(args) == 3 {
		agentName = args[0]
		oldProfile = args[1]
		newProfile = args[2]
	} else {
		fmt.Println("Usage: aim rename [agent] <old-profile> <new-profile>")
		return 1
	}

	if pm == nil {
		fmt.Fprintf(os.Stderr, "Error: profile manager not initialized\n")
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

	if agentName != "" {
		if !cfg.HasAgent(oldProfile, agentName) {
			fmt.Fprintf(os.Stderr, "Agent %q is not associated with profile %q\n", agentName, oldProfile)
			return 1
		}
	}

	err := pm.RenameProfile(oldProfile, newProfile, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Update cached usage reports for the renamed profile
	baseDir := config.BaseDir()
	if pm != nil && pm.BaseDir != "" {
		baseDir = pm.BaseDir
	}
	cache := usage.NewCacheStore(baseDir, usage.DefaultTTL)
	cache.Rename(oldProfile, newProfile)

	fmt.Printf("Renamed profile '%s' to '%s' successfully.\n", oldProfile, newProfile)
	return 0
}

func runRename(reg *agents.Registry, pm *profile.ProfileManager, args []string) error {
	if code := executeRename(reg, pm, args); code != 0 {
		return fmt.Errorf("rename failed with exit code %d", code)
	}
	return nil
}
