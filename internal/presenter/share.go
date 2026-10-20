package presenter

import (
	"fmt"
	"io"

	"github.com/aim-cli/aim/internal/share"
	"github.com/aim-cli/aim/internal/tui"
	"github.com/charmbracelet/lipgloss"
)

// RenderSharedProfilesTable renders the shared profiles in a bounded Lipgloss table.
func RenderSharedProfilesTable(w io.Writer, profiles []share.SharedProfileConfig) string {
	t := NewTable(w).
		Border(lipgloss.HiddenBorder()).
		Headers("PROFILE", "AGENT", "AUTO-APPROVE", "MAX TTL", "STATUS")

	for _, p := range profiles {
		status := lipgloss.NewStyle().Foreground(tui.StatusGreen).Render("Shared")
		auto := lipgloss.NewStyle().Foreground(tui.TextDim).Render("Manual")
		if p.AutoApprove {
			auto = lipgloss.NewStyle().Foreground(tui.AccentCyan).Render("Auto")
		}
		t.Row(
			lipgloss.NewStyle().Foreground(tui.TextBright).Render(p.Name),
			p.Agent,
			auto,
			p.MaxLeaseTTL.String(),
			status,
		)
	}
	return t.Render()
}

// RenderPeersTable renders discovered network peers in a bounded Lipgloss table.
func RenderPeersTable(w io.Writer, peers []share.PeerInfo) string {
	t := NewTable(w).
		Border(lipgloss.HiddenBorder()).
		Headers("PEER", "VPN IP", "PROFILES", "LAST SEEN")

	for _, p := range peers {
		profs := fmt.Sprintf("%v", p.Profiles)
		t.Row(
			lipgloss.NewStyle().Foreground(tui.TextBright).Render(p.HostName),
			lipgloss.NewStyle().Foreground(tui.AccentCyan).Render(p.VPNIP),
			profs,
			p.LastSeen.Format("15:04:05"),
		)
	}
	return t.Render()
}
