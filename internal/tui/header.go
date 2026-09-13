package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderTabBar() string {
	agents := m.getRegisteredAgentNames()
	var tabs []string

	for i, ag := range agents {
		num := i + 1
		var label string
		switch ag {
		case "agy":
			label = fmt.Sprintf("[%d] Antigravity (agy)", num)
		case "gemini":
			label = fmt.Sprintf("[%d] Gemini", num)
		case "codex":
			label = fmt.Sprintf("[%d] Codex", num)
		default:
			disp := ag
			if m.reg != nil {
				if ad, err := m.reg.Get(ag); err == nil && ad != nil {
					disp = ad.DisplayName()
				}
			}
			label = fmt.Sprintf("[%d] %s", num, disp)
		}

		if m.agent == ag {
			tabs = append(tabs, TabActiveStyle.Render(label))
		} else {
			tabs = append(tabs, TabInactiveStyle.Render(label))
		}
	}

	hasClaude := false
	for _, ag := range agents {
		if ag == "claude" {
			hasClaude = true
			break
		}
	}
	if !hasClaude {
		tabs = append(tabs, TabInactiveStyle.Render(fmt.Sprintf("[%d] Claude", len(agents)+1)))
	}

	return fmt.Sprintf("  %s\n\n", strings.Join(tabs, "   "))
}

func (m Model) formatVersionTag() string {
	v := m.Version()
	if v == "" {
		return ""
	}
	if !strings.HasPrefix(v, "v") {
		return "v" + v
	}
	return v
}

func (m Model) renderHeader() string {
	logoStyle := lipgloss.NewStyle().Bold(true).Foreground(StatusGreen)
	urlStyle := lipgloss.NewStyle().Foreground(AccentPurple)
	tagStyle := lipgloss.NewStyle().Foreground(StatusGreen)
	versionStyle := lipgloss.NewStyle().Foreground(TextSecondary)

	logoLines := []string{
		"    _    ___ __  __ ",
		"   / \\  |_ _|  \\/  |",
		"  / _ \\  | || |\\/| |",
		" / ___ \\ | || |  | |",
		"/_/   \\_\\___|_|  |_|",
	}

	tagLine := tagStyle.Render("AIM — AI Multiplexer")
	if v := m.formatVersionTag(); v != "" {
		tagLine += " " + versionStyle.Render(v)
	}

	var out strings.Builder
	if m.width > 0 && m.width < 70 {
		for _, l := range logoLines {
			out.WriteString(logoStyle.Render(l) + "\n")
		}
		out.WriteString("  " + urlStyle.Render("https://github.com/adrijshikhar/aim") + "\n")
		out.WriteString("  " + tagLine + "\n\n")
		return out.String()
	}

	for i, l := range logoLines {
		renderedLogo := logoStyle.Render(l)
		if i == 2 {
			out.WriteString(fmt.Sprintf("%s   %s\n", renderedLogo, urlStyle.Render("https://github.com/adrijshikhar/aim")))
		} else if i == 3 {
			out.WriteString(fmt.Sprintf("%s   %s\n", renderedLogo, tagLine))
		} else {
			out.WriteString(renderedLogo + "\n")
		}
	}
	out.WriteString("\n")
	return out.String()
}
