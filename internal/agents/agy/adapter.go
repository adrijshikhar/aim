package agy

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/logger"
	"github.com/aim-cli/aim/internal/oauth"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/usage"
	"github.com/otiai10/copy"
)

var (
	// DefaultClientID and DefaultClientSecret can be injected at build time via -ldflags:
	// -X github.com/aim-cli/aim/internal/agents/agy.DefaultClientID=...
	// -X github.com/aim-cli/aim/internal/agents/agy.DefaultClientSecret=...
	DefaultClientID     = os.Getenv("AIM_AGY_CLIENT_ID")
	DefaultClientSecret = os.Getenv("AIM_AGY_CLIENT_SECRET")

	// ClientID and ClientSecret provide backward-compatible access to default OAuth credentials.
	ClientID     = DefaultClientID
	ClientSecret = DefaultClientSecret
)

func getOAuthCredentials() (string, string) {
	clientID := os.Getenv("AIM_AGY_CLIENT_ID")
	if clientID == "" {
		clientID = DefaultClientID
	}
	clientSecret := os.Getenv("AIM_AGY_CLIENT_SECRET")
	if clientSecret == "" {
		clientSecret = DefaultClientSecret
	}
	return clientID, clientSecret
}

// Compile-time assertion that Adapter implements agents.AgentAdapter.
var _ agents.AgentAdapter = (*Adapter)(nil)

type Adapter struct{}

// AntigravityAdapter is an alias for Adapter.
type AntigravityAdapter = Adapter

func NewAdapter() *Adapter {
	return &Adapter{}
}

// NewAntigravityAdapter creates a new AntigravityAdapter.
func NewAntigravityAdapter() *AntigravityAdapter {
	return NewAdapter()
}

func (a *Adapter) Name() string        { return "agy" }
func (a *Adapter) DisplayName() string { return "Antigravity CLI" }
func (a *Adapter) Aliases() []string   { return []string{"antigravity"} }
func (a *Adapter) BinaryName() string  { return "agy" }

func (a *Adapter) TokenPath(profileDir string) string {
	return filepath.Join(profileDir, ".gemini", "antigravity-cli", "antigravity-oauth-token")
}

func (a *Adapter) HasCredentials(profileDir string) bool {
	p := a.TokenPath(profileDir)
	data, err := os.ReadFile(p)
	if err == nil && len(data) > 0 {
		_, _ = validateAndRepairTokenJSON(p, data)
		logger.Debug("[agy] HasCredentials: true (valid token at %s)", p)
		return true
	}

	// Check if Google Application Default Credentials (ADC) exist
	adcPath := filepath.Join(profileDir, ".config", "gcloud", "application_default_credentials.json")
	if fi, err := os.Stat(adcPath); err == nil && !fi.IsDir() && fi.Size() > 0 {
		logger.Debug("[agy] HasCredentials: true (ADC at %s)", adcPath)
		return true
	}

	// Auto-seed credentials if eligible profile doesn't have credentials yet
	profileName := filepath.Base(profileDir)
	if a.SeedDefaultCredentials(profileName, profileDir) {
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() && fi.Size() > 0 {
			logger.Debug("[agy] HasCredentials: true (auto-seeded token at %s)", p)
			return true
		}
	}

	logger.Debug("[agy] HasCredentials: false (no token at %s or ADC at %s)", p, adcPath)
	return false
}

func isProfileEligibleForSeeding(profileName string) bool {
	switch profileName {
	case "personal", "default", "p", "me", "main":
		return true
	}

	cfg, err := config.LoadConfig()
	if err == nil && cfg != nil {
		if cfg.DefaultProfile != "" && cfg.DefaultProfile == profileName {
			return true
		}
		if len(cfg.Profiles) <= 1 {
			return true
		}
	}
	return false
}

func copyHostSettings(realHome, tokenDir string) {
	realSettings := filepath.Join(realHome, ".gemini", "antigravity-cli", "settings.json")
	destSettings := filepath.Join(tokenDir, "settings.json")
	if _, err := os.Stat(destSettings); os.IsNotExist(err) {
		if sData, err := os.ReadFile(realSettings); err == nil {
			_ = os.WriteFile(destSettings, sData, 0644)
		}
	}
}

// SeedDefaultCredentials copies the host's existing Antigravity CLI token to an
// eligible profile (such as "personal", "default", "p", or the configured default profile)
// if the profile doesn't have credentials yet.
func (a *Adapter) SeedDefaultCredentials(profileName, profileDir string) bool {
	if !isProfileEligibleForSeeding(profileName) {
		return false
	}
	p := a.TokenPath(profileDir)
	if fi, err := os.Stat(p); err == nil && fi.Size() > 0 {
		return false
	}
	realHome := config.RealHomeDir()
	realToken := filepath.Join(realHome, ".gemini", "antigravity-cli", "antigravity-oauth-token")

	// 1. Try copying token from host filesystem
	if filepath.Clean(realToken) != filepath.Clean(p) {
		if realFi, err := os.Stat(realToken); err == nil && !realFi.IsDir() {
			if rData, err := os.ReadFile(realToken); err == nil && len(rData) > 0 {
				perm := realFi.Mode().Perm()
				if perm == 0 {
					perm = 0600
				}
				_ = os.MkdirAll(filepath.Dir(p), 0700)
				if err := os.WriteFile(p, rData, perm); err == nil {
					logger.Debug("[agy] Seeded profile %q from host token file %s", profileName, realToken)
					copyHostSettings(realHome, filepath.Dir(p))
					return true
				}
			}
		}
	}

	// 2. Try harvesting token from host macOS Keychain
	if profile.HarvestKeychainTokenToProfile(a.Name(), profileDir) {
		logger.Debug("[agy] Seeded profile %q from host macOS Keychain", profileName)
		copyHostSettings(realHome, filepath.Dir(p))
		return true
	}

	return false
}

func validateAndRepairTokenJSON(path string, data []byte) ([]byte, bool) {
	var tok struct {
		Token struct {
			RefreshToken string `json:"refresh_token"`
			AccessToken  string `json:"access_token"`
		} `json:"token"`
		RefreshToken string `json:"refresh_token"`
		AccessToken  string `json:"access_token"`
	}
	if err := json.Unmarshal(data, &tok); err == nil {
		hasTok := tok.Token.RefreshToken != "" || tok.Token.AccessToken != "" ||
			tok.RefreshToken != "" || tok.AccessToken != ""
		return data, hasTok
	}
	// Attempt to repair trailing garbage or extra braces
	trimmed := strings.TrimSpace(string(data))
	for {
		lastIdx := strings.LastIndexByte(trimmed, '}')
		if lastIdx == -1 {
			break
		}
		candidate := []byte(trimmed[:lastIdx+1])
		if err := json.Unmarshal(candidate, &tok); err == nil {
			hasTok := tok.Token.RefreshToken != "" || tok.Token.AccessToken != "" ||
				tok.RefreshToken != "" || tok.AccessToken != ""
			if hasTok {
				_ = os.WriteFile(path, append(candidate, '\n'), 0600)
				return candidate, true
			}
		}
		trimmed = trimmed[:lastIdx]
	}
	return data, false
}

func (a *Adapter) Login(ctx context.Context, profileName, profileDir string) error {
	clientID, clientSecret := getOAuthCredentials()
	if clientID == "" || clientSecret == "" {
		bin, err := exec.LookPath(a.BinaryName())
		if err != nil {
			realHome := config.RealHomeDir()
			fallback := filepath.Join(realHome, ".local", "bin", a.BinaryName())
			if _, sErr := os.Stat(fallback); sErr == nil {
				bin = fallback
			} else {
				return fmt.Errorf("OAuth client credentials not configured and '%s' binary not found in PATH", a.BinaryName())
			}
		}
		realHome := config.RealHomeDir()
		_ = bridgeSharedState(realHome, profileDir)
		cmd := exec.CommandContext(ctx, bin)
		cmd.Dir = profileDir
		cmd.Env = append(os.Environ(),
			"HOME="+profileDir,
			"AIM_AGENT="+a.Name(),
			"AIM_PROFILE="+profileName,
			"SSH_CONNECTION=127.0.0.1 50000 127.0.0.1 22",
		)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
		_ = profile.HarvestKeychainTokenToProfile(a.Name(), profileDir)
		return nil
	}
	cfg := oauth.ProviderConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL:     "https://oauth2.googleapis.com/token",
		UserInfoURL:  "https://www.googleapis.com/oauth2/v2/userinfo",
		Scopes: []string{
			"openid",
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/cloud-platform",
		},
		ExtraAuthParams: map[string]string{
			"access_type": "offline",
			"prompt":      "consent",
		},
	}

	tok, err := oauth.Authenticate(ctx, cfg, true)
	if err != nil {
		return fmt.Errorf("oauth login failed: %w", err)
	}

	tokenDir := filepath.Join(profileDir, ".gemini", "antigravity-cli")
	if err := os.MkdirAll(tokenDir, 0700); err != nil {
		return err
	}

	payload := map[string]any{
		"token": map[string]any{
			"access_token":  tok.AccessToken,
			"token_type":    tok.TokenType,
			"refresh_token": tok.RefreshToken,
			"expiry":        tok.ExpiresAt.Format(time.RFC3339),
		},
		"auth_method": "consumer",
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}

	dest := a.TokenPath(profileDir)
	tmp := dest + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	if err := os.Rename(tmp, dest); err != nil {
		return err
	}

	realHome, err := os.UserHomeDir()
	if err == nil {
		realSettings := filepath.Join(realHome, ".gemini", "antigravity-cli", "settings.json")
		destSettings := filepath.Join(tokenDir, "settings.json")
		if _, err := os.Stat(destSettings); os.IsNotExist(err) {
			if sData, err := os.ReadFile(realSettings); err == nil {
				_ = os.WriteFile(destSettings, sData, 0644)
			}
		}
	}

	fmt.Printf("[Success] Authenticated profile '%s' as %s\n", profileName, tok.UserEmail)
	return nil
}

func (a *Adapter) PrepareEnv(profileName, profileDir string) (agents.LaunchEnv, error) {
	tokenDir := filepath.Join(profileDir, ".gemini", "antigravity-cli")
	if err := os.MkdirAll(tokenDir, 0700); err != nil {
		return agents.LaunchEnv{}, err
	}

	realHome := config.RealHomeDir()
	_ = bridgeSharedState(realHome, profileDir)

	// Auto-seed credentials for personal/default profile if missing in profileDir but available on host
	_ = a.SeedDefaultCredentials(profileName, profileDir)

	// Also copy settings.json from host if not present in profile
	copyHostSettings(realHome, tokenDir)

	bin, err := exec.LookPath(a.BinaryName())
	if err != nil {
		fallback := filepath.Join(realHome, ".local", "bin", a.BinaryName())
		if _, sErr := os.Stat(fallback); sErr == nil {
			bin = fallback
		} else {
			bin = a.BinaryName()
		}
	}
	logger.Debug("[agy] Resolved binary: %s", bin)

	envMap := make(map[string]string)
	for _, e := range os.Environ() {
		for i := 0; i < len(e); i++ {
			if e[i] == '=' {
				envMap[e[:i]] = e[i+1:]
				break
			}
		}
	}

	envMap["HOME"] = profileDir
	envMap["SSH_CONNECTION"] = "127.0.0.1 50000 127.0.0.1 22"
	delete(envMap, "SSH_CLIENT")
	delete(envMap, "SSH_TTY")
	envMap["AIM_AGENT"] = a.Name()
	envMap["AIM_PROFILE"] = profileName
	envMap["AIM_HOME"] = config.BaseDir()
	delete(envMap, "GEMINI_CLI_HOME")

	logger.Debug("[agy] Launch env: HOME=%s, AIM_AGENT=%s, AIM_PROFILE=%s, AIM_HOME=%s", profileDir, a.Name(), profileName, config.BaseDir())

	cwd, _ := os.Getwd()
	return agents.LaunchEnv{
		BinaryPath: bin,
		Env:        envMap,
		WorkingDir: cwd,
	}, nil
}

func bridgeSharedState(realHome, profileDir string) error {
	var sharedDir string
	realAgyDir := filepath.Join(realHome, ".gemini", "antigravity-cli")
	if _, err := os.Stat(realAgyDir); err == nil {
		sharedDir = realAgyDir
	} else {
		sharedDir = filepath.Join(config.BaseDir(), "shared", "antigravity-cli")
	}

	_ = os.MkdirAll(filepath.Join(sharedDir, "conversations"), 0755)
	_ = os.MkdirAll(filepath.Join(sharedDir, "brain"), 0755)

	tokenDir := filepath.Join(profileDir, ".gemini", "antigravity-cli")
	if err := os.MkdirAll(tokenDir, 0700); err != nil {
		return err
	}

	// 1. conversations
	pConv := filepath.Join(tokenDir, "conversations")
	sConv := filepath.Join(sharedDir, "conversations")
	bridgeDir(pConv, sConv)

	// 2. brain
	pBrain := filepath.Join(tokenDir, "brain")
	sBrain := filepath.Join(sharedDir, "brain")
	bridgeDir(pBrain, sBrain)

	// 3. conversation_summaries.db
	pDb := filepath.Join(tokenDir, "conversation_summaries.db")
	sDb := filepath.Join(sharedDir, "conversation_summaries.db")
	bridgeFile(pDb, sDb)

	// 4. history.jsonl
	pHist := filepath.Join(tokenDir, "history.jsonl")
	sHist := filepath.Join(sharedDir, "history.jsonl")
	bridgeFile(pHist, sHist)

	return nil
}

func bridgeDir(profilePath, sharedPath string) {
	fi, err := os.Lstat(profilePath)
	if err != nil {
		if os.IsNotExist(err) {
			_ = os.Symlink(sharedPath, profilePath)
		}
		return
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		return
	}
	// Migrate existing items to sharedPath
	entries, rErr := os.ReadDir(profilePath)
	if rErr == nil {
		for _, e := range entries {
			src := filepath.Join(profilePath, e.Name())
			dst := filepath.Join(sharedPath, e.Name())
			if _, sErr := os.Lstat(dst); os.IsNotExist(sErr) {
				if e.IsDir() {
					_ = copyDirectory(src, dst)
				} else {
					_ = copySingleFile(src, dst)
				}
			}
		}
	}
	if rmErr := os.RemoveAll(profilePath); rmErr == nil {
		_ = os.Symlink(sharedPath, profilePath)
	} else {
		sharedEntries, _ := os.ReadDir(sharedPath)
		for _, se := range sharedEntries {
			dst := filepath.Join(profilePath, se.Name())
			src := filepath.Join(sharedPath, se.Name())
			if _, dErr := os.Lstat(dst); os.IsNotExist(dErr) {
				_ = os.Symlink(src, dst)
			}
		}
	}
}

func bridgeFile(profileFile, sharedFile string) {
	fi, err := os.Lstat(profileFile)
	if err != nil {
		if os.IsNotExist(err) {
			if _, sErr := os.Stat(sharedFile); sErr == nil {
				_ = os.Symlink(sharedFile, profileFile)
			}
		}
		return
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		return
	}
	if _, sErr := os.Stat(sharedFile); os.IsNotExist(sErr) {
		_ = copySingleFile(profileFile, sharedFile)
	}
	_ = os.Remove(profileFile)
	_ = os.Symlink(sharedFile, profileFile)
}

func copySingleFile(src, dst string) error {
	return copy.Copy(src, dst)
}

func copyDirectory(src, dst string) error {
	return copy.Copy(src, dst, copy.Options{
		OnSymlink: func(src string) copy.SymlinkAction {
			return copy.Shallow
		},
	})
}

func (a *Adapter) Doctor(ctx context.Context, profileName, profileDir string) []agents.DiagnosticResult {
	var results []agents.DiagnosticResult
	bin, err := exec.LookPath(a.BinaryName())
	if err != nil {
		realHome, _ := os.UserHomeDir()
		fallback := filepath.Join(realHome, ".local", "bin", a.BinaryName())
		if _, sErr := os.Stat(fallback); sErr == nil {
			bin = fallback
			err = nil
		}
	}
	if err != nil {
		results = append(results, agents.DiagnosticResult{
			Category: "Binary",
			Status:   "WARN",
			Message:  fmt.Sprintf("'%s' not found in PATH", a.BinaryName()),
		})
	} else {
		results = append(results, agents.DiagnosticResult{
			Category: "Binary",
			Status:   "OK",
			Message:  fmt.Sprintf("Found %s at %s", a.BinaryName(), bin),
		})
	}

	_ = a.SeedDefaultCredentials(profileName, profileDir)
	tokFile := a.TokenPath(profileDir)
	if data, err := os.ReadFile(tokFile); err != nil {
		adcPath := filepath.Join(profileDir, ".config", "gcloud", "application_default_credentials.json")
		if adcFi, adcErr := os.Stat(adcPath); adcErr == nil && !adcFi.IsDir() && adcFi.Size() > 0 {
			results = append(results, agents.DiagnosticResult{
				Category: "Token",
				Status:   "OK",
				Message:  "Google Cloud Application Default Credentials (ADC) active",
			})
		} else {
			results = append(results, agents.DiagnosticResult{
				Category: "Token",
				Status:   "FAIL",
				Message:  fmt.Sprintf("Missing token file (run: aim login %s %s)", a.Name(), profileName),
			})
		}
	} else {
		cleanedData, valid := validateAndRepairTokenJSON(tokFile, data)
		var tok struct {
			Token struct {
				AccessToken  string `json:"access_token"`
				RefreshToken string `json:"refresh_token"`
				Expiry       string `json:"expiry"`
			} `json:"token"`
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
			Expiry       string `json:"expiry"`
		}
		refreshToken := ""
		expiryStr := ""
		if valid {
			_ = json.Unmarshal(cleanedData, &tok)
			refreshToken = tok.Token.RefreshToken
			if refreshToken == "" {
				refreshToken = tok.RefreshToken
			}
			expiryStr = tok.Token.Expiry
			if expiryStr == "" {
				expiryStr = tok.Expiry
			}
		}
		if !valid || refreshToken == "" {
			results = append(results, agents.DiagnosticResult{
				Category: "Token",
				Status:   "FAIL",
				Message:  "Corrupt token file; missing refresh_token",
			})
		} else {
			expiryTime, pErr := time.Parse(time.RFC3339, expiryStr)
			if pErr == nil {
				if time.Now().After(expiryTime) {
					results = append(results, agents.DiagnosticResult{
						Category: "Token",
						Status:   "OK",
						Message:  fmt.Sprintf("Access token expired (%s), refresh_token ready for auto-refresh", expiryTime.Format("15:04:05")),
					})
				} else {
					remaining := time.Until(expiryTime).Round(time.Minute)
					results = append(results, agents.DiagnosticResult{
						Category: "Token",
						Status:   "OK",
						Message:  fmt.Sprintf("Access token valid for %s (expires %s)", remaining, expiryTime.Format("15:04:05")),
					})
				}
			} else {
				results = append(results, agents.DiagnosticResult{
					Category: "Token",
					Status:   "OK",
					Message:  "Valid token file present",
				})
			}

			if fi, sErr := os.Stat(tokFile); sErr == nil {
				perm := fi.Mode().Perm()
				if perm == 0600 {
					results = append(results, agents.DiagnosticResult{
						Category: "Permissions",
						Status:   "OK",
						Message:  "Token file permissions 0600 OK",
					})
				} else {
					results = append(results, agents.DiagnosticResult{
						Category: "Permissions",
						Status:   "WARN",
						Message:  fmt.Sprintf("Token file permissions %#o (recommended 0600)", perm),
					})
				}
			}
		}
	}

	if profileDir != "" {
		gitconfig := filepath.Join(profileDir, ".gitconfig")
		if _, err := os.Lstat(gitconfig); err == nil {
			results = append(results, agents.DiagnosticResult{
				Category: "Dotfiles",
				Status:   "OK",
				Message:  ".gitconfig linked correctly",
			})
		} else {
			results = append(results, agents.DiagnosticResult{
				Category: "Dotfiles",
				Status:   "WARN",
				Message:  ".gitconfig not linked in profile",
			})
		}

		dbPath := filepath.Join(profileDir, ".gemini", "antigravity-cli", "conversation_summaries.db")
		if fi, err := os.Stat(dbPath); err == nil {
			results = append(results, agents.DiagnosticResult{
				Category: "Database",
				Status:   "OK",
				Message:  fmt.Sprintf("conversation_summaries.db healthy (%d bytes)", fi.Size()),
			})
		}
	}
	return results
}

func ParseUsageTSV(raw string) []usage.LimitWindow {
	var windows []usage.LimitWindow
	scanner := bufio.NewScanner(strings.NewReader(raw))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 3 {
			continue
		}
		category := parts[0]
		name := parts[1]
		pctStr := strings.TrimSuffix(strings.TrimSpace(parts[2]), "%")
		var pct int
		if strings.EqualFold(pctStr, "disabled") {
			pct = 0
		} else {
			var pErr error
			pct, pErr = strconv.Atoi(pctStr)
			if pErr != nil {
				continue
			}
		}

		var resetsAt time.Time
		var resetsIn time.Duration
		if len(parts) >= 4 {
			if t, tErr := time.Parse(time.RFC3339, strings.TrimSpace(parts[3])); tErr == nil {
				resetsAt = t
				resetsIn = time.Until(t)
				if resetsIn < 0 {
					resetsIn = 0
				}
			}
		}

		windows = append(windows, usage.LimitWindow{
			Category:     category,
			Name:         name,
			RemainingPct: pct,
			ResetsAt:     resetsAt,
			ResetsIn:     resetsIn,
		})
	}
	return windows
}

func (a *Adapter) GetUsage(ctx context.Context, profileName, profileDir string) (*usage.Report, error) {
	_ = a.SeedDefaultCredentials(profileName, profileDir)
	if !a.HasCredentials(profileDir) {
		return &usage.Report{
			Agent:     a.Name(),
			Profile:   profileName,
			Status:    usage.StatusUnknown,
			FetchedAt: time.Now(),
			Error:     "no credentials",
		}, nil
	}

	bin, err := exec.LookPath(a.BinaryName())
	if err != nil {
		realHome, _ := os.UserHomeDir()
		fallback := filepath.Join(realHome, ".local", "bin", a.BinaryName())
		if _, sErr := os.Stat(fallback); sErr == nil {
			bin = fallback
			err = nil
		}
	}
	if err != nil {
		return &usage.Report{
			Agent:     a.Name(),
			Profile:   profileName,
			Status:    usage.StatusUnknown,
			FetchedAt: time.Now(),
			Error:     "binary not found",
		}, nil
	}

	env := append(os.Environ(),
		"HOME="+profileDir,
		"SSH_CONNECTION=127.0.0.1 50000 127.0.0.1 22",
		"AIM_AGENT="+a.Name(),
		"AIM_PROFILE="+profileName,
	)

	var (
		usageOut   []byte
		usageErr   error
		creditsOut []byte
		creditsErr error
		wg         sync.WaitGroup
	)

	wg.Add(2)
	go func() {
		defer wg.Done()
		cmd := exec.CommandContext(ctx, bin, "-p", "/usage")
		cmd.Env = env
		usageOut, usageErr = cmd.Output()
	}()

	go func() {
		defer wg.Done()
		cmd := exec.CommandContext(ctx, bin, "-p", "/credits")
		cmd.Env = env
		creditsOut, creditsErr = cmd.Output()
	}()

	wg.Wait()

	if usageErr != nil {
		// Fallback: check if local conversation db exists
		dbPath := filepath.Join(profileDir, ".gemini", "antigravity-cli", "conversation_summaries.db")
		summary := "Offline"
		if fi, sErr := os.Stat(dbPath); sErr == nil && fi.Size() > 0 {
			summary = "Offline (local session cache present)"
		}
		return &usage.Report{
			Agent:     a.Name(),
			Profile:   profileName,
			Status:    usage.StatusUnknown,
			Summary:   summary,
			FetchedAt: time.Now(),
			Error:     usageErr.Error(),
		}, nil
	}

	windows := ParseUsageTSV(string(usageOut))
	status := usage.CalculateStatus(windows)

	credits := "0"
	if creditsErr == nil {
		for _, line := range strings.Split(string(creditsOut), "\n") {
			parts := strings.Split(line, "\t")
			if len(parts) >= 2 && strings.Contains(strings.ToLower(parts[0]), "remaining") {
				credits = strings.TrimSpace(parts[1])
				break
			}
		}
	}

	report := &usage.Report{
		Agent:     a.Name(),
		Profile:   profileName,
		Status:    status,
		Windows:   windows,
		Credits:   credits,
		FetchedAt: time.Now(),
	}

	// Build summary
	var parts []string
	if pw := report.PrimaryWindow(); pw != nil {
		parts = append(parts, usage.FormatWindowSummary(pw))
	}
	if ww := report.WeeklyWindow(); ww != nil && (report.PrimaryWindow() == nil || ww.Name != report.PrimaryWindow().Name) {
		parts = append(parts, usage.FormatWindowSummary(ww))
	}
	report.Summary = strings.Join(parts, ", ")
	if len(report.Windows) == 0 && report.Summary == "" {
		report.Summary = "0 windows parsed"
	}

	return report, nil
}
