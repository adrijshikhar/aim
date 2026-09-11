package profile

import (
	"errors"
	"os"
	"path/filepath"
)

var bridgedDotfiles = []string{
	".gitconfig",
	".ssh",
	".zshrc",
	".bashrc",
}

func EnsureDotfiles(realHome, profileDir string) error {
	var errs []error
	for _, name := range bridgedDotfiles {
		src := filepath.Join(realHome, name)
		if _, err := os.Lstat(src); err != nil {
			continue
		}
		dest := filepath.Join(profileDir, name)
		if _, err := os.Lstat(dest); err == nil {
			continue
		}
		if err := os.Symlink(src, dest); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
