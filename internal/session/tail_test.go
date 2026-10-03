package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadTurnCounts(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "transcript.jsonl")

	content := strings.Join([]string{
		`{"type":"system","message":"Session started"}`,
		`{"type":"user","message":"First user request"}`,
		`{"type":"assistant","message":"Assistant reply 1"}`,
		`{"type":"USER_INPUT","message":"Second user request"}`,
		`{"type":"assistant","message":"Assistant reply 2"}`,
		`{"type":"tool_call","message":"Calling tool"}`,
		`{"type":"user","message":"Third user request"}`,
	}, "\n") + "\n"

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	f, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer f.Close()

	turns, err := ReadTurnCounts(f)
	if err != nil {
		t.Fatalf("ReadTurnCounts failed: %v", err)
	}
	if turns != 3 {
		t.Errorf("expected 3 turns (2 'user' + 1 'USER_INPUT'), got %d", turns)
	}
}

func TestReadMessageCounts(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "transcript.jsonl")

	content := strings.Join([]string{
		`{"type":"system","message":"Session started"}`,
		`{"type":"user","message":"First user request"}`,
		`{"type":"assistant","message":"Assistant reply 1"}`,
		`{"type":"USER_INPUT","message":"Second user request"}`,
		`{"type":"assistant","message":"Assistant reply 2"}`,
		`{"type":"tool_call","message":"Calling tool"}`,
		`{"type":"user","message":"Third user request"}`,
	}, "\n") + "\n"

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	f, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer f.Close()

	counts, err := ReadMessageCounts(f)
	if err != nil {
		t.Fatalf("ReadMessageCounts failed: %v", err)
	}
	if counts != 5 {
		t.Errorf("expected 5 messages (3 user + 2 assistant), got %d", counts)
	}
}

func TestReadTurnCounts_LargeLineSafety(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "large.jsonl")

	// 2MB line padding
	largePadding := strings.Repeat("x", 2*1024*1024)
	content := strings.Join([]string{
		`{"type":"user","payload":"` + largePadding + `"}`,
		`{"type":"assistant","payload":"ok"}`,
		`{"type":"USER_INPUT","payload":"next"}`,
	}, "\n") + "\n"

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write large file: %v", err)
	}

	f, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("failed to open large file: %v", err)
	}
	defer f.Close()

	turns, err := ReadTurnCounts(f)
	if err != nil {
		t.Fatalf("ReadTurnCounts failed on large line: %v", err)
	}
	if turns != 2 {
		t.Errorf("expected 2 turns, got %d", turns)
	}
}

func TestReadTurnCounts_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "empty.jsonl")
	_ = os.WriteFile(filePath, []byte(""), 0644)

	f, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("failed to open empty file: %v", err)
	}
	defer f.Close()

	turns, err := ReadTurnCounts(f)
	if err != nil {
		t.Fatalf("expected nil error on empty file, got: %v", err)
	}
	if turns != 0 {
		t.Errorf("expected 0 turns for empty file, got %d", turns)
	}
}
