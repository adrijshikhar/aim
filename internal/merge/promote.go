package merge

import (
	"errors"
	"os"
	"path/filepath"
)

var errHostChanged = errors.New("host changed since the session started")

// hostTarget resolves a symlinked host file (stow, chezmoi) so the write
// replaces the link's target, not the link. A missing host is written at path;
// a dangling link is refused.
func hostTarget(path string) (string, error) {
	p, err := filepath.EvalSymlinks(path)
	if !errors.Is(err, os.ErrNotExist) {
		return p, err
	}
	if fi, lerr := os.Lstat(path); lerr == nil && fi.Mode()&os.ModeSymlink != 0 {
		return "", errors.New(path + " is a dangling symlink")
	}
	return path, nil
}

// promote writes one change to the host file through the collection's helper
// (never a native command, so no secret reaches argv), under the host lock. It
// refuses when the host changed since the session started, and confirms the
// write by re-reading the host.
func (e *Engine) promote(ch Change, st *State) error {
	c := ch.Collection
	id := c.ID()
	hl, err := LockExclusive(e.Store.HostLockPath(), lockTimeout)
	if err != nil {
		return err
	}
	defer hl.Unlock()
	path, err := hostTarget(c.HostPath)
	if err != nil {
		return err
	}
	host, _, err := c.Read(path)
	if err != nil {
		return err
	}
	norm := c.norm()
	cur := ""
	if host.Has(ch.Name) {
		cur = Hash(norm, host.Values[ch.Name])
	}
	switch ch.Kind {
	case Removed, Edited:
		if cur != st.Active[id][ch.Name] {
			return errHostChanged
		}
	case Added:
		if host.Has(ch.Name) && cur != Hash(norm, ch.Value) {
			return errHostChanged
		}
	}
	if ch.Kind == Removed {
		host.Delete(ch.Name)
	} else {
		host.Set(ch.Name, ch.Value, ch.Raw)
	}
	if _, err := c.Write(path, host); err != nil {
		return err
	}
	host, _, err = c.Read(path)
	if err != nil {
		return err
	}
	if ch.Kind == Removed {
		if host.Has(ch.Name) {
			return errors.New("host still has the item")
		}
		delete(st.Active[id], ch.Name)
		return nil
	}
	want := Hash(norm, ch.Value)
	if !host.Has(ch.Name) || Hash(norm, host.Values[ch.Name]) != want {
		return errors.New("host did not accept the promoted value")
	}
	// The profile copy now equals a host item: strip removes it at rest.
	setHash(st.Active, id, ch.Name, want)
	delete(st.StartProfile[id], ch.Name)
	return nil
}
