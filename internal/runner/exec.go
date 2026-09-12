package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"

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

	cmd.Env = os.Environ()
	for k, v := range launch.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	profileName := launch.Env["AIM_PROFILE"]
	agentName := launch.Env["AIM_AGENT"]
	if agentName != "" && profileName != "" {
		fmt.Fprintf(os.Stdout, "\033]0;AIM: [%s] %s\007", agentName, profileName)
	}

	if agentName != "" {
		cfg, _ := config.LoadConfig()
		var customServices []string
		if cfg != nil {
			customServices = cfg.CustomIgnoredKeychains
		}
		_ = profile.PurgeIgnoredKeychains(agentName, customServices...)
		defer func() {
			_ = profile.PurgeIgnoredKeychains(agentName, customServices...)
		}()
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
