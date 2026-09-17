package codex

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
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

	// Copy host config.toml to profile if not present (with hook trust path rewriting)
	realHome := config.RealHomeDir()
	copyHostConfig(realHome, codexDir)

	// Probe and auto-start local sidecar proxy daemons (e.g. Caveman) if configured
	ensureSidecarDaemons(realHome, codexDir)

	// Bridge host plugins and hooks (e.g. Catalyst) into profile
	bridgePluginsAndHooks(realHome, codexDir)

	// Bridge host cxstatusline config into profile
	bridgeCxStatusline(realHome, profileDir)

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

// deduplicateTomlTables removes duplicate table headers [foo] in TOML files while preserving
// array-of-tables [[foo]] and returning the cleaned string and count of duplicate sections removed.
func deduplicateTomlTables(content string) (string, int) {
	lines := strings.Split(content, "\n")
	seenTables := make(map[string]bool)
	var output []string
	skipping := false
	dupes := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Array of tables [[foo]] can appear multiple times in valid TOML
		isArrayTable := strings.HasPrefix(trimmed, "[[") && strings.HasSuffix(trimmed, "]]")
		// Standard table [foo] cannot appear multiple times in valid TOML
		isSingleTable := !isArrayTable && strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")

		if isSingleTable {
			tableName := strings.TrimSpace(trimmed[1 : len(trimmed)-1])
			if seenTables[tableName] {
				skipping = true
				dupes++
				continue
			}
			skipping = false
			seenTables[tableName] = true
			output = append(output, line)
		} else if isArrayTable {
			skipping = false
			output = append(output, line)
		} else {
			if !skipping {
				output = append(output, line)
			}
		}
	}
	return strings.Join(output, "\n"), dupes
}

func copyHostConfig(realHome, profileCodexDir string) {
	hostConfig := filepath.Join(realHome, ".codex", "config.toml")
	destConfig := filepath.Join(profileCodexDir, "config.toml")
	hostHooksJSON := filepath.Join(realHome, ".codex", "hooks.json")
	destHooksJSON := filepath.Join(profileCodexDir, "hooks.json")

	if _, err := os.Stat(destConfig); os.IsNotExist(err) {
		if data, err := os.ReadFile(hostConfig); err == nil && len(data) > 0 {
			// Rewrite hook trust hashes keyed by host hooks.json path to the profile's hooks.json path
			rewritten := strings.ReplaceAll(string(data), hostHooksJSON, destHooksJSON)
			cleaned, _ := deduplicateTomlTables(rewritten)
			_ = os.WriteFile(destConfig, []byte(cleaned), 0644)
		}
	} else {
		// Migrate any existing hostHooksJSON paths and deduplicate tables
		if data, err := os.ReadFile(destConfig); err == nil {
			destStr := string(data)
			changed := false
			if strings.Contains(destStr, hostHooksJSON) {
				destStr = strings.ReplaceAll(destStr, hostHooksJSON, destHooksJSON)
				changed = true
			}
			cleaned, dupes := deduplicateTomlTables(destStr)
			if changed || dupes > 0 {
				_ = os.WriteFile(destConfig, []byte(cleaned), 0644)
			}
		}
	}
}

func ensureSidecarDaemons(realHome, profileCodexDir string) {
	cfgPath := filepath.Join(profileCodexDir, "config.toml")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return
	}
	cfgStr := string(data)
	if !strings.Contains(cfgStr, "127.0.0.1:8787") && !strings.Contains(cfgStr, "localhost:8787") && !strings.Contains(cfgStr, "model_provider = \"caveman\"") {
		return
	}

	// 1. Probe port 8787
	conn, err := net.DialTimeout("tcp", "127.0.0.1:8787", 250*time.Millisecond)
	if err == nil {
		conn.Close()
		return
	}

	// 2. Port is unreachable; locate caveman-proxy binary
	bin := filepath.Join(realHome, ".caveman", "bin", "caveman-proxy")
	if _, err := os.Stat(bin); err != nil {
		if path, err := exec.LookPath("caveman-proxy"); err == nil {
			bin = path
		} else {
			logger.Debug("[codex] caveman-proxy binary not found, cannot auto-start")
			return
		}
	}

	logger.Debug("[codex] Caveman proxy on 127.0.0.1:8787 is not reachable. Auto-starting %s...", bin)
	cmd := exec.Command(bin)
	cmd.Dir = realHome
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	if err := cmd.Start(); err != nil {
		logger.Debug("[codex] Failed to start caveman-proxy daemon: %v", err)
		return
	}

	// Wait up to 1.5s for the proxy to start listening
	deadline := time.Now().Add(1500 * time.Millisecond)
	for time.Now().Before(deadline) {
		time.Sleep(100 * time.Millisecond)
		conn, err := net.DialTimeout("tcp", "127.0.0.1:8787", 150*time.Millisecond)
		if err == nil {
			conn.Close()
			logger.Debug("[codex] caveman-proxy auto-started and listening on 127.0.0.1:8787")
			return
		}
	}
	logger.Debug("[codex] caveman-proxy did not become ready within timeout")
}

func bridgePluginsAndHooks(realHome, profileCodexDir string) {
	hostPlugins := filepath.Join(realHome, ".codex", "plugins")
	destPlugins := filepath.Join(profileCodexDir, "plugins")
	if fi, err := os.Stat(hostPlugins); err == nil && fi.IsDir() {
		if _, err := os.Lstat(destPlugins); os.IsNotExist(err) {
			_ = os.Symlink(hostPlugins, destPlugins)
		}
	}

	hostHooksDir := filepath.Join(realHome, ".codex", "hooks")
	destHooksDir := filepath.Join(profileCodexDir, "hooks")
	if fi, err := os.Stat(hostHooksDir); err == nil && fi.IsDir() {
		if _, err := os.Lstat(destHooksDir); os.IsNotExist(err) {
			_ = os.Symlink(hostHooksDir, destHooksDir)
		}
	}

	hostHooksJSON := filepath.Join(realHome, ".codex", "hooks.json")
	destHooksJSON := filepath.Join(profileCodexDir, "hooks.json")
	if fi, err := os.Stat(hostHooksJSON); err == nil && !fi.IsDir() {
		if _, err := os.Lstat(destHooksJSON); os.IsNotExist(err) {
			_ = os.Symlink(hostHooksJSON, destHooksJSON)
		}
	}
}

func bridgeCxStatusline(realHome, profileDir string) {
	hostConfig := filepath.Join(realHome, ".config", "cxstatusline")
	destConfig := filepath.Join(profileDir, ".config", "cxstatusline")
	if fi, err := os.Stat(hostConfig); err == nil && fi.IsDir() {
		lfi, err := os.Lstat(destConfig)
		if os.IsNotExist(err) {
			_ = os.MkdirAll(filepath.Dir(destConfig), 0755)
			_ = os.Symlink(hostConfig, destConfig)
		} else if err == nil && lfi.IsDir() && (lfi.Mode()&os.ModeSymlink == 0) {
			// If destConfig exists as a non-symlink directory, check if it only has an auto-generated settings.json or is empty.
			// In that case, replace it with a symlink to hostConfig so custom statusline settings are shared.
			entries, readErr := os.ReadDir(destConfig)
			if readErr == nil && (len(entries) == 0 || (len(entries) == 1 && entries[0].Name() == "settings.json")) {
				_ = os.RemoveAll(destConfig)
				_ = os.Symlink(hostConfig, destConfig)
			}
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

	// 4. Sidecar Check (Caveman / local proxy if configured)
	cfgPath := filepath.Join(codexDir, "config.toml")
	if cfgData, err := os.ReadFile(cfgPath); err == nil {
		cfgStr := string(cfgData)
		if strings.Contains(cfgStr, "127.0.0.1:8787") || strings.Contains(cfgStr, "localhost:8787") || strings.Contains(cfgStr, "model_provider = \"caveman\"") {
			conn, err := net.DialTimeout("tcp", "127.0.0.1:8787", 250*time.Millisecond)
			if err == nil {
				conn.Close()
				results = append(results, agents.DiagnosticResult{
					Category: "Sidecar",
					Status:   "OK",
					Message:  "Caveman proxy listening on 127.0.0.1:8787",
				})
			} else {
				results = append(results, agents.DiagnosticResult{
					Category: "Sidecar",
					Status:   "WARN",
					Message:  "Caveman proxy on 127.0.0.1:8787 unreachable (will auto-start on run)",
				})
			}
		}

		// 5. Hooks Trust Check
		realHome := config.RealHomeDir()
		hostHooksJSON := filepath.Join(realHome, ".codex", "hooks.json")
		if strings.Contains(cfgStr, hostHooksJSON) {
			copyHostConfig(realHome, codexDir)
			results = append(results, agents.DiagnosticResult{
				Category: "Hooks",
				Status:   "OK",
				Message:  "Auto-migrated host hook trust paths in config.toml to profile",
			})
		}

		// 6. TOML Integrity Check (Deduplication)
		if _, dupes := deduplicateTomlTables(cfgStr); dupes > 0 {
			copyHostConfig(realHome, codexDir)
			results = append(results, agents.DiagnosticResult{
				Category: "Config",
				Status:   "OK",
				Message:  fmt.Sprintf("Auto-repaired %d duplicate table key(s) in config.toml", dupes),
			})
		}
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

	rateLimitsList := getAllLatestCodexRateLimits(profileName, profileDir)
	var windows []usage.LimitWindow
	var credits string

	for _, rl := range rateLimitsList {
		if rl == nil {
			continue
		}

		cat := "Codex"
		if rl.LimitName != "" {
			cat = rl.LimitName
		} else if rl.LimitID != "" && rl.LimitID != "codex" {
			cat = rl.LimitID
		}

		makeWindow := func(w *codexRateLimitWindow, defaultName string) *usage.LimitWindow {
			if w == nil {
				return nil
			}
			var resetsAt time.Time
			var resetsIn time.Duration
			if w.ResetsAt > 0 {
				resetsAt = time.Unix(w.ResetsAt, 0)
				resetsIn = time.Until(resetsAt)
				if resetsIn < 0 {
					resetsIn = 0
				}
			}
			remaining := int(math.Round(100.0 - w.UsedPercent))
			if remaining < 0 {
				remaining = 0
			} else if remaining > 100 {
				remaining = 100
			}
			if !resetsAt.IsZero() && time.Now().After(resetsAt) {
				remaining = 100
				resetsIn = 0
			}

			name := defaultName
			if w.WindowMinutes > 0 {
				wm := formatWindowMinutes(w.WindowMinutes)
				if strings.Contains(strings.ToLower(wm), "5h") || strings.Contains(strings.ToLower(wm), "hour") {
					name = "5h Limit"
				} else if strings.Contains(strings.ToLower(wm), "week") || strings.Contains(strings.ToLower(wm), "7d") {
					name = "Weekly Limit"
				} else {
					name = fmt.Sprintf("%s Limit", wm)
				}
			}

			return &usage.LimitWindow{
				Category:     cat,
				Name:         name,
				RemainingPct: remaining,
				ResetsAt:     resetsAt,
				ResetsIn:     resetsIn,
			}
		}

		seenNames := make(map[string]bool)
		addWin := func(w *codexRateLimitWindow, defaultName string) {
			if win := makeWindow(w, defaultName); win != nil {
				if !seenNames[win.Name] {
					seenNames[win.Name] = true
					windows = append(windows, *win)
				}
			}
		}

		if rl.Primary != nil {
			defaultName := "5h Limit"
			if rl.Primary.WindowMinutes >= 10080 {
				defaultName = "Weekly Limit"
			}
			addWin(rl.Primary, defaultName)
		}

		if rl.Secondary != nil {
			addWin(rl.Secondary, "Weekly Limit")
		}

		if rl.Credits != nil && credits == "" {
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

func parseSessionFileAllRateLimits(filePath string) map[string]*codexRateLimits {
	f, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 10*1024*1024)
	limits := make(map[string]*codexRateLimits)
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
				key := rl.LimitName
				if key == "" {
					key = rl.LimitID
				}
				if key == "" {
					key = "codex"
				}
				limits[key] = rl
			}
		}
	}
	if err := scanner.Err(); err != nil {
		logger.Debug("[codex] scanner error reading session file %s: %v", filePath, err)
	}
	return limits
}

func getAllLatestCodexRateLimits(profileName, profileDir string) []*codexRateLimits {
	// 1. Check profileDir/.codex/sessions
	profileSessionsDir := filepath.Join(profileDir, ".codex", "sessions")
	files := findRecentSessionFiles(profileSessionsDir)

	collected := make(map[string]*codexRateLimits)
	maxCheck := 20
	limit := maxCheck
	if len(files) < limit {
		limit = len(files)
	}
	for i := 0; i < limit; i++ {
		fileLimits := parseSessionFileAllRateLimits(files[i].path)
		for k, rl := range fileLimits {
			if _, exists := collected[k]; !exists {
				collected[k] = rl
			}
		}
		if len(collected) >= 2 {
			break
		}
	}

	// 2. If fewer than 2 models found, and profile is eligible for seeding (e.g. primary profile),
	// also check host ~/.codex/sessions
	if len(collected) < 2 && isProfileEligibleForSeeding(profileName) {
		hostSessionsDir := filepath.Join(config.RealHomeDir(), ".codex", "sessions")
		hostFiles := findRecentSessionFiles(hostSessionsDir)
		hostLimit := maxCheck
		if len(hostFiles) < hostLimit {
			hostLimit = len(hostFiles)
		}
		for i := 0; i < hostLimit; i++ {
			fileLimits := parseSessionFileAllRateLimits(hostFiles[i].path)
			for k, rl := range fileLimits {
				if _, exists := collected[k]; !exists {
					collected[k] = rl
				}
			}
			if len(collected) >= 2 {
				break
			}
		}
	}

	var results []*codexRateLimits
	// Guarantee stable ordering: "codex" (default) first, then Spark / others
	if codexRL, ok := collected["codex"]; ok {
		results = append(results, codexRL)
	}
	var otherKeys []string
	for k := range collected {
		if k != "codex" {
			otherKeys = append(otherKeys, k)
		}
	}
	sort.Strings(otherKeys)
	for _, k := range otherKeys {
		results = append(results, collected[k])
	}
	return results
}
