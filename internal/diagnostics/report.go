package diagnostics

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
	"github.com/aim-cli/aim/internal/profile"
)

// GenerateReport creates a sanitized, anonymized Markdown diagnostic report
// suitable for GitHub issue submission and feedback reporting.
func GenerateReport(reg *agents.Registry, pm *profile.ProfileManager, cfg *config.Config, agentName, version, commit string) string {
	var b strings.Builder
	homeDir, _ := os.UserHomeDir()
	realHome := config.RealHomeDir()
	sanitize := func(s string) string {
		return SanitizeText(s, homeDir, realHome)
	}

	if cfg == nil {
		cfg, _ = config.LoadConfig()
	}

	b.WriteString("# AIM Diagnostic Report\n\n")

	b.WriteString("## System & Environment\n")
	b.WriteString(fmt.Sprintf("- AIM Version: %s (commit: %s)\n", version, commit))
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
