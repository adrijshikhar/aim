package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/mcp"
	"github.com/aim-cli/aim/internal/presenter"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/tui"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
)

func newMCPCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	if reg == nil || len(reg.All()) == 0 {
		reg = defaultRegistry()
	}
	mcpCmd := &cobra.Command{
		Use:   "mcp",
		Short: "Model Context Protocol (MCP) server and integration commands",
		Long: `Model Context Protocol (MCP) integration for AI coding agents.
Provides a read-only stdio MCP server for agent self-inspection (whoami, quota, profile list).`,
	}

	mcpCmd.AddCommand(
		newMCPServeCmd(reg, pm),
		newMCPDoctorCmd(reg, pm),
		newMCPInstallCmd(reg, pm),
		newMCPUninstallCmd(reg, pm),
		newMCPListCmd(reg, pm),
		newMCPAddCmd(reg, pm),
		newMCPRemoveCmd(reg, pm),
	)

	return mcpCmd
}

type ProfileTarget struct {
	Name string
	Dir  string
}

func resolveProfileTargets(pm *profile.ProfileManager, profileFlag string, allFlag bool) ([]ProfileTarget, error) {
	if allFlag {
		if pm == nil {
			return nil, fmt.Errorf("profile manager not initialized")
		}
		profs, err := pm.ListProfiles()
		if err != nil {
			return nil, err
		}
		var targets []ProfileTarget
		for _, name := range profs {
			targets = append(targets, ProfileTarget{Name: name, Dir: pm.ProfileDir(name)})
		}
		return targets, nil
	}
	if profileFlag != "" {
		if pm == nil {
			return nil, fmt.Errorf("profile manager not initialized")
		}
		var targets []ProfileTarget
		for _, name := range strings.Split(profileFlag, ",") {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			pDir, err := pm.ResolveProfile(name)
			if err != nil {
				return nil, err
			}
			if !pm.ProfileExists(name) {
				return nil, fmt.Errorf("profile %q does not exist", name)
			}
			targets = append(targets, ProfileTarget{Name: name, Dir: pDir})
		}
		if len(targets) == 0 {
			return nil, fmt.Errorf("no valid profile names specified in --profile")
		}
		return targets, nil
	}
	return nil, nil // Nil slice explicitly signifies Host target
}

func resolveTargetAdapters(reg *agents.Registry, clientsFlag string, realHome string) ([]agents.MCPConfigManager, error) {
	if reg == nil || len(reg.All()) == 0 {
		reg = defaultRegistry()
	}
	clientsFlag = strings.TrimSpace(clientsFlag)

	allAdapters := reg.All()

	adapterByName := make(map[string]agents.MCPConfigManager)
	for _, a := range allAdapters {
		if mgr, ok := a.(agents.MCPConfigManager); ok {
			adapterByName[strings.ToLower(a.Name())] = mgr
			for _, alias := range a.Aliases() {
				adapterByName[strings.ToLower(alias)] = mgr
			}
		}
	}

	var targets []agents.MCPConfigManager
	if clientsFlag == "" || clientsFlag == "all" {
		for _, a := range allAdapters {
			if mgr, ok := a.(agents.MCPConfigManager); ok {
				if mgr.DetectInstalled(realHome) {
					targets = appendUniqueManager(targets, mgr)
				}
			}
		}
		if len(targets) == 0 {
			for _, a := range allAdapters {
				if mgr, ok := a.(agents.MCPConfigManager); ok {
					targets = appendUniqueManager(targets, mgr)
				}
			}
		}
		return targets, nil
	}

	parts := strings.Split(clientsFlag, ",")
	for _, part := range parts {
		name := strings.ToLower(strings.TrimSpace(part))
		if name == "" {
			continue
		}
		mgr, ok := adapterByName[name]
		if !ok {
			return nil, fmt.Errorf("unsupported or unconfigurable agent client %q", part)
		}
		targets = appendUniqueManager(targets, mgr)
	}
	return targets, nil
}

func appendUniqueManager(list []agents.MCPConfigManager, item agents.MCPConfigManager) []agents.MCPConfigManager {
	if item == nil {
		return list
	}
	for _, existing := range list {
		if a1, ok1 := existing.(agents.AgentAdapter); ok1 {
			if a2, ok2 := item.(agents.AgentAdapter); ok2 && a1.Name() == a2.Name() {
				return list
			}
		}
	}
	return append(list, item)
}

func resolveProfileDir(pm *profile.ProfileManager, profileFlag string) (string, error) {
	if profileFlag == "" {
		return "", nil
	}
	if pm == nil {
		return "", fmt.Errorf("profile manager not initialized")
	}
	pDir, err := pm.ResolveProfile(profileFlag)
	if err != nil {
		return "", fmt.Errorf("invalid profile %q: %w", profileFlag, err)
	}
	if !pm.ProfileExists(profileFlag) {
		return "", fmt.Errorf("profile %q does not exist", profileFlag)
	}
	return pDir, nil
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
				if _, err := resolveProfileDir(pm, profileFlag); err != nil {
					return err
				}
				_ = os.Setenv("AIM_PROFILE", profileFlag)
			}
			if agentFlag != "" {
				_ = os.Setenv("AIM_AGENT", agentFlag)
			}

			baseDir := config.BaseDir()
			srv := mcp.NewServer(reg, pm, baseDir)
			srv.SetSessionManager(getSessionManager(cmd.Context()))
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
2. Client audit: Checks detected AI coding agents (Claude Code, Codex, Antigravity) for valid configuration and 'aim' registration.
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

func newMCPInstallCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	var profileFlag string

	installCmd := &cobra.Command{
		Use:   "install [client]",
		Short: "Register aim MCP server into agent configuration (e.g. claude, codex, agy)",
		Long: `Registers 'aim mcp serve' as an MCP server in target agent configuration files.
Supports 'claude' (Claude Code), 'codex', 'agy', or 'all'.
When run without arguments, scans and auto-detects installed agents.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			homeDir := config.RealHomeDir()
			profileDir, err := resolveProfileDir(pm, profileFlag)
			if err != nil {
				return err
			}

			targetClient := "all"
			if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
				targetClient = strings.ToLower(strings.TrimSpace(args[0]))
			}

			targets, err := resolveTargetAdapters(reg, targetClient, homeDir)
			if err != nil {
				if targetClient != "all" {
					fmt.Fprintf(cmd.OutOrStdout(), "Auto-install is currently supported for 'claude', 'codex', 'agy'.\nFor %s, register:\n  Command: aim\n  Args: [\"mcp\", \"serve\"]\n", targetClient)
					return nil
				}
				return err
			}
			if len(targets) == 0 {
				return fmt.Errorf("no target agent clients found to register server into")
			}

			aimEntry := agents.MCPServerConfig{
				Command: "aim",
				Args:    []string{"mcp", "serve"},
			}

			ctx := cmd.Context()
			out := cmd.OutOrStdout()

			for _, adapter := range targets {
				adName := "agent"
				if a, ok := adapter.(agents.AgentAdapter); ok {
					adName = a.DisplayName()
				}

				if profileDir == "" {
					if err := adapter.AddHostMCPServer(ctx, "aim", aimEntry); err != nil {
						return fmt.Errorf("failed to register into %s (host): %w", adName, err)
					}
					fmt.Fprintf(out, "Installed aim MCP server into %s at %s\n", adName, homeDir)
				} else {
					if err := adapter.AddProfileMCPServer(ctx, profileFlag, profileDir, "aim", aimEntry); err != nil {
						return fmt.Errorf("failed to register into %s (profile %s): %w", adName, profileFlag, err)
					}
					fmt.Fprintf(out, "Installed aim MCP server into %s at %s\n", adName, profileDir)
				}
			}

			return nil
		},
	}

	installCmd.Flags().StringVar(&profileFlag, "profile", "", "Target specific profile instead of global host config")

	return installCmd
}

func newMCPUninstallCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	var profileFlag string

	uninstallCmd := &cobra.Command{
		Use:   "uninstall [client]",
		Short: "Remove aim MCP server from agent configuration (e.g. claude, codex, agy)",
		Long: `Removes the 'aim' MCP server definition from agent configuration files without disturbing other servers.
Supports 'claude', 'codex', 'agy', or 'all'.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			homeDir := config.RealHomeDir()
			profileDir, err := resolveProfileDir(pm, profileFlag)
			if err != nil {
				return err
			}

			targetClient := "all"
			if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
				targetClient = strings.ToLower(strings.TrimSpace(args[0]))
			}

			targets, err := resolveTargetAdapters(reg, targetClient, homeDir)
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			out := cmd.OutOrStdout()
			removedCount := 0

			for _, adapter := range targets {
				adName := "agent"
				if a, ok := adapter.(agents.AgentAdapter); ok {
					adName = a.DisplayName()
				}

				if profileDir == "" {
					if err := adapter.RemoveHostMCPServer(ctx, "aim"); err == nil {
						fmt.Fprintf(out, "Removed aim MCP server from %s at %s\n", adName, homeDir)
						removedCount++
					} else if !errors.Is(err, os.ErrNotExist) {
						return fmt.Errorf("failed to remove aim from %s: %w", adName, err)
					}
				} else {
					if err := adapter.RemoveProfileMCPServer(ctx, profileFlag, profileDir, "aim"); err == nil {
						fmt.Fprintf(out, "Removed aim MCP server from %s at %s\n", adName, profileDir)
						removedCount++
					} else if !errors.Is(err, os.ErrNotExist) {
						return fmt.Errorf("failed to remove aim from %s: %w", adName, err)
					}
				}
			}

			if removedCount == 0 {
				fmt.Fprintln(out, "No active 'aim' MCP server registrations found.")
			}
			return nil
		},
	}

	uninstallCmd.Flags().StringVar(&profileFlag, "profile", "", "Target specific profile instead of global host config")
	return uninstallCmd
}

func newMCPListCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	var profileFlag string
	var clientFlag string
	var clientsFlag string
	var jsonFlag bool

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List configured MCP servers across agent configurations",
		Long:  `Displays all Model Context Protocol (MCP) servers configured across detected AI coding agents and profiles.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			homeDir := config.RealHomeDir()
			profileDir, err := resolveProfileDir(pm, profileFlag)
			if err != nil {
				return err
			}

			type serverRow struct {
				Name    string `json:"name"`
				Client  string `json:"client"`
				Command string `json:"command"`
				Args    string `json:"args,omitempty"`
				Status  string `json:"status"`
			}

			var rows []serverRow
			ctx := cmd.Context()

			selectedClient := clientFlag
			if selectedClient == "" && clientsFlag != "" {
				selectedClient = clientsFlag
			}

			if reg == nil || len(reg.All()) == 0 {
				reg = defaultRegistry()
			}

			var targets []agents.MCPConfigManager
			if selectedClient == "" || selectedClient == "all" {
				for _, a := range reg.All() {
					if mgr, ok := a.(agents.MCPConfigManager); ok {
						targets = appendUniqueManager(targets, mgr)
					}
				}
			} else {
				var err error
				targets, err = resolveTargetAdapters(reg, selectedClient, homeDir)
				if err != nil {
					return err
				}
			}

			for _, adapter := range targets {
				adName := "Agent"
				if a, ok := adapter.(agents.AgentAdapter); ok {
					adName = a.DisplayName()
				}
				scopeLabel := adName
				if profileFlag != "" {
					scopeLabel += " (" + profileFlag + ")"
				}

				servers, err := adapter.ListMCPServers(ctx, profileFlag, profileDir)
				if err != nil || len(servers) == 0 {
					continue
				}

				for _, s := range servers {
					status := s.Status
					if status == "" {
						status = "enabled"
					}
					rows = append(rows, serverRow{
						Name:    s.Name,
						Client:  scopeLabel,
						Command: s.Target,
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
				tableRows = append(tableRows, []string{
					r.Name,
					r.Client,
					r.Command,
					r.Status,
				})
			}

			t := presenter.NewTable(out).
				Border(lipgloss.RoundedBorder()).
				BorderStyle(lipgloss.NewStyle().Foreground(tui.TextDim)).
				Headers("SERVER", "CLIENT / SCOPE", "COMMAND / TARGET", "STATUS").
				Rows(tableRows...)

			t.StyleFunc(func(row, col int) lipgloss.Style {
				if row == table.HeaderRow {
					return lipgloss.NewStyle().Bold(true).Foreground(tui.AccentBlue)
				}
				switch col {
				case 0:
					return lipgloss.NewStyle().Bold(true).Foreground(tui.TextBright)
				case 1:
					return lipgloss.NewStyle().Foreground(tui.AccentCyan)
				case 2:
					return lipgloss.NewStyle().Foreground(tui.TextPrimary)
				case 3:
					if row >= 0 && row < len(rows) && rows[row].Status == "enabled" {
						return lipgloss.NewStyle().Foreground(tui.StatusGreen)
					}
					return lipgloss.NewStyle().Foreground(tui.TextMuted)
				default:
					return lipgloss.NewStyle().Foreground(tui.TextSecondary)
				}
			})

			fmt.Fprintln(out, t.Render())
			return nil
		},
	}

	listCmd.Flags().StringVar(&profileFlag, "profile", "", "Filter or inspect specific profile")
	listCmd.Flags().StringVar(&clientFlag, "client", "", "Filter by client (e.g. 'claude', 'codex')")
	listCmd.Flags().StringVar(&clientsFlag, "clients", "", "Filter by clients (e.g. 'claude', 'codex')")
	listCmd.Flags().BoolVar(&jsonFlag, "json", false, "Output results in JSON format")

	return listCmd
}

func newMCPAddCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	var clientsFlag string
	var profileFlag string
	var allFlag bool

	addCmd := &cobra.Command{
		Use:   "add <name> <command> [args...]",
		Short: "Register an MCP server across agent configurations",
		Long: `Registers a new MCP server definition across selected or detected agent configurations.
Example:
  aim mcp add fetch uvx mcp-server-fetch --clients codex,claude
  aim mcp add fetch uvx mcp-server-fetch --profile work,staging
  aim mcp add fetch uvx mcp-server-fetch --all`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			homeDir := config.RealHomeDir()
			profileTargets, err := resolveProfileTargets(pm, profileFlag, allFlag)
			if err != nil {
				return err
			}

			serverName := args[0]
			serverCommand := args[1]
			var serverArgs []string
			if len(args) > 2 {
				serverArgs = args[2:]
			}

			targets, err := resolveTargetAdapters(reg, clientsFlag, homeDir)
			if err != nil {
				return err
			}
			if len(targets) == 0 {
				return fmt.Errorf("no target agent clients found to register server into")
			}

			cfg := agents.MCPServerConfig{}
			if strings.HasPrefix(serverCommand, "http://") || strings.HasPrefix(serverCommand, "https://") {
				cfg.URL = serverCommand
			} else {
				cfg.Command = serverCommand
				cfg.Args = serverArgs
			}

			ctx := cmd.Context()
			out := cmd.OutOrStdout()

			if allFlag && len(profileTargets) == 0 {
				fmt.Fprintln(out, "No profiles configured.")
				return nil
			}

			for _, adapter := range targets {
				adName := "agent"
				if a, ok := adapter.(agents.AgentAdapter); ok {
					adName = a.DisplayName()
				}

				if !allFlag && profileFlag == "" {
					if err := adapter.AddHostMCPServer(ctx, serverName, cfg); err != nil {
						return fmt.Errorf("failed to add server to %s (host): %w", adName, err)
					}
					fmt.Fprintf(out, "Added MCP server '%s' to %s at %s\n", serverName, adName, homeDir)
				} else {
					for _, p := range profileTargets {
						if err := adapter.AddProfileMCPServer(ctx, p.Name, p.Dir, serverName, cfg); err != nil {
							return fmt.Errorf("failed to add server to %s (profile %s): %w", adName, p.Name, err)
						}
						fmt.Fprintf(out, "Added MCP server '%s' to %s (profile %s)\n", serverName, adName, p.Name)
					}
				}
			}
			return nil
		},
	}

	addCmd.Flags().StringVar(&clientsFlag, "clients", "all", "Comma-separated list of target clients (e.g. 'claude,codex') or 'all'")
	addCmd.Flags().StringVar(&profileFlag, "profile", "", "Comma-separated list of target profiles (e.g. 'work,dev')")
	addCmd.Flags().BoolVar(&allFlag, "all", false, "Apply to all configured profiles")

	return addCmd
}

func newMCPRemoveCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	var clientsFlag string
	var profileFlag string
	var allFlag bool

	removeCmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Unregister an MCP server from agent configurations",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			homeDir := config.RealHomeDir()
			profileTargets, err := resolveProfileTargets(pm, profileFlag, allFlag)
			if err != nil {
				return err
			}

			serverName := args[0]
			targets, err := resolveTargetAdapters(reg, clientsFlag, homeDir)
			if err != nil {
				return err
			}
			if len(targets) == 0 {
				return fmt.Errorf("no target agent clients found")
			}

			ctx := cmd.Context()
			out := cmd.OutOrStdout()

			if allFlag && len(profileTargets) == 0 {
				fmt.Fprintln(out, "No profiles configured.")
				return nil
			}

			removedCount := 0

			for _, adapter := range targets {
				adName := "agent"
				if a, ok := adapter.(agents.AgentAdapter); ok {
					adName = a.DisplayName()
				}

				if !allFlag && profileFlag == "" {
					if err := adapter.RemoveHostMCPServer(ctx, serverName); err == nil {
						fmt.Fprintf(out, "Removed MCP server '%s' from %s at %s\n", serverName, adName, homeDir)
						removedCount++
					} else if !errors.Is(err, os.ErrNotExist) {
						return fmt.Errorf("failed to remove MCP server %q from %s: %w", serverName, adName, err)
					}
				} else {
					for _, p := range profileTargets {
						if err := adapter.RemoveProfileMCPServer(ctx, p.Name, p.Dir, serverName); err == nil {
							fmt.Fprintf(out, "Removed MCP server '%s' from %s (profile %s)\n", serverName, adName, p.Name)
							removedCount++
						} else if !errors.Is(err, os.ErrNotExist) {
							return fmt.Errorf("failed to remove MCP server %q from %s (profile %s): %w", serverName, adName, p.Name, err)
						}
					}
				}
			}

			if removedCount == 0 {
				fmt.Fprintf(out, "Server '%s' was not found in selected agent configs.\n", serverName)
			}
			return nil
		},
	}

	removeCmd.Flags().StringVar(&clientsFlag, "clients", "all", "Comma-separated list of target clients or 'all'")
	removeCmd.Flags().StringVar(&profileFlag, "profile", "", "Comma-separated list of target profiles")
	removeCmd.Flags().BoolVar(&allFlag, "all", false, "Apply to all configured profiles")

	return removeCmd
}
