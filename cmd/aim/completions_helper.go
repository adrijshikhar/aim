package main

import (
	"fmt"
	"strings"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/spf13/cobra"
)

func completeAgents(reg *agents.Registry, toComplete string) []string {
	if reg == nil {
		return nil
	}
	var res []string
	toCompleteLower := strings.ToLower(toComplete)
	for _, a := range reg.All() {
		if strings.HasPrefix(strings.ToLower(a.Name()), toCompleteLower) {
			if a.DisplayName() != "" {
				res = append(res, fmt.Sprintf("%s\t%s", a.Name(), a.DisplayName()))
			} else {
				res = append(res, a.Name())
			}
		}
	}
	return res
}

func completeProfilesForAgent(pm *profile.ProfileManager, reg *agents.Registry, agent string, toComplete string) []string {
	if pm == nil {
		return nil
	}
	if reg != nil {
		if ad, err := reg.Get(agent); err == nil {
			agent = ad.Name()
		}
	}
	cfg, _ := config.LoadConfig()
	profs, err := pm.ListProfilesForAgent(agent, cfg, reg)
	if err != nil {
		return nil
	}
	var res []string
	toCompleteLower := strings.ToLower(toComplete)
	for _, p := range profs {
		if strings.HasPrefix(strings.ToLower(p), toCompleteLower) {
			res = append(res, p)
		}
	}
	return res
}

func completeAllProfiles(pm *profile.ProfileManager, toComplete string) []string {
	if pm == nil {
		return nil
	}
	profs, err := pm.ListProfiles()
	if err != nil {
		return nil
	}
	var res []string
	toCompleteLower := strings.ToLower(toComplete)
	for _, p := range profs {
		if strings.HasPrefix(strings.ToLower(p), toCompleteLower) {
			res = append(res, p)
		}
	}
	return res
}

func completeAgentAndProfile(reg *agents.Registry, pm *profile.ProfileManager, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		return completeAgents(reg, toComplete), cobra.ShellCompDirectiveNoFileComp
	}
	if len(args) == 1 {
		return completeProfilesForAgent(pm, reg, args[0], toComplete), cobra.ShellCompDirectiveNoFileComp
	}
	return nil, cobra.ShellCompDirectiveNoFileComp
}
