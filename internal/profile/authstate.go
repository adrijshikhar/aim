package profile

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/aim-cli/aim/internal/merge"
)

const (
	logoutMarkerFileName = ".aim-logged-out"

	agentClaude           = "claude"
	agentCodex            = "codex"
	agentAgy              = "agy"
	agentAliasAntigravity = "antigravity"
	agentGemini           = "gemini"
	agentCursor           = "cursor"
)

func canonicalAuthAgent(agentName string) string {
	switch strings.ToLower(strings.TrimSpace(agentName)) {
	case agentClaude, "cc", "claude-code":
		return agentClaude
	case agentCodex, "codex-cli", "openai-codex":
		return agentCodex
	case agentAgy, agentAliasAntigravity:
		return agentAgy
	case agentGemini, "gemini-cli":
		return agentGemini
	case agentCursor, "cur":
		return agentCursor
	default:
		return strings.ToLower(strings.TrimSpace(agentName))
	}
}

func agentStateSubdir(agentName string) string {
	switch canonicalAuthAgent(agentName) {
	case agentClaude:
		return ".claude"
	case agentCodex:
		return ".codex"
	case agentAgy:
		return filepath.Join(".gemini", "antigravity-cli")
	case agentGemini:
		return ".gemini"
	case agentCursor:
		return ".cursor"
	default:
		return "."
	}
}

// LogoutMarkerPath returns the profile-scoped marker file path recording an
// explicit logout for the given agent within profileDir.
func LogoutMarkerPath(profileDir, agentName string) string {
	if strings.TrimSpace(profileDir) == "" || strings.TrimSpace(agentName) == "" {
		return ""
	}
	sub := agentStateSubdir(agentName)
	if sub == "." {
		return filepath.Join(profileDir, logoutMarkerFileName+"-"+strings.ToLower(strings.TrimSpace(agentName)))
	}
	return filepath.Join(profileDir, sub, logoutMarkerFileName)
}

// IsAgentLoggedOut reports whether the profile holds an explicit logout marker for agentName.
func IsAgentLoggedOut(profileDir, agentName string) bool {
	p := LogoutMarkerPath(profileDir, agentName)
	if p == "" {
		return false
	}
	fi, err := os.Stat(p)
	return err == nil && !fi.IsDir()
}

// MarkAgentLoggedOut records an explicit logout marker so host credential seeding
// and background Keychain harvesting never silently overwrite a logged-out profile.
func MarkAgentLoggedOut(profileDir, agentName string) error {
	p := LogoutMarkerPath(profileDir, agentName)
	if p == "" {
		return errors.New("profile directory and agent name are required to mark logout")
	}
	return merge.AtomicWrite(p, []byte("logged_out\n"), 0o600)
}

// ClearAgentLoggedOut removes the explicit logout marker upon a successful login.
func ClearAgentLoggedOut(profileDir, agentName string) error {
	p := LogoutMarkerPath(profileDir, agentName)
	if p == "" {
		return nil
	}
	err := os.Remove(p)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// PurgeLocalAgentCredentials removes local profile-scoped credential files for agentName
// inside profileDir without touching other profiles, host HOME, or session history.
func PurgeLocalAgentCredentials(profileDir, agentName string) error {
	if strings.TrimSpace(profileDir) == "" {
		return errors.New("profile directory is required")
	}
	var targets []string
	switch canonicalAuthAgent(agentName) {
	case agentClaude:
		targets = []string{
			filepath.Join(profileDir, ".claude", ".credentials.json"),
			filepath.Join(profileDir, ".claude", "auth.json"),
		}
	case agentCodex:
		targets = []string{
			filepath.Join(profileDir, ".codex", "auth.json"),
		}
	case agentAgy:
		targets = []string{
			filepath.Join(profileDir, ".gemini", "antigravity-cli", "antigravity-oauth-token"),
			filepath.Join(profileDir, ".config", "gcloud", "application_default_credentials.json"),
		}
	case agentGemini:
		targets = []string{
			filepath.Join(profileDir, ".gemini", "gemini-oauth-token"),
		}
	case agentCursor:
		targets = []string{
			filepath.Join(profileDir, ".cursor", "user-data", "User", "globalStorage", "state.vscdb"),
		}
	}

	var errs []error
	for _, p := range targets {
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
