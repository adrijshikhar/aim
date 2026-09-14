package catalyst_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/session"
	"github.com/aim-cli/aim/internal/session/catalyst"
)

func TestBridge_WriteAndReadHandoffBrief(t *testing.T) {
	tmpDir := t.TempDir()
	branch := "feat/awesome-feature"

	b := catalyst.NewBridge()
	s := session.NewSession("775e6ada-1595-4e7e-84fa-ce0ea71e3007", "Build Auth Module", "agy", "work", false, time.Now())
	s.Summary = "Initial tokens and middleware configured."

	briefPath, err := b.WriteHandoffBrief(tmpDir, branch, &s)
	if err != nil {
		t.Fatalf("WriteHandoffBrief failed: %v", err)
	}

	expectedKey := "feat-awesome-feature.json"
	expectedPath := filepath.Join(tmpDir, ".catalyst", "handoffs", expectedKey)
	if briefPath != expectedPath {
		t.Errorf("expected brief path %s, got %s", expectedPath, briefPath)
	}

	// Verify file exists and has valid JSON
	data, err := os.ReadFile(briefPath)
	if err != nil {
		t.Fatalf("failed to read brief: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("invalid json: %v", err)
	}

	if raw["schema_version"] != float64(1) {
		t.Errorf("expected schema_version 1, got %v", raw["schema_version"])
	}
	if raw["goal"] != "Continue session: Build Auth Module" {
		t.Errorf("expected goal 'Continue session: Build Auth Module', got %v", raw["goal"])
	}

	// Verify ReadHandoffBrief
	readBrief, err := b.ReadHandoffBrief(tmpDir, branch)
	if err != nil {
		t.Fatalf("ReadHandoffBrief failed: %v", err)
	}
	if readBrief.Branch != branch {
		t.Errorf("expected branch %s, got %s", branch, readBrief.Branch)
	}
	if readBrief.Notes != s.Summary {
		t.Errorf("expected notes %q, got %q", s.Summary, readBrief.Notes)
	}
}
