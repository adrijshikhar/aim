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
	"github.com/aim-cli/aim/internal/config"
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

	if agentName != "" {
		cfg, _ := config.LoadConfig()
		var customServices []string
		if cfg != nil {
			customServices = cfg.CustomIgnoredKeychains
		}

		profileDir := launch.Env["HOME"]
		// Check whether the profile is authenticated in keyring bypass mode.
		// For agy, if SSH_CONNECTION is omitted, the session is unauthenticated, expired, or logging in.
		hasCreds := false
		if profileDir != "" {
			if _, hasSSH := launch.Env["SSH_CONNECTION"]; hasSSH {
				hasCreds = true
			} else if agentName != "agy" {
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

		// Only unauthenticated sessions interact with the macOS Keychain.
		// Authenticated profiles run with SSH_CONNECTION (keyring bypass mode) and never touch the Keychain.
		// Skipping purge for authenticated profiles prevents destroying active sessions or concurrent logins!
		if !hasCreds {
			_ = profile.PurgeIgnoredKeychains(agentName, customServices...)

			// Start background watcher that harvests the token into the profile immediately
			// once the user completes authentication in the browser, and immediately purges
			// the token from the host Keychain so it never lingers or races with other profiles.
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
							_ = profile.PurgeIgnoredKeychains(agentName, customServices...)
							logger.Debug("[runner] Successfully harvested token and purged host keychain during active session")
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
				_ = profile.PurgeIgnoredKeychains(agentName, customServices...)
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
// (SSH variables, HOME, AIM_* variables) unless explicitly provided in launchEnv, and applying overrides.
// This prevents the host environment from leaking SSH connection variables that would suppress browser auto-open.
func BuildEnv(environ []string, launchEnv map[string]string) []string {
	env := make([]string, 0, len(environ)+len(launchEnv))
	for _, e := range environ {
		idx := strings.IndexByte(e, '=')
		if idx == -1 {
			continue
		}
		key := e[:idx]
		if key == "SSH_CONNECTION" || key == "SSH_CLIENT" || key == "SSH_TTY" || key == "GEMINI_CLI_HOME" || key == "HOME" || key == "AIM_AGENT" || key == "AIM_PROFILE" || key == "AIM_HOME" || key == "AIM_SESSION_ID" {
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
