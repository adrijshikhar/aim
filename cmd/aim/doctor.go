package main

import (
	"context"
	"fmt"
	"runtime"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/logger"
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
	logger.Debug("[doctor] Running diagnostics (agent=%q)", agentName)
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
		diagnosePlatform(agentName, cfg)
		return
	}

	for _, adapter := range reg.All() {
		diagnoseAdapter(adapter, pm, cfg, reg)
	}
	diagnosePlatform("", cfg)
}

func diagnosePlatform(agentName string, cfg *config.Config) {
	if runtime.GOOS != "darwin" {
		return
	}

	fmt.Println("\n[Platform Diagnostics (macOS)]")
	var customServices []string
	if cfg != nil {
		customServices = cfg.CustomIgnoredKeychains
	}

	lingering := profile.FindIgnoredKeychains(agentName, customServices...)
	if len(lingering) > 0 {
		fmt.Printf("  [WARN] Keychain: %d agent token(s) detected in macOS Keychain (potential profile isolation risk):\n", len(lingering))
		for _, entry := range lingering {
			fmt.Printf("         - %s (service: %q)\n", entry.Description, entry.Service)
		}
		fmt.Println("         Auto-purging ignored agent keychains to enforce profile isolation...")
		if err := profile.PurgeIgnoredKeychains(agentName, customServices...); err != nil {
			fmt.Printf("  [WARN] Keychain: Could not purge some entries: %v\n", err)
		} else {
			fmt.Println("  [OK]   Keychain: Agent credentials successfully purged from macOS Keychain.")
		}
	} else {
		fmt.Println("  [OK]   Keychain: No lingering agent tokens in macOS Keychain (clean isolation).")
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
