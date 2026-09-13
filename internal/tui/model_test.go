package tui

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/usage"
	tea "github.com/charmbracelet/bubbletea"
)

func newTestModel(t *testing.T, profileNames ...string) Model {
	t.Helper()
	baseDir := t.TempDir()
	t.Setenv("AIM_HOME", baseDir)
	pm := profile.NewProfileManager(baseDir)
	cfg := config.NewDefaultConfig()
	for _, p := range profileNames {
		_, _ = pm.EnsureProfile(p)
		cfg.AddProfileAgent(p, "agy")
	}
	return NewModel(nil, pm, cfg)
}

func TestTUI_TabSwitchingFiltersProfiles(t *testing.T) {
	baseDir := t.TempDir()
	pm := profile.NewProfileManager(baseDir)
	_, _ = pm.EnsureProfile("work")
	_, _ = pm.EnsureProfile("bot")

	cfg := config.NewDefaultConfig()
	cfg.AddProfileAgent("bot", "agy")
	cfg.AddProfileAgent("work", "agy")
	cfg.AddProfileAgent("work", "gemini")

	reg := agents.DefaultRegistry()
	m := NewModel(reg, pm, cfg)

	// Default agent is agy: should have bot and work
	if len(m.Profiles()) != 2 {
		t.Fatalf("expected 2 profiles for agy, got %d", len(m.Profiles()))
	}

	// Press "2" to switch to gemini: should only have work
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m2 := updated.(Model)
	if m2.SelectedAgent() != "gemini" {
		t.Errorf("expected agent gemini, got %s", m2.SelectedAgent())
	}
	if len(m2.Profiles()) != 1 || m2.Profiles()[0] != "work" {
		t.Errorf("expected [work] for gemini, got %v", m2.Profiles())
	}
}

func TestTUI_MultiAgentBadge(t *testing.T) {
	baseDir := t.TempDir()
	pm := profile.NewProfileManager(baseDir)
	_, _ = pm.EnsureProfile("single")
	_, _ = pm.EnsureProfile("multi")

	cfg := config.NewDefaultConfig()
	cfg.AddProfileAgent("single", "agy")
	cfg.AddProfileAgent("multi", "agy")
	cfg.AddProfileAgent("multi", "claude")

	m := NewModel(nil, pm, cfg)
	view := m.View()

	if !strings.Contains(view, "multi [claude]") {
		t.Errorf("expected view to contain multi-agent badge 'multi [claude]', got:\n%s", view)
	}
	if strings.Contains(view, "multi [agy") {
		t.Errorf("expected active agent 'agy' to be omitted from badge, got:\n%s", view)
	}
	if strings.Contains(view, "single [") {
		t.Errorf("expected single-agent profile not to have badge, got:\n%s", view)
	}
}

func TestTUI_EmptyProfilesForAgent(t *testing.T) {
	baseDir := t.TempDir()
	pm := profile.NewProfileManager(baseDir)
	_, _ = pm.EnsureProfile("work")

	cfg := config.NewDefaultConfig()
	cfg.AddProfileAgent("work", "agy")

	m := NewModel(nil, pm, cfg)

	// Switch to gemini (which has 0 profiles)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m2 := updated.(Model)
	view := m2.View()
	expected := "(no profiles configured for gemini - press 'l' to log in)"
	if !strings.Contains(view, expected) {
		t.Errorf("expected view to contain %q, got:\n%s", expected, view)
	}
}

func TestTUIModelNavigation(t *testing.T) {
	m := newTestModel(t, "alpha", "beta")
	if m.cursor != 0 {
		t.Errorf("expected initial cursor 0, got %d", m.cursor)
	}

	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = newM.(Model)
	if m.cursor != 1 {
		t.Errorf("expected cursor 1 after down key, got %d", m.cursor)
	}

	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = newM.(Model)
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 after up key, got %d", m.cursor)
	}

	// Test tab switching
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m = newM.(Model)
	if m.agent != "gemini" {
		t.Errorf("expected agent 'gemini' after pressing 2, got '%s'", m.agent)
	}
}

func TestTUIModelBoundsAndVimKeys(t *testing.T) {
	m := newTestModel(t, "alpha", "beta")

	// Up key at 0 should remain 0
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = newM.(Model)
	if m.cursor != 0 {
		t.Errorf("expected cursor 0, got %d", m.cursor)
	}

	// 'j' moves down
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = newM.(Model)
	if m.cursor != 1 {
		t.Errorf("expected cursor 1 after 'j', got %d", m.cursor)
	}

	// 'j' at bottom remains at bottom
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = newM.(Model)
	if m.cursor != 1 {
		t.Errorf("expected cursor 1 at bottom, got %d", m.cursor)
	}

	// 'k' moves up
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = newM.(Model)
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 after 'k', got %d", m.cursor)
	}
}

func TestTUIModelTabSwitching(t *testing.T) {
	baseDir := t.TempDir()
	pm := profile.NewProfileManager(baseDir)
	cfg := config.NewDefaultConfig()
	_, _ = pm.EnsureProfile("p1")
	_, _ = pm.EnsureProfile("p2")
	cfg.AddProfileAgent("p1", "agy")
	cfg.AddProfileAgent("p2", "agy")
	cfg.AddProfileAgent("p1", "gemini")
	cfg.AddProfileAgent("p2", "gemini")
	m := NewModel(nil, pm, cfg)

	// Move cursor down
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = newM.(Model)
	if m.cursor != 1 {
		t.Fatalf("expected cursor 1, got %d", m.cursor)
	}

	// Press 'tab' -> gemini, cursor resets to 0
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = newM.(Model)
	if m.SelectedAgent() != "gemini" {
		t.Errorf("expected agent 'gemini', got '%s'", m.SelectedAgent())
	}
	if m.cursor != 0 {
		t.Errorf("expected cursor reset to 0, got %d", m.cursor)
	}

	// Press 'tab' again -> codex
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = newM.(Model)
	if m.SelectedAgent() != "codex" {
		t.Errorf("expected agent 'codex', got '%s'", m.SelectedAgent())
	}

	// Press 'tab' again -> agy
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = newM.(Model)
	if m.SelectedAgent() != "agy" {
		t.Errorf("expected agent 'agy', got '%s'", m.SelectedAgent())
	}

	// Press '3' -> codex
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	m = newM.(Model)
	if m.SelectedAgent() != "codex" {
		t.Errorf("expected agent 'codex', got '%s'", m.SelectedAgent())
	}

	// Press 'shift+tab' -> gemini
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = newM.(Model)
	if m.SelectedAgent() != "gemini" {
		t.Errorf("expected agent 'gemini' via shift+tab, got '%s'", m.SelectedAgent())
	}

	// Press '1' -> agy
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	m = newM.(Model)
	if m.SelectedAgent() != "agy" {
		t.Errorf("expected agent 'agy', got '%s'", m.SelectedAgent())
	}
}

func TestTUI_MultiAgentTabBarRendering(t *testing.T) {
	m := newTestModel(t, "alpha")
	view := m.View()

	if !strings.Contains(view, "[1] Antigravity (agy)") {
		t.Errorf("expected view to contain [1] Antigravity (agy), got:\n%s", view)
	}
	if !strings.Contains(view, "[2] Gemini") {
		t.Errorf("expected view to contain [2] Gemini, got:\n%s", view)
	}
	if !strings.Contains(view, "[3] Codex") {
		t.Errorf("expected view to contain [3] Codex, got:\n%s", view)
	}
	if !strings.Contains(view, "[4] Claude") {
		t.Errorf("expected view to contain [4] Claude, got:\n%s", view)
	}

	// Switch to codex and verify active tab
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	codexModel := newM.(Model)
	if codexModel.SelectedAgent() != "codex" {
		t.Errorf("expected selected agent 'codex', got %q", codexModel.SelectedAgent())
	}
	codexView := codexModel.View()
	if !strings.Contains(codexView, "PROFILES (codex):") {
		t.Errorf("expected view to show PROFILES (codex), got:\n%s", codexView)
	}
}

func TestTUIModelActions(t *testing.T) {
	// Test ActionRun (enter)
	m := newTestModel(t, "alpha", "beta")
	if m.Init() == nil {
		t.Errorf("Init() should return refresh command, got nil")
	}
	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newM.(Model)
	if m.Outcome() != ActionRun {
		t.Errorf("expected ActionRun, got %v", m.Outcome())
	}
	if m.SelectedProfile() != "alpha" {
		t.Errorf("expected selected profile 'alpha', got '%s'", m.SelectedProfile())
	}
	if cmd == nil {
		t.Errorf("expected non-nil tea.Quit cmd")
	}

	// Test ActionShell ('s')
	m = newTestModel(t, "alpha", "beta")
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = newM.(Model)
	newM, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = newM.(Model)
	if m.Outcome() != ActionShell {
		t.Errorf("expected ActionShell, got %v", m.Outcome())
	}
	if m.SelectedProfile() != "beta" {
		t.Errorf("expected selected profile 'beta', got '%s'", m.SelectedProfile())
	}
	if cmd == nil {
		t.Errorf("expected non-nil tea.Quit cmd")
	}

	// Test ActionLogin ('l')
	m = newTestModel(t, "alpha", "beta")
	newM, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	m = newM.(Model)
	if m.Outcome() != ActionLogin {
		t.Errorf("expected ActionLogin, got %v", m.Outcome())
	}
	if cmd == nil {
		t.Errorf("expected non-nil tea.Quit cmd")
	}

	// Test Quit ('q')
	m = newTestModel(t, "alpha", "beta")
	newM, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = newM.(Model)
	if m.Outcome() != ActionNone {
		t.Errorf("expected ActionNone, got %v", m.Outcome())
	}
	if cmd == nil {
		t.Errorf("expected non-nil tea.Quit cmd")
	}

	// Test Quit ('esc')
	m = newTestModel(t, "alpha", "beta")
	newM, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = newM.(Model)
	if m.Outcome() != ActionNone {
		t.Errorf("expected ActionNone, got %v", m.Outcome())
	}
	if cmd == nil {
		t.Errorf("expected non-nil tea.Quit cmd")
	}
}

func TestTUIModelEmptyProfiles(t *testing.T) {
	m := NewModel(nil, nil, nil)

	// Enter on empty profiles does not trigger Run
	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newM.(Model)
	if m.Outcome() != ActionNone {
		t.Errorf("expected ActionNone on empty profiles Enter, got %v", m.Outcome())
	}
	if cmd != nil {
		t.Errorf("expected nil cmd on empty profiles Enter")
	}

	// 's' on empty profiles does not trigger Shell
	newM, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = newM.(Model)
	if m.Outcome() != ActionNone {
		t.Errorf("expected ActionNone on empty profiles 's', got %v", m.Outcome())
	}
	if cmd != nil {
		t.Errorf("expected nil cmd on empty profiles 's'")
	}

	// View output includes empty prompt
	view := m.View()
	expected := "(no profiles configured for agy - press 'l' to log in)"
	if !strings.Contains(view, expected) {
		t.Errorf("expected view to contain %q, got:\n%s", expected, view)
	}
}

func TestTUIModelViewRendering(t *testing.T) {
	m := newTestModel(t, "default", "prod")
	view := m.View()

	if !strings.Contains(view, "AIM — AI Multiplexer") {
		t.Errorf("expected title in view")
	}
	if !strings.Contains(view, "https://github.com/adrijshikhar/aim") {
		t.Errorf("expected repo URL in view")
	}
	if !strings.Contains(view, "Antigravity (agy)") {
		t.Errorf("expected Antigravity in view")
	}
	if !strings.Contains(view, "Gemini") {
		t.Errorf("expected Gemini in view")
	}
	if !strings.Contains(view, "default") || !strings.Contains(view, "prod") {
		t.Errorf("expected profiles in view")
	}
	if strings.Contains(view, "(active profile)") {
		t.Errorf("expected view not to contain '(active profile)', got:\n%s", view)
	}
	if !strings.Contains(view, ">") {
		t.Errorf("expected cursor indicator '>' in view, got:\n%s", view)
	}
	if !strings.Contains(view, "[Enter]") || !strings.Contains(view, "[Tab]") {
		t.Errorf("expected hints in view")
	}
}

func TestTUIUsageUpdateAndKeybinding(t *testing.T) {
	baseDir := t.TempDir()
	pm := profile.NewProfileManager(baseDir)
	cfg := config.NewDefaultConfig()
	_, _ = pm.EnsureProfile("default")
	_, _ = pm.EnsureProfile("prod")
	cfg.AddProfileAgent("default", "agy")
	cfg.AddProfileAgent("prod", "agy")

	reg := agents.DefaultRegistry()
	m := NewModel(reg, pm, cfg)

	now := time.Now()
	repDefault := usage.Report{
		Agent:   "agy",
		Profile: "default",
		Status:  usage.StatusOK,
		Windows: []usage.LimitWindow{
			{
				Name:         "5-hour limit",
				RemainingPct: 82,
				ResetsIn:     2*time.Hour + 15*time.Minute,
				ResetsAt:     now.Add(2*time.Hour + 15*time.Minute),
			},
			{
				Name:         "Weekly limit",
				RemainingPct: 90,
				ResetsIn:     5*24*time.Hour + 14*time.Hour,
				ResetsAt:     now.Add(5*24*time.Hour + 14*time.Hour),
			},
		},
		Credits:   "$25.00",
		FetchedAt: now.Add(-5 * time.Minute),
		FromCache: true,
	}

	repProd := usage.Report{
		Agent:   "agy",
		Profile: "prod",
		Status:  usage.StatusWarning,
		Windows: []usage.LimitWindow{
			{
				Name:         "5-hour limit",
				RemainingPct: 40,
				ResetsIn:     45 * time.Minute,
				ResetsAt:     now.Add(45 * time.Minute),
			},
		},
		FetchedAt: now,
	}

	// 1. Send usageBatchMsg to Update
	updated, cmd := m.Update(usageBatchMsg{repDefault, repProd})
	if cmd != nil {
		t.Errorf("expected nil cmd on usageBatchMsg, got %v", cmd)
	}
	m = updated.(Model)

	// Check reports stored in m.reports
	if len(m.reports) != 2 {
		t.Fatalf("expected 2 reports, got %d", len(m.reports))
	}
	if rep, ok := m.reports["agy:default"]; !ok || rep.Status != usage.StatusOK {
		t.Errorf("expected agy:default report with StatusOK, got %v (ok=%v)", rep, ok)
	}

	// 2. View in wide terminal (m.width >= 85 or default 0)
	view := m.View()
	if !strings.Contains(view, "[82%]") {
		t.Errorf("expected minimal quota badge [82%%] in wide view, got:\n%s", view)
	}
	if !strings.Contains(view, "[40%]") {
		t.Errorf("expected minimal quota badge [40%%] in wide view, got:\n%s", view)
	}
	if !strings.Contains(view, "Profile Details: default") {
		t.Errorf("expected profile details header for highlighted profile 'default', got:\n%s", view)
	}
	if !strings.Contains(view, "5-hour limit:") {
		t.Errorf("expected inspector to contain '5-hour limit:', got:\n%s", view)
	}
	if !strings.Contains(view, "Weekly limit:") {
		t.Errorf("expected inspector to contain 'Weekly limit:', got:\n%s", view)
	}
	if !strings.Contains(view, "Resets At:") {
		t.Errorf("expected inspector to contain 'Resets At:', got:\n%s", view)
	}
	if !strings.Contains(view, "Credits:") || !strings.Contains(view, "$25.00") {
		t.Errorf("expected inspector to contain Credits $25.00, got:\n%s", view)
	}
	if !strings.Contains(view, "Refreshed:") || !strings.Contains(view, "5m ago") {
		t.Errorf("expected inspector to contain Refreshed 5m ago, got:\n%s", view)
	}
	if !strings.Contains(view, "[r]") || !strings.Contains(view, "Refresh Quota") {
		t.Errorf("expected hotkey hints to contain [r] Refresh Quota, got:\n%s", view)
	}

	// 3. Narrow terminal fallback (< 85 cols)
	updated, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	mNarrow := updated.(Model)
	narrowView := mNarrow.View()
	if !strings.Contains(narrowView, "[82%]") {
		t.Errorf("expected minimal badge '[82%%]', got:\n%s", narrowView)
	}

	// 4. 'r' key triggers refresh command non-destructively (preserves cache/reports while loading)
	if m.cache != nil {
		_ = m.cache.Put(repDefault)
	}
	updated, refreshCmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	mRefreshed := updated.(Model)
	if refreshCmd == nil {
		t.Fatalf("expected non-nil refresh command on 'r'")
	}
	if !mRefreshed.loading {
		t.Errorf("expected loading to be true after 'r'")
	}
	if _, ok := mRefreshed.reports["agy:default"]; !ok {
		t.Errorf("expected agy:default report to be preserved in m.reports during 'r' refresh")
	}

	// Execute refresh command
	msg := refreshCmd()
	switch msg := msg.(type) {
	case usageReportMsg, usageStreamClosedMsg:
		// OK
	case tea.BatchMsg:
		if len(msg) == 0 {
			t.Fatalf("expected non-empty batch command")
		}
	default:
		t.Fatalf("expected usageReportMsg, usageStreamClosedMsg, or BatchMsg from refreshCmd, got %T", msg)
	}
}

func TestTUIUsage_SingleWindowBadge(t *testing.T) {
	baseDir := t.TempDir()
	pm := profile.NewProfileManager(baseDir)
	cfg := config.NewDefaultConfig()
	_, _ = pm.EnsureProfile("gemini-only")
	cfg.AddProfileAgent("gemini-only", "agy")

	reg := agents.DefaultRegistry()
	m := NewModel(reg, pm, cfg)

	now := time.Now()
	rep := usage.Report{
		Agent:   "agy",
		Profile: "gemini-only",
		Status:  usage.StatusOK,
		Windows: []usage.LimitWindow{
			{
				Name:         "Weekly limit",
				RemainingPct: 75,
				ResetsIn:     3*24*time.Hour + 8*time.Hour,
				ResetsAt:     now.Add(3*24*time.Hour + 8*time.Hour),
			},
		},
		FetchedAt: now,
	}

	updated, _ := m.Update(usageBatchMsg{rep})
	m = updated.(Model)

	view := m.View()
	if !strings.Contains(view, "[75%]") {
		t.Errorf("expected single window badge '[75%%]', got:\n%s", view)
	}
	// Verify inspector: for weekly-only profile, Primary Limit is omitted and Weekly limit is populated
	if strings.Contains(view, "Primary Limit") {
		t.Errorf("expected inspector not to contain 'Primary Limit' for single weekly window profile, got:\n%s", view)
	}
	if !strings.Contains(view, "Weekly limit:  [███████░░░] 75% (resets in 3d 8h)") {
		t.Errorf("expected inspector Weekly Limit to show weekly quota, got:\n%s", view)
	}
}

func TestTUIUsage_FullCapacityOmitResetCountdown(t *testing.T) {
	baseDir := t.TempDir()
	pm := profile.NewProfileManager(baseDir)
	cfg := config.NewDefaultConfig()
	_, _ = pm.EnsureProfile("full-cap")
	cfg.AddProfileAgent("full-cap", "agy")

	reg := agents.DefaultRegistry()
	m := NewModel(reg, pm, cfg)

	now := time.Now()
	rep := usage.Report{
		Agent:   "agy",
		Profile: "full-cap",
		Status:  usage.StatusOK,
		Windows: []usage.LimitWindow{
			{
				Name:         "5h limit",
				RemainingPct: 100,
				ResetsIn:     2 * time.Hour,
				ResetsAt:     now.Add(2 * time.Hour),
			},
			{
				Name:         "Weekly limit",
				RemainingPct: 100,
				ResetsIn:     5 * 24 * time.Hour,
				ResetsAt:     now.Add(5 * 24 * time.Hour),
			},
		},
		FetchedAt: now,
	}

	updated, _ := m.Update(usageBatchMsg{rep})
	m = updated.(Model)

	// Wide mode (default)
	wideView := m.View()
	if !strings.Contains(wideView, "[100%]") {
		t.Errorf("expected minimal badge without countdown '[100%%]', got:\n%s", wideView)
	}
	if !strings.Contains(wideView, "5h limit:      [██████████] 100%") {
		t.Errorf("expected inspector to omit countdown at 100%% capacity, got:\n%s", wideView)
	}
	if strings.Contains(wideView, "(2h)") || strings.Contains(wideView, "(5d)") {
		t.Errorf("expected inspector to omit countdowns at 100%% capacity, got:\n%s", wideView)
	}

	// Narrow mode (< 85 cols)
	updated, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	mNarrow := updated.(Model)
	narrowView := mNarrow.View()
	if !strings.Contains(narrowView, "[100%]") {
		t.Errorf("expected narrow badge without countdown '[100%%]', got:\n%s", narrowView)
	}
	if strings.Contains(narrowView, "(2h)") {
		t.Errorf("expected narrow badge to omit countdown at 100%% capacity, got:\n%s", narrowView)
	}
}

func TestTUIUsage_GaugeColorStyles(t *testing.T) {
	if GaugeGreenStyle.GetForeground() != GaugeGreenStyle.Foreground(StatusGreen).GetForeground() {
		t.Errorf("expected GaugeGreenStyle color StatusGreen")
	}
	if GaugeYellowStyle.GetForeground() != GaugeYellowStyle.Foreground(StatusYellow).GetForeground() {
		t.Errorf("expected GaugeYellowStyle color StatusYellow")
	}
	if GaugeRedStyle.GetForeground() != GaugeRedStyle.Foreground(StatusRed).GetForeground() {
		t.Errorf("expected GaugeRedStyle color StatusRed")
	}
	if GaugeDimStyle.GetForeground() != GaugeDimStyle.Foreground(StatusDim).GetForeground() {
		t.Errorf("expected GaugeDimStyle color StatusDim")
	}

	if GaugeStyleForStatus(usage.StatusOK).GetForeground() != GaugeGreenStyle.GetForeground() {
		t.Errorf("expected GaugeGreenStyle for StatusOK")
	}
	if GaugeStyleForStatus(usage.StatusWarning).GetForeground() != GaugeYellowStyle.GetForeground() {
		t.Errorf("expected GaugeYellowStyle for StatusWarning")
	}
	if GaugeStyleForStatus(usage.StatusCritical).GetForeground() != GaugeRedStyle.GetForeground() {
		t.Errorf("expected GaugeRedStyle for StatusCritical")
	}
	if GaugeStyleForStatus(usage.StatusExhausted).GetForeground() != GaugeRedStyle.GetForeground() {
		t.Errorf("expected GaugeRedStyle for StatusExhausted")
	}
	if GaugeStyleForStatus(usage.StatusUnknown).GetForeground() != GaugeDimStyle.GetForeground() {
		t.Errorf("expected GaugeDimStyle for StatusUnknown")
	}
}

func TestTUIUsage_TabSwitchTriggersRefresh(t *testing.T) {
	baseDir := t.TempDir()
	pm := profile.NewProfileManager(baseDir)
	cfg := config.NewDefaultConfig()
	_, _ = pm.EnsureProfile("p1")
	cfg.AddProfileAgent("p1", "agy")
	cfg.AddProfileAgent("p1", "gemini")

	reg := agents.DefaultRegistry()
	m := NewModel(reg, pm, cfg)

	// Init() should return refreshCmd
	if m.Init() == nil {
		t.Errorf("expected non-nil cmd on Init()")
	}

	// '2' should switch to gemini and return refreshCmd
	_, cmd2 := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	if cmd2 == nil {
		t.Errorf("expected non-nil cmd on pressing '2'")
	}

	// 'tab' should switch to gemini and return refreshCmd
	_, cmdTab := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if cmdTab == nil {
		t.Errorf("expected non-nil cmd on pressing 'tab'")
	}

	// '1' should switch to agy and return refreshCmd
	_, cmd1 := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	if cmd1 == nil {
		t.Errorf("expected non-nil cmd on pressing '1'")
	}
}

func TestTUIUsage_TrueStreaming(t *testing.T) {
	baseDir := t.TempDir()
	pm := profile.NewProfileManager(baseDir)
	cfg := config.NewDefaultConfig()
	_, _ = pm.EnsureProfile("p1")
	_, _ = pm.EnsureProfile("p2")
	cfg.AddProfileAgent("p1", "agy")
	cfg.AddProfileAgent("p2", "agy")

	reg := agents.DefaultRegistry()
	m := NewModel(reg, pm, cfg)

	ch := make(chan usage.Report, 2)
	m.usageStream.ch = ch

	// Wait on stream
	cmd := waitForUsageReport(ch)

	r1 := usage.Report{
		Agent:   "agy",
		Profile: "p1",
		Status:  usage.StatusOK,
		Windows: []usage.LimitWindow{
			{Name: "5h", RemainingPct: 90},
		},
	}
	ch <- r1

	msg1 := cmd()
	repMsg, ok := msg1.(usageReportMsg)
	if !ok {
		t.Fatalf("expected usageReportMsg, got %T", msg1)
	}

	updated, nextCmd := m.Update(repMsg)
	mUpdated := updated.(Model)
	if _, found := mUpdated.reports["agy:p1"]; !found {
		t.Fatalf("expected report agy:p1 to be stored incrementally")
	}
	if nextCmd == nil {
		t.Fatalf("expected non-nil nextCmd for continuation of streaming")
	}

	r2 := usage.Report{
		Agent:   "agy",
		Profile: "p2",
		Status:  usage.StatusWarning,
		Windows: []usage.LimitWindow{
			{Name: "5h", RemainingPct: 40},
		},
	}
	ch <- r2
	close(ch)

	msg2 := nextCmd()
	repMsg2, ok := msg2.(usageReportMsg)
	if !ok {
		t.Fatalf("expected usageReportMsg for second item, got %T", msg2)
	}

	updated2, finalCmd := mUpdated.Update(repMsg2)
	mUpdated2 := updated2.(Model)
	if _, found := mUpdated2.reports["agy:p2"]; !found {
		t.Fatalf("expected report agy:p2 to be stored incrementally")
	}
	if finalCmd == nil {
		t.Fatalf("expected finalCmd waiting for stream close")
	}

	msgClose := finalCmd()
	if _, isClosed := msgClose.(usageStreamClosedMsg); !isClosed {
		t.Fatalf("expected usageStreamClosedMsg when channel closes, got %T", msgClose)
	}

	updatedDone, nilCmd := mUpdated2.Update(msgClose)
	_ = updatedDone
	if nilCmd != nil {
		t.Fatalf("expected nil cmd after usageStreamClosedMsg")
	}
}

func TestTUI_MultiModelInspectorBreakdown(t *testing.T) {
	m := newTestModel(t, "bot")
	m.reports = map[string]usage.Report{
		"agy:bot": {
			Agent:   "agy",
			Profile: "bot",
			Status:  usage.StatusWarning,
			Windows: []usage.LimitWindow{
				{Category: "Gemini Models", Name: "5-hour limit", RemainingPct: 23, ResetsIn: 2*time.Hour + 15*time.Minute},
				{Category: "Gemini Models", Name: "weekly limit", RemainingPct: 67, ResetsIn: 5*24*time.Hour + 14*time.Hour},
				{Category: "Claude and GPT models", Name: "5-hour limit", RemainingPct: 100, ResetsIn: 0},
				{Category: "Claude and GPT models", Name: "weekly limit", RemainingPct: 84, ResetsIn: 4*24*time.Hour + 12*time.Hour},
			},
			Credits: "150",
		},
	}

	view := m.View()

	// Should contain model rows with 5h and Wk side-by-side
	if !strings.Contains(view, "Gemini:") {
		t.Errorf("expected view to contain 'Gemini:', got:\n%s", view)
	}
	if !strings.Contains(view, "Claude & GPT:") {
		t.Errorf("expected view to contain 'Claude & GPT:', got:\n%s", view)
	}
	// Progress bars and values should be rendered
	if !strings.Contains(view, "5h: [██░░░░░░░░] 23% (2h 15m)") {
		t.Errorf("expected view to contain '5h: [██░░░░░░░░] 23%% (2h 15m)', got:\n%s", view)
	}
	if !strings.Contains(view, "Wk: [██████░░░░] 67% (5d 14h)") {
		t.Errorf("expected view to contain 'Wk: [██████░░░░] 67%% (5d 14h)', got:\n%s", view)
	}
	if !strings.Contains(view, "5h: [██████████] 100%") {
		t.Errorf("expected view to contain full capacity '5h: [██████████] 100%%', got:\n%s", view)
	}
	if !strings.Contains(view, "Wk: [████████░░] 84% (4d 12h)") {
		t.Errorf("expected view to contain 'Wk: [████████░░] 84%% (4d 12h)', got:\n%s", view)
	}
	if !strings.Contains(view, "Credits:") || !strings.Contains(view, "150") {
		t.Errorf("expected view to contain 'Credits:' and '150', got:\n%s", view)
	}
	// Should NOT contain the old "Primary Limit:" or "Weekly Limit:" lines when multi-model is present
	if strings.Contains(view, "Primary Limit:") {
		t.Errorf("expected view NOT to contain 'Primary Limit:' for multi-model profile, got:\n%s", view)
	}
}

func TestTUI_PeriodicAutoRefreshTick(t *testing.T) {
	m := newTestModel(t, "work")

	// Init should return a batch cmd (refresh + ticker)
	cmd := m.Init()
	if cmd == nil {
		t.Fatalf("expected non-nil cmd from Init()")
	}

	// Update with tickMsg should return a non-nil cmd
	updated, nextCmd := m.Update(tickMsg(time.Now()))
	if updated == nil {
		t.Fatalf("expected non-nil updated model")
	}
	if nextCmd == nil {
		t.Fatalf("expected non-nil nextCmd from tickMsg")
	}
}

func TestTUI_SpinnerAnimation(t *testing.T) {
	m := newTestModel(t, "work")
	m.loading = true
	m.spinnerIdx = 0

	// View should render spinner next to [r] when loading
	view := m.View()
	if !strings.Contains(view, spinnerFrames[0]) {
		t.Errorf("expected view to contain spinner frame '%s', got:\n%s", spinnerFrames[0], view)
	}

	// spinnerTickMsg should advance spinner frame
	updated, cmd := m.Update(spinnerTickMsg(time.Now()))
	m2 := updated.(Model)
	if m2.spinnerIdx != 1 {
		t.Errorf("expected spinnerIdx to advance to 1, got %d", m2.spinnerIdx)
	}
	if cmd == nil {
		t.Fatalf("expected next spinner tick command")
	}

	// spinner.Tick() should advance m.spinner.View() to next frame
	updatedSpin, cmdSpin := m.Update(m.spinner.Tick())
	mSpin := updatedSpin.(Model)
	if cmdSpin == nil {
		t.Fatalf("expected next spinner tick command from bubbles spinner")
	}
	viewSpin := mSpin.View()
	if !strings.Contains(viewSpin, spinnerFrames[1]) {
		t.Errorf("expected view to contain advanced spinner frame '%s', got:\n%s", spinnerFrames[1], viewSpin)
	}

	// usageStreamClosedMsg should stop loading
	updatedDone, cmdDone := m2.Update(usageStreamClosedMsg{})
	mDone := updatedDone.(Model)
	if mDone.loading {
		t.Errorf("expected loading to be false after stream closed")
	}
	if cmdDone != nil {
		t.Errorf("expected nil cmd after stream closed")
	}
}

func TestTUI_DeleteModal_TriggerAndCancel(t *testing.T) {
	m := newTestModel(t, "alpha", "beta")

	// Pressing 'x' opens the delete modal for current cursor (alpha)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	m2 := updated.(Model)

	if !m2.IsDeleteModalActive() {
		t.Fatalf("expected delete modal to be active after pressing 'x'")
	}
	if m2.DeleteModalTarget() != "alpha" {
		t.Errorf("expected modal target to be 'alpha', got %q", m2.DeleteModalTarget())
	}
	if m2.DeleteModalFocusedIndex() != 1 {
		t.Errorf("expected default focus on Cancel (index 1), got %d", m2.DeleteModalFocusedIndex())
	}

	// Pressing Esc cancels and closes modal without deleting
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m3 := updated2.(Model)

	if m3.IsDeleteModalActive() {
		t.Errorf("expected modal to be inactive after Esc")
	}
	if len(m3.Profiles()) != 2 {
		t.Errorf("expected profiles to remain intact, got %v", m3.Profiles())
	}

	// Also test cancelling via 'n'
	updatedOpen, _ := m3.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	mOpen := updatedOpen.(Model)
	updatedCancelN, _ := mOpen.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	mCancelN := updatedCancelN.(Model)
	if mCancelN.IsDeleteModalActive() {
		t.Errorf("expected modal to be inactive after 'n'")
	}
}

func TestTUI_DeleteModal_SingleAgent_Confirm(t *testing.T) {
	m := newTestModel(t, "work")

	// Press 'x' to open modal
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	m2 := updated.(Model)

	// Navigate left to select [Delete Profile] (index 0)
	updatedNav, _ := m2.Update(tea.KeyMsg{Type: tea.KeyLeft})
	mNav := updatedNav.(Model)
	if mNav.DeleteModalFocusedIndex() != 0 {
		t.Fatalf("expected focus index 0 (Delete Profile), got %d", mNav.DeleteModalFocusedIndex())
	}

	// Press Enter to confirm delete
	updatedDel, _ := mNav.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mDel := updatedDel.(Model)

	if mDel.IsDeleteModalActive() {
		t.Errorf("expected modal to close after confirm")
	}
	if len(mDel.Profiles()) != 0 {
		t.Errorf("expected 0 profiles after deletion, got %v", mDel.Profiles())
	}
}

func TestTUI_DeleteModal_SingleAgent_QuickConfirmY(t *testing.T) {
	m := newTestModel(t, "quick_del")

	// Press 'x' then 'y'
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	m2 := updated.(Model)
	updatedY, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	mY := updatedY.(Model)

	if mY.IsDeleteModalActive() {
		t.Errorf("expected modal to close after 'y'")
	}
	if len(mY.Profiles()) != 0 {
		t.Errorf("expected profile to be deleted via 'y'")
	}
}

func TestTUI_DeleteModal_SharedProfile_UnlinkAndEntireDelete(t *testing.T) {
	baseDir := t.TempDir()
	t.Setenv("AIM_HOME", baseDir)
	pm := profile.NewProfileManager(baseDir)
	_, _ = pm.EnsureProfile("shared")

	cfg := config.NewDefaultConfig()
	cfg.AddProfileAgent("shared", "agy")
	cfg.AddProfileAgent("shared", "gemini")

	m := NewModel(nil, pm, cfg)

	// 1. Open modal on agy tab for 'shared'
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	mOpen := updated.(Model)

	if !mOpen.IsDeleteModalActive() {
		t.Fatalf("expected modal to be active")
	}
	if mOpen.DeleteModalFocusedIndex() != 2 {
		t.Errorf("expected default focus on Cancel (index 2 for shared), got %d", mOpen.DeleteModalFocusedIndex())
	}

	// Press '1' to unlink agy only
	updatedUnlink, _ := mOpen.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	mUnlinked := updatedUnlink.(Model)

	// agy profiles should now be empty
	if len(mUnlinked.Profiles()) != 0 {
		t.Errorf("expected 0 profiles for agy after unlinking, got %v", mUnlinked.Profiles())
	}
	// config should still have shared profile with gemini
	if !cfg.HasAgent("shared", "gemini") {
		t.Errorf("expected gemini to remain on shared profile")
	}

	// 2. Switch to gemini (which now has shared as single agent)
	updatedGem, _ := mUnlinked.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	mGem := updatedGem.(Model)
	if len(mGem.Profiles()) != 1 || mGem.Profiles()[0] != "shared" {
		t.Fatalf("expected [shared] for gemini, got %v", mGem.Profiles())
	}

	// Delete shared from gemini via 'x' and 'y'
	updatedX, _ := mGem.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	mGemX := updatedX.(Model)
	updatedDelGem, _ := mGemX.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	mGemDel := updatedDelGem.(Model)

	if len(mGemDel.Profiles()) != 0 {
		t.Errorf("expected 0 profiles for gemini after final delete")
	}
}

func TestTUI_DeleteModal_CursorClamping(t *testing.T) {
	m := newTestModel(t, "first", "second")

	// Move cursor to second (index 1)
	mDown, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m2 := mDown.(Model)
	if m2.cursor != 1 {
		t.Fatalf("expected cursor at 1, got %d", m2.cursor)
	}

	// Delete 'second'
	mOpen, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	mDeleted, _ := mOpen.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	mFinal := mDeleted.(Model)

	if len(mFinal.Profiles()) != 1 || mFinal.Profiles()[0] != "first" {
		t.Fatalf("expected only 'first' remaining, got %v", mFinal.Profiles())
	}
	if mFinal.cursor != 0 {
		t.Errorf("expected cursor clamped to 0, got %d", mFinal.cursor)
	}
}

func TestTUI_DeleteModal_EmptyList_NoOp(t *testing.T) {
	m := newTestModel(t) // 0 profiles
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	m2 := updated.(Model)

	if m2.IsDeleteModalActive() {
		t.Errorf("expected modal NOT to activate when profile list is empty")
	}
}

func TestTUI_DeleteModal_ViewRendering(t *testing.T) {
	m := newTestModel(t, "my-profile")

	// In normal view, hint bar should contain [x] Delete
	normalView := m.View()
	if !strings.Contains(normalView, "[x]") || !strings.Contains(normalView, "Delete") {
		t.Errorf("expected normal view to contain [x] Delete hint, got:\n%s", normalView)
	}

	// Open delete modal
	mOpen, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	modalView := mOpen.(Model).View()

	if !strings.Contains(modalView, "Confirm Deletion: my-profile") {
		t.Errorf("expected modal view to contain 'Confirm Deletion: my-profile', got:\n%s", modalView)
	}
	if !strings.Contains(modalView, "[ Delete Profile ]") || !strings.Contains(modalView, "[ Cancel ]") {
		t.Errorf("expected modal view to render buttons, got:\n%s", modalView)
	}
}

func TestTUI_DoctorDrawer_ToggleAndClose(t *testing.T) {
	m := newTestModel(t, "alpha")

	// 1. Press 'd' to open drawer
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m2 := updated.(Model)

	if !m2.IsDoctorDrawerActive() {
		t.Fatalf("expected doctor drawer to be active after 'd'")
	}
	if m2.DoctorDrawerTargetProfile() != "alpha" {
		t.Errorf("expected target profile 'alpha', got %q", m2.DoctorDrawerTargetProfile())
	}
	if len(m2.DoctorDrawerResults()) == 0 {
		t.Errorf("expected non-empty doctor results")
	}

	// 2. Press 'd' again to toggle close
	updatedCloseD, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	mCloseD := updatedCloseD.(Model)
	if mCloseD.IsDoctorDrawerActive() {
		t.Errorf("expected doctor drawer to close after pressing 'd' again")
	}

	// 3. Re-open and close with Esc
	mReopened, _ := mCloseD.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	mEsc, _ := mReopened.(Model).Update(tea.KeyMsg{Type: tea.KeyEsc})
	if mEsc.(Model).IsDoctorDrawerActive() {
		t.Errorf("expected doctor drawer to close after Esc")
	}

	// 4. Re-open and close with 'q'
	mReopened2, _ := mEsc.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	mQ, _ := mReopened2.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if mQ.(Model).IsDoctorDrawerActive() {
		t.Errorf("expected doctor drawer to close after 'q'")
	}
}

func TestTUI_DoctorDrawer_NavigationUpdatesDiagnostics(t *testing.T) {
	m := newTestModel(t, "first", "second")

	// Open drawer on first
	mOpen, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m1 := mOpen.(Model)
	if m1.DoctorDrawerTargetProfile() != "first" {
		t.Fatalf("expected first profile, got %s", m1.DoctorDrawerTargetProfile())
	}

	// Navigate down while drawer is open
	mDown, _ := m1.Update(tea.KeyMsg{Type: tea.KeyDown})
	m2 := mDown.(Model)
	if m2.DoctorDrawerTargetProfile() != "second" {
		t.Errorf("expected drawer target to update to 'second', got %s", m2.DoctorDrawerTargetProfile())
	}

	// Navigate back up
	mUp, _ := m2.Update(tea.KeyMsg{Type: tea.KeyUp})
	m3 := mUp.(Model)
	if m3.DoctorDrawerTargetProfile() != "first" {
		t.Errorf("expected drawer target to update back to 'first', got %s", m3.DoctorDrawerTargetProfile())
	}
}

func TestTUI_DoctorDrawer_TabSwitchUpdatesDiagnostics(t *testing.T) {
	baseDir := t.TempDir()
	pm := profile.NewProfileManager(baseDir)
	_, _ = pm.EnsureProfile("work")

	cfg := config.NewDefaultConfig()
	cfg.AddProfileAgent("work", "agy")
	cfg.AddProfileAgent("work", "gemini")

	reg := agents.DefaultRegistry()
	m := NewModel(reg, pm, cfg)

	// Open drawer on agy
	mOpen, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m1 := mOpen.(Model)
	if m1.SelectedAgent() != "agy" {
		t.Fatalf("expected agy agent")
	}

	// Switch to gemini while drawer is open
	mSwitch, _ := m1.Update(tea.KeyMsg{Type: tea.KeyTab})
	m2 := mSwitch.(Model)
	if m2.SelectedAgent() != "gemini" {
		t.Fatalf("expected gemini agent after Tab, got %s", m2.SelectedAgent())
	}
	if !m2.IsDoctorDrawerActive() {
		t.Errorf("expected doctor drawer to stay open after Tab")
	}
}

func TestTUI_DoctorDrawer_ViewRendering(t *testing.T) {
	m := newTestModel(t, "alpha")

	// Check hint bar in normal view
	normalView := m.View()
	if !strings.Contains(normalView, "[d]") || !strings.Contains(normalView, "Doctor") {
		t.Errorf("expected normal view to contain [d] Doctor, got:\n%s", normalView)
	}

	// Open drawer and check view
	mOpen, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	drawerView := mOpen.(Model).View()

	if !strings.Contains(drawerView, "Diagnostics: agy / alpha") {
		t.Errorf("expected drawer view to contain 'Diagnostics: agy / alpha', got:\n%s", drawerView)
	}
	if !strings.Contains(drawerView, "Close Drawer") {
		t.Errorf("expected drawer view to contain 'Close Drawer', got:\n%s", drawerView)
	}
}

func TestTUI_DoctorDrawer_EmptyProfiles(t *testing.T) {
	m := newTestModel(t) // 0 profiles
	mOpen, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m2 := mOpen.(Model)

	if !m2.IsDoctorDrawerActive() {
		t.Fatalf("expected drawer to open even on empty profiles")
	}
	drawerView := m2.View()
	if !strings.Contains(drawerView, "No profiles configured") {
		t.Errorf("expected view to contain 'No profiles configured', got:\n%s", drawerView)
	}
}

func TestTUI_DoctorDrawer_ConfigOverrides(t *testing.T) {
	tempDir := t.TempDir()
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("overridden")

	cfg := config.NewDefaultConfig()
	cfg.AddProfileAgent("overridden", "agy")
	cfg.SetProfileEnv("overridden", map[string]string{
		"MY_ENV": "1",
	})
	cfg.SetProfileArgs("overridden", []string{"--flag1", "--flag2"})

	reg := agents.DefaultRegistry()

	m := NewModel(reg, pm, cfg)
	mOpen, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m2 := mOpen.(Model)

	results := m2.DoctorDrawerResults()
	hasEnvMsg := false
	hasArgMsg := false
	for _, r := range results {
		if r.Category == "Config" && strings.Contains(r.Message, "1 custom env var(s)") {
			hasEnvMsg = true
		}
		if r.Category == "Config" && strings.Contains(r.Message, "2 custom launch arg(s)") {
			hasArgMsg = true
		}
	}
	if !hasEnvMsg {
		t.Errorf("expected doctor drawer to have config env result, got %v", results)
	}
	if !hasArgMsg {
		t.Errorf("expected doctor drawer to have config args result, got %v", results)
	}
}

func TestTUI_RenameModal_TriggerAndCancel(t *testing.T) {
	m := newTestModel(t, "alpha", "beta")

	// Press 'm' to open rename modal
	mOpen, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	m2 := mOpen.(Model)

	if !m2.IsRenameModalActive() {
		t.Fatalf("expected rename modal to be active after pressing 'm'")
	}
	if m2.RenameModalTarget() != "alpha" {
		t.Errorf("expected target 'alpha', got %q", m2.RenameModalTarget())
	}

	// Press 'esc' to cancel
	mCancel, _ := m2.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m3 := mCancel.(Model)

	if m3.IsRenameModalActive() {
		t.Errorf("expected rename modal to be closed after 'esc'")
	}

	// Test trigger with 'R'
	mOpenR, _ := m3.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'R'}})
	m4 := mOpenR.(Model)
	if !m4.IsRenameModalActive() {
		t.Fatalf("expected rename modal to be active after pressing 'R'")
	}
}

func TestTUI_RenameModal_ValidationErrors(t *testing.T) {
	m := newTestModel(t, "alpha", "beta")

	// Open modal
	mOpen, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	m2 := mOpen.(Model)

	// Press Enter without typing anything
	mEmpty, _ := m2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m3 := mEmpty.(Model)
	if !m3.IsRenameModalActive() {
		t.Fatalf("expected modal to remain active on empty input")
	}
	if m3.RenameModalError() != "Profile name cannot be empty" {
		t.Errorf("expected empty error message, got %q", m3.RenameModalError())
	}

	// Type "alpha" (same name)
	mTyping := m3
	for _, r := range "alpha" {
		up, _ := mTyping.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		mTyping = up.(Model)
	}
	mSame, _ := mTyping.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m4 := mSame.(Model)
	if !m4.IsRenameModalActive() {
		t.Fatalf("expected modal to remain active on same name")
	}
	if m4.RenameModalError() != "New profile name must be different from current name" {
		t.Errorf("expected same name error, got %q", m4.RenameModalError())
	}

	// Type "../invalid" (path traversal)
	mTyping = m4
	mTyping.renameModal.input.SetValue("../invalid")
	mTrav, _ := mTyping.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m5 := mTrav.(Model)
	if !m5.IsRenameModalActive() {
		t.Fatalf("expected modal to remain active on traversal input")
	}
	if !strings.Contains(m5.RenameModalError(), "cannot contain slashes") {
		t.Errorf("expected slash validation error, got %q", m5.RenameModalError())
	}
}

func TestTUI_RenameModal_Success(t *testing.T) {
	m := newTestModel(t, "alpha", "beta")

	// Set mock reports for multiple agents in reports and cache
	m.reports["agy:alpha"] = usage.Report{
		Agent:   "agy",
		Profile: "alpha",
		Status:  usage.StatusOK,
	}
	m.reports["gemini:alpha"] = usage.Report{
		Agent:   "gemini",
		Profile: "alpha",
		Status:  usage.StatusOK,
	}
	if m.cache != nil {
		_ = m.cache.Put(usage.Report{
			Agent:   "agy",
			Profile: "alpha",
			Status:  usage.StatusOK,
		})
		_ = m.cache.Put(usage.Report{
			Agent:   "gemini",
			Profile: "alpha",
			Status:  usage.StatusOK,
		})
	}

	// Open modal on "alpha"
	mOpen, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	mCurrent := mOpen.(Model)

	// Type "charlie"
	for _, r := range "charlie" {
		up, _ := mCurrent.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		mCurrent = up.(Model)
	}

	if mCurrent.RenameModalInputValue() != "charlie" {
		t.Fatalf("expected input value 'charlie', got %q", mCurrent.RenameModalInputValue())
	}

	// Press Enter to submit
	mSubmit, _ := mCurrent.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mDone := mSubmit.(Model)

	if mDone.IsRenameModalActive() {
		t.Fatalf("expected modal to close after successful rename, err: %s", mDone.RenameModalError())
	}

	// Profiles should contain "beta" and "charlie", but not "alpha"
	foundCharlie := false
	foundAlpha := false
	for _, p := range mDone.Profiles() {
		if p == "charlie" {
			foundCharlie = true
		}
		if p == "alpha" {
			foundAlpha = true
		}
	}
	if !foundCharlie || foundAlpha {
		t.Errorf("expected profiles to contain 'charlie' and not 'alpha', got %v", mDone.Profiles())
	}

	// Cursor should point to "charlie"
	if mDone.Profiles()[mDone.cursor] != "charlie" {
		t.Errorf("expected cursor on 'charlie', got %q", mDone.Profiles()[mDone.cursor])
	}

	// Cache and reports should be updated to "charlie" for ALL agents
	if _, ok := mDone.reports["agy:charlie"]; !ok {
		t.Errorf("expected report for 'agy:charlie' to exist")
	}
	if _, ok := mDone.reports["agy:alpha"]; ok {
		t.Errorf("expected old report for 'agy:alpha' to be deleted")
	}
	if _, ok := mDone.reports["gemini:charlie"]; !ok {
		t.Errorf("expected report for 'gemini:charlie' to exist")
	}
	if _, ok := mDone.reports["gemini:alpha"]; ok {
		t.Errorf("expected old report for 'gemini:alpha' to be deleted")
	}
	if mDone.cache != nil {
		if _, found := mDone.cache.Get("agy", "charlie"); !found {
			t.Errorf("expected cache entry for 'agy:charlie'")
		}
		if _, found := mDone.cache.Get("agy", "alpha"); found {
			t.Errorf("expected cache entry for 'agy:alpha' to be deleted")
		}
		if _, found := mDone.cache.Get("gemini", "charlie"); !found {
			t.Errorf("expected cache entry for 'gemini:charlie'")
		}
		if _, found := mDone.cache.Get("gemini", "alpha"); found {
			t.Errorf("expected cache entry for 'gemini:alpha' to be deleted")
		}
	}
}

func TestTUI_RenameModal_EmptyList_NoOp(t *testing.T) {
	m := newTestModel(t) // 0 profiles
	mOpen, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	m2 := mOpen.(Model)

	if m2.IsRenameModalActive() {
		t.Errorf("expected rename modal not to open on empty profile list")
	}
}

func TestTUI_RenameModal_ViewRendering(t *testing.T) {
	m := newTestModel(t, "alpha")

	// Check footer hints in normal view
	normalView := m.View()
	if !strings.Contains(normalView, "[m]") || !strings.Contains(normalView, "Rename") {
		t.Errorf("expected normal view to contain [m] Rename, got:\n%s", normalView)
	}

	// Open rename modal
	mOpen, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	modalView := mOpen.(Model).View()

	if !strings.Contains(modalView, "Rename Profile: alpha") {
		t.Errorf("expected modal view to contain 'Rename Profile: alpha', got:\n%s", modalView)
	}
	if !strings.Contains(modalView, "Enter new name for profile:") {
		t.Errorf("expected modal view to contain input prompt, got:\n%s", modalView)
	}
	if !strings.Contains(modalView, "[Enter] Confirm") || !strings.Contains(modalView, "[Esc] Cancel") {
		t.Errorf("expected modal view to contain confirm/cancel hints, got:\n%s", modalView)
	}
}

func TestTUI_ProfileDetails_AccountEmailDisplay(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("AIM_HOME", tmpDir)
	pm := profile.NewProfileManager(tmpDir)
	cfg := config.NewDefaultConfig()

	pDir, err := pm.EnsureProfile("myprofile")
	if err != nil {
		t.Fatalf("EnsureProfile failed: %v", err)
	}
	cfg.AddProfileAgent("myprofile", "agy")

	// Write token with mock JWT id_token containing email
	claimsJSON := `{"email":"engineer@company.com","name":"Alice Engineer"}`
	b64Claims := base64.RawURLEncoding.EncodeToString([]byte(claimsJSON))
	mockToken := fmt.Sprintf(`{"token":{"access_token":"ya29.mock"},"auth_method":"consumer","id_token":"header.%s.sig"}`, b64Claims)
	tokenFile := filepath.Join(pDir, ".gemini", "antigravity-cli", "antigravity-oauth-token")
	_ = os.MkdirAll(filepath.Dir(tokenFile), 0700)
	_ = os.WriteFile(tokenFile, []byte(mockToken), 0600)

	reg := agents.DefaultRegistry()
	m := NewModel(reg, pm, cfg)

	view := m.View()
	if !strings.Contains(view, "Profile Details: myprofile") {
		t.Errorf("expected view to display 'Profile Details: myprofile', got:\n%s", view)
	}
	if !strings.Contains(view, "engineer@company.com") {
		t.Errorf("expected view to display account email 'engineer@company.com', got:\n%s", view)
	}
	if !strings.Contains(view, "Alice Engineer") {
		t.Errorf("expected view to display account name 'Alice Engineer', got:\n%s", view)
	}
	if !strings.Contains(view, "Google OAuth (consumer)") {
		t.Errorf("expected view to display auth method, got:\n%s", view)
	}
}

func TestTUI_NoProfiles_NoDefaultProfileLoaded(t *testing.T) {
	emptyDir := t.TempDir()
	t.Setenv("AIM_HOME", emptyDir)
	pm := profile.NewProfileManager(emptyDir)
	cfg := config.NewDefaultConfig()

	reg := agents.DefaultRegistry()
	m := NewModel(reg, pm, cfg)

	if len(m.Profiles()) != 0 {
		t.Errorf("expected 0 profiles when directory is empty, got %v", m.Profiles())
	}

	view := m.View()
	if strings.Contains(view, "default") {
		t.Errorf("expected view not to contain 'default' profile, got:\n%s", view)
	}
	if !strings.Contains(view, "no profiles configured for agy") {
		t.Errorf("expected view to indicate no profiles configured, got:\n%s", view)
	}
}

func TestTUI_FormatBadge_BottleneckWindow(t *testing.T) {
	rep := usage.Report{
		Status: usage.StatusWarning,
		Windows: []usage.LimitWindow{
			{Category: "Gemini", Name: "Five Hour", RemainingPct: 98},
			{Category: "Claude", Name: "Five Hour", RemainingPct: 25},
			{Category: "Gemini", Name: "Weekly", RemainingPct: 90},
		},
	}
	badge := formatBadge(rep, false)
	if badge != "[25%]" {
		t.Errorf("expected badge to reflect bottleneck 25%%, got %s", badge)
	}
}

func TestTUI_Header_VersionDisplay(t *testing.T) {
	baseDir := t.TempDir()
	pm := profile.NewProfileManager(baseDir)
	cfg := config.NewDefaultConfig()
	reg := agents.DefaultRegistry()

	// Default model inherits package Version
	m := NewModel(reg, pm, cfg)
	if m.Version() != Version {
		t.Errorf("expected default model version to be %q, got %q", Version, m.Version())
	}
	view := m.View()
	if !strings.Contains(view, "AIM — AI Multiplexer") {
		t.Errorf("expected 'AIM — AI Multiplexer' in view")
	}
	if !strings.Contains(view, "v"+Version) {
		t.Errorf("expected 'v%s' in view, got:\n%s", Version, view)
	}

	// Custom version without 'v' prefix
	m2 := NewModel(reg, pm, cfg).WithVersion("0.3.1")
	if m2.Version() != "0.3.1" {
		t.Errorf("expected version '0.3.1', got %q", m2.Version())
	}
	view2 := m2.View()
	if !strings.Contains(view2, "AIM — AI Multiplexer") || !strings.Contains(view2, "v0.3.1") {
		t.Errorf("expected 'AIM — AI Multiplexer' and 'v0.3.1' in view, got:\n%s", view2)
	}

	// Custom version with 'v' prefix
	m3 := NewModel(reg, pm, cfg).WithVersion("v1.2.0")
	view3 := m3.View()
	if !strings.Contains(view3, "v1.2.0") {
		t.Errorf("expected 'v1.2.0' in view, got:\n%s", view3)
	}

	// Narrow layout (<70 width)
	mNarrow := NewModel(reg, pm, cfg).WithVersion("0.3.1")
	mNarrow.width = 60
	viewNarrow := mNarrow.View()
	if !strings.Contains(viewNarrow, "AIM — AI Multiplexer") || !strings.Contains(viewNarrow, "v0.3.1") {
		t.Errorf("expected version in narrow view, got:\n%s", viewNarrow)
	}

	// SetVersion pointer method
	m4 := NewModel(reg, pm, cfg)
	m4.SetVersion("2.0.0")
	if m4.Version() != "2.0.0" {
		t.Errorf("expected version '2.0.0', got %q", m4.Version())
	}
	if !strings.Contains(m4.View(), "v2.0.0") {
		t.Errorf("expected 'v2.0.0' in view after SetVersion")
	}
}

func TestTUI_ProfileList_OmitsActiveAgentInBrackets(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("rs")
	_, _ = pm.EnsureProfile("bby")

	cfg := config.NewDefaultConfig()
	cfg.AddProfileAgent("rs", "agy")
	cfg.AddProfileAgent("rs", "codex")
	cfg.AddProfileAgent("bby", "agy")

	reg := agents.DefaultRegistry()

	// Tab: agy
	mAgy := NewModel(reg, pm, cfg)
	mAgy.width = 100
	mAgyView := mAgy.View()

	// bby only has agy, so it should not have brackets
	if strings.Contains(mAgyView, "bby [") {
		t.Errorf("expected 'bby' to have no bracket, got:\n%s", mAgyView)
	}
	// rs is shared by agy and codex. On the agy tab, it should show [codex] and NOT [agy
	if !strings.Contains(mAgyView, "rs [codex]") {
		t.Errorf("expected 'rs [codex]' on agy tab, got:\n%s", mAgyView)
	}
	if strings.Contains(mAgyView, "rs [agy") {
		t.Errorf("did not expect 'agy' in rs brackets on agy tab, got:\n%s", mAgyView)
	}

	// Switch to codex tab
	mCodex, _ := mAgy.switchAgent("codex")
	mCodex.width = 100
	mCodexView := mCodex.View()
	// On codex tab, rs should show [agy] and NOT [codex
	if !strings.Contains(mCodexView, "rs [agy]") {
		t.Errorf("expected 'rs [agy]' on codex tab, got:\n%s", mCodexView)
	}
	if strings.Contains(mCodexView, "rs [codex") {
		t.Errorf("did not expect 'codex' in rs brackets on codex tab, got:\n%s", mCodexView)
	}
}

