//go:build !windows

package runner

import (
	"os"
	"os/signal"
	"syscall"
)

func SetupWindowSizeListener(onResize func()) func() {
	sigChan := make(chan os.Signal, 1)
	done := make(chan struct{})
	signal.Notify(sigChan, syscall.SIGWINCH)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-sigChan:
				if onResize != nil {
					onResize()
				}
			}
		}
	}()
	return func() {
		signal.Stop(sigChan)
		close(done)
	}
}

func forwardWindowSize(proc *os.Process) func() {
	return SetupWindowSizeListener(func() {
		if proc != nil {
			_ = proc.Signal(syscall.SIGWINCH)
		}
	})
}
