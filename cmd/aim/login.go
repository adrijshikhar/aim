package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/logger"
	"github.com/aim-cli/aim/internal/oauth"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/telemetry"
	"github.com/aim-cli/aim/internal/usage"
	"github.com/spf13/cobra"
)

func newLoginCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	return &cobra.Command{
		Use:   "login <agent> <profile>",
		Short: "Authenticate new account via OAuth PKCE",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			exitCode := executeLoginWithIO(ctx, cmd.OutOrStdout(), cmd.ErrOrStderr(), reg, pm, args[0], args[1])
			if exitCode != 0 {
				return &ExitError{Code: exitCode}
			}
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completeAgentAndProfile(reg, pm, args, toComplete)
		},
	}
}

func executeLogin(reg *agents.Registry, pm *profile.ProfileManager, agentName, profileName string) int {
	return executeLoginWithIO(context.Background(), os.Stdout, os.Stderr, reg, pm, agentName, profileName)
}

func executeLoginWithContext(ctx context.Context, reg *agents.Registry, pm *profile.ProfileManager, agentName, profileName string) int {
	return executeLoginWithIO(ctx, os.Stdout, os.Stderr, reg, pm, agentName, profileName)
}

func executeLoginWithIO(ctx context.Context, _ io.Writer, errOut io.Writer, reg *agents.Registry, pm *profile.ProfileManager, agentName, profileName string) int {
	if ctx == nil {
		ctx = context.Background()
	}
	if errOut == nil {
		errOut = os.Stderr
	}
	if reg == nil {
		reg = defaultRegistry()
	}
	if pm == nil {
		_, _ = fmt.Fprintln(errOut, "Error: profile manager not initialized")
		return 1
	}

	logger.Debug("[login] Starting login for agent=%q, profile=%q", agentName, profileName)
	adapter, err := reg.Get(agentName)
	if err != nil {
		_, _ = fmt.Fprintf(errOut, "Error: %v\n", err)
		return 1
	}
	auth, ok := adapter.(agents.Authenticator)
	if !ok {
		_, _ = fmt.Fprintf(errOut, "Agent %s does not support login\n", adapter.Name())
		return 1
	}

	cfg, err := loadOrRecoverConfig(errOut)
	if err != nil {
		_, _ = fmt.Fprintf(errOut, "Error loading config: %v\n", err)
		return 1
	}

	pDir, err := pm.EnsureProfile(profileName)
	if err != nil {
		_, _ = fmt.Fprintf(errOut, "Error preparing profile: %v\n", err)
		return 1
	}
	if err := auth.Login(ctx, profileName, pDir); err != nil {
		_, _ = fmt.Fprintf(errOut, "Login failed: %v\n", err)
		return 1
	}

	return finalizeLoginState(ctx, errOut, pm, adapter, cfg, agentName, profileName, pDir)
}

func finalizeLoginState(
	ctx context.Context,
	errOut io.Writer,
	pm *profile.ProfileManager,
	adapter agents.AgentAdapter,
	cfg *config.Config,
	agentName, profileName, pDir string,
) int {
	// Only clear logout state and finalize credentials once Login succeeds cleanly.
	_ = profile.ClearAgentLoggedOut(pDir, adapter.Name())
	if !adapter.HasCredentials(pDir) {
		if harvester, ok := adapter.(agents.CredentialHarvester); ok {
			_ = harvester.HarvestCredentials(ctx, profileName, pDir)
		} else {
			_ = profile.HarvestKeychainTokenToProfile(adapter.Name(), pDir)
		}
	}

	acc := profile.GetProfileAccountInfoForAgent(pDir, adapter.Name())
	telemetry.IdentifyAccount(adapter.Name(), acc)
	warnDuplicateAccountEmail(errOut, pm, adapter.Name(), acc.Email, profileName)

	baseDir := config.BaseDir()
	if pm.BaseDir != "" {
		baseDir = pm.BaseDir
	}
	refStore := oauth.NewRefusalStore(baseDir)
	_ = refStore.Clear(adapter.Name(), profileName)
	cacheStore := usage.NewCacheStore(baseDir, usage.DefaultTTL)
	cacheStore.Delete(adapter.Name(), profileName)
	if agentName != adapter.Name() {
		_ = refStore.Clear(agentName, profileName)
		cacheStore.Delete(agentName, profileName)
	}

	cfg.AddProfileAgent(profileName, adapter.Name())
	if err := config.SaveConfig(cfg); err != nil {
		_, _ = fmt.Fprintf(errOut, "Error saving config: %v\n", err)
		return 1
	}
	return 0
}

func warnDuplicateAccountEmail(errOut io.Writer, pm *profile.ProfileManager, agentName, email, profileName string) {
	if email == "" || pm == nil {
		return
	}
	if errOut == nil {
		errOut = os.Stderr
	}
	existingProfiles, err := pm.ListConfiguredProfiles()
	if err != nil {
		return
	}
	for _, otherProf := range existingProfiles {
		if otherProf == profileName {
			continue
		}
		otherDir := pm.ProfileDir(otherProf)
		otherAcc := profile.GetProfileAccountInfoForAgent(otherDir, agentName)
		if otherAcc.Email != "" && strings.EqualFold(otherAcc.Email, email) {
			_, _ = fmt.Fprintf(errOut, "\nNotice: Profile %q authenticated with %s (already used by profile %q)\n", profileName, email, otherProf)
			return
		}
	}
}
