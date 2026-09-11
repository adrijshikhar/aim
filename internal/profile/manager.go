package profile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/otiai10/copy"
)

type ProfileManager struct {
	BaseDir string
}

func NewProfileManager(baseDir string) *ProfileManager {
	return &ProfileManager{BaseDir: baseDir}
}

func validateProfileName(name string) error {
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}
	if name == "." || name == ".." {
		return fmt.Errorf("invalid profile name %q", name)
	}
	if strings.ContainsAny(name, "/\\") || filepath.Base(name) != name {
		return fmt.Errorf("invalid profile name %q: path traversal not allowed", name)
	}
	return nil
}

func (m *ProfileManager) ProfilesRoot() string {
	return filepath.Join(m.BaseDir, "profiles")
}

func (m *ProfileManager) ProfileDir(name string) string {
	return filepath.Join(m.ProfilesRoot(), name)
}

func (m *ProfileManager) EnsureProfile(name string) (string, error) {
	if err := validateProfileName(name); err != nil {
		return "", err
	}
	pDir := m.ProfileDir(name)
	if err := os.MkdirAll(pDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create profile directory: %w", err)
	}
	realHome := config.RealHomeDir()
	var extraPaths []string
	cfg, _ := config.LoadConfig()
	if cfg != nil && len(cfg.CustomBridgedPaths) > 0 {
		extraPaths = cfg.CustomBridgedPaths
	}
	_ = EnsureDotfiles(realHome, pDir, extraPaths...)
	return pDir, nil
}

// EnsureAllProfilesDotfiles ensures that all existing profiles on disk have up-to-date
// bridged developer configurations, tools, and keychains.
func (m *ProfileManager) EnsureAllProfilesDotfiles() error {
	profiles, err := m.ListProfiles()
	if err != nil {
		return err
	}
	realHome := config.RealHomeDir()
	var extraPaths []string
	cfg, _ := config.LoadConfig()
	if cfg != nil && len(cfg.CustomBridgedPaths) > 0 {
		extraPaths = cfg.CustomBridgedPaths
	}
	var errs []error
	for _, p := range profiles {
		pDir := m.ProfileDir(p)
		if err := EnsureDotfiles(realHome, pDir, extraPaths...); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (m *ProfileManager) ListProfiles() ([]string, error) {
	root := m.ProfilesRoot()
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	var res []string
	for _, e := range entries {
		if e.IsDir() {
			res = append(res, e.Name())
		}
	}
	sort.Strings(res)
	return res, nil
}

func (m *ProfileManager) RemoveProfile(name string) error {
	if err := validateProfileName(name); err != nil {
		return err
	}
	pDir := m.ProfileDir(name)
	if _, err := os.Stat(pDir); os.IsNotExist(err) {
		return fmt.Errorf("profile '%s' does not exist", name)
	}
	return os.RemoveAll(pDir)
}

// DeleteProfile removes the profile directory on disk and cleans up its entry in config.
func (m *ProfileManager) DeleteProfile(profileName string, cfg *config.Config) error {
	if err := validateProfileName(profileName); err != nil {
		return err
	}
	if cfg == nil {
		var err error
		cfg, err = config.LoadConfig()
		if err != nil {
			cfg = config.NewDefaultConfig()
		}
	}
	pDir := m.ProfileDir(profileName)
	dirExists := false
	if _, err := os.Stat(pDir); err == nil {
		dirExists = true
	}
	cfgExists := false
	if cfg.Profiles != nil {
		_, cfgExists = cfg.Profiles[profileName]
	}
	if !dirExists && !cfgExists {
		return fmt.Errorf("profile '%s' does not exist", profileName)
	}

	if dirExists {
		if err := m.RemoveProfile(profileName); err != nil {
			return err
		}
	}
	cfg.DeleteProfile(profileName)
	return config.SaveConfig(cfg)
}

// RemoveAgent unlinks an agent from a profile in config.
// If other agents remain associated with the profile, the directory is preserved and cleanedUp is false.
// If no agents remain, the entire profile directory is deleted and cleanedUp is true.
func (m *ProfileManager) RemoveAgent(profileName, agentName string, cfg *config.Config) (bool, error) {
	if err := validateProfileName(profileName); err != nil {
		return false, err
	}
	if cfg == nil {
		var err error
		cfg, err = config.LoadConfig()
		if err != nil {
			cfg = config.NewDefaultConfig()
		}
	}
	pDir := m.ProfileDir(profileName)
	dirExists := false
	if _, err := os.Stat(pDir); err == nil {
		dirExists = true
	}
	cfgExists := false
	if cfg.Profiles != nil {
		_, cfgExists = cfg.Profiles[profileName]
	}
	if !dirExists && !cfgExists {
		return false, fmt.Errorf("profile '%s' does not exist", profileName)
	}

	if !cfg.HasAgent(profileName, agentName) {
		return false, fmt.Errorf("agent %q is not associated with profile %q", agentName, profileName)
	}

	cfg.RemoveProfileAgent(profileName, agentName)
	remaining := cfg.GetProfileAgents(profileName)
	if len(remaining) > 0 {
		return false, config.SaveConfig(cfg)
	}

	if dirExists {
		if err := m.RemoveProfile(profileName); err != nil {
			return false, err
		}
	}
	cfg.DeleteProfile(profileName)
	return true, config.SaveConfig(cfg)
}

// ListProfilesForAgent returns profile names associated with the given agent.
// Resolution logic:
// 1. If agentName is empty or "all", returns all profiles.
// 2. Checks if profile has agentName in cfg.Profiles[p].Agents.
// 3. Falls back to detecting credentials via adapter.HasCredentials(profileDir).
// 4. Returns sorted, deduplicated slice of matching profile names.
func (m *ProfileManager) ListProfilesForAgent(agentName string, cfg *config.Config, reg *agents.Registry) ([]string, error) {
	if m == nil {
		return nil, nil
	}
	all, err := m.ListProfiles()
	if err != nil {
		return nil, err
	}
	if agentName == "" || agentName == "all" {
		return all, nil
	}

	var adapter agents.AgentAdapter
	if reg != nil {
		adapter, _ = reg.Get(agentName)
	}

	matched := make([]string, 0)
	for _, p := range all {
		if cfg != nil && cfg.HasAgent(p, agentName) {
			matched = append(matched, p)
			continue
		}
		if adapter != nil && adapter.HasCredentials(m.ProfileDir(p)) {
			matched = append(matched, p)
			continue
		}
	}
	sort.Strings(matched)
	return matched, nil
}

func isSensitiveProfileFile(name string) bool {
	lower := strings.ToLower(name)
	return strings.Contains(lower, "token") || strings.Contains(lower, "credential") || strings.HasSuffix(lower, ".db")
}

// CloneProfile duplicates a profile's non-sensitive configuration and dotfile symlinks to a new profile.
// Authentication tokens and session history databases are intentionally omitted.
func (m *ProfileManager) CloneProfile(sourceProfile, newProfile, agentName string, cfg *config.Config) error {
	if err := validateProfileName(sourceProfile); err != nil {
		return fmt.Errorf("invalid source profile: %w", err)
	}
	if err := validateProfileName(newProfile); err != nil {
		return fmt.Errorf("invalid new profile: %w", err)
	}
	if sourceProfile == newProfile {
		return fmt.Errorf("source and destination profile names cannot be the same")
	}

	if cfg == nil {
		var err error
		cfg, err = config.LoadConfig()
		if err != nil {
			cfg = config.NewDefaultConfig()
		}
	}

	srcDir := m.ProfileDir(sourceProfile)
	_, srcDirErr := os.Stat(srcDir)
	srcDirExists := srcDirErr == nil
	_, srcCfgExists := cfg.Profiles[sourceProfile]
	if !srcDirExists && !srcCfgExists {
		return fmt.Errorf("source profile '%s' does not exist", sourceProfile)
	}

	dstDir := m.ProfileDir(newProfile)
	_, dstDirErr := os.Stat(dstDir)
	_, dstCfgExists := cfg.Profiles[newProfile]
	if dstDirErr == nil || dstCfgExists {
		return fmt.Errorf("profile '%s' already exists", newProfile)
	}

	if agentName != "" {
		if !cfg.HasAgent(sourceProfile, agentName) {
			return fmt.Errorf("agent %q is not associated with profile %q", agentName, sourceProfile)
		}
	}

	// Create destination directory
	if err := os.MkdirAll(dstDir, 0700); err != nil {
		return fmt.Errorf("failed to create destination profile directory: %w", err)
	}

	// If source directory exists, copy non-sensitive files and symlinks
	if srcDirExists {
		opt := copy.Options{
			Skip: func(srcinfo os.FileInfo, src, dest string) (bool, error) {
				return isSensitiveProfileFile(filepath.Base(src)), nil
			},
			OnSymlink: func(src string) copy.SymlinkAction {
				return copy.Shallow
			},
			PreserveTimes: true,
			PreserveOwner: true,
		}
		_ = copy.Copy(srcDir, dstDir, opt)
	}

	// Guarantee dotfiles are linked
	realHome := config.RealHomeDir()
	var extraPaths []string
	if cfg != nil && len(cfg.CustomBridgedPaths) > 0 {
		extraPaths = cfg.CustomBridgedPaths
	}
	_ = EnsureDotfiles(realHome, dstDir, extraPaths...)

	// Update config
	srcAgents := cfg.GetProfileAgents(sourceProfile)
	var dstAgents []string
	if agentName != "" {
		dstAgents = []string{agentName}
	} else if len(srcAgents) > 0 {
		dstAgents = append([]string(nil), srcAgents...)
	}

	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]config.ProfileConfig)
	}
	srcProf := cfg.Profiles[sourceProfile]
	dstProf := config.ProfileConfig{
		Agents: dstAgents,
	}
	if len(srcProf.Env) > 0 {
		dstProf.Env = make(map[string]string, len(srcProf.Env))
		for k, v := range srcProf.Env {
			dstProf.Env[k] = v
		}
	}
	if len(srcProf.Args) > 0 {
		dstProf.Args = append([]string(nil), srcProf.Args...)
	}
	cfg.Profiles[newProfile] = dstProf

	return config.SaveConfig(cfg)
}
