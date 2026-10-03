package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/diagnostics"
	"github.com/aim-cli/aim/internal/logger"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/telemetry"
	"github.com/aim-cli/aim/internal/updater"
	"time"
)

func dispatchWithContext(ctx context.Context, args []string, reg *agents.Registry, pm *profile.ProfileManager) (exitCode int) {
	startTime := time.Now()
	telemetry.MaybeDisplayFirstRunNotice(config.BaseDir())
	logger.Init(config.BaseDir())
	defer logger.Close()

	defer func() {
		if r := recover(); r != nil {
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

	cfg, _ := config.LoadConfig()
	if cfg != nil && cfg.Debug {
		env := strings.TrimSpace(strings.ToLower(os.Getenv("AIM_DEBUG")))
		if env != "0" && env != "false" && env != "no" && env != "off" {
			logger.SetDebug(true)
		}
	}

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
	if len(args) > 0 && args[0] == "__complete" {
		return
	}
	for _, a := range args {
		if a == "--json" || a == "-j" {
			return
		}
	}
	cached := updater.CheckCached(Version, config.CacheDir())
	if cached != nil && cached.UpdateAvailable {
		fmt.Fprintf(os.Stderr, "\nA new version of aim is available: %s → %s (run 'brew upgrade aim')\n", Version, cached.LatestVersion)
	}
	updater.MaybeTriggerBackgroundCheck(Version, config.CacheDir())
}

func main() {
	reg := defaultRegistry()

	pm := profile.NewProfileManager(config.BaseDir())

	os.Exit(dispatch(os.Args[1:], reg, pm))
}
