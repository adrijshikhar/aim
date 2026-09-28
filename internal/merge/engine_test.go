package merge

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type fixture struct {
	t        *testing.T
	host, pf string
	col      Collection
	eng      *Engine
	out      *bytes.Buffer
}

func newFixture(t *testing.T) *fixture {
	d := t.TempDir()
	f := &fixture{t: t, host: filepath.Join(d, "host.json"), pf: filepath.Join(d, "profile.json"), out: &bytes.Buffer{}}
	f.col = Collection{Agent: "claude", Name: "mcpServers", Format: JSON, Key: "mcpServers", HostPath: f.host, ProfilePath: f.pf}
	f.eng = &Engine{Store: Store{Dir: filepath.Join(d, "state")}, Out: f.out, Now: func() time.Time { return time.Unix(0, 0) }}
	return f
}

func (f *fixture) files(host, profile string) {
	_ = os.WriteFile(f.host, []byte(host), 0o600)
	_ = os.WriteFile(f.pf, []byte(profile), 0o600)
}

func (f *fixture) start() *Session {
	f.t.Helper()
	s, err := f.eng.Start("work", "claude", []Collection{f.col}, StartOptions{Enabled: allOn})
	if err != nil {
		f.t.Fatal(err)
	}
	return s
}

func (f *fixture) names(path string) []string {
	e, _, err := f.col.Read(path)
	if err != nil {
		f.t.Fatal(err)
	}
	return e.Order
}

func (f *fixture) profileNames() []string { return f.names(f.pf) }

func (f *fixture) hostValue(name string) string {
	e, _, _ := f.col.Read(f.host)
	if !e.Has(name) {
		return ""
	}
	return Hash(f.col.norm(), e.Values[name])
}

func keepAll(c []Change) []Decision { return make([]Decision, len(c)) }

func allOn(Collection) bool  { return true }
func allOff(Collection) bool { return false }

// sessionsLockFree fails the test when a session still holds the shared
// sessions lock for (profile, agent).
func sessionsLockFree(t *testing.T, eng *Engine, profile, agent string) {
	t.Helper()
	probe, ok, err := TryUpgrade(eng.Store.SessionsLockPath(profile, agent))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("the sessions lock for %s/%s is still held", profile, agent)
	}
	_ = probe.Unlock()
}

func decideBy(m map[string]Decision) func([]Change) []Decision {
	return func(c []Change) []Decision {
		d := make([]Decision, len(c))
		for i, x := range c {
			d[i] = m[x.Name]
		}
		return d
	}
}

func kinds(ch []Change) string {
	var got []string
	for _, c := range ch {
		got = append(got, c.Name+":"+string(c.Kind))
	}
	return strings.Join(got, ",")
}

// sameAtRest checks the engine's guarantee: the collection holds the same items
// (by normalised hash) and every other top-level member is byte-identical.
func sameAtRest(t *testing.T, c Collection, before, after []byte) {
	t.Helper()
	bm, err := scanTop(before)
	if err != nil {
		t.Fatal(err)
	}
	am, err := scanTop(after)
	if err != nil {
		t.Fatal(err)
	}
	other := func(data []byte, ms []member) map[string]string {
		o := map[string]string{}
		for _, m := range ms {
			if m.key != c.Key {
				o[m.key] = string(data[m.valStart:m.end])
			}
		}
		return o
	}
	if fmt.Sprint(other(before, bm)) != fmt.Sprint(other(after, am)) {
		t.Fatalf("other keys changed:\n%s\n%s", before, after)
	}
	be, _ := decodeCollection(before, c.Key)
	ae, _ := decodeCollection(after, c.Key)
	if len(be.Order) != len(ae.Order) {
		t.Fatalf("items differ: %v vs %v", be.Order, ae.Order)
	}
	for _, n := range be.Order {
		if Hash(c.norm(), be.Values[n]) != Hash(c.norm(), ae.Values[n]) {
			t.Fatalf("item %s differs", n)
		}
	}
}

func decodeCollection(data []byte, key string) (Entries, error) {
	ms, err := scanTop(data)
	if err != nil {
		return Entries{}, err
	}
	for _, m := range ms {
		if m.key == key {
			return decodeObjectEntries(data[m.valStart:m.end])
		}
	}
	return NewEntries(), nil
}

func TestEngine_NoChangeRestoresItemsAndOtherKeys(t *testing.T) {
	f := newFixture(t)
	f.files(`{"mcpServers":{"jev":{"command":"npx"}}}`, `{"x":1,"mcpServers":{"mine":{"command":"a"}}}`)
	before, _ := os.ReadFile(f.pf)
	s := f.start()
	if got := f.profileNames(); len(got) != 2 {
		t.Fatalf("during session = %v", got)
	}
	ch, _ := s.Diff()
	if len(ch) != 0 {
		t.Fatalf("changes = %+v", ch)
	}
	if err := s.Finish(ch, keepAll); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(f.pf)
	sameAtRest(t, f.col, before, after)
}

func TestEngine_KeyCreatedByMergeIsRemovedAgain(t *testing.T) {
	f := newFixture(t)
	f.files(`{"mcpServers":{"jev":{"command":"npx"}}}`, `{"x": 1}`)
	s := f.start()
	ch, _ := s.Diff()
	if err := s.Finish(ch, keepAll); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(f.pf); string(b) != `{"x": 1}` {
		t.Fatalf("at rest = %s", b)
	}
}

func TestEngine_AgentNoiseIsNotAChange(t *testing.T) {
	f := newFixture(t)
	f.col.Normalise = func(v map[string]any) map[string]any {
		v = DropEmpty(v)
		if v["type"] == "stdio" {
			delete(v, "type")
		}
		return v
	}
	f.files(`{"mcpServers":{"jev":{"command":"npx"}}}`, `{"mcpServers":{}}`)
	s := f.start()
	_ = os.WriteFile(f.pf, []byte(`{"mcpServers":{"jev":{"type":"stdio","command":"npx","env":{}}}}`), 0o600)
	ch, _ := s.Diff()
	if len(ch) != 0 {
		t.Fatalf("noise reported as change: %+v", ch)
	}
	if err := s.Finish(ch, keepAll); err != nil {
		t.Fatal(err)
	}
}

func TestEngine_AddedEditedRemoved_KeepAndPromote(t *testing.T) {
	f := newFixture(t)
	f.files(`{"mcpServers":{"jev":{"command":"npx"},"ctx":{"command":"c"}}}`, `{"mcpServers":{}}`)
	s := f.start()
	_ = os.WriteFile(f.pf, []byte(`{"mcpServers":{"jev":{"command":"npx2"},"foo":{"command":"f"}}}`), 0o600)
	ch, _ := s.Diff()
	if kinds(ch) != "ctx:removed,foo:added,jev:edited" {
		t.Fatalf("changes = %v", kinds(ch))
	}
	for _, c := range ch {
		if c.HostChanged {
			t.Fatalf("host did not change: %+v", c)
		}
	}
	// promote foo; keep jev (the profile's own value now) and ctx (comes back next session)
	if err := s.Finish(ch, decideBy(map[string]Decision{"foo": Promote})); err != nil {
		t.Fatal(err)
	}
	if f.hostValue("foo") != Hash(f.col.norm(), map[string]any{"command": "f"}) {
		t.Fatal("host must hold the promoted foo")
	}
	if names := f.profileNames(); strings.Join(names, ",") != "jev" {
		t.Fatalf("at rest = %v (want only jev: foo is a host item now, ctx was removed)", names)
	}
	// next session: ctx comes back, jev stays the profile's, foo comes from the host
	s2 := f.start()
	if names := f.profileNames(); strings.Join(names, ",") != "jev,ctx,foo" {
		t.Fatalf("second session sees %v", names)
	}
	ch2, _ := s2.Diff()
	if len(ch2) != 0 {
		t.Fatalf("second session must start clean: %+v", ch2)
	}
	_ = s2.Finish(ch2, keepAll)
	if names := f.profileNames(); strings.Join(names, ",") != "jev" {
		t.Fatalf("at rest after second session = %v", names)
	}
}

func TestEngine_PromoteEditedReplacesHostValue(t *testing.T) {
	f := newFixture(t)
	f.files(`{"mcpServers":{"jev":{"command":"npx"}}}`, `{"mcpServers":{}}`)
	s := f.start()
	_ = os.WriteFile(f.pf, []byte(`{"mcpServers":{"jev":{"command":"npx2"}}}`), 0o600)
	ch, _ := s.Diff()
	if err := s.Finish(ch, decideBy(map[string]Decision{"jev": Promote})); err != nil {
		t.Fatal(err)
	}
	if f.hostValue("jev") != Hash(f.col.norm(), map[string]any{"command": "npx2"}) {
		t.Fatal("host must hold exactly the promoted value")
	}
	if names := f.profileNames(); len(names) != 0 {
		t.Fatalf("a promoted edit is a host item and leaves the profile at rest: %v", names)
	}
}

func TestEngine_PromoteRemovedRemovesFromHost(t *testing.T) {
	f := newFixture(t)
	f.files(`{"mcpServers":{"ctx":{"command":"c"}}}`, `{"mcpServers":{}}`)
	s := f.start()
	_ = os.WriteFile(f.pf, []byte(`{"mcpServers":{}}`), 0o600)
	ch, _ := s.Diff()
	if err := s.Finish(ch, decideBy(map[string]Decision{"ctx": Promote})); err != nil {
		t.Fatal(err)
	}
	if f.hostValue("ctx") != "" {
		t.Fatal("ctx must be gone from the host")
	}
}

func TestEngine_PromoteHostChangedIsKept(t *testing.T) {
	f := newFixture(t)
	f.files(`{"mcpServers":{"jev":{"command":"npx"}}}`, `{"mcpServers":{}}`)
	s := f.start()
	_ = os.WriteFile(f.pf, []byte(`{"mcpServers":{"jev":{"command":"npx2"}}}`), 0o600)
	_ = os.WriteFile(f.host, []byte(`{"mcpServers":{"jev":{"command":"npx3"}}}`), 0o600) // host changed meanwhile
	ch, _ := s.Diff()
	if len(ch) != 1 || !ch[0].HostChanged {
		t.Fatalf("diff must flag the host change: %+v", ch)
	}
	if err := s.Finish(ch, decideBy(map[string]Decision{"jev": Promote})); err != nil {
		t.Fatal(err)
	}
	if f.hostValue("jev") != Hash(f.col.norm(), map[string]any{"command": "npx3"}) {
		t.Fatal("promote must never overwrite a host value that changed during the session")
	}
	if names := f.profileNames(); strings.Join(names, ",") != "jev" {
		t.Fatalf("the edit stays the profile's: %v", names)
	}
	if !strings.Contains(f.out.String(), "host changed") {
		t.Fatalf("user must be told: %q", f.out.String())
	}
}

func TestEngine_ProfileNameWinsEvenWhenIdentical(t *testing.T) {
	f := newFixture(t)
	f.files(`{"mcpServers":{"jev":{"command":"npx"}}}`, `{"mcpServers":{"jev":{"command":"mine"}}}`)
	s := f.start() // migration: jev differs, so it stays the profile's
	if got := f.profileNames(); strings.Join(got, ",") != "jev" {
		t.Fatalf("the profile's jev wins; the host's is not merged: %v", got)
	}
	ch, _ := s.Diff()
	if len(ch) != 0 {
		t.Fatalf("changes = %+v", ch)
	}
	_ = s.Finish(ch, keepAll)
	// after migration, even an identical copy is the profile's own
	_ = os.WriteFile(f.pf, []byte(`{"mcpServers":{"jev":{"command":"npx"}}}`), 0o600)
	s2 := f.start()
	if got := f.profileNames(); strings.Join(got, ",") != "jev" {
		t.Fatalf("identical name still wins: %v", got)
	}
	ch2, _ := s2.Diff()
	if len(ch2) != 0 {
		t.Fatalf("a profile-defined name is never a change: %+v", ch2)
	}
	_ = s2.Finish(ch2, keepAll)
	if got := f.profileNames(); strings.Join(got, ",") != "jev" {
		t.Fatalf("never stripped: %v", got)
	}
}

func TestEngine_StripSkipsMissingFile(t *testing.T) {
	f := newFixture(t)
	f.files(`{"mcpServers":{"jev":{"command":"npx"}}}`, `{"mcpServers":{}}`)
	s := f.start()
	_ = os.Remove(f.pf) // the agent deleted its config during the session
	ch, _ := s.Diff()
	_ = s.Finish(ch, keepAll)
	if _, err := os.Stat(f.pf); !os.IsNotExist(err) {
		t.Fatal("strip must not recreate a file the agent deleted")
	}
	st, _ := f.eng.Store.Load("work")
	if len(st.Active) != 0 || len(st.StartProfile) != 0 {
		t.Fatalf("state must be cleared: %+v", st)
	}
}

func TestEngine_PromoteValueMovedSincePromptIsKept(t *testing.T) {
	f := newFixture(t)
	f.files(`{"mcpServers":{}}`, `{"mcpServers":{}}`)
	s := f.start()
	_ = os.WriteFile(f.pf, []byte(`{"mcpServers":{"foo":{"command":"f"}}}`), 0o600)
	ch, _ := s.Diff() // the user is shown foo = f
	// the value moves before apply
	_ = os.WriteFile(f.pf, []byte(`{"mcpServers":{"foo":{"command":"g"}}}`), 0o600)
	if err := s.Finish(ch, decideBy(map[string]Decision{"foo": Promote})); err != nil {
		t.Fatal(err)
	}
	if f.hostValue("foo") != "" {
		t.Fatal("promote must not write a value the user did not see")
	}
	if !strings.Contains(f.out.String(), "changed since the prompt") {
		t.Fatalf("out = %q", f.out.String())
	}
	if got := f.profileNames(); strings.Join(got, ",") != "foo" {
		t.Fatalf("kept in the profile: %v", got)
	}
}

func TestEngine_TwoSessionsOnlyLastStrips(t *testing.T) {
	f := newFixture(t)
	f.files(`{"mcpServers":{"jev":{"command":"npx"}}}`, `{"mcpServers":{}}`)
	a := f.start()
	b := f.start()
	cha, _ := a.Diff()
	_ = a.Finish(cha, keepAll)
	if got := f.profileNames(); len(got) != 1 {
		t.Fatalf("first exit stripped while second session runs: %v", got)
	}
	chb, _ := b.Diff()
	_ = b.Finish(chb, keepAll)
	if got := f.profileNames(); len(got) != 0 {
		t.Fatalf("last exit must strip: %v", got)
	}
}

func TestEngine_TwoSessionsKeptChangeIsNotReportedTwice(t *testing.T) {
	f := newFixture(t)
	f.files(`{"mcpServers":{"jev":{"command":"npx"}}}`, `{"mcpServers":{}}`)
	a := f.start()
	b := f.start()
	_ = os.WriteFile(f.pf, []byte(`{"mcpServers":{"jev":{"command":"npx2"},"foo":{"command":"f"}}}`), 0o600)
	cha, _ := a.Diff()
	if len(cha) != 2 {
		t.Fatalf("first session changes = %+v", cha)
	}
	_ = a.Finish(cha, keepAll)
	chb, _ := b.Diff()
	if len(chb) != 0 {
		t.Fatalf("changes kept by the first session reported again: %+v", chb)
	}
	_ = b.Finish(chb, keepAll)
	if got := f.profileNames(); strings.Join(got, ",") != "jev,foo" {
		t.Fatalf("at rest = %v", got)
	}
}

func TestEngine_SecondSessionDoesNotRemerge(t *testing.T) {
	f := newFixture(t)
	f.files(`{"mcpServers":{}}`, `{"mcpServers":{}}`) // nothing to merge: Active stays empty
	a := f.start()
	_ = os.WriteFile(f.pf, []byte(`{"mcpServers":{"foo":{"command":"f"}}}`), 0o600)
	b := f.start() // must not overwrite StartProfile with foo
	cha, _ := a.Diff()
	if kinds(cha) != "foo:added" {
		t.Fatalf("first session must still see foo as added: %+v", cha)
	}
	_ = a.Finish(cha, keepAll)
	chb, _ := b.Diff()
	_ = b.Finish(chb, keepAll)
}

func TestEngine_SessionsArePerAgent(t *testing.T) {
	f := newFixture(t)
	f.files(`{"mcpServers":{"jev":{"command":"npx"}}}`, `{"mcpServers":{}}`)
	other := filepath.Join(t.TempDir(), "codex.json")
	_ = os.WriteFile(other, []byte(`{"mcpServers":{}}`), 0o600)
	codexCol := Collection{Agent: "codex", Name: "mcpServers", Format: JSON, Key: "mcpServers", HostPath: f.host, ProfilePath: other}
	c, err := f.eng.Start("work", "codex", []Collection{codexCol}, StartOptions{Enabled: allOn})
	if err != nil {
		t.Fatal(err)
	}
	a := f.start()
	cha, _ := a.Diff()
	_ = a.Finish(cha, keepAll)
	if got := f.profileNames(); len(got) != 0 {
		t.Fatalf("a codex session must not keep claude's merge alive: %v", got)
	}
	chc, _ := c.Diff()
	_ = c.Finish(chc, keepAll)
	if got := f.names(other); len(got) != 0 {
		t.Fatalf("codex at rest = %v", got)
	}
}

func TestEngine_RecoverAfterCrash(t *testing.T) {
	f := newFixture(t)
	f.files(`{"mcpServers":{"jev":{"command":"npx"}}}`, `{"mcpServers":{}}`)
	s := f.start()
	_ = os.WriteFile(f.pf, []byte(`{"mcpServers":{"jev":{"command":"npx"},"foo":{"command":"f"}}}`), 0o600)
	s.abandon() // simulate kill -9: the sessions lock is released without Finish
	s2 := f.start()
	if !strings.Contains(f.out.String(), "1 change(s) kept in work") {
		t.Fatalf("recovery must say what it kept: %q", f.out.String())
	}
	if names := f.profileNames(); strings.Join(names, ",") != "foo,jev" && strings.Join(names, ",") != "jev,foo" {
		t.Fatalf("second session sees %v", names)
	}
	ch, _ := s2.Diff()
	if len(ch) != 0 {
		t.Fatalf("recovered session must start clean: %+v", ch)
	}
	_ = s2.Finish(ch, keepAll)
	if names := f.profileNames(); strings.Join(names, ",") != "foo" {
		t.Fatalf("the kept item stays the profile's own: %v", names)
	}
}

func TestEngine_BackgroundThenRecover(t *testing.T) {
	f := newFixture(t)
	f.files(`{"mcpServers":{"jev":{"command":"npx"}}}`, `{"mcpServers":{"mine":{"command":"m"}}}`)
	bg, err := f.eng.Start("work", "claude", []Collection{f.col}, StartOptions{Enabled: allOn, Background: true})
	if err != nil {
		t.Fatal(err)
	}
	if got := f.profileNames(); strings.Join(got, ",") != "mine,jev" {
		t.Fatalf("background launch must merge: %v", got)
	}
	ch, _ := bg.Diff()
	if len(ch) != 0 {
		t.Fatalf("background Diff is a no-op: %+v", ch)
	}
	if err := bg.Finish(ch, keepAll); err != nil {
		t.Fatal(err)
	}
	if got := f.profileNames(); strings.Join(got, ",") != "mine,jev" {
		t.Fatalf("background Finish must not strip: %v", got)
	}
	// the background session wrote a change back before the next launch
	_ = os.WriteFile(f.pf, []byte(`{"mcpServers":{"mine":{"command":"m"},"jev":{"command":"npx"},"bgadd":{"command":"b"}}}`), 0o600)
	s := f.start() // recovers: keeps bgadd, strips jev, re-merges jev
	if got := f.profileNames(); strings.Join(got, ",") != "mine,bgadd,jev" {
		t.Fatalf("after recovery + merge = %v", got)
	}
	chs, _ := s.Diff()
	if len(chs) != 0 {
		t.Fatalf("recovered launch starts clean: %+v", chs)
	}
	_ = s.Finish(chs, keepAll)
	if got := f.profileNames(); strings.Join(got, ",") != "mine,bgadd" {
		t.Fatalf("at rest = %v", got)
	}
}

func TestEngine_MigrationBackupAndMessage(t *testing.T) {
	f := newFixture(t)
	f.files(`{"mcpServers":{"a":{"command":"1"},"b":{"command":"2"}}}`,
		`{"mcpServers":{"a":{"command":"1"},"b":{"command":"X"},"own":{"command":"o"}}}`)
	s := f.start()
	if !strings.Contains(f.out.String(), "claude: 1 server(s) in work differ from the host's (b) and stay the profile's own") {
		t.Fatalf("migration message = %q", f.out.String())
	}
	matches, _ := filepath.Glob(f.pf + ".aim-backup-*")
	if len(matches) != 1 {
		t.Fatalf("migration must back up the profile file once: %v", matches)
	}
	if !strings.Contains(f.out.String(), "claude: 1 server(s) in work matched the host's (a) and were removed; backup: "+matches[0]+"\n") {
		t.Fatalf("the removal must name the backup: %q", f.out.String())
	}
	ch, _ := s.Diff()
	_ = s.Finish(ch, keepAll)
	if names := f.profileNames(); strings.Join(names, ",") != "b,own" {
		t.Fatalf("at rest = %v (a was a host copy)", names)
	}
	f.out.Reset()
	s2 := f.start()
	if strings.Contains(f.out.String(), "differ from the host") || strings.Contains(f.out.String(), "backup:") {
		t.Fatal("the migration messages are printed once")
	}
	if matches, _ := filepath.Glob(f.pf + ".aim-backup-*"); len(matches) != 1 {
		t.Fatalf("no second backup: %v", matches)
	}
	ch2, _ := s2.Diff()
	_ = s2.Finish(ch2, keepAll)
}

func TestEngine_MigrationBackupNeverOverwrites(t *testing.T) {
	f := newFixture(t)
	f.files(`{"mcpServers":{"a":{"command":"1"}}}`, `{"mcpServers":{"a":{"command":"1"}}}`)
	old := f.pf + ".aim-backup-19700101T000000Z"
	_ = os.WriteFile(old, []byte("earlier backup"), 0o600)
	s := f.start()
	ch, _ := s.Diff()
	_ = s.Finish(ch, keepAll)
	if b, _ := os.ReadFile(old); string(b) != "earlier backup" {
		t.Fatal("an existing backup must never be overwritten")
	}
	if b, _ := os.ReadFile(old + "-2"); string(b) != `{"mcpServers":{"a":{"command":"1"}}}` {
		t.Fatalf("new backup = %s", b)
	}
}

func TestEngine_SymlinkedProfileFileIsSkipped(t *testing.T) {
	f := newFixture(t)
	target := filepath.Join(t.TempDir(), "shared.json")
	_ = os.WriteFile(f.host, []byte(`{"mcpServers":{"jev":{"command":"npx"}}}`), 0o600)
	_ = os.WriteFile(target, []byte(`{"mcpServers":{}}`), 0o600)
	if err := os.Symlink(target, f.pf); err != nil {
		t.Fatal(err)
	}
	s := f.start()
	if !strings.Contains(f.out.String(), "symlink") {
		t.Fatalf("want a symlink warning, got %q", f.out.String())
	}
	ch, _ := s.Diff()
	_ = s.Finish(ch, keepAll)
	if b, _ := os.ReadFile(target); string(b) != `{"mcpServers":{}}` {
		t.Fatalf("symlink target changed: %s", b)
	}
	if fi, _ := os.Lstat(f.pf); fi.Mode()&os.ModeSymlink == 0 {
		t.Fatal("the symlink must stay a symlink")
	}
}

func TestEngine_TOMLNoOpSessionsDoNotGrowFile(t *testing.T) {
	d := t.TempDir()
	host := filepath.Join(d, "host.toml")
	prof := filepath.Join(d, "config.toml")
	_ = os.WriteFile(host, []byte("[mcp_servers.jev]\ncommand = \"npx\"\n"), 0o600)
	src := "model = \"x\"\n\n[mcp_servers.own]\ncommand = \"o\"\n"
	_ = os.WriteFile(prof, []byte(src), 0o600)
	col := Collection{Agent: "codex", Name: "mcp_servers", Format: TOML, Key: "mcp_servers", HostPath: host, ProfilePath: prof}
	eng := &Engine{Store: Store{Dir: filepath.Join(d, "state")}, Out: &bytes.Buffer{}, Now: time.Now}
	for i := 0; i < 4; i++ {
		s, err := eng.Start("work", "codex", []Collection{col}, StartOptions{Enabled: allOn})
		if err != nil {
			t.Fatal(err)
		}
		if b, _ := os.ReadFile(prof); !strings.Contains(string(b), "[mcp_servers.jev]") {
			t.Fatalf("session %d does not see the host server", i+1)
		}
		ch, _ := s.Diff()
		if err := s.Finish(ch, keepAll); err != nil {
			t.Fatal(err)
		}
		if b, _ := os.ReadFile(prof); string(b) != src {
			t.Fatalf("session %d changed the file (%d → %d bytes):\n%s", i+1, len(src), len(b), b)
		}
	}
}

func TestEngine_DisabledStillRecovers(t *testing.T) {
	f := newFixture(t)
	f.files(`{"mcpServers":{"jev":{"command":"npx"}}}`, `{"mcpServers":{"mine":{"command":"m"}}}`)
	s := f.start()
	s.abandon() // crash with jev merged
	off, err := f.eng.Start("work", "claude", []Collection{f.col}, StartOptions{Enabled: allOff})
	if err != nil {
		t.Fatal(err)
	}
	if got := f.profileNames(); strings.Join(got, ",") != "mine" {
		t.Fatalf("mcp_global=false must still strip a crashed merge, and merge nothing: %v", got)
	}
	sessionsLockFree(t, f.eng, "work", "claude")
	ch, _ := off.Diff()
	if len(ch) != 0 {
		t.Fatalf("disabled Diff = %+v", ch)
	}
	_ = os.WriteFile(f.pf, []byte(`{"mcpServers":{"mine":{"command":"m2"}}}`), 0o600)
	if err := off.Finish(ch, keepAll); err != nil {
		t.Fatal(err)
	}
	if got := f.profileNames(); strings.Join(got, ",") != "mine" {
		t.Fatalf("disabled Finish touches nothing: %v", got)
	}
}

func TestEngine_HostUnreadableSkipsCollection(t *testing.T) {
	f := newFixture(t)
	f.files(`{not json`, `{"mcpServers":{"own":{"command":"o"}}}`)
	s := f.start()
	if got := f.profileNames(); len(got) != 1 {
		t.Fatalf("profile changed despite unreadable host: %v", got)
	}
	if !strings.Contains(f.out.String(), "skipped") {
		t.Fatalf("want a warning, got %q", f.out.String())
	}
	ch, _ := s.Diff()
	_ = s.Finish(ch, keepAll)
}

func TestEngine_HostMissingIsSilentSkip(t *testing.T) {
	f := newFixture(t)
	_ = os.WriteFile(f.pf, []byte(`{"mcpServers":{"own":{"command":"o"}}}`), 0o600)
	s := f.start()
	if f.out.Len() != 0 {
		t.Fatalf("no host file is not worth a warning: %q", f.out.String())
	}
	_ = os.WriteFile(f.pf, []byte(`{"mcpServers":{"own":{"command":"o"},"new":{"command":"n"}}}`), 0o600)
	ch, _ := s.Diff()
	if len(ch) != 0 {
		t.Fatalf("a skipped collection reports no changes: %+v", ch)
	}
	_ = s.Finish(ch, keepAll)
}

func TestEngine_JoinerSkipsCollectionTheFirstSessionSkipped(t *testing.T) {
	f := newFixture(t)
	_ = os.WriteFile(f.pf, []byte(`{"mcpServers":{"own":{"command":"o"}}}`), 0o600)
	a := f.start() // no host file: nothing merged, no session state
	b := f.start() // joins a
	chb, _ := b.Diff()
	if len(chb) != 0 {
		t.Fatalf("a joiner must not report the profile's own servers: %+v", chb)
	}
	if err := b.Finish(chb, decideBy(map[string]Decision{"own": Promote})); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(f.host); !os.IsNotExist(err) {
		t.Fatal("promote must not create a host file for a collection no session merged")
	}
	cha, _ := a.Diff()
	_ = a.Finish(cha, keepAll)
	if got := f.profileNames(); strings.Join(got, ",") != "own" {
		t.Fatalf("at rest = %v", got)
	}
}

func TestEngine_PromoteFollowsSymlinkedHost(t *testing.T) {
	f := newFixture(t)
	target := filepath.Join(t.TempDir(), "dotfiles.json")
	_ = os.WriteFile(target, []byte(`{"mcpServers":{"jev":{"command":"npx"}}}`), 0o600)
	if err := os.Symlink(target, f.host); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(f.pf, []byte(`{"mcpServers":{}}`), 0o600)
	s := f.start()
	_ = os.WriteFile(f.pf, []byte(`{"mcpServers":{"jev":{"command":"npx"},"foo":{"command":"f"}}}`), 0o600)
	ch, _ := s.Diff()
	if err := s.Finish(ch, decideBy(map[string]Decision{"foo": Promote})); err != nil {
		t.Fatal(err)
	}
	if fi, err := os.Lstat(f.host); err != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Fatal("promote must keep a symlinked host a symlink")
	}
	e, _, err := f.col.Read(target)
	if err != nil || !e.Has("foo") {
		t.Fatalf("the link target must hold the promoted foo: %v %v", e.Order, err)
	}
	if fi, _ := os.Stat(target); fi.Mode().Perm() != 0o600 {
		t.Fatalf("target mode = %v", fi.Mode().Perm())
	}
}

func TestEngine_SaveFailureRollsBackMerge(t *testing.T) {
	f := newFixture(t)
	sabotage := false
	f.col.Normalise = func(v map[string]any) map[string]any {
		if sabotage {
			// The state file becomes a directory after Load: Save cannot rename over it.
			sabotage = false
			st := f.eng.Store.path("work")
			_ = os.Remove(st)
			_ = os.MkdirAll(filepath.Join(st, "x"), 0o700)
		}
		return DropEmpty(v)
	}
	f.files(`{"mcpServers":{"jev":{"command":"npx"}}}`, `{"mcpServers":{"own":{"command":"o"}}}`)
	s := f.start() // migrates
	ch, _ := s.Diff()
	if err := s.Finish(ch, keepAll); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(f.pf)
	sabotage = true
	if _, err := f.eng.Start("work", "claude", []Collection{f.col}, StartOptions{Enabled: allOn}); err == nil {
		t.Fatal("Start must report the failed Save")
	}
	if after, _ := os.ReadFile(f.pf); string(after) != string(before) {
		t.Fatalf("an unrecorded merge must be rolled back:\n%s\n%s", before, after)
	}
}

func TestEngine_FailedProfileWriteRecordsNoState(t *testing.T) {
	d := t.TempDir()
	host := filepath.Join(d, "host.toml")
	prof := filepath.Join(d, "config.toml")
	_ = os.WriteFile(host, []byte("[mcp_servers.jev]\ncommand = \"npx\"\n"), 0o600)
	// an inline table cannot be extended by a spliced [mcp_servers.jev] block
	_ = os.WriteFile(prof, []byte("mcp_servers = {}\n"), 0o600)
	col := Collection{Agent: "codex", Name: "mcp_servers", Format: TOML, Key: "mcp_servers", HostPath: host, ProfilePath: prof}
	out := &bytes.Buffer{}
	eng := &Engine{Store: Store{Dir: filepath.Join(d, "state")}, Out: out, Now: time.Now}
	s, err := eng.Start("work", "codex", []Collection{col}, StartOptions{Enabled: allOn})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "skipped") {
		t.Fatalf("want a write-failure warning, got %q", out.String())
	}
	st, _ := eng.Store.Load("work")
	if _, ok := st.Active[col.ID()]; ok {
		t.Fatalf("a failed write must record no Active names: %+v", st.Active)
	}
	ch, _ := s.Diff()
	_ = s.Finish(ch, keepAll)
	out.Reset()
	s2, err := eng.Start("work", "codex", []Collection{col}, StartOptions{Enabled: allOn})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "did not exit through aim") {
		t.Fatalf("no spurious recovery: %q", out.String())
	}
	ch2, _ := s2.Diff()
	_ = s2.Finish(ch2, keepAll)
}

// One collection failing must not undo or block another, in either order:
// collection 1 is merged, recorded and stripped back to its pre-Start bytes;
// collection 2 is skipped with a warning.
func TestEngine_PartialFailureMergesTheOtherCollection(t *testing.T) {
	for _, badFirst := range []bool{false, true} {
		t.Run(fmt.Sprintf("badFirst=%v", badFirst), func(t *testing.T) {
			f := newFixture(t)
			d := t.TempDir()
			// in renderObject's layout, so the strip reproduces the file byte for byte
			src := "{\"x\": 1, \"mcpServers\": {\n    \"own\": {\"command\": \"o\"}\n  }}\n"
			f.files(`{"mcpServers":{"jev":{"command":"npx"}}}`, src)
			bad := Collection{Agent: "claude", Name: "servers", Format: JSON, Key: "servers",
				HostPath: filepath.Join(d, "host2.json"), ProfilePath: filepath.Join(d, "profile2.json")}
			_ = os.WriteFile(bad.HostPath, []byte(`{not json`), 0o600)
			_ = os.WriteFile(bad.ProfilePath, []byte(`{"servers":{}}`), 0o600)
			cols := []Collection{f.col, bad}
			if badFirst {
				cols = []Collection{bad, f.col}
			}
			s, err := f.eng.Start("work", "claude", cols, StartOptions{Enabled: allOn})
			if err != nil {
				t.Fatalf("a failing collection must not fail Start: %v", err)
			}
			if out := f.out.String(); !strings.Contains(out, bad.ID()) || !strings.Contains(out, "skipped") || strings.Contains(out, f.col.ID()) {
				t.Fatalf("want a warning for %s only, got %q", bad.ID(), out)
			}
			if got := f.profileNames(); strings.Join(got, ",") != "own,jev" {
				t.Fatalf("collection 1 during session = %v", got)
			}
			st, _ := f.eng.Store.Load("work")
			if st.Active[f.col.ID()]["jev"] == "" {
				t.Fatalf("collection 1's merge must be recorded: %+v", st.Active)
			}
			if _, ok := st.Active[bad.ID()]; ok {
				t.Fatalf("the skipped collection must record nothing: %+v", st.Active)
			}
			ch, err := s.Diff()
			if err != nil || len(ch) != 0 {
				t.Fatalf("changes = %+v, err = %v", ch, err)
			}
			if err := s.Finish(ch, keepAll); err != nil {
				t.Fatal(err)
			}
			if b, _ := os.ReadFile(f.pf); string(b) != src {
				t.Fatalf("collection 1 at rest:\n%s\nwant:\n%s", b, src)
			}
			if b, _ := os.ReadFile(bad.ProfilePath); string(b) != `{"servers":{}}` {
				t.Fatalf("the skipped collection's profile changed: %s", b)
			}
		})
	}
}

// pluginsCol is a second collection in the fixture's files, the shape Claude's
// enabledPlugins has: bool values in the same settings file as mcpServers.
func (f *fixture) pluginsCol() Collection {
	return Collection{Agent: "claude", Name: "enabledPlugins", Format: JSON, Key: "enabledPlugins",
		HostPath: f.host, ProfilePath: f.pf, Group: "plugins", Noun: "plugin"}
}

func TestEngine_DisabledGroupIsRecoveredNotMerged(t *testing.T) {
	f := newFixture(t)
	f.col.Group = "mcp"
	plug := f.pluginsCol()
	f.files(`{"mcpServers":{"jev":{"command":"npx"}},"enabledPlugins":{"x@m":true}}`, `{"mcpServers":{}}`)
	cols := []Collection{f.col, plug}
	if _, err := f.eng.Start("work", "claude", cols, StartOptions{Enabled: allOn, Background: true}); err != nil {
		t.Fatal(err)
	}
	if got := f.names(f.pf); strings.Join(got, ",") != "jev" {
		t.Fatalf("background mcp merge = %v", got)
	}
	if e, _, _ := plug.Read(f.pf); strings.Join(e.Order, ",") != "x@m" {
		t.Fatalf("background plugin merge = %v", e.Order)
	}
	mcpOnly := func(c Collection) bool { return c.Group != "plugins" }
	s, err := f.eng.Start("work", "claude", cols, StartOptions{Enabled: mcpOnly})
	if err != nil {
		t.Fatal(err)
	}
	if got := f.profileNames(); strings.Join(got, ",") != "jev" {
		t.Fatalf("the enabled group must be merged again: %v", got)
	}
	if has, _ := JSONHasKey(f.pf, "enabledPlugins"); has {
		b, _ := os.ReadFile(f.pf)
		t.Fatalf("the disabled group must be recovered, key and all:\n%s", b)
	}
	st, _ := f.eng.Store.Load("work")
	if _, ok := st.Active[plug.ID()]; ok || st.Active[f.col.ID()]["jev"] == "" {
		t.Fatalf("Active = %+v", st.Active)
	}
	if len(s.cols) != 1 || s.cols[0].ID() != f.col.ID() {
		t.Fatalf("session collections = %+v", s.cols)
	}
	ch, _ := s.Diff()
	if err := s.Finish(ch, keepAll); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(f.pf); string(b) != `{"mcpServers":{}}` {
		t.Fatalf("at rest = %s", b)
	}
}

func TestEngine_AllDisabledAfterBackgroundRecoversWithoutLock(t *testing.T) {
	f := newFixture(t)
	plug := f.pluginsCol()
	f.files(`{"mcpServers":{"jev":{"command":"npx"}},"enabledPlugins":{"x@m":true}}`, `{"mcpServers":{"mine":{"command":"m"}}}`)
	cols := []Collection{f.col, plug}
	if _, err := f.eng.Start("work", "claude", cols, StartOptions{Enabled: allOn, Background: true}); err != nil {
		t.Fatal(err)
	}
	off, err := f.eng.Start("work", "claude", cols, StartOptions{Enabled: allOff})
	if err != nil {
		t.Fatal(err)
	}
	if has, _ := JSONHasKey(f.pf, "enabledPlugins"); has || strings.Join(f.profileNames(), ",") != "mine" {
		b, _ := os.ReadFile(f.pf)
		t.Fatalf("every group must be recovered:\n%s", b)
	}
	sessionsLockFree(t, f.eng, "work", "claude")
	if len(off.cols) != 0 {
		t.Fatalf("nothing merged, got %+v", off.cols)
	}
	st, _ := f.eng.Store.Load("work")
	if len(st.Active) != 0 {
		t.Fatalf("Active = %+v", st.Active)
	}
	// switched back on, the next launch is alone and merges again
	on := f.start()
	if got := f.profileNames(); strings.Join(got, ",") != "mine,jev" {
		t.Fatalf("re-enabled launch = %v", got)
	}
	ch, _ := on.Diff()
	_ = on.Finish(ch, keepAll)
}

func TestEngine_NilEnabledMergesEveryCollection(t *testing.T) {
	f := newFixture(t)
	plug := f.pluginsCol()
	f.files(`{"mcpServers":{"jev":{"command":"npx"}},"enabledPlugins":{"x@m":true}}`, `{}`)
	s, err := f.eng.Start("work", "claude", []Collection{f.col, plug}, StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(s.cols) != 2 {
		t.Fatalf("a nil predicate means every collection: %+v", s.cols)
	}
	ch, _ := s.Diff()
	_ = s.Finish(ch, keepAll)
	if b, _ := os.ReadFile(f.pf); string(b) != `{}` {
		t.Fatalf("at rest = %s", b)
	}
}

func TestEngine_MigrationMessageUsesNoun(t *testing.T) {
	f := newFixture(t)
	plug := f.pluginsCol()
	f.files(`{"enabledPlugins":{"shared@m":true,"diff@m":true}}`, `{"enabledPlugins":{"shared@m":true,"diff@m":false}}`)
	s, err := f.eng.Start("work", "claude", []Collection{plug}, StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	out := f.out.String()
	if !strings.Contains(out, "claude: 1 plugin(s) in work matched the host's (shared@m) and were removed; backup: ") ||
		!strings.Contains(out, "claude: 1 plugin(s) in work differ from the host's (diff@m) and stay the profile's own") {
		t.Fatalf("migration message = %q", out)
	}
	ch, _ := s.Diff()
	_ = s.Finish(ch, keepAll)
}

// Two collections share one profile file and only the second migrates this
// Start: the first one's merge must not reach the backup, or restoring it makes
// the host's servers (and their secrets) the profile's own.
func TestEngine_MigrationBackupIsThePreStartFile(t *testing.T) {
	f := newFixture(t)
	f.col.Group = "mcp"
	plug := f.pluginsCol()
	src := `{"enabledPlugins":{"x@m":true}}`
	f.files(`{"mcpServers":{"jev":{"command":"npx","env":{"TOKEN":"secret"}}},"enabledPlugins":{"x@m":true}}`, src)
	st, _ := f.eng.Store.Load("work")
	st.Migrated[f.col.ID()] = time.Unix(0, 0)
	if err := f.eng.Store.Save("work", st); err != nil {
		t.Fatal(err)
	}
	s, err := f.eng.Start("work", "claude", []Collection{f.col, plug}, StartOptions{Enabled: allOn})
	if err != nil {
		t.Fatal(err)
	}
	matches, _ := filepath.Glob(f.pf + ".aim-backup-*")
	if len(matches) != 1 {
		t.Fatalf("want one backup, got %v", matches)
	}
	if b, _ := os.ReadFile(matches[0]); string(b) != src {
		t.Fatalf("backup =\n%s\nwant the pre-Start file\n%s", b, src)
	}
	ch, _ := s.Diff()
	_ = s.Finish(ch, keepAll)
}

// A joiner with fewer groups switched on can be the last session out: its
// strip must cover every group the running sessions merged, not just its own.
func TestEngine_LastSessionStripsGroupsItDidNotJoin(t *testing.T) {
	f := newFixture(t)
	f.col.Group = "mcp"
	plug := f.pluginsCol()
	src := `{"mcpServers":{}}`
	f.files(`{"mcpServers":{"jev":{"command":"npx"}},"enabledPlugins":{"x@m":true}}`, src)
	cols := []Collection{f.col, plug}
	a, err := f.eng.Start("work", "claude", cols, StartOptions{Enabled: allOn})
	if err != nil {
		t.Fatal(err)
	}
	mcpOnly := func(c Collection) bool { return c.Group != "plugins" }
	b, err := f.eng.Start("work", "claude", cols, StartOptions{Enabled: mcpOnly})
	if err != nil {
		t.Fatal(err)
	}
	if len(b.cols) != 1 || b.cols[0].ID() != f.col.ID() {
		t.Fatalf("the joiner diffs its own groups only: %+v", b.cols)
	}
	cha, _ := a.Diff()
	if err := a.Finish(cha, keepAll); err != nil {
		t.Fatal(err)
	}
	chb, _ := b.Diff()
	if err := b.Finish(chb, keepAll); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(f.pf); string(got) != src {
		t.Fatalf("at rest =\n%s\nwant\n%s", got, src)
	}
	st, _ := f.eng.Store.Load("work")
	if len(st.Active) != 0 || len(st.AddedKey) != 0 {
		t.Fatalf("session state left behind: Active=%+v AddedKey=%+v", st.Active, st.AddedKey)
	}
}
