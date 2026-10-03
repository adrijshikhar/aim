package presenter

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/aim-cli/aim/internal/tui"
	"github.com/charmbracelet/lipgloss"
)

// Spinner manages an ANSI terminal spinner for long-running CLI operations.
type Spinner struct {
	out    io.Writer
	frames []string
	delay  time.Duration
	color  lipgloss.Color

	mu     sync.Mutex
	text   string
	done   chan struct{}
	active bool
	wg     sync.WaitGroup
}

// NewSpinner creates a new terminal progress spinner.
func NewSpinner(out io.Writer, initialText string) *Spinner {
	if out == nil {
		out = os.Stderr
	}
	return &Spinner{
		out:    out,
		frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		delay:  80 * time.Millisecond,
		color:  tui.AccentPurple,
		text:   initialText,
		done:   make(chan struct{}),
	}
}

// Start launches the background goroutine rendering spinner frames.
func (s *Spinner) Start() {
	s.mu.Lock()
	if s.active {
		s.mu.Unlock()
		return
	}
	s.active = true
	s.done = make(chan struct{})
	s.wg.Add(1)
	s.mu.Unlock()

	go func() {
		defer s.wg.Done()
		frameIdx := 0
		ticker := time.NewTicker(s.delay)
		defer ticker.Stop()
		for {
			select {
			case <-s.done:
				s.mu.Lock()
				fmt.Fprintf(s.out, "\r\033[K")
				s.mu.Unlock()
				return
			case <-ticker.C:
				s.mu.Lock()
				txt := s.text
				spin := lipgloss.NewStyle().Foreground(s.color).Render(s.frames[frameIdx%len(s.frames)])
				fmt.Fprintf(s.out, "\r\033[K%s %s", spin, txt)
				s.mu.Unlock()
				frameIdx++
			}
		}
	}()
}

// UpdateText safely updates the status message shown next to the spinner.
func (s *Spinner) UpdateText(text string) {
	s.mu.Lock()
	s.text = text
	s.mu.Unlock()
}

// Stop terminates the spinner, clears its terminal line, and waits for the background worker to exit.
func (s *Spinner) Stop() {
	s.mu.Lock()
	if !s.active {
		s.mu.Unlock()
		return
	}
	s.active = false
	close(s.done)
	s.mu.Unlock()

	s.wg.Wait()
}
