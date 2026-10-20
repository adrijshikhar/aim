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
	// Sanitize sourcePeer for clean directory naming: strip .local and replace dots/colons
	cleanPeer := strings.TrimSuffix(strings.TrimSuffix(sourcePeer, ".local"), ".LOCAL")
	sanitizedPeer := strings.ReplaceAll(strings.ReplaceAll(cleanPeer, ".", "-"), ":", "-")
	localProfileName := fmt.Sprintf("p2p-%s-%s", sanitizedPeer, profileName)
	targetDir := filepath.Join(m.profilesDir, localProfileName)

	if err := os.MkdirAll(targetDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create ephemeral profile dir: %w", err)
	}

	// Write metadata
	meta := LeaseMeta{
		LeaseID:       resp.LeaseID,
		SourcePeer:    sourcePeer,
		SourceProfile: profileName,
		Agent:         resp.Agent,
		GrantedAt:     time.Now(),
		ExpiresAt:     resp.ExpiresAt,
	}
	metaData, err := json.MarshalIndent(meta, "", "  ")
	if err == nil {
		_ = os.WriteFile(filepath.Join(targetDir, ".aim-lease.json"), metaData, 0600)
	}

	// Write credential files
	if resp.Payload != nil && resp.Payload.Files != nil {
		for relPath, b64 := range resp.Payload.Files {
			fullPath := filepath.Join(targetDir, relPath)
			if err := os.MkdirAll(filepath.Dir(fullPath), 0700); err != nil {
				return "", fmt.Errorf("failed to create dir for %s: %w", relPath, err)
			}
			decoded, err := base64.StdEncoding.DecodeString(b64)
			if err != nil {
				continue
			}
			if err := os.WriteFile(fullPath, decoded, 0600); err != nil {
				return "", fmt.Errorf("failed to write %s: %w", relPath, err)
			}
		}
	}

	return localProfileName, nil
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
	targetDir := filepath.Join(m.profilesDir, localProfileName)
	return os.RemoveAll(targetDir)
}
