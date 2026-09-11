package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/usage"
	"github.com/spf13/cobra"
)

func newListCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	return &cobra.Command{
		Use:   "list [agent]",
		Short: "List all profiles and status",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			agent := ""
			if len(args) > 0 {
				agent = args[0]
			}
			return runList(reg, pm, agent)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return completeAgents(reg, toComplete), cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}
}

func formatProfileUsageBadge(cache *usage.CacheStore, agent, profileName string) string {
	if cache == nil {
		return ""
	}
	rep, ok := cache.Get(agent, profileName)
	if !ok {
		return ""
	}
	if rep.Error != "" || rep.Status == usage.StatusUnknown {
		errLower := strings.ToLower(rep.Error)
		summaryLower := strings.ToLower(rep.Summary)
		if strings.Contains(errLower, "credential") || strings.Contains(summaryLower, "credential") {
			return "[no credentials]"
		}
		if strings.Contains(errLower, "offline") || strings.Contains(summaryLower, "offline") ||
			strings.Contains(errLower, "connect") || strings.Contains(errLower, "network") ||
			strings.Contains(errLower, "timeout") {
			return "[offline]"
		}
		if rep.Error != "" {
			return fmt.Sprintf("[%s]", strings.ToLower(rep.Error))
		}
		return "[unknown]"
	}

	var parts []string
	pw := rep.PrimaryWindow()
	if pw != nil {
		if s := usage.FormatWindowSummary(pw); s != "" {
			parts = append(parts, s)
		}
	}
	ww := rep.WeeklyWindow()
	if ww != nil && (pw == nil || ww.Name != pw.Name) {
		if s := usage.FormatWindowSummary(ww); s != "" {
			parts = append(parts, s)
		}
	}
	if len(parts) > 0 {
		return strings.Join(parts, ", ")
	}
	if len(rep.Windows) > 0 {
		for i := range rep.Windows {
			if s := usage.FormatWindowSummary(&rep.Windows[i]); s != "" {
				parts = append(parts, s)
			}
		}
		if len(parts) > 0 {
			return strings.Join(parts, ", ")
		}
	}
	return ""
}

func runList(reg *agents.Registry, pm *profile.ProfileManager, agentName string) error {
	if pm == nil {
		return nil
	}
	cfg, _ := config.LoadConfig()
	baseDir := config.BaseDir()
	if pm != nil && pm.BaseDir != "" {
		baseDir = pm.BaseDir
	}
	cache := usage.NewCacheStore(baseDir, usage.DefaultTTL)

	if agentName != "" {
		canonicalAgent := agentName
		if reg != nil {
			if ad, err := reg.Get(agentName); err == nil {
				canonicalAgent = ad.Name()
			}
		}
		profiles, err := pm.ListProfilesForAgent(canonicalAgent, cfg, reg)
		if err != nil {
			return err
		}
		fmt.Printf("=== Configured Profiles (%s) ===\n", canonicalAgent)
		if len(profiles) == 0 {
			fmt.Printf("  No profiles found for agent %q. Run: aim login %s <profile>\n", canonicalAgent, canonicalAgent)
			return nil
		}
		hasStale := false
		for i, p := range profiles {
			if cache.IsStale(canonicalAgent, p, 10*time.Minute) {
				hasStale = true
			}
			badgeStr := ""
			if summary := formatProfileUsageBadge(cache, canonicalAgent, p); summary != "" {
				if strings.HasPrefix(summary, "[") && strings.HasSuffix(summary, "]") {
					badgeStr = fmt.Sprintf(" %s", summary)
				} else {
					badgeStr = fmt.Sprintf(" (%s)", summary)
				}
			}
			fmt.Printf("  %d. %s%s (%s)\n", i+1, p, badgeStr, pm.ProfileDir(p))
		}
		if hasStale {
			triggerPrewarmAsync(baseDir, canonicalAgent)
		}
		return nil
	}

	// List all profiles with attached agent tags
	profiles, err := pm.ListProfiles()
	if err != nil {
		return err
	}
	fmt.Println("=== Configured Profiles ===")
	if len(profiles) == 0 {
		fmt.Println("  No profiles found. Run: aim login agy <profile>")
		return nil
	}
	hasStale := false
	for i, p := range profiles {
		agentsList := cfg.GetProfileAgents(p)
		agentStr := ""
		if len(agentsList) > 0 {
			agentStr = fmt.Sprintf(" [%s]", strings.Join(agentsList, ", "))
		}

		var summaries []string
		if len(agentsList) > 0 {
			for _, ag := range agentsList {
				if cache.IsStale(ag, p, 10*time.Minute) {
					hasStale = true
				}
				if s := formatProfileUsageBadge(cache, ag, p); s != "" {
					if len(agentsList) > 1 {
						summaries = append(summaries, fmt.Sprintf("%s: %s", ag, s))
					} else {
						summaries = append(summaries, s)
					}
				}
			}
		} else if reg != nil {
			for _, ad := range reg.All() {
				if cache.IsStale(ad.Name(), p, 10*time.Minute) {
					hasStale = true
				}
				if s := formatProfileUsageBadge(cache, ad.Name(), p); s != "" {
					summaries = append(summaries, s)
				}
			}
		}
		badgeStr := ""
		if len(summaries) == 1 && strings.HasPrefix(summaries[0], "[") && strings.HasSuffix(summaries[0], "]") {
			badgeStr = fmt.Sprintf(" %s", summaries[0])
		} else if len(summaries) > 0 {
			badgeStr = fmt.Sprintf(" (%s)", strings.Join(summaries, ", "))
		}

		fmt.Printf("  %d. %s%s%s (%s)\n", i+1, p, agentStr, badgeStr, pm.ProfileDir(p))
	}
	if hasStale {
		triggerPrewarmAsync(baseDir, "")
	}
	return nil
}
