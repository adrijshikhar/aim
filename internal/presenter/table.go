package presenter

import (
	"fmt"
	"io"
	"strings"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/tui"
	"github.com/aim-cli/aim/internal/usage"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// RenderMCPListTable formats and writes the configured MCP servers list as a formatted Lipgloss table.
func RenderMCPListTable(w io.Writer, agentName, profileName string, servers []agents.MCPServerInfo) {
	banner := fmt.Sprintf("=== Configured MCP Servers (%s: %s) ===", agentName, profileName)
	fmt.Fprintf(w, "\n%s\n\n", lipgloss.NewStyle().Bold(true).Foreground(tui.AccentBlue).Render(banner))

	if len(servers) == 0 {
		fmt.Fprintf(w, "  %s\n\n", lipgloss.NewStyle().Foreground(tui.TextMuted).Render("(no MCP servers configured)"))
		return
	}

	t := table.New().
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
	t := table.New().
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
