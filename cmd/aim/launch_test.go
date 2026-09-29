package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/merge"
	"github.com/aim-cli/aim/internal/profile"
)

// collectionMock is a mockAdapter with an MCP collection.
type collectionMock struct {
	mockAdapter
	cols    []merge.Collection
	bg      bool
	bgWords []string // background first words, matched like a real adapter
}

func (m *collectionMock) MCPCollections(profileDir, realHome string) []merge.Collection {
	return m.cols
}
func (m *collectionMock) IsBackground(profileArgs, args []string) bool {
	return m.bg || agents.ArgsMatch(profileArgs, args, nil, m.bgWords)
}
func (m *collectionMock) IsSession(profileArgs, args []string) bool { return true }

// isolate points every aim and agent path at temp dirs (this machine has a
// legacy ~/.aim that config.StateDir() would otherwise resolve to).
func isolate(t *testing.T) string {
	t.Helper()
	d := t.TempDir()
	t.Setenv("AIM_HOME", d)
	t.Setenv("AIM_REAL_HOME", filepath.Join(d, "realhome"))
	t.Setenv("AIM_MOCK_KEYCHAIN", "1")
	return d
}

func mockWithFiles(t *testing.T, d string) (*collectionMock, string, string) {
	t.Helper()
	host := filepath.Join(d, "host.json")
	prof := filepath.Join(d, "prof.json")
	_ = os.WriteFile(host, []byte(`{"mcpServers":{"jev":{"command":"npx"}}}`), 0o600)
	_ = os.WriteFile(prof, []byte(`{"mcpServers":{}}`), 0o600)
	return &collectionMock{mockAdapter: mockAdapter{name: "mock"}, cols: []merge.Collection{{Agent: "mock", Name: "mcpServers",
		Format: merge.JSON, Key: "mcpServers", HostPath: host, ProfilePath: prof, Group: "mcp"}}}, host, prof
}

func TestSplitRunFlags_OnlyBeforeDoubleDash(t *testing.T) {
	auto, rest := splitRunFlags([]string{"claude", "work", "-y", "--", "npx", "-y", "pkg"})
	if !auto {
		t.Fatal("-y before -- must set autoCreate")
	}
	if want := []string{"claude", "work", "--", "npx", "-y", "pkg"}; !reflect.DeepEqual(rest, want) {
		t.Fatalf("rest = %v", rest)
	}
	if auto, _ := splitRunFlags([]string{"claude", "work", "--", "-y"}); auto {
		t.Fatal("-y after -- belongs to the agent")
	}
}

func TestWithSessionMerge_MergesDuringRunAndStripsAfter(t *testing.T) {
	d := isolate(t)
	ad, _, prof := mockWithFiles(t, d)
	store := merge.Store{Dir: filepath.Join(d, "state")}
	var during []byte
	code := withSessionMerge(ad, store, "work", d, nil, nil, func() int {
		during, _ = os.ReadFile(prof)
		return 0
	})
	if code != 0 {
		t.Fatal(code)
	}
	if !strings.Contains(string(during), "jev") {
		t.Fatalf("host server missing during run: %s", during)
	}
	if after, _ := os.ReadFile(prof); string(after) != `{"mcpServers":{}}` {
		t.Fatalf("not restored: %s", after)
	}
	if _, err := os.Stat(filepath.Join(store.Dir, "work.json")); err != nil {
		t.Fatalf("state must be written to the injected store: %v", err)
	}
}

func TestWithSessionMerge_NonTTYKeepsAndPrints(t *testing.T) {
	d := isolate(t)
	ad, _, prof := mockWithFiles(t, d)
	orig := stdinIsTerminal
	stdinIsTerminal = func() bool { return false }
	t.Cleanup(func() { stdinIsTerminal = orig })
	store := merge.Store{Dir: filepath.Join(d, "state")}
	_, out := captureOutput(t, func() { // stderr: the prompt writes there
		withSessionMerge(ad, store, "work", d, nil, nil, func() int {
			_ = os.WriteFile(prof, []byte(`{"mcpServers":{"jev":{"command":"npx"},"foo":{"command":"f"}}}`), 0o600)
			return 0
		})
	})
	if !strings.Contains(out, "1 change(s) kept in work (mock)") {
		t.Fatalf("output = %q", out)
	}
	if b, _ := os.ReadFile(prof); !strings.Contains(string(b), `"foo"`) || strings.Contains(string(b), `"jev"`) {
		t.Fatalf("at rest = %s", b)
	}
}

func TestWithSessionMerge_BackgroundKeepsMerge(t *testing.T) {
	d := isolate(t)
	ad, _, prof := mockWithFiles(t, d)
	ad.bg = true
	store := merge.Store{Dir: filepath.Join(d, "state")}
	withSessionMerge(ad, store, "work", d, nil, []string{"--bg"}, func() int { return 0 })
	if b, _ := os.ReadFile(prof); !strings.Contains(string(b), `"jev"`) {
		t.Fatalf("a background launch keeps the merge in place: %s", b)
	}
	st, _ := store.Load("work")
	if st.Active["mock/mcpServers"]["jev"] == "" {
		t.Fatal("state must remember the merge so the next launch recovers")
	}
}

// A background first word on the CLI still counts when profiles.<p>.args
// (here flags) precede it in the launched command: `codex --model o3
// app-server` must keep the merge, not strip it under the running daemon.
func TestWithSessionMerge_BackgroundFirstWordAfterProfileArgs(t *testing.T) {
	d := isolate(t)
	ad, _, prof := mockWithFiles(t, d)
	ad.bgWords = []string{"app-server"}
	cfg := &config.Config{Profiles: map[string]config.ProfileConfig{"work": {Args: []string{"--model", "o3"}}}}
	store := merge.Store{Dir: filepath.Join(d, "state")}
	withSessionMerge(ad, store, "work", d, cfg, []string{"app-server"}, func() int { return 0 })
	if b, _ := os.ReadFile(prof); !strings.Contains(string(b), `"jev"`) {
		t.Fatalf("a background launch after profile args keeps the merge in place: %s", b)
	}
}

func TestWithSessionMerge_DisabledMergesNothing(t *testing.T) {
	d := isolate(t)
	ad, _, prof := mockWithFiles(t, d)
	off := false
	cfg := &config.Config{Profiles: map[string]config.ProfileConfig{"work": {MCPGlobal: &off}}}
	store := merge.Store{Dir: filepath.Join(d, "state")}
	var during []byte
	withSessionMerge(ad, store, "work", d, cfg, nil, func() int { during, _ = os.ReadFile(prof); return 0 })
	if strings.Contains(string(during), "jev") {
		t.Fatalf("mcp_global=false must not merge: %s", during)
	}
}

func TestExecuteRun_WrapsSessionWithMerge(t *testing.T) {
	d := isolate(t)
	host := filepath.Join(d, "host.json")
	seen := filepath.Join(d, "seen.json")
	_ = os.WriteFile(host, []byte(`{"mcpServers":{"jev":{"command":"npx"}}}`), 0o600)
	pm := profile.NewProfileManager(d)
	pDir, _ := pm.EnsureProfile("work")
	prof := filepath.Join(pDir, "mcp.json")
	_ = os.WriteFile(prof, []byte(`{"mcpServers":{}}`), 0o600)
	bin := filepath.Join(d, "agent.sh")
	_ = os.WriteFile(bin, []byte("#!/bin/sh\ncat \""+prof+"\" > \""+seen+"\"\n"), 0o755)
	reg := agents.NewRegistry()
	reg.Register(&collectionMock{mockAdapter: mockAdapter{name: "mock", binaryPath: bin}, cols: []merge.Collection{{Agent: "mock",
		Name: "mcpServers", Format: merge.JSON, Key: "mcpServers", HostPath: host, ProfilePath: prof, Group: "mcp"}}})
	if code := executeRun(reg, pm, "mock", "work", nil); code != 0 {
		t.Fatalf("exit %d", code)
	}
	if b, _ := os.ReadFile(seen); !strings.Contains(string(b), "jev") {
		t.Fatalf("the agent must see the host server: %s", b)
	}
	if b, _ := os.ReadFile(prof); string(b) != `{"mcpServers":{}}` {
		t.Fatalf("not restored: %s", b)
	}
	if _, err := os.Stat(filepath.Join(d, "profile-merge", "work.json")); err != nil {
		t.Fatal("state must live under AIM_HOME/profile-merge")
	}
}

// groupMock declares an MCP and a plugins collection in one settings file.
func groupMock(t *testing.T, d string) (*collectionMock, string) {
	t.Helper()
	host := filepath.Join(d, "host.json")
	prof := filepath.Join(d, "prof.json")
	_ = os.WriteFile(host, []byte(`{"mcpServers":{"jev":{"command":"npx"}},"enabledPlugins":{"x@m":true}}`), 0o600)
	_ = os.WriteFile(prof, []byte(`{"mcpServers":{}}`), 0o600)
	col := func(name, group string) merge.Collection {
		return merge.Collection{Agent: "mock", Name: name, Format: merge.JSON, Key: name, HostPath: host, ProfilePath: prof, Group: group}
	}
	return &collectionMock{mockAdapter: mockAdapter{name: "mock"},
		cols: []merge.Collection{col("mcpServers", "mcp"), col("enabledPlugins", "plugins")}}, prof
}

func TestWithSessionMerge_PluginsGlobalOffRecoversAfterBackground(t *testing.T) {
	d := isolate(t)
	ad, prof := groupMock(t, d)
	store := merge.Store{Dir: filepath.Join(d, "state")}
	ad.bg = true
	withSessionMerge(ad, store, "work", d, nil, []string{"--bg"}, func() int { return 0 })
	if b, _ := os.ReadFile(prof); !strings.Contains(string(b), `"x@m"`) {
		t.Fatalf("the background launch merges plugins: %s", b)
	}
	ad.bg = false
	off := false
	cfg := &config.Config{Profiles: map[string]config.ProfileConfig{"work": {PluginsGlobal: &off}}}
	var during []byte
	withSessionMerge(ad, store, "work", d, cfg, nil, func() int { during, _ = os.ReadFile(prof); return 0 })
	if strings.Contains(string(during), "enabledPlugins") || !strings.Contains(string(during), `"jev"`) {
		t.Fatalf("plugins_global=false must recover the plugins and merge only servers: %s", during)
	}
	if after, _ := os.ReadFile(prof); strings.Contains(string(after), "jev") || strings.Contains(string(after), "enabledPlugins") {
		t.Fatalf("at rest = %s", after)
	}
}

func TestWithSessionMerge_MCPGlobalOffStillMergesPlugins(t *testing.T) {
	d := isolate(t)
	ad, prof := groupMock(t, d)
	off := false
	cfg := &config.Config{Profiles: map[string]config.ProfileConfig{"work": {MCPGlobal: &off}}}
	store := merge.Store{Dir: filepath.Join(d, "state")}
	var during []byte
	withSessionMerge(ad, store, "work", d, cfg, nil, func() int { during, _ = os.ReadFile(prof); return 0 })
	if strings.Contains(string(during), "jev") || !strings.Contains(string(during), `"x@m"`) {
		t.Fatalf("mcp_global=false switches off servers only: %s", during)
	}
}

// GroupEnabled treats an unknown group as always on, so every collection an
// adapter merges must name a group a profile setting switches, and a non-MCP
// one its own Noun. No adapter name may hold a dot (Store.ActiveSessions).
func TestAdapters_CollectionsDeclareKnownGroup(t *testing.T) {
	d := t.TempDir()
	n := 0
	for _, a := range defaultRegistry().All() {
		if strings.Contains(a.Name(), ".") {
			t.Errorf("adapter name %q holds a dot: Store.ActiveSessions skips a dotted agent in <profile>.<agent>.sessions, so the running-session refusal would miss it", a.Name())
		}
		mp, ok := a.(agents.MCPProvider)
		if !ok {
			continue
		}
		for _, c := range mp.MCPCollections(filepath.Join(d, "profile"), filepath.Join(d, "home")) {
			n++
			if c.Group != "mcp" && c.Group != "plugins" {
				t.Errorf("%s: %s declares group %q, want mcp or plugins", a.Name(), c.ID(), c.Group)
			}
			// An empty Noun reads "server" in messages, right only for MCP.
			if c.Group != "mcp" && c.Noun == "" {
				t.Errorf("%s: %s must declare a Noun", a.Name(), c.ID())
			}
		}
	}
	if n == 0 {
		t.Fatal("no adapter declared a collection")
	}
}
