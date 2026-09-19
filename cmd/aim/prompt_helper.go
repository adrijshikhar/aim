package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aim-cli/aim/internal/profile"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

var isInteractiveFunc = func(r io.Reader) bool {
	if f, ok := r.(*os.File); ok {
		return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
	}
	return false
}

// confirmProfileExists checks if a profile exists.
// If it does not exist:
//   - If running interactively, prompts the user:
//     "Profile \"<name>\" does not exist. Do you want to create it and <actionDesc>? [y/N]: "
//     If profileName contains "--", prints a warning suggesting the user might have missed a space before a flag.
//     Returns (true, nil) if the user confirms with "y"/"yes", or (false, nil) if the user cancels.
//   - If non-interactive, returns an error to prevent silent profile pollution in automated scripts.
func confirmProfileExists(cmd *cobra.Command, pm *profile.ProfileManager, profileName, actionDesc string) (bool, error) {
	if pm.ProfileExists(profileName) {
		return true, nil
	}

	if isInteractiveFunc(cmd.InOrStdin()) {
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

	return false, fmt.Errorf("profile %q does not exist; run 'aim login <agent> %s' or run interactively to create it", profileName, profileName)
}
