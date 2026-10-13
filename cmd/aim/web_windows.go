//go:build windows

package main

import (
	"os"
	"os/exec"
)

func setSysProcAttr(cmd *exec.Cmd) {
	// No-op or standard Windows detachment
}

func isProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Windows FindProcess checks process existence
	return proc != nil
}

func terminateProcess(pid int) error {
	if pid <= 0 {
		return nil
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Kill()
}
