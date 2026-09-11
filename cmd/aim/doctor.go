package main

import (
	"context"
	"fmt"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/spf13/cobra"
)

func newDoctorCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor [agent]",
		Short: "Diagnose environment, tokens, and binaries",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			agent := ""
			if len(args) > 0 {
				agent = args[0]
			}
			runDoctor(reg, pm, agent)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return completeAgents(reg, toComplete), cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}
}

func runDoctor(reg *agents.Registry, pm *profile.ProfileManager, agentName string) {
	fmt.Println("=== AIM Doctor Diagnostics ===")
	if reg == nil {
		return
	}
	if pm != nil {
		_ = pm.EnsureAllProfilesDotfiles()
	}
	cfg, _ := config.LoadConfig()

	if agentName != "" {
		adapter, err := reg.Get(agentName)
		if err != nil {
			fmt.Printf("[FAIL] Unknown agent: %s\n", agentName)
			return
		}
		diagnoseAdapter(adapter, pm, cfg, reg)
		return
	}

	for _, adapter := range reg.All() {
		diagnoseAdapter(adapter, pm, cfg, reg)
	}
}

func diagnoseAdapter(adapter agents.AgentAdapter, pm *profile.ProfileManager, cfg *config.Config, reg *agents.Registry) {
	fmt.Printf("\n[%s (%s)]\n", adapter.DisplayName(), adapter.Name())
	profiles, _ := pm.ListProfilesForAgent(adapter.Name(), cfg, reg)
	if len(profiles) == 0 {
		fmt.Printf("  No profiles configured for agent %q. Run: aim login %s <profile>\n", adapter.Name(), adapter.Name())
		return
	}
	for _, p := range profiles {
		fmt.Printf("\nProfile: %s\n", p)
		results := adapter.Doctor(context.Background(), p, pm.ProfileDir(p))
		if cfg != nil {
			if env := cfg.GetProfileEnv(p); len(env) > 0 {
				results = append(results, agents.DiagnosticResult{
					Category: "Config",
					Status:   "OK",
					Message:  fmt.Sprintf("%d custom env var(s) configured", len(env)),
				})
			}
			if args := cfg.GetProfileArgs(p); len(args) > 0 {
				results = append(results, agents.DiagnosticResult{
					Category: "Config",
					Status:   "OK",
					Message:  fmt.Sprintf("%d custom launch arg(s) configured", len(args)),
				})
			}
		}
		for _, r := range results {
			fmt.Printf("  [%s] %s: %s\n", r.Status, r.Category, r.Message)
		}
	}
}
