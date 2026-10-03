package session

import (
	"bufio"
	"bytes"
	"io"
	"os"
)

const maxLineLength = 4 * 1024 * 1024 // 4MB safety ceiling

// ReadTurnCounts extracts session start and turn count without full in-memory JSON parsing.
func ReadTurnCounts(f *os.File) (int, error) {
	_, _ = f.Seek(0, io.SeekStart)
	scanner := bufio.NewScanner(f)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, maxLineLength)

	turns := 0
	for scanner.Scan() {
		line := scanner.Bytes()
		if bytes.Contains(line, []byte(`"type":"USER_INPUT"`)) ||
			bytes.Contains(line, []byte(`"type":"user"`)) {
			turns++
		}
	}
	return turns, scanner.Err()
}

// ReadMessageCounts extracts total user and assistant message count without full in-memory JSON parsing.
func ReadMessageCounts(f *os.File) (int, error) {
	_, _ = f.Seek(0, io.SeekStart)
	scanner := bufio.NewScanner(f)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, maxLineLength)

	count := 0
	for scanner.Scan() {
		line := scanner.Bytes()
		if bytes.Contains(line, []byte(`"type":"user"`)) ||
			bytes.Contains(line, []byte(`"type":"assistant"`)) ||
			bytes.Contains(line, []byte(`"type":"USER_INPUT"`)) {
			count++
		}
	}
	return count, scanner.Err()
}
