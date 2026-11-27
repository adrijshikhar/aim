package tui

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/session"
	"github.com/aim-cli/aim/internal/usage"
	"github.com/charmbracelet/bubbles/textinput"
)

func newFilteredTestModel(t *testing.T, query string) (Model, *doctorRecorder) {
	t.Helper()
	base := t.TempDir()
	t.Setenv("AIM_HOME", base)
	t.Setenv("AIM_REAL_HOME", t.TempDir())
	t.Setenv("AIM_MOCK_KEYCHAIN", "1")
	pm := profile.NewProfileManager(base)
	for _, p := range []string{"alpha", "work"} {
		if _, err := pm.EnsureProfile(p); err != nil {
			t.Fatalf("EnsureProfile(%s): %v", p, err)
		}
	}
	cfg := config.NewDefaultConfig()
	cfg.Profiles = map[string]config.ProfileConfig{
		"alpha": {Agents: []string{"agy"}},
		"work":  {Agents: []string{"agy", "claude"}},
	}
	rec := &doctorRecorder{}
	reg := agents.NewRegistry()
	reg.Register(rec)

	m := NewModel(reg, pm, cfg)
	m.filter.input.SetValue(query)
	m.cursor = 0
	return m, rec
}

func TestFilteredCursorTargetsMatchFilteredListAcrossActions(t *testing.T) {
	m, rec := newFilteredTestModel(t, "work")

	renamed, _ := m.openRenameModal()
	if renamed.RenameModalTarget() != "work" {
		t.Fatalf("openRenameModal target = %q, want %q", renamed.RenameModalTarget(), "work")
	}

	deleted, _ := m.openDeleteModal()
	if deleted.DeleteModalTarget() != "work" {
		t.Fatalf("openDeleteModal target = %q, want %q", deleted.DeleteModalTarget(), "work")
	}

	diagModel := m.fetchDoctorDiagnostics()
	if diagModel.doctorDrawer.targetProfile != "work" {
		t.Fatalf("doctor targetProfile = %q, want %q", diagModel.doctorDrawer.targetProfile, "work")
	}
	if len(rec.dirs) == 0 || !strings.HasSuffix(rec.dirs[0], "/work") && !strings.HasSuffix(rec.dirs[0], `\work`) {
		t.Fatalf("Doctor called with unexpected dirs: %+v", rec.dirs)
	}
}

func TestDeleteModalClearsAllAgentCachesOnFullProfileDelete(t *testing.T) {
	m, _ := newFilteredTestModel(t, "work")
	for _, ag := range []string{"agy", "claude"} {
		if err := m.cache.Put(usage.Report{Agent: ag, Profile: "work", Summary: "ok"}); err != nil {
			t.Fatalf("cache.Put(%s): %v", ag, err)
		}
		m.reports[ag+":work"] = usage.Report{Agent: ag, Profile: "work", Summary: "ok"}
	}

	opened, _ := m.openDeleteModal()
	after, _ := opened.executeDeleteChoice(1) // idx 1 == Delete entire shared profile
	if _, ok := after.cache.GetStale("agy", "work"); ok {
		t.Fatal("expected agy cache entry deleted")
	}
	if _, ok := after.cache.GetStale("claude", "work"); ok {
		t.Fatal("expected claude cache entry deleted")
	}
	if _, ok := after.reports["claude:work"]; ok {
		t.Fatal("expected claude:work report deleted from memory map")
	}
}

func TestMoveModalDeletesSourceCacheEntry(t *testing.T) {
	m, _ := newFilteredTestModel(t, "")
	if err := m.cache.Put(usage.Report{Agent: "agy", Profile: "alpha", Summary: "70%"}); err != nil {
		t.Fatalf("cache.Put: %v", err)
	}
	opened, _ := m.openMoveModal()
	opened.moveModal.input.SetValue("dest")
	after, _ := opened.executeMove()
	if after.MoveModalError() != "" {
		t.Fatalf("executeMove error: %s", after.MoveModalError())
	}
	if _, ok := after.cache.GetStale("agy", "alpha"); ok {
		t.Fatal("expected stale source cache entry deleted after move")
	}
	rep, ok := after.cache.GetStale("agy", "dest")
	if !ok || rep.Profile != "dest" {
		t.Fatalf("expected migrated cache entry for dest, got ok=%v rep=%+v", ok, rep)
	}
}

func TestSessionsDrawer_SetSessionsClampsToFilteredAndSafeUTF8Wrap(t *testing.T) {
	ti := textinput.New()
	ti.SetValue("needle")
	d := SessionsDrawer{filterInput: ti}
	d.cursor = 4
	d.setSessions([]session.Session{
		{ID: "1", Summary: "other"},
		{ID: "2", Summary: "needle match"},
	})
	if d.cursor != 0 {
		t.Fatalf("cursor = %d, want 0 (clamped to filtered len-1)", d.cursor)
	}

	for _, w := range []int{0, 1, 2, 3, 4, 8} {
		if got := truncateString("日本語テストメッセージ", w); !utf8.ValidString(got) || len([]rune(got)) > w {
			t.Fatalf("truncateString(width=%d) = %q (runes=%d)", w, got, len([]rune(got)))
		}
		lines := wrapText("日本語テスト 漢字ワード alpha beta\n日本語テスト", w, 2)
		for _, line := range lines {
			if !utf8.ValidString(line) || len([]rune(line)) > w {
				t.Fatalf("wrapText(width=%d) produced %q (runes=%d)", w, line, len([]rune(line)))
			}
		}
	}

	paraLines := wrapText("line one\nline two\nline three", 40, 2)
	if len(paraLines) != 2 || paraLines[0] != "line one" || paraLines[1] != "line two..." {
		t.Fatalf("expected truncated second paragraph with ellipsis, got %#v", paraLines)
	}
}
