package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/tui"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

var isInteractiveFunc = func(r io.Reader) bool {
	if f, ok := r.(*os.File); ok {
		return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
	}
	return false
}

func isAutoCreateEnv() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv("AIM_AUTO_CREATE")))
	if v == "1" || v == "true" || v == "yes" {
		return true
	}
	y := strings.TrimSpace(strings.ToLower(os.Getenv("AIM_YES")))
	return y == "1" || y == "true" || y == "yes"
}

// confirmProfileExists checks if a profile exists.
// If it does not exist:
//   - If autoCreate is true or AIM_AUTO_CREATE=1 / AIM_YES=1 is set, creates without prompting.
//   - If running interactively, prompts the user:
//     "Profile \"<name>\" does not exist. Do you want to create it and <actionDesc>? [y/N]: "
//     If profileName contains "--", prints a warning suggesting the user might have missed a space before a flag.
//     Returns (true, nil) if the user confirms with "y"/"yes", or (false, nil) if the user cancels.
//   - If non-interactive, returns an error to prevent silent profile pollution in automated scripts.
func confirmProfileExists(cmd *cobra.Command, pm *profile.ProfileManager, profileName, actionDesc string, autoCreate bool) (bool, error) {
	if pm.ProfileExists(profileName) {
		return true, nil
	}

	if strings.Contains(profileName, "--") && (autoCreate || isAutoCreateEnv()) {
		parts := strings.SplitN(profileName, "--", 2)
		return false, fmt.Errorf("refusing to auto-create profile %q containing \"--\" (did you mean %q with flag %q?)", profileName, parts[0], "--"+parts[1])
	}

	if autoCreate || isAutoCreateEnv() {
		return true, nil
	}

	if isInteractiveFunc(cmd.InOrStdin()) {
		isTTY := false
		if f, ok := cmd.OutOrStdout().(*os.File); ok {
			isTTY = isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
		}

		if isTTY {
			if strings.Contains(profileName, "--") {
				parts := strings.SplitN(profileName, "--", 2)
				warnBadge := lipgloss.NewStyle().Foreground(tui.StatusYellow).Bold(true).Render("▲ Warning:")
				profStyle := lipgloss.NewStyle().Foreground(tui.AccentCyan).Render(fmt.Sprintf("%q", profileName))
				meanProf := lipgloss.NewStyle().Foreground(tui.AccentCyan).Bold(true).Render(fmt.Sprintf("%q", parts[0]))
				meanFlag := lipgloss.NewStyle().Foreground(tui.TextBright).Bold(true).Render(fmt.Sprintf("%q", "--"+parts[1]))
				fmt.Fprintf(cmd.OutOrStdout(), "%s profile name %s contains \"--\". Did you mean %s with flag %s?\n", warnBadge, profStyle, meanProf, meanFlag)
			}
			promptBadge := lipgloss.NewStyle().Foreground(tui.AccentBlue).Bold(true).Render("?")
			profStyle := lipgloss.NewStyle().Foreground(tui.AccentCyan).Bold(true).Render(fmt.Sprintf("%q", profileName))
			actionStyle := lipgloss.NewStyle().Foreground(tui.TextBright).Render(actionDesc)
			choiceStyle := lipgloss.NewStyle().Foreground(tui.TextMuted).Render("[y/N]: ")
			fmt.Fprintf(cmd.OutOrStdout(), "%s Profile %s does not exist. Do you want to create it and %s? %s", promptBadge, profStyle, actionStyle, choiceStyle)
			reader := bufio.NewReader(cmd.InOrStdin())
			ans, _ := reader.ReadString('\n')
			ans = strings.TrimSpace(strings.ToLower(ans))
			if ans == "y" || ans == "yes" {
				return true, nil
			}
			abortBadge := lipgloss.NewStyle().Foreground(tui.StatusRed).Bold(true).Render("✖")
			mutedStyle := lipgloss.NewStyle().Foreground(tui.TextMuted)
			fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", abortBadge, mutedStyle.Render("Profile creation aborted."))
			return false, nil
		}

		// Plain text fallback (for non-TTY stdout, pipes, or tests)
		if strings.Contains(profileName, "--") {
			parts := strings.SplitN(profileName, "--", 2)
			fmt.Fprintf(cmd.OutOrStdout(), "Warning: profile name %q contains \"--\". Did you mean %q with flag %q?\n", profileName, parts[0], "--"+parts[1])
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Profile %q does not exist. Do you want to create it and %s? [y/N]: ", profileName, actionDesc)
		reader := bufio.NewReader(cmd.InOrStdin())
		ans, _ := reader.ReadString('\n')
		ans = strings.TrimSpace(strings.ToLower(ans))
		if ans == "y" || ans == "yes" {
			return true, nil
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Profile creation aborted.")
		return false, nil
	}

	if strings.Contains(profileName, "--") {
		parts := strings.SplitN(profileName, "--", 2)
		return false, fmt.Errorf("profile %q does not exist (did you mean %q with flag %q?); run 'aim login <agent> %s' or run interactively to create it", profileName, parts[0], "--"+parts[1], profileName)
	}

	return false, fmt.Errorf("profile %q does not exist; run 'aim login <agent> %s' or run interactively to create it", profileName, profileName)
}
