package oauth

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// StashedCredential represents an OAuth token pair held in local staging
// before being committed to profile storage or macOS Keychain.
type StashedCredential struct {
	SpentTokenHash string    `json:"spent_token_hash"`
	Agent          string    `json:"agent"`
	Profile        string    `json:"profile"`
	AccessToken    string    `json:"access_token"`
	RefreshToken   string    `json:"refresh_token"`
	ExpiresAt      time.Time `json:"expires_at"`
	CreatedAt      time.Time `json:"created_at"`
}

// StashStore manages unclaimed OAuth credentials in ~/.aim/unclaimed/.
type StashStore struct {
	Dir string
}

// NewStashStore initializes a StashStore rooted in the specified base directory.
func NewStashStore(baseDir string) *StashStore {
	return &StashStore{
		Dir: filepath.Join(baseDir, "unclaimed"),
	}
}

// HashToken returns the SHA-256 hexadecimal digest of a token string.
func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// StashPath returns the path to the stash file for a given spent refresh token.
func (s *StashStore) StashPath(spentToken string) string {
	return filepath.Join(s.Dir, HashToken(spentToken)+".json")
}

// Save atomically writes an uncommitted token pair to the stash directory.
func (s *StashStore) Save(spentToken, agent, profile, accessToken, refreshToken string, expiresAt time.Time) error {
	if err := os.MkdirAll(s.Dir, 0700); err != nil {
		return fmt.Errorf("failed to create unclaimed directory: %w", err)
	}

	cred := StashedCredential{
		SpentTokenHash: HashToken(spentToken),
		Agent:          agent,
		Profile:        profile,
		AccessToken:    accessToken,
		RefreshToken:   refreshToken,
		ExpiresAt:      expiresAt,
		CreatedAt:      time.Now().UTC(),
	}

	data, err := json.MarshalIndent(cred, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode stashed credential: %w", err)
	}

	targetPath := s.StashPath(spentToken)
	tmpPath := targetPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write temporary stash file: %w", err)
	}

	if err := os.Rename(tmpPath, targetPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to finalize stash file: %w", err)
	}

	return nil
}

// Find retrieves a stashed credential matching the spent token fingerprint.
func (s *StashStore) Find(spentToken string) (*StashedCredential, error) {
	targetPath := s.StashPath(spentToken)
	data, err := os.ReadFile(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read stash file: %w", err)
	}

	var cred StashedCredential
	if err := json.Unmarshal(data, &cred); err != nil {
		return nil, fmt.Errorf("corrupt stash record: %w", err)
	}

	return &cred, nil
}

// Adopt retrieves and validates an unclaimed credential for the target agent and profile.
func (s *StashStore) Adopt(spentToken, agent, profile string) (*StashedCredential, error) {
	cred, err := s.Find(spentToken)
	if err != nil {
		return nil, err
	}
	if cred == nil {
		return nil, nil
	}

	if cred.Agent != agent || cred.Profile != profile {
		return nil, fmt.Errorf("stash credential mismatch: expected %s/%s, found %s/%s",
			agent, profile, cred.Agent, cred.Profile)
	}

	return cred, nil
}

// Remove deletes a stashed credential file after successful persistence.
func (s *StashStore) Remove(spentToken string) error {
	targetPath := s.StashPath(spentToken)
	if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove stash file: %w", err)
	}
	return nil
}

// List returns all currently unclaimed credentials in the stash directory.
func (s *StashStore) List() ([]*StashedCredential, error) {
	entries, err := os.ReadDir(s.Dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read unclaimed directory: %w", err)
	}

	var creds []*StashedCredential
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.Dir, entry.Name()))
		if err != nil {
			continue
		}
		var cred StashedCredential
		if err := json.Unmarshal(data, &cred); err != nil {
			continue
		}
		creds = append(creds, &cred)
	}
	return creds, nil
}

// Prune removes stash records older than maxAge to prevent stale disk accumulation.
func (s *StashStore) Prune(maxAge time.Duration) (int, error) {
	entries, err := os.ReadDir(s.Dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to read unclaimed directory for pruning: %w", err)
	}

	cutoff := time.Now().Add(-maxAge)
	prunedCount := 0

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			if err := os.Remove(filepath.Join(s.Dir, entry.Name())); err == nil {
				prunedCount++
			}
		}
	}
	return prunedCount, nil
}
