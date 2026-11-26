package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/logger"
	"github.com/aim-cli/aim/internal/presenter"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/usage"
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
			return executeUsageTo(cmd.OutOrStdout(), cmd.ErrOrStderr(), reg, pm, args, opts)
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
	return executeUsageTo(os.Stdout, os.Stderr, reg, pm, args, opts)
}

func executeUsageTo(out, errOut io.Writer, reg *agents.Registry, pm *profile.ProfileManager, args []string, opts usageOptions) error {
	targetAgent, targetProfile := "", ""
	if len(args) > 0 {
		targetAgent = args[0]
	}
	if len(args) > 1 {
		targetProfile = args[1]
	}
	logger.Debug("[usage] Querying usage (agent=%q, profile=%q, refresh=%t, json=%t)", targetAgent, targetProfile, opts.refresh, opts.jsonOutput)

	if pm == nil {
		return writeEmptyUsage(out, targetAgent, opts.jsonOutput, true)
	}
	if reg == nil {
		reg = defaultRegistry()
	}

	cfg, _ := config.LoadConfig()
	baseDir := config.BaseDir()
	if pm.BaseDir != "" {
		baseDir = pm.BaseDir
	}
	cache := usage.NewCacheStore(baseDir, usage.DefaultTTL)

	canonicalAgent, adapters, err := resolveUsageAdapters(reg, targetAgent)
	if err != nil {
		return err
	}
	if opts.refresh {
		flushAgent := ""
		if targetAgent != "" {
			flushAgent = canonicalAgent
		}
		cache.Flush(flushAgent)
	}

	targets := collectUsageTargets(pm, cfg, reg, adapters, targetProfile)
	if len(targets) == 0 {
		return writeEmptyUsage(out, targetAgent, opts.jsonOutput, false)
	}

	useColor := isTerminalWriter(out) && !opts.noColor
	reports := collectUsageReports(errOut, targets, cache, useColor, opts.jsonOutput)

	if opts.jsonOutput {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(reports)
	}

	renderUsageReports(out, reports, useColor)
	return nil
}

func writeEmptyUsage(out io.Writer, targetAgent string, jsonOutput, noProfiles bool) error {
	if jsonOutput {
		_, err := fmt.Fprintln(out, "[]")
		return err
	}
	if !noProfiles && targetAgent != "" {
		_, err := fmt.Fprintf(out, "No profiles found for agent %s.\n", targetAgent)
		return err
	}
	_, err := fmt.Fprintln(out, "No profiles configured.")
	return err
}

func resolveUsageAdapters(reg *agents.Registry, targetAgent string) (string, []agents.AgentAdapter, error) {
	if targetAgent == "" {
		return "", reg.All(), nil
	}
	ad, err := reg.Get(targetAgent)
	if err != nil {
		return "", nil, fmt.Errorf("unknown agent %q", targetAgent)
	}
	return ad.Name(), []agents.AgentAdapter{ad}, nil
}

func collectUsageTargets(pm *profile.ProfileManager, cfg *config.Config, reg *agents.Registry, adapters []agents.AgentAdapter, targetProfile string) []usage.TargetProfile {
	var targets []usage.TargetProfile
	for _, ad := range adapters {
		profs, _ := pm.ListProfilesForAgent(ad.Name(), cfg, reg)
		for _, p := range profs {
			if targetProfile != "" && p != targetProfile {
				continue
			}
			var usageFn func(context.Context, string, string) (*usage.Report, error)
			if up, ok := ad.(agents.UsageProvider); ok {
				usageFn = up.GetUsage
			}
			var cachedUsageFn func(context.Context, string, string) (*usage.Report, error)
			if cp, ok := ad.(agents.CachedUsageProvider); ok {
				cachedUsageFn = cp.GetCachedUsage
			}
			targets = append(targets, usage.TargetProfile{
				Agent:            ad.Name(),
				Profile:          p,
				ProfileDir:       pm.ProfileDir(p),
				GetUsageFn:       usageFn,
				GetCachedUsageFn: cachedUsageFn,
			})
		}
	}
	return targets
}

func collectUsageReports(errOut io.Writer, targets []usage.TargetProfile, cache *usage.CacheStore, useColor, jsonOutput bool) []usage.Report {
	reportsChan := usage.RefreshAsync(context.Background(), targets, cache)
	reports := make([]usage.Report, 0, len(targets))

	if useColor && !jsonOutput {
		spin := presenter.NewSpinner(errOut, fmt.Sprintf("Querying usage quotas for %d profile(s)...", len(targets)))
		spin.Start()
		count := 0
		for r := range reportsChan {
			reports = append(reports, r)
			count++
			spin.UpdateText(fmt.Sprintf("Querying usage quotas... [%d/%d] %s:%s", count, len(targets), r.Agent, r.Profile))
		}
		spin.Stop()
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
	return reports
}

func renderUsageReports(out io.Writer, reports []usage.Report, useColor bool) {
	hasCategory, hasAccount, has5h := detectUsageColumns(reports)
	columns := usageTableColumns{hasAccount: hasAccount, hasCategory: hasCategory, has5h: has5h}
	rows := buildUsageRows(reports, columns, useColor)
	presenter.PrintTable(out, columns.headers(), rows, useColor)
}

func detectUsageColumns(reports []usage.Report) (hasCategory, hasAccount, has5h bool) {
	for _, r := range reports {
		if r.AccountEmail != "" {
			hasAccount = true
		}
		for i := range r.Windows {
			if strings.TrimSpace(r.Windows[i].Category) != "" {
				hasCategory = true
			}
			if r.Windows[i].IsHourly() {
				has5h = true
			}
		}
	}
	return hasCategory, hasAccount, has5h
}

func buildUsageRows(reports []usage.Report, columns usageTableColumns, useColor bool) [][]string {
	var rows [][]string
	for _, r := range reports {
		checked := "just now"
		if r.FromCache {
			checked = usage.FormatDuration(time.Since(r.FetchedAt)) + " ago"
		}
		accStr := "—"
		if r.AccountEmail != "" {
			accStr = r.AccountEmail
		}

		if len(r.Windows) == 0 {
			rows = append(rows, columns.row(usageRowValues{agent: r.Agent, profile: r.Profile, account: accStr, category: "—", status: presenter.RenderCLIStatus(r.Status, useColor), primary: "—", primaryReset: "—", weekly: "—", weeklyReset: "—", checked: checked}))
			continue
		}

		rows = append(rows, buildUsageCategoryRows(r, accStr, checked, columns, useColor)...)
	}
	return rows
}

func buildUsageCategoryRows(r usage.Report, accStr, checked string, columns usageTableColumns, useColor bool) [][]string {
	type catGroup struct {
		category string
		windows  []usage.LimitWindow
	}
	var groups []catGroup
	catIdx := make(map[string]int)

	for _, w := range r.Windows {
		cat := strings.TrimSpace(w.Category)
		if idx, exists := catIdx[cat]; !exists {
			catIdx[cat] = len(groups)
			groups = append(groups, catGroup{category: cat, windows: []usage.LimitWindow{w}})
		} else {
			groups[idx].windows = append(groups[idx].windows, w)
		}
	}

	rows := make([][]string, 0, len(groups))
	for _, g := range groups {
		catStatus := usage.CalculateStatus(g.windows)
		catName := usage.CleanModelCategory(g.category)
		primary, weekly := findCategoryWindows(g.windows)

		pStr, pReset := formatUsageWindow(primary, catStatus, useColor)
		wStr, wReset := formatUsageWindow(weekly, catStatus, useColor)
		statusStr := presenter.RenderCLIStatus(catStatus, useColor)
		rows = append(rows, columns.row(usageRowValues{agent: r.Agent, profile: r.Profile, account: accStr, category: catName, status: statusStr, primary: pStr, primaryReset: pReset, weekly: wStr, weeklyReset: wReset, checked: checked}))
	}
	return rows
}

func formatUsageWindow(w *usage.LimitWindow, catStatus usage.Status, useColor bool) (bar, reset string) {
	if w == nil {
		return "—", "—"
	}
	bar = presenter.RenderCLIBar(w.RemainingPct, 10, catStatus, useColor)
	reset = "—"
	if w.RemainingPct < 100 && w.ResetsIn > 0 {
		reset = usage.FormatDuration(w.ResetsIn)
	}
	return bar, reset
}

func isTerminal() bool {
	return isTerminalWriter(os.Stdout)
}

func isTerminalWriter(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return false
	}
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
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
