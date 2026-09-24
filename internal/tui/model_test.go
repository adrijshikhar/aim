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
	"github.com/aim-cli/aim/internal/session"
	"github.com/aim-cli/aim/internal/usage"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

	if !strings.Contains(view, "multi [agy, claude]") {
		t.Errorf("expected view to contain multi-agent badge 'multi [agy, claude]', got:\n%s", view)
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
	if !strings.Contains(codexView, "PROFILES:") {
		t.Errorf("expected view to show PROFILES:, got:\n%s", codexView)
	}
	if strings.Contains(codexView, "PROFILES (") {
		t.Errorf("expected view not to contain agent in parentheses in PROFILES: header, got:\n%s", codexView)
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

	// Test ActionShell ('S')
	m = newTestModel(t, "alpha", "beta")
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = newM.(Model)
	newM, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}})
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

	// 'S' on empty profiles does not trigger Shell
	newM, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}})
	m = newM.(Model)
	if m.Outcome() != ActionNone {
		t.Errorf("expected ActionNone on empty profiles 'S', got %v", m.Outcome())
	}
	if cmd != nil {
		t.Errorf("expected nil cmd on empty profiles 'S'")
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

	// Stream reports through the same message path used in production.
	updated, cmd := m.Update(usageReportMsg(repDefault))
	if cmd != nil {
		t.Errorf("expected nil cmd on usageReportMsg, got %v", cmd)
	}
	m = updated.(Model)
	updated, _ = m.Update(usageReportMsg(repProd))
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

	updated, _ := m.Update(usageReportMsg(rep))
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

	updated, _ := m.Update(usageReportMsg(rep))
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

	// View should render spinner next to [r] when loading
	view := m.View()
	if !strings.Contains(view, m.spinner.View()) {
		t.Errorf("expected view to contain spinner, got:\n%s", view)
	}

	// Bubble Tea's spinner tick advances the rendered spinner.
	previousFrame := m.spinner.View()
	updatedSpin, cmdSpin := m.Update(m.spinner.Tick())
	mSpin := updatedSpin.(Model)
	if mSpin.spinner.View() == previousFrame {
		t.Error("expected spinner tick to advance the visible frame")
	}
	if cmdSpin == nil {
		t.Fatalf("expected next spinner tick command from bubbles spinner")
	}
	viewSpin := mSpin.View()
	if !strings.Contains(viewSpin, mSpin.spinner.View()) {
		t.Errorf("expected view to contain updated spinner, got:\n%s", viewSpin)
	}

	// usageStreamClosedMsg should stop loading
	updatedDone, cmdDone := mSpin.Update(usageStreamClosedMsg{})
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

func TestTUI_ProfilesHeader_NoAgentParentheses(t *testing.T) {
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

	// Header should just be "PROFILES:" without "(agy):"
	if !strings.Contains(mAgyView, "PROFILES:") {
		t.Errorf("expected 'PROFILES:' in view, got:\n%s", mAgyView)
	}
	if strings.Contains(mAgyView, "PROFILES (") {
		t.Errorf("expected no agent parentheses in PROFILES: header, got:\n%s", mAgyView)
	}

	// bby only has agy, so it should not have brackets
	if strings.Contains(mAgyView, "bby [") {
		t.Errorf("expected 'bby' to have no bracket, got:\n%s", mAgyView)
	}
	// rs is shared by agy and codex: should show [agy, codex]
	if !strings.Contains(mAgyView, "rs [agy, codex]") {
		t.Errorf("expected 'rs [agy, codex]' on agy tab, got:\n%s", mAgyView)
	}

	// Switch to codex tab
	mCodex, _ := mAgy.switchAgent("codex")
	mCodex.width = 100
	mCodexView := mCodex.View()

	// Header should also be "PROFILES:" without "(codex):"
	if !strings.Contains(mCodexView, "PROFILES:") {
		t.Errorf("expected 'PROFILES:' in codex view, got:\n%s", mCodexView)
	}
	if strings.Contains(mCodexView, "PROFILES (") {
		t.Errorf("expected no agent parentheses in PROFILES: header on codex tab, got:\n%s", mCodexView)
	}

	// On codex tab, rs is still shared across agy and codex, so it shows [agy, codex]
	if !strings.Contains(mCodexView, "rs [agy, codex]") {
		t.Errorf("expected 'rs [agy, codex]' on codex tab, got:\n%s", mCodexView)
	}

	// Agents line is rendered even for single-agent profiles (bby)
	if !strings.Contains(mAgyView, "Agents:") || !strings.Contains(mAgyView, "agy") {
		t.Errorf("expected 'Agents:' line with 'agy' for single-agent profile, got:\n%s", mAgyView)
	}
}

func TestTUI_DecomposedComponents(t *testing.T) {
	// 1. Keys
	km := DefaultKeyMap()
	if len(km.Up.Keys()) == 0 || len(km.Down.Keys()) == 0 || len(km.Quit.Keys()) == 0 {
		t.Fatalf("expected non-empty keybindings in DefaultKeyMap")
	}
	if len(km.ShortHelp()) == 0 || len(km.FullHelp()) == 0 {
		t.Fatalf("expected non-empty ShortHelp and FullHelp in KeyMap")
	}

	m := newTestModel(t, "test-prof")
	if len(m.KeyMap().Quit.Keys()) == 0 {
		t.Errorf("expected m.KeyMap() to return initialized KeyMap")
	}

	// 2. Header and TabBar
	hdr := m.renderHeader()
	if !strings.Contains(hdr, "AIM — AI Multiplexer") {
		t.Errorf("expected header to contain 'AIM — AI Multiplexer', got:\n%s", hdr)
	}
	tabBar := m.renderTabBar()
	if !strings.Contains(tabBar, "Antigravity (agy)") {
		t.Errorf("expected tab bar to contain 'Antigravity (agy)', got:\n%s", tabBar)
	}
	if tag := m.formatVersionTag(); tag == "" || !strings.HasPrefix(tag, "v") {
		t.Errorf("expected version tag starting with 'v', got %q", tag)
	}

	// 3. Inspector and formatWin
	w := usage.LimitWindow{
		Name:         "test-win",
		RemainingPct: 80,
		ResetsIn:     time.Hour,
	}
	winFormatted := formatWin(w, "5h")
	if !strings.Contains(winFormatted, "80%") || !strings.Contains(winFormatted, "5h:") {
		t.Errorf("expected formatWin to contain '80%%' and '5h:', got %q", winFormatted)
	}

	insp := m.renderInspector("test-prof")
	if !strings.Contains(insp, "Profile Details: test-prof") {
		t.Errorf("expected inspector to contain 'Profile Details: test-prof', got:\n%s", insp)
	}
	if empty := m.renderInspector(""); empty != "" {
		t.Errorf("expected empty string for empty profile, got %q", empty)
	}
}

func TestTUI_DecomposedModalsAndDrawers(t *testing.T) {
	m := newTestModel(t, "test-prof")

	// 1. Delete Modal decomposition
	mDel, _ := m.openDeleteModal()
	if !mDel.IsDeleteModalActive() {
		t.Fatalf("expected delete modal to be active after openDeleteModal()")
	}
	delView := mDel.renderDeleteModal()
	if !strings.Contains(delView, "Confirm Deletion: test-prof") {
		t.Errorf("expected delete modal view to contain 'Confirm Deletion: test-prof', got:\n%s", delView)
	}
	mDelClosed, _ := mDel.updateDeleteModal(tea.KeyMsg{Type: tea.KeyEsc})
	if mDelClosed.IsDeleteModalActive() {
		t.Errorf("expected delete modal to be closed after updateDeleteModal(Esc)")
	}

	// 2. Rename Modal decomposition
	mRen, _ := m.openRenameModal()
	if !mRen.IsRenameModalActive() {
		t.Fatalf("expected rename modal to be active after openRenameModal()")
	}
	renView := mRen.renderRenameModal()
	if !strings.Contains(renView, "Rename Profile: test-prof") {
		t.Errorf("expected rename modal view to contain 'Rename Profile: test-prof', got:\n%s", renView)
	}
	mRenClosed, _ := mRen.updateRenameModal(tea.KeyMsg{Type: tea.KeyEsc})
	if mRenClosed.IsRenameModalActive() {
		t.Errorf("expected rename modal to be closed after updateRenameModal(Esc)")
	}

	// 3. Doctor Drawer decomposition
	mDoc, _ := m.openDoctorDrawer()
	if !mDoc.IsDoctorDrawerActive() {
		t.Fatalf("expected doctor drawer to be active after openDoctorDrawer()")
	}
	docView := mDoc.renderDoctorDrawer()
	if !strings.Contains(docView, "Diagnostics") {
		t.Errorf("expected doctor drawer view to contain 'Diagnostics', got:\n%s", docView)
	}
	mDocClosed, _ := mDoc.updateDoctorDrawer(tea.KeyMsg{Type: tea.KeyEsc})
	if mDocClosed.IsDoctorDrawerActive() {
		t.Errorf("expected doctor drawer to be closed after updateDoctorDrawer(Esc)")
	}

	// 4. Help Overlay decomposition
	mHelp, _ := m.openHelpOverlay()
	if !mHelp.IsHelpActive() {
		t.Fatalf("expected help overlay to be active after openHelpOverlay()")
	}
	helpView := mHelp.renderHelpOverlay()
	if !strings.Contains(helpView, "Keyboard Shortcuts") {
		t.Errorf("expected help overlay view to contain 'Keyboard Shortcuts', got:\n%s", helpView)
	}
	mHelpClosed, _ := mHelp.updateHelpOverlay(tea.KeyMsg{Type: tea.KeyEsc})
	if mHelpClosed.IsHelpActive() {
		t.Errorf("expected help overlay to be closed after updateHelpOverlay(Esc)")
	}

	// 5. Filter bar decomposition
	mFilter, cmdFilter := m.openFilter()
	if !mFilter.IsFilterActive() {
		t.Fatalf("expected filter to be active after openFilter()")
	}
	if cmdFilter == nil {
		t.Errorf("expected non-nil focus cmd on openFilter()")
	}
	filterBarView := mFilter.renderFilterBar()
	if !strings.Contains(filterBarView, "FILTER:") || !strings.Contains(filterBarView, "Enter to apply") {
		t.Errorf("expected filter bar to contain prompt and hints, got:\n%s", filterBarView)
	}
	mFilterClosed, _ := mFilter.updateFilter(tea.KeyMsg{Type: tea.KeyEsc})
	if mFilterClosed.IsFilterActive() {
		t.Errorf("expected filter to be closed after updateFilter(Esc)")
	}
}

func TestTUI_HelpOverlayToggle(t *testing.T) {
	m := newTestModel(t, "test-prof")

	// Initially help overlay is inactive
	if m.IsHelpActive() {
		t.Fatalf("expected help overlay to be initially inactive")
	}

	// Normal view should have [?] Help hint
	normalView := m.View()
	if !strings.Contains(normalView, "[?]") || !strings.Contains(normalView, "Help") {
		t.Errorf("expected footer to contain [?] Help hint, got:\n%s", normalView)
	}

	// Press '?' to toggle open help overlay
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	mHelp := updated.(Model)
	if !mHelp.IsHelpActive() {
		t.Fatalf("expected help overlay to be active after pressing '?'")
	}

	// Help overlay view contains cheatsheet content
	view := mHelp.View()
	if !strings.Contains(view, "Keyboard Shortcuts") {
		t.Errorf("expected help overlay view to contain 'Keyboard Shortcuts', got:\n%s", view)
	}
	if !strings.Contains(view, "Navigation & Tabs") || !strings.Contains(view, "Actions & Commands") {
		t.Errorf("expected help overlay view to contain cheatsheet categories, got:\n%s", view)
	}

	// Press 'Esc' to close help overlay
	closed, _ := mHelp.Update(tea.KeyMsg{Type: tea.KeyEsc})
	mClosed := closed.(Model)
	if mClosed.IsHelpActive() {
		t.Errorf("expected help overlay to close after pressing Esc")
	}

	// Press '?' to open again, then '?' to toggle close
	reopened, _ := mClosed.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	mReopened := reopened.(Model)
	if !mReopened.IsHelpActive() {
		t.Fatalf("expected help overlay to be active after second '?'")
	}
	reclosed, _ := mReopened.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	mReclosed := reclosed.(Model)
	if mReclosed.IsHelpActive() {
		t.Errorf("expected help overlay to toggle closed after pressing '?' again")
	}

	// Press '?' to open again, then 'q' to close
	openForQ, _ := mReclosed.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	mOpenForQ := openForQ.(Model)
	if !mOpenForQ.IsHelpActive() {
		t.Fatalf("expected help overlay to be active before testing 'q' close")
	}
	closedByQ, _ := mOpenForQ.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	mClosedByQ := closedByQ.(Model)
	if mClosedByQ.IsHelpActive() {
		t.Errorf("expected help overlay to close after pressing 'q'")
	}

	// Non-closing keys should not close help overlay or trigger background actions
	openAgain, _ := mClosedByQ.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	mOpenAgain := openAgain.(Model)
	unhandled, cmd := mOpenAgain.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	mUnhandled := unhandled.(Model)
	if !mUnhandled.IsHelpActive() {
		t.Errorf("expected help overlay to remain active on unhandled key")
	}
	if cmd != nil {
		t.Errorf("expected nil cmd for unhandled key in help overlay")
	}

	// Narrow view layout check (< 70 width)
	mNarrow := mOpenAgain
	mNarrow.width = 60
	narrowView := mNarrow.View()
	if !strings.Contains(narrowView, "Keyboard Shortcuts") {
		t.Errorf("expected narrow help view to contain 'Keyboard Shortcuts', got:\n%s", narrowView)
	}
}

func TestTUI_FilterProfiles(t *testing.T) {
	// Baseline matching test specified in plan doc
	{
		m := NewModel(nil, nil, nil)
		m.profiles = []string{"personal", "work-backend", "work-frontend"}
		m.filter.active = true
		m.filter.input.SetValue("backend")

		filtered := m.filteredProfiles()
		if len(filtered) != 1 || filtered[0] != "work-backend" {
			t.Fatalf("expected only 'work-backend', got %v", filtered)
		}
	}

	baseDir := t.TempDir()
	pm := profile.NewProfileManager(baseDir)
	cfg := config.NewDefaultConfig()

	// Setup profiles
	profiles := []string{"personal", "work-backend", "work-frontend", "prod-cluster"}
	for _, p := range profiles {
		_, _ = pm.EnsureProfile(p)
	}
	cfg.AddProfileAgent("personal", "agy")
	cfg.AddProfileAgent("work-backend", "agy")
	cfg.AddProfileAgent("work-backend", "codex")
	cfg.AddProfileAgent("work-frontend", "agy")
	cfg.AddProfileAgent("work-frontend", "claude")
	cfg.AddProfileAgent("prod-cluster", "agy")
	cfg.AddProfileAgent("prod-cluster", "gemini")

	m := NewModel(nil, pm, cfg)
	if len(m.Profiles()) != 4 {
		t.Fatalf("expected 4 profiles, got %d", len(m.Profiles()))
	}

	// Initially filter is inactive, filteredProfiles returns all 4
	if m.IsFilterActive() {
		t.Fatalf("expected filter to be initially inactive")
	}
	if len(m.filteredProfiles()) != 4 {
		t.Fatalf("expected 4 filtered profiles initially, got %d", len(m.filteredProfiles()))
	}

	// 1. Filtering by profile name
	m.filter.input.SetValue("backend")
	filtered := m.filteredProfiles()
	if len(filtered) != 1 || filtered[0] != "work-backend" {
		t.Fatalf("expected only 'work-backend' for query 'backend', got %v", filtered)
	}

	// 2. Filtering by attached agent name
	m.filter.input.SetValue("claude")
	filtered = m.filteredProfiles()
	if len(filtered) != 1 || filtered[0] != "work-frontend" {
		t.Fatalf("expected only 'work-frontend' for query 'claude', got %v", filtered)
	}

	m.filter.input.SetValue("gemini")
	filtered = m.filteredProfiles()
	if len(filtered) != 1 || filtered[0] != "prod-cluster" {
		t.Fatalf("expected only 'prod-cluster' for query 'gemini', got %v", filtered)
	}

	// 3. Case-insensitivity (both name and agent)
	m.filter.input.SetValue("BACKEND")
	filtered = m.filteredProfiles()
	if len(filtered) != 1 || filtered[0] != "work-backend" {
		t.Fatalf("expected 'work-backend' for query 'BACKEND', got %v", filtered)
	}

	m.filter.input.SetValue("CLAUDE")
	filtered = m.filteredProfiles()
	if len(filtered) != 1 || filtered[0] != "work-frontend" {
		t.Fatalf("expected 'work-frontend' for query 'CLAUDE', got %v", filtered)
	}

	// Substring matching multiple profiles
	m.filter.input.SetValue("work")
	filtered = m.filteredProfiles()
	if len(filtered) != 2 || filtered[0] != "work-backend" || filtered[1] != "work-frontend" {
		t.Fatalf("expected ['work-backend', 'work-frontend'] for query 'work', got %v", filtered)
	}

	// Non-matching query
	m.filter.input.SetValue("nonexistent")
	filtered = m.filteredProfiles()
	if len(filtered) != 0 {
		t.Fatalf("expected 0 filtered profiles for query 'nonexistent', got %v", filtered)
	}
	viewNonExistent := m.View()
	if !strings.Contains(viewNonExistent, "(no profiles matching \"nonexistent\")") {
		t.Errorf("expected view to show no profiles matching message, got:\n%s", viewNonExistent)
	}

	// 4. Interactive flow: '/' shortcut, typing, Enter lock-in, and list navigation
	// Reset filter
	m.filter = newFilterState()

	// Press '/' to activate filter
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(Model)
	if !m.IsFilterActive() {
		t.Fatalf("expected filter to become active after pressing '/'")
	}
	if cmd == nil {
		t.Errorf("expected focus command when activating filter")
	}

	// Check filter bar rendering while active
	bar := m.renderFilterBar()
	if !strings.Contains(bar, "FILTER:") || !strings.Contains(bar, "(Enter to apply, Esc to clear)") {
		t.Errorf("expected filter bar to contain prompt and hints, got:\n%s", bar)
	}

	// Type 'w', 'o', 'r', 'k'
	for _, r := range []rune{'w', 'o', 'r', 'k'} {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(Model)
	}
	if m.FilterInputValue() != "work" {
		t.Fatalf("expected input value 'work', got %q", m.FilterInputValue())
	}
	if len(m.filteredProfiles()) != 2 {
		t.Fatalf("expected 2 filtered profiles while typing, got %d", len(m.filteredProfiles()))
	}

	// View should contain filter bar and only filtered profiles
	view := m.View()
	if !strings.Contains(view, "FILTER:") || !strings.Contains(view, "/work") {
		t.Errorf("expected view to contain filter bar with '/work', got:\n%s", view)
	}
	if !strings.Contains(view, "work-backend") || !strings.Contains(view, "work-frontend") {
		t.Errorf("expected view to contain filtered profiles, got:\n%s", view)
	}
	if strings.Contains(view, "personal") || strings.Contains(view, "prod-cluster") {
		t.Errorf("expected view not to contain non-matching profiles, got:\n%s", view)
	}

	// Press Enter to lock in the filter
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if m.IsFilterActive() {
		t.Fatalf("expected filter to be locked in (inactive) after pressing Enter")
	}
	if m.FilterInputValue() != "work" {
		t.Fatalf("expected filter query to remain 'work' after lock-in, got %q", m.FilterInputValue())
	}
	if len(m.filteredProfiles()) != 2 {
		t.Fatalf("expected 2 filtered profiles after lock-in, got %d", len(m.filteredProfiles()))
	}

	// Navigate within filtered list
	if m.cursor != 0 {
		t.Fatalf("expected cursor 0, got %d", m.cursor)
	}
	// Down to index 1 ("work-frontend")
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(Model)
	if m.cursor != 1 {
		t.Fatalf("expected cursor 1 after Down, got %d", m.cursor)
	}
	// Down again should be bounded at index 1 (not go to 2 or 3)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(Model)
	if m.cursor != 1 {
		t.Fatalf("expected cursor to remain 1 at bottom of filtered list, got %d", m.cursor)
	}
	// Up back to index 0
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(Model)
	if m.cursor != 0 {
		t.Fatalf("expected cursor 0 after Up, got %d", m.cursor)
	}

	// Press Enter while filter is locked in -> triggers ActionRun on selected profile ("work-backend")
	updated, runCmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mRun := updated.(Model)
	if mRun.Outcome() != ActionRun {
		t.Fatalf("expected ActionRun on Enter when filter is locked in, got %v", mRun.Outcome())
	}
	if mRun.SelectedProfile() != "work-backend" {
		t.Fatalf("expected selected profile 'work-backend', got %q", mRun.SelectedProfile())
	}
	if runCmd == nil {
		t.Errorf("expected tea.Quit command on ActionRun")
	}

	// 5. Esc clears locked-in filter
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.FilterInputValue() != "" {
		t.Fatalf("expected filter text to be cleared on Esc, got %q", m.FilterInputValue())
	}
	if len(m.filteredProfiles()) != 4 {
		t.Fatalf("expected all 4 profiles after clearing filter, got %d", len(m.filteredProfiles()))
	}
	if strings.Contains(m.View(), "FILTER:") {
		t.Errorf("expected filter bar to disappear when filter is empty and inactive")
	}

	// Esc clear while active
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(Model)
	if !m.IsFilterActive() {
		t.Fatalf("expected filter active")
	}
	for _, r := range []rune{'t', 'e', 's', 't'} {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(Model)
	}
	if m.FilterInputValue() != "test" {
		t.Fatalf("expected 'test', got %q", m.FilterInputValue())
	}
	// Press Esc while active -> clears filter and closes filter mode
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.IsFilterActive() {
		t.Fatalf("expected filter inactive after Esc")
	}
	if m.FilterInputValue() != "" {
		t.Fatalf("expected filter text empty after Esc, got %q", m.FilterInputValue())
	}
	if len(m.filteredProfiles()) != 4 {
		t.Fatalf("expected 4 profiles, got %d", len(m.filteredProfiles()))
	}
}

func TestSessionsDrawer_OpenAndClose(t *testing.T) {
	m := newTestModel(t, "alpha", "beta")

	// Press 's' to open sessions drawer
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = updated.(Model)

	if !m.IsSessionsDrawerActive() {
		t.Fatalf("expected sessions drawer to be active after pressing 's'")
	}

	view := m.View()
	if !strings.Contains(view, "Sessions Explorer") {
		t.Errorf("expected view to contain 'Sessions Explorer', got:\n%s", view)
	}

	// Press Esc to close
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)

	if m.IsSessionsDrawerActive() {
		t.Fatalf("expected sessions drawer to be closed after pressing Esc")
	}
}

func TestSessionsDrawer_Interactions(t *testing.T) {
	m := newTestModel(t, "alpha", "beta")

	// Open sessions drawer
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = updated.(Model)

	// Inject mock sessions
	s1 := session.NewSession("11111111-2222-3333-4444-555566667777", "Session One", "agy", "alpha", false, time.Now())
	s2 := session.NewSession("88888888-9999-aaaa-bbbb-ccccddddeeee", "Session Two", "agy", "beta", false, time.Now())
	m.SetSessionsForTest([]session.Session{s1, s2})

	// Test navigation: Down (j)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)

	// Press Enter on second session to open resume modal
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	if !m.IsResumeModalActive() {
		t.Fatalf("expected ResumeModal to be active")
	}
	if m.ResumeModalSession() == nil || m.ResumeModalSession().ID != s2.ID {
		t.Fatalf("expected resume modal session %s, got %v", s2.ID, m.ResumeModalSession())
	}

	// Press Enter again inside resume modal to confirm pre-selected original profile
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	if m.Outcome() != ActionResumeExact {
		t.Fatalf("expected ActionResumeExact, got %v", m.Outcome())
	}
	if m.SelectedSession() == nil || m.SelectedSession().ID != s2.ID {
		t.Fatalf("expected selected session %s, got %v", s2.ID, m.SelectedSession())
	}
	if m.SelectedProfile() != "beta" {
		t.Fatalf("expected selected profile to be beta, got %s", m.SelectedProfile())
	}
	if cmd == nil {
		t.Errorf("expected tea.Quit command on resume")
	}

	// Test Catalyst resume ('c')
	m = newTestModel(t, "alpha", "beta")
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = updated.(Model)
	m.SetSessionsForTest([]session.Session{s1, s2})

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	m = updated.(Model)

	if !m.IsResumeModalActive() {
		t.Fatalf("expected ResumeModal to be active for catalyst resume")
	}

	// Confirm in modal
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	if m.Outcome() != ActionResumeCatalyst {
		t.Fatalf("expected ActionResumeCatalyst, got %v", m.Outcome())
	}
	if m.SelectedSession() == nil || m.SelectedSession().ID != s1.ID {
		t.Fatalf("expected selected session %s, got %v", s1.ID, m.SelectedSession())
	}
	if m.SelectedProfile() != "alpha" {
		t.Fatalf("expected selected profile alpha, got %s", m.SelectedProfile())
	}

	// Test Fork resume ('b')
	m = newTestModel(t, "alpha", "beta")
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = updated.(Model)
	m.SetSessionsForTest([]session.Session{s1, s2})

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	m = updated.(Model)

	if !m.IsResumeModalActive() {
		t.Fatalf("expected ResumeModal to be active for fork resume")
	}

	// Confirm in modal
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	if m.Outcome() != ActionResumeExact {
		t.Fatalf("expected ActionResumeExact for fork, got %v", m.Outcome())
	}
	if !m.IsForkResume() {
		t.Fatalf("expected IsForkResume to be true")
	}
}

func TestResumeModal_ProfileNavigationAndCustom(t *testing.T) {
	m := newTestModel(t, "alpha", "beta", "gamma")

	// Open sessions drawer
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = updated.(Model)

	s1 := session.NewSession("11111111-2222-3333-4444-555566667777", "Session One", "agy", "beta", false, time.Now())
	m.SetSessionsForTest([]session.Session{s1})

	// Open modal via Enter
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	if !m.IsResumeModalActive() {
		t.Fatalf("expected resume modal active")
	}
	// Original profile "beta" should be first
	profiles := m.ResumeModalProfiles()
	if len(profiles) == 0 || profiles[0] != "beta" {
		t.Fatalf("expected first profile to be original 'beta', got %v", profiles)
	}

	// Navigate down to second profile
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)
	if m.ResumeModalCursor() != 1 {
		t.Fatalf("expected cursor 1, got %d", m.ResumeModalCursor())
	}

	// Confirm with second profile
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if cmd == nil {
		t.Fatalf("expected quit command")
	}
	if m.SelectedProfile() != profiles[1] {
		t.Fatalf("expected selected profile %s, got %s", profiles[1], m.SelectedProfile())
	}

	// Test Esc dismisses modal back to drawer
	m = newTestModel(t, "alpha", "beta")
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = updated.(Model)
	m.SetSessionsForTest([]session.Session{s1})
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if !m.IsResumeModalActive() {
		t.Fatalf("expected resume modal active")
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.IsResumeModalActive() {
		t.Fatalf("expected resume modal dismissed")
	}
	if !m.sessionsDrawer.active {
		t.Fatalf("expected sessions drawer still active")
	}

	// Test Custom Profile input
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	// Cursor to bottom ("Enter custom profile name...")
	totalOpts := len(m.ResumeModalProfiles()) + 1
	for i := 0; i < totalOpts; i++ {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		m = updated.(Model)
	}
	// Enter custom mode
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if !m.ResumeModalIsCustomMode() {
		t.Fatalf("expected customMode true")
	}

	// Type custom profile name: "work-profile"
	for _, r := range "work-profile" {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(Model)
	}
	// Confirm custom profile
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if cmd == nil {
		t.Fatalf("expected quit command")
	}
	if m.SelectedProfile() != "work-profile" {
		t.Fatalf("expected selected 'work-profile', got '%s'", m.SelectedProfile())
	}
}

func TestResumeModal_QuotaAndActiveBadges(t *testing.T) {
	m := newTestModel(t, "alpha", "beta")
	m.agent = "agy"

	// Inject usage reports
	m.reports["agy:alpha"] = usage.Report{
		Agent:   "agy",
		Profile: "alpha",
		Status:  usage.StatusOK,
		Windows: []usage.LimitWindow{
			{Name: "5-hour limit", RemainingPct: 85},
		},
	}
	m.reports["agy:beta"] = usage.Report{
		Agent:   "agy",
		Profile: "beta",
		Status:  usage.StatusCritical,
		Windows: []usage.LimitWindow{
			{Name: "5-hour limit", RemainingPct: 10},
		},
	}

	// Open sessions drawer
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = updated.(Model)

	// Inject sessions: s1 idle in alpha, s2 active in beta
	s1 := session.NewSession("11111111-2222-3333-4444-555566667777", "Session One", "agy", "alpha", false, time.Now())
	s2 := session.NewSession("88888888-9999-aaaa-bbbb-ccccddddeeee", "Session Two", "agy", "beta", false, time.Now())
	s2.Status = session.StatusActive
	m.SetSessionsForTest([]session.Session{s1, s2})

	// Open resume modal on s1 (original: alpha)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	if !m.IsResumeModalActive() {
		t.Fatalf("expected resume modal active")
	}

	view := m.renderResumeModal()
	if !strings.Contains(view, "[85%]") {
		t.Errorf("expected view to contain [85%%], got:\n%s", view)
	}
	if !strings.Contains(view, "[10%]") {
		t.Errorf("expected view to contain [10%%], got:\n%s", view)
	}
	if !strings.Contains(view, "(1 active)") {
		t.Errorf("expected view to contain '(1 active)', got:\n%s", view)
	}

	// Move cursor down to beta (index 1)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)

	// Verify matching width and border integrity:
	modalView := m.renderResumeModal()
	modalWidth := lipgloss.Width(modalView)
	drawerView := m.renderSessionsDrawer()
	drawerWidth := lipgloss.Width(drawerView)

	if modalWidth != drawerWidth {
		t.Errorf("expected modal width (%d) to equal drawer width (%d)", modalWidth, drawerWidth)
	}

	lines := strings.Split(modalView, "\n")
	for i, l := range lines {
		if strings.TrimSpace(l) == "" {
			continue
		}
		w := lipgloss.Width(l)
		if w != modalWidth {
			t.Errorf("modal line %d width %d does not match expected box width %d: %q", i, w, modalWidth, l)
		}
	}

	drawerLines := strings.Split(drawerView, "\n")
	for i, l := range drawerLines {
		if strings.TrimSpace(l) == "" {
			continue
		}
		w := lipgloss.Width(l)
		if w != drawerWidth {
			t.Errorf("drawer line %d width %d does not match expected box width %d: %q", i, w, drawerWidth, l)
		}
	}
}

func TestSessionsDrawer_WrapText(t *testing.T) {
	// Test 1: Empty text
	if lines := wrapText("", 80, 3); len(lines) != 0 {
		t.Errorf("expected 0 lines for empty text, got %d", len(lines))
	}

	// Test 2: Very long unbroken word (e.g. file path) must be clamped to maxWidth
	longWord := "/Users/nemesis/Projects/my-projects/aim/.superpowers/sdd/2026-09-14-k9s-architecture-and-code-health/task-6-brief.md"
	wrapped := wrapText(longWord, 40, 3)
	for idx, l := range wrapped {
		if lipgloss.Width(l) > 40 {
			t.Errorf("line %d width %d exceeded maxWidth 40: %s", idx, lipgloss.Width(l), l)
		}
	}

	// Test 2: Text that wraps across 3 lines
	longPrompt := "Please perform a thorough, comprehensive code review of the pull request changes on branch 'feat/sessions-and-cross-profile-resume' in /Users/nemesis/Projects/my-projects/aim against main."
	lines := wrapText(longPrompt, 80, 3)
	if len(lines) != 3 {
		t.Fatalf("expected 3 wrapped lines, got %d: %+v", len(lines), lines)
	}

	if !strings.Contains(lines[0], "Please perform") {
		t.Errorf("expected line 1 to start with 'Please perform', got: %s", lines[0])
	}
	if !strings.Contains(lines[2], "against main.") {
		t.Errorf("expected line 3 to contain 'against main.', got: %s", lines[2])
	}

	// Test 3: Very long text exceeding 3 lines should end with ...
	veryLong := "Word " + strings.Repeat("test ", 100)
	linesLong := wrapText(veryLong, 50, 3)
	if len(linesLong) != 3 {
		t.Fatalf("expected 3 lines for very long text, got %d", len(linesLong))
	}
	if !strings.HasSuffix(linesLong[2], "...") {
		t.Errorf("expected line 3 to end with '...', got: %s", linesLong[2])
	}
}
