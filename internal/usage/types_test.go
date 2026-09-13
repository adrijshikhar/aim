package usage_test

import (
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/usage"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d        time.Duration
		expected string
	}{
		{45 * time.Minute, "45m"},
		{2*time.Hour + 15*time.Minute, "2h 15m"},
		{5*24*time.Hour + 14*time.Hour, "5d 14h"},
		{-5 * time.Minute, "0m"},
	}

	for _, tc := range tests {
		got := usage.FormatDuration(tc.d)
		if got != tc.expected {
			t.Errorf("FormatDuration(%v) = %q, want %q", tc.d, got, tc.expected)
		}
	}
}

func TestRenderBar(t *testing.T) {
	bar100 := usage.RenderBar(100, 10)
	if bar100 != "[██████████]" {
		t.Errorf("RenderBar(100, 10) = %q, want [██████████]", bar100)
	}

	bar80 := usage.RenderBar(80, 10)
	if bar80 != "[████████░░]" {
		t.Errorf("RenderBar(80, 10) = %q, want [████████░░]", bar80)
	}

	bar0 := usage.RenderBar(0, 10)
	if bar0 != "[░░░░░░░░░░]" {
		t.Errorf("RenderBar(0, 10) = %q, want [░░░░░░░░░░]", bar0)
	}
}

func TestReportWindowsAndStatus(t *testing.T) {
	rep := usage.Report{
		Agent:   "agy",
		Profile: "work",
		Windows: []usage.LimitWindow{
			{
				Category:     "Gemini Models",
				Name:         "Weekly Limit Remaining",
				RemainingPct: 80,
				ResetsAt:     time.Now().Add(5 * 24 * time.Hour),
				ResetsIn:     5 * 24 * time.Hour,
			},
			{
				Category:     "Gemini Models",
				Name:         "Five Hour Limit Remaining",
				RemainingPct: 40,
				ResetsAt:     time.Now().Add(45 * time.Minute),
				ResetsIn:     45 * time.Minute,
			},
		},
	}

	primary := rep.PrimaryWindow()
	if primary == nil || primary.Name != "Five Hour Limit Remaining" {
		t.Fatalf("expected 5-hour primary window, got %+v", primary)
	}

	weekly := rep.WeeklyWindow()
	if weekly == nil || weekly.Name != "Weekly Limit Remaining" {
		t.Fatalf("expected weekly window, got %+v", weekly)
	}

	status := usage.CalculateStatus(rep.Windows)
	if status != usage.StatusWarning {
		t.Errorf("expected StatusWarning for 40%% remaining, got %s", status)
	}

	// Test 100% full window suppression
	fullWin := &usage.LimitWindow{
		Name:         "Five Hour Limit Remaining",
		RemainingPct: 100,
		ResetsIn:     10 * time.Minute,
	}
	formatted := usage.FormatWindowSummary(fullWin)
	if formatted != "5h: 100%" {
		t.Errorf("expected '5h: 100%%', got %q", formatted)
	}

	subWin := &usage.LimitWindow{
		Name:         "Five Hour Limit Remaining",
		RemainingPct: 82,
		ResetsIn:     2*time.Hour + 15*time.Minute,
	}
	formattedSub := usage.FormatWindowSummary(subWin)
	if formattedSub != "5h: 82% [2h 15m]" {
		t.Errorf("expected '5h: 82%% [2h 15m]', got %q", formattedSub)
	}
}

func TestEdgeCases(t *testing.T) {
	// FormatDuration exact hour / exact day
	if got := usage.FormatDuration(2 * time.Hour); got != "2h" {
		t.Errorf("FormatDuration(2h) = %q, want '2h'", got)
	}
	if got := usage.FormatDuration(3 * 24 * time.Hour); got != "3d" {
		t.Errorf("FormatDuration(3d) = %q, want '3d'", got)
	}

	// RenderBar edge cases
	if got := usage.RenderBar(50, 0); got != "" {
		t.Errorf("RenderBar(50, 0) = %q, want empty string", got)
	}
	if got := usage.RenderBar(-10, 5); got != "[░░░░░]" {
		t.Errorf("RenderBar(-10, 5) = %q, want [░░░░░]", got)
	}
	if got := usage.RenderBar(150, 5); got != "[█████]" {
		t.Errorf("RenderBar(150, 5) = %q, want [█████]", got)
	}

	// CalculateStatus edge cases
	if got := usage.CalculateStatus(nil); got != usage.StatusUnknown {
		t.Errorf("CalculateStatus(nil) = %s, want %s", got, usage.StatusUnknown)
	}
	if got := usage.CalculateStatus([]usage.LimitWindow{{RemainingPct: 0}}); got != usage.StatusExhausted {
		t.Errorf("CalculateStatus(0%%) = %s, want %s", got, usage.StatusExhausted)
	}
	if got := usage.CalculateStatus([]usage.LimitWindow{{RemainingPct: 10}}); got != usage.StatusCritical {
		t.Errorf("CalculateStatus(10%%) = %s, want %s", got, usage.StatusCritical)
	}
	if got := usage.CalculateStatus([]usage.LimitWindow{{RemainingPct: 50}}); got != usage.StatusWarning {
		t.Errorf("CalculateStatus(50%%) = %s, want %s", got, usage.StatusWarning)
	}
	if got := usage.CalculateStatus([]usage.LimitWindow{{RemainingPct: 75}}); got != usage.StatusOK {
		t.Errorf("CalculateStatus(75%%) = %s, want %s", got, usage.StatusOK)
	}

	// Report helper edge cases
	emptyRep := usage.Report{}
	if emptyRep.PrimaryWindow() != nil {
		t.Errorf("emptyRep.PrimaryWindow() expected nil, got %+v", emptyRep.PrimaryWindow())
	}
	if emptyRep.WeeklyWindow() != nil {
		t.Errorf("emptyRep.WeeklyWindow() expected nil, got %+v", emptyRep.WeeklyWindow())
	}

	fallbackRep := usage.Report{
		Windows: []usage.LimitWindow{
			{Name: "General Quota", RemainingPct: 60},
		},
	}
	if fallbackRep.PrimaryWindow() == nil || fallbackRep.PrimaryWindow().Name != "General Quota" {
		t.Errorf("fallbackRep.PrimaryWindow() expected 'General Quota', got %+v", fallbackRep.PrimaryWindow())
	}
	if fallbackRep.WeeklyWindow() != nil {
		t.Errorf("fallbackRep.WeeklyWindow() expected nil, got %+v", fallbackRep.WeeklyWindow())
	}

	// FormatWindowSummary edge cases
	if got := usage.FormatWindowSummary(nil); got != "" {
		t.Errorf("FormatWindowSummary(nil) = %q, want empty string", got)
	}
	genWin := &usage.LimitWindow{Name: "Generic", RemainingPct: 65}
	if got := usage.FormatWindowSummary(genWin); got != "limit: 65%" {
		t.Errorf("FormatWindowSummary(generic) = %q, want 'limit: 65%%'", got)
	}
	wkWin := &usage.LimitWindow{Name: "7d Limit", RemainingPct: 40, ResetsIn: 24 * time.Hour}
	if got := usage.FormatWindowSummary(wkWin); got != "wk: 40% [1d]" {
		t.Errorf("FormatWindowSummary(7d) = %q, want 'wk: 40%% [1d]'", got)
	}
	// Nil receiver tests
	var nilRep *usage.Report
	if nilRep.PrimaryWindow() != nil {
		t.Errorf("nilRep.PrimaryWindow() expected nil, got %+v", nilRep.PrimaryWindow())
	}
	if nilRep.WeeklyWindow() != nil {
		t.Errorf("nilRep.WeeklyWindow() expected nil, got %+v", nilRep.WeeklyWindow())
	}

	// Clamped percentage tests in FormatWindowSummary
	overWin := &usage.LimitWindow{Name: "Limit", RemainingPct: 150}
	if got := usage.FormatWindowSummary(overWin); got != "limit: 100%" {
		t.Errorf("FormatWindowSummary(150%%) = %q, want 'limit: 100%%'", got)
	}
	underWin := &usage.LimitWindow{Name: "Limit", RemainingPct: -10}
	if got := usage.FormatWindowSummary(underWin); got != "limit: 0%" {
		t.Errorf("FormatWindowSummary(-10%%) = %q, want 'limit: 0%%'", got)
	}
}

func TestCleanModelCategory(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"Gemini Models", "Gemini"},
		{"gemini", "Gemini"},
		{"Claude and GPT models", "Claude & GPT"},
		{"claude", "Claude & GPT"},
		{"gpt", "Claude & GPT"},
		{"Custom", "Custom"},
		{"", "—"},
		{"   ", "—"},
	}
	for _, tc := range cases {
		got := usage.CleanModelCategory(tc.input)
		if got != tc.want {
			t.Errorf("CleanModelCategory(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestModelGroups(t *testing.T) {
	rep := usage.Report{
		Windows: []usage.LimitWindow{
			{Category: "gemini", Name: "5h", RemainingPct: 50},
			{Category: "gemini", Name: "weekly", RemainingPct: 80},
			{Category: "claude", Name: "5h", RemainingPct: 100},
			{Category: "claude", Name: "weekly", RemainingPct: 90},
		},
	}
	groups := rep.ModelGroups()
	if len(groups) != 2 {
		t.Fatalf("expected 2 model groups, got %d", len(groups))
	}
	if groups[0].Category != "Gemini" || len(groups[0].Windows) != 2 {
		t.Errorf("unexpected group 0: %+v", groups[0])
	}
	if groups[1].Category != "Claude & GPT" || len(groups[1].Windows) != 2 {
		t.Errorf("unexpected group 1: %+v", groups[1])
	}

	// Empty / no categories
	emptyRep := usage.Report{
		Windows: []usage.LimitWindow{
			{Name: "general", RemainingPct: 50},
		},
	}
	if len(emptyRep.ModelGroups()) != 0 {
		t.Errorf("expected 0 model groups for un-categorized report, got %d", len(emptyRep.ModelGroups()))
	}
}

func TestLimitWindow_Classifications(t *testing.T) {
	w5h := usage.LimitWindow{Name: "5-hour limit", RemainingPct: 40}
	if !w5h.IsHourly() {
		t.Errorf("expected w5h to be classified as hourly")
	}
	if w5h.IsWeekly() {
		t.Errorf("expected w5h not to be weekly")
	}

	wwk := usage.LimitWindow{Name: "7d rolling limit", RemainingPct: 90}
	if !wwk.IsWeekly() {
		t.Errorf("expected wwk to be classified as weekly")
	}
	if wwk.IsHourly() {
		t.Errorf("expected wwk not to be hourly")
	}

	// Comprehensive variations for hourly
	hourlyNames := []string{
		"five hour", "FIVE", "5h window", "5-hour rolling", "5-h quota", "1 hour limit",
	}
	for _, name := range hourlyNames {
		w := usage.LimitWindow{Name: name}
		if !w.IsHourly() {
			t.Errorf("expected %q to be classified as hourly", name)
		}
	}

	// Comprehensive variations for weekly
	weeklyNames := []string{
		"weekly limit", "WEEK", "7d rolling", "7D", "wk quota", "WK",
	}
	for _, name := range weeklyNames {
		w := usage.LimitWindow{Name: name}
		if !w.IsWeekly() {
			t.Errorf("expected %q to be classified as weekly", name)
		}
	}

	// Nil receiver tests
	var nilWindow *usage.LimitWindow
	if nilWindow.IsHourly() {
		t.Errorf("nil window IsHourly() should return false")
	}
	if nilWindow.IsWeekly() {
		t.Errorf("nil window IsWeekly() should return false")
	}

	// Non-matching window
	other := usage.LimitWindow{Name: "Monthly Quota"}
	if other.IsHourly() {
		t.Errorf("expected Monthly Quota not to be hourly")
	}
	if other.IsWeekly() {
		t.Errorf("expected Monthly Quota not to be weekly")
	}
}

func TestReport_BottleneckPct(t *testing.T) {
	rep := usage.Report{
		Windows: []usage.LimitWindow{
			{Name: "5h", RemainingPct: 85},
			{Name: "wk", RemainingPct: 30},
		},
	}
	if got := rep.BottleneckPct(); got != 30 {
		t.Fatalf("expected bottleneck 30, got %d", got)
	}

	// Empty windows defaults to 100
	emptyRep := usage.Report{}
	if got := emptyRep.BottleneckPct(); got != 100 {
		t.Errorf("expected empty report bottleneck to be 100, got %d", got)
	}

	// Nil report receiver defaults to 100
	var nilRep *usage.Report
	if got := nilRep.BottleneckPct(); got != 100 {
		t.Errorf("expected nil report bottleneck to be 100, got %d", got)
	}

	// Clamped below 0
	underRep := usage.Report{
		Windows: []usage.LimitWindow{
			{Name: "5h", RemainingPct: -15},
		},
	}
	if got := underRep.BottleneckPct(); got != 0 {
		t.Errorf("expected clamped bottleneck 0, got %d", got)
	}

	// Clamped above 100
	overRep := usage.Report{
		Windows: []usage.LimitWindow{
			{Name: "5h", RemainingPct: 150},
		},
	}
	if got := overRep.BottleneckPct(); got != 100 {
		t.Errorf("expected clamped bottleneck 100, got %d", got)
	}
}
