package telemetry

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var knownAgents = map[string]bool{
	"agy":         true,
	"antigravity": true,
	"claude":      true,
	"codex":       true,
	"gemini":      true,
}

// DurationBucket maps execution duration to coarse privacy-safe buckets.
func DurationBucket(d time.Duration) string {
	switch {
	case d < 500*time.Millisecond:
		return "<500ms"
	case d < 2*time.Second:
		return "500ms-2s"
	case d < 10*time.Second:
		return "2s-10s"
	default:
		return ">10s"
	}
}

// SanitizeCommand returns the canonical command and detected agent name,
// stripping any profile names, filesystem paths, prompt contents, or custom flags.
func SanitizeCommand(rawCmd string, rawArgs []string) (cleanCmd string, cleanAgent string) {
	cleanCmd = strings.TrimSpace(rawCmd)
	for _, arg := range rawArgs {
		argClean := strings.ToLower(strings.TrimSpace(arg))
		if knownAgents[argClean] {
			cleanAgent = argClean
			break
		}
	}
	return cleanCmd, cleanAgent
}

// AnonymousMachineID returns a pseudonymous, non-reversible SHA-256 machine identifier.
// It is generated once and stored locally at ~/.aim/telemetry_id.
func AnonymousMachineID(baseDir string) string {
	idPath := filepath.Join(baseDir, "telemetry_id")
	if data, err := os.ReadFile(idPath); err == nil {
		id := strings.TrimSpace(string(data))
		if len(id) == 64 {
			return id
		}
	}

	// Generate 32 random bytes and hash them
	buf := make([]byte, 32)
	_, _ = rand.Read(buf)
	h := sha256.Sum256(buf)
	id := hex.EncodeToString(h[:])

	_ = os.MkdirAll(baseDir, 0755)
	_ = os.WriteFile(idPath, []byte(id), 0600)
	return id
}
