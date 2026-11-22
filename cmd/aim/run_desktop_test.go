package main

import (
	"testing"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/agents/claudedesktop"
	"github.com/aim-cli/aim/internal/agents/codexdesktop"
	"github.com/aim-cli/aim/internal/agents/cursor"
)

func TestRun_DesktopAppDetection(t *testing.T) {
	reg := agents.NewRegistry()
	reg.Register(claudedesktop.NewAdapter())
	reg.Register(codexdesktop.NewAdapter())
	reg.Register(cursor.NewAdapter())

	for _, name := range []string{"claude-desktop", "codex-desktop", "cursor"} {
		a, err := reg.Get(name)
		if err != nil {
			t.Fatalf("unexpected error getting %s: %v", name, err)
		}
		desktopApp, ok := a.(agents.DesktopApp)
		if !ok || !desktopApp.IsDesktop() {
			t.Fatalf("expected %s to be detected as DesktopApp", name)
		}
		if desktopApp.AppBundleName() == "" {
			t.Errorf("expected non-empty AppBundleName for %s", name)
		}
	}
}

func TestRun_HasForceFlag(t *testing.T) {
	if !hasForceFlag([]string{"--force"}) {
		t.Error("expected true for --force")
	}
	if !hasForceFlag([]string{"-f"}) {
		t.Error("expected true for -f")
	}
	if !hasForceFlag([]string{"run", "something", "--force", "arg"}) {
		t.Error("expected true for mixed flags with --force")
	}
	if hasForceFlag([]string{"--foo", "bar"}) {
		t.Error("expected false when --force is absent")
	}
}
