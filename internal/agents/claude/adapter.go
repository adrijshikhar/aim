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

// ClaudeAdapter is an alias for Adapter.
type ClaudeAdapter = Adapter

func NewAdapter() *Adapter {
	return &Adapter{}
}

// NewClaudeAdapter creates a new ClaudeAdapter.
func NewClaudeAdapter() *ClaudeAdapter {
	return NewAdapter()
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

// HasCredentials returns true if ANTHROPIC_API_KEY is in the environment,
// or if the profile has a valid auth.json, .credentials.json, or .claude.json credentials file,
// or if macOS Keychain credentials can be harvested.
func (a *Adapter) HasCredentials(profileDir string) bool {
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		return true
	}
	authPath := filepath.Join(profileDir, ".claude", "auth.json")
	if fi, err := os.Stat(authPath); err == nil && !fi.IsDir() && fi.Size() > 0 {
		return true
	}
	credsPath := filepath.Join(profileDir, ".claude", ".credentials.json")
	if fi, err := os.Stat(credsPath); err == nil && !fi.IsDir() && fi.Size() > 0 {
		return true
	}
	for _, p := range []string{
		filepath.Join(profileDir, ".claude.json"),
		filepath.Join(profileDir, ".claude", ".claude.json"),
	} {
		if data, err := os.ReadFile(p); err == nil && len(data) > 0 {
			if strings.Contains(string(data), "oauthAccount") {
				return true
			}
		}
	}
	if profile.HarvestKeychainTokenToProfile(a.Name(), profileDir) {
		return true
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

	cleanEnv := make([]string, 0, len(os.Environ())+5)
	for _, env := range os.Environ() {
		idx := strings.IndexByte(env, '=')
		if idx == -1 {
			continue
		}
		key := env[:idx]
		if key == "HOME" || key == "CLAUDE_CONFIG_DIR" || key == "AIM_AGENT" || key == "AIM_PROFILE" || key == "AIM_HOME" || key == "AIM_SESSION_ID" {
			continue
		}
		cleanEnv = append(cleanEnv, env)
	}

	cmd.Env = append(cleanEnv,
		"HOME="+profileDir,
		"CLAUDE_CONFIG_DIR="+claudeDir,
		"AIM_AGENT="+a.Name(),
		"AIM_PROFILE="+profileName,
		"AIM_HOME="+config.BaseDir(),
	)

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
	envMap := make(map[string]string)
	for _, e := range os.Environ() {
		if idx := strings.IndexByte(e, '='); idx != -1 {
			envMap[e[:idx]] = e[idx+1:]
		}
	}

	envMap["HOME"] = profileDir
	envMap["CLAUDE_CONFIG_DIR"] = claudeDir
	envMap["AIM_AGENT"] = a.Name()
	envMap["AIM_PROFILE"] = profileName
	envMap["AIM_HOME"] = config.BaseDir()
	delete(envMap, "AIM_SESSION_ID")

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
			_ = os.WriteFile(destSettings, []byte(destContent), 0644)
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
	_ = os.WriteFile(destSettings, []byte(content), 0644)
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
		_ = os.WriteFile(dest, data, 0600)
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
