package main

import (
	"bytes"
	"testing"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/profile"
)

func TestShareCommandHelp(t *testing.T) {
	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(t.TempDir())
	cmd := newShareCmd(reg, pm)

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("Share AIM profiles across teammates")) {
		t.Errorf("expected share help text, got %s", buf.String())
	}
}

func TestShareSubcommandsExist(t *testing.T) {
	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(t.TempDir())
	cmd := newShareCmd(reg, pm)

	subcommands := []string{"enable", "disable", "status", "list", "request", "release", "peer", "serve"}
	for _, sub := range subcommands {
		found := false
		for _, c := range cmd.Commands() {
			if c.Name() == sub {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected subcommand %q to exist", sub)
		}
	}
}
