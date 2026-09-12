package profile

import (
	"io"
	"os"
	"path/filepath"
)

func copyFileFallback(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

// cloneDirectory recursively copies files and symlinks from srcDir to dstDir,
// using APFS clonefile (reflink) on macOS when possible and skipping sensitive files.
func cloneDirectory(srcDir, dstDir string) error {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		name := entry.Name()
		if isSensitiveProfileFile(name) {
			continue
		}
		srcPath := filepath.Join(srcDir, name)
		dstPath := filepath.Join(dstDir, name)

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(srcPath)
			if err != nil {
				continue
			}
			_ = os.Remove(dstPath)
			_ = os.Symlink(target, dstPath)
			continue
		}

		if entry.IsDir() {
			if err := os.MkdirAll(dstPath, info.Mode().Perm()); err != nil {
				return err
			}
			if err := cloneDirectory(srcPath, dstPath); err != nil {
				return err
			}
			_ = os.Chtimes(dstPath, info.ModTime(), info.ModTime())
			continue
		}

		// Regular file
		_ = os.Remove(dstPath)
		if err := cloneOrCopyFile(srcPath, dstPath); err != nil {
			continue
		}
		_ = os.Chmod(dstPath, info.Mode().Perm())
		_ = os.Chtimes(dstPath, info.ModTime(), info.ModTime())
	}
	return nil
}
