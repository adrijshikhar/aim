package gemini

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/usage"
)

// Compile-time assertion that Adapter implements agents.AgentAdapter.
var _ agents.AgentAdapter = (*Adapter)(nil)

type Adapter struct{}

// GeminiAdapter is an alias for Adapter.
type GeminiAdapter = Adapter

func NewAdapter() *Adapter { return &Adapter{} }

// NewGeminiAdapter creates a new GeminiAdapter.
func NewGeminiAdapter() *GeminiAdapter { return NewAdapter() }

func (a *Adapter) Name() string        { return "gemini" }
func (a *Adapter) DisplayName() string { return "Gemini CLI" }
func (a *Adapter) Aliases() []string   { return []string{"gemini-cli"} }
func (a *Adapter) BinaryName() string  { return "gemini" }

func (a *Adapter) TokenPath(profileDir string) string {
	return filepath.Join(profileDir, ".gemini", "gemini-oauth-token")
}

func (a *Adapter) HasCredentials(profileDir string) bool {
	p := a.TokenPath(profileDir)
	fi, err := os.Stat(p)
	return err == nil && !fi.IsDir() && fi.Size() > 0
}

func (a *Adapter) Login(ctx context.Context, profileName, profileDir string) error {
	return fmt.Errorf("gemini interactive login adapter coming soon")
}

func (a *Adapter) PrepareEnv(profileName, profileDir string) (agents.LaunchEnv, error) {
	bin, _ := exec.LookPath(a.BinaryName())
	if bin == "" {
		bin = a.BinaryName()
	}
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
	envMap["GEMINI_CLI_HOME"] = filepath.Join(profileDir, ".gemini")
	envMap["AIM_AGENT"] = a.Name()
	envMap["AIM_PROFILE"] = profileName
	delete(envMap, "AIM_SESSION_ID")
	cwd, _ := os.Getwd()
	return agents.LaunchEnv{
		BinaryPath: bin,
		Env:        envMap,
		WorkingDir: cwd,
	}, nil
}

func (a *Adapter) Doctor(ctx context.Context, profileName, profileDir string) []agents.DiagnosticResult {
	return []agents.DiagnosticResult{
		{Category: "Adapter", Status: "OK", Message: "Gemini adapter registered"},
	}
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

	return &usage.Report{
		Agent:     a.Name(),
		Profile:   profileName,
		Status:    usage.StatusOK,
		Summary:   "Active credentials",
		FetchedAt: time.Now(),
	}, nil
}
