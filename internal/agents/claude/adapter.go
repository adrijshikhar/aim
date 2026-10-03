package claude

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/logger"
	"github.com/aim-cli/aim/internal/merge"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/usage"
)

// Compile-time assertion that Adapter implements agents.AgentAdapter.
var _ agents.AgentAdapter = (*Adapter)(nil)

type Adapter struct {
	agents.BaseAdapter
}

func NewAdapter() *Adapter {
	return &Adapter{
		BaseAdapter: agents.NewBaseAdapter("claude", "Claude Code", "claude", []string{"cc", "claude-code"}),
	}
}

// HasCredentials returns true if the profile has valid authentication credentials:
// - ANTHROPIC_API_KEY in environment
// - auth.json file exists and is non-empty
// - .credentials.json exists and contains valid claudeAiOauth tokens
// - Scoped Keychain service contains valid claudeAiOauth tokens
func (a *Adapter) HasCredentials(profileDir string) bool {
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		return true
	}
	claudeDir := filepath.Join(profileDir, ".claude")
	authPath := filepath.Join(claudeDir, "auth.json")
	if fi, err := os.Stat(authPath); err == nil && !fi.IsDir() && fi.Size() > 0 {
		return true
	}
	credsPath := filepath.Join(claudeDir, ".credentials.json")
	if data, err := os.ReadFile(credsPath); err == nil && len(data) > 0 {
		if profile.HasClaudeCredentials(data) {
			return true
		}
	}
	if profile.ShouldSeedCredentials(filepath.Base(profileDir)) {
		if profile.HarvestKeychainTokenToProfile(a.Name(), profileDir) {
			return true
		}
	}
	return false
}

// Login launches the interactive `claude auth login` flow scoped to the profile directory.
func (a *Adapter) Login(ctx context.Context, profileName, profileDir string) error {
	claudeDir := filepath.Join(profileDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0700); err != nil {
		return fmt.Errorf("failed to create claude config dir: %w", err)
	}

	seedClaudeJSON(config.RealHomeDir(), profileDir)

	bin := a.ResolveBinary()
	cmd := exec.CommandContext(ctx, bin, "auth", "login")
	cmd.Dir = profileDir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	storageEnv := config.StorageEnv()
	cleanEnv := make([]string, 0, len(os.Environ())+5)
	for _, env := range os.Environ() {
		idx := strings.IndexByte(env, '=')
		if idx == -1 {
			continue
		}
		key := env[:idx]
		_, storageKey := storageEnv[key]
		if storageKey || key == "HOME" || key == "CLAUDE_CONFIG_DIR" || key == "AIM_AGENT" || key == "AIM_PROFILE" || key == "AIM_SESSION_ID" ||
			key == "CLAUDE_CODE_OAUTH_TOKEN" || key == "ANTHROPIC_API_KEY" || key == "CLAUDE_CODE_OAUTH_REFRESH_TOKEN" {
			continue
		}
		cleanEnv = append(cleanEnv, env)
	}

	cmd.Env = append(cleanEnv,
		"HOME="+profileDir,
		"CLAUDE_CONFIG_DIR="+claudeDir,
		"AIM_AGENT="+a.Name(),
		"AIM_PROFILE="+profileName,
	)
	for key, value := range storageEnv {
		cmd.Env = append(cmd.Env, key+"="+value)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("claude login failed: %w", err)
	}
	return nil
}

// PrepareEnv sets up the execution environment for a Claude Code session.
func (a *Adapter) PrepareEnv(ctx context.Context, profileName, profileDir string) (agents.LaunchEnv, error) {
	claudeDir := filepath.Join(profileDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0700); err != nil {
		return agents.LaunchEnv{}, fmt.Errorf("failed to create claude config dir: %w", err)
	}

	seedClaudeJSON(config.RealHomeDir(), profileDir)
	rewriteSettingsHooks(config.RealHomeDir(), profileDir)
	_ = profile.HarvestKeychainTokenToProfile(a.Name(), profileDir)

	bin := a.ResolveBinary()
	envMap := a.BaseLaunchEnv(profileName, profileDir, map[string]string{
		"CLAUDE_CONFIG_DIR": claudeDir,
	})

	logger.Debug("[claude] Launch env: HOME=%s, CLAUDE_CONFIG_DIR=%s, AIM_AGENT=%s, AIM_PROFILE=%s",
		profileDir, claudeDir, a.Name(), profileName)

	cwd, _ := os.Getwd()
	return agents.LaunchEnv{
		BinaryPath: bin,
		Env:        envMap,
		WorkingDir: cwd,
	}, nil
}

// writeSettings writes the profile's first settings.json. Replaced in tests
// (they count calls), so those tests must not use t.Parallel.
var writeSettings = merge.AtomicWrite

// rewriteSettingsHooks points host hook paths in the profile's settings.json at
// the profile. The first copy from the host carries no plugin enablement: that
// is merged per session (G5). Only the hooks member is rewritten, so a
// marketplace path under the host .claude is never turned into a profile path.
func rewriteSettingsHooks(hostHome, profileDir string) {
	destDir := filepath.Join(profileDir, ".claude")
	_ = os.MkdirAll(destDir, 0700)
	destSettings := filepath.Join(destDir, "settings.json")
	hostClaude := filepath.Join(hostHome, ".claude")
	destClaude := filepath.Join(profileDir, ".claude")

	_, err := os.Stat(destSettings)
	switch {
	case errors.Is(err, os.ErrNotExist):
		seedSettings(filepath.Join(hostClaude, "settings.json"), destSettings, hostClaude, destClaude)
		return
	case err != nil:
		logger.Warn("[claude] %s: %v; its hook paths are not rewritten", destSettings, err)
		return
	}
	if _, err := merge.ReplaceInJSONMember(destSettings, "hooks", hostClaude, destClaude, 0o600); err != nil {
		logger.Debug("[claude] Failed to update rewritten settings.json at %s: %v", destSettings, err)
	}
}

// seedSettings makes the profile's first settings.json from the host's, in
// memory and written once (0600): plugin enablement and marketplaces dropped,
// hook paths pointed at the profile. A host file that cannot be stripped
// safely is not copied at all; Claude then starts with its defaults rather
// than with the host's enablement.
func seedSettings(hostSettings, dest, hostClaude, destClaude string) {
	data, err := os.ReadFile(hostSettings)
	if err != nil {
		return
	}
	if len(bytes.TrimSpace(data)) > 0 {
		for _, key := range []string{"enabledPlugins", "extraKnownMarketplaces"} {
			if data, err = merge.DeleteJSONKeyBytes(data, key); err != nil {
				break
			}
		}
		if err == nil {
			data, _, err = merge.ReplaceInJSONMemberBytes(data, "hooks", hostClaude, destClaude)
		}
		if err != nil {
			logger.Warn("[claude] %s not copied to the profile (%v); Claude starts with its default settings", hostSettings, err)
			return
		}
	}
	if err := writeSettings(dest, data, 0o600); err != nil {
		logger.Warn("[claude] Failed to write settings.json to %s: %v", dest, err)
	}
}

// seedClaudeJSON initializes the profile's .claude/.claude.json from the host's
// ~/.claude.json if it does not already exist. It strips credentials (oauthAccount)
// and static mcpServers (which Spec A merges dynamically per-session), while
// preserving user onboarding state (hasCompletedOnboarding, lastOnboardingVersion),
// theme, editor preferences, and projects trust (hasTrustDialogAccepted, hasClaudeMdExternalIncludesApproved).
func seedClaudeJSON(hostHome, profileDir string) {
	destDir := filepath.Join(profileDir, ".claude")
	_ = os.MkdirAll(destDir, 0700)
	dest := filepath.Join(destDir, ".claude.json")

	if _, err := os.Stat(dest); err == nil {
		return // already exists; do not overwrite profile overrides
	}

	src := filepath.Join(hostHome, ".claude.json")
	data, err := os.ReadFile(src)
	if err != nil || len(bytes.TrimSpace(data)) == 0 {
		return
	}

	for _, key := range []string{"oauthAccount", "mcpServers"} {
		if data, err = merge.DeleteJSONKeyBytes(data, key); err != nil {
			logger.Warn("[claude] Failed to strip %s from seeded .claude.json: %v", key, err)
			return
		}
	}

	if err := writeSettings(dest, data, 0o600); err != nil {
		logger.Warn("[claude] Failed to write .claude.json to %s: %v", dest, err)
	}
}

// Doctor performs diagnostics on the Claude Code installation and profile state.
func (a *Adapter) Doctor(ctx context.Context, profileName, profileDir string) []agents.DiagnosticResult {
	var results []agents.DiagnosticResult
	bin := a.ResolveBinary()
	out, err := exec.CommandContext(ctx, bin, "--version").Output()
	if err != nil {
		results = append(results, agents.DiagnosticResult{
			Category: "Binary",
			Status:   "FAIL",
			Message:  fmt.Sprintf("Claude CLI not found or failed: %v", err),
		})
	} else {
		results = append(results, agents.DiagnosticResult{
			Category: "Binary",
			Status:   "OK",
			Message:  fmt.Sprintf("Claude installed (%s at %s)", strings.TrimSpace(string(out)), bin),
		})
	}

	if _, err := os.Stat(filepath.Join(profileDir, ".claude.json")); err == nil {
		results = append(results, agents.DiagnosticResult{Category: "Config", Status: "OK",
			Message: "~/.claude.json in this profile is used only when Claude runs without CLAUDE_CONFIG_DIR"})
	}
	if a.HasCredentials(profileDir) {
		results = append(results, agents.DiagnosticResult{
			Category: "Auth",
			Status:   "OK",
			Message:  "Claude credentials detected",
		})
	} else {
		results = append(results, agents.DiagnosticResult{
			Category: "Auth",
			Status:   "WARN",
			Message:  fmt.Sprintf("No Claude credentials found (run 'aim login claude %s')", profileName),
		})
	}

	claudeDir := filepath.Join(profileDir, ".claude")
	if fi, err := os.Stat(claudeDir); err == nil && fi.Mode().Perm() == 0700 {
		results = append(results, agents.DiagnosticResult{
			Category: "Storage",
			Status:   "OK",
			Message:  "Storage directory permissions OK (0700)",
		})
	} else if err == nil {
		results = append(results, agents.DiagnosticResult{
			Category: "Storage",
			Status:   "OK",
			Message:  "Storage directory exists",
		})
	} else {
		results = append(results, agents.DiagnosticResult{
			Category: "Storage",
			Status:   "WARN",
			Message:  fmt.Sprintf("Storage directory not yet created (%s)", claudeDir),
		})
	}

	// Hooks Check
	profileSettings := filepath.Join(claudeDir, "settings.json")
	hostClaude := filepath.Join(config.RealHomeDir(), ".claude")
	if changed, err := merge.ReplaceInJSONMember(profileSettings, "hooks", hostClaude, claudeDir, 0o600); err == nil && changed {
		results = append(results, agents.DiagnosticResult{
			Category: "Hooks",
			Status:   "OK",
			Message:  "Auto-migrated host hook paths in settings.json to profile",
		})
	}

	return results
}

// GetUsage returns the usage report for the Claude Code agent.
func (a *Adapter) GetUsage(ctx context.Context, profileName, profileDir string) (*usage.Report, error) {
	acc := profile.GetProfileAccountInfoForAgent(profileDir, a.Name())

	if !a.HasCredentials(profileDir) {
		return &usage.Report{
			Agent:        a.Name(),
			Profile:      profileName,
			Status:       usage.StatusUnknown,
			FetchedAt:    time.Now(),
			Error:        "no credentials",
			AccountEmail: acc.Email,
			AccountName:  acc.Name,
			AuthMethod:   acc.AuthMethod,
		}, nil
	}

	bin := a.ResolveBinary()
	if _, err := exec.LookPath(bin); err != nil {
		if _, statErr := os.Stat(bin); statErr != nil {
			return &usage.Report{
				Agent:        a.Name(),
				Profile:      profileName,
				Status:       usage.StatusUnknown,
				FetchedAt:    time.Now(),
				Error:        "binary not found",
				AccountEmail: acc.Email,
				AccountName:  acc.Name,
				AuthMethod:   acc.AuthMethod,
			}, nil
		}
	}

	cmd := exec.CommandContext(ctx, bin, "-p", "/usage")
	cmd.Dir = profileDir
	cmd.Env = append(os.Environ(),
		"HOME="+profileDir,
		"AIM_AGENT="+a.Name(),
		"AIM_PROFILE="+profileName,
	)

	out, err := cmd.Output()
	if err != nil {
		report := &usage.Report{
			Agent:        a.Name(),
			Profile:      profileName,
			Status:       usage.StatusUnknown,
			FetchedAt:    time.Now(),
			Error:        err.Error(),
			AccountEmail: acc.Email,
			AccountName:  acc.Name,
			AuthMethod:   acc.AuthMethod,
			Summary:      "Offline",
		}
		return report, nil
	}

	now := time.Now()
	windows := ParseClaudeUsage(string(out), now)
	status := usage.CalculateStatus(windows)

	report := &usage.Report{
		Agent:        a.Name(),
		Profile:      profileName,
		Status:       status,
		Windows:      windows,
		Credits:      "0",
		AccountEmail: acc.Email,
		AccountName:  acc.Name,
		AuthMethod:   acc.AuthMethod,
		FetchedAt:    now,
	}

	var parts []string
	seen := make(map[string]bool)
	for _, w := range windows {
		var label string
		if w.IsHourly() {
			label = "5h"
		} else if w.IsWeekly() {
			label = "wk"
		}
		if label != "" {
			if w.Category != "" && w.Category != "All Models" && w.Category != "Claude" {
				label = fmt.Sprintf("%s %s", w.Category, label)
			}
			if !seen[label] {
				seen[label] = true
				parts = append(parts, fmt.Sprintf("%s: %d%%", label, w.RemainingPct))
			}
		}
	}
	if len(parts) > 0 {
		report.Summary = strings.Join(parts, ", ")
	} else if acc.AuthMethod != "" {
		report.Summary = acc.AuthMethod
	} else {
		report.Summary = "Active credentials"
	}

	return report, nil
}

// ParseClaudeUsage parses the output of "claude -p /usage" into a slice of LimitWindow.
func ParseClaudeUsage(output string, now time.Time) []usage.LimitWindow {
	var windows []usage.LimitWindow
	rePct := regexp.MustCompile(`(\d+(?:\.\d+)?)%`)
	reResets := regexp.MustCompile(`(?i)resets\s+([^·\n]+)`)

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if !strings.Contains(line, ":") || !strings.Contains(line, "%") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		header := strings.TrimSpace(parts[0])
		body := strings.TrimSpace(parts[1])

		headerLower := strings.ToLower(header)
		if strings.Contains(headerLower, "last 24h") || strings.Contains(headerLower, "last 7d") || strings.Contains(headerLower, "approximate") {
			continue
		}

		pctMatch := rePct.FindStringSubmatch(body)
		if len(pctMatch) < 2 {
			continue
		}

		usedPct, err := strconv.ParseFloat(pctMatch[1], 64)
		if err != nil {
			continue
		}

		remainingPct := int(math.Round(100.0 - usedPct))
		if remainingPct < 0 {
			remainingPct = 0
		} else if remainingPct > 100 {
			remainingPct = 100
		}

		var windowName string
		if strings.Contains(headerLower, "session") || strings.Contains(headerLower, "5h") || strings.Contains(headerLower, "hour") {
			windowName = "5h Limit"
		} else if strings.Contains(headerLower, "week") || strings.Contains(headerLower, "7d") || strings.Contains(headerLower, "wk") {
			windowName = "Weekly Limit"
		} else {
			windowName = header
		}

		category := "All Models"
		if idxOpen := strings.Index(header, "("); idxOpen != -1 {
			if idxClose := strings.Index(header[idxOpen:], ")"); idxClose != -1 {
				inside := strings.TrimSpace(header[idxOpen+1 : idxOpen+idxClose])
				if strings.EqualFold(inside, "all models") || strings.EqualFold(inside, "all") {
					category = "All Models"
				} else if inside != "" {
					category = inside
				}
			}
		}

		var resetsAt time.Time
		var resetsIn time.Duration
		if resetMatch := reResets.FindStringSubmatch(body); len(resetMatch) >= 2 {
			rawReset := strings.TrimSpace(resetMatch[1])
			resetsAt, resetsIn = parseClaudeResetTime(rawReset, now)
		}

		if !resetsAt.IsZero() && now.After(resetsAt) {
			remainingPct = 100
			resetsIn = 0
		}

		windows = append(windows, usage.LimitWindow{
			Category:     category,
			Name:         windowName,
			RemainingPct: remainingPct,
			ResetsAt:     resetsAt,
			ResetsIn:     resetsIn,
		})
	}

	return windows
}

func parseClaudeResetTime(s string, now time.Time) (time.Time, time.Duration) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, 0
	}

	loc := time.Local
	if idxOpen := strings.LastIndex(s, "("); idxOpen != -1 {
		if idxClose := strings.LastIndex(s, ")"); idxClose > idxOpen {
			tzStr := strings.TrimSpace(s[idxOpen+1 : idxClose])
			if l, err := time.LoadLocation(tzStr); err == nil {
				loc = l
			}
			s = strings.TrimSpace(s[:idxOpen])
		}
	}

	clean := strings.ReplaceAll(s, " at ", " ")
	clean = strings.ReplaceAll(clean, ",", "")
	clean = strings.Join(strings.Fields(clean), " ")

	nowInLoc := now.In(loc)

	layoutsWithYear := []string{
		"Jan 2 2006 3:04pm",
		"Jan 2 2006 3:04PM",
		"Jan 2 2006 15:04",
		"2006-01-02 15:04",
	}
	for _, l := range layoutsWithYear {
		if t, err := time.ParseInLocation(l, clean, loc); err == nil {
			rem := t.Sub(nowInLoc)
			if rem < 0 {
				rem = 0
			}
			return t, rem
		}
	}

	layoutsWithoutYear := []string{
		"Jan 2 3:04pm",
		"Jan 2 3:04PM",
		"Jan 2 15:04",
		"January 2 3:04pm",
		"January 2 3:04PM",
		"January 2 15:04",
	}
	for _, l := range layoutsWithoutYear {
		if t, err := time.ParseInLocation(l, clean, loc); err == nil {
			res := time.Date(nowInLoc.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, loc)
			if res.Before(nowInLoc.Add(-24 * time.Hour)) {
				res = res.AddDate(1, 0, 0)
			}
			rem := res.Sub(nowInLoc)
			if rem < 0 {
				rem = 0
			}
			return res, rem
		}
	}

	return time.Time{}, 0
}
