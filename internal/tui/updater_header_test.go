package tui

import (
	"strings"
	"testing"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
)

func TestTUI_HeaderDisplaysUpdateNoticeWhenAvailable(t *testing.T) {
	tempBase := t.TempDir()
	pm := profile.NewProfileManager(tempBase)
	reg := agents.NewRegistry()
	cfg := config.NewDefaultConfig()

	m := NewModel(reg, pm, cfg)
	m.version = "0.12.1"

	// Initially no update available
	viewBefore := m.renderHeader()
	if strings.Contains(viewBefore, "Update Available") {
		t.Errorf("expected no update badge initially, got:\n%s", viewBefore)
	}

	// UpdateAvailable message received
	updated, _ := m.Update(updateAvailableMsg{latestVersion: "v0.13.0"})
	mUpdated := updated.(Model)

	viewAfter := mUpdated.renderHeader()
	if !strings.Contains(viewAfter, "[Update Available: v0.13.0]") {
		t.Errorf("expected update available badge in header, got:\n%s", viewAfter)
	}
}
