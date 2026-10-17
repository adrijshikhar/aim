package updater

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// InstallMethod indicates how the current AIM binary was installed.
type InstallMethod string

const (
	InstallMethodHomebrew   InstallMethod = "homebrew"
	InstallMethodStandalone InstallMethod = "standalone"
)

// DetectInstallMethod determines if the binary at execPath is managed by Homebrew
// or is a standalone installation.
func DetectInstallMethod(execPath string) InstallMethod {
	if execPath == "" {
		return InstallMethodStandalone
	}

	cleanPath := filepath.Clean(execPath)
	realPath := cleanPath
	if resolved, err := filepath.EvalSymlinks(cleanPath); err == nil {
		realPath = filepath.Clean(resolved)
	}

	pathsToCheck := []string{cleanPath, realPath}

	brewPrefix := os.Getenv("HOMEBREW_PREFIX")
	if brewPrefix != "" {
		brewPrefix = filepath.Clean(brewPrefix)
	}

	for _, p := range pathsToCheck {
		normalized := filepath.ToSlash(p)

		// Homebrew Caskroom or Cellar paths
		if strings.Contains(normalized, "/Caskroom/aim/") || strings.Contains(normalized, "/Cellar/aim/") {
			return InstallMethodHomebrew
		}

		// Standard macOS Apple Silicon Homebrew paths
		if strings.HasPrefix(normalized, "/opt/homebrew/") {
			return InstallMethodHomebrew
		}

		// Custom or env HOMEBREW_PREFIX
		if brewPrefix != "" && strings.HasPrefix(p, brewPrefix) {
			if strings.Contains(normalized, "/Cellar/") || strings.Contains(normalized, "/Caskroom/") {
				return InstallMethodHomebrew
			}
		}

		// Standard Intel Mac / Linuxbrew paths
		if strings.HasPrefix(normalized, "/usr/local/Homebrew/") || strings.HasPrefix(normalized, "/usr/local/Caskroom/") || strings.HasPrefix(normalized, "/usr/local/Cellar/") {
			return InstallMethodHomebrew
		}
		if strings.HasPrefix(normalized, "/home/linuxbrew/.linuxbrew/") {
			return InstallMethodHomebrew
		}
	}

	return InstallMethodStandalone
}

// UpgradeHomebrew invokes Homebrew to upgrade the AIM cask.
func UpgradeHomebrew(ctx context.Context, stdout, stderr io.Writer) error {
	brewBin, err := exec.LookPath("brew")
	if err != nil {
		return fmt.Errorf("homebrew executable 'brew' not found in PATH: %w", err)
	}

	cmd := exec.CommandContext(ctx, brewBin, "upgrade", "adrijshikhar/tap/aim")
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if stdout != nil {
		cmd.Stdin = os.Stdin
	}

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		// Fallback attempt: brew upgrade aim
		fallbackCmd := exec.CommandContext(ctx, brewBin, "upgrade", "aim")
		fallbackCmd.Stdout = stdout
		fallbackCmd.Stderr = stderr
		if stdout != nil {
			fallbackCmd.Stdin = os.Stdin
		}
		if fallbackErr := fallbackCmd.Run(); fallbackErr != nil {
			return fmt.Errorf("brew upgrade failed: %w", fallbackErr)
		}
	}
	return nil
}

// SelfUpdateOptions configures standalone binary update.
type SelfUpdateOptions struct {
	TargetVersion  string
	CurrentVersion string
	ExecPath       string
	DownloadURL    string
	ChecksumURL    string
	HTTPClient     *http.Client
	Stdout         io.Writer
}

// SelfUpdate downloads, verifies, and atomically replaces the running AIM binary.
func SelfUpdate(ctx context.Context, opts SelfUpdateOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	destPath := opts.ExecPath
	if destPath == "" {
		var err error
		destPath, err = os.Executable()
		if err != nil {
			return fmt.Errorf("failed to determine executable path: %w", err)
		}
	}
	destPath = filepath.Clean(destPath)
	if resolved, err := filepath.EvalSymlinks(destPath); err == nil {
		destPath = filepath.Clean(resolved)
	}

	targetVer := opts.TargetVersion
	if targetVer == "" {
		return errors.New("target version cannot be empty")
	}
	if !strings.HasPrefix(targetVer, "v") {
		targetVer = "v" + targetVer
	}
	cleanVer := strings.TrimPrefix(targetVer, "v")

	goos := runtime.GOOS
	goarch := runtime.GOARCH
	tarballName := fmt.Sprintf("aim_%s_%s_%s.tar.gz", cleanVer, goos, goarch)

	tarballURL := opts.DownloadURL
	if tarballURL == "" {
		tarballURL = fmt.Sprintf("https://github.com/adrijshikhar/aim/releases/download/%s/%s", targetVer, tarballName)
	}

	checksumURL := opts.ChecksumURL
	if checksumURL == "" {
		checksumURL = fmt.Sprintf("https://github.com/adrijshikhar/aim/releases/download/%s/checksums.txt", targetVer)
	}

	client := opts.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}

	if opts.Stdout != nil {
		fmt.Fprintf(opts.Stdout, "Downloading checksums for AIM %s...\n", targetVer)
	}

	// 1. Fetch checksums.txt
	expectedChecksum, err := fetchExpectedChecksum(ctx, client, checksumURL, tarballName)
	if err != nil {
		return fmt.Errorf("failed to fetch or verify checksums: %w", err)
	}

	if opts.Stdout != nil {
		fmt.Fprintf(opts.Stdout, "Downloading %s...\n", tarballName)
	}

	// 2. Download tarball
	tarballData, err := fetchAssetData(ctx, client, tarballURL)
	if err != nil {
		return fmt.Errorf("failed to download release asset: %w", err)
	}

	// 3. Verify SHA256
	actualChecksum := fmt.Sprintf("%x", sha256.Sum256(tarballData))
	if !strings.EqualFold(expectedChecksum, actualChecksum) {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	if opts.Stdout != nil {
		fmt.Fprintln(opts.Stdout, "Checksum verified. Extracting binary...")
	}

	// 4. Extract "aim" binary
	extracted, err := extractBinary(bytes.NewReader(tarballData), "aim")
	if err != nil {
		return fmt.Errorf("failed to extract binary from release archive: %w", err)
	}

	// 5. Check write permissions and atomically replace
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0700); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	tmpFile, err := os.CreateTemp(destDir, ".aim.update.*.tmp")
	if err != nil {
		return fmt.Errorf("permission denied or failed to create temporary file in %s (run with sudo if needed): %w", destDir, err)
	}
	tmpName := tmpFile.Name()
	defer func() {
		_ = os.Remove(tmpName)
	}()

	if _, err := tmpFile.Write(extracted); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("failed to write new binary to temp file: %w", err)
	}

	if err := tmpFile.Chmod(0755); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("failed to set binary permissions: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("failed to sync binary to disk: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Atomic replace
	if err := os.Rename(tmpName, destPath); err != nil {
		return fmt.Errorf("failed to replace executable %s (run with sudo if required): %w", destPath, err)
	}

	// 6. Strip macOS quarantine if on Darwin
	if runtime.GOOS == "darwin" {
		if xattr, err := exec.LookPath("xattr"); err == nil {
			_ = exec.CommandContext(ctx, xattr, "-dr", "com.apple.quarantine", destPath).Run()
		}
	}

	return nil
}

const (
	maxAssetSize  = 250 * 1024 * 1024 // 250 MB
	maxBinarySize = 250 * 1024 * 1024 // 250 MB
)

func fetchAssetData(ctx context.Context, client *http.Client, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "aim-updater")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status %d from %s", resp.StatusCode, url)
	}

	return io.ReadAll(io.LimitReader(resp.Body, maxAssetSize))
}

func fetchExpectedChecksum(ctx context.Context, client *http.Client, url, targetFilename string) (string, error) {
	data, err := fetchAssetData(ctx, client, url)
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			hash := fields[0]
			rawName := strings.TrimPrefix(fields[1], "*")
			name := strings.TrimPrefix(filepath.Base(rawName), "*")
			if name == targetFilename || rawName == targetFilename {
				return hash, nil
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("file %s not found in checksums manifest", targetFilename)
}

func extractBinary(r io.Reader, binaryName string) ([]byte, error) {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("gzip reader error: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("tar reader error: %w", err)
		}

		if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeRegA {
			continue
		}

		cleanName := filepath.Clean(header.Name)
		if cleanName == binaryName || cleanName == "./"+binaryName {
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, io.LimitReader(tr, maxBinarySize)); err != nil {
				return nil, fmt.Errorf("failed to read binary from archive: %w", err)
			}
			return buf.Bytes(), nil
		}
	}

	return nil, fmt.Errorf("binary %q not found in archive", binaryName)
}
