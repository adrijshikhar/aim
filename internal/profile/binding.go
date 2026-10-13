package profile

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const ProfileFileName = ".aim-profile"

// BindingResolution holds the resolved profile, source .aim-profile path, and containing directory.
type BindingResolution struct {
	Profile    string `json:"profile"`
	SourceFile string `json:"source_file"`
	Directory  string `json:"directory"`
}

// ResolveDirProfile ascends from dir upwards looking for an .aim-profile file.
// It stops ascending if it encounters a .git boundary or the filesystem root.
// It parses the file line-by-line, ignoring comments (#...) and blank lines,
// and validates the profile name against path traversal.
func ResolveDirProfile(dir string) (*BindingResolution, error) {
	curr, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	for {
		targetFile := filepath.Join(curr, ProfileFileName)
		data, err := os.ReadFile(targetFile)
		if err == nil {
			scanner := bufio.NewScanner(bytes.NewReader(data))
			for scanner.Scan() {
				line := scanner.Text()
				if idx := strings.Index(line, "#"); idx >= 0 {
					line = line[:idx]
				}
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				if err := ValidateProfileName(line); err != nil {
					return nil, fmt.Errorf("invalid profile name %q in %s: %w", line, targetFile, err)
				}
				return &BindingResolution{
					Profile:    line,
					SourceFile: targetFile,
					Directory:  curr,
				}, nil
			}
			if err := scanner.Err(); err != nil {
				return nil, fmt.Errorf("error reading %s: %w", targetFile, err)
			}
		}

		// Stop if reached .git repository boundary
		if _, err := os.Stat(filepath.Join(curr, ".git")); err == nil {
			break
		}

		// Stop if reached filesystem root
		parent := filepath.Dir(curr)
		if parent == curr {
			break
		}
		curr = parent
	}

	return nil, nil
}

// BindDirectory writes the specified profile name into .aim-profile in dir.
func BindDirectory(dir, profileName string) (string, error) {
	cleanName := strings.TrimSpace(profileName)
	if err := ValidateProfileName(cleanName); err != nil {
		return "", err
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	targetPath := filepath.Join(absDir, ProfileFileName)
	content := fmt.Sprintf("%s\n", cleanName)
	if err := os.WriteFile(targetPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write %s: %w", ProfileFileName, err)
	}

	return targetPath, nil
}

// UnbindDirectory removes the .aim-profile file in dir if present.
func UnbindDirectory(dir string) (string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	targetPath := filepath.Join(absDir, ProfileFileName)
	if err := os.Remove(targetPath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("no %s found in %s", ProfileFileName, absDir)
		}
		return "", err
	}
	return targetPath, nil
}
