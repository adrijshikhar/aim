package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/mcp"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
)

func newMCPCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	mcpCmd := &cobra.Command{
		Use:   "mcp",
		Short: "Model Context Protocol (MCP) server and integration commands",
		Long: `Model Context Protocol (MCP) integration for AI coding agents.
Provides a read-only stdio MCP server for agent self-inspection (whoami, quota, profile list).`,
	}

	mcpCmd.AddCommand(
		newMCPServeCmd(reg, pm),
		newMCPDoctorCmd(reg, pm),
		newMCPInstallCmd(pm),
		newMCPUninstallCmd(pm),
		newMCPListCmd(pm),
		newMCPAddCmd(pm),
		newMCPRemoveCmd(pm),
	)

	return mcpCmd
}

func newMCPServeCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	var profileFlag string
	var agentFlag string

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Run read-only stdio MCP server for agent self-inspection",
		Long: `Starts a JSON-RPC 2.0 Model Context Protocol server over stdio.
Exposes read-only tools: aim_whoami, aim_get_quota, and aim_list_profiles.
Guarantees strict read-only operation: no profile switching, creation, or deletion capabilities exist.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if profileFlag != "" {
				if pm != nil {
					if _, err := pm.ResolveProfile(profileFlag); err != nil {
						return fmt.Errorf("invalid profile %q: %w", profileFlag, err)
					}
					if !pm.ProfileExists(profileFlag) {
						return fmt.Errorf("profile %q does not exist", profileFlag)
					}
				}
				_ = os.Setenv("AIM_PROFILE", profileFlag)
			}
			if agentFlag != "" {
				_ = os.Setenv("AIM_AGENT", agentFlag)
			}

			baseDir := config.BaseDir()
			srv := mcp.NewServer(reg, pm, baseDir)
			srv.SetSessionManager(NewDefaultSessionManager())
			return srv.Serve(cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}

	serveCmd.Flags().StringVar(&profileFlag, "profile", "", "Override active profile name")
	serveCmd.Flags().StringVar(&agentFlag, "agent", "", "Override active agent name")

	return serveCmd
}

func newMCPDoctorCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	var check bool
	doctorCmd := &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose MCP protocol conformance, client configs, and binaries",
		Long: `Performs a multi-stage diagnostic check of the AIM MCP ecosystem:
1. Self-test: Verifies in-process stdio JSON-RPC 2.0 protocol handshake and tool definitions.
2. Client audit: Checks detected AI coding clients (Claude Code, Cursor, Claude Desktop, Windsurf) for valid configuration and 'aim' registration.
3. Binary audit: Verifies that third-party MCP server executables exist in PATH.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			homeDir := config.RealHomeDir()
			report := mcp.RunMCPDoctor(cmd.Context(), reg, pm, homeDir)

			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "AIM Model Context Protocol (MCP) Doctor")
			fmt.Fprintln(out, "=======================================")

			const (
				mcpTagOK   = "[OK]    "
				mcpTagWARN = "[WARN]  "
				mcpTagFAIL = "[FAIL]  "
			)

			for _, res := range report.Results {
				tag := mcpTagOK
				switch res.Status {
				case mcp.StatusWarn:
					tag = mcpTagWARN
				case mcp.StatusFail:
					tag = mcpTagFAIL
				}
				fmt.Fprintf(out, "%s %s: %s\n", tag, res.Category, res.Message)
				if res.Remedy != "" {
					fmt.Fprintf(out, "       Remedy: %s\n", res.Remedy)
				}
			}

			fmt.Fprintf(out, "\nAll MCP checks completed in %dms (Status: %s)\n", report.DurationMs, report.OverallStatus)

			if check && report.OverallStatus == mcp.StatusFail {
				return fmt.Errorf("mcp doctor diagnostics reported failures")
			}
			return nil
		},
	}

	doctorCmd.Flags().BoolVar(&check, "check", false, "Exit with non-zero status if any check fails")
	return doctorCmd
}

func newMCPInstallCmd(pm *profile.ProfileManager) *cobra.Command {
	var profileFlag string
	var yesFlag bool

	installCmd := &cobra.Command{
		Use:   "install [client]",
		Short: "Register aim MCP server into agent/IDE configuration (e.g. cursor, claude)",
		Long: `Registers 'aim mcp serve' as an MCP server in target client configuration files.
Supports 'claude' (Claude Code), 'claude-desktop', 'cursor', 'windsurf', or 'all'.
When run without arguments, scans and auto-detects installed clients.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			homeDir := config.RealHomeDir()
			var profileDir string
			if profileFlag != "" {
				if pm == nil {
					return fmt.Errorf("profile manager not initialized")
				}
				pDir, err := pm.ResolveProfile(profileFlag)
				if err != nil {
					return fmt.Errorf("invalid profile %q: %w", profileFlag, err)
				}
				if !pm.ProfileExists(profileFlag) {
					return fmt.Errorf("profile %q does not exist", profileFlag)
				}
				profileDir = pDir
			}

			targetClient := ""
			if len(args) > 0 {
				targetClient = strings.ToLower(strings.TrimSpace(args[0]))
			}

			var targets []mcp.ClientConfigDriver
			if targetClient != "" && targetClient != "all" {
				driver, ok := mcp.GetDriver(mcp.ClientID(targetClient))
				if !ok {
					fmt.Fprintf(cmd.OutOrStdout(), "Auto-install is currently supported for 'claude', 'claude-desktop', 'cursor', 'windsurf'.\nFor %s, register:\n  Command: aim\n  Args: [\"mcp\", \"serve\"]\n", targetClient)
					return nil
				}
				targets = append(targets, driver)
			} else {
				detected := mcp.DetectInstalledClients(homeDir)
				if len(detected) == 0 {
					// Fall back to claude driver if none detected
					if d, ok := mcp.GetDriver(mcp.ClientClaudeCode); ok {
						detected = append(detected, d)
					}
				}
				targets = detected
			}

			aimEntry := mcp.RawMCPEntry{
				Command: "aim",
				Args:    []string{"mcp", "serve"},
			}

			for _, driver := range targets {
				cfgPath, err := driver.ConfigPath(homeDir, profileDir)
				if err != nil {
					return err
				}
				if err := driver.WriteServer(cfgPath, "aim", aimEntry); err != nil {
					return fmt.Errorf("failed to register into %s at %s: %w", driver.Name(), cfgPath, err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Installed aim MCP server into %s at %s\n", driver.Name(), cfgPath)
			}

			return nil
		},
	}

	installCmd.Flags().StringVar(&profileFlag, "profile", "", "Target specific profile instead of global host config")
	installCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Confirm installation to all detected clients without prompt")

	return installCmd
}

func newMCPUninstallCmd(pm *profile.ProfileManager) *cobra.Command {
	var profileFlag string

	uninstallCmd := &cobra.Command{
		Use:   "uninstall [client]",
		Short: "Remove aim MCP server from client configuration (e.g. cursor, claude)",
		Long: `Removes the 'aim' MCP server definition from client configuration files without disturbing other servers.
Supports 'claude', 'claude-desktop', 'cursor', 'windsurf', or 'all'.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			homeDir := config.RealHomeDir()
			var profileDir string
			if profileFlag != "" && pm != nil {
				pDir, err := pm.ResolveProfile(profileFlag)
				if err == nil {
					profileDir = pDir
				}
			}

			targetClient := "all"
			if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
				targetClient = strings.ToLower(strings.TrimSpace(args[0]))
			}

			var targets []mcp.ClientConfigDriver
			if targetClient != "all" {
				d, ok := mcp.GetDriver(mcp.ClientID(targetClient))
				if !ok {
					return fmt.Errorf("unsupported client %q", targetClient)
				}
				targets = append(targets, d)
			} else {
				targets = mcp.AllDrivers()
			}

			removedCount := 0
			for _, driver := range targets {
				cfgPath, err := driver.ConfigPath(homeDir, profileDir)
				if err != nil {
					continue
				}
				servers, err := driver.ReadServers(cfgPath)
				if err != nil || servers == nil {
					continue
				}
				if _, exists := servers["aim"]; exists {
					if err := driver.DeleteServer(cfgPath, "aim"); err == nil {
						fmt.Fprintf(cmd.OutOrStdout(), "Removed aim MCP server from %s at %s\n", driver.Name(), cfgPath)
						removedCount++
					}
				}
			}

			if removedCount == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No active 'aim' MCP server registrations found.")
			}
			return nil
		},
	}

	uninstallCmd.Flags().StringVar(&profileFlag, "profile", "", "Target specific profile instead of global host config")
	return uninstallCmd
}

func newMCPListCmd(pm *profile.ProfileManager) *cobra.Command {
	var profileFlag string
	var clientFlag string
	var jsonFlag bool

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List configured MCP servers across client configurations",
		Long:  `Displays all Model Context Protocol (MCP) servers configured across detected AI coding clients and profiles.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			homeDir := config.RealHomeDir()
			var profileDir string
			if profileFlag != "" && pm != nil {
				pDir, err := pm.ResolveProfile(profileFlag)
				if err == nil {
					profileDir = pDir
				}
			}

			type serverRow struct {
				Name    string `json:"name"`
				Client  string `json:"client"`
				Command string `json:"command"`
				Args    string `json:"args,omitempty"`
				Status  string `json:"status"`
			}

			var rows []serverRow

			drivers := mcp.AllDrivers()
			if clientFlag != "" {
				d, ok := mcp.GetDriver(mcp.ClientID(strings.ToLower(clientFlag)))
				if ok {
					drivers = []mcp.ClientConfigDriver{d}
				}
			}

			for _, driver := range drivers {
				cfgPath, err := driver.ConfigPath(homeDir, profileDir)
				if err != nil {
					continue
				}
				servers, err := driver.ReadServers(cfgPath)
				if err != nil || len(servers) == 0 {
					continue
				}

				scopeLabel := driver.Name()
				if profileFlag != "" {
					scopeLabel += " (" + profileFlag + ")"
				}

				for sName, entry := range servers {
					target := entry.Command
					if target == "" {
						target = entry.ServerURL
						if target == "" {
							target = entry.URL
						}
					}
					status := "enabled"
					if entry.Disabled {
						status = "disabled"
					}
					rows = append(rows, serverRow{
						Name:    sName,
						Client:  scopeLabel,
						Command: target,
						Args:    strings.Join(entry.Args, " "),
						Status:  status,
					})
				}
			}

			sort.Slice(rows, func(i, j int) bool {
				if rows[i].Name == rows[j].Name {
					return rows[i].Client < rows[j].Client
				}
				return rows[i].Name < rows[j].Name
			})

			out := cmd.OutOrStdout()
			if jsonFlag {
				data, err := json.MarshalIndent(rows, "", "  ")
				if err != nil {
					return err
				}
				fmt.Fprintln(out, string(data))
				return nil
			}

			if len(rows) == 0 {
				fmt.Fprintln(out, "No MCP servers configured.")
				return nil
			}

			var tableRows [][]string
			for _, r := range rows {
				cmdDisplay := r.Command
				if r.Args != "" {
					cmdDisplay += " " + r.Args
				}
				tableRows = append(tableRows, []string{
					r.Name,
					r.Client,
					cmdDisplay,
					r.Status,
				})
			}

			t := table.New().
				Border(lipgloss.RoundedBorder()).
				BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
				Headers("SERVER", "CLIENT / SCOPE", "COMMAND / TARGET", "STATUS").
				Rows(tableRows...)

			fmt.Fprintln(out, t.Render())
			return nil
		},
	}

	listCmd.Flags().StringVar(&profileFlag, "profile", "", "Filter or inspect specific profile")
	listCmd.Flags().StringVar(&clientFlag, "client", "", "Filter by client (e.g. 'cursor', 'claude')")
	listCmd.Flags().BoolVar(&jsonFlag, "json", false, "Output results in JSON format")

	return listCmd
}

func newMCPAddCmd(pm *profile.ProfileManager) *cobra.Command {
	var clientsFlag string
	var profileFlag string

	addCmd := &cobra.Command{
		Use:   "add <name> <command> [args...]",
		Short: "Register a third-party MCP server across client configurations",
		Long: `Registers a new MCP server definition across selected or all detected AI client configurations.
Example:
  aim mcp add fetch uvx mcp-server-fetch --clients cursor,claude`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			homeDir := config.RealHomeDir()
			var profileDir string
			if profileFlag != "" && pm != nil {
				pDir, err := pm.ResolveProfile(profileFlag)
				if err == nil {
					profileDir = pDir
				}
			}

			serverName := args[0]
			serverCommand := args[1]
			var serverArgs []string
			if len(args) > 2 {
				serverArgs = args[2:]
			}

			targets := resolveTargetDrivers(clientsFlag, homeDir)
			if len(targets) == 0 {
				return fmt.Errorf("no target clients found to register server into")
			}

			entry := mcp.RawMCPEntry{
				Command: serverCommand,
				Args:    serverArgs,
			}

			for _, driver := range targets {
				cfgPath, err := driver.ConfigPath(homeDir, profileDir)
				if err != nil {
					return err
				}
				if err := driver.WriteServer(cfgPath, serverName, entry); err != nil {
					return fmt.Errorf("failed to add server to %s: %w", driver.Name(), err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Added MCP server '%s' to %s at %s\n", serverName, driver.Name(), cfgPath)
			}
			return nil
		},
	}

	addCmd.Flags().StringVar(&clientsFlag, "clients", "all", "Comma-separated list of target clients (e.g. 'cursor,claude') or 'all'")
	addCmd.Flags().StringVar(&profileFlag, "profile", "", "Target specific profile instead of global host config")

	return addCmd
}

func newMCPRemoveCmd(pm *profile.ProfileManager) *cobra.Command {
	var clientsFlag string
	var profileFlag string

	removeCmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Unregister an MCP server from client configurations",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			homeDir := config.RealHomeDir()
			var profileDir string
			if profileFlag != "" && pm != nil {
				pDir, err := pm.ResolveProfile(profileFlag)
				if err == nil {
					profileDir = pDir
				}
			}

			serverName := args[0]
			targets := resolveTargetDrivers(clientsFlag, homeDir)
			if len(targets) == 0 {
				return fmt.Errorf("no target clients found")
			}

			removedCount := 0
			for _, driver := range targets {
				cfgPath, err := driver.ConfigPath(homeDir, profileDir)
				if err != nil {
					continue
				}
				if err := driver.DeleteServer(cfgPath, serverName); err == nil {
					fmt.Fprintf(cmd.OutOrStdout(), "Removed MCP server '%s' from %s at %s\n", serverName, driver.Name(), cfgPath)
					removedCount++
				}
			}

			if removedCount == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "Server '%s' was not found in selected client configs.\n", serverName)
			}
			return nil
		},
	}

	removeCmd.Flags().StringVar(&clientsFlag, "clients", "all", "Comma-separated list of target clients or 'all'")
	removeCmd.Flags().StringVar(&profileFlag, "profile", "", "Target specific profile instead of global host config")

	return removeCmd
}

func resolveTargetDrivers(clientsFlag string, homeDir string) []mcp.ClientConfigDriver {
	clientsFlag = strings.TrimSpace(clientsFlag)
	if clientsFlag == "" || clientsFlag == "all" {
		detected := mcp.DetectInstalledClients(homeDir)
		if len(detected) > 0 {
			return detected
		}
		return mcp.AllDrivers()
	}

	var targets []mcp.ClientConfigDriver
	parts := strings.Split(clientsFlag, ",")
	for _, part := range parts {
		cID := mcp.ClientID(strings.ToLower(strings.TrimSpace(part)))
		if d, ok := mcp.GetDriver(cID); ok {
			targets = append(targets, d)
		}
	}
	return targets
}
