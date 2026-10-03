package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/session"
)

func TestNewIndexedSession(t *testing.T) {
	s := session.Session{
		ID:       "12345678-ABCD-EF01-2345-6789ABCDEF01",
		ShortID:  "12345678",
		Title:    "Fixing Memory Leak in TUI",
		Summary:  "Investigating RSS growth",
		Goal:     "Stabilize memory footprint",
		Progress: "Profiling pprof heap",
		Recent:   "Found large slice allocations",
		Cwd:      "/Users/dev/Projects/aim",
		Profile:  "Work-Profile",
		Agent:    "Claude",
	}

	idx := NewIndexedSession(s)

	if idx.Session.ID != s.ID {
		t.Fatalf("expected embedded Session.ID to match %q, got %q", s.ID, idx.Session.ID)
	}

	for _, field := range []string{
		"12345678-abcd-ef01-2345-6789abcdef01",
		"12345678",
		"fixing memory leak in tui",
		"investigating rss growth",
		"stabilize memory footprint",
		"profiling pprof heap",
		"found large slice allocations",
		"/users/dev/projects/aim",
		"work-profile",
		"claude",
	} {
		if !strings.Contains(idx.searchCorpus, field) {
			t.Errorf("expected searchCorpus to contain %q, but got %q", field, idx.searchCorpus)
		}
	}
}

func TestSessionsIndex_Search(t *testing.T) {
	s1 := session.Session{
		ID:      "1111",
		ShortID: "1111",
		Title:   "Frontend React Redesign",
		Profile: "personal",
		Agent:   "agy",
		Status:  session.StatusIdle,
		Cwd:     "/code/frontend",
	}
	s2 := session.Session{
		ID:      "2222",
		ShortID: "2222",
		Title:   "Backend Go Microservice",
		Profile: "work",
		Agent:   "codex",
		Status:  session.StatusActive,
		Cwd:     "/code/backend",
	}
	s3 := session.Session{
		ID:      "3333",
		ShortID: "3333",
		Title:   "CLI Architecture & Scalability",
		Profile: "work",
		Agent:   "claude",
		Status:  session.StatusActive,
		Cwd:     "/code/aim",
	}

	index := NewSessionsIndex([]session.Session{s1, s2, s3})

	// 1. Empty query returns all sessions
	all := index.Search("", false)
	if len(all) != 3 {
		t.Fatalf("expected 3 sessions, got %d", len(all))
	}

	// 2. Active only
	active := index.Search("", true)
	if len(active) != 2 {
		t.Fatalf("expected 2 active sessions, got %d", len(active))
	}
	for _, s := range active {
		if s.Status != session.StatusActive {
			t.Errorf("expected session to be active, got %s", s.Status)
		}
	}

	// 3. Query match across fields (case-insensitive)
	reactMatches := index.Search("REACT", false)
	if len(reactMatches) != 1 || reactMatches[0].ID != "1111" {
		t.Fatalf("expected 1 match for REACT (1111), got %+v", reactMatches)
	}

	goMatches := index.Search("backend", false)
	if len(goMatches) != 1 || goMatches[0].ID != "2222" {
		t.Fatalf("expected 1 match for backend, got %+v", goMatches)
	}

	claudeMatches := index.Search("claude", false)
	if len(claudeMatches) != 1 || claudeMatches[0].ID != "3333" {
		t.Fatalf("expected 1 match for claude, got %+v", claudeMatches)
	}

	// 4. SearchWithProfile
	workMatches := index.SearchWithProfile("", false, "work")
	if len(workMatches) != 2 {
		t.Fatalf("expected 2 sessions for profile 'work', got %d", len(workMatches))
	}

	workActiveClaude := index.SearchWithProfile("scalability", true, "work")
	if len(workActiveClaude) != 1 || workActiveClaude[0].ID != "3333" {
		t.Fatalf("expected 1 match for scalability in work active, got %+v", workActiveClaude)
	}

	// 5. No matches
	noMatches := index.Search("nonexistent-term-xyz", false)
	if len(noMatches) != 0 {
		t.Fatalf("expected 0 matches, got %d", len(noMatches))
	}
}

func TestSessionsIndex_NilAndEmptySafe(t *testing.T) {
	var nilIndex *SessionsIndex
	if res := nilIndex.Search("test", false); res != nil {
		t.Fatalf("expected nil for nil index search, got %v", res)
	}

	emptyIndex := NewSessionsIndex(nil)
	if res := emptyIndex.Search("test", false); len(res) != 0 {
		t.Fatalf("expected empty slice for empty index search, got %v", res)
	}
}

func BenchmarkSessionsIndex_Search(b *testing.B) {
	// Create 1,000 synthetic sessions
	sessions := make([]session.Session, 1000)
	for i := 0; i < 1000; i++ {
		s := session.Session{
			ID:        fmt.Sprintf("session-%04d-uuid-000000000000", i),
			ShortID:   fmt.Sprintf("%04d", i),
			Title:     fmt.Sprintf("Feature Implementation Task #%d", i),
			Summary:   fmt.Sprintf("Building component %d for subsystem", i%10),
			Goal:      fmt.Sprintf("Deliver module #%d on time", i),
			Progress:  fmt.Sprintf("Completed %d percent", i%100),
			Recent:    fmt.Sprintf("Committed changes to branch-%d", i),
			Cwd:       fmt.Sprintf("/Users/workspace/project-%d", i%5),
			Profile:   fmt.Sprintf("profile-%d", i%8),
			Agent:     "agy",
			StartedAt: time.Now(),
		}
		if i%3 == 0 {
			s.Status = session.StatusActive
		} else {
			s.Status = session.StatusIdle
		}
		sessions[i] = s
	}

	index := NewSessionsIndex(sessions)
	b.ResetTimer()

	b.Run("MatchQuarter", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = index.Search("subsystem", false)
		}
	})

	b.Run("MatchSingle", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = index.Search("Task #42", false)
		}
	})
}
