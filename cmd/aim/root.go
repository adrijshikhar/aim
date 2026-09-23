package main

import (
	"fmt"
	"runtime/debug"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/agents/agy"
	"github.com/aim-cli/aim/internal/agents/claude"
	"github.com/aim-cli/aim/internal/agents/codex"
	"github.com/aim-cli/aim/internal/agents/gemini"
	"github.com/aim-cli/aim/internal/logger"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/tui"
	"github.com/spf13/cobra"
)

var (
	// Version is the current version of AIM, injected at build time via -ldflags.
	Version = "0.4.0"
	// Commit is the git commit hash at build time.
	Commit = "none"
	// Date is the build timestamp.
	Date = "unknown"

	debugFlag bool
)

func init() {
	if Commit == "none" {
		if info, ok := debug.ReadBuildInfo(); ok {
			for _, setting := range info.Settings {
				if setting.Key == "vcs.revision" && Commit == "none" {
					Commit = setting.Value
					if len(Commit) > 7 {
						Commit = Commit[:7]
					}
				}
				if setting.Key == "vcs.time" && Date == "unknown" {
					Date = setting.Value
				}
			}
		}
	}
	tui.Version = Version
}

// ExitError represents an explicit process exit code from command execution.
type ExitError struct {
	Code int
}

func (e *ExitError) Error() string {
	return fmt.Sprintf("exit code %d", e.Code)
}

// defaultRegistry returns a new registry populated with all core agent adapters.
func defaultRegistry() *agents.Registry {
	reg := agents.NewRegistry()
	reg.Register(agy.NewAdapter())
	reg.Register(gemini.NewAdapter())
	reg.Register(codex.NewAdapter())
	reg.Register(claude.NewAdapter())
	return reg
}

func newRootCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	if reg == nil {
		reg = defaultRegistry()
	}
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
  sessions [agent]           List active and past conversation sessions
  resume <agent> <profile>   Resume an existing session under a profile
  usage [agent] [profile]    Display remaining quota and usage limits
  doctor [agent]             Diagnose environment, tokens, and binaries
  remove [agent] <profile>   Delete profile credentials and state
  clone [agent] <src> <dst>  Duplicate profile settings without copying tokens
  mv <agent> <src> <dst>     Move agent account and credentials between profiles
  whoami                     Show active profile, agent, session, and quota
  completion <shell>         Generate shell completion script (zsh, bash, fish)

Flags:
  -v, --version              Show AIM version
  -h, --help                 Show help documentation
      --debug                Enable verbose debug logging`,
		Version:       Version,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if debugFlag {
				logger.SetDebug(true)
			}
		},
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
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Enable verbose debug logging")

	// Add subcommands
	rootCmd.AddCommand(
		newRunCmd(reg, pm),
		newShellCmd(reg, pm),
		newLoginCmd(reg, pm),
		newListCmd(reg, pm),
		newSessionsCmd(reg, pm),
		newResumeCmd(reg, pm),
		newUsageCmd(reg, pm),
		newDoctorCmd(reg, pm),
		newRemoveCmd(reg, pm),
		newCloneCmd(reg, pm),
		newMvCmd(reg, pm),
		newWhoamiCmd(reg, pm),
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
