package agents

import (
	"os"
	"os/exec"
	"path/filepath"
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
func (b *BaseAdapter) Aliases() []string   { return b.aliases }

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
		if key == "SSH_CONNECTION" || key == "SSH_CLIENT" || key == "SSH_TTY" ||
			key == "GEMINI_CLI_HOME" || key == "CODEX_HOME" || key == "CLAUDE_CONFIG_DIR" ||
			key == "HOME" || key == "AIM_AGENT" || key == "AIM_PROFILE" || key == "AIM_HOME" ||
			key == "AIM_SESSION_ID" || key == "CLAUDE_CODE_OAUTH_TOKEN" ||
			key == "ANTHROPIC_API_KEY" || key == "CLAUDE_CODE_OAUTH_REFRESH_TOKEN" {
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
