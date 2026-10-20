package presenter

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/share"
)

func TestRenderSharedProfilesTable(t *testing.T) {
	var buf bytes.Buffer
	profiles := []share.SharedProfileConfig{
		{
			Name:        "work",
			Agent:       "claude",
			AutoApprove: true,
			MaxLeaseTTL: 4 * time.Hour,
			Enabled:     true,
		},
	}
	out := RenderSharedProfilesTable(&buf, profiles)
	if !strings.Contains(out, "work") || !strings.Contains(out, "claude") {
		t.Errorf("expected table output with work and claude, got %s", out)
	}
}

func TestRenderPeersTable(t *testing.T) {
	var buf bytes.Buffer
	peers := []share.PeerInfo{
		{
			ID:       "node-1",
			HostName: "alice-laptop",
			VPNIP:    "10.8.19.251",
			Port:     4319,
			LastSeen: time.Now(),
			Profiles: []string{"work"},
		},
	}
	out := RenderPeersTable(&buf, peers)
	if !strings.Contains(out, "alice-laptop") || !strings.Contains(out, "10.8.19.251") {
		t.Errorf("expected table output with alice-laptop and 10.8.19.251, got %s", out)
	}
}
