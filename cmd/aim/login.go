package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/logger"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/spf13/cobra"
)

func newLoginCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	return &cobra.Command{
		Use:   "login <agent> <profile>",
		Short: "Authenticate new account via OAuth PKCE",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			exitCode := executeLogin(reg, pm, args[0], args[1])
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

func executeLogin(reg *agents.Registry, pm *profile.ProfileManager, agentName, profileName string) int {
	logger.Debug("[login] Starting login for agent=%q, profile=%q", agentName, profileName)
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

	cfg, err := config.LoadConfig()
	if err != nil || cfg == nil {
		cfg = config.NewDefaultConfig()
	}



	// Harvest any agent credentials into the profile directory after login completes.
	// Never purge host keychains, as doing so breaks host tools (CodexBar, host CLIs) and triggers security prompts.
	defer func() {
		_ = profile.HarvestKeychainTokenToProfile(agentName, pDir)
	}()

	if err := adapter.Login(context.Background(), profileName, pDir); err != nil {
		fmt.Fprintf(os.Stderr, "Login failed: %v\n", err)
		return 1
	}

	cfg.AddProfileAgent(profileName, adapter.Name())
	_ = config.SaveConfig(cfg)
	return 0
}
