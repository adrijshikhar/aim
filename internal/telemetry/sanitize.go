package telemetry

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aim-cli/aim/internal/merge"
)

var machineIDMu sync.Mutex

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
	if id := readValidTelemetryID(baseDir); id != "" {
		return id
	}

	machineIDMu.Lock()
	defer machineIDMu.Unlock()

	if id := readValidTelemetryID(baseDir); id != "" {
		return id
	}

	// Generate 32 random bytes and hash them
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	h := sha256.Sum256(buf)
	id := hex.EncodeToString(h[:])

	_ = os.MkdirAll(baseDir, 0o700)
	if root, err := os.OpenRoot(baseDir); err == nil {
		f, openErr := root.OpenFile("telemetry_id", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		_ = root.Close()
		if openErr == nil {
			_, _ = f.WriteString(id)
			_ = f.Close()
			return id
		}
	}
	if existing := readValidTelemetryID(baseDir); existing != "" {
		return existing
	}
	_ = merge.AtomicWrite(idPath, []byte(id), 0o600)
	return id
}

func readValidTelemetryID(baseDir string) string {
	root, err := os.OpenRoot(baseDir)
	if err != nil {
		return ""
	}
	defer func() { _ = root.Close() }()
	data, err := root.ReadFile("telemetry_id")
	if err != nil {
		return ""
	}
	id := strings.TrimSpace(string(data))
	if !isValidHexID(id) {
		return ""
	}
	return id
}

func isValidHexID(id string) bool {
	if len(id) != 64 {
		return false
	}
	for i := range id {
		c := id[i]
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}
