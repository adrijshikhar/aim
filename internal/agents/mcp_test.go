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

func TestNormalizeAuth(t *testing.T) {
	cases := map[string]string{
		"o_auth":        "OAuth",
		"OAuth":         "OAuth",
		"connected":     "OAuth",
		"not_logged_in": "auth required",
		"auth_required": "auth required",
		"unsupported":   "unsupported",
		"none":          "unsupported",
		"":              "unsupported",
		"custom":        "custom",
	}
	for in, want := range cases {
		if got := NormalizeAuth(in); got != want {
			t.Errorf("NormalizeAuth(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCollapseCommand(t *testing.T) {
	cases := []struct {
		cmd  string
		args []string
		want string
	}{
		{
			cmd:  "npx",
			args: []string{"@playwright/mcp@latest"},
			want: "npx @playwright/mcp@latest",
		},
		{
			cmd:  "sh",
			args: []string{"-c", `url="${LOCAL_GRAFANA_URL:-http://localhost:3000}"; exec env GRAFANA_URL="$url" uvx mcp-grafana`},
			want: "uvx mcp-grafana",
		},
		{
			cmd:  "sh",
			args: []string{"-c", `key="${CORALOGIX_API_KEY:?required}"; exec npx -y mcp-remote@latest https://api.eu1.coralogix.com/mgmt/api/v1/mcp --header "Authorization: Bearer $key"`},
			want: "npx (api.eu1.coralogix.com)",
		},
		{
			cmd:  "sh",
			args: []string{"-c", `url="${SKRULL_BASE_URL:-https://skrull.me}"; exec env SKRULL_BASE_URL="$url" uvx --from 'skrull[mcp] @ git+https://github.com/hevoio/skrull@dev' skrull-mcp`},
			want: "uvx skrull-mcp",
		},
		{
			cmd:  "node",
			args: []string{"-e", `const f=require('fs'); ... if(f.existsSync(p.join(r,'scripts','mcp-server.cjs'))) ...`},
			want: "node .../mcp-server.cjs",
		},
		{
			cmd:  "/Applications/ChatGPT.app/Contents/Resources/cua_node/bin/node_repl",
			args: nil,
			want: "node_repl",
		},
		{
			cmd:  "/Users/nemesis/.caveman/bin/caveman-mcp",
			args: nil,
			want: "caveman-mcp",
		},
		{
			cmd:  "python",
			args: []string{"-c", `import hevo_mcp.server; server.main()`},
			want: "python -m hevo_mcp.server",
		},
	}

	for _, c := range cases {
		got := CollapseCommand(c.cmd, c.args)
		if got != c.want {
			t.Errorf("CollapseCommand(%q, %v) = %q, want %q", c.cmd, c.args, got, c.want)
		}
	}
}
