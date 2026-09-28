package merge

import (
	"errors"
	"os"

	toml "github.com/pelletier/go-toml/v2"
)

type Format int

const (
	JSON Format = iota
	TOML
)

// Collection is one named map inside an agent's config file (spec §4).
type Collection struct {
	Agent, Name string
	HostPath    string
	ProfilePath string
	Format      Format
	Key         string
	Normalise   Normaliser
}

func (c Collection) ID() string { return c.Agent + "/" + c.Name }

func (c Collection) norm() Normaliser {
	if c.Normalise != nil {
		return c.Normalise
	}
	return DropEmpty
}

func (c Collection) Read(path string) (Entries, bool, error) {
	if c.Format == TOML {
		return ReadTOMLKey(path, c.Key)
	}
	return ReadJSONKey(path, c.Key)
}

// Write rewrites the collection's key; the file is never left looser than 0600.
func (c Collection) Write(path string, e Entries) ([]string, error) {
	if c.Format == TOML {
		return WriteTOMLKey(path, c.Key, e, 0o600)
	}
	return nil, WriteJSONKey(path, c.Key, e, os.FileMode(0o600))
}

// HasKey reports whether the file defines the collection's key at all.
func (c Collection) HasKey(path string) (bool, error) {
	if c.Format == JSON {
		return JSONHasKey(path, c.Key)
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	var doc map[string]any
	if err := toml.Unmarshal(data, &doc); err != nil {
		return false, err
	}
	_, ok := doc[c.Key]
	return ok, nil
}

// RemoveKey deletes the (empty) collection key. A TOML table disappears with
// its last block, so only JSON needs a key delete.
func (c Collection) RemoveKey(path string) error {
	if c.Format == JSON {
		return DeleteJSONKey(path, c.Key, 0o600)
	}
	_, err := WriteTOMLKey(path, c.Key, NewEntries(), 0o600)
	return err
}
