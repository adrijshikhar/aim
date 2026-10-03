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
	s.mu.Unlock()

	go func() {
		frameIdx := 0
		ticker := time.NewTicker(s.delay)
		defer ticker.Stop()
		for {
			select {
			case <-s.done:
				fmt.Fprintf(s.out, "\r\033[K")
				return
			case <-ticker.C:
				s.mu.Lock()
				txt := s.text
				s.mu.Unlock()
				spin := lipgloss.NewStyle().Foreground(s.color).Render(s.frames[frameIdx%len(s.frames)])
				fmt.Fprintf(s.out, "\r\033[K%s %s", spin, txt)
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

// Stop terminates the spinner and clears its terminal line.
func (s *Spinner) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.active {
		return
	}
	s.active = false
	close(s.done)
}
