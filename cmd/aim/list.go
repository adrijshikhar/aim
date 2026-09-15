package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/logger"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/tui"
	"github.com/aim-cli/aim/internal/usage"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
)

func newListCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	return &cobra.Command{
		Use:   "list [agent]",
		Short: "List all profiles and status",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			agent := ""
			if len(args) > 0 {
				agent = args[0]
			}
			return runList(reg, pm, agent)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return completeAgents(reg, toComplete), cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}
}

func formatProfileUsageBadge(cache *usage.CacheStore, agent, profileName string) string {
	if cache == nil {
		return ""
	}
	rep, ok := cache.Get(agent, profileName)
	if !ok {
		return ""
	}
	if rep.Error != "" || rep.Status == usage.StatusUnknown {
		errLower := strings.ToLower(rep.Error)
		summaryLower := strings.ToLower(rep.Summary)
		if strings.Contains(errLower, "credential") || strings.Contains(summaryLower, "credential") {
			return "[no credentials]"
		}
		if strings.Contains(errLower, "offline") || strings.Contains(summaryLower, "offline") ||
			strings.Contains(errLower, "connect") || strings.Contains(errLower, "network") ||
			strings.Contains(errLower, "timeout") {
			return "[offline]"
		}
		if rep.Error != "" {
			return fmt.Sprintf("[%s]", strings.ToLower(rep.Error))
		}
		return "[unknown]"
	}

	groups := rep.ModelGroups()
	if len(groups) > 1 {
		var groupParts []string
		for _, g := range groups {
			catShort := strings.ToLower(g.Category)
			if strings.Contains(catShort, "claude") {
				catShort = "claude"
			} else if strings.Contains(catShort, "gemini") {
				catShort = "gemini"
			}
			var gWindows []string
			for i := range g.Windows {
				w := &g.Windows[i]
				if w.IsHourly() {
					if s := usage.FormatWindowSummary(w); s != "" {
						gWindows = append(gWindows, s)
					}
				}
			}
			for i := range g.Windows {
				w := &g.Windows[i]
				if w.IsWeekly() {
					if s := usage.FormatWindowSummary(w); s != "" {
						gWindows = append(gWindows, s)
					}
				}
			}
			if len(gWindows) > 0 {
				groupParts = append(groupParts, fmt.Sprintf("%s: %s", catShort, strings.Join(gWindows, ", ")))
			}
		}
		if len(groupParts) > 0 {
			return strings.Join(groupParts, " | ")
		}
	}

	var parts []string
	pw := rep.PrimaryWindow()
	if pw != nil {
		if s := usage.FormatWindowSummary(pw); s != "" {
			parts = append(parts, s)
		}
	}
	ww := rep.WeeklyWindow()
	if ww != nil && (pw == nil || ww.Name != pw.Name) {
		if s := usage.FormatWindowSummary(ww); s != "" {
			parts = append(parts, s)
		}
	}
	if len(parts) > 0 {
		return strings.Join(parts, ", ")
	}
	if len(rep.Windows) > 0 {
		for i := range rep.Windows {
			if s := usage.FormatWindowSummary(&rep.Windows[i]); s != "" {
				parts = append(parts, s)
			}
		}
		if len(parts) > 0 {
			return strings.Join(parts, ", ")
		}
	}
	return ""
}

func runList(reg *agents.Registry, pm *profile.ProfileManager, agentName string) error {
	if pm == nil {
		return nil
	}
	logger.Debug("[list] Listing profiles (agent=%q)", agentName)
	cfg, _ := config.LoadConfig()
	baseDir := config.BaseDir()
	if pm != nil && pm.BaseDir != "" {
		baseDir = pm.BaseDir
	}
	cache := usage.NewCacheStore(baseDir, usage.DefaultTTL)
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(tui.AccentBlue)

	if agentName != "" {
		canonicalAgent := agentName
		if reg != nil {
			if ad, err := reg.Get(agentName); err == nil {
				canonicalAgent = ad.Name()
			}
		}
		profiles, err := pm.ListProfilesForAgent(canonicalAgent, cfg, reg)
		if err != nil {
			return err
		}
		fmt.Println(headerStyle.Render(fmt.Sprintf("=== Configured Profiles (%s) ===", canonicalAgent)))
		if len(profiles) == 0 {
			emptyStyle := lipgloss.NewStyle().Foreground(tui.TextMuted)
			fmt.Println(emptyStyle.Render(fmt.Sprintf("  No profiles found for agent %q. Run: aim login %s <profile>", canonicalAgent, canonicalAgent)))
			return nil
		}
		hasStale := false
		t := table.New().
			Border(lipgloss.HiddenBorder()).
			Headers("#", "PROFILE", "QUOTA STATUS", "STORAGE PATH")

		for i, p := range profiles {
			if cache.IsStale(canonicalAgent, p, 10*time.Minute) {
				hasStale = true
			}
			badgeStr := ""
			if summary := formatProfileUsageBadge(cache, canonicalAgent, p); summary != "" {
				if strings.HasPrefix(summary, "[") && strings.HasSuffix(summary, "]") {
					badgeStr = summary
				} else {
					badgeStr = fmt.Sprintf("(%s)", summary)
				}
			}
			t.Row(
				fmt.Sprintf("%d.", i+1),
				p,
				badgeStr,
				pm.ProfileDir(p),
			)
		}

		t.StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return lipgloss.NewStyle().Bold(true).Foreground(tui.AccentBlue)
			}
			switch col {
			case 0:
				return lipgloss.NewStyle().Foreground(tui.TextMuted)
			case 1:
				return lipgloss.NewStyle().Bold(true).Foreground(tui.TextBright)
			case 2:
				return lipgloss.NewStyle().Foreground(tui.AccentCyan)
			default:
				return lipgloss.NewStyle().Foreground(tui.TextDim)
			}
		})

		fmt.Println(t.Render())
		if hasStale {
			triggerPrewarmAsync(baseDir, canonicalAgent)
		}
		return nil
	}

	// List all profiles with attached agent tags
	profiles, err := pm.ListProfiles()
	if err != nil {
		return err
	}
	fmt.Println(headerStyle.Render("=== Configured Profiles ==="))
	if len(profiles) == 0 {
		emptyStyle := lipgloss.NewStyle().Foreground(tui.TextMuted)
		fmt.Println(emptyStyle.Render("  No profiles found. Run: aim login agy <profile>"))
		return nil
	}
	hasStale := false
	t := table.New().
		Border(lipgloss.HiddenBorder()).
		Headers("#", "PROFILE", "AGENTS", "QUOTA STATUS", "STORAGE PATH")

	for i, p := range profiles {
		agentsList := cfg.GetProfileAgents(p)
		agentStr := ""
		if len(agentsList) > 0 {
			agentStr = fmt.Sprintf("[%s]", strings.Join(agentsList, ", "))
		}

		var summaries []string
		if len(agentsList) > 0 {
			for _, ag := range agentsList {
				if cache.IsStale(ag, p, 10*time.Minute) {
					hasStale = true
				}
				if s := formatProfileUsageBadge(cache, ag, p); s != "" {
					if len(agentsList) > 1 {
						summaries = append(summaries, fmt.Sprintf("%s: %s", ag, s))
					} else {
						summaries = append(summaries, s)
					}
				}
			}
		} else if reg != nil {
			for _, ad := range reg.All() {
				if cache.IsStale(ad.Name(), p, 10*time.Minute) {
					hasStale = true
				}
				if s := formatProfileUsageBadge(cache, ad.Name(), p); s != "" {
					summaries = append(summaries, s)
				}
			}
		}
		badgeStr := ""
		if len(summaries) == 1 && strings.HasPrefix(summaries[0], "[") && strings.HasSuffix(summaries[0], "]") {
			badgeStr = summaries[0]
		} else if len(summaries) > 0 {
			badgeStr = fmt.Sprintf("(%s)", strings.Join(summaries, ", "))
		}

		t.Row(
			fmt.Sprintf("%d.", i+1),
			p,
			agentStr,
			badgeStr,
			pm.ProfileDir(p),
		)
	}

	t.StyleFunc(func(row, col int) lipgloss.Style {
		if row == table.HeaderRow {
			return lipgloss.NewStyle().Bold(true).Foreground(tui.AccentBlue)
		}
		switch col {
		case 0:
			return lipgloss.NewStyle().Foreground(tui.TextMuted)
		case 1:
			return lipgloss.NewStyle().Bold(true).Foreground(tui.TextBright)
		case 2:
			return lipgloss.NewStyle().Foreground(tui.AccentPurple)
		case 3:
			return lipgloss.NewStyle().Foreground(tui.AccentCyan)
		default:
			return lipgloss.NewStyle().Foreground(tui.TextDim)
		}
	})

	fmt.Println(t.Render())
	if hasStale {
		triggerPrewarmAsync(baseDir, "")
	}
	return nil
}
