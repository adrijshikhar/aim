package codex

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
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

type codexRateLimitWindow struct {
	UsedPercent   float64 `json:"used_percent"`
	WindowMinutes int     `json:"window_minutes"`
	ResetsAt      int64   `json:"resets_at"`
}

type codexCredits struct {
	HasCredits bool   `json:"has_credits"`
	Unlimited  bool   `json:"unlimited"`
	Balance    string `json:"balance"`
}

type codexRateLimits struct {
	LimitID   string                `json:"limit_id"`
	LimitName string                `json:"limit_name"`
	Primary   *codexRateLimitWindow `json:"primary"`
	Secondary *codexRateLimitWindow `json:"secondary"`
	Credits   *codexCredits         `json:"credits"`
}

type codexSessionEvent struct {
	Type    string `json:"type"`
	Payload struct {
		Type       string           `json:"type"`
		RateLimits *codexRateLimits `json:"rate_limits"`
	} `json:"payload"`
	RateLimits      *codexRateLimits `json:"rate_limits"`
	RateLimitsCamel *codexRateLimits `json:"rateLimits"`
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

	rl := getLatestCodexRateLimits(profileName, profileDir)
	var windows []usage.LimitWindow
	var credits string

	if rl != nil {
		cat := rl.LimitName
		if cat == "" {
			cat = "Codex"
		}

		if rl.Primary != nil {
			var resetsAt time.Time
			var resetsIn time.Duration
			if rl.Primary.ResetsAt > 0 {
				resetsAt = time.Unix(rl.Primary.ResetsAt, 0)
				resetsIn = time.Until(resetsAt)
				if resetsIn < 0 {
					resetsIn = 0
				}
			}
			remaining := int(math.Round(100.0 - rl.Primary.UsedPercent))
			if remaining < 0 {
				remaining = 0
			} else if remaining > 100 {
				remaining = 100
			}
			if !resetsAt.IsZero() && time.Now().After(resetsAt) {
				remaining = 100
				resetsIn = 0
			}

			name := "5h (Primary)"
			if rl.Primary.WindowMinutes > 0 {
				wm := formatWindowMinutes(rl.Primary.WindowMinutes)
				if strings.Contains(strings.ToLower(wm), "5h") {
					name = "5h (Primary)"
				} else {
					name = fmt.Sprintf("%s (Primary)", wm)
				}
			}

			windows = append(windows, usage.LimitWindow{
				Category:     cat,
				Name:         name,
				RemainingPct: remaining,
				ResetsAt:     resetsAt,
				ResetsIn:     resetsIn,
			})
		}

		if rl.Secondary != nil {
			var resetsAt time.Time
			var resetsIn time.Duration
			if rl.Secondary.ResetsAt > 0 {
				resetsAt = time.Unix(rl.Secondary.ResetsAt, 0)
				resetsIn = time.Until(resetsAt)
				if resetsIn < 0 {
					resetsIn = 0
				}
			}
			remaining := int(math.Round(100.0 - rl.Secondary.UsedPercent))
			if remaining < 0 {
				remaining = 0
			} else if remaining > 100 {
				remaining = 100
			}
			if !resetsAt.IsZero() && time.Now().After(resetsAt) {
				remaining = 100
				resetsIn = 0
			}

			name := "Weekly (Secondary)"
			if rl.Secondary.WindowMinutes > 0 {
				wm := formatWindowMinutes(rl.Secondary.WindowMinutes)
				if strings.Contains(strings.ToLower(wm), "week") || strings.Contains(strings.ToLower(wm), "7d") {
					name = "Weekly (Secondary)"
				} else {
					name = fmt.Sprintf("%s (Secondary)", wm)
				}
			}

			windows = append(windows, usage.LimitWindow{
				Category:     cat,
				Name:         name,
				RemainingPct: remaining,
				ResetsAt:     resetsAt,
				ResetsIn:     resetsIn,
			})
		}

		if rl.Credits != nil {
			if rl.Credits.Unlimited {
				credits = "Unlimited"
			} else if rl.Credits.Balance != "" {
				credits = rl.Credits.Balance
			}
		}
	}

	status := usage.StatusOK
	if len(windows) > 0 {
		status = usage.CalculateStatus(windows)
		var parts []string
		for _, w := range windows {
			parts = append(parts, fmt.Sprintf("%s: %d%%", w.Name, w.RemainingPct))
		}
		summary += " • " + strings.Join(parts, ", ")
	}

	return &usage.Report{
		Agent:        a.Name(),
		Profile:      profileName,
		Status:       status,
		Windows:      windows,
		Credits:      credits,
		AccountEmail: accInfo.Email,
		AccountName:  accInfo.Name,
		AuthMethod:   accInfo.AuthMethod,
		Summary:      summary,
		FetchedAt:    time.Now(),
	}, nil
}

func formatWindowMinutes(minutes int) string {
	if minutes <= 0 {
		return ""
	}
	if minutes%(60*24) == 0 {
		days := minutes / (60 * 24)
		if days == 7 {
			return "Weekly"
		}
		return fmt.Sprintf("%dd", days)
	}
	if minutes%60 == 0 {
		return fmt.Sprintf("%dh", minutes/60)
	}
	return fmt.Sprintf("%dm", minutes)
}

type sessionFileInfo struct {
	path    string
	modTime time.Time
}

func findRecentSessionFiles(dir string) []sessionFileInfo {
	var files []sessionFileInfo
	if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
		return nil
	}
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".jsonl") {
			info, err := d.Info()
			if err == nil {
				files = append(files, sessionFileInfo{
					path:    path,
					modTime: info.ModTime(),
				})
			}
		}
		return nil
	})
	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime.After(files[j].modTime)
	})
	return files
}

func parseSessionFileRateLimits(filePath string) *codexRateLimits {
	f, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 10*1024*1024)
	var latest *codexRateLimits
	for scanner.Scan() {
		line := scanner.Bytes()
		if !strings.Contains(string(line), "rate_limits") && !strings.Contains(string(line), "rateLimits") {
			continue
		}
		var msg codexSessionEvent
		if err := json.Unmarshal(line, &msg); err == nil {
			rl := msg.Payload.RateLimits
			if rl == nil {
				rl = msg.RateLimits
			}
			if rl == nil {
				rl = msg.RateLimitsCamel
			}
			if rl != nil {
				latest = rl
			}
		}
	}
	if err := scanner.Err(); err != nil {
		logger.Debug("[codex] scanner error reading session file %s: %v", filePath, err)
	}
	return latest
}

func getLatestCodexRateLimits(profileName, profileDir string) *codexRateLimits {
	// 1. Check profileDir/.codex/sessions
	profileSessionsDir := filepath.Join(profileDir, ".codex", "sessions")
	files := findRecentSessionFiles(profileSessionsDir)

	// 2. If no session files found in profile, and profile is eligible for seeding (e.g. primary profile),
	// fall back to host ~/.codex/sessions
	if len(files) == 0 && isProfileEligibleForSeeding(profileName) {
		hostSessionsDir := filepath.Join(config.RealHomeDir(), ".codex", "sessions")
		files = findRecentSessionFiles(hostSessionsDir)
	}

	maxCheck := 5
	if len(files) < maxCheck {
		maxCheck = len(files)
	}
	for i := 0; i < maxCheck; i++ {
		if rl := parseSessionFileRateLimits(files[i].path); rl != nil {
			return rl
		}
	}
	return nil
}
