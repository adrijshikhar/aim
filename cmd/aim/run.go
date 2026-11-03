package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/logger"
	"github.com/aim-cli/aim/internal/presenter"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/runner"
	"github.com/aim-cli/aim/internal/session"
	"github.com/aim-cli/aim/internal/share"
	"github.com/aim-cli/aim/internal/telemetry"
	"github.com/aim-cli/aim/internal/tui"
	"github.com/charmbracelet/lipgloss"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

var (
	leaseMu                sync.RWMutex
	newShareClient         = share.NewClient
	leaseHeartbeatInterval = 60 * time.Second
)

func getShareClient(port int) *share.Client {
	leaseMu.RLock()
	defer leaseMu.RUnlock()
	return newShareClient(port)
}

func getHeartbeatInterval() time.Duration {
	leaseMu.RLock()
	defer leaseMu.RUnlock()
	if env := os.Getenv("AIM_LEASE_HEARTBEAT_INTERVAL"); env != "" {
		if d, err := time.ParseDuration(env); err == nil && d > 0 {
			return d
		}
	}
	if leaseHeartbeatInterval <= 0 {
		return 60 * time.Second
	}
	return leaseHeartbeatInterval
}

func newRunCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:                "run <agent> <profile> [flags...] [-- args...]",
		Short:              "Execute agent under isolated profile",
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			autoCreate, cleanArgs := splitRunFlags(args)
			if len(cleanArgs) > 0 && (cleanArgs[0] == "-h" || cleanArgs[0] == "--help") {
				return cmd.Help()
			}
			if len(cleanArgs) == 0 {
				return fmt.Errorf("run requires <agent> and <profile>")
			}
			agentName := cleanArgs[0]
			var profileName string
			var extraArgs []string

			if len(cleanArgs) >= 2 && !strings.HasPrefix(cleanArgs[1], "-") {
				profileName = cleanArgs[1]
				if len(cleanArgs) > 2 {
					extraArgs = cleanArgs[2:]
				}
			} else {
				// No explicit profile argument provided; attempt repository directory binding resolution
				cwd, _ := os.Getwd()
				res, err := profile.ResolveDirProfile(cwd)
				if err != nil {
					return fmt.Errorf("resolving repository profile: %w", err)
				}
				if res != nil && res.Profile != "" {
					profileName = res.Profile
					fmt.Fprintf(cmd.ErrOrStderr(), "[aim] Using repository-bound profile '%s' from %s\n", profileName, filepath.Base(res.SourceFile))
				} else {
					return fmt.Errorf("run requires <agent> and <profile>")
				}
				if len(cleanArgs) > 1 {
					extraArgs = cleanArgs[1:]
				}
			}

			if len(extraArgs) > 0 && extraArgs[0] == "--" {
				extraArgs = extraArgs[1:]
			}
			if profile.IsLeased(profileName) || (pm != nil && pm.IsLeased(profileName)) {
				sanitized, err := sanitizeLeasedRunFlags(extraArgs)
				if err != nil {
					return err
				}
				extraArgs = sanitized
			}
			if !strings.Contains(profileName, ":") {
				ok, err := confirmProfileExists(cmd, pm, profileName, fmt.Sprintf("start %s", agentName), autoCreate)
				if err != nil {
					return err
				}
				if !ok {
					return nil
				}
			}

			if reg != nil {
				if ad, err := reg.Get(agentName); err == nil {
					agentName = ad.Name()
				}
			}

			mgr := getSessionManager(cmd.Context())
			resolvedID, err := ensureRunSessionHydrated(cmd, pm, mgr, agentName, profileName, extraArgs)
			if err != nil {
				return err
			}
			if resolvedID != "" {
				sessID := runner.ExtractSessionID(extraArgs)
				extraArgs = runner.ReplaceSessionID(extraArgs, sessID, resolvedID)
			}

			sessID := runner.ExtractSessionID(extraArgs)
			exitCode := executeRunWithSession(cmd.Context(), reg, pm, agentName, profileName, sessID, extraArgs)
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

type RunOptions struct {
	AgentName   string
	ProfileName string
	SessionID   string
	WorkingDir  string
	ExtraArgs   []string
}

func executeRun(reg *agents.Registry, pm *profile.ProfileManager, agentName, profileName string, extraArgs []string) int {
	return executeRunWithSession(context.Background(), reg, pm, agentName, profileName, "", extraArgs)
}

func executeRunWithSession(ctx context.Context, reg *agents.Registry, pm *profile.ProfileManager, agentName, profileName, sessionID string, extraArgs []string) int {
	return executeRunWithOptions(ctx, reg, pm, RunOptions{
		AgentName:   agentName,
		ProfileName: profileName,
		SessionID:   sessionID,
		ExtraArgs:   extraArgs,
	})
}

func executeRunWithOptions(ctx context.Context, reg *agents.Registry, pm *profile.ProfileManager, opts RunOptions) int {
	if ctx == nil {
		ctx = context.Background()
	}
	logger.Debug("[run] Executing agent %q with profile %q (sessionID=%s, workingDir=%s, extraArgs=%v)", opts.AgentName, opts.ProfileName, opts.SessionID, opts.WorkingDir, opts.ExtraArgs)
	adapter, err := reg.Get(opts.AgentName)
	if err != nil {
		logger.Debug("[run] Failed to get adapter for agent %q: %v", opts.AgentName, err)
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	profileName := opts.ProfileName
	var pDir string
	if strings.Contains(profileName, ":") {
		parts := strings.SplitN(profileName, ":", 2)
		hostPart := parts[0]
		profPart := parts[1]
		mgr := share.NewLeaseManager(pm.ProfilesRoot())
		candidateDir := filepath.Join(pm.ProfilesRoot(), profileName)
		leaseFile := filepath.Join(candidateDir, ".aim-lease.json")
		if _, err := os.Stat(leaseFile); os.IsNotExist(err) {
			logger.Debug("[run] Leased profile %q not found locally, attempting auto-lease from %s...", profileName, hostPart)
			client := share.NewClient(4319)
			reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			resp, reqErr := client.RequestLease(reqCtx, hostPart, profPart, adapter.Name(), 2)
			cancel()
			if reqErr != nil && !strings.Contains(hostPart, ".") {
				retryCtx, retryCancel := context.WithTimeout(ctx, 5*time.Second)
				resp, reqErr = client.RequestLease(retryCtx, hostPart+".local", profPart, adapter.Name(), 2)
				retryCancel()
			}
			if reqErr != nil || resp == nil || resp.Status != "approved" {
				errReason := "lease request failed"
				if reqErr != nil {
					errReason = reqErr.Error()
				} else if resp != nil {
					errReason = resp.Status
					if resp.Message != "" {
						errReason += ": " + resp.Message
					}
				}
				fmt.Fprintf(os.Stderr, "Error: Could not acquire lease for %q from %s: %s\n", profPart, hostPart, errReason)
				return 1
			}
			installedName, installErr := mgr.InstallLease(hostPart, profPart, resp)
			if installErr != nil {
				fmt.Fprintf(os.Stderr, "Error: Failed to install lease: %v\n", installErr)
				return 1
			}
			profileName = installedName
			opts.ProfileName = profileName
			pDir = filepath.Join(pm.ProfilesRoot(), profileName)
			fmt.Printf("%s Auto-acquired lease for %s from %s (expires %s)\n",
				lipgloss.NewStyle().Foreground(tui.StatusGreen).Render("✔"),
				profPart, hostPart, resp.ExpiresAt.Format("15:04:05"))
		} else {
			pDir = candidateDir
		}
		if mgr.IsLeaseExpired(profileName) {
			_ = mgr.RevokeLease(profileName)
			fmt.Fprintf(os.Stderr, "Error: Leased profile %q has expired or was revoked. Please renew the lease via 'aim share request %s %s' or the web dashboard.\n", profileName, hostPart, profPart)
			return 1
		}
		// Active verification with peer host if connected
		meta, metaErr := mgr.GetLeaseMeta(profileName)
		if metaErr == nil && meta != nil && meta.LeaseID != "" {
			client := share.NewClient(4319)
			checkCtx, checkCancel := context.WithTimeout(ctx, 600*time.Millisecond)
			st, checkErr := client.CheckLeaseStatus(checkCtx, hostPart, meta.LeaseID)
			checkCancel()
			if checkErr == nil && st != nil && (st.Revoked || !st.Active) {
				_ = mgr.RevokeLease(profileName)
				fmt.Fprintf(os.Stderr, "Error: Leased profile %q has been revoked by host %s.\n", profileName, hostPart)
				return 1
			}
		}
	} else {
		var err error
		pDir, err = pm.EnsureProfile(profileName)
		if err != nil {
			logger.Debug("[run] Failed to ensure profile %q: %v", profileName, err)
			fmt.Fprintf(os.Stderr, "Error preparing profile: %v\n", err)
			return 1
		}
		mgr := share.NewLeaseManager(pm.ProfilesRoot())
		if mgr.HasLease(profileName) && mgr.IsLeaseExpired(profileName) {
			_ = mgr.RevokeLease(profileName)
			fmt.Fprintf(os.Stderr, "Error: Leased profile %q has expired or was revoked. Please renew or release the lease.\n", profileName)
			return 1
		}
	}
	logger.Debug("[run] Profile %q directory: %s", profileName, pDir)

	cfg, _ := config.LoadConfig()
	if cfg != nil {
		cfg.AddProfileAgent(opts.ProfileName, adapter.Name())
		_ = config.SaveConfig(cfg)
	}

	launchEnv, err := adapter.PrepareEnv(ctx, opts.ProfileName, pDir)
	if err != nil {
		logger.Debug("[run] PrepareEnv failed for %q: %v", opts.ProfileName, err)
		fmt.Fprintf(os.Stderr, "Error preparing launch environment: %v\n", err)
		return 1
	}

	if profile.IsLeased(opts.ProfileName) {
		if mcpMgr, ok := adapter.(agents.MCPConfigManager); ok {
			_ = mcpMgr.AddProfileMCPServer(ctx, opts.ProfileName, pDir, "aim_share", agents.MCPServerConfig{
				Command: "aim",
				Args:    []string{"mcp", "serve-share", "--profile", opts.ProfileName},
			})
		}
	}

	if opts.WorkingDir != "" {
		if fi, err := os.Stat(opts.WorkingDir); err == nil && fi.IsDir() {
			launchEnv.WorkingDir = opts.WorkingDir
		}
	}

	logger.Debug("[run] LaunchEnv: binary=%s, workingDir=%s, args=%v, envVars=%d", launchEnv.BinaryPath, launchEnv.WorkingDir, launchEnv.Args, len(launchEnv.Env))

	// Ensure AIM_SESSION_ID is set in launchEnv if resuming or running with a known session ID
	sessionID := opts.SessionID
	if sessionID == "" {
		sessionID = runner.ExtractSessionID(opts.ExtraArgs)
	}
	if sessionID != "" {
		if launchEnv.Env == nil {
			launchEnv.Env = make(map[string]string)
		}
		if launchEnv.Env["AIM_SESSION_ID"] == "" {
			launchEnv.Env["AIM_SESSION_ID"] = sessionID
		}
		mgr := getSessionManager(ctx)
		if prov := mgr.Provider(opts.AgentName); prov != nil {
			if sanitizable, ok := prov.(interface {
				SanitizeSession(context.Context, string, string) error
			}); ok {
				_ = sanitizable.SanitizeSession(ctx, sessionID, pDir)
			}
		}
	}

	applyProfileOverrides(&launchEnv, cfg, opts.ProfileName, true)

	if profile.IsLeased(opts.ProfileName) || (pm != nil && pm.IsLeased(opts.ProfileName)) {
		sanitized, sErr := sanitizeLeasedRunFlags(opts.ExtraArgs)
		if sErr != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", sErr)
			return 1
		}
		opts.ExtraArgs = sanitized
	}

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	if profile.IsLeased(opts.ProfileName) || (pm != nil && pm.IsLeased(opts.ProfileName)) {
		startLeaseSupervisor(runCtx, pDir, opts.ProfileName, cancel)
	}

	r := runner.NewRunner(adapter)
	logger.Debug("[run] Invoking runner.Run with extraArgs=%v", opts.ExtraArgs)
	code := withSessionMerge(adapter, pm.MergeStateStore(), opts.ProfileName, pDir, cfg, opts.ExtraArgs, func() int {
		if isMCPListInvocation(opts.ExtraArgs) {
			if listProv, ok := adapter.(agents.MCPListProvider); ok {
				servers, listErr := listProv.ListMCPServers(runCtx, opts.ProfileName, pDir)
				if listErr == nil {
					presenter.RenderMCPListTable(os.Stdout, adapter.Name(), opts.ProfileName, servers)
					return 0
				}
			}
		}
		c, runErr := r.Run(runCtx, launchEnv, opts.ExtraArgs)
		err = runErr
		return c
	})
	if err != nil {
		logger.Debug("[run] Runner.Run returned error: %v", err)
		fmt.Fprintf(os.Stderr, "Execution error: %v\n", err)
		return 1
	}
	logger.Debug("[run] Process finished with exit code %d", code)
	telemetry.CaptureAccountEvent(adapter.Name(), profile.GetProfileAccountInfoForAgent(pDir, adapter.Name()), "agent_run_completed", map[string]any{
		"exit_code":        code,
		"session_attached": sessionID != "",
	})

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
		mgr = getSessionManager(cmd.Context())
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

func isMCPListInvocation(extraArgs []string) bool {
	if len(extraArgs) < 2 {
		return false
	}
	if extraArgs[0] == "mcp" && extraArgs[1] == "list" {
		for _, arg := range extraArgs[2:] {
			if arg == "--json" || arg == "--raw" || arg == "-h" || arg == "--help" {
				return false
			}
		}
		return true
	}
	return false
}

func sanitizeLeasedRunFlags(extraArgs []string) ([]string, error) {
	banned := map[string]bool{
		"--no-hooks":                     true,
		"--hooks=false":                  true,
		"--dangerously-skip-permissions": true,
		"--config-dir":                   true,
		"--profile":                      true,
	}
	for _, arg := range extraArgs {
		if banned[arg] || strings.HasPrefix(arg, "--hooks=") || strings.HasPrefix(arg, "--config-dir=") || strings.HasPrefix(arg, "--profile=") {
			return nil, fmt.Errorf("flag %q is prohibited on leased profiles to maintain security compliance", arg)
		}
	}
	return extraArgs, nil
}

func startLeaseSupervisor(ctx context.Context, profDir, profileName string, cancel context.CancelFunc) {
	go runLeaseSupervisor(ctx, profDir, profileName, cancel)
}

func isHookFileName(name string) bool {
	base := filepath.Base(name)
	return base == "hooks.json" || base == "settings.json"
}

func checkHookTampered(monitoredHooks map[string]bool, event fsnotify.Event) bool {
	if isHookFileName(event.Name) {
		if event.Op&(fsnotify.Remove|fsnotify.Rename) != 0 {
			data, err := os.ReadFile(event.Name)
			if err != nil || len(data) == 0 || !strings.Contains(string(data), "aim hook guard") {
				return true
			}
		}
		if event.Op&(fsnotify.Write|fsnotify.Chmod) != 0 {
			data, err := os.ReadFile(event.Name)
			if err != nil || len(data) == 0 || !strings.Contains(string(data), "aim hook guard") {
				return true
			}
		}
	}

	for h := range monitoredHooks {
		data, err := os.ReadFile(h)
		if err != nil || len(data) == 0 || !strings.Contains(string(data), "aim hook guard") {
			return true
		}
	}

	return false
}

func runLeaseSupervisor(ctx context.Context, profDir, profileName string, cancel context.CancelFunc) {
	var watcher *fsnotify.Watcher
	if w, err := fsnotify.NewWatcher(); err == nil {
		watcher = w
		defer watcher.Close()

		watchDirs := []string{
			filepath.Join(profDir, ".gemini", "config"),
			filepath.Join(profDir, ".codex"),
			filepath.Join(profDir, ".claude"),
			profDir,
		}
		for _, d := range watchDirs {
			if fi, err := os.Stat(d); err == nil && fi.IsDir() {
				_ = watcher.Add(d)
			}
		}
	} else {
		logger.Debug("[supervisor] fsnotify watcher init error: %v", err)
	}

	candidateHooks := []string{
		filepath.Join(profDir, ".gemini", "config", "hooks.json"),
		filepath.Join(profDir, ".codex", "hooks.json"),
		filepath.Join(profDir, ".claude", "settings.json"),
		filepath.Join(profDir, "hooks.json"),
		filepath.Join(profDir, "settings.json"),
	}

	monitoredHooks := make(map[string]bool)
	for _, h := range candidateHooks {
		if data, err := os.ReadFile(h); err == nil && strings.Contains(string(data), "aim hook guard") {
			monitoredHooks[h] = true
		}
	}

	interval := getHeartbeatInterval()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	var eventsChan <-chan fsnotify.Event
	var errorsChan <-chan error
	if watcher != nil {
		eventsChan = watcher.Events
		errorsChan = watcher.Errors
	}

	for {
		select {
		case <-ctx.Done():
			return

		case event, ok := <-eventsChan:
			if !ok {
				eventsChan = nil
				continue
			}
			// Dynamically watch newly created subdirectories
			if event.Op&fsnotify.Create != 0 && watcher != nil {
				if fi, err := os.Stat(event.Name); err == nil && fi.IsDir() {
					_ = watcher.Add(event.Name)
				}
			}

			// If a new hook file was created with aim hook guard, track it
			if isHookFileName(event.Name) && (event.Op&(fsnotify.Create|fsnotify.Write) != 0) {
				if data, err := os.ReadFile(event.Name); err == nil && strings.Contains(string(data), "aim hook guard") {
					monitoredHooks[event.Name] = true
				}
			}

			if event.Op&(fsnotify.Write|fsnotify.Remove|fsnotify.Rename|fsnotify.Chmod) != 0 {
				if checkHookTampered(monitoredHooks, event) {
					fmt.Fprintf(os.Stderr, "[aim] ⚠️ Tamper detected on hook configuration! Session terminated.\n")
					cancel()
					return
				}
			}

		case err, ok := <-errorsChan:
			if !ok {
				errorsChan = nil
				continue
			}
			logger.Debug("[supervisor] fsnotify watcher error: %v", err)

		case <-ticker.C:
			if performLeaseHeartbeat(ctx, profDir, profileName, cancel) {
				return
			}
		}
	}
}

func performLeaseHeartbeat(ctx context.Context, profDir, profileName string, cancel context.CancelFunc) bool {
	leaseFile := filepath.Join(profDir, ".aim-lease.json")
	data, err := os.ReadFile(leaseFile)
	if err != nil {
		return false
	}
	var meta share.LeaseMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return false
	}

	if meta.Revoked {
		mgr := share.NewLeaseManager(filepath.Dir(profDir))
		_ = mgr.RevokeLease(profileName)
		fmt.Fprintf(os.Stderr, "[aim] 🔒 Lease has been revoked by lender. Terminating session.\n")
		cancel()
		return true
	}

	targetPeer := meta.SourcePeer
	if targetPeer == "" {
		targetPeer = meta.LenderNodeID
	}
	if targetPeer == "" || meta.LeaseID == "" {
		return false
	}

	client := getShareClient(4319)
	checkCtx, checkCancel := context.WithTimeout(ctx, 5*time.Second)
	st, checkErr := client.CheckLeaseStatus(checkCtx, targetPeer, meta.LeaseID)
	checkCancel()

	if checkErr == nil && st != nil {
		if st.Revoked || !st.Active {
			mgr := share.NewLeaseManager(filepath.Dir(profDir))
			_ = mgr.RevokeLease(profileName)
			fmt.Fprintf(os.Stderr, "[aim] 🔒 Lease has been revoked by lender. Terminating session.\n")
			cancel()
			return true
		}

		meta.LastVerifiedAt = time.Now()
		if updated, err := json.MarshalIndent(meta, "", "  "); err == nil {
			tmp := fmt.Sprintf("%s.tmp.%d", leaseFile, time.Now().UnixNano())
			if err := os.WriteFile(tmp, updated, 0600); err == nil {
				_ = os.Rename(tmp, leaseFile)
			}
		}
	}
	return false
}
