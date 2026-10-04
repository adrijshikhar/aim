package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/updater"
)

func TestUpdater_NoticeOnStderrWhenUpdateAvailable(t *testing.T) {
	tempBase := t.TempDir()
	origHome := os.Getenv("AIM_HOME")
	defer func() { _ = os.Setenv("AIM_HOME", origHome) }()
	_ = os.Setenv("AIM_HOME", tempBase)

	cacheDir := filepath.Join(tempBase, "cache")
	_ = os.MkdirAll(cacheDir, 0755)

	// Write cached update info: current is 0.12.1, latest is 0.13.0
	cacheData := updater.CacheData{
		LatestVersion: "v0.13.0",
		ReleaseURL:    "https://github.com/adrijshikhar/aim/releases/tag/v0.13.0",
		CheckedAt:     time.Now(),
	}
	data, _ := json.Marshal(cacheData)
	_ = os.WriteFile(filepath.Join(cacheDir, "update_check.json"), data, 0644)

	pm := profile.NewProfileManager(tempBase)
	reg := agents.NewRegistry()

	origVer := Version
	defer func() { Version = origVer }()
	Version = "0.12.1"

	_, errOut := captureOutput(t, func() {
		code := dispatch([]string{"version"}, reg, pm)
		if code != 0 {
			t.Fatalf("expected code 0, got %d", code)
		}
	})

	if !strings.Contains(errOut, "A new version of aim is available: 0.12.1 → v0.13.0") {
		t.Errorf("expected update notice on stderr, got:\n%s", errOut)
	}
}

func TestUpdater_NoNoticeWhenJsonFlag(t *testing.T) {
	tempBase := t.TempDir()
	origHome := os.Getenv("AIM_HOME")
	defer func() { _ = os.Setenv("AIM_HOME", origHome) }()
	_ = os.Setenv("AIM_HOME", tempBase)

	cacheDir := filepath.Join(tempBase, "cache")
	_ = os.MkdirAll(cacheDir, 0755)

	cacheData := updater.CacheData{
		LatestVersion: "v0.13.0",
		ReleaseURL:    "https://github.com/adrijshikhar/aim/releases/tag/v0.13.0",
		CheckedAt:     time.Now(),
	}
	data, _ := json.Marshal(cacheData)
	_ = os.WriteFile(filepath.Join(cacheDir, "update_check.json"), data, 0644)

	pm := profile.NewProfileManager(tempBase)
	reg := agents.NewRegistry()

	origVer := Version
	defer func() { Version = origVer }()
	Version = "0.12.1"

	_, errOut := captureOutput(t, func() {
		_ = dispatch([]string{"list", "--json"}, reg, pm)
	})

	if strings.Contains(errOut, "A new version of aim is available") {
		t.Errorf("expected no update notice when --json flag passed, got:\n%s", errOut)
	}
}
