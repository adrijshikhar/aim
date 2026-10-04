package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/daemon"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/tui"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

func newDaemonCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "daemon [command]",
		Short: "Manage periodic background daemon tasks (e.g. quota refresh)",
		Long: `Manage the AIM background daemon for periodic tasks such as pre-warming quota caches.

The daemon integrates with the host operating system's native service manager
(launchd on macOS, systemd user timers on Linux) to execute periodic tasks
every 15 minutes without keeping a persistent background process in memory.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		newDaemonInstallCmd(),
		newDaemonUninstallCmd(),
		newDaemonStatusCmd(),
		newDaemonRunCmd(reg, pm),
	)

	return cmd
}

func newDaemonInstallCmd() *cobra.Command {
	var binaryPath string
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install and start the background daemon service",
		RunE: func(cmd *cobra.Command, args []string) error {
			baseDir := config.BaseDir()
			info, err := daemon.Install(binaryPath, baseDir)
			if err != nil {
				return fmt.Errorf("failed to install daemon service: %w", err)
			}

			fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(tui.StatusGreen).Render("Background daemon service installed successfully!"))
			fmt.Printf("  Service:  %s\n", lipgloss.NewStyle().Bold(true).Foreground(tui.AccentCyan).Render(info.Label))
			fmt.Printf("  Interval: %s\n", info.Interval)
			if info.ConfigPath != "" {
				fmt.Printf("  Config:   %s\n", info.ConfigPath)
			}
			if info.LogPath != "" {
				fmt.Printf("  Log:      %s\n", info.LogPath)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&binaryPath, "binary", "", "Path to the aim executable (defaults to current binary)")
	return cmd
}

func newDaemonUninstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Stop and remove the background daemon service",
		RunE: func(cmd *cobra.Command, args []string) error {
			baseDir := config.BaseDir()
			if err := daemon.Uninstall(baseDir); err != nil {
				return fmt.Errorf("failed to uninstall daemon service: %w", err)
			}
			fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(tui.StatusGreen).Render("Background daemon service uninstalled successfully."))
			return nil
		},
	}
	return cmd
}

func newDaemonStatusCmd() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Display background daemon operational status and last run execution",
		RunE: func(cmd *cobra.Command, args []string) error {
			baseDir := config.BaseDir()
			info, err := daemon.Status(baseDir)
			if err != nil {
				return fmt.Errorf("failed to check daemon status: %w", err)
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(info)
			}

			fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(tui.AccentBlue).Render("=== AIM Background Daemon Status ==="))

			installedStr := "No"
			if info.Installed {
				installedStr = "Yes"
			}
			fmt.Printf("  Installed:    %s\n", installedStr)

			activeStr := lipgloss.NewStyle().Foreground(tui.TextMuted).Render("Inactive (Not running)")
			if info.Active {
				activeStr = lipgloss.NewStyle().Bold(true).Foreground(tui.StatusGreen).Render("Active (Running)")
			}
			fmt.Printf("  Active:       %s\n", activeStr)
			fmt.Printf("  Service:      %s\n", info.Label)
			fmt.Printf("  Interval:     %s\n", info.Interval)
			if info.ConfigPath != "" {
				fmt.Printf("  Config File:  %s\n", info.ConfigPath)
			}
			if info.LogPath != "" {
				fmt.Printf("  Log File:     %s\n", info.LogPath)
			}

			lastRunStr := "Never"
			if !info.LastRun.IsZero() {
				lastRunStr = fmt.Sprintf("%s (%s)", info.LastRun.Format(time.RFC3339), formatDaemonTimeAgo(info.LastRun))
			}
			fmt.Printf("  Last Run:     %s\n", lastRunStr)
			if info.LastRunMessage != "" {
				fmt.Printf("  Last Message: %s\n", lipgloss.NewStyle().Foreground(tui.TextMuted).Render(info.LastRunMessage))
			}

			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output status in JSON format")
	return cmd
}

func newDaemonRunCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Execute scheduled background tasks once (invoked by OS service manager)",
		RunE: func(cmd *cobra.Command, args []string) error {
			baseDir := config.BaseDir()
			return daemon.RunOnce(cmd.Context(), reg, pm, baseDir)
		},
	}
	return cmd
}

func formatDaemonTimeAgo(t time.Time) string {
	diff := time.Since(t)
	if diff < 0 {
		diff = 0
	}
	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	} else if diff < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	}
	days := int(diff.Hours() / 24)
	return fmt.Sprintf("%dd ago", days)
}
