package codex_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/session"
	"github.com/aim-cli/aim/internal/session/providers/codex"
)

func setupMockCodex(t *testing.T) string {
	t.Helper()
	sqliteBin, err := exec.LookPath("sqlite3")
	if err != nil {
		t.Skip("sqlite3 binary not available in PATH")
	}

	tmpDir := t.TempDir()
	codexDir := filepath.Join(tmpDir, ".codex")
	if err := os.MkdirAll(codexDir, 0755); err != nil {
		t.Fatalf("failed to create mock codex dir: %v", err)
	}

	dbPath := filepath.Join(codexDir, "state_5.sqlite")
	schema := `
CREATE TABLE threads (
	id TEXT PRIMARY KEY,
	rollout_path TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL,
	title TEXT NOT NULL,
	preview TEXT NOT NULL DEFAULT ''
);
INSERT INTO threads (id, rollout_path, created_at, updated_at, title, preview)
VALUES 
('01a09eb7-2f6c-7c52-895f-218f9ac9eecd', '/tmp/rollout1.jsonl', 1726300000, 1726300000, 'Refactor router middleware', 'Clean up routes'),
('01a04baf-4b95-7681-82bd-81df4f995452', '/tmp/rollout2.jsonl', 1726200000, 1726200000, 'Update docs', 'Fix outdated markdown');
`
	if err := exec.Command(sqliteBin, dbPath, schema).Run(); err != nil {
		t.Fatalf("failed to seed mock sqlite DB: %v", err)
	}

	// Also write a mock session_index.jsonl
	indexFile := filepath.Join(codexDir, "session_index.jsonl")
	indexContent := `{"id":"01a09eb7-2f6c-7c52-895f-218f9ac9eecd","thread_name":"Refactor router middleware","updated_at":"2026-09-14T07:12:52Z"}
{"id":"01a04baf-4b95-7681-82bd-81df4f995452","thread_name":"Update docs","updated_at":"2026-09-13T04:03:40Z"}
`
	_ = os.WriteFile(indexFile, []byte(indexContent), 0644)

	return tmpDir
}

func TestProvider_ListSessions(t *testing.T) {
	mockProfileDir := setupMockCodex(t)
	p := codex.NewProvider()

	if p.Agent() != "codex" {
		t.Fatalf("expected agent 'codex', got %q", p.Agent())
	}

	ctx := context.Background()
	sessions, err := p.ListSessions(ctx, mockProfileDir, false)
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}

	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}

	if sessions[0].ID != "01a09eb7-2f6c-7c52-895f-218f9ac9eecd" {
		t.Errorf("expected first session 01a09eb7..., got %s", sessions[0].ID)
	}
	if sessions[0].ShortID != "01a09eb7" {
		t.Errorf("expected ShortID 01a09eb7, got %s", sessions[0].ShortID)
	}
	if sessions[0].Title != "Refactor router middleware" {
		t.Errorf("expected title 'Refactor router middleware', got %q", sessions[0].Title)
	}
}

func TestProvider_GetSession(t *testing.T) {
	mockProfileDir := setupMockCodex(t)
	p := codex.NewProvider()
	ctx := context.Background()

	s, err := p.GetSession(ctx, "01a09eb7", mockProfileDir, false)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if s == nil || s.ID != "01a09eb7-2f6c-7c52-895f-218f9ac9eecd" {
		t.Fatalf("expected session 01a09eb7..., got %+v", s)
	}
}

func TestProvider_Hydrate(t *testing.T) {
	srcDir := setupMockCodex(t)
	_ = srcDir
	targetDir := t.TempDir()

	p := codex.NewProvider()
	ctx := context.Background()

	// Create a dummy rollout file
	dummyRollout := filepath.Join(t.TempDir(), "dummy-rollout.jsonl")
	_ = os.WriteFile(dummyRollout, []byte(`{"turn": 1}`), 0644)

	srcSession := session.NewSession("01a09eb7-2f6c-7c52-895f-218f9ac9eecd", "Test Hydrate", "codex", "<host>", true, time.Now())
	srcSession.StoragePath = dummyRollout
	srcSession.Summary = "Preview text"

	newID, err := p.Hydrate(ctx, &srcSession, targetDir, false)
	if err != nil {
		t.Fatalf("Hydrate failed: %v", err)
	}
	if newID != srcSession.ID {
		t.Errorf("expected newID %s, got %s", srcSession.ID, newID)
	}

	// Verify target has session file and sqlite entry
	targetDB := filepath.Join(targetDir, ".codex", "state_5.sqlite")
	if _, err := os.Stat(targetDB); err != nil {
		t.Errorf("expected target state_5.sqlite to exist: %v", err)
	}
}

func TestProvider_Hydrate_Fork(t *testing.T) {
	setupMockCodex(t)
	targetDir := t.TempDir()

	p := codex.NewProvider()
	ctx := context.Background()

	dummyRollout := filepath.Join(t.TempDir(), "dummy-rollout.jsonl")
	_ = os.WriteFile(dummyRollout, []byte(`{"turn": 1}`), 0644)

	srcSession := session.NewSession("01a09eb7-2f6c-7c52-895f-218f9ac9eecd", "Fork Test", "codex", "<host>", true, time.Now())
	srcSession.StoragePath = dummyRollout

	forkedID, err := p.Hydrate(ctx, &srcSession, targetDir, true)
	if err != nil {
		t.Fatalf("Hydrate fork failed: %v", err)
	}
	if forkedID == srcSession.ID {
		t.Errorf("expected forked ID to be different from source ID, got %s", forkedID)
	}
	if len(forkedID) != 36 {
		t.Errorf("expected 36-character UUID for forked ID, got %s", forkedID)
	}

	// Verify rollout file was copied with forked ID
	forkedRollout := filepath.Join(targetDir, ".codex", "sessions", "rollout-"+forkedID+".jsonl")
	if _, err := os.Stat(forkedRollout); err != nil {
		t.Errorf("expected forked rollout file to exist: %v", err)
	}
}

func TestProvider_SQLInjectionSafety(t *testing.T) {
	mockProfileDir := setupMockCodex(t)
	p := codex.NewProvider()
	ctx := context.Background()

	// 1. Malicious session ID in Hydrate
	maliciousSession := session.NewSession("'; DROP TABLE threads; --", "Hacked Title", "codex", "<host>", true, time.Now())
	targetDir := t.TempDir()
	_, err := p.Hydrate(ctx, &maliciousSession, targetDir, false)
	if err == nil {
		t.Errorf("expected error for malicious session ID in Hydrate, got nil")
	}

	// 2. Querying with malicious prefix
	s, err := p.GetSession(ctx, "'; DROP TABLE threads; --", mockProfileDir, false)
	if s != nil {
		t.Errorf("malicious prefix unexpectedly returned a session")
	}
	_ = err
}

func setupMockCodexFull(t *testing.T) (string, string, string) {
	t.Helper()
	sqliteBin, err := exec.LookPath("sqlite3")
	if err != nil {
		t.Skip("sqlite3 binary not available in PATH")
	}

	tmpDir := t.TempDir()
	codexDir := filepath.Join(tmpDir, ".codex")
	sessionsDir := filepath.Join(codexDir, "sessions", "2026", "09", "17")
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		t.Fatalf("failed to create mock sessions dir: %v", err)
	}

	parentID := "01a0a91d-9187-7021-a499-0ff13f9df264"
	childID := "01a0aedb-a8f2-71e2-85d8-cf0479089899"

	parentRolloutPath := filepath.Join(sessionsDir, fmt.Sprintf("rollout-%s.jsonl", parentID))
	childRolloutPath := filepath.Join(sessionsDir, fmt.Sprintf("rollout-%s.jsonl", childID))

	parentContent := fmt.Sprintf("{\"type\":\"session_meta\",\"payload\":{\"id\":\"%s\"}}\n{\"type\":\"event_msg\",\"payload\":{\"thread_id\":\"%s\"}}\n", parentID, parentID)
	childContent := fmt.Sprintf("{\"type\":\"session_meta\",\"payload\":{\"id\":\"%s\",\"history_base\":{\"thread_id\":\"%s\"}}}\n{\"type\":\"event_msg\",\"payload\":{\"thread_id\":\"%s\"}}\n", childID, parentID, childID)

	_ = os.WriteFile(parentRolloutPath, []byte(parentContent), 0644)
	_ = os.WriteFile(childRolloutPath, []byte(childContent), 0644)

	// Create state_5.sqlite
	dbPath := filepath.Join(codexDir, "state_5.sqlite")
	stateSchema := fmt.Sprintf(`
CREATE TABLE _sqlx_migrations (
    version BIGINT PRIMARY KEY,
    description TEXT NOT NULL,
    installed_on TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    success BOOLEAN NOT NULL,
    checksum BLOB NOT NULL,
    execution_time BIGINT NOT NULL
);
INSERT INTO _sqlx_migrations (version, description, success, checksum, execution_time) VALUES (1, 'init', 1, X'00', 1);

CREATE TABLE threads (
    id TEXT PRIMARY KEY,
    rollout_path TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    source TEXT NOT NULL,
    model_provider TEXT NOT NULL,
    cwd TEXT NOT NULL,
    title TEXT NOT NULL,
    sandbox_policy TEXT NOT NULL,
    approval_mode TEXT NOT NULL,
    history_mode TEXT NOT NULL DEFAULT 'legacy'
);

INSERT INTO threads (id, rollout_path, created_at, updated_at, source, model_provider, cwd, title, sandbox_policy, approval_mode, history_mode)
VALUES 
('%s', '%s', 1726000000, 1726000000, 'cli', 'caveman', '/workspace/project', 'Parent Session', 'workspace-write', 'ask', 'paginated'),
('%s', '%s', 1726100000, 1726100000, 'cli', 'caveman', '/workspace/project', 'Child Session', 'workspace-write', 'ask', 'paginated');
`, parentID, parentRolloutPath, childID, childRolloutPath)

	if err := exec.Command(sqliteBin, dbPath, stateSchema).Run(); err != nil {
		t.Fatalf("failed to create mock state_5.sqlite: %v", err)
	}

	// Create thread_history_1.sqlite
	historyDB := filepath.Join(codexDir, "thread_history_1.sqlite")
	historySchema := fmt.Sprintf(`
CREATE TABLE _sqlx_migrations (
    version BIGINT PRIMARY KEY,
    description TEXT NOT NULL,
    installed_on TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    success BOOLEAN NOT NULL,
    checksum BLOB NOT NULL,
    execution_time BIGINT NOT NULL
);
INSERT INTO _sqlx_migrations (version, description, success, checksum, execution_time) VALUES (1, 'init', 1, X'00', 1);

CREATE TABLE thread_turns (
    thread_id TEXT NOT NULL,
    turn_id TEXT NOT NULL,
    rollout_ordinal INTEGER NOT NULL,
    status TEXT NOT NULL,
    PRIMARY KEY (thread_id, turn_id)
);

CREATE TABLE thread_items (
    thread_id TEXT NOT NULL,
    turn_id TEXT NOT NULL,
    item_id TEXT NOT NULL,
    rollout_ordinal INTEGER NOT NULL,
    item_json TEXT NOT NULL,
    PRIMARY KEY (thread_id, turn_id, item_id)
);

INSERT INTO thread_turns (thread_id, turn_id, rollout_ordinal, status)
VALUES 
('%s', 'turn-p1', 1, 'completed'),
('%s', 'turn-c1', 1, 'completed');

INSERT INTO thread_items (thread_id, turn_id, item_id, rollout_ordinal, item_json)
VALUES
('%s', 'turn-p1', 'item-p1', 1, '{"content":"parent message"}'),
('%s', 'turn-c1', 'item-c1', 1, '{"content":"child message"}');
`, parentID, childID, parentID, childID)

	if err := exec.Command(sqliteBin, historyDB, historySchema).Run(); err != nil {
		t.Fatalf("failed to create mock thread_history_1.sqlite: %v", err)
	}

	return tmpDir, parentID, childID
}

func TestProvider_Hydrate_FullFidelity(t *testing.T) {
	srcDir, parentID, childID := setupMockCodexFull(t)
	targetDir := t.TempDir()

	p := codex.NewProvider()
	ctx := context.Background()

	childRollout := filepath.Join(srcDir, ".codex", "sessions", "2026", "09", "17", fmt.Sprintf("rollout-%s.jsonl", childID))
	srcSession := session.NewSession(childID, "Child Session", "codex", "mockprofile", false, time.Now())
	srcSession.StoragePath = childRollout

	hydratedID, err := p.Hydrate(ctx, &srcSession, targetDir, false)
	if err != nil {
		t.Fatalf("Hydrate failed: %v", err)
	}
	if hydratedID != childID {
		t.Errorf("expected hydratedID %s, got %s", childID, hydratedID)
	}

	sqliteBin, err := exec.LookPath("sqlite3")
	if err != nil {
		t.Skip("sqlite3 binary not available in PATH")
	}
	targetDB := filepath.Join(targetDir, ".codex", "state_5.sqlite")
	targetHistoryDB := filepath.Join(targetDir, ".codex", "thread_history_1.sqlite")

	// 1. Verify child thread in state_5.sqlite
	out, err := exec.Command(sqliteBin, targetDB, fmt.Sprintf("SELECT cwd, model_provider, history_mode FROM threads WHERE id = '%s';", childID)).Output()
	if err != nil {
		t.Fatalf("failed to query target state_5: %v", err)
	}
	res := strings.TrimSpace(string(out))
	if !strings.Contains(res, "/workspace/project") {
		t.Errorf("expected preserved cwd /workspace/project, got %q", res)
	}
	if !strings.Contains(res, "caveman") {
		t.Errorf("expected preserved model_provider caveman, got %q", res)
	}
	if !strings.Contains(res, "paginated") {
		t.Errorf("expected preserved history_mode paginated, got %q", res)
	}

	// 2. Verify parent thread also copied into state_5.sqlite
	out, err = exec.Command(sqliteBin, targetDB, fmt.Sprintf("SELECT COUNT(*) FROM threads WHERE id = '%s';", parentID)).Output()
	if err != nil || strings.TrimSpace(string(out)) != "1" {
		t.Errorf("expected parent thread in target state_5: %v, out: %s", err, string(out))
	}

	// 3. Verify turns and items copied into target thread_history_1.sqlite
	out, err = exec.Command(sqliteBin, targetHistoryDB, fmt.Sprintf("SELECT COUNT(*) FROM thread_turns WHERE thread_id = '%s';", childID)).Output()
	if err != nil || strings.TrimSpace(string(out)) != "1" {
		t.Errorf("expected child turns in target thread_history_1: %v, out: %s", err, string(out))
	}
	out, err = exec.Command(sqliteBin, targetHistoryDB, fmt.Sprintf("SELECT COUNT(*) FROM thread_turns WHERE thread_id = '%s';", parentID)).Output()
	if err != nil || strings.TrimSpace(string(out)) != "1" {
		t.Errorf("expected parent turns in target thread_history_1: %v, out: %s", err, string(out))
	}

	// 4. Verify rollout files exist
	destChildRollout := filepath.Join(targetDir, ".codex", "sessions", "2026", "09", "17", fmt.Sprintf("rollout-%s.jsonl", childID))
	if _, err := os.Stat(destChildRollout); err != nil {
		t.Errorf("expected dest child rollout file %s to exist: %v", destChildRollout, err)
	}
	destParentRollout := filepath.Join(targetDir, ".codex", "sessions", "2026", "09", "17", fmt.Sprintf("rollout-%s.jsonl", parentID))
	if _, err := os.Stat(destParentRollout); err != nil {
		t.Errorf("expected dest parent rollout file %s to exist: %v", destParentRollout, err)
	}
}

func TestProvider_Hydrate_AncestorRolloutUpdated(t *testing.T) {
	srcDir, parentID, childID := setupMockCodexFull(t)
	targetDir := t.TempDir()

	// Pre-create destination parent rollout with smaller/stale content
	destParentDir := filepath.Join(targetDir, ".codex", "sessions", "2026", "09", "17")
	if err := os.MkdirAll(destParentDir, 0755); err != nil {
		t.Fatalf("failed to create destParentDir: %v", err)
	}
	destParentRollout := filepath.Join(destParentDir, fmt.Sprintf("rollout-%s.jsonl", parentID))
	staleContent := `{"type":"event_msg","payload":{"type":"task_started"}}` + "\n"
	if err := os.WriteFile(destParentRollout, []byte(staleContent), 0644); err != nil {
		t.Fatalf("failed to write stale parent rollout: %v", err)
	}

	p := codex.NewProvider()
	ctx := context.Background()

	childRollout := filepath.Join(srcDir, ".codex", "sessions", "2026", "09", "17", fmt.Sprintf("rollout-%s.jsonl", childID))
	srcSession := session.NewSession(childID, "Child Session", "codex", "mockprofile", false, time.Now())
	srcSession.StoragePath = childRollout

	hydratedID, err := p.Hydrate(ctx, &srcSession, targetDir, false)
	if err != nil {
		t.Fatalf("Hydrate failed: %v", err)
	}
	if hydratedID != childID {
		t.Errorf("expected hydratedID %s, got %s", childID, hydratedID)
	}

	// Verify destParentRollout was updated to the full content from srcDir
	srcParentRollout := filepath.Join(srcDir, ".codex", "sessions", "2026", "09", "17", fmt.Sprintf("rollout-%s.jsonl", parentID))
	srcStat, err := os.Stat(srcParentRollout)
	if err != nil {
		t.Fatalf("failed to stat src parent rollout: %v", err)
	}
	destStat, err := os.Stat(destParentRollout)
	if err != nil {
		t.Fatalf("failed to stat dest parent rollout: %v", err)
	}

	if destStat.Size() != srcStat.Size() {
		t.Errorf("expected dest parent rollout size %d, got stale size %d (stale was %d)", srcStat.Size(), destStat.Size(), len(staleContent))
	}
}

func TestProvider_Hydrate_Fork_FullFidelity(t *testing.T) {
	srcDir, _, childID := setupMockCodexFull(t)
	targetDir := t.TempDir()

	p := codex.NewProvider()
	ctx := context.Background()

	childRollout := filepath.Join(srcDir, ".codex", "sessions", "2026", "09", "17", fmt.Sprintf("rollout-%s.jsonl", childID))
	srcSession := session.NewSession(childID, "Child Session", "codex", "mockprofile", false, time.Now())
	srcSession.StoragePath = childRollout

	forkedID, err := p.Hydrate(ctx, &srcSession, targetDir, true)
	if err != nil {
		t.Fatalf("Hydrate fork failed: %v", err)
	}
	if forkedID == childID {
		t.Fatalf("expected forked ID different from child ID")
	}

	sqliteBin, err := exec.LookPath("sqlite3")
	if err != nil {
		t.Skip("sqlite3 binary not available in PATH")
	}
	targetDB := filepath.Join(targetDir, ".codex", "state_5.sqlite")
	targetHistoryDB := filepath.Join(targetDir, ".codex", "thread_history_1.sqlite")

	// 1. Verify forked thread in state_5.sqlite
	out, err := exec.Command(sqliteBin, targetDB, fmt.Sprintf("SELECT cwd, model_provider FROM threads WHERE id = '%s';", forkedID)).Output()
	if err != nil {
		t.Fatalf("failed to query target state_5 for forked thread: %v", err)
	}
	res := strings.TrimSpace(string(out))
	if !strings.Contains(res, "/workspace/project") {
		t.Errorf("expected preserved cwd /workspace/project, got %q", res)
	}
	if !strings.Contains(res, "caveman") {
		t.Errorf("expected preserved model_provider caveman, got %q", res)
	}

	// 2. Verify turns copied with new forked ID in thread_history_1.sqlite
	out, err = exec.Command(sqliteBin, targetHistoryDB, fmt.Sprintf("SELECT COUNT(*) FROM thread_turns WHERE thread_id = '%s';", forkedID)).Output()
	if err != nil || strings.TrimSpace(string(out)) != "1" {
		t.Errorf("expected forked turns in target thread_history_1: %v, out: %s", err, string(out))
	}

	// 3. Verify forked rollout file has replaced ID
	destForkedRollout := filepath.Join(targetDir, ".codex", "sessions", "2026", "09", "17", fmt.Sprintf("rollout-%s.jsonl", forkedID))
	content, err := os.ReadFile(destForkedRollout)
	if err != nil {
		t.Fatalf("failed to read forked rollout file: %v", err)
	}
	if strings.Contains(string(content), childID) {
		t.Errorf("forked rollout still contains original childID %s", childID)
	}
	if !strings.Contains(string(content), forkedID) {
		t.Errorf("forked rollout does not contain forkedID %s", forkedID)
	}
}

func TestProvider_ListSessions_SchemaDetectionAndFilterSubagents(t *testing.T) {
	sqliteBin, err := exec.LookPath("sqlite3")
	if err != nil {
		t.Skip("sqlite3 binary not available in PATH")
	}

	tmpDir := t.TempDir()
	codexDir := filepath.Join(tmpDir, ".codex")
	if err := os.MkdirAll(codexDir, 0755); err != nil {
		t.Fatalf("failed to create codex dir: %v", err)
	}

	dbPath := filepath.Join(codexDir, "state_5.sqlite")
	schema := `
CREATE TABLE threads (
	id TEXT PRIMARY KEY,
	name TEXT,
	title TEXT NOT NULL,
	first_user_message TEXT,
	preview TEXT NOT NULL DEFAULT '',
	cwd TEXT,
	thread_source TEXT,
	archived INTEGER DEFAULT 0,
	updated_at INTEGER NOT NULL,
	rollout_path TEXT NOT NULL
);
INSERT INTO threads (id, name, title, first_user_message, preview, cwd, thread_source, archived, updated_at, rollout_path)
VALUES 
('01a0b351-2f2f-7d22-8176-49e45bde8f9b', 'cc-ov2', 'Model Generated Title', 'First prompt text', 'Preview text', '/workspace/repo', 'user', 0, 1726000000, '/tmp/rollout1.jsonl'),
('01a0sub1-1111-2222-3333-444444444444', NULL, 'Subagent Maxwell', 'Subagent prompt', '', '/workspace/repo', 'subagent', 0, 1726100000, '/tmp/rollout2.jsonl'),
('01a0arch-5555-6666-7777-888888888888', 'archived-sess', 'Old Session', 'Old prompt', '', '/workspace/repo', 'user', 1, 1726050000, '/tmp/rollout3.jsonl');
`
	if err := exec.Command(sqliteBin, dbPath, schema).Run(); err != nil {
		t.Fatalf("failed to seed mock sqlite DB: %v", err)
	}

	p := codex.NewProvider()
	ctx := context.Background()

	sessions, err := p.ListSessions(ctx, tmpDir, false)
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}

	if len(sessions) != 1 {
		t.Fatalf("expected exactly 1 session (subagent and archived filtered out), got %d", len(sessions))
	}

	s := sessions[0]
	if s.ID != "01a0b351-2f2f-7d22-8176-49e45bde8f9b" {
		t.Errorf("expected session ID 01a0b351..., got %s", s.ID)
	}
	if s.Title != "cc-ov2" {
		t.Errorf("expected custom name 'cc-ov2' as title, got %q", s.Title)
	}
	if s.Cwd != "/workspace/repo" {
		t.Errorf("expected Cwd '/workspace/repo', got %q", s.Cwd)
	}

	// Also verify GetSession retrieves it directly with title 'cc-ov2'
	retrieved, err := p.GetSession(ctx, "01a0b351", tmpDir, false)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if retrieved == nil || retrieved.Title != "cc-ov2" {
		t.Fatalf("expected retrieved session with title 'cc-ov2', got %+v", retrieved)
	}
	if retrieved.Cwd != "/workspace/repo" {
		t.Errorf("expected retrieved Cwd '/workspace/repo', got %q", retrieved.Cwd)
	}

	// Verify GetSession retrieves by session name 'cc-ov2'
	retrievedByName, err := p.GetSession(ctx, "cc-ov2", tmpDir, false)
	if err != nil {
		t.Fatalf("GetSession by name failed: %v", err)
	}
	if retrievedByName == nil || retrievedByName.ID != "01a0b351-2f2f-7d22-8176-49e45bde8f9b" {
		t.Fatalf("expected session 01a0b351 retrieved by name, got %+v", retrievedByName)
	}
}

func TestProvider_MultilinePromptAndLookupByName(t *testing.T) {
	sqliteBin, err := exec.LookPath("sqlite3")
	if err != nil {
		t.Skip("sqlite3 binary not available in PATH")
	}

	tmpDir := t.TempDir()
	codexDir := filepath.Join(tmpDir, ".codex")
	if err := os.MkdirAll(codexDir, 0755); err != nil {
		t.Fatalf("failed to create mock codex dir: %v", err)
	}

	dbPath := filepath.Join(codexDir, "state_5.sqlite")
	schema := `
CREATE TABLE threads (
	id TEXT PRIMARY KEY,
	name TEXT,
	title TEXT,
	first_user_message TEXT,
	preview TEXT,
	cwd TEXT,
	thread_source TEXT,
	archived INTEGER DEFAULT 0,
	updated_at INTEGER NOT NULL,
	rollout_path TEXT NOT NULL
);
INSERT INTO threads (id, name, title, first_user_message, preview, cwd, thread_source, archived, updated_at, rollout_path)
VALUES 
('01a0b8b4-3942-7912-a78e-a1cd76748eb7', 'dsl-delete', 'Review a three-repo change that adds DSL connector deletion to Hevo.\nIt replaces an earlier soft-delete design.\n\n## Read in this order\n1. Design + plan', 'First prompt\nwith multiple lines\nand markdown', 'Preview line 1\nPreview line 2', '/Users/nemesis/Projects/hevo-data/dsl-connector', 'user', 0, 1726744883, '/tmp/rollout-dsl.jsonl');
`
	if err := exec.Command(sqliteBin, dbPath, schema).Run(); err != nil {
		t.Fatalf("failed to seed mock sqlite DB: %v", err)
	}

	p := codex.NewProvider()
	ctx := context.Background()

	// 1. Verify ListSessions returns the multiline session
	sessions, err := p.ListSessions(ctx, tmpDir, false)
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	if sessions[0].ID != "01a0b8b4-3942-7912-a78e-a1cd76748eb7" {
		t.Errorf("expected session 01a0b8b4..., got %s", sessions[0].ID)
	}
	if sessions[0].Title != "dsl-delete" {
		t.Errorf("expected title 'dsl-delete', got %q", sessions[0].Title)
	}

	// 2. Verify GetSession by prefix
	s1, err := p.GetSession(ctx, "01a0b8b4", tmpDir, false)
	if err != nil {
		t.Fatalf("GetSession by prefix failed: %v", err)
	}
	if s1 == nil || s1.ID != "01a0b8b4-3942-7912-a78e-a1cd76748eb7" {
		t.Fatalf("expected session 01a0b8b4, got %+v", s1)
	}

	// 3. Verify GetSession by name
	s2, err := p.GetSession(ctx, "dsl-delete", tmpDir, false)
	if err != nil {
		t.Fatalf("GetSession by name failed: %v", err)
	}
	if s2 == nil || s2.ID != "01a0b8b4-3942-7912-a78e-a1cd76748eb7" {
		t.Fatalf("expected session 01a0b8b4, got %+v", s2)
	}
}
