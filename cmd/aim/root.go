package main

import (
	"fmt"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/spf13/cobra"
)

var (
	// Version is the current version of AIM, injected at build time via -ldflags.
	Version = "0.1.0"
	// Commit is the git commit hash at build time.
	Commit = "none"
	// Date is the build timestamp.
	Date = "unknown"
)

// ExitError represents an explicit process exit code from command execution.
type ExitError struct {
	Code int
}

func (e *ExitError) Error() string {
	return fmt.Sprintf("exit code %d", e.Code)
}

func newRootCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "aim [command] [agent] [profile] [flags] [-- args...]",
		Short: "AIM (AI Multiplexer) - Isolated Profile Manager for AI Agents",
		Long: `AIM (AI Multiplexer) v` + Version + `
Usage:
  aim [command] [agent] [profile] [flags] [-- args...]

Primary Commands:
  run <agent> <profile>      Execute agent under isolated profile
  shell <agent> <profile>    Launch subshell with profile environment
  login <agent> <profile>    Authenticate new account via OAuth PKCE
  list [agent]               List all profiles and status
  usage [agent] [profile]    Display remaining quota and usage limits
  doctor [agent]             Diagnose environment, tokens, and binaries
  remove [agent] <profile>   Delete profile credentials and state
  clone [agent] <src> <dst>  Duplicate profile settings without copying tokens
  rename <old> <new>         Rename profile preserving all tokens and state
  whoami                     Show active profile, agent, session, and quota
  ui                         Open interactive TUI dashboard (default)
  completion <shell>         Generate shell completion script (zsh, bash, fish)

Flags:
  -v, --version              Show AIM version
  -h, --help                 Show help documentation`,
		Version:       Version,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				code := runTUI(reg, pm)
				if code != 0 {
					return &ExitError{Code: code}
				}
				return nil
			}
			return cmd.Help()
		},
	}

	rootCmd.SetVersionTemplate(fmt.Sprintf("aim version %s\n", Version))

	// Add subcommands
	rootCmd.AddCommand(
		newRunCmd(reg, pm),
		newShellCmd(reg, pm),
		newLoginCmd(reg, pm),
		newListCmd(reg, pm),
		newUsageCmd(reg, pm),
		newDoctorCmd(reg, pm),
		newRemoveCmd(reg, pm),
		newCloneCmd(reg, pm),
		newRenameCmd(reg, pm),
		newWhoamiCmd(reg, pm),
		newUICmd(reg, pm),
		newPrewarmCmd(reg, pm),
		newCompletionCmd(rootCmd),
		newVersionCmd(),
	)

	return rootCmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show AIM version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("aim version %s\n", Version)
		},
	}
}

func printUsage() {
	cmd := newRootCmd(nil, nil)
	_ = cmd.Help()
}
