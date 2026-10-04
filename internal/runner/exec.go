package runner

import (
	"context"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/logger"
	"github.com/charmbracelet/x/term"
)

type Runner struct {
	adapter agents.AgentAdapter
}

func NewRunner(adapter ...agents.AgentAdapter) *Runner {
	var a agents.AgentAdapter
	if len(adapter) > 0 {
		a = adapter[0]
	}
	return &Runner{adapter: a}
}

func (r *Runner) Run(ctx context.Context, launch agents.LaunchEnv, extraArgs []string) (int, error) {
	var oldState *term.State
	isTerm := term.IsTerminal(os.Stdin.Fd())
	if isTerm {
		if state, err := term.GetState(os.Stdin.Fd()); err == nil {
			oldState = state
			defer func() {
				if oldState != nil {
					_ = term.Restore(os.Stdin.Fd(), oldState)
				}
			}()
		}
	}

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

	// Dispatch optional PostLauncher hook (e.g. background token synchronization or harvesting)
	var postLauncher agents.PostLauncher
	if r != nil && r.adapter != nil {
		if pl, ok := r.adapter.(agents.PostLauncher); ok {
			postLauncher = pl
		}
	}
	if postLauncher == nil && launch.PostLauncher != nil {
		postLauncher = launch.PostLauncher
	}
	var postWg sync.WaitGroup
	if postLauncher != nil {
		postCtx, postCancel := context.WithCancel(ctx)
		defer func() {
			postCancel()
			done := make(chan struct{}, 1)
			go func() {
				postWg.Wait()
				close(done)
			}()
			select {
			case <-done:
			case <-time.After(500 * time.Millisecond):
				logger.Debug("[runner] PostLaunch shutdown timed out after 500ms")
			}
		}()
		postWg.Add(1)
		go func() {
			defer postWg.Done()
			postLauncher.PostLaunch(postCtx, profileName, launch.Env["HOME"])
		}()
	}

	if err := cmd.Start(); err != nil {
		logger.Debug("[runner] Process start failed: %v", err)
		return 1, err
	}
	logger.Debug("[runner] Process started with PID %d", cmd.Process.Pid)

	stopSignals := setupSignalForwarding(cmd.Process, isTerm)
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

// BuildEnv constructs the execution environment by filtering out sensitive/managed variables
// (SSH variables, HOME, AIM_* variables, agent-specific ambient tokens) unless explicitly provided in launchEnv,
// and applying overrides. This prevents host tokens from inadvertently leaking into profile executions.
func BuildEnv(environ []string, launchEnv map[string]string) []string {
	return agents.BuildEnv(environ, launchEnv)
}
