//go:build darwin

package profile

import (
	"golang.org/x/sys/unix"
)

// cloneOrCopyFile attempts an APFS copy-on-write clone (reflink) first on macOS.
// If clonefile fails (e.g. across filesystems), it falls back to streaming copy.
func cloneOrCopyFile(src, dst string) error {
	err := unix.Clonefile(src, dst, unix.CLONE_NOFOLLOW)
	if err == nil {
		return nil
	}
	return copyFileFallback(src, dst)
}
