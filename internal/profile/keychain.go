package profile

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/aim-cli/aim/internal/logger"
)

// IgnoredKeychainEntry defines an agent credential entry in the macOS Keychain
// that should be scrubbed and ignored to maintain strict profile sandboxing.
type IgnoredKeychainEntry struct {
	Agent       string `json:"agent"`
	Service     string `json:"service"`
	Account     string `json:"account,omitempty"`
	Description string `json:"description"`
}

// DefaultIgnoredKeychainEntries lists agent credentials known to store tokens
// in the macOS Keychain which would otherwise lead to cross-profile token leakage.
var DefaultIgnoredKeychainEntries = []IgnoredKeychainEntry{
	// Antigravity CLI / Gemini CLI
	{
		Agent:       "agy",
		Service:     "gemini",
		Account:     "antigravity",
		Description: "Antigravity OAuth token (service: gemini, account: antigravity)",
	},
	{
		Agent:       "agy",
		Service:     "gemini",
		Account:     "",
		Description: "Gemini CLI token (service: gemini)",
	},
	{
		Agent:       "agy",
		Service:     "antigravity",
		Account:     "",
		Description: "Antigravity CLI token (service: antigravity)",
	},
	{
		Agent:       "agy",
		Service:     "antigravity-oauth-token",
		Account:     "",
		Description: "Antigravity token key",
	},

	// Claude Code CLI
	{
		Agent:       "claude",
		Service:     "claude",
		Account:     "",
		Description: "Claude CLI token",
	},
	{
		Agent:       "claude",
		Service:     "claude-code",
		Account:     "",
		Description: "Claude Code CLI token",
	},
	{
		Agent:       "claude",
		Service:     "@anthropic-ai/claude-code",
		Account:     "",
		Description: "Anthropic Claude Code token",
	},
	{
		Agent:       "claude",
		Service:     "Claude Code-credentials",
		Account:     "",
		Description: "Claude Code credentials",
	},

	// OpenAI Codex CLI
	{
		Agent:       "codex",
		Service:     "codex",
		Account:     "",
		Description: "Codex CLI token",
	},
	{
		Agent:       "codex",
		Service:     "openai-codex",
		Account:     "",
		Description: "OpenAI Codex CLI token",
	},
	{
		Agent:       "codex",
		Service:     "openai",
		Account:     "",
		Description: "OpenAI CLI token",
	},
}

// GetIgnoredKeychainEntries returns the full list of ignored keychain entries,
// combining default agent entries with any custom services specified in config.
func GetIgnoredKeychainEntries(customServices ...string) []IgnoredKeychainEntry {
	entries := make([]IgnoredKeychainEntry, len(DefaultIgnoredKeychainEntries))
	copy(entries, DefaultIgnoredKeychainEntries)
	for _, custom := range customServices {
		clean := strings.TrimSpace(custom)
		if clean != "" {
			entries = append(entries, IgnoredKeychainEntry{
				Agent:       "custom",
				Service:     clean,
				Description: fmt.Sprintf("Custom ignored keychain service: %s", clean),
			})
		}
	}
	return entries
}

// PurgeIgnoredKeychains scrubs agent credentials from the macOS Keychain for a given agent
// (or all agents if agent is empty) so that profile isolation is preserved even when
// Library/Keychains is mounted into the profile.
func PurgeIgnoredKeychains(agent string, customServices ...string) error {
	if runtime.GOOS != "darwin" {
		return nil
	}

	entries := GetIgnoredKeychainEntries(customServices...)
	var errs []error

	for _, entry := range entries {
		if agent != "" && entry.Agent != "" && entry.Agent != agent && entry.Agent != "custom" {
			continue
		}

		if err := deleteGenericPassword(entry.Service, entry.Account); err != nil {
			errs = append(errs, err)
		}
		if err := deleteInternetPassword(entry.Service, entry.Account); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func deleteGenericPassword(service, account string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []string{"delete-generic-password", "-s", service}
	if account != "" {
		args = append(args, "-a", account)
	}

	cmd := exec.CommandContext(ctx, "security", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Exit code 44 means item not found in keychain, which is completely expected
			if exitErr.ExitCode() == 44 {
				return nil
			}
		}
		logger.Debug("[keychain] security delete-generic-password error: %v (output: %s)", err, strings.TrimSpace(string(out)))
		return fmt.Errorf("security delete-generic-password: %w", err)
	}
	logger.Debug("[keychain] Successfully purged generic password from macOS Keychain: service=%q, account=%q", service, account)
	return nil
}

func deleteInternetPassword(service, account string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []string{"delete-internet-password", "-s", service}
	if account != "" {
		args = append(args, "-a", account)
	}

	cmd := exec.CommandContext(ctx, "security", args...)
	_, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 44 {
				return nil
			}
		}
		return nil
	}
	logger.Debug("[keychain] Successfully purged internet password from macOS Keychain: service=%q, account=%q", service, account)
	return nil
}

// FindIgnoredKeychains returns any ignored keychain entries that currently exist in the macOS Keychain.
func FindIgnoredKeychains(agent string, customServices ...string) []IgnoredKeychainEntry {
	if runtime.GOOS != "darwin" {
		return nil
	}

	var found []IgnoredKeychainEntry
	entries := GetIgnoredKeychainEntries(customServices...)

	for _, entry := range entries {
		if agent != "" && entry.Agent != "" && entry.Agent != agent && entry.Agent != "custom" {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		args := []string{"find-generic-password", "-s", entry.Service}
		if entry.Account != "" {
			args = append(args, "-a", entry.Account)
		}
		cmd := exec.CommandContext(ctx, "security", args...)
		foundItem := (cmd.Run() == nil)
		if !foundItem {
			args[0] = "find-internet-password"
			cmd2 := exec.CommandContext(ctx, "security", args...)
			foundItem = (cmd2.Run() == nil)
		}
		if foundItem {
			found = append(found, entry)
		}
		cancel()
	}
	return found
}
