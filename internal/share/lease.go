package share

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aim-cli/aim/internal/merge"
)

var (
	leaseOpMu sync.Mutex
)

type LeaseMeta struct {
	LeaseID       string    `json:"lease_id"`
	SourcePeer    string    `json:"source_peer"`
	SourceProfile string    `json:"source_profile"`
	Agent         string    `json:"agent"`
	GrantedAt     time.Time `json:"granted_at"`
	ExpiresAt     time.Time `json:"expires_at"`
}

type LeaseManager struct {
	profilesDir string
}

func NewLeaseManager(profilesDir string) *LeaseManager {
	return &LeaseManager{profilesDir: profilesDir}
}

func (m *LeaseManager) InstallLease(sourcePeer, profileName string, resp *LeaseResponse) (string, error) {
	// Format as host:<profile_name> (e.g., Adrijs-MacBook-Pro:work)
	cleanPeer := strings.TrimSuffix(strings.TrimSuffix(sourcePeer, ".local"), ".LOCAL")
	sanitizedPeer := strings.ReplaceAll(cleanPeer, ".", "-")
	sanitizedPeer = strings.ReplaceAll(sanitizedPeer, "/", "-")
	sanitizedPeer = strings.ReplaceAll(sanitizedPeer, "\\", "-")
	sanitizedProfile := strings.ReplaceAll(profileName, "/", "-")
	sanitizedProfile = strings.ReplaceAll(sanitizedProfile, "\\", "-")
	localProfileName := fmt.Sprintf("%s:%s", sanitizedPeer, sanitizedProfile)
	if sanitizedPeer == "" {
		localProfileName = sanitizedProfile
	}
	targetDir := filepath.Join(m.profilesDir, localProfileName)

	cleanTargetDir := filepath.Clean(targetDir)
	cleanProfilesRoot := filepath.Clean(m.profilesDir)
	if !strings.HasPrefix(cleanTargetDir, cleanProfilesRoot+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid profile target path traversal: %s", localProfileName)
	}

	leaseOpMu.Lock()
	defer leaseOpMu.Unlock()

	lockName := fmt.Sprintf(".lease-%s.lock", strings.ReplaceAll(localProfileName, ":", "-"))
	lockPath := filepath.Join(cleanProfilesRoot, lockName)
	lk, err := merge.LockExclusive(lockPath, 10*time.Second)
	if err != nil {
		return "", fmt.Errorf("failed to acquire lease lock for %s: %w", localProfileName, err)
	}
	defer lk.Unlock()

	// Protect existing permanent user profiles from being overwritten or deleted as leases
	if info, err := os.Stat(cleanTargetDir); err == nil && info.IsDir() {
		leaseFile := filepath.Join(cleanTargetDir, ".aim-lease.json")
		if _, err := os.Stat(leaseFile); os.IsNotExist(err) {
			entries, _ := os.ReadDir(cleanTargetDir)
			if len(entries) > 0 {
				return "", fmt.Errorf("target profile %q already exists and is not a leased profile", localProfileName)
			}
		}
	}

	allowedCredentialFiles := map[string]bool{
		".claude/.credentials.json":                           true,
		".claude/auth.json":                                   true,
		".claude.json":                                        true,
		".claude/.claude.json":                                true,
		".codex/auth.json":                                    true,
		".codex/config.toml":                                  true,
		".config/codex/auth.json":                             true,
		".gemini/antigravity-cli/antigravity-oauth-token":     true,
		".gemini/antigravity-cli/config.json":                 true,
		".config/gcloud/application_default_credentials.json": true,
		".gemini/gemini-oauth-token":                          true,
		".gemini/oauth.json":                                  true,
	}

	type stagedFile struct {
		relPath string
		content []byte
	}
	var staged []stagedFile

	// Validate all payload files before modifying disk or creating directories
	if resp.Payload != nil && resp.Payload.Files != nil {
		for relPath, b64 := range resp.Payload.Files {
			cleanRel := filepath.Clean(relPath)
			if filepath.IsAbs(cleanRel) || strings.HasPrefix(cleanRel, "..") {
				return "", fmt.Errorf("illegal relative path in payload: %s", relPath)
			}
			slashRel := filepath.ToSlash(cleanRel)
			if !allowedCredentialFiles[slashRel] {
				return "", fmt.Errorf("disallowed credential path in payload: %s", relPath)
			}

			fullPath := filepath.Join(cleanTargetDir, cleanRel)
			if !strings.HasPrefix(filepath.Clean(fullPath), cleanTargetDir+string(filepath.Separator)) {
				return "", fmt.Errorf("path traversal in credential path: %s", relPath)
			}

			decoded, err := base64.StdEncoding.DecodeString(b64)
			if err != nil {
				return "", fmt.Errorf("invalid base64 encoding in file %s: %w", relPath, err)
			}
			staged = append(staged, stagedFile{relPath: slashRel, content: decoded})
		}
	}

	meta := LeaseMeta{
		LeaseID:       resp.LeaseID,
		SourcePeer:    sourcePeer,
		SourceProfile: profileName,
		Agent:         resp.Agent,
		GrantedAt:     time.Now(),
		ExpiresAt:     resp.ExpiresAt,
	}
	metaData, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to encode lease metadata: %w", err)
	}

	// Check if this is a renewal of an existing lease
	existingLease := filepath.Join(cleanTargetDir, ".aim-lease.json")
	if _, err := os.Stat(existingLease); err == nil {
		// Renewal: update credential files in-place and refresh lease metadata,
		// preserving the borrower's local conversation history, databases, and custom settings.
		type fileBackup struct {
			path    string
			bakPath string
			existed bool
		}
		type stagedItem struct {
			destPath string
			tmpPath  string
		}

		var backups []fileBackup
		var stagedItems []stagedItem
		var createdPaths []string

		trackTemp := func(p string) {
			createdPaths = append(createdPaths, p)
		}

		cleanTemps := func() {
			for _, p := range createdPaths {
				_ = os.Remove(p)
			}
		}

		nowNano := time.Now().UnixNano()
		pid := os.Getpid()

		// 1a. Stage credential files and create backups of existing files
		for _, f := range staged {
			destPath := filepath.Join(cleanTargetDir, filepath.FromSlash(f.relPath))
			if err := os.MkdirAll(filepath.Dir(destPath), 0700); err != nil {
				cleanTemps()
				return "", fmt.Errorf("failed to create dir for %s: %w", destPath, err)
			}

			if _, err := os.Stat(destPath); err == nil {
				oldData, err := os.ReadFile(destPath)
				if err != nil {
					cleanTemps()
					return "", fmt.Errorf("failed to read existing file for backup %s: %w", destPath, err)
				}
				bakPath := fmt.Sprintf("%s.bak.%d.%d", destPath, pid, nowNano)
				trackTemp(bakPath)
				if err := os.WriteFile(bakPath, oldData, 0600); err != nil {
					cleanTemps()
					return "", fmt.Errorf("failed to backup existing file %s: %w", destPath, err)
				}
				backups = append(backups, fileBackup{path: destPath, bakPath: bakPath, existed: true})
			} else {
				backups = append(backups, fileBackup{path: destPath, existed: false})
			}

			tmpFile := fmt.Sprintf("%s.tmp.%d.%d", destPath, pid, nowNano)
			trackTemp(tmpFile)
			if err := os.WriteFile(tmpFile, f.content, 0600); err != nil {
				cleanTemps()
				return "", fmt.Errorf("failed to stage write %s: %w", destPath, err)
			}
			stagedItems = append(stagedItems, stagedItem{destPath: destPath, tmpPath: tmpFile})
		}

		// 1b. Stage lease metadata and create backup of existing metadata
		oldMeta, err := os.ReadFile(existingLease)
		if err != nil {
			cleanTemps()
			return "", fmt.Errorf("failed to read existing lease metadata for backup: %w", err)
		}
		bakMeta := fmt.Sprintf("%s.bak.%d.%d", existingLease, pid, nowNano)
		trackTemp(bakMeta)
		if err := os.WriteFile(bakMeta, oldMeta, 0600); err != nil {
			cleanTemps()
			return "", fmt.Errorf("failed to backup existing lease metadata: %w", err)
		}
		backups = append(backups, fileBackup{path: existingLease, bakPath: bakMeta, existed: true})

		tmpMeta := fmt.Sprintf("%s.tmp.%d.%d", existingLease, pid, nowNano)
		trackTemp(tmpMeta)
		if err := os.WriteFile(tmpMeta, metaData, 0600); err != nil {
			cleanTemps()
			return "", fmt.Errorf("failed to stage write lease metadata: %w", err)
		}
		stagedItems = append(stagedItems, stagedItem{destPath: existingLease, tmpPath: tmpMeta})

		// Stage 2: Atomic commit of all staged items.
		// If any rename fails, rollback restores every file from its backup.
		rollback := func() {
			for _, item := range stagedItems {
				_ = os.Remove(item.tmpPath)
			}
			for _, b := range backups {
				if b.existed {
					_ = os.Rename(b.bakPath, b.path)
				} else {
					_ = os.Remove(b.path)
				}
			}
		}

		for _, item := range stagedItems {
			if err := os.Rename(item.tmpPath, item.destPath); err != nil {
				rollback()
				return "", fmt.Errorf("failed to commit renewal file %s: %w", item.destPath, err)
			}
		}

		// Success! Clean up backup files
		for _, b := range backups {
			if b.existed {
				_ = os.Remove(b.bakPath)
			}
		}

		return localProfileName, nil
	}

	// First install: write into a clean staging directory first so failed/interrupted installs leave no partial directory behind
	stagingDir := fmt.Sprintf("%s.tmp.%d.%d", cleanTargetDir, os.Getpid(), time.Now().UnixNano())
	if err := os.MkdirAll(stagingDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create staging dir: %w", err)
	}
	success := false
	defer func() {
		if !success {
			_ = os.RemoveAll(stagingDir)
		}
	}()

	// Write credential files into staging directory
	for _, f := range staged {
		destPath := filepath.Join(stagingDir, filepath.FromSlash(f.relPath))
		if err := os.MkdirAll(filepath.Dir(destPath), 0700); err != nil {
			return "", fmt.Errorf("failed to create dir for %s: %w", destPath, err)
		}
		if err := os.WriteFile(destPath, f.content, 0600); err != nil {
			return "", fmt.Errorf("failed to write %s: %w", destPath, err)
		}
	}

	// Write metadata into staging directory
	if err := os.WriteFile(filepath.Join(stagingDir, ".aim-lease.json"), metaData, 0600); err != nil {
		return "", fmt.Errorf("failed to write lease metadata: %w", err)
	}

	// Atomically replace target directory
	if _, err := os.Stat(cleanTargetDir); err == nil {
		_ = os.RemoveAll(cleanTargetDir)
	}
	if err := os.Rename(stagingDir, cleanTargetDir); err != nil {
		return "", fmt.Errorf("failed to publish lease: %w", err)
	}
	success = true

	return localProfileName, nil
}

func (m *LeaseManager) HasLease(localProfileName string) bool {
	metaFile := filepath.Join(m.profilesDir, localProfileName, ".aim-lease.json")
	_, err := os.Stat(metaFile)
	return err == nil
}

func (m *LeaseManager) IsLeaseExpired(localProfileName string) bool {
	metaFile := filepath.Join(m.profilesDir, localProfileName, ".aim-lease.json")
	data, err := os.ReadFile(metaFile)
	if err != nil {
		return false
	}
	var meta LeaseMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return false
	}
	return time.Now().After(meta.ExpiresAt)
}

func (m *LeaseManager) ReleaseLease(localProfileName string) error {
	return m.releaseLeaseInternal(localProfileName, false)
}

func (m *LeaseManager) releaseLeaseInternal(localProfileName string, onlyIfExpired bool) error {
	if localProfileName == "" || localProfileName == "." || localProfileName == ".." ||
		strings.Contains(localProfileName, "/") || strings.Contains(localProfileName, "\\") {
		return fmt.Errorf("invalid profile name: %q", localProfileName)
	}

	targetDir := filepath.Join(m.profilesDir, localProfileName)
	cleanTarget := filepath.Clean(targetDir)
	cleanProfilesRoot := filepath.Clean(m.profilesDir)
	if !strings.HasPrefix(cleanTarget, cleanProfilesRoot+string(filepath.Separator)) {
		return fmt.Errorf("profile path escapes profiles root: %q", localProfileName)
	}

	leaseOpMu.Lock()
	defer leaseOpMu.Unlock()

	lockName := fmt.Sprintf(".lease-%s.lock", strings.ReplaceAll(localProfileName, ":", "-"))
	lockPath := filepath.Join(cleanProfilesRoot, lockName)
	lk, err := merge.LockExclusive(lockPath, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to acquire lease lock for %s: %w", localProfileName, err)
	}
	defer lk.Unlock()

	// Strictly require .aim-lease.json so permanent non-leased profiles cannot be removed
	leaseFile := filepath.Join(cleanTarget, ".aim-lease.json")
	if _, err := os.Stat(leaseFile); os.IsNotExist(err) {
		return fmt.Errorf("profile %q is not an ephemeral leased profile (missing .aim-lease.json)", localProfileName)
	}

	if onlyIfExpired {
		data, err := os.ReadFile(leaseFile)
		if err != nil {
			return err
		}
		var meta LeaseMeta
		if err := json.Unmarshal(data, &meta); err != nil {
			return err
		}
		if !time.Now().After(meta.ExpiresAt) {
			return fmt.Errorf("lease %q is not expired", localProfileName)
		}
	}

	return os.RemoveAll(cleanTarget)
}

func (m *LeaseManager) CleanExpiredLeases() (int, error) {
	entries, err := os.ReadDir(m.profilesDir)
	if err != nil {
		return 0, err
	}
	cleaned := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if m.IsLeaseExpired(e.Name()) {
			if err := m.releaseLeaseInternal(e.Name(), true); err == nil {
				cleaned++
			}
		}
	}
	return cleaned, nil
}
