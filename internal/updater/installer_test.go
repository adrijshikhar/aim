package updater

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func createMockTarGz(t *testing.T, binaryName string, content []byte) []byte {
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

func TestDetectInstallMethod(t *testing.T) {
	tests := []struct {
		name     string
		execPath string
		setup    func(t *testing.T) string
		want     InstallMethod
	}{
		{
			name:     "opt homebrew caskroom",
			execPath: "/opt/homebrew/Caskroom/aim/0.13.0/aim",
			want:     InstallMethodHomebrew,
		},
		{
			name:     "opt homebrew bin",
			execPath: "/opt/homebrew/bin/aim",
			want:     InstallMethodHomebrew,
		},
		{
			name:     "usr local caskroom",
			execPath: "/usr/local/Caskroom/aim/0.14.0/aim",
			want:     InstallMethodHomebrew,
		},
		{
			name:     "usr local cellar",
			execPath: "/usr/local/Cellar/aim/0.14.0/bin/aim",
			want:     InstallMethodHomebrew,
		},
		{
			name:     "user local bin",
			execPath: "/Users/user/.local/bin/aim",
			want:     InstallMethodStandalone,
		},
		{
			name:     "usr local bin standalone",
			execPath: "/usr/local/bin/aim",
			want:     InstallMethodStandalone,
		},
		{
			name: "symlink into homebrew",
			setup: func(t *testing.T) string {
				tmp := t.TempDir()
				brewFake := filepath.Join(tmp, "opt", "homebrew", "Caskroom", "aim", "0.13.0")
				if err := os.MkdirAll(brewFake, 0755); err != nil {
					t.Fatalf("failed to create fake brew dir: %v", err)
				}
				targetBin := filepath.Join(brewFake, "aim")
				if err := os.WriteFile(targetBin, []byte("fake"), 0755); err != nil {
					t.Fatalf("failed to write fake binary: %v", err)
				}
				symlink := filepath.Join(tmp, "aim-symlink")
				if err := os.Symlink(targetBin, symlink); err != nil {
					t.Fatalf("failed to create symlink: %v", err)
				}
				return symlink
			},
			want: InstallMethodHomebrew,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := tc.execPath
			if tc.setup != nil {
				path = tc.setup(t)
			}
			got := DetectInstallMethod(path)
			if got != tc.want {
				t.Errorf("DetectInstallMethod(%q) = %v, want %v", path, got, tc.want)
			}
		})
	}
}

func TestSelfUpdate_Success(t *testing.T) {
	newBinaryContent := []byte("#!/bin/sh\necho updated aim v0.99.0\n")
	tarballData := createMockTarGz(t, "aim", newBinaryContent)
	tarballSum := fmt.Sprintf("%x", sha256.Sum256(tarballData))

	targetVersion := "v0.99.0"
	tarballName := fmt.Sprintf("aim_0.99.0_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	checksumsContent := fmt.Sprintf("%s  %s\n", tarballSum, tarballName)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/tarball":
			w.Header().Set("Content-Type", "application/gzip")
			_, _ = w.Write(tarballData)
		case "/checksums.txt":
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte(checksumsContent))
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

	opts := SelfUpdateOptions{
		TargetVersion: targetVersion,
		ExecPath:      destPath,
		DownloadURL:   server.URL + "/tarball",
		ChecksumURL:   server.URL + "/checksums.txt",
		HTTPClient:    server.Client(),
	}

	err := SelfUpdate(context.Background(), opts)
	if err != nil {
		t.Fatalf("SelfUpdate failed unexpectedly: %v", err)
	}

	updated, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("failed to read updated binary: %v", err)
	}
	if !bytes.Equal(updated, newBinaryContent) {
		t.Errorf("expected binary content %q, got %q", string(newBinaryContent), string(updated))
	}

	info, err := os.Stat(destPath)
	if err != nil {
		t.Fatalf("failed to stat updated binary: %v", err)
	}
	if info.Mode().Perm() != 0755 {
		t.Errorf("expected permissions 0755, got %04o", info.Mode().Perm())
	}
}

func TestSelfUpdate_ChecksumMismatch(t *testing.T) {
	newBinaryContent := []byte("corrupted or tampered content")
	tarballData := createMockTarGz(t, "aim", newBinaryContent)

	targetVersion := "v0.99.0"
	tarballName := fmt.Sprintf("aim_0.99.0_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	// Provide wrong checksum
	checksumsContent := fmt.Sprintf("0000000000000000000000000000000000000000000000000000000000000000  %s\n", tarballName)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/tarball":
			w.Header().Set("Content-Type", "application/gzip")
			_, _ = w.Write(tarballData)
		case "/checksums.txt":
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte(checksumsContent))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	destDir := t.TempDir()
	destPath := filepath.Join(destDir, "aim")
	originalContent := []byte("original aim binary")
	if err := os.WriteFile(destPath, originalContent, 0755); err != nil {
		t.Fatalf("failed to write original binary: %v", err)
	}

	opts := SelfUpdateOptions{
		TargetVersion: targetVersion,
		ExecPath:      destPath,
		DownloadURL:   server.URL + "/tarball",
		ChecksumURL:   server.URL + "/checksums.txt",
		HTTPClient:    server.Client(),
	}

	err := SelfUpdate(context.Background(), opts)
	if err == nil {
		t.Fatal("expected checksum mismatch error, got nil")
	}

	// Verify original file was left untouched
	remaining, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("failed to read destination binary: %v", err)
	}
	if !bytes.Equal(remaining, originalContent) {
		t.Errorf("destination binary was modified despite checksum mismatch!")
	}
}

func TestExtractBinary_NotFoundInArchive(t *testing.T) {
	// Archive containing "not-aim" instead of "aim"
	tarballData := createMockTarGz(t, "not-aim", []byte("other"))
	_, err := extractBinary(bytes.NewReader(tarballData), "aim")
	if err == nil {
		t.Fatalf("expected error when binary 'aim' is missing in archive, got nil")
	}
}
