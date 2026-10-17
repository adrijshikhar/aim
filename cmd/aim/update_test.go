package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/updater"
)

func createTestTarGz(t *testing.T, binaryName string, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	hdr := &tar.Header{
		Name: binaryName,
		Mode: 0755,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("failed to write tar header: %v", err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatalf("failed to write tar content: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("failed to close tar writer: %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("failed to close gzip writer: %v", err)
	}
	return buf.Bytes()
}

func TestUpdateCmd_Help(t *testing.T) {
	tempBase := t.TempDir()
	t.Setenv("AIM_HOME", tempBase)

	pm := profile.NewProfileManager(tempBase)
	reg := agents.NewRegistry()

	out, _ := captureOutput(t, func() {
		code := dispatch([]string{"update", "--help"}, reg, pm)
		if code != 0 {
			t.Fatalf("expected code 0, got %d", code)
		}
	})

	for _, expected := range []string{"--check", "--force", "--background", "Check for and install updates to AIM"} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected help output to contain %q, got:\n%s", expected, out)
		}
	}
}

func TestUpdateCmd_Check_UpToDate(t *testing.T) {
	tempBase := t.TempDir()
	t.Setenv("AIM_HOME", tempBase)

	cacheDir := filepath.Join(tempBase, "cache")
	_ = os.MkdirAll(cacheDir, 0700)

	// Write cache with current version matching latest
	cacheData := updater.CacheData{
		LatestVersion: "v1.0.0",
		ReleaseURL:    "https://github.com/adrijshikhar/aim/releases/tag/v1.0.0",
		CheckedAt:     time.Now(),
	}
	data, _ := json.Marshal(cacheData)
	_ = os.WriteFile(filepath.Join(cacheDir, "update_check.json"), data, 0600)

	pm := profile.NewProfileManager(tempBase)
	reg := agents.NewRegistry()

	origVer := GetVersion()
	defer func() { SetVersion(origVer) }()
	SetVersion("1.0.0")

	out, _ := captureOutput(t, func() {
		code := dispatch([]string{"update", "--check"}, reg, pm)
		if code != 0 {
			t.Fatalf("expected code 0, got %d", code)
		}
	})

	if !strings.Contains(out, "up-to-date") && !strings.Contains(out, "already on the latest") {
		t.Errorf("expected up-to-date message, got:\n%s", out)
	}
}

func TestUpdateCmd_Check_UpdateAvailable(t *testing.T) {
	tempBase := t.TempDir()
	t.Setenv("AIM_HOME", tempBase)

	cacheDir := filepath.Join(tempBase, "cache")
	_ = os.MkdirAll(cacheDir, 0700)

	cacheData := updater.CacheData{
		LatestVersion: "v1.1.0",
		ReleaseURL:    "https://github.com/adrijshikhar/aim/releases/tag/v1.1.0",
		CheckedAt:     time.Now(),
	}
	data, _ := json.Marshal(cacheData)
	_ = os.WriteFile(filepath.Join(cacheDir, "update_check.json"), data, 0600)

	pm := profile.NewProfileManager(tempBase)
	reg := agents.NewRegistry()

	origVer := GetVersion()
	defer func() { SetVersion(origVer) }()
	SetVersion("1.0.0")

	out, _ := captureOutput(t, func() {
		code := dispatch([]string{"update", "--check"}, reg, pm)
		if code != 0 {
			t.Fatalf("expected code 0, got %d", code)
		}
	})

	if !strings.Contains(out, "A new version of aim is available: 1.0.0 → v1.1.0") {
		t.Errorf("expected update available message, got:\n%s", out)
	}
}

func TestUpdateCmd_Run_AlreadyUpToDate_WithoutForce(t *testing.T) {
	tempBase := t.TempDir()
	t.Setenv("AIM_HOME", tempBase)

	cacheDir := filepath.Join(tempBase, "cache")
	_ = os.MkdirAll(cacheDir, 0700)

	cacheData := updater.CacheData{
		LatestVersion: "v1.0.0",
		ReleaseURL:    "https://github.com/adrijshikhar/aim/releases/tag/v1.0.0",
		CheckedAt:     time.Now(),
	}
	data, _ := json.Marshal(cacheData)
	_ = os.WriteFile(filepath.Join(cacheDir, "update_check.json"), data, 0600)

	pm := profile.NewProfileManager(tempBase)
	reg := agents.NewRegistry()

	origVer := GetVersion()
	defer func() { SetVersion(origVer) }()
	SetVersion("1.0.0")

	out, _ := captureOutput(t, func() {
		code := dispatch([]string{"update"}, reg, pm)
		if code != 0 {
			t.Fatalf("expected code 0, got %d", code)
		}
	})

	if !strings.Contains(out, "already on the latest version") {
		t.Errorf("expected already on latest version message, got:\n%s", out)
	}
}

func TestUpdateCmd_Run_Standalone_Success(t *testing.T) {
	newBinary := []byte("#!/bin/sh\necho updated aim v2.0.0\n")
	tarData := createTestTarGz(t, "aim", newBinary)
	tarSum := fmt.Sprintf("%x", sha256.Sum256(tarData))

	tarName := fmt.Sprintf("aim_2.0.0_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	checksums := fmt.Sprintf("%s  %s\n", tarSum, tarName)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/tarball":
			w.Header().Set("Content-Type", "application/gzip")
			_, _ = w.Write(tarData)
		case "/checksums.txt":
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte(checksums))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	destDir := t.TempDir()
	destPath := filepath.Join(destDir, "aim")
	if err := os.WriteFile(destPath, []byte("old aim binary"), 0755); err != nil {
		t.Fatalf("failed to write original binary: %v", err)
	}

	opts := updater.SelfUpdateOptions{
		TargetVersion: "v2.0.0",
		ExecPath:      destPath,
		DownloadURL:   server.URL + "/tarball",
		ChecksumURL:   server.URL + "/checksums.txt",
		HTTPClient:    server.Client(),
	}

	err := updater.SelfUpdate(t.Context(), opts)
	if err != nil {
		t.Fatalf("SelfUpdate failed: %v", err)
	}

	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("failed to read destination: %v", err)
	}
	if !bytes.Equal(data, newBinary) {
		t.Errorf("expected updated content %q, got %q", string(newBinary), string(data))
	}
}
