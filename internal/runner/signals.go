package runner

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func setupSignalForwarding(proc *os.Process, isTerminal bool) func() {
	sigChan := make(chan os.Signal, 1)
	done := make(chan struct{})
	// When connected to an interactive terminal, DO NOT forward SIGINT (os.Interrupt)
	// because the terminal line discipline already delivers it to the foreground pgid.
	if isTerminal {
		signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGHUP)
	} else {
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	}

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

	var once sync.Once
	return func() {
		once.Do(func() {
			signal.Stop(sigChan)
			close(done)
			if stopResize != nil {
				stopResize()
			}
		})
	}
}

