package profile

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
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
		Service:     "Claude Safe Storage",
		Account:     "",
		Description: "Claude Safe Storage credentials",
	},
	{
		Agent:       "claude",
		Service:     "Claude Code-credentials",
		Account:     "",
		Description: "Claude Code credentials",
	},
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
	var (
		mu   sync.Mutex
		errs []error
		wg   sync.WaitGroup
	)

	for _, entry := range entries {
		if agent != "" && entry.Agent != "" && entry.Agent != agent && entry.Agent != "custom" {
			continue
		}

		ent := entry
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := deleteGenericPassword(ent.Service, ent.Account); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
			if err := deleteInternetPassword(ent.Service, ent.Account); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

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

var getGenericPasswordFn = getGenericPasswordReal

func getGenericPasswordReal(service, account string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	args := []string{"find-generic-password", "-s", service}
	if account != "" {
		args = append(args, "-a", account)
	}
	args = append(args, "-w")

	cmd := exec.CommandContext(ctx, "security", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// DecodeKeychainPassword parses a raw keychain password string, decoding any
// go-keyring-base64: prefixes if present and ensuring valid JSON.
func DecodeKeychainPassword(raw string) []byte {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	const prefix = "go-keyring-base64:"
	if strings.HasPrefix(raw, prefix) {
		encoded := strings.TrimPrefix(raw, prefix)
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err == nil && len(decoded) > 0 {
			var js json.RawMessage
			if json.Unmarshal(decoded, &js) == nil {
				return decoded
			}
		}
		return nil
	}
	var js json.RawMessage
	if json.Unmarshal([]byte(raw), &js) == nil {
		return []byte(raw)
	}
	return nil
}

// GetAgentKeychainToken searches the macOS Keychain for any credentials
// belonging to the specified agent and returns the decoded token JSON.
func GetAgentKeychainToken(agent string) ([]byte, error) {
	if runtime.GOOS != "darwin" && os.Getenv("AIM_MOCK_KEYCHAIN") == "" {
		return nil, errors.New("keychain is only supported on darwin")
	}

	entries := GetIgnoredKeychainEntries()
	for _, entry := range entries {
		if entry.Agent != agent {
			continue
		}
		raw, err := getGenericPasswordFn(entry.Service, entry.Account)
		if err != nil || raw == "" {
			continue
		}
		data := DecodeKeychainPassword(raw)
		if len(data) > 0 {
			logger.Debug("[keychain] Found valid keychain token for agent %q (service=%q, account=%q)", agent, entry.Service, entry.Account)
			return data, nil
		}
	}
	return nil, fmt.Errorf("no keychain credentials found for agent %q", agent)
}

// HarvestKeychainTokenToProfile extracts any existing agent credentials from the macOS
// Keychain and saves them into the isolated profile directory.
func HarvestKeychainTokenToProfile(agent, profileDir string) bool {
	if runtime.GOOS != "darwin" && os.Getenv("AIM_MOCK_KEYCHAIN") == "" {
		return false
	}
	var destTokenPath string
	switch agent {
	case "agy":
		destTokenPath = filepath.Join(profileDir, ".gemini", "antigravity-cli", "antigravity-oauth-token")
	default:
		return false
	}

	tokData, err := GetAgentKeychainToken(agent)
	if err == nil && len(tokData) > 0 {
		if err := os.MkdirAll(filepath.Dir(destTokenPath), 0700); err != nil {
			logger.Debug("[keychain] Failed to create directory for token %s: %v", destTokenPath, err)
			return false
		}
		if err := os.WriteFile(destTokenPath, tokData, 0600); err != nil {
			logger.Debug("[keychain] Failed to write token to %s: %v", destTokenPath, err)
			return false
		}
		logger.Debug("[keychain] Successfully harvested keychain token for %s into %s", agent, destTokenPath)
		return true
	}

	if fi, err := os.Stat(destTokenPath); err == nil && fi.Size() > 0 {
		return true
	}

	return false
}

var claudeKeychainServices = []string{
	"Claude Safe Storage",
	"Claude Code-credentials",
}

// KnownKeychainServices returns a list of known keychain service names.
// If an agent is specified, only services associated with that agent are returned.
// If no agent is specified, all known keychain services across all agents are returned.
func KnownKeychainServices(agent ...string) []string {
	var targetAgent string
	if len(agent) > 0 {
		targetAgent = agent[0]
	}

	var services []string
	seen := make(map[string]bool)

	for _, entry := range DefaultIgnoredKeychainEntries {
		if targetAgent != "" && entry.Agent != targetAgent {
			continue
		}
		if !seen[entry.Service] {
			seen[entry.Service] = true
			services = append(services, entry.Service)
		}
	}

	// Ensure claudeKeychainServices are present when querying for claude or all
	if targetAgent == "" || targetAgent == "claude" {
		for _, s := range claudeKeychainServices {
			if !seen[s] {
				seen[s] = true
				services = append(services, s)
			}
		}
	}

	return services
}

// PurgeAgentKeychain scrubs credentials for the specified agent from the macOS Keychain.
func PurgeAgentKeychain(agent string) error {
	return PurgeIgnoredKeychains(agent)
}
