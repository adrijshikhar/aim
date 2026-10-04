package updater

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const DefaultGitHubReleaseEndpoint = "https://api.github.com/repos/adrijshikhar/aim/releases/latest"

// UpdateInfo holds the outcome of a release check.
type UpdateInfo struct {
	CurrentVersion  string    `json:"current_version"`
	LatestVersion   string    `json:"latest_version"`
	UpdateAvailable bool      `json:"update_available"`
	ReleaseURL      string    `json:"release_url"`
	CheckedAt       time.Time `json:"checked_at"`
}

// Checker checks GitHub releases for updates with 24-hour local caching.
type Checker struct {
	Endpoint   string
	HTTPClient *http.Client
}

// NewChecker creates a default Checker targeting official releases.
func NewChecker() *Checker {
	return &Checker{
		Endpoint: DefaultGitHubReleaseEndpoint,
		HTTPClient: &http.Client{
			Timeout: 1 * time.Second,
		},
	}
}

// Check evaluates whether a newer version exists. If cache is fresh (<24h) and not forced,
// it avoids making a network request.
func (c *Checker) Check(ctx context.Context, currentVersion, cacheDir string, force bool) (*UpdateInfo, error) {
	if !force {
		cached, err := readCache(cacheDir)
		if err == nil && cached != nil && time.Since(cached.CheckedAt) < 24*time.Hour {
			newer := IsNewerVersion(cached.LatestVersion, currentVersion)
			return &UpdateInfo{
				CurrentVersion:  currentVersion,
				LatestVersion:   cached.LatestVersion,
				UpdateAvailable: newer,
				ReleaseURL:      cached.ReleaseURL,
				CheckedAt:       cached.CheckedAt,
			}, nil
		}
	}

	endpoint := c.Endpoint
	if endpoint == "" {
		endpoint = DefaultGitHubReleaseEndpoint
	}

	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 1 * time.Second}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return &UpdateInfo{CurrentVersion: currentVersion, UpdateAvailable: false}, nil
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "aim-updater/"+currentVersion)

	resp, err := client.Do(req)
	if err != nil {
		// Network unreachable or timeout: degrade gracefully
		return &UpdateInfo{CurrentVersion: currentVersion, UpdateAvailable: false}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &UpdateInfo{CurrentVersion: currentVersion, UpdateAvailable: false}, nil
	}

	var ghRelease struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ghRelease); err != nil {
		return &UpdateInfo{CurrentVersion: currentVersion, UpdateAvailable: false}, nil
	}

	now := time.Now().UTC()
	_ = writeCache(cacheDir, CacheData{
		LatestVersion: ghRelease.TagName,
		ReleaseURL:    ghRelease.HTMLURL,
		CheckedAt:     now,
	})

	newer := IsNewerVersion(ghRelease.TagName, currentVersion)
	return &UpdateInfo{
		CurrentVersion:  currentVersion,
		LatestVersion:   ghRelease.TagName,
		UpdateAvailable: newer,
		ReleaseURL:      ghRelease.HTMLURL,
		CheckedAt:       now,
	}, nil
}

// CheckForUpdate is a package-level helper that checks for updates.
func CheckForUpdate(ctx context.Context, currentVersion, cacheDir string, force bool) (*UpdateInfo, error) {
	return NewChecker().Check(ctx, currentVersion, cacheDir, force)
}

// CheckCached inspects the local cache without making any network calls.
func CheckCached(currentVersion, cacheDir string) *UpdateInfo {
	cached, err := readCache(cacheDir)
	if err != nil || cached == nil {
		return nil
	}
	newer := IsNewerVersion(cached.LatestVersion, currentVersion)
	return &UpdateInfo{
		CurrentVersion:  currentVersion,
		LatestVersion:   cached.LatestVersion,
		UpdateAvailable: newer,
		ReleaseURL:      cached.ReleaseURL,
		CheckedAt:       cached.CheckedAt,
	}
}

var (
	bgCheckWg sync.WaitGroup
)

// MaybeTriggerBackgroundCheck runs an asynchronous release check if cache is older than 24h.
func MaybeTriggerBackgroundCheck(currentVersion, cacheDir string) {
	cached, err := readCache(cacheDir)
	if err == nil && cached != nil && time.Since(cached.CheckedAt) < 24*time.Hour {
		return
	}
	bgCheckWg.Add(1)
	go func() {
		defer bgCheckWg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_, _ = CheckForUpdate(ctx, currentVersion, cacheDir, true)
	}()
}

// AwaitBackgroundCheck waits up to timeout for any in-flight background check to finish.
func AwaitBackgroundCheck(timeout time.Duration) {
	done := make(chan struct{}, 1)
	go func() {
		bgCheckWg.Wait()
		done <- struct{}{}
	}()
	select {
	case <-done:
	case <-time.After(timeout):
	}
}

// IsNewerVersion returns true if latest is strictly higher than current according to semver.
// If current is "dev", returns false.
func IsNewerVersion(latest, current string) bool {
	current = strings.TrimSpace(current)
	latest = strings.TrimSpace(latest)

	if current == "dev" || current == "" {
		return false
	}

	latestParts := parseSemver(latest)
	currentParts := parseSemver(current)

	if latestParts == nil || currentParts == nil {
		return false
	}

	for i := 0; i < 3; i++ {
		if latestParts[i] > currentParts[i] {
			return true
		}
		if latestParts[i] < currentParts[i] {
			return false
		}
	}

	return false
}

func parseSemver(v string) []int {
	v = strings.TrimPrefix(v, "v")
	// Strip any prerelease or build metadata (-beta, +build)
	if idx := strings.IndexAny(v, "-+"); idx != -1 {
		v = v[:idx]
	}
	parts := strings.Split(v, ".")
	if len(parts) < 3 {
		return nil
	}
	res := make([]int, 3)
	for i := 0; i < 3; i++ {
		n, err := strconv.Atoi(parts[i])
		if err != nil {
			return nil
		}
		res[i] = n
	}
	return res
}
