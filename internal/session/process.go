package session

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aim-cli/aim/internal/logger"
)

var (
	// Matches: aim run <agent> <profile> ... --conversation[= ]<id>
	aimRunConvRegex = regexp.MustCompile(`aim\s+run\s+([a-zA-Z0-9_-]+)\s+([a-zA-Z0-9_-]+).*?--conversation[=\s]([a-zA-Z0-9_-]+)`)
	// Matches: agy ... --conversation[= ]<id>
	agyConvRegex = regexp.MustCompile(`(?:^|/|\s)agy(?:\.exe)?\s+.*?--conversation[=\s]([a-zA-Z0-9_-]+)`)
	// Matches: codex ... resume <id>
	codexResumeRegex = regexp.MustCompile(`(?:^|/|\s)codex(?:\.exe)?\s+.*?resume\s+([a-zA-Z0-9_-]+)`)
)

// DefaultProcessScanner scans local OS processes via `ps -eo pid,command`.
type DefaultProcessScanner struct{}

func NewDefaultProcessScanner() *DefaultProcessScanner {
	return &DefaultProcessScanner{}
}

func (s *DefaultProcessScanner) ScanActiveProcesses(ctx context.Context) (map[string]ActiveProcessInfo, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctxTimeout, "ps", "-eo", "pid,command")
	out, err := cmd.Output()
	if err != nil {
		return map[string]ActiveProcessInfo{}, nil
	}

	return ParseProcessOutput(bytes.NewReader(out)), nil
}

// ParseProcessOutput parses tabular `ps -eo pid,command` output.
func ParseProcessOutput(r io.Reader) map[string]ActiveProcessInfo {
	result := make(map[string]ActiveProcessInfo)
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "PID") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		pid, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}

		cmdline := strings.Join(fields[1:], " ")

		// 1. Check for `aim run <agent> <profile> ... --conversation=<id>`
		if matches := aimRunConvRegex.FindStringSubmatch(cmdline); len(matches) >= 4 {
			agent := matches[1]
			profile := matches[2]
			convID := matches[3]
			result[convID] = ActiveProcessInfo{
				PID:            pid,
				Agent:          agent,
				Profile:        profile,
				ConversationID: convID,
			}
			continue
		}

		// 2. Check for native `agy ... --conversation=<id>`
		if matches := agyConvRegex.FindStringSubmatch(cmdline); len(matches) >= 2 {
			convID := matches[1]
			// Only set if not already set by parent aim process
			if _, exists := result[convID]; !exists {
				result[convID] = ActiveProcessInfo{
					PID:            pid,
					Agent:          "agy",
					Profile:        "<host>",
					ConversationID: convID,
				}
			}
			continue
		}

		// 3. Check for `codex ... resume <id>`
		if matches := codexResumeRegex.FindStringSubmatch(cmdline); len(matches) >= 2 {
			convID := matches[1]
			if _, exists := result[convID]; !exists {
				result[convID] = ActiveProcessInfo{
					PID:            pid,
					Agent:          "codex",
					Profile:        "<host>",
					ConversationID: convID,
				}
			}
			continue
		}
	}
	if err := scanner.Err(); err != nil {
		logger.Debug("[session/process] scanner error parsing process output: %v", err)
	}

	return result
}
