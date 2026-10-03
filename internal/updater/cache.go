package updater

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// CacheData records the cached update check details.
type CacheData struct {
	LatestVersion string    `json:"latest_version"`
	ReleaseURL    string    `json:"release_url"`
	CheckedAt     time.Time `json:"checked_at"`
}

func readCache(cacheDir string) (*CacheData, error) {
	cachePath := filepath.Join(cacheDir, "update_check.json")
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, err
	}
	var c CacheData
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func writeCache(cacheDir string, c CacheData) error {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	cachePath := filepath.Join(cacheDir, "update_check.json")
	return os.WriteFile(cachePath, data, 0644)
}
