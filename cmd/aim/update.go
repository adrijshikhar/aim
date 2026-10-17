package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/logger"
	"github.com/aim-cli/aim/internal/tui"
	"github.com/aim-cli/aim/internal/updater"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
	var (
		checkFlag      bool
		forceFlag      bool
		backgroundFlag bool
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Check for and install updates to AIM",
		Long: `Check for and install updates to AIM.

Detects whether AIM is managed by Homebrew or is a standalone installation,
and performs the appropriate upgrade.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(cmd.Context(), checkFlag, forceFlag, backgroundFlag)
		},
	}

	cmd.Flags().BoolVarP(&checkFlag, "check", "c", false, "Check for updates without installing")
	cmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Force reinstall or update even if already up-to-date")
	cmd.Flags().BoolVarP(&backgroundFlag, "background", "b", false, "Run update silently in background")

	return cmd
}

func runUpdate(ctx context.Context, check, force, background bool) error {
	curVer := GetVersion()
	cacheDir := config.CacheDir()

	info, err := updater.CheckForUpdate(ctx, curVer, cacheDir, force)
	if err != nil {
		if background {
			logger.Debug("background update check failed: %v", err)
			return nil
		}
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if info == nil {
		if background {
			return nil
		}
		return fmt.Errorf("could not determine update status")
	}

	hasUpdate := info.UpdateAvailable
	if !hasUpdate && !force {
		if background {
			return nil
		}
		if check {
			fmt.Printf("AIM is up-to-date (version %s)\n", curVer)
			return nil
		}
		fmt.Printf("AIM is already on the latest version (%s)\n", curVer)
		return nil
	}

	if check {
		fmt.Printf("A new version of aim is available: %s → %s\nRun 'aim update' to install.\n", curVer, info.LatestVersion)
		return nil
	}

	execPath, err := os.Executable()
	if err != nil {
		if background {
			logger.Debug("failed to determine executable path for update: %v", err)
			return nil
		}
		return fmt.Errorf("failed to determine executable path: %w", err)
	}

	method := updater.DetectInstallMethod(execPath)

	if method == updater.InstallMethodHomebrew {
		if !background {
			bannerStyle := lipgloss.NewStyle().Foreground(tui.AccentBlue).Bold(true)
			fmt.Println(bannerStyle.Render("Detected Homebrew installation."))
			fmt.Println("Running 'brew upgrade adrijshikhar/tap/aim'...")
		}

		var stdout, stderr io.Writer
		if !background {
			stdout = os.Stdout
			stderr = os.Stderr
		}

		if err := updater.UpgradeHomebrew(ctx, stdout, stderr); err != nil {
			if background {
				logger.Debug("background homebrew upgrade failed: %v", err)
				return nil
			}
			return fmt.Errorf("homebrew upgrade failed: %w", err)
		}
	} else {
		if !background {
			bannerStyle := lipgloss.NewStyle().Foreground(tui.AccentBlue).Bold(true)
			fmt.Println(bannerStyle.Render(fmt.Sprintf("Updating aim %s → %s...", curVer, info.LatestVersion)))
		}

		var stdout io.Writer
		if !background {
			stdout = os.Stdout
		}

		opts := updater.SelfUpdateOptions{
			TargetVersion:  info.LatestVersion,
			CurrentVersion: curVer,
			ExecPath:       execPath,
			Stdout:         stdout,
		}

		if err := updater.SelfUpdate(ctx, opts); err != nil {
			if background {
				logger.Debug("background self-update failed: %v", err)
				return nil
			}
			return fmt.Errorf("update failed: %w", err)
		}
	}

	if !background {
		successBadge := lipgloss.NewStyle().Foreground(tui.StatusGreen).Bold(true).Render("✔")
		fmt.Printf("%s Successfully updated aim to %s!\n", successBadge, info.LatestVersion)
	}

	return nil
}
