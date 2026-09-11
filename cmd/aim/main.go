package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/agents/agy"
	"github.com/aim-cli/aim/internal/agents/gemini"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
)

func dispatch(args []string, reg *agents.Registry, pm *profile.ProfileManager) int {
	rootCmd := newRootCmd(reg, pm)

	normalizedArgs := args
	if len(normalizedArgs) > 0 && normalizedArgs[0] == "__complete" {
		if len(normalizedArgs) == 1 {
			normalizedArgs = []string{"__complete", ""}
		} else {
			last := normalizedArgs[len(normalizedArgs)-1]
			if last != "" {
				foundCmd, _, err := rootCmd.Find([]string{last})
				isSub := err == nil && foundCmd != nil && foundCmd != rootCmd
				isAg := false
				if reg != nil {
					_, err := reg.Get(last)
					isAg = err == nil
				}
				if isSub || isAg {
					normalizedArgs = append(normalizedArgs, "")
				}
			}
		}
	}

	rootCmd.SetArgs(normalizedArgs)

	err := rootCmd.Execute()
	if err == nil {
		return 0
	}

	var exitErr *ExitError
	if errors.As(err, &exitErr) {
		return exitErr.Code
	}

	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	return 1
}

func main() {
	reg := agents.NewRegistry()
	reg.Register(agy.NewAdapter())
	reg.Register(gemini.NewAdapter())

	pm := profile.NewProfileManager(config.BaseDir())

	os.Exit(dispatch(os.Args[1:], reg, pm))
}
