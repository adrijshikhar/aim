package claude

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/logger"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/usage"
)

// Compile-time assertion that Adapter implements agents.AgentAdapter.
var _ agents.AgentAdapter = (*Adapter)(nil)

type Adapter struct{}

func NewAdapter() *Adapter {
	return &Adapter{}
}

func (a *Adapter) Name() string        { return "claude" }
func (a *Adapter) DisplayName() string { return "Claude Code" }
func (a *Adapter) Aliases() []string   { return []string{"cc", "claude-code"} }
func (a *Adapter) BinaryName() string  { return "claude" }

// ResolveBinary locates the claude executable on the system.
func (a *Adapter) ResolveBinary() string {
	bin, err := exec.LookPath(a.BinaryName())
	if err == nil {
		return bin
	}
	fallbacks := []string{
		filepath.Join(config.RealHomeDir(), ".local", "bin", a.BinaryName()),
		filepath.Join("/opt/homebrew/bin", a.BinaryName()),
		filepath.Join("/usr/local/bin", a.BinaryName()),
	}
	for _, fb := range fallbacks {
		if _, err := os.Stat(fb); err == nil {
			return fb
		}
	}
	return a.BinaryName()
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
func (a *Adapter) PrepareEnv(profileName, profileDir string) (agents.LaunchEnv, error) {
	claudeDir := filepath.Join(profileDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0700); err != nil {
		return agents.LaunchEnv{}, fmt.Errorf("failed to create claude config dir: %w", err)
	}

	rewriteSettingsHooks(config.RealHomeDir(), profileDir)
	copyClaudeJSON(config.RealHomeDir(), profileDir)
	_ = profile.HarvestKeychainTokenToProfile(a.Name(), profileDir)

	bin := a.ResolveBinary()
	envMap := config.StorageEnv()
	envMap["HOME"] = profileDir
	envMap["CLAUDE_CONFIG_DIR"] = claudeDir
	envMap["AIM_AGENT"] = a.Name()
	envMap["AIM_PROFILE"] = profileName

	logger.Debug("[claude] Launch env: HOME=%s, CLAUDE_CONFIG_DIR=%s, AIM_AGENT=%s, AIM_PROFILE=%s",
		profileDir, claudeDir, a.Name(), profileName)

	cwd, _ := os.Getwd()
	return agents.LaunchEnv{
		BinaryPath: bin,
		Env:        envMap,
		WorkingDir: cwd,
	}, nil
}

// rewriteSettingsHooks copies or rewrites settings.json hooks from the host home to the profile.
func rewriteSettingsHooks(hostHome, profileDir string) {
	destDir := filepath.Join(profileDir, ".claude")
	_ = os.MkdirAll(destDir, 0700)
	destSettings := filepath.Join(destDir, "settings.json")
	hostClaude := filepath.Join(hostHome, ".claude")
	destClaude := filepath.Join(profileDir, ".claude")

	// If destSettings already exists, check and rewrite any host paths in it
	if destData, err := os.ReadFile(destSettings); err == nil {
		destContent := string(destData)
		if strings.Contains(destContent, hostClaude) {
			destContent = strings.ReplaceAll(destContent, hostClaude, destClaude)
			if err := os.WriteFile(destSettings, []byte(destContent), 0644); err != nil {
				logger.Debug("[claude] Failed to update rewritten settings.json at %s: %v", destSettings, err)
			}
		}
		return
	}

	// Otherwise, copy from hostSettings if it exists, rewriting host paths
	hostSettings := filepath.Join(hostHome, ".claude", "settings.json")
	data, err := os.ReadFile(hostSettings)
	if err != nil {
		return
	}
	content := string(data)
	if strings.Contains(content, hostClaude) {
		content = strings.ReplaceAll(content, hostClaude, destClaude)
	}
	if err := os.WriteFile(destSettings, []byte(content), 0644); err != nil {
		logger.Debug("[claude] Failed to write settings.json to %s: %v", destSettings, err)
	}
}

// copyClaudeJSON copies .claude.json from host to the profile directory if it does not
// already exist or if the destination file lacks oauthAccount.
func copyClaudeJSON(hostHome, profileDir string) {
	dest := filepath.Join(profileDir, ".claude.json")
	if data, err := os.ReadFile(dest); err == nil && len(data) > 0 {
		if strings.Contains(string(data), "oauthAccount") {
			return
		}
	}
	src := filepath.Join(hostHome, ".claude.json")
	if data, err := os.ReadFile(src); err == nil && len(data) > 0 {
		if err := os.WriteFile(dest, data, 0600); err != nil {
			logger.Debug("[claude] Failed to copy .claude.json to %s: %v", dest, err)
		}
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

	copyClaudeJSON(config.RealHomeDir(), profileDir)
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
	if data, err := os.ReadFile(profileSettings); err == nil {
		hostClaude := filepath.Join(config.RealHomeDir(), ".claude")
		if strings.Contains(string(data), hostClaude) {
			rewriteSettingsHooks(config.RealHomeDir(), profileDir)
			results = append(results, agents.DiagnosticResult{
				Category: "Hooks",
				Status:   "OK",
				Message:  "Auto-migrated host hook paths in settings.json to profile",
			})
		}
	}

	return results
}

// GetUsage returns the usage report for the Claude Code agent.
func (a *Adapter) GetUsage(ctx context.Context, profileName, profileDir string) (*usage.Report, error) {
	return &usage.Report{
		Agent:     a.Name(),
		Profile:   profileName,
		Status:    usage.StatusUnknown,
		FetchedAt: time.Now(),
	}, nil
}
