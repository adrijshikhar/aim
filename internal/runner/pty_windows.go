//go:build windows

package runner

import "os"

func SetupWindowSizeListener(onResize func()) func() {
	return func() {}
}

func forwardWindowSize(proc *os.Process) func() {
	return func() {}
}
