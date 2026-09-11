package runner

import (
	"os"
	"os/signal"
	"syscall"
)

func setupSignalForwarding(proc *os.Process) func() {
	sigChan := make(chan os.Signal, 1)
	done := make(chan struct{})
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	stopResize := forwardWindowSize(proc)

	go func() {
		for {
			select {
			case <-done:
				return
			case sig := <-sigChan:
				if proc != nil {
					_ = proc.Signal(sig)
				}
			}
		}
	}()

	return func() {
		signal.Stop(sigChan)
		close(done)
		if stopResize != nil {
			stopResize()
		}
	}
}
