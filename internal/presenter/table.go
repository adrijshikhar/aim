package presenter

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/tui"
	"github.com/aim-cli/aim/internal/usage"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/charmbracelet/x/term"
)

// IsTerminal returns true and the terminal width if the writer is an active terminal
// or if an explicit terminal width is configured via COLUMNS or AIM_TERMINAL_WIDTH.
// It returns 0, false if the writer is redirected, piped, or not an interactive terminal.
func IsTerminal(w io.Writer) (int, bool) {
	for _, envVar := range []string{"COLUMNS", "AIM_TERMINAL_WIDTH"} {
		if val := strings.TrimSpace(os.Getenv(envVar)); val != "" {
			if n, err := strconv.Atoi(val); err == nil && n > 0 {
				return n, true
			}
		}
	}
	if f, ok := w.(*os.File); ok {
		if term.IsTerminal(f.Fd()) {
			if width, _, err := term.GetSize(f.Fd()); err == nil && width > 0 {
				return width, true
			}
		}
		return 0, false
	}
	if w == nil {
		if term.IsTerminal(os.Stdout.Fd()) {
			if width, _, err := term.GetSize(os.Stdout.Fd()); err == nil && width > 0 {
				return width, true
			}
		}
	}
	return 0, false
}

// TerminalWidth returns the detected terminal width from the given writer or stdout,
// defaulting to fallback (or 100 if fallback <= 0) when width cannot be determined.
func TerminalWidth(w io.Writer, fallback int) int {
	if fallback <= 0 {
		fallback = 100
	}
	if width, ok := IsTerminal(w); ok {
		return width
	}
	return fallback
}

// NewTable returns a lipgloss Table configured for the given output writer.
// When rendering to an interactive terminal, the table is bounded to the terminal width
// with automatic truncation (Wrap(false)) to prevent line wrapping and border corruption.
// When output is piped, redirected to a file, or non-terminal, table width is unconstrained
// so full paths, URLs, and fields are preserved for scripting and pipeline consumption.
func NewTable(w io.Writer) *table.Table {
	t := table.New()
	if width, ok := IsTerminal(w); ok {
		return t.Width(width).Wrap(false)
	}
	return t
}

// RenderMCPListTable formats and writes the configured MCP servers list as a formatted Lipgloss table.
func RenderMCPListTable(w io.Writer, agentName, profileName string, servers []agents.MCPServerInfo) {
	banner := fmt.Sprintf("=== Configured MCP Servers (%s: %s) ===", agentName, profileName)
	fmt.Fprintf(w, "\n%s\n\n", lipgloss.NewStyle().Bold(true).Foreground(tui.AccentBlue).Render(banner))

	if len(servers) == 0 {
		fmt.Fprintf(w, "  %s\n\n", lipgloss.NewStyle().Foreground(tui.TextMuted).Render("(no MCP servers configured)"))
		return
	}

	t := NewTable(w).
		Border(lipgloss.HiddenBorder()).
		Headers("#", "NAME", "STATUS", "AUTH", "TYPE", "TARGET / COMMAND")

	enabledCount := 0
	disabledCount := 0

	for i, s := range servers {
		numStr := fmt.Sprintf("%d", i+1)

		statusStyled := lipgloss.NewStyle().Foreground(tui.StatusGreen).Render("enabled")
		if s.Status == "disabled" {
			statusStyled = lipgloss.NewStyle().Foreground(tui.StatusRed).Render("disabled")
			disabledCount++
		} else {
			enabledCount++
		}

		authStyled := lipgloss.NewStyle().Foreground(tui.TextMuted).Render(s.Auth)
		if s.Auth == "OAuth" || s.Auth == "connected" {
			authStyled = lipgloss.NewStyle().Foreground(tui.StatusGreen).Render(s.Auth)
		} else if s.Auth == "auth required" {
			authStyled = lipgloss.NewStyle().Foreground(tui.StatusYellow).Render(s.Auth)
		}

		typeStyled := lipgloss.NewStyle().Foreground(tui.AccentCyan).Render(s.Type)

		t.Row(
			numStr,
			s.Name,
			statusStyled,
			authStyled,
			typeStyled,
			s.Target,
		)
	}

	t.StyleFunc(func(row, col int) lipgloss.Style {
		if row == table.HeaderRow {
			return lipgloss.NewStyle().Bold(true).Foreground(tui.AccentBlue)
		}
		switch col {
		case 0:
			return lipgloss.NewStyle().Foreground(tui.TextDim)
		case 1:
			return lipgloss.NewStyle().Bold(true).Foreground(tui.TextBright)
		case 5:
			return lipgloss.NewStyle().Foreground(tui.TextPrimary)
		default:
			return lipgloss.NewStyle()
		}
	})

	fmt.Fprintln(w, t.Render())
	summary := fmt.Sprintf("Total: %d server(s) configured (%d enabled, %d disabled)", len(servers), enabledCount, disabledCount)
	fmt.Fprintf(w, "\n%s\n\n", lipgloss.NewStyle().Foreground(tui.TextMuted).Render(summary))
}

// PrintTable renders headers and data rows into a styled Lipgloss table.
func PrintTable(w io.Writer, headers []string, rows [][]string, useColor bool) {
	t := NewTable(w).
		Border(lipgloss.HiddenBorder()).
		Headers(headers...).
		Rows(rows...)

	if useColor {
		t.StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return lipgloss.NewStyle().Bold(true).Foreground(tui.AccentCyan)
			}
			return lipgloss.NewStyle().Foreground(tui.TextPrimary)
		})
	}
	fmt.Fprintln(w, t.Render())
}

// RenderCLIBar formats a graphical ASCII / ANSI percentage meter.
func RenderCLIBar(pct int, width int, st usage.Status, useColor bool) string {
	if !useColor {
		return fmt.Sprintf("%s %d%%", usage.RenderBar(pct, width), pct)
	}

	if width <= 0 {
		return fmt.Sprintf("%d%%", pct)
	}

	clamped := pct
	if clamped < 0 {
		clamped = 0
	} else if clamped > 100 {
		clamped = 100
	}

	filled := (clamped * width) / 100
	empty := width - filled

	fillColor := tui.StatusGreen
	if clamped <= 15 {
		fillColor = tui.StatusRed
	} else if clamped <= 50 {
		fillColor = tui.StatusYellow
	}

	fillStyle := lipgloss.NewStyle().Foreground(fillColor)
	emptyStyle := lipgloss.NewStyle().Foreground(tui.TextMuted)
	bracketStyle := lipgloss.NewStyle().Foreground(tui.TextDim)

	bar := bracketStyle.Render("[") +
		fillStyle.Render(strings.Repeat("█", filled)) +
		emptyStyle.Render(strings.Repeat("░", empty)) +
		bracketStyle.Render("]")

	return fmt.Sprintf("%s %s", bar, fillStyle.Render(fmt.Sprintf("%d%%", pct)))
}

// RenderCLIStatus styles a usage status string with matching ANSI colors.
func RenderCLIStatus(st usage.Status, useColor bool) string {
	str := string(st)
	if !useColor {
		return str
	}
	return tui.GaugeStyleForStatus(st).Render(str)
}

// RenderDiagnosticResult styles a single DiagnosticResult row for terminal output.
func RenderDiagnosticResult(r agents.DiagnosticResult) (string, string, string) {
	var badgeStyle lipgloss.Style
	switch r.Status {
	case "OK":
		badgeStyle = tui.GaugeGreenStyle
	case "WARN":
		badgeStyle = tui.GaugeYellowStyle
	case "FAIL":
		badgeStyle = tui.GaugeRedStyle
	default:
		badgeStyle = tui.GaugeDimStyle
	}
	badge := badgeStyle.Width(8).Render(fmt.Sprintf("[%s]", r.Status))
	cat := lipgloss.NewStyle().Bold(true).Foreground(tui.TextPrimary).Render(r.Category + ":")
	msg := lipgloss.NewStyle().Foreground(tui.TextSecondary).Render(r.Message)
	return badge, cat, msg
}
