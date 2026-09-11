package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/runner"
	"github.com/spf13/cobra"
)

func newShellCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	return &cobra.Command{
		Use:   "shell <agent> <profile>",
		Short: "Launch subshell with profile environment",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			exitCode := executeShell(reg, pm, args[0], args[1])
			if exitCode != 0 {
				return &ExitError{Code: exitCode}
			}
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completeAgentAndProfile(reg, pm, args, toComplete)
		},
	}
}

func executeShell(reg *agents.Registry, pm *profile.ProfileManager, agentName, profileName string) int {
	adapter, err := reg.Get(agentName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	pDir, err := pm.EnsureProfile(profileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error preparing profile: %v\n", err)
		return 1
	}
	launchEnv, err := adapter.PrepareEnv(profileName, pDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error preparing launch environment: %v\n", err)
		return 1
	}

	// Apply profile configuration environment overrides
	cfg, _ := config.LoadConfig()
	if cfg != nil {
		if profEnv := cfg.GetProfileEnv(profileName); len(profEnv) > 0 {
			if launchEnv.Env == nil {
				launchEnv.Env = make(map[string]string)
			}
			for k, v := range profEnv {
				launchEnv.Env[k] = v
			}
		}
	}

	r := runner.NewRunner()
	code, err := r.RunShell(context.Background(), launchEnv)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Shell execution error: %v\n", err)
		return 1
	}
	return code
}
