package claude

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

var uuidV4Regex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestProvider_Metadata(t *testing.T) {
	p := NewProvider()
	if p.Agent() != "claude" {
		t.Fatalf("expected agent 'claude', got %q", p.Agent())
	}
}

func TestProvider_ListSessions(t *testing.T) {
	tempDir := t.TempDir()
	profileDir := filepath.Join(tempDir, "profiles", "work")
	projectSlug := "-Users-nemesis-Projects-aim"
	projDir := filepath.Join(profileDir, ".claude", "projects", projectSlug)
	if err := os.MkdirAll(projDir, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}

	sessionUUID := "d24699d0-820f-435f-b4e4-eaf8440311b4"
	jsonlContent := strings.Join([]string{
		`{"type":"last-prompt","leafUuid":"leaf-1","sessionId":"` + sessionUUID + `"}`,
		`{"type":"user","message":{"role":"user","content":"Optimize database query"},"timestamp":"2026-09-23T12:00:00.000Z","sessionId":"` + sessionUUID + `"}`,
		`{"type":"assistant","message":{"role":"assistant","content":"I will analyze the query."},"timestamp":"2026-09-23T12:01:00.000Z","sessionId":"` + sessionUUID + `"}`,
	}, "\n") + "\n"

	sessionFile := filepath.Join(projDir, sessionUUID+".jsonl")
	if err := os.WriteFile(sessionFile, []byte(jsonlContent), 0644); err != nil {
		t.Fatalf("failed to write session file: %v", err)
	}

	// Also write a non-jsonl file to ensure it is ignored
	_ = os.WriteFile(filepath.Join(projDir, "notes.txt"), []byte("random text"), 0644)

	p := NewProvider()
	ctx := context.Background()

	sessions, err := p.ListSessions(ctx, profileDir, false)
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}

	s := sessions[0]
	if s.ID != sessionUUID {
		t.Errorf("expected ID %s, got %s", sessionUUID, s.ID)
	}
	if s.ShortID != sessionUUID[:8] {
		t.Errorf("expected ShortID %s, got %s", sessionUUID[:8], s.ShortID)
	}
	if s.Agent != "claude" {
		t.Errorf("expected Agent 'claude', got %s", s.Agent)
	}
	if s.Profile != "work" {
		t.Errorf("expected Profile 'work', got %s", s.Profile)
	}
	if s.IsHost {
		t.Errorf("expected IsHost false, got true")
	}
	if s.Title != "Optimize database query" {
		t.Errorf("expected Title 'Optimize database query', got %q", s.Title)
	}
	if s.Summary != "Optimize database query" {
		t.Errorf("expected Summary 'Optimize database query', got %q", s.Summary)
	}
	if s.StoragePath != sessionFile {
		t.Errorf("expected StoragePath %s, got %s", sessionFile, s.StoragePath)
	}
	if s.MessageCount != 2 {
		t.Errorf("expected MessageCount 2 (1 user + 1 assistant), got %d", s.MessageCount)
	}
	expectedStartTime, _ := time.Parse(time.RFC3339, "2026-09-23T12:00:00.000Z")
	if !s.StartedAt.Equal(expectedStartTime) {
		t.Errorf("expected StartedAt %v, got %v", expectedStartTime, s.StartedAt)
	}
	expectedLastActive, _ := time.Parse(time.RFC3339, "2026-09-23T12:01:00.000Z")
	if !s.LastActiveAt.Equal(expectedLastActive) {
		t.Errorf("expected LastActiveAt %v, got %v", expectedLastActive, s.LastActiveAt)
	}
}

func TestProvider_ListSessions_HostDirectory(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("AIM_REAL_HOME", tempHome)

	projDir := filepath.Join(tempHome, ".claude", "projects", "-test-host-project")
	if err := os.MkdirAll(projDir, 0755); err != nil {
		t.Fatalf("failed to create host project dir: %v", err)
	}

	sessionUUID := "a1b2c3d4-e5f6-4789-8012-3456789abcde"
	jsonlContent := `{"type":"user","message":{"role":"user","content":"Host prompt"},"timestamp":"2026-09-23T15:30:00Z"}` + "\n"
	_ = os.WriteFile(filepath.Join(projDir, sessionUUID+".jsonl"), []byte(jsonlContent), 0644)

	p := NewProvider()
	ctx := context.Background()

	sessions, err := p.ListSessions(ctx, tempHome, true)
	if err != nil {
		t.Fatalf("ListSessions on host failed: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session on host, got %d", len(sessions))
	}

	s := sessions[0]
	if !s.IsHost {
		t.Errorf("expected IsHost true")
	}
	if s.Profile != "<host>" {
		t.Errorf("expected Profile '<host>', got %s", s.Profile)
	}
	if s.Title != "Host prompt" {
		t.Errorf("expected Title 'Host prompt', got %q", s.Title)
	}
}

func TestProvider_ListSessions_ContentVariations(t *testing.T) {
	tempDir := t.TempDir()
	profileDir := filepath.Join(tempDir, "profile")
	projDir := filepath.Join(profileDir, ".claude", "projects", "-sample-project")
	_ = os.MkdirAll(projDir, 0755)

	cases := []struct {
		name          string
		jsonl         string
		expectedTitle string
		messageCount  int
	}{
		{
			name: "content as block array",
			jsonl: strings.Join([]string{
				`{"type":"user","message":{"role":"user","content":[{"type":"text","text":"Array text block prompt"}]},"timestamp":"2026-09-23T10:00:00Z"}`,
				`{"type":"assistant","message":{"role":"assistant","content":"Done"},"timestamp":"2026-09-23T10:01:00Z"}`,
			}, "\n"),
			expectedTitle: "Array text block prompt",
			messageCount:  2,
		},
		{
			name: "content as top-level string on user line",
			jsonl: strings.Join([]string{
				`{"type":"user","content":"Direct string content","timestamp":"2026-09-23T10:00:00Z"}`,
			}, "\n"),
			expectedTitle: "Direct string content",
			messageCount:  1,
		},
		{
			name: "untitled fallback when no user message exists",
			jsonl: strings.Join([]string{
				`{"type":"file-read","path":"/tmp/foo.txt","timestamp":"2026-09-23T10:00:00Z"}`,
			}, "\n"),
			expectedTitle: "Untitled Session",
			messageCount:  0,
		},
		{
			name: "first non-empty user message selected when initial user event is empty",
			jsonl: strings.Join([]string{
				`{"type":"user","message":{"role":"user","content":""}}`,
				`{"type":"user","message":{"role":"user","content":"   \n  \t "}}`,
				`{"type":"user","message":{"role":"user","content":"Valid user question"}}`,
			}, "\n"),
			expectedTitle: "Valid user question",
			messageCount:  3,
		},
	}

	p := NewProvider()
	ctx := context.Background()

	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id := fmt.Sprintf("11111111-2222-4333-8444-55555555555%d", i)
			filePath := filepath.Join(projDir, id+".jsonl")
			if err := os.WriteFile(filePath, []byte(tc.jsonl+"\n"), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			s, err := p.GetSession(ctx, id, profileDir, false)
			if err != nil {
				t.Fatalf("GetSession failed: %v", err)
			}
			if s == nil {
				t.Fatalf("expected session not to be nil")
			}
			if s.Title != tc.expectedTitle {
				t.Errorf("expected Title %q, got %q", tc.expectedTitle, s.Title)
			}
			if s.MessageCount != tc.messageCount {
				t.Errorf("expected MessageCount %d, got %d", tc.messageCount, s.MessageCount)
			}
		})
	}
}

func TestProvider_ListSessions_MalformedAndLargeLines(t *testing.T) {
	tempDir := t.TempDir()
	profileDir := filepath.Join(tempDir, "profile")
	projDir := filepath.Join(profileDir, ".claude", "projects", "-sample-project")
	_ = os.MkdirAll(projDir, 0755)

	id := "22222222-3333-4444-8555-666666666666"
	filePath := filepath.Join(projDir, id+".jsonl")

	// Create a line with ~2MB content to test 10MB scanner buffer support
	largePadding := strings.Repeat("A", 2*1024*1024)
	jsonlContent := strings.Join([]string{
		`{not valid json line here`,
		`{"type":"user","message":{"role":"user","content":"Valid prompt after malformed line"},"timestamp":"2026-09-23T11:00:00Z"}`,
		`{"type":"assistant","message":{"role":"assistant","content":"` + largePadding + `"},"timestamp":"2026-09-23T11:05:00Z"}`,
		`another bad line}`,
	}, "\n") + "\n"

	if err := os.WriteFile(filePath, []byte(jsonlContent), 0644); err != nil {
		t.Fatalf("failed to write large jsonl: %v", err)
	}

	p := NewProvider()
	ctx := context.Background()

	s, err := p.GetSession(ctx, id, profileDir, false)
	if err != nil {
		t.Fatalf("GetSession failed on file with malformed & large lines: %v", err)
	}
	if s == nil {
		t.Fatalf("expected session not to be nil")
	}
	if s.Title != "Valid prompt after malformed line" {
		t.Errorf("expected valid prompt title, got %q", s.Title)
	}
	if s.MessageCount != 2 {
		t.Errorf("expected 2 valid messages counted, got %d", s.MessageCount)
	}
}

func TestProvider_GetSession(t *testing.T) {
	tempDir := t.TempDir()
	profileDir := filepath.Join(tempDir, "profile")
	projDir := filepath.Join(profileDir, ".claude", "projects", "-sample-project")
	_ = os.MkdirAll(projDir, 0755)

	id1 := "33333333-aaaa-4bbb-8ccc-111111111111"
	id2 := "33333333-bbbb-4bbb-8ccc-222222222222"
	id3 := "44444444-cccc-4ddd-8eee-333333333333"

	for _, id := range []string{id1, id2, id3} {
		_ = os.WriteFile(filepath.Join(projDir, id+".jsonl"), []byte(`{"type":"user","content":"Prompt for `+id+`"}`+"\n"), 0644)
	}

	p := NewProvider()
	ctx := context.Background()

	// 1. Exact match
	s, err := p.GetSession(ctx, id1, profileDir, false)
	if err != nil {
		t.Fatalf("GetSession exact failed: %v", err)
	}
	if s == nil || s.ID != id1 {
		t.Errorf("expected session %s, got %+v", id1, s)
	}

	// 2. Non-ambiguous short prefix match
	s, err = p.GetSession(ctx, "44444444", profileDir, false)
	if err != nil {
		t.Fatalf("GetSession short prefix failed: %v", err)
	}
	if s == nil || s.ID != id3 {
		t.Errorf("expected session %s, got %+v", id3, s)
	}

	// 3. Ambiguous prefix matching both id1 and id2
	_, err = p.GetSession(ctx, "33333333", profileDir, false)
	if err == nil {
		t.Fatalf("expected error for ambiguous prefix '33333333', got nil")
	}
	if !strings.Contains(err.Error(), "ambiguous session ID prefix") {
		t.Errorf("expected error message to contain 'ambiguous session ID prefix', got: %v", err)
	}

	// 4. Not found
	s, err = p.GetSession(ctx, "99999999", profileDir, false)
	if err != nil {
		t.Fatalf("expected no error for nonexistent session, got: %v", err)
	}
	if s != nil {
		t.Errorf("expected nil session for nonexistent ID, got %+v", s)
	}

	// 5. Empty ID
	s, err = p.GetSession(ctx, "", profileDir, false)
	if err != nil {
		t.Fatalf("expected no error for empty ID, got: %v", err)
	}
	if s != nil {
		t.Errorf("expected nil session for empty ID, got %+v", s)
	}
}

func TestProvider_Hydrate_ForkFalse(t *testing.T) {
	tempDir := t.TempDir()
	srcProfile := filepath.Join(tempDir, "src")
	destProfile := filepath.Join(tempDir, "dest")
	projectSlug := "-Users-nemesis-Projects-aim"
	srcProjDir := filepath.Join(srcProfile, ".claude", "projects", projectSlug)
	_ = os.MkdirAll(srcProjDir, 0755)

	origUUID := "d24699d0-820f-435f-b4e4-eaf8440311b4"
	jsonlContent := strings.Join([]string{
		`{"type":"last-prompt","leafUuid":"leaf-1","sessionId":"` + origUUID + `"}`,
		`{"type":"user","message":{"role":"user","content":"Initial prompt"},"sessionId":"` + origUUID + `"}`,
	}, "\n") + "\n"
	srcFile := filepath.Join(srcProjDir, origUUID+".jsonl")
	_ = os.WriteFile(srcFile, []byte(jsonlContent), 0644)

	p := NewProvider()
	ctx := context.Background()

	srcSession, err := p.GetSession(ctx, origUUID, srcProfile, false)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if srcSession == nil {
		t.Fatalf("source session is nil")
	}

	hydratedID, err := p.Hydrate(ctx, srcSession, destProfile, false)
	if err != nil {
		t.Fatalf("Hydrate failed: %v", err)
	}
	if hydratedID != origUUID {
		t.Errorf("expected hydratedID %s, got %s", origUUID, hydratedID)
	}

	destFile := filepath.Join(destProfile, ".claude", "projects", projectSlug, origUUID+".jsonl")
	destContent, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("failed to read hydrated destination file: %v", err)
	}
	if string(destContent) != jsonlContent {
		t.Errorf("expected exact byte-for-byte content copy")
	}
}

func TestProvider_Hydrate_ForkTrue(t *testing.T) {
	tempDir := t.TempDir()
	srcProfile := filepath.Join(tempDir, "src")
	destProfile := filepath.Join(tempDir, "dest")
	projectSlug := "-Users-nemesis-Projects-aim"
	srcProjDir := filepath.Join(srcProfile, ".claude", "projects", projectSlug)
	_ = os.MkdirAll(srcProjDir, 0755)

	origUUID := "d24699d0-820f-435f-b4e4-eaf8440311b4"
	jsonlContent := strings.Join([]string{
		`{"type":"last-prompt","leafUuid":"leaf-1","sessionId":"` + origUUID + `"}`,
		`{"type":"user","message":{"role":"user","content":"Initial prompt"},"sessionId":"` + origUUID + `"}`,
		`{"type":"assistant","message":{"role":"assistant","content":"Response"},"sessionId":"` + origUUID + `"}`,
	}, "\n") + "\n"
	srcFile := filepath.Join(srcProjDir, origUUID+".jsonl")
	_ = os.WriteFile(srcFile, []byte(jsonlContent), 0644)

	p := NewProvider()
	ctx := context.Background()

	srcSession, err := p.GetSession(ctx, origUUID, srcProfile, false)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}

	forkedID, err := p.Hydrate(ctx, srcSession, destProfile, true)
	if err != nil {
		t.Fatalf("Hydrate with fork=true failed: %v", err)
	}
	if forkedID == origUUID {
		t.Errorf("expected new UUID for fork, got same UUID %s", origUUID)
	}
	if !uuidV4Regex.MatchString(forkedID) {
		t.Errorf("expected valid RFC 4122 v4 UUID, got %q", forkedID)
	}

	forkFile := filepath.Join(destProfile, ".claude", "projects", projectSlug, forkedID+".jsonl")
	data, err := os.ReadFile(forkFile)
	if err != nil {
		t.Fatalf("failed to read forked destination file: %v", err)
	}

	forkedContent := string(data)
	if strings.Contains(forkedContent, origUUID) {
		t.Errorf("forked file still contains original UUID %s", origUUID)
	}
	if !strings.Contains(forkedContent, forkedID) {
		t.Errorf("forked file does not contain new UUID %s", forkedID)
	}
	if strings.Count(forkedContent, forkedID) != 3 {
		t.Errorf("expected 3 replacements of UUID, got %d", strings.Count(forkedContent, forkedID))
	}
}

func TestProvider_Hydrate_NilSession(t *testing.T) {
	p := NewProvider()
	ctx := context.Background()
	_, err := p.Hydrate(ctx, nil, "/tmp/dest", false)
	if err == nil {
		t.Fatalf("expected error for nil session, got nil")
	}
}

func TestProvider_MissingProjectsDirectory(t *testing.T) {
	tempDir := t.TempDir()
	p := NewProvider()
	ctx := context.Background()

	sessions, err := p.ListSessions(ctx, tempDir, false)
	if err != nil {
		t.Fatalf("expected nil error on nonexistent projects dir, got: %v", err)
	}
	if len(sessions) != 0 {
		t.Fatalf("expected 0 sessions on nonexistent projects dir, got %d", len(sessions))
	}
}

func TestProvider_Hydrate_SameFileNoTruncate(t *testing.T) {
	tempDir := t.TempDir()
	profileDir := filepath.Join(tempDir, "work")
	projectSlug := "-Users-nemesis-Projects-aim"
	projDir := filepath.Join(profileDir, ".claude", "projects", projectSlug)
	if err := os.MkdirAll(projDir, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}

	sessionUUID := "d24699d0-820f-435f-b4e4-eaf8440311b4"
	jsonlContent := `{"type":"user","message":{"role":"user","content":"Do not truncate me"},"sessionId":"` + sessionUUID + `"}` + "\n"
	sessionFile := filepath.Join(projDir, sessionUUID+".jsonl")
	if err := os.WriteFile(sessionFile, []byte(jsonlContent), 0644); err != nil {
		t.Fatalf("failed to write session file: %v", err)
	}

	p := NewProvider()
	ctx := context.Background()

	sess, err := p.GetSession(ctx, sessionUUID, profileDir, false)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}

	// Hydrate into the same profile directory without fork
	returnedID, err := p.Hydrate(ctx, sess, profileDir, false)
	if err != nil {
		t.Fatalf("Hydrate on same file failed: %v", err)
	}
	if returnedID != sessionUUID {
		t.Errorf("expected %s, got %s", sessionUUID, returnedID)
	}

	data, err := os.ReadFile(sessionFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(data) != jsonlContent {
		t.Errorf("file was corrupted or truncated: got %q", string(data))
	}
}

func TestProvider_ListSessions_ContextCanceled(t *testing.T) {
	tempDir := t.TempDir()
	projDir := filepath.Join(tempDir, ".claude", "projects", "slug1")
	_ = os.MkdirAll(projDir, 0755)

	p := NewProvider()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancel context

	_, err := p.ListSessions(ctx, tempDir, false)
	if err == nil {
		t.Fatalf("expected context cancellation error, got nil")
	}
}
