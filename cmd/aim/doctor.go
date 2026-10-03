package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/diagnostics"
	"github.com/aim-cli/aim/internal/logger"
	"github.com/aim-cli/aim/internal/presenter"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/telemetry"
	"github.com/aim-cli/aim/internal/tui"
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
				fmt.Print(generateDiagnosticReport(reg, pm, agent))
				telClient := telemetry.NewClient(config.BaseDir(), config.CacheDir(), Version, nil)
				telClient.Track(telemetry.EventDoctorReportGenerated, map[string]any{})
				_ = telClient.Close()
				return nil
			}
			ok := runDoctor(reg, pm, agent)
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
	logger.Debug("[doctor] Running diagnostics (agent=%q)", agentName)
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(tui.AccentBlue).Render("=== AIM Doctor Diagnostics ==="))
	if reg == nil {
		return true
	}
	if pm != nil {
		if err := pm.EnsureAllProfilesDotfiles(); err != nil {
			logger.Debug("[doctor] Failed to ensure dotfiles across profiles: %v", err)
		}
	}
	cfg, _ := config.LoadConfig()

	if agentName != "" {
		adapter, err := reg.Get(agentName)
		if err != nil {
			failBadge := tui.GaugeRedStyle.Width(8).Render("[FAIL]")
			fmt.Printf("%s Unknown agent: %s\n", failBadge, agentName)
			return false
		}
		ok := diagnoseAdapter(adapter, pm, cfg, reg, map[string]bool{})
		diagnosePlatform(agentName, cfg)
		return ok
	}

	allOK := true
	bridged := map[string]bool{}
	for _, adapter := range reg.All() {
		if !diagnoseAdapter(adapter, pm, cfg, reg, bridged) {
			allOK = false
		}
	}
	diagnosePlatform("", cfg)
	return allOK
}

func diagnosePlatform(agentName string, cfg *config.Config) {
	if runtime.GOOS != "darwin" {
		return
	}

	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(tui.TextBright).Render("[Platform Diagnostics (macOS)]"))
	var customServices []string
	if cfg != nil {
		customServices = cfg.CustomIgnoredKeychains
	}

	lingering := profile.FindIgnoredKeychains(agentName, customServices...)
	if len(lingering) > 0 {
		okBadge := tui.GaugeGreenStyle.Width(8).Render("[OK]")
		keychainCat := lipgloss.NewStyle().Bold(true).Foreground(tui.TextPrimary).Render("Keychain:")
		fmt.Printf("  %s %s %d host agent token(s) detected in macOS Keychain (AIM uses file-based profile isolation):\n",
			okBadge,
			keychainCat,
			len(lingering),
		)
		for _, entry := range lingering {
			fmt.Printf("         - %s %s\n",
				lipgloss.NewStyle().Foreground(tui.TextPrimary).Render(entry.Description),
				lipgloss.NewStyle().Foreground(tui.TextMuted).Render(fmt.Sprintf("(service: %q)", entry.Service)),
			)
		}
	} else {
		okBadge := tui.GaugeGreenStyle.Width(8).Render("[OK]")
		keychainCat := lipgloss.NewStyle().Bold(true).Foreground(tui.TextPrimary).Render("Keychain:")
		fmt.Printf("  %s %s No host agent tokens in macOS Keychain.\n",
			okBadge,
			keychainCat,
		)
	}
}

// diagnoseAdapter prints one agent's block. The Bridge results describe the
// profile rather than the agent, so they are shown only under the first agent
// that lists the profile; bridged records the profiles already shown.
func diagnoseAdapter(adapter agents.AgentAdapter, pm *profile.ProfileManager, cfg *config.Config, reg *agents.Registry, bridged map[string]bool) bool {
	allOK := true
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(tui.TextBright).Render(fmt.Sprintf("[%s (%s)]", adapter.DisplayName(), adapter.Name())))
	profiles, _ := pm.ListProfilesForAgent(adapter.Name(), cfg, reg)
	if len(profiles) == 0 {
		fmt.Println(lipgloss.NewStyle().Foreground(tui.TextMuted).Render(fmt.Sprintf("  No profiles configured for agent %q. Run: aim login %s <profile>", adapter.Name(), adapter.Name())))
		return true
	}
	for _, p := range profiles {
		fmt.Println()
		fmt.Printf("Profile: %s\n", lipgloss.NewStyle().Bold(true).Foreground(tui.AccentCyan).Render(p))
		var results []agents.DiagnosticResult
		if diag, ok := adapter.(agents.Diagnostician); ok {
			results = diag.Doctor(context.Background(), p, pm.ProfileDir(p))
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
			fmt.Printf("  %s %s %s\n", badge, cat, msg)
		}
	}
	return allOK
}

func generateDiagnosticReport(reg *agents.Registry, pm *profile.ProfileManager, agentName string) string {
	var b strings.Builder
	homeDir, _ := os.UserHomeDir()
	realHome := config.RealHomeDir()
	sanitize := func(s string) string {
		return diagnostics.SanitizeText(s, homeDir, realHome)
	}

	cfg, _ := config.LoadConfig()

	b.WriteString("# AIM Diagnostic Report\n\n")

	b.WriteString("## System & Environment\n")
	b.WriteString(fmt.Sprintf("- AIM Version: %s (commit: %s)\n", Version, Commit))
	b.WriteString(fmt.Sprintf("- Environment: %s/%s (%s)\n", runtime.GOOS, runtime.GOARCH, runtime.Version()))
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "unknown"
	}
	b.WriteString(fmt.Sprintf("- Shell: %s\n", sanitize(shell)))
	if pm != nil {
		b.WriteString(fmt.Sprintf("- AIM Base Directory: %s\n", sanitize(pm.BaseDir)))
	}
	b.WriteString("\n")

	b.WriteString("## Agents & Tooling\n")
	b.WriteString("| Agent | Installed | Binary Path | Configured Profiles |\n")
	b.WriteString("| --- | --- | --- | --- |\n")

	var adapters []agents.AgentAdapter
	if reg != nil {
		if agentName != "" {
			if a, err := reg.Get(agentName); err == nil {
				adapters = append(adapters, a)
			}
		} else {
			adapters = reg.All()
		}
	}

	for _, adapter := range adapters {
		binPath, err := exec.LookPath(adapter.BinaryName())
		installed := "Yes"
		if err != nil {
			installed = "No"
			binPath = "not found"
		} else {
			binPath = sanitize(binPath)
		}
		profCount := 0
		if pm != nil {
			profs, _ := pm.ListProfilesForAgent(adapter.Name(), cfg, reg)
			profCount = len(profs)
		}
		b.WriteString(fmt.Sprintf("| %s | %s | %s | %d |\n", adapter.Name(), installed, binPath, profCount))
	}
	b.WriteString("\n")

	totalProfiles := 0
	if pm != nil {
		allProfs, _ := pm.ListProfiles()
		totalProfiles = len(allProfs)
	}
	b.WriteString("## Profiles Configuration\n")
	b.WriteString(fmt.Sprintf("- Total Configured Profiles: %d\n", totalProfiles))
	if cfg != nil {
		b.WriteString(fmt.Sprintf("- Custom Bridged Paths: %d\n", len(cfg.CustomBridgedPaths)))
		b.WriteString(fmt.Sprintf("- Custom Ignored Keychains: %d\n", len(cfg.CustomIgnoredKeychains)))
	}
	b.WriteString("\n")

	b.WriteString("## Diagnostics Checks\n")
	b.WriteString("| Agent | Category | Status | Message |\n")
	b.WriteString("| --- | --- | --- | --- |\n")

	bridged := map[string]bool{}
	for _, adapter := range adapters {
		if pm == nil {
			continue
		}
		profiles, _ := pm.ListProfilesForAgent(adapter.Name(), cfg, reg)
		for idx, p := range profiles {
			var results []agents.DiagnosticResult
			if diag, ok := adapter.(agents.Diagnostician); ok {
				results = diag.Doctor(context.Background(), p, pm.ProfileDir(p))
			}
			if !bridged[p] {
				bridged[p] = true
				var extraPaths []string
				if cfg != nil {
					extraPaths = cfg.CustomBridgedPaths
				}
				results = append(results, profile.BridgeDiagnostics(p, config.RealHomeDir(), pm.ProfileDir(p), extraPaths...)...)
			}
			profRegex := regexp.MustCompile(`\b` + regexp.QuoteMeta(p) + `\b`)
			for _, r := range results {
				sanitizedMsg := sanitize(r.Message)
				sanitizedMsg = profRegex.ReplaceAllString(sanitizedMsg, fmt.Sprintf("[profile-%d]", idx+1))
				b.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", adapter.Name(), r.Category, r.Status, sanitizedMsg))
			}
		}
	}

	return b.String()
}

