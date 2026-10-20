package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/daemon"
	"github.com/aim-cli/aim/internal/diagnostics"
	"github.com/aim-cli/aim/internal/logger"
	"github.com/aim-cli/aim/internal/oauth"
	"github.com/aim-cli/aim/internal/presenter"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/share"
	"github.com/aim-cli/aim/internal/telemetry"
	"github.com/aim-cli/aim/internal/tui"
	"github.com/aim-cli/aim/internal/usage"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

func newDoctorCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	var check bool
	var report bool
	cmd := &cobra.Command{
		Use:   "doctor [agent]",
		Short: "Diagnose environment, tokens, and binaries",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			agent := ""
			if len(args) > 0 {
				agent = args[0]
			}
			if report {
				fmt.Fprint(cmd.OutOrStdout(), generateDiagnosticReport(reg, pm, agent))
				telClient := telemetry.NewClient(config.BaseDir(), config.CacheDir(), Version, nil)
				telClient.Track(telemetry.EventDoctorReportGenerated, map[string]any{})
				_ = telClient.Close()
				return nil
			}
			ok := runDoctorTo(cmd.OutOrStdout(), reg, pm, agent)
			if check && !ok {
				return fmt.Errorf("doctor diagnostics reported failures")
			}
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return completeAgents(reg, toComplete), cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}
	cmd.Flags().BoolVar(&check, "check", false, "Exit with non-zero status if any diagnostic check fails")
	cmd.Flags().BoolVar(&report, "report", false, "Generate an anonymized Markdown diagnostic report for issue submission")
	return cmd
}

func runDoctor(reg *agents.Registry, pm *profile.ProfileManager, agentName string) bool {
	return runDoctorTo(os.Stdout, reg, pm, agentName)
}

func runDoctorTo(out io.Writer, reg *agents.Registry, pm *profile.ProfileManager, agentName string) bool {
	logger.Debug("[doctor] Running diagnostics (agent=%q)", agentName)
	fmt.Fprintln(out, lipgloss.NewStyle().Bold(true).Foreground(tui.AccentBlue).Render("=== AIM Doctor Diagnostics ==="))
	if reg == nil {
		return true
	}
	if pm != nil {
		if err := pm.EnsureAllProfilesDotfiles(); err != nil {
			logger.Debug("[doctor] Failed to ensure dotfiles across profiles: %v", err)
		}
	}
	cfg, _ := config.LoadConfig()

	baseDir := config.BaseDir()
	if pm != nil && pm.BaseDir != "" {
		baseDir = pm.BaseDir
	}
	refStore := oauth.NewRefusalStore(baseDir)
	_ = refStore.Prune()

	if agentName != "" {
		adapter, err := reg.Get(agentName)
		if err != nil {
			failBadge := tui.GaugeRedStyle.Width(8).Render("[FAIL]")
			fmt.Fprintf(out, "%s Unknown agent: %s\n", failBadge, agentName)
			return false
		}
		ok := diagnoseAdapterTo(out, adapter, pm, cfg, reg, map[string]bool{})
		diagnosePlatformTo(out, agentName, cfg)
		diagnoseDaemonTo(out)
		diagnoseVPNTo(out)
		diagnoseUnclaimedTokensTo(out, baseDir, agentName)
		diagnoseBindingTo(out)
		return ok
	}

	allOK := true
	bridged := map[string]bool{}
	for _, adapter := range reg.All() {
		if !diagnoseAdapterTo(out, adapter, pm, cfg, reg, bridged) {
			allOK = false
		}
	}
	diagnosePlatformTo(out, "", cfg)
	diagnoseDaemonTo(out)
	diagnoseVPNTo(out)
	diagnoseUnclaimedTokensTo(out, baseDir, "")
	diagnoseBindingTo(out)
	return allOK
}

func diagnoseDaemon() {
	diagnoseDaemonTo(os.Stdout)
}

func diagnoseDaemonTo(out io.Writer) {
	fmt.Fprintln(out)
	fmt.Fprintln(out, lipgloss.NewStyle().Bold(true).Foreground(tui.TextBright).Render("[Background Daemon]"))
	info, err := daemon.Status(config.BaseDir())
	if err != nil || info == nil || !info.Installed {
		infoBadge := lipgloss.NewStyle().Bold(true).Foreground(tui.TextMuted).Width(8).Render("[INFO]")
		cat := lipgloss.NewStyle().Bold(true).Foreground(tui.TextPrimary).Render("Daemon:")
		fmt.Fprintf(out, "  %s %s Not installed (run 'aim daemon install' to enable background tasks)\n", infoBadge, cat)
		return
	}

	if info.Active {
		okBadge := tui.GaugeGreenStyle.Width(8).Render("[OK]")
		cat := lipgloss.NewStyle().Bold(true).Foreground(tui.TextPrimary).Render("Daemon:")
		lastRunStr := "never"
		if !info.LastRun.IsZero() {
			lastRunStr = formatDaemonTimeAgo(info.LastRun)
		}
		serviceManager := "launchd"
		if runtime.GOOS == "linux" {
			serviceManager = "systemd"
		}
		fmt.Fprintf(out, "  %s %s Active (%s, interval: 15m, last run: %s)\n", okBadge, cat, serviceManager, lastRunStr)
	} else {
		warnBadge := tui.GaugeYellowStyle.Width(8).Render("[WARN]")
		cat := lipgloss.NewStyle().Bold(true).Foreground(tui.TextPrimary).Render("Daemon:")
		fmt.Fprintf(out, "  %s %s Installed but inactive (%s)\n", warnBadge, cat, info.ConfigPath)
	}
}

func diagnoseVPNTo(out io.Writer) {
	fmt.Fprintln(out)
	fmt.Fprintln(out, lipgloss.NewStyle().Bold(true).Foreground(tui.TextBright).Render("[P2P Sharing & VPN Tunnel]"))
	vpn, err := share.DetectVPN()
	cat := lipgloss.NewStyle().Bold(true).Foreground(tui.TextPrimary).Render("VPN Tunnel:")
	if err != nil || vpn == nil || !vpn.Connected {
		infoBadge := lipgloss.NewStyle().Bold(true).Foreground(tui.TextMuted).Width(8).Render("[INFO]")
		fmt.Fprintf(out, "  %s %s Inactive (Connect to Pritunl/OpenVPN to enable P2P profile sharing)\n", infoBadge, cat)
		return
	}

	okBadge := tui.GaugeGreenStyle.Width(8).Render("[OK]")
	fmt.Fprintf(out, "  %s %s Active (%s - %s, broadcast: %s)\n",
		okBadge, cat, vpn.InterfaceName, vpn.IP, vpn.BroadcastIP)
}

func diagnosePlatform(agentName string, cfg *config.Config) {
	diagnosePlatformTo(os.Stdout, agentName, cfg)
}

func diagnosePlatformTo(out io.Writer, agentName string, cfg *config.Config) {
	if runtime.GOOS != "darwin" {
		return
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, lipgloss.NewStyle().Bold(true).Foreground(tui.TextBright).Render("[Platform Diagnostics (macOS)]"))
	var customServices []string
	if cfg != nil {
		customServices = cfg.CustomIgnoredKeychains
	}

	lingering := profile.FindIgnoredKeychains(agentName, customServices...)
	if len(lingering) > 0 {
		okBadge := tui.GaugeGreenStyle.Width(8).Render("[OK]")
		keychainCat := lipgloss.NewStyle().Bold(true).Foreground(tui.TextPrimary).Render("Keychain:")
		fmt.Fprintf(out, "  %s %s %d host agent token(s) detected in macOS Keychain (AIM uses file-based profile isolation):\n",
			okBadge,
			keychainCat,
			len(lingering),
		)
		for _, entry := range lingering {
			fmt.Fprintf(out, "         - %s %s\n",
				lipgloss.NewStyle().Foreground(tui.TextPrimary).Render(entry.Description),
				lipgloss.NewStyle().Foreground(tui.TextMuted).Render(fmt.Sprintf("(service: %q)", entry.Service)),
			)
		}
	} else {
		okBadge := tui.GaugeGreenStyle.Width(8).Render("[OK]")
		keychainCat := lipgloss.NewStyle().Bold(true).Foreground(tui.TextPrimary).Render("Keychain:")
		fmt.Fprintf(out, "  %s %s No host agent tokens in macOS Keychain.\n",
			okBadge,
			keychainCat,
		)
	}
}

func diagnoseUnclaimedTokensTo(out io.Writer, baseDir string, agentName string) {
	store := oauth.NewStashStore(baseDir)
	pruned, _ := store.Prune(7 * 24 * time.Hour)

	creds, err := store.List()
	if err != nil {
		fmt.Fprintln(out)
		fmt.Fprintln(out, lipgloss.NewStyle().Bold(true).Foreground(tui.TextBright).Render("[Unclaimed Tokens]"))
		warnBadge := tui.GaugeYellowStyle.Width(8).Render("[WARN]")
		cat := lipgloss.NewStyle().Bold(true).Foreground(tui.TextPrimary).Render("Unclaimed Tokens:")
		fmt.Fprintf(out, "  %s %s Failed to read unclaimed tokens store: %v\n", warnBadge, cat, err)
		return
	}

	var filtered []*oauth.StashedCredential
	for _, c := range creds {
		if agentName == "" || c.Agent == agentName {
			filtered = append(filtered, c)
		}
	}

	if len(filtered) == 0 && pruned == 0 {
		return
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, lipgloss.NewStyle().Bold(true).Foreground(tui.TextBright).Render("[Unclaimed Tokens]"))
	if pruned > 0 {
		infoBadge := lipgloss.NewStyle().Bold(true).Foreground(tui.TextMuted).Width(8).Render("[INFO]")
		cat := lipgloss.NewStyle().Bold(true).Foreground(tui.TextPrimary).Render("Unclaimed Tokens:")
		fmt.Fprintf(out, "  %s %s Pruned %d expired unclaimed token record(s) older than 7 days\n", infoBadge, cat, pruned)
	}
	if len(filtered) > 0 {
		warnBadge := tui.GaugeYellowStyle.Width(8).Render("[WARN]")
		cat := lipgloss.NewStyle().Bold(true).Foreground(tui.TextPrimary).Render("Unclaimed Tokens:")
		fmt.Fprintf(out, "  %s %s Unclaimed OAuth tokens found (%d uncommitted credential(s) in %s)\n",
			warnBadge, cat, len(filtered), store.Dir)
		for _, c := range filtered {
			fmt.Fprintf(out, "         - Agent: %s, Profile: %s (spent token hash: %.12s...)\n",
				lipgloss.NewStyle().Foreground(tui.AccentCyan).Render(c.Agent),
				lipgloss.NewStyle().Foreground(tui.TextPrimary).Render(c.Profile),
				c.SpentTokenHash)
		}
	}
}

func diagnoseAdapter(adapter agents.AgentAdapter, pm *profile.ProfileManager, cfg *config.Config, reg *agents.Registry, bridged map[string]bool) bool {
	return diagnoseAdapterTo(os.Stdout, adapter, pm, cfg, reg, bridged)
}

func diagnoseAdapterTo(out io.Writer, adapter agents.AgentAdapter, pm *profile.ProfileManager, cfg *config.Config, reg *agents.Registry, bridged map[string]bool) bool {
	allOK := true
	fmt.Fprintln(out)
	fmt.Fprintln(out, lipgloss.NewStyle().Bold(true).Foreground(tui.TextBright).Render(fmt.Sprintf("[%s (%s)]", adapter.DisplayName(), adapter.Name())))
	profiles, _ := pm.ListProfilesForAgent(adapter.Name(), cfg, reg)
	if len(profiles) == 0 {
		fmt.Fprintln(out, lipgloss.NewStyle().Foreground(tui.TextMuted).Render(fmt.Sprintf("  No profiles configured for agent %q. Run: aim login %s <profile>", adapter.Name(), adapter.Name())))
		return true
	}
	baseDir := config.BaseDir()
	if pm != nil && pm.BaseDir != "" {
		baseDir = pm.BaseDir
	}
	refStore := oauth.NewRefusalStore(baseDir)
	usageCache := usage.NewCacheStore(baseDir, usage.DefaultTTL)

	for _, p := range profiles {
		fmt.Fprintln(out)
		fmt.Fprintf(out, "Profile: %s\n", lipgloss.NewStyle().Bold(true).Foreground(tui.AccentCyan).Render(p))
		var results []agents.DiagnosticResult
		if diag, ok := adapter.(agents.Diagnostician); ok {
			results = diag.Doctor(context.Background(), p, pm.ProfileDir(p))
		}
		if rec, ok := refStore.IsRefused(adapter.Name(), p); ok && rec != nil {
			filtered := make([]agents.DiagnosticResult, 0, len(results))
			for _, r := range results {
				if (r.Category == "Auth" || r.Category == "Token") && r.Status == "OK" {
					continue
				}
				filtered = append(filtered, r)
			}
			results = append(filtered, agents.DiagnosticResult{
				Category: "Auth",
				Status:   "FAIL",
				Message:  fmt.Sprintf("Re-login Required (%s)\n         Remedy: Run 'aim login %s %s'", rec.Code, adapter.Name(), p),
			})
		} else {
			profDir := ""
			if pm != nil {
				profDir = pm.ProfileDir(p)
			}
			target := usage.TargetProfile{
				Agent:      adapter.Name(),
				Profile:    p,
				ProfileDir: profDir,
			}
			if cachedProv, ok := adapter.(agents.CachedUsageProvider); ok {
				target.GetCachedUsageFn = cachedProv.GetCachedUsage
			}
			rep := usage.InspectProfile(context.Background(), target, usageCache)

			authFound := false
			hasAuthCategory := false
			for i := range results {
				if results[i].Category == "Auth" || results[i].Category == "Token" {
					hasAuthCategory = true
					if results[i].Status == "OK" {
						if rep.Status != usage.StatusOK {
							results[i].Message = "Credentials configured (Idle - will renew on next turn)"
						}
						authFound = true
					}
				}
			}
			if !authFound && !hasAuthCategory && rep.Status == usage.StatusIdle {
				insertIdx := len(results)
				for i, r := range results {
					if r.Category == "Binary" {
						insertIdx = i + 1
						break
					}
				}
				idleResult := agents.DiagnosticResult{
					Category: "Auth",
					Status:   "OK",
					Message:  "Credentials configured (Idle - will renew on next turn)",
				}
				if insertIdx >= len(results) {
					results = append(results, idleResult)
				} else {
					results = append(results, agents.DiagnosticResult{})
					copy(results[insertIdx+1:], results[insertIdx:])
					results[insertIdx] = idleResult
				}
			}
		}
		if !bridged[p] {
			bridged[p] = true
			var extraPaths []string
			if cfg != nil {
				extraPaths = cfg.CustomBridgedPaths
			}
			results = append(results, profile.BridgeDiagnostics(p, config.RealHomeDir(), pm.ProfileDir(p), extraPaths...)...)
		}
		if cfg != nil {
			if env := cfg.GetProfileEnv(p); len(env) > 0 {
				results = append(results, agents.DiagnosticResult{
					Category: "Config",
					Status:   "OK",
					Message:  fmt.Sprintf("%d custom env var(s) configured", len(env)),
				})
			}
			if args := cfg.GetProfileArgs(p); len(args) > 0 {
				results = append(results, agents.DiagnosticResult{
					Category: "Config",
					Status:   "OK",
					Message:  fmt.Sprintf("%d custom launch arg(s) configured", len(args)),
				})
			}
		}
		for _, r := range results {
			if r.Status == "FAIL" {
				allOK = false
			}
			badge, cat, msg := presenter.RenderDiagnosticResult(r)
			fmt.Fprintf(out, "  %s %s %s\n", badge, cat, msg)
		}
	}
	return allOK
}

func generateDiagnosticReport(reg *agents.Registry, pm *profile.ProfileManager, agentName string) string {
	return diagnostics.GenerateReport(reg, pm, nil, agentName, Version, Commit)
}

func diagnoseBindingTo(out io.Writer) {
	cwd, err := os.Getwd()
	if err != nil {
		return
	}
	res, err := profile.ResolveDirProfile(cwd)
	if err != nil {
		fmt.Fprintln(out)
		fmt.Fprintln(out, lipgloss.NewStyle().Bold(true).Foreground(tui.TextBright).Render("[Repository Binding]"))
		errBadge := tui.GaugeRedStyle.Width(8).Render("[FAIL]")
		cat := lipgloss.NewStyle().Bold(true).Foreground(tui.TextPrimary).Render("Binding:")
		fmt.Fprintf(out, "  %s %s %v\n", errBadge, cat, err)
		return
	}
	if res == nil {
		return
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, lipgloss.NewStyle().Bold(true).Foreground(tui.TextBright).Render("[Repository Binding]"))
	okBadge := tui.GaugeGreenStyle.Width(8).Render("[OK]")
	cat := lipgloss.NewStyle().Bold(true).Foreground(tui.TextPrimary).Render("Binding:")
	fmt.Fprintf(out, "  %s %s Directory bound to profile '%s' (via %s)\n", okBadge, cat, res.Profile, filepath.Base(res.SourceFile))
}
