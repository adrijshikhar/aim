package agents_test

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/aim-cli/aim/internal/agents"
)

func TestBaseAdapter_Metadata(t *testing.T) {
	b := agents.NewBaseAdapter("test-agent", "Test Agent", "test-bin", []string{"ta", "test"}, "/custom/bin")

	if b.Name() != "test-agent" {
		t.Fatalf("expected Name() = 'test-agent', got %q", b.Name())
	}
	if b.DisplayName() != "Test Agent" {
		t.Fatalf("expected DisplayName() = 'Test Agent', got %q", b.DisplayName())
	}
	if b.BinaryName() != "test-bin" {
		t.Fatalf("expected BinaryName() = 'test-bin', got %q", b.BinaryName())
	}
	if !slices.Equal(b.Aliases(), []string{"ta", "test"}) {
		t.Fatalf("expected Aliases() = [ta, test], got %v", b.Aliases())
	}
}

func TestBaseAdapter_ResolveBinary(t *testing.T) {
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "custom-bin")
	if err := os.WriteFile(binPath, []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatalf("failed to write test binary: %v", err)
	}

	b := agents.NewBaseAdapter("custom", "Custom", "custom-bin", nil, binPath)
	resolved := b.ResolveBinary()
	if resolved != binPath {
		t.Fatalf("expected ResolveBinary() = %q, got %q", binPath, resolved)
	}

	// Non-existent binary returns binaryName
	bMissing := agents.NewBaseAdapter("missing", "Missing", "non-existent-binary-12345", nil)
	if bMissing.ResolveBinary() != "non-existent-binary-12345" {
		t.Fatalf("expected ResolveBinary() = 'non-existent-binary-12345', got %q", bMissing.ResolveBinary())
	}
}

func TestBaseAdapter_BaseLaunchEnv(t *testing.T) {
	b := agents.NewBaseAdapter("claude", "Claude Code", "claude", nil)
	envMap := b.BaseLaunchEnv("work", "/profiles/work", map[string]string{
		"EXTRA_VAR": "val123",
	})

	if envMap["HOME"] != "/profiles/work" {
		t.Fatalf("expected HOME = '/profiles/work', got %q", envMap["HOME"])
	}
	if envMap["AIM_AGENT"] != "claude" {
		t.Fatalf("expected AIM_AGENT = 'claude', got %q", envMap["AIM_AGENT"])
	}
	if envMap["AIM_PROFILE"] != "work" {
		t.Fatalf("expected AIM_PROFILE = 'work', got %q", envMap["AIM_PROFILE"])
	}
	if envMap["EXTRA_VAR"] != "val123" {
		t.Fatalf("expected EXTRA_VAR = 'val123', got %q", envMap["EXTRA_VAR"])
	}
}

func TestBaseAdapter_BuildCleanEnv(t *testing.T) {
	b := agents.NewBaseAdapter("codex", "Codex CLI", "codex", nil)
	envSlice := b.BuildCleanEnv("dev", "/profiles/dev", map[string]string{
		"CUSTOM_KEY": "custom_val",
	})

	var foundHome, foundAgent, foundProfile, foundCustom bool
	for _, kv := range envSlice {
		if kv == "HOME=/profiles/dev" {
			foundHome = true
		}
		if kv == "AIM_AGENT=codex" {
			foundAgent = true
		}
		if kv == "AIM_PROFILE=dev" {
			foundProfile = true
		}
		if kv == "CUSTOM_KEY=custom_val" {
			foundCustom = true
		}
		if strings.HasPrefix(kv, "SSH_CONNECTION=") {
			t.Fatalf("SSH_CONNECTION leaked into clean env: %s", kv)
		}
	}

	if !foundHome || !foundAgent || !foundProfile || !foundCustom {
		t.Fatalf("clean env missing expected vars: home=%v, agent=%v, profile=%v, custom=%v",
			foundHome, foundAgent, foundProfile, foundCustom)
	}
}
