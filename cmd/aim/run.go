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

func newRunCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:                "run <agent> <profile> [flags...] [-- args...]",
		Short:              "Execute agent under isolated profile",
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
				return cmd.Help()
			}
			if len(args) < 2 {
				return fmt.Errorf("run requires <agent> and <profile>")
			}
			agentName := args[0]
			profileName := args[1]
			var extraArgs []string
			if len(args) > 2 {
				extraArgs = args[2:]
				if len(extraArgs) > 0 && extraArgs[0] == "--" {
					extraArgs = extraArgs[1:]
				}
			}
			exitCode := executeRun(reg, pm, agentName, profileName, extraArgs)
			if exitCode != 0 {
				return &ExitError{Code: exitCode}
			}
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completeAgentAndProfile(reg, pm, args, toComplete)
		},
	}
	return cmd
}

func executeRun(reg *agents.Registry, pm *profile.ProfileManager, agentName, profileName string, extraArgs []string) int {
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

	cfg, _ := config.LoadConfig()
	if cfg != nil {
		cfg.AddProfileAgent(profileName, adapter.Name())
		_ = config.SaveConfig(cfg)
	}

	launchEnv, err := adapter.PrepareEnv(profileName, pDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error preparing launch environment: %v\n", err)
		return 1
	}

	// Apply profile configuration overrides (custom environment variables & launch arguments)
	if cfg != nil {
		if profEnv := cfg.GetProfileEnv(profileName); len(profEnv) > 0 {
			if launchEnv.Env == nil {
				launchEnv.Env = make(map[string]string)
			}
			for k, v := range profEnv {
				launchEnv.Env[k] = v
			}
		}
		if profArgs := cfg.GetProfileArgs(profileName); len(profArgs) > 0 {
			launchEnv.Args = append(launchEnv.Args, profArgs...)
		}
	}

	r := runner.NewRunner()
	code, err := r.Run(context.Background(), launchEnv, extraArgs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Execution error: %v\n", err)
		return 1
	}

	// Trigger asynchronous cache pre-warming upon session exit so quotas reflect recent usage
	triggerPrewarmAsync(config.BaseDir(), adapter.Name())

	return code
}
