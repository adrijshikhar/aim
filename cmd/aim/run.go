package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/logger"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/runner"
	"github.com/aim-cli/aim/internal/session"
	"github.com/aim-cli/aim/internal/tui"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

func newRunCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:                "run <agent> <profile> [flags...] [-- args...]",
		Short:              "Execute agent under isolated profile",
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var autoCreate bool
			var cleanArgs []string
			for _, a := range args {
				if a == "-y" || a == "--yes" || a == "--create" {
					autoCreate = true
				} else {
					cleanArgs = append(cleanArgs, a)
				}
			}
			if len(cleanArgs) > 0 && (cleanArgs[0] == "-h" || cleanArgs[0] == "--help") {
				return cmd.Help()
			}
			if len(cleanArgs) < 2 {
				return fmt.Errorf("run requires <agent> and <profile>")
			}
			agentName := cleanArgs[0]
			profileName := cleanArgs[1]
			var extraArgs []string
			if len(cleanArgs) > 2 {
				extraArgs = cleanArgs[2:]
				if len(extraArgs) > 0 && extraArgs[0] == "--" {
					extraArgs = extraArgs[1:]
				}
			}
			ok, err := confirmProfileExists(cmd, pm, profileName, fmt.Sprintf("start %s", agentName), autoCreate)
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}

			if reg != nil {
				if ad, err := reg.Get(agentName); err == nil {
					agentName = ad.Name()
				}
			}

			mgr := defaultSessionManager()
			resolvedID, err := ensureRunSessionHydrated(cmd, pm, mgr, agentName, profileName, extraArgs)
			if err != nil {
				return err
			}
			if resolvedID != "" {
				sessID := runner.ExtractSessionID(extraArgs)
				extraArgs = runner.ReplaceSessionID(extraArgs, sessID, resolvedID)
			}

			sessID := runner.ExtractSessionID(extraArgs)
			exitCode := executeRunWithSession(reg, pm, agentName, profileName, sessID, extraArgs)
			if exitCode != 0 {
				return &ExitError{Code: exitCode}
			}
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completeAgentAndProfile(reg, pm, args, toComplete)
		},
	}
	return cmd
}

func executeRun(reg *agents.Registry, pm *profile.ProfileManager, agentName, profileName string, extraArgs []string) int {
	return executeRunWithSession(reg, pm, agentName, profileName, "", extraArgs)
}

func executeRunWithSession(reg *agents.Registry, pm *profile.ProfileManager, agentName, profileName, sessionID string, extraArgs []string) int {
	logger.Debug("[run] Executing agent %q with profile %q (sessionID=%s, extraArgs=%v)", agentName, profileName, sessionID, extraArgs)
	adapter, err := reg.Get(agentName)
	if err != nil {
		logger.Debug("[run] Failed to get adapter for agent %q: %v", agentName, err)
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	pDir, err := pm.EnsureProfile(profileName)
	if err != nil {
		logger.Debug("[run] Failed to ensure profile %q: %v", profileName, err)
		fmt.Fprintf(os.Stderr, "Error preparing profile: %v\n", err)
		return 1
	}
	logger.Debug("[run] Profile %q directory: %s", profileName, pDir)

	cfg, _ := config.LoadConfig()
	if cfg != nil {
		cfg.AddProfileAgent(profileName, adapter.Name())
		_ = config.SaveConfig(cfg)
	}

	launchEnv, err := adapter.PrepareEnv(profileName, pDir)
	if err != nil {
		logger.Debug("[run] PrepareEnv failed for %q: %v", profileName, err)
		fmt.Fprintf(os.Stderr, "Error preparing launch environment: %v\n", err)
		return 1
	}
	logger.Debug("[run] LaunchEnv: binary=%s, workingDir=%s, args=%v, envVars=%d", launchEnv.BinaryPath, launchEnv.WorkingDir, launchEnv.Args, len(launchEnv.Env))

	// Ensure AIM_SESSION_ID is set in launchEnv if resuming or running with a known session ID
	if sessionID == "" {
		sessionID = runner.ExtractSessionID(extraArgs)
	}
	if sessionID != "" {
		if launchEnv.Env == nil {
			launchEnv.Env = make(map[string]string)
		}
		if launchEnv.Env["AIM_SESSION_ID"] == "" {
			launchEnv.Env["AIM_SESSION_ID"] = sessionID
		}
	}

	applyProfileOverrides(&launchEnv, cfg, profileName, true)

	r := runner.NewRunner()
	logger.Debug("[run] Invoking runner.Run with extraArgs=%v", extraArgs)
	code, err := r.Run(context.Background(), launchEnv, extraArgs)
	if err != nil {
		logger.Debug("[run] Runner.Run returned error: %v", err)
		fmt.Fprintf(os.Stderr, "Execution error: %v\n", err)
		return 1
	}
	logger.Debug("[run] Process finished with exit code %d", code)

	// Trigger asynchronous cache pre-warming upon session exit so quotas reflect recent usage
	triggerPrewarmAsync(config.BaseDir(), adapter.Name())

	return code
}

func applyProfileOverrides(launch *agents.LaunchEnv, cfg *config.Config, profileName string, includeArgs bool) {
	if launch == nil || cfg == nil {
		return
	}
	if env := cfg.GetProfileEnv(profileName); len(env) > 0 {
		if launch.Env == nil {
			launch.Env = make(map[string]string)
		}
		for key, value := range env {
			launch.Env[key] = value
		}
	}
	if includeArgs {
		launch.Args = append(launch.Args, cfg.GetProfileArgs(profileName)...)
	}
}

func ensureRunSessionHydrated(cmd *cobra.Command, pm *profile.ProfileManager, mgr *session.Manager, agentName, profileName string, extraArgs []string) (string, error) {
	sessionID := runner.ExtractSessionID(extraArgs)
	if sessionID == "" {
		return "", nil
	}

	if mgr == nil {
		mgr = defaultSessionManager()
	}

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	latest, err := findLatestSessionAcrossProfiles(ctx, mgr, agentName, sessionID)
	if err != nil || latest == nil {
		logger.Debug("[run] Session %q not found across profiles during auto-hydrate check: %v", sessionID, err)
		return "", nil
	}

	// If the latest copy is already native to the target profile, no hydration is needed,
	// but return the resolved full ID so caller can expand short prefixes in extraArgs.
	if latest.Profile == profileName && !latest.IsHost {
		return latest.ID, nil
	}

	pDir, err := pm.EnsureProfile(profileName)
	if err != nil {
		return "", fmt.Errorf("failed to ensure profile %q: %w", profileName, err)
	}

	prov := mgr.Provider(agentName)
	if prov == nil {
		logger.Debug("[run] No session provider registered for agent %q", agentName)
		return latest.ID, nil
	}

	destSess, err := prov.GetSession(ctx, sessionID, pDir, false)
	needsHydrate := false
	if err != nil || destSess == nil {
		needsHydrate = true
	} else if latest.LastActiveAt.After(destSess.LastActiveAt) {
		needsHydrate = true
	}

	if needsHydrate {
		arrowStyle := lipgloss.NewStyle().Foreground(tui.AccentBlue).Bold(true)
		cyanStyle := lipgloss.NewStyle().Foreground(tui.AccentCyan)
		boldStyle := lipgloss.NewStyle().Bold(true).Foreground(tui.TextBright)
		sourceDesc := latest.Profile
		if latest.IsHost {
			sourceDesc = "host"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s Syncing latest %s session %s from %s into %q...\n",
			arrowStyle.Render("➜"),
			boldStyle.Render(agentName),
			cyanStyle.Render(latest.ShortID),
			sourceDesc,
			profileName,
		)

		_, err = prov.Hydrate(ctx, latest, pDir, false)
		if err != nil {
			return "", fmt.Errorf("failed to auto-hydrate session %q into profile %q: %w", sessionID, profileName, err)
		}
	}

	return latest.ID, nil
}
