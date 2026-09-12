//go:build !darwin

package profile

// cloneOrCopyFile falls back to streaming copy on non-Darwin platforms.
func cloneOrCopyFile(src, dst string) error {
	return copyFileFallback(src, dst)
}
