package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/session"
	"github.com/spf13/cobra"
)

func completeAgents(reg *agents.Registry, toComplete string) []string {
	if reg == nil {
		reg = defaultRegistry()
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

func completeSessionsForAgent(reg *agents.Registry, agent string, toComplete string) []string {
	if reg != nil {
		if ad, err := reg.Get(agent); err == nil {
			agent = ad.Name()
		}
	}
	mgr := defaultSessionManager()
	if mgr == nil {
		return nil
	}
	sessions, err := mgr.ListSessions(context.Background(), agent, "", false)
	if err != nil {
		return nil
	}
	var res []string
	toCompleteLower := strings.ToLower(toComplete)
	for _, s := range sessions {
		shortID := s.ShortID
		if shortID == "" && s.ID != "" {
			shortID = session.ComputeShortID(s.ID)
		}
		if shortID == "" {
			continue
		}
		if toComplete != "" {
			if !strings.HasPrefix(strings.ToLower(shortID), toCompleteLower) &&
				!strings.HasPrefix(strings.ToLower(s.ID), toCompleteLower) {
				continue
			}
		}
		var tag string
		if s.Title != "" {
			tag = fmt.Sprintf("%s\t%s [%s]", shortID, s.Title, s.Profile)
		} else {
			tag = fmt.Sprintf("%s\t[%s]", shortID, s.Profile)
		}
		res = append(res, tag)
	}
	return res
}

func completeAgentProfileAndSession(reg *agents.Registry, pm *profile.ProfileManager, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		return completeAgents(reg, toComplete), cobra.ShellCompDirectiveNoFileComp
	}
	if len(args) == 1 {
		return completeProfilesForAgent(pm, reg, args[0], toComplete), cobra.ShellCompDirectiveNoFileComp
	}
	if len(args) == 2 {
		return completeSessionsForAgent(reg, args[0], toComplete), cobra.ShellCompDirectiveNoFileComp
	}
	return nil, cobra.ShellCompDirectiveNoFileComp
}
