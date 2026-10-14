package oauth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// RefusalExpiry defines the backoff window (24 hours) after which a fatal OAuth
// refusal expires, allowing background daemons and usage engines to retry auto-refresh.
const RefusalExpiry = 24 * time.Hour

// RefusalRecord represents a persisted refusal event for an agent profile.
type RefusalRecord struct {
	Agent       string      `json:"agent"`
	Profile     string      `json:"profile"`
	Code        RefusalCode `json:"code"`
	Description string      `json:"description"`
	RecordedAt  time.Time   `json:"recorded_at"`
}

// RefusalStore manages persistent storage of OAuth refusal records in ~/.aim/refusals.json.
// Mutating operations are synchronized with a Mutex and file writes are performed atomically.
type RefusalStore struct {
	filePath string
	mu       sync.Mutex
}

// NewRefusalStore returns a RefusalStore rooted in baseDir (storing records in baseDir/refusals.json).
func NewRefusalStore(baseDir string) *RefusalStore {
	return &RefusalStore{
		filePath: filepath.Join(baseDir, "refusals.json"),
	}
}

// FilePath returns the absolute path to the underlying refusals.json file.
func (s *RefusalStore) FilePath() string {
	return s.filePath
}

func refusalKey(agent, profile string) string {
	return strings.ToLower(strings.TrimSpace(agent)) + ":" + strings.TrimSpace(profile)
}

// Record persists an OAuth refusal event for the given agent and profile.
// It also cleans up expired refusal entries to keep storage compact.
func (s *RefusalStore) Record(agent, profile string, code RefusalCode, desc string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	records := s.loadRecords()

	// Clean expired entries on write
	for k, r := range records {
		if time.Since(r.RecordedAt) > RefusalExpiry {
			delete(records, k)
		}
	}

	key := refusalKey(agent, profile)
	records[key] = RefusalRecord{
		Agent:       agent,
		Profile:     profile,
		Code:        code,
		Description: desc,
		RecordedAt:  time.Now().UTC(),
	}

	return s.saveRecords(records)
}

// IsRefused checks whether the agent and profile currently have an active (non-expired)
// refusal record. If a record exists but has passed the 24-hour expiry window,
// it returns nil, false. Reads are lockless and read directly from disk via atomic snapshots.
func (s *RefusalStore) IsRefused(agent, profile string) (*RefusalRecord, bool) {
	records := s.loadRecords()
	key := refusalKey(agent, profile)
	rec, ok := records[key]
	if !ok {
		return nil, false
	}

	// 24-hour backoff check
	if time.Since(rec.RecordedAt) > RefusalExpiry {
		return nil, false
	}

	return &rec, true
}

// Clear removes any refusal record for the specified agent and profile.
// If the key does not exist, it avoids unnecessary disk writes.
func (s *RefusalStore) Clear(agent, profile string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	records := s.loadRecords()
	key := refusalKey(agent, profile)
	if _, ok := records[key]; !ok {
		return nil
	}
	delete(records, key)

	return s.saveRecords(records)
}

// Prune purges all expired refusal records from disk. If maxAge is provided and > 0,
// it uses maxAge; otherwise it defaults to RefusalExpiry.
func (s *RefusalStore) Prune(maxAge ...time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	expiry := RefusalExpiry
	if len(maxAge) > 0 && maxAge[0] > 0 {
		expiry = maxAge[0]
	}

	records := s.loadRecords()
	changed := false
	for k, r := range records {
		if time.Since(r.RecordedAt) > expiry {
			delete(records, k)
			changed = true
		}
	}

	if !changed {
		return nil
	}

	return s.saveRecords(records)
}

func (s *RefusalStore) loadRecords() map[string]RefusalRecord {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return make(map[string]RefusalRecord)
	}

	var records map[string]RefusalRecord
	if err := json.Unmarshal(data, &records); err != nil || records == nil {
		return make(map[string]RefusalRecord)
	}

	return records
}

func (s *RefusalStore) saveRecords(records map[string]RefusalRecord) error {
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create refusal store directory: %w", err)
	}
	_ = os.Chmod(dir, 0700)

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode refusal records: %w", err)
	}

	tmpFile, err := os.CreateTemp(dir, "refusals-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temporary refusal file: %w", err)
	}
	tmpName := tmpFile.Name()

	// Ensure 0600 permissions
	_ = tmpFile.Chmod(0600)

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("failed to write temporary refusal file: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("failed to sync temporary refusal file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("failed to close temporary refusal file: %w", err)
	}

	if err := os.Rename(tmpName, s.filePath); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("failed to commit refusal file: %w", err)
	}

	_ = os.Chmod(s.filePath, 0600)
	return nil
}
