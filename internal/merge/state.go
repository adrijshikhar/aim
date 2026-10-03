package merge

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ChangeKind string

const (
	Added   ChangeKind = "added"
	Edited  ChangeKind = "edited"
	Removed ChangeKind = "removed"
)

// State is one profile's merge state (spec §5).
type State struct {
	Version int `json:"version"`
	// Active: collection ID → host item name → hash of the value merged in.
	Active map[string]map[string]string `json:"active,omitempty"`
	// StartProfile: collection ID → the profile's own items when the first
	// running session of that agent started.
	StartProfile map[string]map[string]string `json:"start_profile,omitempty"`
	Migrated     map[string]time.Time         `json:"migrated,omitempty"`
	// AddedKey marks collections whose top-level key the merge created; strip
	// removes the key again when it ends empty.
	AddedKey map[string]bool `json:"added_key,omitempty"`
}

func newState() *State {
	return &State{Version: 1, Active: map[string]map[string]string{}, StartProfile: map[string]map[string]string{},
		Migrated: map[string]time.Time{}, AddedKey: map[string]bool{}}
}

// Store keeps per-profile merge state and lock files in Dir (StateDir()/profile-merge).
type Store struct{ Dir string }

func (s Store) path(p string) string { return filepath.Join(s.Dir, p+".json") }

// MergeLockPath is held exclusively, briefly, around every merge, diff, apply and strip.
func (s Store) MergeLockPath(p string) string { return filepath.Join(s.Dir, p+".lock") }

// SessionsLockPath is held shared by every running foreground session of one
// agent in one profile; a session is the last one when it can take it exclusively.
func (s Store) SessionsLockPath(p, agent string) string {
	return filepath.Join(s.Dir, p+"."+agent+".sessions")
}

func (s Store) HostLockPath() string { return filepath.Join(s.Dir, "host.lock") }

func (s Store) Load(profile string) (*State, error) {
	data, err := os.ReadFile(s.path(profile))
	if errors.Is(err, os.ErrNotExist) {
		return newState(), nil
	}
	if err != nil {
		return nil, err
	}
	st := newState()
	if err := json.Unmarshal(data, st); err != nil {
		return nil, err
	}
	return st, nil
}

func (s Store) Save(profile string, st *State) error {
	if err := os.MkdirAll(s.Dir, 0o700); err != nil {
		return err
	}
	_ = os.Chmod(s.Dir, 0o700)
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return AtomicWrite(s.path(profile), b, 0o600)
}

func (s Store) Remove(profile string) error {
	err := os.Remove(s.path(profile))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (s Store) Rename(from, to string) error {
	err := os.Rename(s.path(from), s.path(to))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

// ActiveSessions returns the agents with a running foreground session in the
// profile: those whose sessions lock cannot be taken exclusively. The probe
// lock is released at once.
func (s Store) ActiveSessions(profile string) ([]string, error) {
	entries, err := os.ReadDir(s.Dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var running []string
	for _, e := range entries {
		agent, ok := strings.CutPrefix(e.Name(), profile+".")
		agent, ok2 := strings.CutSuffix(agent, ".sessions")
		// Agent names hold no dot, so "work.x.claude.sessions" is profile work.x.
		if !ok || !ok2 || agent == "" || strings.Contains(agent, ".") {
			continue
		}
		l, free, err := TryUpgrade(filepath.Join(s.Dir, e.Name()))
		if err != nil {
			return nil, err
		}
		if !free {
			running = append(running, agent)
			continue
		}
		_ = l.Unlock()
	}
	return running, nil
}
