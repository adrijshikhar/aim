package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/logger"
	"github.com/aim-cli/aim/internal/profile"
)

type Runner struct{}

func NewRunner() *Runner {
	return &Runner{}
}

func (r *Runner) Run(ctx context.Context, launch agents.LaunchEnv, extraArgs []string) (int, error) {
	args := append(launch.Args, extraArgs...)
	logger.Debug("[runner] Executing binary %s with %d args (cwd: %s)", launch.BinaryPath, len(args), launch.WorkingDir)
	logger.Debug("[runner] Environment: HOME=%s, AIM_PROFILE=%s, AIM_AGENT=%s", launch.Env["HOME"], launch.Env["AIM_PROFILE"], launch.Env["AIM_AGENT"])

	cmd := exec.CommandContext(ctx, launch.BinaryPath, args...)
	cmd.Dir = launch.WorkingDir

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Env = BuildEnv(os.Environ(), launch.Env)

	profileName := launch.Env["AIM_PROFILE"]
	agentName := launch.Env["AIM_AGENT"]
	if agentName != "" && profileName != "" {
		sessionID := launch.Env["AIM_SESSION_ID"]
		if sessionID == "" {
			sessionID = ExtractSessionID(args)
		}
		_ = SetTerminalTitle(os.Stdout, FormatTitle(agentName, profileName, sessionID))
		defer ResetTerminalTitle(os.Stdout)
	}

	// Keyring bypass and keychain harvesting for agents with global shared keychains (e.g. agy).
	// Profiles use file-based credentials and never purge host keychains.
	if agentName == "agy" {
		profileDir := launch.Env["HOME"]
		hasCreds := false
		if profileDir != "" {
			if _, hasSSH := launch.Env["SSH_CONNECTION"]; hasSSH {
				hasCreds = true
			} else {
				tokenPath := filepath.Join(profileDir, ".gemini", "antigravity-cli", "antigravity-oauth-token")
				if fi, err := os.Stat(tokenPath); err == nil && fi.Size() > 0 {
					hasCreds = true
				} else {
					adcPath := filepath.Join(profileDir, ".config", "gcloud", "application_default_credentials.json")
					if fi, err := os.Stat(adcPath); err == nil && fi.Size() > 0 {
						hasCreds = true
					}
				}
			}
		}

		// Only unauthenticated sessions might need to harvest credentials from the macOS Keychain.
		// Authenticated profiles run with SSH_CONNECTION (keyring bypass mode) and never touch the Keychain.
		// Never purge host keychains, as doing so breaks host tools (CodexBar, host CLIs) and triggers security prompts.
		if !hasCreds {
			// Start background watcher that harvests the token into the profile once
			// the user completes authentication in the browser.
			stopWatcher := make(chan struct{})
			doneWatcher := make(chan struct{})
			go func() {
				defer close(doneWatcher)
				ticker := time.NewTicker(1 * time.Second)
				defer ticker.Stop()
				for {
					select {
					case <-stopWatcher:
						return
					case <-ticker.C:
						if profile.HarvestKeychainTokenToProfile(agentName, profileDir) {
							logger.Debug("[runner] Successfully harvested token during active session")
							return
						}
					}
				}
			}()

			defer func() {
				close(stopWatcher)
				<-doneWatcher
				if profileDir != "" {
					_ = profile.HarvestKeychainTokenToProfile(agentName, profileDir)
				}
			}()
		}
	}

	if err := cmd.Start(); err != nil {
		logger.Debug("[runner] Process start failed: %v", err)
		return 1, err
	}
	logger.Debug("[runner] Process started with PID %d", cmd.Process.Pid)

	stopSignals := setupSignalForwarding(cmd.Process)
	defer stopSignals()

	err := cmd.Wait()
	if err != nil {
		logger.Debug("[runner] Process wait returned: %v", err)
		if exitErr, ok := err.(*exec.ExitError); ok {
			if ws, ok := exitErr.ProcessState.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
				return 128 + int(ws.Signal()), nil
			}
			return exitErr.ExitCode(), nil
		}
		return 1, err
	}
	logger.Debug("[runner] Process exited cleanly (code 0)")
	return 0, nil
}

func (r *Runner) RunShell(ctx context.Context, launch agents.LaunchEnv) (int, error) {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	launch.BinaryPath = shell
	launch.Args = nil
	if launch.Env == nil {
		launch.Env = make(map[string]string)
	}
	profileName := launch.Env["AIM_PROFILE"]
	agentName := launch.Env["AIM_AGENT"]
	launch.Env["PS1"] = fmt.Sprintf("[aim:%s:%s] $ ", agentName, profileName)
	logger.Debug("[runner] Launching interactive shell %s for profile %q", shell, profileName)
	return r.Run(ctx, launch, nil)
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
