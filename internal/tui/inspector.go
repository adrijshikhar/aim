package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/usage"
	"github.com/charmbracelet/lipgloss"
)

// formatWin formats a single usage limit window with gauge bar, percentage, and reset info.
func formatWin(w usage.LimitWindow, label string) string {
	bar := usage.RenderBar(w.RemainingPct, 10)
	winStatus := usage.CalculateStatus([]usage.LimitWindow{w})
	gaugeStyle := GaugeStyleForStatus(winStatus)

	resetInfo := ""
	if w.RemainingPct < 100 && w.ResetsIn > 0 {
		resetInfo = lipgloss.NewStyle().Foreground(TextMuted).Render(fmt.Sprintf(" (%s)", usage.FormatDuration(w.ResetsIn)))
	}
	prefix := ""
	if label != "" {
		prefix = lipgloss.NewStyle().Foreground(TextDim).Render(label + ": ")
	}
	return fmt.Sprintf("%s%s %s%s",
		prefix,
		gaugeStyle.Render(bar),
		gaugeStyle.Render(fmt.Sprintf("%d%%", w.RemainingPct)),
		resetInfo,
	)
}

// renderInspector renders the bottom profile detail card for curProfile.
func (m Model) renderInspector(curProfile string) string {
	if curProfile == "" {
		return ""
	}

	var s strings.Builder
	s.WriteString("\n  " + lipgloss.NewStyle().Foreground(TextDim).Render("── Profile Details: "+curProfile+" ──") + "\n")
	rep, hasReport := m.getReport(curProfile)
	lblStyle := lipgloss.NewStyle().Width(14).Foreground(TextSecondary)

	// Resolve account info from credentials or report
	var pDir string
	if m.pm != nil {
		pDir = m.pm.ProfileDir(curProfile)
	}
	accInfo := profile.GetProfileAccountInfoForAgent(pDir, m.agent)
	accountEmail := accInfo.Email
	accountName := accInfo.Name
	authMethod := accInfo.AuthMethod

	if accountEmail == "" && hasReport && rep.AccountEmail != "" {
		accountEmail = rep.AccountEmail
		accountName = rep.AccountName
		authMethod = rep.AuthMethod
	}

	hasCreds := false
	if m.reg != nil && pDir != "" {
		if ad, err := m.reg.Get(m.agent); err == nil {
			hasCreds = ad.HasCredentials(pDir)
		}
	}

	if accountEmail != "" {
		accountStr := lipgloss.NewStyle().Foreground(AccentCyan).Bold(true).Render(accountEmail)
		if accountName != "" {
			accountStr += " " + lipgloss.NewStyle().Foreground(TextMuted).Render("("+accountName+")")
		}
		s.WriteString(fmt.Sprintf("    %s %s\n", lblStyle.Render("Account:"), accountStr))
	} else if hasCreds {
		accountStr := lipgloss.NewStyle().Foreground(TextMuted).Render("active (local credentials)")
		s.WriteString(fmt.Sprintf("    %s %s\n", lblStyle.Render("Account:"), accountStr))
	} else {
		accountStr := lipgloss.NewStyle().Foreground(TextMuted).Render("[no credentials - press 'l' to log in]")
		s.WriteString(fmt.Sprintf("    %s %s\n", lblStyle.Render("Account:"), accountStr))
	}

	if authMethod != "" {
		s.WriteString(fmt.Sprintf("    %s %s\n", lblStyle.Render("Auth:"), lipgloss.NewStyle().Foreground(TextDim).Render(authMethod)))
	}

	projectID := accInfo.ProjectID
	if projectID == "" && hasReport && rep.ProjectID != "" {
		projectID = rep.ProjectID
	}
	if projectID != "" {
		s.WriteString(fmt.Sprintf("    %s %s\n", lblStyle.Render("Project:"), lipgloss.NewStyle().Foreground(TextDim).Render(projectID)))
	}

	agentsList := []string{m.agent}
	if m.cfg != nil {
		if list := m.cfg.GetProfileAgents(curProfile); len(list) > 0 {
			agentsList = list
		}
	}
	s.WriteString(fmt.Sprintf("    %s %s\n", lblStyle.Render("Agents:"), lipgloss.NewStyle().Foreground(TextDim).Render(strings.Join(agentsList, ", "))))

	if hasReport && rep.Error != "" {
		statusMsg := rep.Summary
		if statusMsg == "" {
			statusMsg = rep.Error
		}
		s.WriteString(fmt.Sprintf("    %s %s\n", lblStyle.Render("Status:"), lipgloss.NewStyle().Foreground(StatusYellow).Render(statusMsg)))
	}

	if hasReport {
		modelGroups := rep.ModelGroups()
		if len(modelGroups) > 0 {
			for _, g := range modelGroups {
				modelShort := usage.CleanModelCategory(g.Category)

				var win5h, winWk *usage.LimitWindow
				for i := range g.Windows {
					w := &g.Windows[i]
					if w.IsHourly() {
						win5h = w
					} else if w.IsWeekly() {
						winWk = w
					}
				}

				var lineContent string
				if win5h != nil && winWk != nil {
					str5h := formatWin(*win5h, "5h")
					strWk := formatWin(*winWk, "Wk")
					lineContent = lipgloss.NewStyle().Width(33).Render(str5h) + strWk
				} else if win5h != nil {
					lineContent = formatWin(*win5h, "5h")
				} else if winWk != nil {
					lineContent = formatWin(*winWk, "Wk")
				} else {
					var parts []string
					for _, w := range g.Windows {
						parts = append(parts, formatWin(w, w.Name))
					}
					lineContent = strings.Join(parts, "   ")
				}

				s.WriteString(fmt.Sprintf("    %s %s\n", lblStyle.Render(modelShort+":"), lineContent))
			}
		} else if len(rep.Windows) > 0 {
			for _, w := range rep.Windows {
				bar := usage.RenderBar(w.RemainingPct, 10)
				winStatus := usage.CalculateStatus([]usage.LimitWindow{w})
				gaugeStyle := GaugeStyleForStatus(winStatus)
				resetInfo := ""
				if w.RemainingPct < 100 && w.ResetsIn > 0 {
					resetInfo = lipgloss.NewStyle().Foreground(TextMuted).Render(fmt.Sprintf(" (resets in %s)", usage.FormatDuration(w.ResetsIn)))
				}
				primaryStr := fmt.Sprintf("%s %s%s", gaugeStyle.Render(bar), gaugeStyle.Render(fmt.Sprintf("%d%%", w.RemainingPct)), resetInfo)
				s.WriteString(fmt.Sprintf("    %s %s\n", lblStyle.Render(w.Name+":"), primaryStr))
			}
		}

		// Resets At
		resetsAtStr := "None"
		var soonestReset time.Time
		for _, w := range rep.Windows {
			if !w.ResetsAt.IsZero() {
				if soonestReset.IsZero() || (w.ResetsAt.After(time.Now()) && (soonestReset.Before(time.Now()) || w.ResetsAt.Before(soonestReset))) {
					soonestReset = w.ResetsAt
				}
			}
		}
		if !soonestReset.IsZero() {
			resetsAtStr = soonestReset.Local().Format("2006-01-02 15:04:05")
		}
		s.WriteString(fmt.Sprintf("    %s %s\n", lblStyle.Render("Resets At:"), lipgloss.NewStyle().Foreground(TextMuted).Render(resetsAtStr)))

		// Credits
		creditsStr := "0"
		if rep.Credits != "" {
			creditsStr = rep.Credits
		}
		s.WriteString(fmt.Sprintf("    %s %s\n", lblStyle.Render("Credits:"), creditsStr))

		// Refreshed
		refreshedStr := "just now"
		if !rep.FetchedAt.IsZero() {
			age := time.Since(rep.FetchedAt)
			if age >= time.Second {
				refreshedStr = usage.FormatDuration(age) + " ago"
			}
		}
		s.WriteString(fmt.Sprintf("    %s %s\n", lblStyle.Render("Refreshed:"), lipgloss.NewStyle().Foreground(TextMuted).Render(refreshedStr)))
	} else if m.loading {
		s.WriteString("    " + lipgloss.NewStyle().Foreground(TextMuted).Render("(fetching quota...)") + "\n")
	} else {
		s.WriteString("    (no quota data available - press 'r' to refresh)\n")
	}

	return s.String()
}
