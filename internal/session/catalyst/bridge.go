package catalyst

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aim-cli/aim/internal/session"
)

// HandoffBrief matches the schema defined in the Catalyst handoff specification (.catalyst/handoffs/<key>.json).
type HandoffBrief struct {
	SchemaVersion    int      `json:"schema_version"`
	Branch           string   `json:"branch"`
	Repo             string   `json:"repo"`
	Goal             string   `json:"goal"`
	DoneWhen         string   `json:"done_when"`
	KeyDecisions     []string `json:"key_decisions"`
	OpenRisks        []string `json:"open_risks"`
	FilesToReadFirst []string `json:"files_to_read_first"`
	Notes            string   `json:"notes"`
}

type Bridge struct{}

func NewBridge() *Bridge {
	return &Bridge{}
}

// BranchToKey converts a branch name into a sanitized handoff key (e.g. feat/foo -> feat-foo).
func BranchToKey(branch string) string {
	cleaned := strings.ReplaceAll(branch, "/", "-")
	cleaned = strings.ReplaceAll(cleaned, "\\", "-")
	if len(cleaned) > 80 {
		cleaned = cleaned[:80]
	}
	if cleaned == "" {
		return "default"
	}
	return cleaned
}

// WriteHandoffBrief serializes a Session's summary into <repoRoot>/.catalyst/handoffs/<branchKey>.json.
func (b *Bridge) WriteHandoffBrief(repoRoot, branch string, s *session.Session) (string, error) {
	if s == nil {
		return "", fmt.Errorf("session is nil")
	}

	storeDir := filepath.Join(repoRoot, ".catalyst", "handoffs")
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		return "", err
	}

	key := BranchToKey(branch)
	targetPath := filepath.Join(storeDir, key+".json")

	goalTitle := s.Title
	if goalTitle == "" {
		goalTitle = "AIM Resumed Session"
	}

	brief := HandoffBrief{
		SchemaVersion: 1,
		Branch:        branch,
		Repo:          repoRoot,
		Goal:          fmt.Sprintf("Continue session: %s", goalTitle),
		DoneWhen:      "Tasks from previous session completed successfully",
		KeyDecisions: []string{
			fmt.Sprintf("Resumed from AIM session %s (%s/%s)", s.ShortID, s.Agent, s.Profile),
		},
		OpenRisks:        []string{},
		FilesToReadFirst: []string{},
		Notes:            s.Summary,
	}

	data, err := json.MarshalIndent(brief, "", "  ")
	if err != nil {
		return "", err
	}

	tmpPath := targetPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return "", err
	}

	if err := os.Rename(tmpPath, targetPath); err != nil {
		_ = os.Remove(tmpPath)
		return "", err
	}

	return targetPath, nil
}

// ReadHandoffBrief reads and deserializes the brief for the given branch.
func (b *Bridge) ReadHandoffBrief(repoRoot, branch string) (*HandoffBrief, error) {
	key := BranchToKey(branch)
	targetPath := filepath.Join(repoRoot, ".catalyst", "handoffs", key+".json")

	data, err := os.ReadFile(targetPath)
	if err != nil {
		return nil, err
	}

	var brief HandoffBrief
	if err := json.Unmarshal(data, &brief); err != nil {
		return nil, err
	}

	return &brief, nil
}

// ResolveRepoAndBranch resolves the root worktree and active git branch for a directory.
func ResolveRepoAndBranch(dir string) (repoRoot, branch string, err error) {
	rootCmd := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel")
	rootOut, err := rootCmd.Output()
	if err != nil {
		return dir, "default", nil
	}
	repoRoot = strings.TrimSpace(string(rootOut))

	branchCmd := exec.Command("git", "-C", dir, "branch", "--show-current")
	branchOut, err := branchCmd.Output()
	if err == nil && len(branchOut) > 0 {
		branch = strings.TrimSpace(string(branchOut))
	} else {
		// Detached HEAD or fallback
		headCmd := exec.Command("git", "-C", dir, "rev-parse", "--short", "HEAD")
		if headOut, err := headCmd.Output(); err == nil {
			branch = strings.TrimSpace(string(headOut))
		} else {
			branch = "default"
		}
	}

	return repoRoot, branch, nil
}
