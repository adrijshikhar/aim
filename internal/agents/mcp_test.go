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
