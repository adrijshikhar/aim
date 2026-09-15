package profile

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/otiai10/copy"
)

// MoveAgent transfers an agent's credentials, tokens, and configuration from sourceProfile to targetProfile.
// If targetProfile already has credentials for the agent, it returns an error unless force is true.
// If sourceProfile has no remaining agents after the move, it is cleaned up.
func (m *ProfileManager) MoveAgent(agentName, sourceProfile, targetProfile string, force bool, cfg *config.Config, reg *agents.Registry) error {
	if err := validateProfileName(sourceProfile); err != nil {
		return fmt.Errorf("invalid source profile: %w", err)
	}
	if err := validateProfileName(targetProfile); err != nil {
		return fmt.Errorf("invalid target profile: %w", err)
	}
	if sourceProfile == targetProfile {
		return fmt.Errorf("source and target profile names cannot be the same")
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

	canonicalAgent := agentName
	var adapter agents.AgentAdapter
	if reg != nil {
		if ad, err := reg.Get(agentName); err == nil {
			adapter = ad
			canonicalAgent = ad.Name()
		}
	}

	hasAgent := cfg.HasAgent(sourceProfile, canonicalAgent)
	if !hasAgent && adapter != nil && srcDirExists && adapter.HasCredentials(srcDir) {
		hasAgent = true
	}
	if !hasAgent {
		return fmt.Errorf("agent %q is not associated with profile %q", canonicalAgent, sourceProfile)
	}

	// Ensure destination profile exists
	dstDir, err := m.EnsureProfile(targetProfile)
	if err != nil {
		return fmt.Errorf("failed to ensure target profile %q: %w", targetProfile, err)
	}

	// Collision check
	hasTargetCreds := false
	if adapter != nil && adapter.HasCredentials(dstDir) {
		hasTargetCreds = true
	}
	if !hasTargetCreds && hasAgentCredentialsOnDisk(canonicalAgent, dstDir) {
		hasTargetCreds = true
	}
	if hasTargetCreds && !force {
		return fmt.Errorf("target profile %q already has credentials for agent %q (use --force to overwrite)", targetProfile, canonicalAgent)
	}

	// Move agent files on disk
	if srcDirExists {
		if err := moveAgentData(canonicalAgent, srcDir, dstDir, force, adapter); err != nil {
			return fmt.Errorf("failed to move agent data: %w", err)
		}
	}

	// Guarantee dotfiles in destination
	realHome := config.RealHomeDir()
	var extraPaths []string
	if len(cfg.CustomBridgedPaths) > 0 {
		extraPaths = cfg.CustomBridgedPaths
	}
	_ = EnsureDotfiles(realHome, dstDir, extraPaths...)

	// Update config
	cfg.RemoveProfileAgent(sourceProfile, canonicalAgent)
	cfg.AddProfileAgent(targetProfile, canonicalAgent)

	remaining := cfg.GetProfileAgents(sourceProfile)
	if len(remaining) == 0 {
		_ = m.RemoveProfile(sourceProfile)
		cfg.DeleteProfile(sourceProfile)
	}

	return config.SaveConfig(cfg)
}

func hasAgentCredentialsOnDisk(agent, dir string) bool {
	switch agent {
	case "codex":
		p := filepath.Join(dir, ".codex", "auth.json")
		fi, err := os.Stat(p)
		return err == nil && !fi.IsDir() && fi.Size() > 0
	case "agy":
		p := filepath.Join(dir, ".gemini", "antigravity-cli", "antigravity-oauth-token")
		fi, err := os.Stat(p)
		if err == nil && !fi.IsDir() && fi.Size() > 0 {
			return true
		}
		adc := filepath.Join(dir, ".config", "gcloud", "application_default_credentials.json")
		fiADC, errADC := os.Stat(adc)
		return errADC == nil && !fiADC.IsDir() && fiADC.Size() > 0
	case "gemini":
		p := filepath.Join(dir, ".gemini", "gemini-oauth-token")
		fi, err := os.Stat(p)
		return err == nil && !fi.IsDir() && fi.Size() > 0
	default:
		return false
	}
}

func moveAgentData(agent, srcDir, dstDir string, force bool, adapter agents.AgentAdapter) error {
	switch agent {
	case "codex":
		src := filepath.Join(srcDir, ".codex")
		dst := filepath.Join(dstDir, ".codex")
		if fi, err := os.Stat(src); err == nil && fi.IsDir() {
			if force {
				_ = os.RemoveAll(dst)
			}
			if err := moveOrCopyDir(src, dst); err != nil {
				return err
			}
		}
	case "agy":
		src := filepath.Join(srcDir, ".gemini", "antigravity-cli")
		dst := filepath.Join(dstDir, ".gemini", "antigravity-cli")
		if fi, err := os.Stat(src); err == nil && fi.IsDir() {
			if force {
				_ = os.RemoveAll(dst)
			}
			_ = os.MkdirAll(filepath.Dir(dst), 0700)
			if err := moveOrCopyDir(src, dst); err != nil {
				return err
			}
		}
		srcGcloud := filepath.Join(srcDir, ".config", "gcloud")
		dstGcloud := filepath.Join(dstDir, ".config", "gcloud")
		if fi, err := os.Stat(srcGcloud); err == nil && fi.IsDir() {
			if force {
				_ = os.RemoveAll(dstGcloud)
			}
			_ = os.MkdirAll(filepath.Dir(dstGcloud), 0700)
			_ = moveOrCopyDir(srcGcloud, dstGcloud)
		}
	case "gemini":
		srcTok := filepath.Join(srcDir, ".gemini", "gemini-oauth-token")
		dstTok := filepath.Join(dstDir, ".gemini", "gemini-oauth-token")
		if fi, err := os.Stat(srcTok); err == nil && !fi.IsDir() {
			if force {
				_ = os.Remove(dstTok)
			}
			_ = os.MkdirAll(filepath.Dir(dstTok), 0700)
			_ = moveOrCopyFile(srcTok, dstTok)
		}
	}

	// Adapter TokenPath fallback if present
	if pather, ok := adapter.(interface{ TokenPath(string) string }); ok {
		srcTok := pather.TokenPath(srcDir)
		if fi, err := os.Stat(srcTok); err == nil && !fi.IsDir() {
			dstTok := pather.TokenPath(dstDir)
			if force {
				_ = os.Remove(dstTok)
			}
			_ = os.MkdirAll(filepath.Dir(dstTok), 0700)
			_ = moveOrCopyFile(srcTok, dstTok)
		}
	}
	return nil
}

func moveOrCopyDir(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	if err := copy.Copy(src, dst); err != nil {
		return err
	}
	return os.RemoveAll(src)
}

func moveOrCopyFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	if err := copy.Copy(src, dst); err != nil {
		return err
	}
	return os.Remove(src)
}
