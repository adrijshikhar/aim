package main

import (
	"fmt"
	"regexp"
	"runtime/debug"
	"strings"

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
	Version = "dev"
	// Commit is the git commit hash at build time.
	Commit = "none"
	// Date is the build timestamp.
	Date = "unknown"

	debugFlag bool

	gitDescribeSuffixRegex = regexp.MustCompile(`(-\d+)?-g[0-9a-fA-F]+([-+]dirty)?$`)
	dirtySuffixRegex       = regexp.MustCompile(`[-+]dirty$`)
	commitHashOnlyRegex    = regexp.MustCompile(`^[0-9a-fA-F]{7,40}$`)
	goPseudoVersionRegex   = regexp.MustCompile(`-(0\.)?\d{14}-[0-9a-fA-F]+([-+]dirty)?$`)
)

// sanitizeVersion normalizes the version string to clean semantic versioning,
// stripping any git describe commit distance/hash suffixes (-g<commit>),
// dirty indicators, and ensuring raw commit hashes or Go module pseudo-versions
// are normalized to "dev".
func sanitizeVersion(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	if v == "" || v == "none" || v == "unknown" {
		return "dev"
	}
	// If the entire version string is just a git commit hash (e.g. from --always without tags), fall back to dev
	if commitHashOnlyRegex.MatchString(v) {
		return "dev"
	}
	// Go module VCS pseudo-version: e.g. v0.0.0-yyyymmddhhmmss-abcdef123456 or v0.12.1-0.yyyymmddhhmmss-abcdef123456+dirty
	if strings.HasPrefix(v, "0.0.0-") || goPseudoVersionRegex.MatchString(v) {
		return "dev"
	}
	// Strip any git describe distance/hash suffix (e.g., "-5-g97df544" or "-g97df544" or "-1-g97df544-dirty")
	v = gitDescribeSuffixRegex.ReplaceAllString(v, "")
	// Strip any standalone -dirty or +dirty suffix
	v = dirtySuffixRegex.ReplaceAllString(v, "")

	if v == "" {
		return "dev"
	}
	return v
}

func init() {
	if info, ok := debug.ReadBuildInfo(); ok {
		// `go install github.com/aim-cli/aim/cmd/aim@vX.Y.Z` records the module version.
		if Version == "dev" && info.Main.Version != "" && info.Main.Version != "(devel)" {
			Version = strings.TrimPrefix(info.Main.Version, "v")
		}
		if Commit == "none" {
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
	Version = sanitizeVersion(Version)
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
  login <agent> <profile>    Authenticate new account via OAuth PKCE
  list [agent]               List all profiles and status
  sessions [agent]           List active and past conversation sessions
  resume <agent> <profile>   Resume an existing session under a profile
  usage [agent] [profile]    Display remaining quota and usage limits
  doctor [agent]             Diagnose environment, tokens, and binaries
  remove [agent] <profile>   Delete profile credentials and state
  mv <agent> <src> <dst>     Move agent account and credentials between profiles
  daemon [cmd]               Manage periodic background daemon tasks
  whoami                     Show active profile, agent, session, and quota
  web                        Start the AIM web dashboard
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
				code := runTUIWithContext(cmd.Context(), reg, pm)
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
		newLoginCmd(reg, pm),
		newListCmd(reg, pm),
		newSessionsCmd(reg, pm),
		newResumeCmd(reg, pm),
		newUsageCmd(reg, pm),
		newDoctorCmd(reg, pm),
		newRemoveCmd(reg, pm),
		newMvCmd(reg, pm),
		newWhoamiCmd(reg, pm),
		newDaemonCmd(reg, pm),
		newPrewarmCmd(reg, pm),
		newFeedbackCmd(reg, pm),
		newCompletionCmd(rootCmd),
		newVersionCmd(),
		newWebCmd(reg, pm),
	)

	return rootCmd
}

func newVersionCmd() *cobra.Command {
	var verbose bool
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show AIM version",
		Run: func(cmd *cobra.Command, args []string) {
			if verbose {
				fmt.Printf("aim version %s (commit: %s, built at: %s)\n", Version, Commit, Date)
			} else {
				fmt.Printf("aim version %s\n", Version)
			}
		},
	}
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed build metadata (commit and date)")
	return cmd
}
