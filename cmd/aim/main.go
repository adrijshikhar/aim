package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/diagnostics"
	"github.com/aim-cli/aim/internal/logger"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/telemetry"
	"github.com/aim-cli/aim/internal/updater"
	"github.com/mattn/go-isatty"
	"github.com/posthog/posthog-go"
)

func dispatchWithContext(ctx context.Context, args []string, reg *agents.Registry, pm *profile.ProfileManager) (exitCode int) {
	startTime := time.Now()
	if len(args) > 0 && args[0] != "__complete" {
		telemetry.MaybeDisplayFirstRunNotice(config.BaseDir())
	}
	logger.Init(config.BaseDir())
	defer logger.Close()

	cfg, _ := config.LoadConfig()
	updater.MaybeTriggerBackgroundCheck(GetVersion(), config.CacheDir())
	if cfg != nil && cfg.Debug {
		env := strings.TrimSpace(strings.ToLower(os.Getenv("AIM_DEBUG")))
		if env != "0" && env != "false" && env != "no" && env != "off" {
			logger.SetDebug(true)
		}
	}

	telemetry.InitPostHog(cfg)
	defer func() {
		if err := telemetry.ClosePostHog(); err != nil {
			logger.Debug("error closing PostHog client: %v", err)
		}
	}()

	telemetry.InitPostHogLogs(cfg)
	defer func() {
		if err := telemetry.ClosePostHogLogs(); err != nil {
			logger.Debug("error closing PostHog log exporter: %v", err)
		}
	}()
	if len(args) == 0 || args[0] != "__complete" {
		telemetry.LogPostHogInfo(ctx, "command_dispatch_started", nil)
	}

	defer func() {
		if r := recover(); r != nil {
			if client := telemetry.PostHogClient(); client != nil {
				client.Enqueue(posthog.NewDefaultException(
					time.Now(),
					telemetry.AnonymousMachineID(config.BaseDir()),
					"UnhandledPanic",
					"AIM encountered an unexpected panic",
				))
			}

			stack := debug.Stack()
			home, _ := os.UserHomeDir()
			issueURL, report := diagnostics.HandlePanic(r, stack, home, Version, Commit)
			logger.Error("Unhandled panic in AIM: %v\n%s", r, report)
			fmt.Fprintf(os.Stderr, "\n\x1b[31;1m⚠️  AIM encountered an unexpected crash: %v\x1b[0m\n", r)
			fmt.Fprintf(os.Stderr, "A sanitized crash report has been logged to %s\n\n", logger.LogFilePath())
			fmt.Fprintf(os.Stderr, "Help improve AIM by reporting this issue:\n\x1b[36m%s\x1b[0m\n\n", issueURL)
			exitCode = 1
		}
	}()

	rootCmd := newRootCmd(reg, pm)

	normalizedArgs := args
	if len(normalizedArgs) > 0 && normalizedArgs[0] == "__complete" {
		if len(normalizedArgs) == 1 {
			normalizedArgs = []string{"__complete", ""}
		} else {
			last := normalizedArgs[len(normalizedArgs)-1]
			if last != "" {
				foundCmd, _, err := rootCmd.Find([]string{last})
				isSub := err == nil && foundCmd != nil && foundCmd != rootCmd
				isAg := false
				if reg != nil {
					_, err := reg.Get(last)
					isAg = err == nil
				}
				if isSub || isAg {
					normalizedArgs = append(normalizedArgs, "")
				}
			}
		}
	}

	rootCmd.SetArgs(normalizedArgs)

	if ctx == nil {
		ctx = context.Background()
	}
	err := rootCmd.ExecuteContext(ctx)
	notifyUpdate(args)

	code := 0
	if err != nil {
		var exitErr *ExitError
		if errors.As(err, &exitErr) {
			code = exitErr.Code
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			code = 1
		}
	}

	if len(args) == 0 || args[0] != "__complete" {
		telemetry.LogPostHogInfo(ctx, "command_dispatch_completed", map[string]any{
			"exit_code": code,
		})
	}
	recordCommandTelemetry(args, code, time.Since(startTime), cfg)
	return code
}

func recordCommandTelemetry(args []string, code int, duration time.Duration, cfg *config.Config) {
	if len(args) > 0 && args[0] == "__complete" {
		return
	}
	cmdName := "tui"
	if len(args) > 0 {
		if strings.HasPrefix(args[0], "-") {
			cmdName = "root"
		} else {
			cmdName = args[0]
		}
	}
	cleanCmd, cleanAgent := telemetry.SanitizeCommand(cmdName, args)
	telClient := telemetry.NewClient(config.BaseDir(), config.CacheDir(), Version, cfg)
	telClient.Track(telemetry.EventCommandExecuted, map[string]any{
		"command":         cleanCmd,
		"agent":           cleanAgent,
		"exit_code":       code,
		"duration_bucket": telemetry.DurationBucket(duration),
	})
	_ = telClient.Close()
}

func dispatch(args []string, reg *agents.Registry, pm *profile.ProfileManager) int {
	return dispatchWithContext(context.Background(), args, reg, pm)
}

func notifyUpdate(args []string) {
	if len(args) > 0 {
		if args[0] == "__complete" || args[0] == "update" {
			return
		}
	}
	for _, a := range args {
		if a == "--json" || a == "-j" {
			return
		}
	}
	updater.AwaitBackgroundCheck(300 * time.Millisecond)
	cached := updater.CheckCached(Version, config.CacheDir())
	if cached != nil && cached.UpdateAvailable {
		cfg, _ := config.LoadConfig()
		mode := cfg.GetAutoUpdate()
		switch mode {
		case config.AutoUpdateOff:
			return
		case config.AutoUpdateAuto:
			execPath, err := os.Executable()
			if err == nil {
				cmd := exec.Command(execPath, "update", "--background")
				_ = cmd.Start()
			}
		case config.AutoUpdatePrompt:
			if isatty.IsTerminal(os.Stdout.Fd()) && isatty.IsTerminal(os.Stdin.Fd()) {
				fmt.Fprintf(os.Stderr, "\n▲ A new version of aim is available: %s → %s\n  Update now? [y/N]: ", Version, cached.LatestVersion)
				reader := bufio.NewReader(os.Stdin)
				input, err := reader.ReadString('\n')
				if err == nil {
					input = strings.TrimSpace(strings.ToLower(input))
					if input == "y" || input == "yes" {
						_ = runUpdate(context.Background(), false, false, false)
						return
					}
				}
			} else {
				fmt.Fprintf(os.Stderr, "\nA new version of aim is available: %s → %s (run 'aim update')\n", Version, cached.LatestVersion)
			}
		case config.AutoUpdateNotify:
			fmt.Fprintf(os.Stderr, "\nA new version of aim is available: %s → %s (run 'aim update')\n", Version, cached.LatestVersion)
		}
	}
}

func main() {
	reg := defaultRegistry()

	pm := profile.NewProfileManager(config.BaseDir())

	os.Exit(dispatch(os.Args[1:], reg, pm))
}
