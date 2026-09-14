package session_test

import (
	"context"
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
	return m.sessions, nil
}

func (m *mockProvider) GetSession(ctx context.Context, idOrPrefix string, profileDir string, isHost bool) (*session.Session, error) {
	for _, s := range m.sessions {
		if s.ID == idOrPrefix || (len(idOrPrefix) >= 4 && len(s.ID) >= len(idOrPrefix) && s.ID[:len(idOrPrefix)] == idOrPrefix) {
			res := s
			return &res, nil
		}
	}
	return nil, nil
}

func (m *mockProvider) Hydrate(ctx context.Context, srcSession *session.Session, destProfileDir string, fork bool) (string, error) {
	return srcSession.ID, nil
}

type mockProcessScanner struct {
	active map[string]session.ActiveProcessInfo
}

func (s *mockProcessScanner) ScanActiveProcesses(ctx context.Context) (map[string]session.ActiveProcessInfo, error) {
	return s.active, nil
}

func TestManager_ResolveSession(t *testing.T) {
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
