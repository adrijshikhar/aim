package main

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/usage"
)

func TestFormatProfileUsageBadge_UTF8ErrorTruncation(t *testing.T) {
	cache := usage.NewCacheStore(t.TempDir(), time.Minute)
	longUTF8Err := strings.Repeat("語", 25) // 25 runes, 75 UTF-8 bytes
	if err := cache.Put(usage.Report{
		Agent:   "claude",
		Profile: "work",
		Error:   longUTF8Err,
	}); err != nil {
		t.Fatalf("cache.Put: %v", err)
	}

	got := formatProfileUsageBadge(cache, agents.NewRegistry(), "claude", "work")
	if !utf8.ValidString(got) {
		t.Fatalf("expected valid UTF-8 badge, got %q", got)
	}
	wantPrefix := "[" + strings.Repeat("語", 20) + "...]"
	if got != wantPrefix {
		t.Fatalf("got %q, want %q", got, wantPrefix)
	}
}
