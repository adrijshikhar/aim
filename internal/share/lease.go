package share

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
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
	targetDir := filepath.Join(m.profilesDir, localProfileName)

	cleanTargetDir := filepath.Clean(targetDir)
	cleanProfilesRoot := filepath.Clean(m.profilesDir)
	if !strings.HasPrefix(cleanTargetDir, cleanProfilesRoot+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid profile target path traversal: %s", localProfileName)
	}

	if err := os.MkdirAll(cleanTargetDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create ephemeral profile dir: %w", err)
	}

	allowedPrefixes := []string{
		".claude",
		".codex",
		".gemini",
		".config/codex",
		".config/gcloud",
	}

	type stagedFile struct {
		fullPath string
		content  []byte
	}
	var staged []stagedFile

	// Validate all payload files before modifying disk or writing metadata
	if resp.Payload != nil && resp.Payload.Files != nil {
		for relPath, b64 := range resp.Payload.Files {
			cleanRel := filepath.Clean(relPath)
			if filepath.IsAbs(cleanRel) || strings.HasPrefix(cleanRel, "..") {
				return "", fmt.Errorf("illegal relative path in payload: %s", relPath)
			}
			slashRel := filepath.ToSlash(cleanRel)
			isAllowed := false
			for _, prefix := range allowedPrefixes {
				if strings.HasPrefix(slashRel, prefix) {
					isAllowed = true
					break
				}
			}
			if !isAllowed {
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
			staged = append(staged, stagedFile{fullPath: fullPath, content: decoded})
		}
	}

	// Write credential files
	for _, f := range staged {
		if err := os.MkdirAll(filepath.Dir(f.fullPath), 0700); err != nil {
			return "", fmt.Errorf("failed to create dir for %s: %w", f.fullPath, err)
		}
		if err := os.WriteFile(f.fullPath, f.content, 0600); err != nil {
			return "", fmt.Errorf("failed to write %s: %w", f.fullPath, err)
		}
	}

	// Publish metadata ONLY after all files are verified and written
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
	tmpMeta := filepath.Join(cleanTargetDir, fmt.Sprintf(".aim-lease.tmp.%d", time.Now().UnixNano()))
	if err := os.WriteFile(tmpMeta, metaData, 0600); err != nil {
		return "", fmt.Errorf("failed to write lease metadata: %w", err)
	}
	if err := os.Rename(tmpMeta, filepath.Join(cleanTargetDir, ".aim-lease.json")); err != nil {
		return "", fmt.Errorf("failed to publish lease metadata: %w", err)
	}

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

	// Strictly require .aim-lease.json so permanent non-leased profiles cannot be removed
	leaseFile := filepath.Join(cleanTarget, ".aim-lease.json")
	if _, err := os.Stat(leaseFile); os.IsNotExist(err) {
		return fmt.Errorf("profile %q is not an ephemeral leased profile (missing .aim-lease.json)", localProfileName)
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
			if err := m.ReleaseLease(e.Name()); err == nil {
				cleaned++
			}
		}
	}
	return cleaned, nil
}
