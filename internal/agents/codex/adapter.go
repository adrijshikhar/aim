package codex

import (
	"bufio"
	"context"
	"encoding/json"
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

// CodexAdapter is an alias for Adapter.
type CodexAdapter = Adapter

func NewAdapter() *Adapter {
	return &Adapter{}
}

// NewCodexAdapter creates a new CodexAdapter.
func NewCodexAdapter() *CodexAdapter {
	return NewAdapter()
}

func (a *Adapter) Name() string        { return "codex" }
func (a *Adapter) DisplayName() string { return "Codex CLI" }
func (a *Adapter) Aliases() []string   { return []string{"codex-cli", "openai-codex"} }
func (a *Adapter) BinaryName() string  { return "codex" }

func (a *Adapter) TokenPath(profileDir string) string {
	return filepath.Join(profileDir, ".codex", "auth.json")
}

func (a *Adapter) HasCredentials(profileDir string) bool {
	p := a.TokenPath(profileDir)
	fi, err := os.Stat(p)
	if err == nil && !fi.IsDir() && fi.Size() > 0 {
		logger.Debug("[codex] HasCredentials: true (auth.json at %s)", p)
		return true
	}

	// Auto-seed credentials if eligible profile doesn't have credentials yet
	profileName := filepath.Base(profileDir)
	if a.SeedDefaultCredentials(profileName, profileDir) {
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() && fi.Size() > 0 {
			logger.Debug("[codex] HasCredentials: true (auto-seeded auth.json at %s)", p)
			return true
		}
	}

	logger.Debug("[codex] HasCredentials: false (no auth.json at %s)", p)
	return false
}

// SeedDefaultCredentials seeds host ~/.codex/auth.json into profile directory for eligible primary profiles.
func (a *Adapter) SeedDefaultCredentials(profileName, profileDir string) bool {
	if !isProfileEligibleForSeeding(profileName) {
		return false
	}
	realHome := config.RealHomeDir()
	hostAuthPath := filepath.Join(realHome, ".codex", "auth.json")
	fi, err := os.Stat(hostAuthPath)
	if err != nil || fi.IsDir() || fi.Size() == 0 {
		return false
	}

	profileAuthDir := filepath.Join(profileDir, ".codex")
	if err := os.MkdirAll(profileAuthDir, 0700); err != nil {
		return false
	}
	destPath := a.TokenPath(profileDir)
	data, err := os.ReadFile(hostAuthPath)
	if err != nil || len(data) == 0 {
		return false
	}
	tmp := destPath + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return false
	}
	if err := os.Rename(tmp, destPath); err != nil {
		_ = os.Remove(tmp)
		return false
	}
	logger.Debug("[codex] Successfully seeded credentials from %s to %s", hostAuthPath, destPath)
	return true
}

func isProfileEligibleForSeeding(profileName string) bool {
	switch profileName {
	case "personal", "p", "me", "main":
		return true
	}

	cfg, err := config.LoadConfig()
	if err == nil && cfg != nil {
		if cfg.DefaultProfile != "" && cfg.DefaultProfile == profileName {
			return true
		}
		if len(cfg.Profiles) == 1 {
			for p := range cfg.Profiles {
				if p == profileName {
					return true
				}
			}
		}
	}
	return false
}

// ResolveBinary locates the codex executable on the system.
func (a *Adapter) ResolveBinary() string {
	bin, err := exec.LookPath(a.BinaryName())
	if err == nil {
		return bin
	}
	realHome := config.RealHomeDir()
	fallback := filepath.Join(realHome, ".local", "bin", a.BinaryName())
	if _, err := os.Stat(fallback); err == nil {
		return fallback
	}
	return a.BinaryName()
}

func (a *Adapter) Login(ctx context.Context, profileName, profileDir string) error {
	codexDir := filepath.Join(profileDir, ".codex")
	if err := os.MkdirAll(codexDir, 0700); err != nil {
		return err
	}

	bin := a.ResolveBinary()
	cmd := exec.CommandContext(ctx, bin, "login")
	cmd.Dir = profileDir

	cleanEnv := make([]string, 0, len(os.Environ())+4)
	for _, env := range os.Environ() {
		idx := strings.IndexByte(env, '=')
		if idx == -1 {
			continue
		}
		key := env[:idx]
		if key == "HOME" || key == "CODEX_HOME" || key == "AIM_AGENT" || key == "AIM_PROFILE" || key == "AIM_HOME" {
			continue
		}
		cleanEnv = append(cleanEnv, env)
	}

	cmd.Env = append(cleanEnv,
		"HOME="+profileDir,
		"CODEX_HOME="+codexDir,
		"AIM_AGENT="+a.Name(),
		"AIM_PROFILE="+profileName,
		"AIM_HOME="+config.BaseDir(),
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("codex login failed: %w", err)
	}

	return nil
}

func (a *Adapter) PrepareEnv(profileName, profileDir string) (agents.LaunchEnv, error) {
	codexDir := filepath.Join(profileDir, ".codex")
	if err := os.MkdirAll(codexDir, 0700); err != nil {
		return agents.LaunchEnv{}, err
	}

	// Auto-seed credentials for eligible profile if missing in profileDir but available on host
	_ = a.SeedDefaultCredentials(profileName, profileDir)

	// Copy host config.toml to profile if not present
	realHome := config.RealHomeDir()
	copyHostConfig(realHome, codexDir)

	bin := a.ResolveBinary()
	logger.Debug("[codex] Resolved binary: %s", bin)

	envMap := make(map[string]string)
	for _, e := range os.Environ() {
		idx := strings.IndexByte(e, '=')
		if idx != -1 {
			envMap[e[:idx]] = e[idx+1:]
		}
	}

	envMap["HOME"] = profileDir
	envMap["CODEX_HOME"] = codexDir
	envMap["AIM_AGENT"] = a.Name()
	envMap["AIM_PROFILE"] = profileName
	envMap["AIM_HOME"] = config.BaseDir()

	logger.Debug("[codex] Launch env: HOME=%s, CODEX_HOME=%s, AIM_AGENT=%s, AIM_PROFILE=%s",
		profileDir, codexDir, a.Name(), profileName)

	cwd, _ := os.Getwd()
	return agents.LaunchEnv{
		BinaryPath: bin,
		Env:        envMap,
		WorkingDir: cwd,
	}, nil
}

func copyHostConfig(realHome, profileCodexDir string) {
	hostConfig := filepath.Join(realHome, ".codex", "config.toml")
	destConfig := filepath.Join(profileCodexDir, "config.toml")
	if _, err := os.Stat(destConfig); os.IsNotExist(err) {
		if data, err := os.ReadFile(hostConfig); err == nil && len(data) > 0 {
			_ = os.WriteFile(destConfig, data, 0644)
		}
	}
}

func (a *Adapter) Doctor(ctx context.Context, profileName, profileDir string) []agents.DiagnosticResult {
	var results []agents.DiagnosticResult

	// 1. Binary Check
	bin := a.ResolveBinary()
	if path, err := exec.LookPath(bin); err == nil {
		verCmd := exec.CommandContext(ctx, path, "--version")
		out, err := verCmd.Output()
		verStr := strings.TrimSpace(string(out))
		if err == nil && verStr != "" {
			results = append(results, agents.DiagnosticResult{
				Category: "Binary",
				Status:   "OK",
				Message:  fmt.Sprintf("Codex installed (%s at %s)", verStr, path),
			})
		} else {
			results = append(results, agents.DiagnosticResult{
				Category: "Binary",
				Status:   "OK",
				Message:  fmt.Sprintf("Codex installed (%s)", path),
			})
		}
	} else {
		results = append(results, agents.DiagnosticResult{
			Category: "Binary",
			Status:   "FAIL",
			Message:  "Codex binary not found in PATH or ~/.local/bin/codex",
		})
	}

	// 2. Credentials Check
	tokenPath := a.TokenPath(profileDir)
	if a.HasCredentials(profileDir) {
		accInfo := profile.GetProfileAccountInfoForAgent(profileDir, a.Name())
		msg := "Active credentials"
		if accInfo.Email != "" {
			msg = fmt.Sprintf("Active credentials (%s", accInfo.Email)
			if accInfo.AuthMethod != "" {
				msg += fmt.Sprintf(", %s", accInfo.AuthMethod)
			}
			msg += ")"
		}
		results = append(results, agents.DiagnosticResult{
			Category: "Auth",
			Status:   "OK",
			Message:  msg,
		})
	} else {
		results = append(results, agents.DiagnosticResult{
			Category: "Auth",
			Status:   "WARN",
			Message:  fmt.Sprintf("No credentials found at %s (run 'aim login codex %s')", tokenPath, profileName),
		})
	}

	// 3. Environment & Storage Check
	codexDir := filepath.Join(profileDir, ".codex")
	if fi, err := os.Stat(codexDir); err == nil && fi.IsDir() {
		results = append(results, agents.DiagnosticResult{
			Category: "Storage",
			Status:   "OK",
			Message:  fmt.Sprintf("Isolated CODEX_HOME directory (%s)", codexDir),
		})
	} else {
		results = append(results, agents.DiagnosticResult{
			Category: "Storage",
			Status:   "WARN",
			Message:  fmt.Sprintf("CODEX_HOME directory not yet created (%s)", codexDir),
		})
	}

	return results
}

func (a *Adapter) GetUsage(ctx context.Context, profileName, profileDir string) (*usage.Report, error) {
	if !a.HasCredentials(profileDir) {
		return &usage.Report{
			Agent:     a.Name(),
			Profile:   profileName,
			Status:    usage.StatusUnknown,
			FetchedAt: time.Now(),
			Error:     "no credentials",
		}, nil
	}

	accInfo := profile.GetProfileAccountInfoForAgent(profileDir, a.Name())
	summary := "Active credentials"
	if accInfo.AuthMethod != "" {
		summary = accInfo.AuthMethod
	}

	sessionsDir := filepath.Join(profileDir, ".codex", "sessions")
	sessionSummary := parseLatestSessionRateLimits(sessionsDir)
	if sessionSummary != "" {
		summary += " • " + sessionSummary
	}

	return &usage.Report{
		Agent:        a.Name(),
		Profile:      profileName,
		Status:       usage.StatusOK,
		AccountEmail: accInfo.Email,
		AccountName:  accInfo.Name,
		AuthMethod:   accInfo.AuthMethod,
		Summary:      summary,
		FetchedAt:    time.Now(),
	}, nil
}

func parseLatestSessionRateLimits(sessionsDir string) string {
	entries, err := os.ReadDir(sessionsDir)
	if err != nil || len(entries) == 0 {
		return ""
	}

	// Find the newest session jsonl file
	var latestFile string
	var latestMod time.Time
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".jsonl") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().After(latestMod) {
			latestMod = info.ModTime()
			latestFile = filepath.Join(sessionsDir, e.Name())
		}
	}

	if latestFile == "" {
		return ""
	}

	f, err := os.Open(latestFile)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 10*1024*1024)
	var lastRateLimitSummary string
	for scanner.Scan() {
		line := scanner.Bytes()
		var msg struct {
			RateLimits struct {
				Primary struct {
					UsedPercent int `json:"used_percent"`
				} `json:"primary"`
			} `json:"rateLimits"`
		}
		if err := json.Unmarshal(line, &msg); err == nil {
			if msg.RateLimits.Primary.UsedPercent > 0 {
				lastRateLimitSummary = fmt.Sprintf("%d%% rate limit used", msg.RateLimits.Primary.UsedPercent)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		logger.Debug("[codex] scanner error reading session file %s: %v", latestFile, err)
	}

	return lastRateLimitSummary
}
