package profile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/aim-cli/aim/internal/logger"
)

var defaultBridgedPaths = []string{
	// Git & Forge CLIs
	".gitconfig",
	".git-credentials",
	".config/git",
	".config/gh",
	".config/glab",

	// Security, Keys & Auth
	".ssh",
	".gnupg",
	".netrc",

	// Shell Profiles & Configurations
	".zshrc",
	".bashrc",
	".bash_profile",
	".profile",
	".config/fish",

	// Package Managers & Runtimes
	".npmrc",
	".yarnrc",
	".yarnrc.yml",
	".pip/pip.conf",
	".cargo/config.toml",
	".cargo/credentials.toml",

	// Containers & Cloud Tooling
	".docker",
	".aws",
	".config/gcloud",
	".kube",
}

// bridgedDotfiles provides backwards compatibility with existing references.
var bridgedDotfiles = defaultBridgedPaths

// GetBridgedPaths returns the full list of paths to bridge from the host home directory.
// On macOS (darwin), it mounts ~/Library/Keychains so native tools like GitHub CLI (gh)
// and git-credential-osxkeychain can access system keychains.
func GetBridgedPaths(extraPaths ...string) []string {
	paths := make([]string, 0, len(defaultBridgedPaths)+1+len(extraPaths))
	paths = append(paths, defaultBridgedPaths...)
	if runtime.GOOS == "darwin" {
		paths = append(paths, filepath.Join("Library", "Keychains"))
	}
	paths = append(paths, extraPaths...)
	return paths
}

// isAllowedBridgedPath validates that the path does not escape the profile
// directory and does not bridge internal AIM or agent credential stores.
func isAllowedBridgedPath(name string) bool {
	clean := filepath.Clean(filepath.FromSlash(name))
	if !filepath.IsLocal(clean) {
		return false
	}
	// Deny AIM internal management state
	if clean == ".aim" || strings.HasPrefix(clean, ".aim"+string(filepath.Separator)) {
		return false
	}
	// Deny agent token storage directories that AIM isolates per-profile
	if clean == ".gemini" || strings.HasPrefix(clean, ".gemini"+string(filepath.Separator)) {
		return false
	}
	if clean == ".claude" || strings.HasPrefix(clean, ".claude"+string(filepath.Separator)) || clean == ".claude.json" {
		return false
	}
	if clean == ".codex" || strings.HasPrefix(clean, ".codex"+string(filepath.Separator)) {
		return false
	}
	return true
}

// EnsureDotfiles links standard developer configurations, tools, and keychains
// from the host home directory into an isolated profile directory so developers keep
// their global developer configs without losing agent isolation.
func EnsureDotfiles(realHome, profileDir string, extraPaths ...string) error {
	if _, err := os.Stat(profileDir); err != nil {
		return fmt.Errorf("profile directory does not exist: %w", err)
	}

	logger.Debug("[symlink] Ensuring dotfiles for profile at %s (host: %s)", profileDir, realHome)
	paths := GetBridgedPaths(extraPaths...)
	var errs []error
	for _, name := range paths {
		if !isAllowedBridgedPath(name) {
			logger.Debug("[symlink] Skipping disallowed path: %s", name)
			continue
		}
		cleanName := filepath.Clean(filepath.FromSlash(name))
		src := filepath.Join(realHome, cleanName)
		if _, err := os.Lstat(src); err != nil {
			continue
		}
		dest := filepath.Join(profileDir, cleanName)
		if _, err := os.Lstat(dest); err == nil {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			errs = append(errs, err)
			continue
		}
		if err := os.Symlink(src, dest); err != nil {
			logger.Debug("[symlink] Failed to bridge %s -> %s: %v", src, dest, err)
			errs = append(errs, err)
		} else {
			logger.Debug("[symlink] Bridged: %s -> %s", cleanName, src)
		}
	}

	return errors.Join(errs...)
}
