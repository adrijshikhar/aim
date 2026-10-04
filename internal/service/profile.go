package service

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/usage"
)

type profileService struct {
	pm    *profile.ProfileManager
	reg   *agents.Registry
	cache *usage.CacheStore
}

// NewProfileService creates a new ProfileService wrapping ProfileManager, agents.Registry, and usage.CacheStore.
func NewProfileService(pm *profile.ProfileManager, reg *agents.Registry, cache ...*usage.CacheStore) ProfileService {
	var c *usage.CacheStore
	if len(cache) > 0 && cache[0] != nil {
		c = cache[0]
	} else if pm != nil && pm.BaseDir != "" {
		c = usage.NewCacheStore(pm.BaseDir, usage.DefaultTTL)
	}
	return &profileService{
		pm:    pm,
		reg:   reg,
		cache: c,
	}
}

// ListProfiles returns all profiles for the specified agent, or all profiles if agent is empty or "all".
func (s *profileService) ListProfiles(ctx context.Context, agent string) ([]ProfileDTO, error) {
	if s.pm == nil {
		return []ProfileDTO{}, nil
	}

	cfg, err := config.LoadConfig()
	if err != nil || cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	agent = strings.TrimSpace(agent)
	if agent != "" && agent != "all" {
		canonicalAgent := agent
		var adapter agents.AgentAdapter
		if s.reg != nil {
			if ad, err := s.reg.Get(agent); err == nil {
				canonicalAgent = ad.Name()
				adapter = ad
			}
		}

		profs, err := s.pm.ListProfilesForAgent(canonicalAgent, cfg, s.reg)
		if err != nil {
			return nil, err
		}

		result := make([]ProfileDTO, 0, len(profs))
		for _, p := range profs {
			dto := s.buildProfileDTO(canonicalAgent, p, adapter, cfg)
			result = append(result, dto)
		}
		return result, nil
	}

	// Listing all profiles across all agents
	allProfs, err := s.pm.ListProfiles()
	if err != nil {
		return nil, err
	}

	result := make([]ProfileDTO, 0, len(allProfs))
	for _, p := range allProfs {
		pDir := s.pm.ProfileDir(p)
		agentsList := cfg.GetProfileAgents(p)
		if len(agentsList) > 0 {
			for _, ag := range agentsList {
				var adapter agents.AgentAdapter
				if s.reg != nil {
					adapter, _ = s.reg.Get(ag)
				}
				result = append(result, s.buildProfileDTO(ag, p, adapter, cfg))
			}
		} else {
			var matchedAdapters []agents.AgentAdapter
			if s.reg != nil {
				for _, ad := range s.reg.All() {
					if ad.HasCredentials(pDir) {
						matchedAdapters = append(matchedAdapters, ad)
					}
				}
			}
			if len(matchedAdapters) > 0 {
				for _, ad := range matchedAdapters {
					result = append(result, s.buildProfileDTO(ad.Name(), p, ad, cfg))
				}
			} else {
				result = append(result, s.buildProfileDTO("", p, nil, cfg))
			}
		}
	}
	return result, nil
}

func (s *profileService) buildProfileDTO(agentName, profileName string, adapter agents.AgentAdapter, cfg *config.Config) ProfileDTO {
	pDir := s.pm.ProfileDir(profileName)
	hasCreds := false
	if adapter != nil {
		hasCreds = adapter.HasCredentials(pDir)
	} else if s.reg != nil && agentName != "" {
		if ad, err := s.reg.Get(agentName); err == nil {
			hasCreds = ad.HasCredentials(pDir)
		}
	}

	acc := profile.GetProfileAccountInfoForAgent(pDir, agentName)
	var accountPtr *profile.AccountInfo
	if acc.Email != "" || acc.Name != "" || acc.AuthMethod != "" || acc.ProjectID != "" {
		accountPtr = &acc
	}

	var quotaPtr *QuotaDTO
	if s.cache != nil && agentName != "" {
		if rep, ok := s.cache.Get(agentName, profileName); ok {
			quotaPtr = reportToQuotaDTO(rep)
		} else if rep, ok := s.cache.GetStale(agentName, profileName); ok {
			quotaPtr = reportToQuotaDTO(rep)
		}
	}

	var mcpGlobal, pluginsGlobal *bool
	if cfg != nil && cfg.Profiles != nil {
		if p, ok := cfg.Profiles[profileName]; ok {
			mcpGlobal = p.MCPGlobal
			pluginsGlobal = p.PluginsGlobal
		}
	}

	return ProfileDTO{
		Agent:          agentName,
		Name:           profileName,
		Path:           pDir,
		HasCredentials: hasCreds,
		Account:        accountPtr,
		Quota:          quotaPtr,
		MCPGlobal:      mcpGlobal,
		PluginsGlobal:  pluginsGlobal,
	}
}

func reportToQuotaDTO(rep usage.Report) *QuotaDTO {
	isExhausted := rep.Status == usage.StatusExhausted || (len(rep.Windows) > 0 && rep.BottleneckPct() <= 0)
	summary := rep.Summary
	if summary == "" && len(rep.Windows) > 0 {
		var parts []string
		if pw := rep.PrimaryWindow(); pw != nil {
			if s := usage.FormatWindowSummary(pw); s != "" {
				parts = append(parts, s)
			}
		}
		if ww := rep.WeeklyWindow(); ww != nil && (rep.PrimaryWindow() == nil || ww.Name != rep.PrimaryWindow().Name) {
			if s := usage.FormatWindowSummary(ww); s != "" {
				parts = append(parts, s)
			}
		}
		summary = strings.Join(parts, ", ")
	}
	if summary == "" && rep.Error != "" {
		summary = rep.Error
	}
	return &QuotaDTO{
		BottleneckPct: rep.BottleneckPct(),
		Summary:       summary,
		IsExhausted:   isExhausted,
	}
}

// CreateProfile scaffolds a new profile or clones from an existing one.
func (s *profileService) CreateProfile(ctx context.Context, req CreateProfileRequest) (*ProfileDTO, error) {
	if s.pm == nil {
		return nil, errors.New("profile manager is not configured")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("profile name cannot be empty")
	}
	if name == "." || name == ".." || strings.ContainsAny(name, "/\\") || filepath.Base(name) != name {
		return nil, fmt.Errorf("invalid profile name %q: path traversal not allowed", name)
	}

	canonicalAgent := strings.TrimSpace(req.Agent)
	var adapter agents.AgentAdapter
	if s.reg != nil && canonicalAgent != "" {
		if ad, err := s.reg.Get(canonicalAgent); err == nil {
			canonicalAgent = ad.Name()
			adapter = ad
		}
	}

	cfg, err := config.LoadConfig()
	if err != nil || cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	if req.CloneFrom != "" {
		cloneSrc := strings.TrimSpace(req.CloneFrom)
		if err := s.pm.CloneProfile(cloneSrc, name, canonicalAgent, cfg); err != nil {
			return nil, err
		}
	} else {
		if s.pm.ProfileExists(name) {
			return nil, fmt.Errorf("profile '%s' already exists", name)
		}
		if _, err := s.pm.EnsureProfile(name); err != nil {
			return nil, err
		}
		if canonicalAgent != "" {
			cfg.AddProfileAgent(name, canonicalAgent)
			if err := config.SaveConfig(cfg); err != nil {
				return nil, err
			}
		}
	}

	dto := s.buildProfileDTO(canonicalAgent, name, adapter, cfg)
	if dto.Account == nil && req.Email != "" {
		dto.Account = &profile.AccountInfo{Email: req.Email}
	}
	return &dto, nil
}

// RemoveProfile deletes an agent's association or the entire profile.
func (s *profileService) RemoveProfile(ctx context.Context, agent, name string) error {
	if s.pm == nil {
		return errors.New("profile manager is not configured")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("profile name cannot be empty")
	}

	cfg, err := config.LoadConfig()
	if err != nil || cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	canonicalAgent := strings.TrimSpace(agent)
	if s.reg != nil && canonicalAgent != "" && canonicalAgent != "all" {
		if ad, err := s.reg.Get(canonicalAgent); err == nil {
			canonicalAgent = ad.Name()
		}
	}

	if canonicalAgent != "" && canonicalAgent != "all" {
		if cfg.HasAgent(name, canonicalAgent) {
			_, err := s.pm.RemoveAgent(name, canonicalAgent, cfg)
			if err != nil {
				return err
			}
			if s.cache != nil {
				s.cache.Delete(canonicalAgent, name)
			}
			return nil
		}

		if !s.pm.ProfileExists(name) {
			return fmt.Errorf("profile '%s' does not exist", name)
		}

		agentsList := cfg.GetProfileAgents(name)
		if len(agentsList) > 0 {
			return fmt.Errorf("agent %q is not associated with profile %q", canonicalAgent, name)
		}

		// Profile exists on disk without explicit config agents - delete it
		if err := s.pm.DeleteProfile(name, cfg); err != nil {
			return err
		}
		if s.cache != nil {
			s.cache.Delete(canonicalAgent, name)
		}
		return nil
	}

	if err := s.pm.DeleteProfile(name, cfg); err != nil {
		return err
	}
	if s.cache != nil {
		if s.reg != nil {
			for _, ad := range s.reg.All() {
				s.cache.Delete(ad.Name(), name)
			}
		}
		s.cache.Delete("", name)
	}
	return nil
}

// RenameProfile renames a profile and synchronizes config and cache.
func (s *profileService) RenameProfile(ctx context.Context, agent, oldName, newName string) error {
	if s.pm == nil {
		return errors.New("profile manager is not configured")
	}
	oldName = strings.TrimSpace(oldName)
	newName = strings.TrimSpace(newName)
	if oldName == "" {
		return errors.New("source profile name cannot be empty")
	}
	if newName == "" {
		return errors.New("target profile name cannot be empty")
	}

	cfg, err := config.LoadConfig()
	if err != nil || cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	canonicalAgent := strings.TrimSpace(agent)
	if s.reg != nil && canonicalAgent != "" && canonicalAgent != "all" {
		if ad, err := s.reg.Get(canonicalAgent); err == nil {
			canonicalAgent = ad.Name()
		}
	}

	if canonicalAgent != "" && canonicalAgent != "all" {
		if !s.pm.ProfileExists(oldName) {
			return fmt.Errorf("profile '%s' does not exist", oldName)
		}
		agentsList := cfg.GetProfileAgents(oldName)
		if len(agentsList) > 0 && !cfg.HasAgent(oldName, canonicalAgent) {
			return fmt.Errorf("agent %q is not associated with profile %q", canonicalAgent, oldName)
		}
	}

	if err := s.pm.RenameProfile(oldName, newName, cfg); err != nil {
		return err
	}

	if s.cache != nil {
		s.cache.Rename(oldName, newName)
	}
	return nil
}

// UpdateProfileConfig updates mcp_global and plugins_global configuration for a profile.
func (s *profileService) UpdateProfileConfig(ctx context.Context, name string, mcpGlobal, pluginsGlobal *bool) error {
	if s.pm == nil {
		return errors.New("profile manager is not configured")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("profile name cannot be empty")
	}

	cfg, err := config.LoadConfig()
	if err != nil || cfg == nil {
		cfg = config.NewDefaultConfig()
	}
	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]config.ProfileConfig)
	}

	p := cfg.Profiles[name]
	if mcpGlobal != nil {
		p.MCPGlobal = mcpGlobal
	}
	if pluginsGlobal != nil {
		p.PluginsGlobal = pluginsGlobal
	}
	cfg.Profiles[name] = p

	return config.SaveConfig(cfg)
}

