package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/logger"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/tui"
	"github.com/aim-cli/aim/internal/usage"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

type usageOptions struct {
	refresh    bool
	jsonOutput bool
	noColor    bool
}

func newUsageCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	var opts usageOptions
	cmd := &cobra.Command{
		Use:   "usage [agent] [profile]",
		Short: "Display remaining quota and usage limits",
		Args:  cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeUsage(reg, pm, args, opts)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completeAgentAndProfile(reg, pm, args, toComplete)
		},
	}

	cmd.Flags().BoolVarP(&opts.refresh, "refresh", "r", false, "Force live refresh from providers")
	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "Output usage reports as JSON")
	cmd.Flags().BoolVar(&opts.noColor, "no-color", false, "Disable colored output")

	return cmd
}

func runUsage(reg *agents.Registry, pm *profile.ProfileManager, args []string) error {
	cmd := newUsageCmd(reg, pm)
	cmd.SetArgs(args)
	return cmd.Execute()
}

func executeUsage(reg *agents.Registry, pm *profile.ProfileManager, args []string, opts usageOptions) error {
	targetAgent := ""
	targetProfile := ""
	if len(args) > 0 {
		targetAgent = args[0]
	}
	if len(args) > 1 {
		targetProfile = args[1]
	}
	logger.Debug("[usage] Querying usage (agent=%q, profile=%q, refresh=%t, json=%t)", targetAgent, targetProfile, opts.refresh, opts.jsonOutput)

	if pm == nil {
		if opts.jsonOutput {
			fmt.Println("[]")
			return nil
		}
		fmt.Println("No profiles configured.")
		return nil
	}
	if reg == nil {
		reg = agents.DefaultRegistry()
	}

	cfg, _ := config.LoadConfig()
	baseDir := config.BaseDir()
	if pm != nil && pm.BaseDir != "" {
		baseDir = pm.BaseDir
	}
	cache := usage.NewCacheStore(baseDir, usage.DefaultTTL)

	canonicalAgent := targetAgent
	var adapters []agents.AgentAdapter
	if targetAgent != "" {
		ad, err := reg.Get(targetAgent)
		if err != nil {
			return fmt.Errorf("unknown agent %q", targetAgent)
		}
		canonicalAgent = ad.Name()
		adapters = []agents.AgentAdapter{ad}
	} else {
		adapters = reg.All()
	}

	if opts.refresh {
		flushAgent := ""
		if targetAgent != "" {
			flushAgent = canonicalAgent
		}
		cache.Flush(flushAgent)
	}

	var targets []usage.TargetProfile
	for _, ad := range adapters {
		profs, _ := pm.ListProfilesForAgent(ad.Name(), cfg, reg)
		for _, p := range profs {
			if targetProfile != "" && p != targetProfile {
				continue
			}
			pDir := pm.ProfileDir(p)
			adapter := ad
			targets = append(targets, usage.TargetProfile{
				Agent:      adapter.Name(),
				Profile:    p,
				ProfileDir: pDir,
				GetUsageFn: adapter.GetUsage,
			})
		}
	}

	if len(targets) == 0 {
		if opts.jsonOutput {
			fmt.Println("[]")
			return nil
		}
		if targetAgent != "" {
			fmt.Printf("No profiles found for agent %s.\n", targetAgent)
		} else {
			fmt.Println("No profiles configured.")
		}
		return nil
	}

	useColor := isTerminal()
	if opts.noColor {
		useColor = false
	}

	reportsChan := usage.RefreshAsync(context.Background(), targets, cache)
	var reports []usage.Report

	// Show an interactive live spinner on stderr when querying live in an interactive terminal
	if useColor && !opts.jsonOutput {
		doneSpinner := make(chan struct{})
		spinnerFrames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		var statusText string
		var mu sync.Mutex

		statusText = fmt.Sprintf("Querying usage quotas for %d profile(s)...", len(targets))

		go func() {
			frameIdx := 0
			ticker := time.NewTicker(80 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-doneSpinner:
					fmt.Fprintf(os.Stderr, "\r\033[K")
					return
				case <-ticker.C:
					mu.Lock()
					txt := statusText
					mu.Unlock()
					spinner := lipgloss.NewStyle().Foreground(tui.AccentPurple).Render(spinnerFrames[frameIdx%len(spinnerFrames)])
					fmt.Fprintf(os.Stderr, "\r\033[K%s %s", spinner, txt)
					frameIdx++
				}
			}
		}()

		count := 0
		for r := range reportsChan {
			reports = append(reports, r)
			count++
			mu.Lock()
			statusText = fmt.Sprintf("Querying usage quotas... [%d/%d] %s:%s", count, len(targets), r.Agent, r.Profile)
			mu.Unlock()
		}
		close(doneSpinner)
	} else {
		for r := range reportsChan {
			reports = append(reports, r)
		}
	}

	sort.Slice(reports, func(i, j int) bool {
		if reports[i].Agent != reports[j].Agent {
			return reports[i].Agent < reports[j].Agent
		}
		return reports[i].Profile < reports[j].Profile
	})

	if opts.jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(reports)
	}

	hasCategory := false
	for _, r := range reports {
		for _, w := range r.Windows {
			if strings.TrimSpace(w.Category) != "" {
				hasCategory = true
				break
			}
		}
		if hasCategory {
			break
		}
	}

	hasAccount := false
	for _, r := range reports {
		if r.AccountEmail != "" {
			hasAccount = true
			break
		}
	}

	has5h := false
	for _, r := range reports {
		for i := range r.Windows {
			if r.Windows[i].IsHourly() {
				has5h = true
				break
			}
		}
		if has5h {
			break
		}
	}

	columns := usageTableColumns{hasAccount: hasAccount, hasCategory: hasCategory, has5h: has5h}
	headers := columns.headers()

	var rows [][]string
	for _, r := range reports {
		checked := "just now"
		if r.FromCache {
			age := time.Since(r.FetchedAt)
			checked = usage.FormatDuration(age) + " ago"
		}

		accStr := "—"
		if r.AccountEmail != "" {
			accStr = r.AccountEmail
		}

		if len(r.Windows) == 0 {
			rows = append(rows, columns.row(usageRowValues{agent: r.Agent, profile: r.Profile, account: accStr, category: "—", status: renderCLIStatus(r.Status, useColor), primary: "—", primaryReset: "—", weekly: "—", weeklyReset: "—", checked: checked}))
			continue
		}

		type catGroup struct {
			category string
			windows  []usage.LimitWindow
		}
		var groups []catGroup
		catIdx := make(map[string]int)

		for _, w := range r.Windows {
			cat := strings.TrimSpace(w.Category)
			idx, exists := catIdx[cat]
			if !exists {
				catIdx[cat] = len(groups)
				groups = append(groups, catGroup{category: cat, windows: []usage.LimitWindow{w}})
			} else {
				groups[idx].windows = append(groups[idx].windows, w)
			}
		}

		for _, g := range groups {
			catStatus := usage.CalculateStatus(g.windows)
			catName := usage.CleanModelCategory(g.category)
			primary, weekly := findCategoryWindows(g.windows)

			pStr := "—"
			pReset := "—"
			if primary != nil {
				pStr = renderCLIBar(primary.RemainingPct, 10, catStatus, useColor)
				if primary.RemainingPct < 100 && primary.ResetsIn > 0 {
					pReset = usage.FormatDuration(primary.ResetsIn)
				}
			}

			wStr := "—"
			wReset := "—"
			if weekly != nil {
				wStr = renderCLIBar(weekly.RemainingPct, 10, catStatus, useColor)
				if weekly.RemainingPct < 100 && weekly.ResetsIn > 0 {
					wReset = usage.FormatDuration(weekly.ResetsIn)
				}
			}

			statusStr := renderCLIStatus(catStatus, useColor)
			rows = append(rows, columns.row(usageRowValues{agent: r.Agent, profile: r.Profile, account: accStr, category: catName, status: statusStr, primary: pStr, primaryReset: pReset, weekly: wStr, weeklyReset: wReset, checked: checked}))
		}
	}

	printTable(headers, rows, useColor)
	return nil
}

func isTerminal() bool {
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return false
	}
	return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}

type usageTableColumns struct{ hasAccount, hasCategory, has5h bool }

type usageRowValues struct {
	agent, profile, account, category, status, primary, primaryReset, weekly, weeklyReset, checked string
}

func (c usageTableColumns) headers() []string {
	headers := []string{"AGENT", "PROFILE"}
	if c.hasAccount {
		headers = append(headers, "ACCOUNT")
	}
	if c.hasCategory {
		headers = append(headers, "MODEL")
	}
	headers = append(headers, "STATUS")
	if c.has5h {
		headers = append(headers, "5H LIMIT", "5H RESET")
	}
	return append(headers, "WEEKLY LIMIT", "WEEKLY RESET", "CHECKED")
}

func (c usageTableColumns) row(v usageRowValues) []string {
	row := []string{v.agent, v.profile}
	if c.hasAccount {
		row = append(row, v.account)
	}
	if c.hasCategory {
		row = append(row, v.category)
	}
	row = append(row, v.status)
	if c.has5h {
		row = append(row, v.primary, v.primaryReset)
	}
	return append(row, v.weekly, v.weeklyReset, v.checked)
}

func findCategoryWindows(windows []usage.LimitWindow) (*usage.LimitWindow, *usage.LimitWindow) {
	var primary *usage.LimitWindow
	var weekly *usage.LimitWindow

	for i := range windows {
		if primary == nil && windows[i].IsHourly() {
			primary = &windows[i]
		}
		if weekly == nil && windows[i].IsWeekly() {
			weekly = &windows[i]
		}
	}
	if primary == nil && len(windows) > 0 {
		if weekly == nil || windows[0].Name != weekly.Name {
			primary = &windows[0]
		}
	}
	return primary, weekly
}

func renderCLIBar(pct int, width int, st usage.Status, useColor bool) string {
	if !useColor {
		return fmt.Sprintf("%s %d%%", usage.RenderBar(pct, width), pct)
	}

	if width <= 0 {
		return fmt.Sprintf("%d%%", pct)
	}

	clamped := pct
	if clamped < 0 {
		clamped = 0
	} else if clamped > 100 {
		clamped = 100
	}

	filled := (clamped * width) / 100
	empty := width - filled

	fillColor := tui.StatusGreen
	if clamped <= 15 {
		fillColor = tui.StatusRed
	} else if clamped <= 50 {
		fillColor = tui.StatusYellow
	}

	fillStyle := lipgloss.NewStyle().Foreground(fillColor)
	emptyStyle := lipgloss.NewStyle().Foreground(tui.TextMuted)
	bracketStyle := lipgloss.NewStyle().Foreground(tui.TextDim)

	bar := bracketStyle.Render("[") +
		fillStyle.Render(strings.Repeat("█", filled)) +
		emptyStyle.Render(strings.Repeat("░", empty)) +
		bracketStyle.Render("]")

	return fmt.Sprintf("%s %s", bar, fillStyle.Render(fmt.Sprintf("%d%%", pct)))
}

func renderCLIStatus(st usage.Status, useColor bool) string {
	str := string(st)
	if !useColor {
		return str
	}
	return tui.GaugeStyleForStatus(st).Render(str)
}

func printTable(headers []string, rows [][]string, useColor bool) {
	t := table.New().
		Border(lipgloss.HiddenBorder()).
		Headers(headers...).
		Rows(rows...)

	if useColor {
		t.StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return lipgloss.NewStyle().Bold(true).Foreground(tui.AccentCyan)
			}
			return lipgloss.NewStyle().Foreground(tui.TextPrimary)
		})
	}
	fmt.Println(t.Render())
}
