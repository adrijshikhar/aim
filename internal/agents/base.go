package agents

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/aim-cli/aim/internal/config"
)

// BaseAdapter provides common metadata, binary resolution, and environment isolation logic
// that all concrete AgentAdapter implementations can embed.
type BaseAdapter struct {
	name        string
	displayName string
	binaryName  string
	aliases     []string
	extraPaths  []string
}

// NewBaseAdapter constructs a BaseAdapter with the given metadata and optional extra search paths.
func NewBaseAdapter(name, displayName, binaryName string, aliases []string, extraPaths ...string) BaseAdapter {
	return BaseAdapter{
		name:        name,
		displayName: displayName,
		binaryName:  binaryName,
		aliases:     aliases,
		extraPaths:  extraPaths,
	}
}

func (b *BaseAdapter) Name() string        { return b.name }
func (b *BaseAdapter) DisplayName() string { return b.displayName }
func (b *BaseAdapter) BinaryName() string  { return b.binaryName }

// Aliases returns a copy of the adapter's CLI aliases.
func (b *BaseAdapter) Aliases() []string { return slices.Clone(b.aliases) }

// ResolveBinary locates the agent's executable on the system, checking PATH first and falling back
// to ~/.local/bin, /opt/homebrew/bin, /usr/local/bin, and any adapter-specific extra paths.
func (b *BaseAdapter) ResolveBinary() string {
	if bin, err := exec.LookPath(b.binaryName); err == nil {
		return bin
	}
	realHome := config.RealHomeDir()
	fallbacks := append([]string{
		filepath.Join(realHome, ".local", "bin", b.binaryName),
		filepath.Join("/opt/homebrew/bin", b.binaryName),
		filepath.Join("/usr/local/bin", b.binaryName),
	}, b.extraPaths...)

	for _, fb := range fallbacks {
		if _, err := os.Stat(fb); err == nil {
			return fb
		}
	}
	return b.binaryName
}

// BaseLaunchEnv constructs the base environment map (HOME, AIM_AGENT, AIM_PROFILE, storage env, and overrides)
// suitable for populating agents.LaunchEnv.Env.
func (b *BaseAdapter) BaseLaunchEnv(profileName, profileDir string, overrides map[string]string) map[string]string {
	launchEnv := config.StorageEnv()
	launchEnv["HOME"] = profileDir
	launchEnv["AIM_AGENT"] = b.name
	launchEnv["AIM_PROFILE"] = profileName
	for k, v := range overrides {
		launchEnv[k] = v
	}
	return launchEnv
}

// BuildCleanEnv constructs the full process environment as a slice of KEY=VALUE strings,
// applying environment isolation and filtering ambient tokens.
func (b *BaseAdapter) BuildCleanEnv(profileName, profileDir string, overrides map[string]string) []string {
	launchEnv := b.BaseLaunchEnv(profileName, profileDir, overrides)
	return BuildEnv(os.Environ(), launchEnv)
}

var isolatedEnvKeys = map[string]struct{}{
	"SSH_CONNECTION":                  {},
	"SSH_CLIENT":                      {},
	"SSH_TTY":                         {},
	"GEMINI_CLI_HOME":                 {},
	"CODEX_HOME":                      {},
	"CLAUDE_CONFIG_DIR":               {},
	"HOME":                            {},
	"AIM_AGENT":                       {},
	"AIM_PROFILE":                     {},
	"AIM_HOME":                        {},
	"AIM_SESSION_ID":                  {},
	"CLAUDE_CODE_OAUTH_TOKEN":         {},
	"CLAUDE_CODE_OAUTH_REFRESH_TOKEN": {},
	"ANTHROPIC_API_KEY":               {},
	"ANTHROPIC_AUTH_TOKEN":            {},
	"OPENAI_API_KEY":                  {},
	"CODEX_API_KEY":                   {},
	"GEMINI_API_KEY":                  {},
	"GOOGLE_API_KEY":                  {},
	"GOOGLE_APPLICATION_CREDENTIALS":  {},
	"GOOGLE_OAUTH_ACCESS_TOKEN":       {},
	"CURSOR_API_KEY":                  {},
	"CURSOR_AUTH_TOKEN":               {},
	"XDG_CONFIG_HOME":                 {},
	"XDG_DATA_HOME":                   {},
	"XDG_CACHE_HOME":                  {},
	"XDG_STATE_HOME":                  {},
}

// BuildEnv constructs the execution environment by filtering out sensitive/managed variables
// (SSH variables, HOME, AIM_* variables, agent-specific ambient tokens) unless explicitly provided in launchEnv,
// and applying overrides. This prevents host tokens from inadvertently leaking into profile executions.
func BuildEnv(environ []string, launchEnv map[string]string) []string {
	env := make([]string, 0, len(environ)+len(launchEnv))
	for _, e := range environ {
		idx := strings.IndexByte(e, '=')
		if idx == -1 {
			continue
		}
		key := e[:idx]
		if _, stripped := isolatedEnvKeys[key]; stripped {
			continue
		}
		if _, overridden := launchEnv[key]; overridden {
			continue
		}
		env = append(env, e)
	}
	for k, v := range launchEnv {
		env = append(env, k+"="+v)
	}
	return env
}
