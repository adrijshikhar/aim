package merge

import (
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"sort"
	"strings"
	"time"
)

// Change is one difference between what a session started with and what it
// left in the profile file.
type Change struct {
	Collection Collection
	Name       string
	Kind       ChangeKind
	Value      map[string]any
	Raw        []byte
	// HostChanged: for edited/removed, the host's copy differs from the value
	// merged at start; for added, the host now defines the name. Promote would
	// overwrite the host, so it is refused (spec §5 End step 3).
	HostChanged bool
}

type Decision int

const (
	// Keep leaves the change in the profile (added/edited: the profile's own;
	// removed: the host item comes back next session).
	Keep Decision = iota
	// Promote writes the change to the host file.
	Promote
)

// StartOptions: Enabled picks the collections to merge (profiles.<p>.mcp_global,
// plugins_global; nil means every one) — recovery covers all of them either
// way. Background means merge, run, no exit step (spec R2).
type StartOptions struct {
	Enabled    func(Collection) bool
	Background bool
}

type Engine struct {
	Store Store
	Out   io.Writer
	Now   func() time.Time
}

type Session struct {
	eng      *Engine
	profile  string
	agent    string
	cols     []Collection // collections this session merged or joined
	all      []Collection // every collection passed to Start: the last-session strip
	sessions *Lock
	active   bool // a foreground session with an exit step
}

const lockTimeout = 10 * time.Second

// errNoHost skips a collection silently: with no host file there is nothing to merge.
var errNoHost = errors.New("host file missing")

// errNotMerged skips a collection silently for a joining session: the first
// session did not merge it, so there is no session state to diff against.
var errNotMerged = errors.New("not merged by the running session")

func (e *Engine) warnf(format string, a ...any) {
	if e.Out != nil {
		fmt.Fprintf(e.Out, format+"\n", a...)
	}
}

func (e *Engine) now() time.Time {
	if e.Now != nil {
		return e.Now()
	}
	return time.Now()
}

func readRetry(c Collection, path string) (Entries, bool, error) {
	ent, ok, err := c.Read(path)
	if err != nil {
		time.Sleep(100 * time.Millisecond)
		ent, ok, err = c.Read(path)
	}
	return ent, ok, err
}

func setHash(m map[string]map[string]string, id, name, h string) {
	if m[id] == nil {
		m[id] = map[string]string{}
	}
	m[id][name] = h
}

// hasSessionState reports whether a session of these collections started and
// has not finished (its Active entry is still present).
func hasSessionState(st *State, cols []Collection) bool {
	for _, c := range cols {
		if _, ok := st.Active[c.ID()]; ok {
			return true
		}
	}
	return false
}

// Start runs spec §5 Start. A foreground session holds the shared sessions
// lock for (profile, agent) when Start returns with active=true; a background
// launch merges and returns a session whose Diff and Finish do nothing.
func (e *Engine) Start(profile, agent string, cols []Collection, opt StartOptions) (*Session, error) {
	s := &Session{eng: e, profile: profile, agent: agent, all: cols}
	ml, err := LockExclusive(e.Store.MergeLockPath(profile), lockTimeout)
	if err != nil {
		e.warnf("aim: host merge skipped for %s (%v)", profile, err)
		return s, nil
	}
	defer ml.Unlock()

	st, err := e.Store.Load(profile)
	if err != nil {
		return s, err
	}
	probe, alone, err := TryUpgrade(e.Store.SessionsLockPath(profile, agent))
	if err != nil {
		return s, err
	}
	if alone {
		_ = probe.Unlock()
		if hasSessionState(st, cols) {
			// The previous session ended without an exit step (crash, or a
			// background launch): keep its changes, strip its host items.
			if err := e.recover(profile, agent, cols, st); err != nil {
				e.warnf("aim: recovery for %s: %v", profile, err)
			}
		}
	}
	var allowed []Collection
	for _, c := range cols {
		if opt.Enabled == nil || opt.Enabled(c) {
			allowed = append(allowed, c)
		}
	}
	if len(allowed) == 0 {
		return s, e.Store.Save(profile, st) // no sessions lock: Diff and Finish do nothing
	}
	bk := newBackups()
	if alone {
		// Before anything is merged: a collection that migrates later in this
		// loop must back up its file as Start found it, not with an earlier
		// collection's host items already in it.
		bk.snapshot(allowed, st)
	}
	var undo []stripWrite
	for _, c := range allowed {
		w, err := e.startCollection(profile, c, st, alone, bk)
		if err != nil {
			if !errors.Is(err, errNoHost) && !errors.Is(err, errNotMerged) {
				e.warnf("aim: %s: %v — skipped", c.ID(), err)
			}
			continue
		}
		s.cols = append(s.cols, c)
		undo = append(undo, w)
	}
	if err := e.Store.Save(profile, st); err != nil {
		// Nothing records the merge: take the host items out again, or they
		// become the profile's own for good.
		for _, w := range undo {
			e.writeStrip(w)
		}
		return s, err
	}
	if opt.Background {
		return s, nil // merged; the next launch of this agent recovers
	}
	s.sessions, err = LockShared(e.Store.SessionsLockPath(profile, agent))
	if err != nil {
		return s, err
	}
	s.active = true
	return s, nil
}

// startCollection merges one collection and returns the write that undoes it
// (skip when nothing was written). State is recorded only once the file holds
// the merge.
func (e *Engine) startCollection(profile string, c Collection, st *State, alone bool, bk *backups) (stripWrite, error) {
	none := stripWrite{c: c, skip: true}
	if fi, err := os.Lstat(c.ProfilePath); err == nil && fi.Mode()&os.ModeSymlink != 0 {
		return none, errors.New(c.ProfilePath + " is a symlink; not merging into a shared file")
	}
	if !alone {
		if _, ok := st.Active[c.ID()]; !ok {
			return none, errNotMerged
		}
		return none, nil // another running session of this agent already merged; join it
	}
	host, hostOK, err := readRetry(c, c.HostPath)
	if err != nil {
		return none, err
	}
	if !hostOK {
		return none, errNoHost
	}
	prof, _, err := readRetry(c, c.ProfilePath)
	if err != nil {
		return none, err
	}
	id := c.ID()
	norm := c.norm()
	if _, done := st.Migrated[id]; !done {
		if err := e.migrate(profile, c, host, &prof, st, bk); err != nil {
			return none, err
		}
	}
	// Every profile item is the profile's own; a name the profile defines wins
	// over the host's (conflict rule) and is never merged over or stripped.
	start, active := map[string]string{}, map[string]string{}
	pre := NewEntries()
	for _, n := range prof.Order {
		start[n] = Hash(norm, prof.Values[n])
		pre.Set(n, prof.Values[n], prof.Raw[n])
	}
	changed := false
	for _, n := range host.Order {
		if prof.Has(n) {
			continue
		}
		prof.Set(n, host.Values[n], host.Raw[n])
		active[n] = Hash(norm, host.Values[n])
		changed = true
	}
	if !changed {
		st.StartProfile[id] = start
		st.Active[id] = active
		return none, nil
	}
	hadKey, err := c.HasKey(c.ProfilePath)
	if err != nil {
		return none, err
	}
	skipped, err := c.Write(c.ProfilePath, prof)
	for _, n := range skipped {
		delete(active, n)
		e.warnf("aim: %s: %s is defined inline and was not merged", id, n)
	}
	if err != nil {
		return none, err
	}
	st.StartProfile[id] = start
	st.Active[id] = active
	if !hadKey {
		st.AddedKey[id] = true
	}
	return stripWrite{c: c, prof: pre, removeKey: !hadKey && len(pre.Order) == 0}, nil
}

// migrate runs once per collection: profile servers identical to the host's are
// old copy-once leftovers (T4) and are removed after a backup; servers that
// differ stay the profile's own (conflict rule). Both are listed once.
func (e *Engine) migrate(profile string, c Collection, host Entries, prof *Entries, st *State, bk *backups) error {
	id := c.ID()
	norm := c.norm()
	var removed, differing []string
	for _, n := range slices.Clone(prof.Order) {
		if !host.Has(n) {
			continue
		}
		if Hash(norm, prof.Values[n]) == Hash(norm, host.Values[n]) {
			prof.Delete(n)
			removed = append(removed, n)
		} else {
			differing = append(differing, n)
		}
	}
	if len(removed) > 0 {
		name, err := e.backup(c.ProfilePath, bk)
		if err != nil {
			return err
		}
		if _, err := c.Write(c.ProfilePath, *prof); err != nil {
			return err
		}
		if name != "" {
			e.warnf("%s: %d %s(s) in %s matched the host's (%s) and were removed; backup: %s",
				c.Agent, len(removed), c.noun(), profile, strings.Join(removed, ", "), name)
		}
	}
	if len(differing) > 0 {
		e.warnf("%s: %d %s(s) in %s differ from the host's (%s) and stay the profile's own",
			c.Agent, len(differing), c.noun(), profile, strings.Join(differing, ", "))
	}
	st.Migrated[id] = e.now()
	return nil
}

// backups is one Start's migration backups: each profile file a collection
// may migrate, as Start found it, and the one backup taken from it.
type backups struct {
	before map[string]snapshot
	names  map[string]string
}

// snapshot is a file's bytes before the merge loop (exists=false: no file).
type snapshot struct {
	data   []byte
	exists bool
	err    error
}

func newBackups() *backups {
	return &backups{before: map[string]snapshot{}, names: map[string]string{}}
}

// snapshot reads every profile file a collection that has not migrated yet
// lives in. Two collections can share one file (Claude's settings.json, Codex's
// config.toml), so the file is read once, before either merges into it.
func (b *backups) snapshot(cols []Collection, st *State) {
	for _, c := range cols {
		if _, done := st.Migrated[c.ID()]; done {
			continue
		}
		if _, ok := b.before[c.ProfilePath]; ok {
			continue
		}
		data, err := os.ReadFile(c.ProfilePath)
		switch {
		case errors.Is(err, os.ErrNotExist):
			b.before[c.ProfilePath] = snapshot{}
		case err != nil:
			b.before[c.ProfilePath] = snapshot{err: err}
		default:
			b.before[c.ProfilePath] = snapshot{data: data, exists: true}
		}
	}
}

// backup writes path's pre-Start bytes to <path>.aim-backup-<UTC>[-n] (0600)
// once per Start and never overwrites an existing backup. It returns the
// backup's name ("" when path did not exist).
func (e *Engine) backup(path string, b *backups) (string, error) {
	if name, ok := b.names[path]; ok {
		return name, nil
	}
	snap, ok := b.before[path]
	if !ok {
		return "", errors.New(path + ": no pre-Start snapshot to back up")
	}
	if snap.err != nil {
		return "", snap.err
	}
	if !snap.exists {
		return "", nil
	}
	base := fmt.Sprintf("%s.aim-backup-%s", path, e.now().UTC().Format("20060102T150405Z"))
	for i := 1; ; i++ {
		name := base
		if i > 1 {
			name = fmt.Sprintf("%s-%d", base, i)
		}
		f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if errors.Is(err, os.ErrExist) {
			continue
		}
		if err != nil {
			return "", err
		}
		_, werr := f.Write(snap.data)
		if cerr := f.Close(); werr == nil {
			werr = cerr
		}
		if werr != nil {
			return "", werr
		}
		b.names[path] = name
		return name, nil
	}
}

// Diff compares each collection with what the session started from (spec §5
// End step 1). It takes the merge lock only while reading.
func (s *Session) Diff() ([]Change, error) {
	if !s.active {
		return nil, nil
	}
	ml, err := LockExclusive(s.eng.Store.MergeLockPath(s.profile), lockTimeout)
	if err != nil {
		return nil, err
	}
	defer ml.Unlock()
	st, err := s.eng.Store.Load(s.profile)
	if err != nil {
		return nil, err
	}
	return s.eng.diff(s.cols, st)
}

// diff returns the changes sorted by collection, then name.
func (e *Engine) diff(cols []Collection, st *State) ([]Change, error) {
	var out []Change
	for _, c := range cols {
		id := c.ID()
		now, _, err := readRetry(c, c.ProfilePath)
		if err != nil {
			return out, err
		}
		host, _, _ := c.Read(c.HostPath) // best effort: an unreadable host counts as changed
		norm := c.norm()
		hostHash := func(n string) string {
			if host.Has(n) {
				return Hash(norm, host.Values[n])
			}
			return ""
		}
		active, start := st.Active[id], st.StartProfile[id]
		for n, h := range active {
			switch {
			case !now.Has(n):
				out = append(out, Change{Collection: c, Name: n, Kind: Removed, HostChanged: hostHash(n) != h})
			case Hash(norm, now.Values[n]) != h:
				out = append(out, Change{Collection: c, Name: n, Kind: Edited, Value: now.Values[n], Raw: now.Raw[n], HostChanged: hostHash(n) != h})
			}
		}
		for _, n := range now.Order {
			if _, isActive := active[n]; isActive {
				continue
			}
			if _, wasProfile := start[n]; wasProfile {
				continue // the profile's own change: kept, no question
			}
			out = append(out, Change{Collection: c, Name: n, Kind: Added, Value: now.Values[n], Raw: now.Raw[n], HostChanged: host.Has(n)})
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if a, b := out[i].Collection.ID(), out[j].Collection.ID(); a != b {
			return a < b
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

func sameChange(a, b Change) bool {
	return a.Collection.ID() == b.Collection.ID() && a.Name == b.Name && a.Kind == b.Kind
}

// sameValue reports whether two changes carry the same value (a removal has none).
func sameValue(a, b Change) bool {
	if a.Kind == Removed {
		return true
	}
	n := a.Collection.norm()
	return Hash(n, a.Value) == Hash(n, b.Value)
}

// Finish runs spec §5 End steps 2–5: decide WITHOUT the merge lock (the prompt
// may wait on the user), then lock, re-diff, apply the decisions to the changes
// still present (anything new is kept), strip if this is the agent's last
// session, save, release.
func (s *Session) Finish(changes []Change, decide func([]Change) []Decision) error {
	if !s.active {
		return nil
	}
	e := s.eng
	var decisions []Decision
	if len(changes) > 0 && decide != nil {
		decisions = decide(changes)
	}
	ml, err := LockExclusive(e.Store.MergeLockPath(s.profile), lockTimeout)
	if err != nil {
		_ = s.sessions.Unlock()
		return err
	}
	defer ml.Unlock()
	st, err := e.Store.Load(s.profile)
	if err != nil {
		_ = s.sessions.Unlock()
		return err
	}
	now, err := e.diff(s.cols, st)
	if err != nil {
		e.warnf("aim: host diff: %v", err)
	}
	for _, ch := range now {
		d := Keep
		for i, old := range changes {
			if i >= len(decisions) || !sameChange(old, ch) {
				continue
			}
			if decisions[i] == Promote && !sameValue(old, ch) {
				e.warnf("aim: %s %s: changed since the prompt — kept in %s", ch.Collection.ID(), ch.Name, s.profile)
				break
			}
			d = decisions[i]
			break
		}
		e.apply(s.profile, ch, d, st)
	}
	_ = s.sessions.Unlock()
	s.active = false
	probe, last, err := TryUpgrade(e.Store.SessionsLockPath(s.profile, s.agent))
	if err != nil || !last {
		return e.Store.Save(s.profile, st)
	}
	_ = probe.Unlock()
	// Last session: plan the strip, save state first, then write (spec §5 End 4).
	// It covers every collection a running session merged, not only this
	// session's: a joiner with a group switched off can still be the last out.
	var writes []stripWrite
	for _, c := range s.all {
		if _, ok := st.Active[c.ID()]; ok || st.AddedKey[c.ID()] {
			writes = append(writes, e.planStrip(c, st))
		}
	}
	if err := e.Store.Save(s.profile, st); err != nil {
		return err
	}
	for _, w := range writes {
		e.writeStrip(w)
	}
	return nil
}

// apply records one decision. A kept addition or edit becomes the profile's
// own (StartProfile), so a session still running does not report it again; a
// kept removal drops the name from Active (it comes back next session). A
// promote that fails falls back to keep.
func (e *Engine) apply(profile string, ch Change, d Decision, st *State) {
	id := ch.Collection.ID()
	if d == Promote {
		err := e.promote(ch, st)
		if err == nil {
			return
		}
		e.warnf("aim: %s %s: %v — kept in %s", id, ch.Name, err, profile)
	}
	switch ch.Kind {
	case Removed:
		delete(st.Active[id], ch.Name)
	case Edited:
		delete(st.Active[id], ch.Name)
		setHash(st.StartProfile, id, ch.Name, Hash(ch.Collection.norm(), ch.Value))
	case Added:
		setHash(st.StartProfile, id, ch.Name, Hash(ch.Collection.norm(), ch.Value))
	}
}

// stripWrite is a strip that has been computed but not yet written.
type stripWrite struct {
	c         Collection
	prof      Entries
	removeKey bool
	skip      bool
}

// planStrip computes the profile file without the host items merged by
// sessions and clears the collection's session state. The file is written
// after the state is saved (a crash in between leaves host copies the user can
// remove; the reverse order would misreport them as kept removals).
func (e *Engine) planStrip(c Collection, st *State) stripWrite {
	id := c.ID()
	w := stripWrite{c: c, skip: true}
	active := st.Active[id]
	if len(active) > 0 || st.AddedKey[id] {
		prof, ok, err := readRetry(c, c.ProfilePath)
		switch {
		case err != nil:
			e.warnf("aim: %s: strip: %v", id, err)
		case !ok:
			// the file vanished during the session: nothing to write
		default:
			for n, h := range active {
				if prof.Has(n) && Hash(c.norm(), prof.Values[n]) == h {
					prof.Delete(n)
				}
			}
			w.prof, w.skip = prof, false
			w.removeKey = st.AddedKey[id] && len(prof.Order) == 0
		}
	}
	delete(st.Active, id)
	delete(st.StartProfile, id)
	delete(st.AddedKey, id)
	return w
}

func (e *Engine) writeStrip(w stripWrite) {
	if w.skip {
		return
	}
	var err error
	if w.removeKey {
		err = w.c.RemoveKey(w.c.ProfilePath)
	} else {
		_, err = w.c.Write(w.c.ProfilePath, w.prof)
	}
	if err != nil {
		e.warnf("aim: %s: strip failed: %v", w.c.ID(), err)
	}
}

// recover finishes a session that ended without an exit step: every change is
// kept, then its host items are stripped.
func (e *Engine) recover(profile, agent string, cols []Collection, st *State) error {
	var mine []Collection
	for _, c := range cols {
		if _, ok := st.Active[c.ID()]; ok {
			mine = append(mine, c)
		}
	}
	changes, err := e.diff(mine, st)
	if err != nil {
		return err
	}
	for _, ch := range changes {
		e.apply(profile, ch, Keep, st)
	}
	if len(changes) > 0 {
		e.warnf("%d change(s) kept in %s (%s) from a session that did not exit through aim", len(changes), profile, agent)
	}
	var writes []stripWrite
	for _, c := range mine {
		writes = append(writes, e.planStrip(c, st))
	}
	if err := e.Store.Save(profile, st); err != nil {
		return err
	}
	for _, w := range writes {
		e.writeStrip(w)
	}
	return nil
}

// abandon simulates a killed session in tests.
func (s *Session) abandon() { _ = s.sessions.Unlock() }
