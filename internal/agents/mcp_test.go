package agents

import "testing"

func TestArgsMatch(t *testing.T) {
	flags := []string{"--bg", "--background"}
	first := []string{"agents", "attach", "respawn"}
	opus := []string{"--model", "opus"}
	cases := []struct {
		profile, args []string
		want          bool
	}{
		{nil, []string{"--bg"}, true},
		{nil, []string{"--model", "opus", "--background"}, true},
		{nil, []string{"agents"}, true},
		{nil, []string{"attach", "id"}, true},
		{nil, []string{"--resume", "x"}, false},
		{nil, []string{"-p", "agents"}, false},      // a prompt that happens to say "agents"
		{nil, []string{"fix", "--", "--bg"}, false}, // after "--" belongs to the agent
		{nil, nil, false},
		// profiles.<p>.args precede the CLI args in the launched command
		{opus, []string{"agents"}, true}, // the CLI's first word, after profile flags
		{opus, []string{"--bg"}, true},
		{[]string{"--bg"}, nil, true},
		{[]string{"agents"}, nil, true},                  // the profile's first word
		{[]string{"agents"}, []string{"--resume"}, true}, // the profile's first word leads
		{opus, nil, false},
		{opus, []string{"fix", "agents"}, false},         // not a first word of either
		{[]string{"--", "x"}, []string{"--bg"}, false},   // the profile's "--" ends the scan
		{[]string{"--", "x"}, []string{"agents"}, false}, // …and makes the CLI args the agent's
	}
	for _, c := range cases {
		if got := ArgsMatch(c.profile, c.args, flags, first); got != c.want {
			t.Errorf("%v + %v: got %v want %v", c.profile, c.args, got, c.want)
		}
	}
}
