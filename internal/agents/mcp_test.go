package agents

import "testing"

func TestBackgroundMatch(t *testing.T) {
	flags := []string{"--bg", "--background"}
	first := []string{"agents", "attach", "respawn"}
	cases := []struct {
		args []string
		want bool
	}{
		{[]string{"--bg"}, true},
		{[]string{"--model", "opus", "--background"}, true},
		{[]string{"agents"}, true},
		{[]string{"attach", "id"}, true},
		{[]string{"--resume", "x"}, false},
		{[]string{"-p", "agents"}, false},      // a prompt that happens to say "agents"
		{[]string{"fix", "--", "--bg"}, false}, // after "--" belongs to the agent
		{nil, false},
	}
	for _, c := range cases {
		if got := BackgroundMatch(c.args, flags, first); got != c.want {
			t.Fatalf("%v: got %v want %v", c.args, got, c.want)
		}
	}
}

func TestIsSession(t *testing.T) {
	cases := []struct {
		args []string
		want bool
	}{
		{nil, true},
		{[]string{"--resume", "x"}, true},
		{[]string{"mcp", "list"}, true}, // shows the merged set, so it merges
		{[]string{"--version"}, false},
		{[]string{"-v"}, false},
		{[]string{"--help"}, false},
		{[]string{"-h"}, false},
		{[]string{"--model", "opus", "--version"}, false}, // profile args precede the CLI's
		{[]string{"help"}, false},
		{[]string{"version"}, false},
		{[]string{"-p", "version"}, true},       // a prompt that happens to say "version"
		{[]string{"fix", "--", "--help"}, true}, // after "--" belongs to the agent
		{[]string{"--", "-v"}, true},
	}
	for _, c := range cases {
		if got := IsSession(c.args); got != c.want {
			t.Fatalf("%v: got %v want %v", c.args, got, c.want)
		}
	}
}
