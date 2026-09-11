package usage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type diskCacheData struct {
	Version   string            `json:"version"`
	UpdatedAt time.Time         `json:"updated_at"`
	Reports   map[string]Report `json:"reports"`
}

const DefaultTTL = 15 * time.Minute

type CacheStore struct {
	mu       sync.RWMutex
	filePath string
	ttl      time.Duration
	reports  map[string]Report
}

func NewCacheStore(baseDir string, ttl time.Duration) *CacheStore {
	filePath := filepath.Join(baseDir, "cache", "usage.json")
	store := &CacheStore{
		filePath: filePath,
		ttl:      ttl,
		reports:  make(map[string]Report),
	}
	store.load()
	return store
}

func key(agent, profile string) string {
	return fmt.Sprintf("%s:%s", agent, profile)
}

func (c *CacheStore) load() {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := os.ReadFile(c.filePath)
	if err != nil {
		return
	}

	var d diskCacheData
	if err := json.Unmarshal(data, &d); err != nil {
		return
	}
	if d.Reports != nil {
		c.reports = d.Reports
	}
}

func (c *CacheStore) Get(agent, profile string) (Report, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	rep, ok := c.reports[key(agent, profile)]
	if !ok {
		return Report{}, false
	}

	if time.Since(rep.FetchedAt) > c.ttl {
		return Report{}, false
	}

	// Deep-copy Windows to avoid mutating shared slice backing array under RLock
	if len(rep.Windows) > 0 {
		windows := make([]LimitWindow, len(rep.Windows))
		copy(windows, rep.Windows)
		for i := range windows {
			if !windows[i].ResetsAt.IsZero() {
				rem := time.Until(windows[i].ResetsAt)
				if rem < 0 {
					rem = 0
				}
				windows[i].ResetsIn = rem
			}
		}
		rep.Windows = windows
	}
	rep.FromCache = true
	return rep, true
}

func (c *CacheStore) Put(report Report) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if report.FetchedAt.IsZero() {
		report.FetchedAt = time.Now()
	}
	report.FromCache = false

	if len(report.Windows) > 0 {
		windows := make([]LimitWindow, len(report.Windows))
		copy(windows, report.Windows)
		report.Windows = windows
	}

	if c.reports == nil {
		c.reports = make(map[string]Report)
	}
	c.reports[key(report.Agent, report.Profile)] = report

	return c.saveAtomic()
}

func (c *CacheStore) Flush(agent string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if agent == "" {
		c.reports = make(map[string]Report)
	} else {
		prefix := agent + ":"
		for k := range c.reports {
			if strings.HasPrefix(k, prefix) {
				delete(c.reports, k)
			}
		}
	}
	_ = c.saveAtomic()
}

// Delete removes a specific profile report from the cache.
func (c *CacheStore) Delete(agent, profile string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.reports != nil {
		delete(c.reports, key(agent, profile))
	}
	_ = c.saveAtomic()
}

func (c *CacheStore) saveAtomic() error {
	dir := filepath.Dir(c.filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	d := diskCacheData{
		Version:   "1",
		UpdatedAt: time.Now(),
		Reports:   c.reports,
	}

	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}

	tmpFile := c.filePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0600); err != nil {
		return err
	}

	if err := os.Rename(tmpFile, c.filePath); err != nil {
		_ = os.Remove(tmpFile)
		return err
	}
	return nil
}

// IsStale returns true if the cached report for agent:profile is missing or older than maxAge.
func (c *CacheStore) IsStale(agent, profile string, maxAge time.Duration) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	rep, ok := c.reports[key(agent, profile)]
	if !ok {
		return true
	}
	return time.Since(rep.FetchedAt) > maxAge
}

// CanPrewarm checks whether a background prewarm should be permitted based on a cooldown lock.
// It checks <baseDir>/cache/.prewarm.lock. If the lock was updated within cooldown, it returns false.
// Otherwise, it updates the lock file timestamp and returns true.
func CanPrewarm(baseDir string, cooldown time.Duration) bool {
	lockFile := filepath.Join(baseDir, "cache", ".prewarm.lock")
	dir := filepath.Dir(lockFile)
	_ = os.MkdirAll(dir, 0700)

	if cooldown > 0 {
		fi, err := os.Stat(lockFile)
		if err == nil {
			if time.Since(fi.ModTime()) < cooldown {
				return false
			}
		}
	}

	// Touch/update lock file
	now := time.Now()
	if err := os.WriteFile(lockFile, []byte(now.Format(time.RFC3339)), 0600); err != nil {
		return false
	}
	_ = os.Chtimes(lockFile, now, now)
	return true
}
