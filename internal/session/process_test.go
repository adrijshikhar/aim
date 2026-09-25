package session_test

import (
	"strings"
	"testing"

	"github.com/aim-cli/aim/internal/session"
)

func TestParseProcessLines(t *testing.T) {
	sampleOutput := `
  PID COMMAND
21484 aim run agy bby --dangerously-skip-permissions --conversation=775e6ada-1595-4e7e-84fa-ce0ea71e3007
21485 /Users/nemesis/.local/bin/agy --dangerously-skip-permissions --conversation=775e6ada-1595-4e7e-84fa-ce0ea71e3007
92974 aim run agy work --dangerously-skip-permissions --conversation=fcdbc2e0-2dc8-4ffa-9ee2-eb5aaa3e556f
33523 agy --dangerously-skip-permissions --conversation=c49a97be-0342-43cc-a90f-85a359959aaf
74382 codex --yolo resume 01a09eb7-2f6c-7c52-895f-218f9ac9eecd
55746 /Applications/ChatGPT.app/Contents/Resources/codex app-server --listen stdio://
21081 aim run codex office --yolo resume 01a0b351-2f2f-7d22-8176-49e45bde8f9b
66778 aim run claude work --resume 55555555-6666-7777-8888-999999999999
`
	active := session.ParseProcessOutput(strings.NewReader(sampleOutput))
	if len(active) == 0 {
		t.Fatal("expected active processes to be found, got none")
	}

	// Verify aim run agy bby
	info1, ok := active["775e6ada-1595-4e7e-84fa-ce0ea71e3007"]
	if !ok {
		t.Errorf("expected 775e6ada-1595-4e7e-84fa-ce0ea71e3007 to be active")
	} else {
		if info1.Agent != "agy" {
			t.Errorf("expected agent 'agy', got %q", info1.Agent)
		}
		if info1.Profile != "bby" {
			t.Errorf("expected profile 'bby', got %q", info1.Profile)
		}
		if info1.PID != 21484 && info1.PID != 21485 {
			t.Errorf("unexpected PID %d", info1.PID)
		}
	}

	// Verify host native agy run
	info2, ok := active["c49a97be-0342-43cc-a90f-85a359959aaf"]
	if !ok {
		t.Errorf("expected c49a97be-0342-43cc-a90f-85a359959aaf to be active")
	} else {
		if info2.Agent != "agy" {
			t.Errorf("expected agent 'agy', got %q", info2.Agent)
		}
		if info2.Profile != "<host>" {
			t.Errorf("expected profile '<host>', got %q", info2.Profile)
		}
		if info2.PID != 33523 {
			t.Errorf("expected PID 33523, got %d", info2.PID)
		}
	}

	// Verify codex resume
	info3, ok := active["01a09eb7-2f6c-7c52-895f-218f9ac9eecd"]
	if !ok {
		t.Errorf("expected 01a09eb7-2f6c-7c52-895f-218f9ac9eecd to be active")
	} else {
		if info3.Agent != "codex" {
			t.Errorf("expected agent 'codex', got %q", info3.Agent)
		}
		if info3.PID != 74382 {
			t.Errorf("expected PID 74382, got %d", info3.PID)
		}
	}

	// Verify aim run codex office resume
	info4, ok := active["01a0b351-2f2f-7d22-8176-49e45bde8f9b"]
	if !ok {
		t.Errorf("expected 01a0b351-2f2f-7d22-8176-49e45bde8f9b to be active")
	} else {
		if info4.Agent != "codex" {
			t.Errorf("expected agent 'codex', got %q", info4.Agent)
		}
		if info4.Profile != "office" {
			t.Errorf("expected profile 'office', got %q", info4.Profile)
		}
		if info4.PID != 21081 {
			t.Errorf("expected PID 21081, got %d", info4.PID)
		}
	}

	// Verify aim run claude work --resume
	info5, ok := active["55555555-6666-7777-8888-999999999999"]
	if !ok {
		t.Errorf("expected 55555555-6666-7777-8888-999999999999 to be active")
	} else {
		if info5.Agent != "claude" {
			t.Errorf("expected agent 'claude', got %q", info5.Agent)
		}
		if info5.Profile != "work" {
			t.Errorf("expected profile 'work', got %q", info5.Profile)
		}
		if info5.PID != 66778 {
			t.Errorf("expected PID 66778, got %d", info5.PID)
		}
	}
}
