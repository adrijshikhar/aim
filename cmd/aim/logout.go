package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/logger"
	"github.com/aim-cli/aim/internal/oauth"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/tui"
	"github.com/aim-cli/aim/internal/usage"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

func newLogoutCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	return &cobra.Command{
		Use:   "logout <agent> <profile>",
		Short: "Clear profile credentials without modifying other profiles or host Keychain",
		Long: `Log out an agent profile cleanly by removing its profile-local credentials,
deleting only its profile-scoped Keychain item (never host or sibling profile items),
and marking it logged out so subsequent runs never silently re-seed host credentials.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			exitCode := executeLogoutTo(cmd.Context(), cmd.OutOrStdout(), cmd.ErrOrStderr(), reg, pm, args[0], args[1])
			if exitCode != 0 {
				return &ExitError{Code: exitCode}
			}
			return nil
		},
		ValidArgsFunction: func(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return completeAgents(reg, toComplete), cobra.ShellCompDirectiveNoFileComp
			}
			if len(args) == 1 {
				return completeProfilesForAgent(pm, reg, args[0], toComplete), cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}
}

func executeLogoutTo(ctx context.Context, out, errOut io.Writer, reg *agents.Registry, pm *profile.ProfileManager, agentName, profileName string) int {
	errBadge := lipgloss.NewStyle().Foreground(tui.StatusRed).Bold(true).Render("✖")
	adapter, resolvedPM, pDir, ok := resolveLogoutTarget(ctx, errOut, errBadge, reg, pm, agentName, profileName)
	if !ok {
		return 1
	}

	logger.Debug("[logout] Logging out agent=%q profile=%q dir=%s", adapter.Name(), profileName, pDir)

	if err := clearProfileAgentAuth(ctx, adapter, profileName, pDir); err != nil {
		_, _ = fmt.Fprintf(errOut, "%s Error: %v\n", errBadge, err)
		return 1
	}

	baseDir := config.BaseDir()
	if resolvedPM.BaseDir != "" {
		baseDir = resolvedPM.BaseDir
	}
	_ = oauth.NewRefusalStore(baseDir).Clear(adapter.Name(), profileName)
	usage.NewCacheStore(baseDir, usage.DefaultTTL).Delete(adapter.Name(), profileName)

	okBadge := lipgloss.NewStyle().Foreground(tui.StatusGreen).Bold(true).Render("✔")
	cyan := lipgloss.NewStyle().Foreground(tui.AccentCyan)
	bold := lipgloss.NewStyle().Bold(true).Foreground(tui.TextBright)
	_, _ = fmt.Fprintf(out, "%s Logged out %s profile %s (profile state preserved)\n",
		okBadge,
		bold.Render(adapter.Name()),
		cyan.Render(profileName),
	)
	return 0
}

func resolveLogoutTarget(_ context.Context, errOut io.Writer, errBadge string, reg *agents.Registry, pm *profile.ProfileManager, agentName, profileName string) (agents.AgentAdapter, *profile.ProfileManager, string, bool) {
	if reg == nil {
		reg = defaultRegistry()
	}
	adapter, err := reg.Get(agentName)
	if err != nil {
		_, _ = fmt.Fprintf(errOut, "%s Error: %v\n", errBadge, err)
		return nil, nil, "", false
	}
	if err := profile.ValidateProfileName(profileName); err != nil {
		_, _ = fmt.Fprintf(errOut, "%s Error: %v\n", errBadge, err)
		return nil, nil, "", false
	}
	if pm == nil {
		pm = profile.NewProfileManager(config.BaseDir())
	}

	pDir := pm.ProfileDir(profileName)
	cfg, _ := config.LoadConfig()
	_, dirErr := os.Stat(pDir)
	dirExists := dirErr == nil
	cfgExists := cfg != nil && cfg.HasAgent(profileName, adapter.Name())
	if !dirExists && !cfgExists {
		_, _ = fmt.Fprintf(errOut, "%s Error: profile %q does not exist\n", errBadge, profileName)
		return nil, nil, "", false
	}
	if !dirExists {
		if _, err := pm.EnsureProfile(profileName); err != nil {
			_, _ = fmt.Fprintf(errOut, "%s Error preparing profile %q: %v\n", errBadge, profileName, err)
			return nil, nil, "", false
		}
	}
	return adapter, pm, pDir, true
}

func clearProfileAgentAuth(ctx context.Context, adapter agents.AgentAdapter, _ string, pDir string) error {
	var purgeErr error
	if purger, ok := adapter.(agents.CredentialPurger); ok {
		purgeErr = profile.PurgeProfileKeychainWithContext(ctx, adapter.Name(), pDir, purger)
	} else {
		purgeErr = profile.PurgeProfileKeychainWithContext(ctx, adapter.Name(), pDir)
	}
	if purgeErr != nil {
		return fmt.Errorf("clear scoped keychain: %w", purgeErr)
	}
	if err := profile.PurgeLocalAgentCredentials(pDir, adapter.Name()); err != nil {
		return fmt.Errorf("clear local credentials: %w", err)
	}
	if err := profile.MarkAgentLoggedOut(pDir, adapter.Name()); err != nil {
		return fmt.Errorf("record logout state: %w", err)
	}
	return nil
}
