package session_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/session"
)

type mockProvider struct {
	agent    string
	sessions []session.Session
}

func (m *mockProvider) Agent() string {
	return m.agent
}

func (m *mockProvider) ListSessions(ctx context.Context, profileDir string, isHost bool) ([]session.Session, error) {
	var filtered []session.Session
	profName := filepath.Base(profileDir)
	for _, s := range m.sessions {
		if isHost && s.IsHost {
			filtered = append(filtered, s)
		} else if !isHost && !s.IsHost && (s.Profile == "" || s.Profile == profName) {
			filtered = append(filtered, s)
		}
	}
	return filtered, nil
}

func (m *mockProvider) GetSession(ctx context.Context, idOrPrefix string, profileDir string, isHost bool) (*session.Session, error) {
	var matches []*session.Session
	profName := filepath.Base(profileDir)
	for _, s := range m.sessions {
		if isHost && !s.IsHost {
			continue
		}
		if !isHost && (s.IsHost || (s.Profile != "" && s.Profile != profName)) {
			continue
		}
		if s.ID == idOrPrefix || strings.HasPrefix(s.ID, idOrPrefix) {
			res := s
			matches = append(matches, &res)
		}
	}
	if len(matches) > 1 {
		var ids []string
		for _, match := range matches {
			ids = append(ids, match.ShortID)
		}
		return nil, fmt.Errorf("ambiguous prefix %q matches multiple sessions: %s", idOrPrefix, strings.Join(ids, ", "))
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	return nil, nil
}

func (m *mockProvider) Hydrate(ctx context.Context, srcSession *session.Session, destProfileDir string, fork bool) (string, error) {
	return srcSession.ID, nil
}

type mockProcessScanner struct {
	active   map[string]session.ActiveProcessInfo
	contexts []context.Context
}

func (s *mockProcessScanner) ScanActiveProcesses(ctx context.Context) (map[string]session.ActiveProcessInfo, error) {
	s.contexts = append(s.contexts, ctx)
	return s.active, nil
}

func TestLatestSessionAndResolveProcessEnrichment(t *testing.T) {
	setupTestProfiles(t)
	scanner := &mockProcessScanner{active: map[string]session.ActiveProcessInfo{
		"session-1": {PID: 42},
	}}
	mgr := session.NewManagerWithScanner(scanner)
	mgr.RegisterProvider(&mockProvider{agent: "agy", sessions: []session.Session{
		session.NewSession("session-1", "Stored session", "agy", "work", false, time.Now()),
	}})
	stored, err := mgr.LatestSession(nil, "agy", "session-1")
	if err != nil {
		t.Fatal(err)
	}
	if stored.Status != session.StatusIdle || len(scanner.contexts) != 0 {
		t.Fatal("stored-session lookup unexpectedly enriched process state")
	}
	active, err := mgr.ResolveSession(nil, "agy", "session-1")
	if err != nil {
		t.Fatal(err)
	}
	if active.Status != session.StatusActive || active.PID != 42 {
		t.Fatalf("expected active process enrichment, got %+v", active)
	}
	if len(scanner.contexts) != 1 || scanner.contexts[0] == nil {
		t.Fatal("expected one process scan with a non-nil context")
	}
}

func setupTestProfiles(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("AIM_HOME", tmpDir)
	for _, prof := range []string{"work", "bby"} {
		p := filepath.Join(tmpDir, "profiles", prof)
		if err := os.MkdirAll(p, 0755); err != nil {
			t.Fatalf("failed to create test profile dir %s: %v", p, err)
		}
	}
	return tmpDir
}

func TestManager_ResolveSession(t *testing.T) {
	setupTestProfiles(t)
	mgr := session.NewManager()

	s1 := session.NewSession("775e6ada-1595-4e7e-84fa-ce0ea71e3007", "Title 1", "agy", "work", false, time.Now())
	s2 := session.NewSession("01a09eb7-2f6c-7c52-895f-218f9ac9eecd", "Title 2", "codex", "<host>", true, time.Now())

	mockAgy := &mockProvider{agent: "agy", sessions: []session.Session{s1}}
	mockCodex := &mockProvider{agent: "codex", sessions: []session.Session{s2}}

	mgr.RegisterProvider(mockAgy)
	mgr.RegisterProvider(mockCodex)

	ctx := context.Background()

	// 1. Resolve exact ID
	res, err := mgr.ResolveSession(ctx, "agy", "775e6ada-1595-4e7e-84fa-ce0ea71e3007")
	if err != nil || res == nil {
		t.Fatalf("expected to resolve s1, got err=%v, res=%v", err, res)
	}

	// 2. Resolve prefix
	resPrefix, err := mgr.ResolveSession(ctx, "agy", "775e6ada")
	if err != nil || resPrefix == nil {
		t.Fatalf("expected to resolve s1 by prefix, got err=%v, res=%v", err, resPrefix)
	}
	if resPrefix.ID != s1.ID {
		t.Errorf("expected resolved ID %s, got %s", s1.ID, resPrefix.ID)
	}

	// 3. Resolve codex session
	resCodex, err := mgr.ResolveSession(ctx, "codex", "01a09eb7")
	if err != nil || resCodex == nil {
		t.Fatalf("expected to resolve s2 by prefix, got err=%v, res=%v", err, resCodex)
	}
	if resCodex.ID != s2.ID {
		t.Errorf("expected resolved ID %s, got %s", s2.ID, resCodex.ID)
	}

	// 4. Resolve non-existent
	notFound, err := mgr.ResolveSession(ctx, "agy", "deadbeef")
	if err == nil || notFound != nil {
		t.Fatalf("expected error for non-existent session, got res=%v", notFound)
	}
}

func TestManager_ActiveProcessCorrelation(t *testing.T) {
	setupTestProfiles(t)
	s1 := session.NewSession("775e6ada-1595-4e7e-84fa-ce0ea71e3007", "Active Conv", "agy", "bby", false, time.Now())
	s2 := session.NewSession("fcdbc2e0-2dc8-4ffa-9ee2-eb5aaa3e556f", "Idle Conv", "agy", "work", false, time.Now().Add(-1*time.Hour))

	mockAgy := &mockProvider{agent: "agy", sessions: []session.Session{s1, s2}}
	scanner := &mockProcessScanner{
		active: map[string]session.ActiveProcessInfo{
			"775e6ada-1595-4e7e-84fa-ce0ea71e3007": {
				PID:            12345,
				Agent:          "agy",
				Profile:        "bby",
				ConversationID: "775e6ada-1595-4e7e-84fa-ce0ea71e3007",
			},
		},
	}

	mgr := session.NewManagerWithScanner(scanner)
	mgr.RegisterProvider(mockAgy)

	ctx := context.Background()
	sessions, err := mgr.ListSessions(ctx, "agy", "", false)
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}

	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}

	var activeFound, idleFound bool
	for _, s := range sessions {
		if s.ID == "775e6ada-1595-4e7e-84fa-ce0ea71e3007" {
			if s.Status != session.StatusActive {
				t.Errorf("expected s1 to be ACTIVE, got %s", s.Status)
			}
			if s.PID != 12345 {
				t.Errorf("expected s1 PID 12345, got %d", s.PID)
			}
			activeFound = true
		}
		if s.ID == "fcdbc2e0-2dc8-4ffa-9ee2-eb5aaa3e556f" {
			if s.Status != session.StatusIdle {
				t.Errorf("expected s2 to be IDLE, got %s", s.Status)
			}
			idleFound = true
		}
	}

	if !activeFound || !idleFound {
		t.Errorf("expected both active and idle sessions to be verified")
	}
}

func TestManager_DeduplicateAndAmbiguity(t *testing.T) {
	setupTestProfiles(t)
	mgr := session.NewManager()
	ctx := context.Background()

	// 1. Profile vs Host deduplication: isolated profile must take priority over host
	sHost := session.NewSession("dup-session-12345678", "Host Copy", "codex", "<host>", true, time.Now().Add(-10*time.Minute))
	sProf := session.NewSession("dup-session-12345678", "Prof Copy", "codex", "work", false, time.Now())

	mockCodex := &mockProvider{
		agent:    "codex",
		sessions: []session.Session{sHost, sProf},
	}
	mgr.RegisterProvider(mockCodex)

	res, err := mgr.ResolveSession(ctx, "codex", "dup-session")
	if err != nil {
		t.Fatalf("expected to resolve session without error, got %v", err)
	}
	if res.IsHost {
		t.Errorf("expected isolated profile to be prioritized over host, got IsHost=true")
	}
	if res.Profile != "work" {
		t.Errorf("expected Profile 'work', got %q", res.Profile)
	}

	// 2. Ambiguous prefix matching multiple distinct sessions
	sA := session.NewSession("ambig-1111-aaaa", "Task A", "codex", "work", false, time.Now())
	sB := session.NewSession("ambig-2222-bbbb", "Task B", "codex", "work", false, time.Now())
	mockCodex.sessions = []session.Session{sA, sB}

	_, errAmbig := mgr.ResolveSession(ctx, "codex", "ambig")
	if errAmbig == nil {
		t.Fatalf("expected error for ambiguous prefix, got nil")
	}
	if !strings.Contains(errAmbig.Error(), "ambiguous prefix") {
		t.Errorf("expected ambiguous prefix error message, got: %v", errAmbig)
	}
}

func TestManager_FindAllSessionsByID(t *testing.T) {
	setupTestProfiles(t)
	mgr := session.NewManager()
	ctx := context.Background()

	sWork := session.NewSession("shared-uuid-1111", "Work Session", "agy", "work", false, time.Now().Add(-1*time.Hour))
	sBby := session.NewSession("shared-uuid-1111", "Bby Session", "agy", "bby", false, time.Now())
	sOther := session.NewSession("different-uuid-2222", "Other Session", "agy", "work", false, time.Now())

	mockAgy := &mockProvider{
		agent:    "agy",
		sessions: []session.Session{sWork, sBby, sOther},
	}
	mgr.RegisterProvider(mockAgy)

	// 1. Matches all instances across profiles without deduplication
	matches, err := mgr.FindAllSessionsByID(ctx, "agy", "shared-uuid-1111")
	if err != nil {
		t.Fatalf("FindAllSessionsByID failed: %v", err)
	}
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches for shared-uuid-1111 across profiles, got %d", len(matches))
	}

	// 2. Prefix matching
	prefixMatches, err := mgr.FindAllSessionsByID(ctx, "agy", "shared-")
	if err != nil {
		t.Fatalf("FindAllSessionsByID prefix failed: %v", err)
	}
	if len(prefixMatches) != 2 {
		t.Fatalf("expected 2 matches for shared- prefix, got %d", len(prefixMatches))
	}

	// 3. Empty ID validation
	_, errEmpty := mgr.FindAllSessionsByID(ctx, "agy", "")
	if errEmpty == nil {
		t.Errorf("expected error for empty session ID, got nil")
	}

	// 4. Ambiguous prefix within a profile surfaces error
	sAmbig1 := session.NewSession("dup-ambig-1111", "Ambig 1", "agy", "work", false, time.Now())
	sAmbig2 := session.NewSession("dup-ambig-2222", "Ambig 2", "agy", "work", false, time.Now())
	mockAgy.sessions = append(mockAgy.sessions, sAmbig1, sAmbig2)

	_, errAmbig := mgr.FindAllSessionsByID(ctx, "agy", "dup-ambig")
	if errAmbig == nil {
		t.Fatalf("expected error for ambiguous prefix, got nil")
	}
	if !strings.Contains(errAmbig.Error(), "ambiguous") {
		t.Errorf("expected error message to contain 'ambiguous', got: %v", errAmbig)
	}

	// 5. Title/name fallback matching
	sNamed := session.NewSession("01a0b351-2f2f-7d22-8176-49e45bde8f9b", "cc-ov2", "agy", "work", false, time.Now())
	mockAgy.sessions = append(mockAgy.sessions, sNamed)

	namedMatches, err := mgr.FindAllSessionsByID(ctx, "agy", "cc-ov2")
	if err != nil {
		t.Fatalf("FindAllSessionsByID by name failed: %v", err)
	}
	if len(namedMatches) != 1 || namedMatches[0].ID != sNamed.ID {
		t.Fatalf("expected 1 match by name cc-ov2, got %+v", namedMatches)
	}

	resolvedNamed, err := mgr.ResolveSession(ctx, "agy", "cc-ov2")
	if err != nil {
		t.Fatalf("ResolveSession by name failed: %v", err)
	}
	if resolvedNamed == nil || resolvedNamed.ID != sNamed.ID {
		t.Fatalf("expected resolved session by name cc-ov2, got %+v", resolvedNamed)
	}
}

func TestManager_ListSessions_CrossProfileLatestActivity(t *testing.T) {
	setupTestProfiles(t)
	mgr := session.NewManagerWithScanner(&mockProcessScanner{})
	ctx := context.Background()

	// Same session ID in work and office, but office has newer timestamp
	tOld := time.Now().Add(-2 * time.Hour)
	tNew := time.Now()

	sWork := session.NewSession("mock-cross-profile-1111", "cc-ov2", "agy", "work", false, tOld)
	sOffice := session.NewSession("mock-cross-profile-1111", "cc-ov2", "agy", "bby", false, tNew)

	mockAgy := &mockProvider{
		agent:    "agy",
		sessions: []session.Session{sWork, sOffice},
	}
	mgr.RegisterProvider(mockAgy)

	sessions, err := mgr.ListSessions(ctx, "agy", "", false)
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}

	// Should deduplicate and keep sOffice
	if len(sessions) != 1 {
		t.Fatalf("expected 1 deduplicated session, got %d", len(sessions))
	}
	if sessions[0].Profile != "bby" {
		t.Errorf("expected session from newer profile 'bby', got %s", sessions[0].Profile)
	}
	if !sessions[0].LastActiveAt.Equal(tNew) {
		t.Errorf("expected newer LastActiveAt, got %v", sessions[0].LastActiveAt)
	}
}
