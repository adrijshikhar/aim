package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/aim-cli/aim/internal/agents"
)

type Runner struct{}

func NewRunner() *Runner {
	return &Runner{}
}

func (r *Runner) Run(ctx context.Context, launch agents.LaunchEnv, extraArgs []string) (int, error) {
	args := append(launch.Args, extraArgs...)
	cmd := exec.CommandContext(ctx, launch.BinaryPath, args...)
	cmd.Dir = launch.WorkingDir

	var envSlice []string
	for k, v := range launch.Env {
		envSlice = append(envSlice, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = envSlice

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	profileName := launch.Env["AIM_PROFILE"]
	agentName := launch.Env["AIM_AGENT"]
	if agentName != "" && profileName != "" {
		fmt.Fprintf(os.Stdout, "\033]0;AIM: [%s] %s\007", agentName, profileName)
	}

	if err := cmd.Start(); err != nil {
		return 1, err
	}

	stopSignals := setupSignalForwarding(cmd.Process)
	defer stopSignals()

	err := cmd.Wait()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if ws, ok := exitErr.ProcessState.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
				return 128 + int(ws.Signal()), nil
			}
			return exitErr.ExitCode(), nil
		}
		return 1, err
	}
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
	return r.Run(ctx, launch, nil)
}
