package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/mcp"
	"github.com/aim-cli/aim/internal/profile"
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
		newMCPInstallCmd(pm),
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
		RunE: func(cmd *cobra.Command, args []string) error {
			if profileFlag != "" {
				_ = os.Setenv("AIM_PROFILE", profileFlag)
			}
			if agentFlag != "" {
				_ = os.Setenv("AIM_AGENT", agentFlag)
			}

			baseDir := config.BaseDir()
			srv := mcp.NewServer(reg, pm, baseDir)
			return srv.Serve(cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}

	serveCmd.Flags().StringVar(&profileFlag, "profile", "", "Override active profile name")
	serveCmd.Flags().StringVar(&agentFlag, "agent", "", "Override active agent name")

	return serveCmd
}

func newMCPInstallCmd(pm *profile.ProfileManager) *cobra.Command {
	var profileFlag string

	installCmd := &cobra.Command{
		Use:   "install [agent]",
		Short: "Register aim MCP server into agent configuration (e.g. claude)",
		Long: `Registers 'aim mcp serve' as an MCP server in the target agent configuration file.
By default installs for 'claude' into ~/.claude.json, or into the specified profile directory with --profile.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetAgent := "claude"
			if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
				targetAgent = strings.ToLower(strings.TrimSpace(args[0]))
			}

			if targetAgent != "claude" {
				fmt.Fprintf(cmd.OutOrStdout(), "Auto-install is currently supported for 'claude'.\nFor %s, register:\n  Command: aim\n  Args: [\"mcp\", \"serve\"]\n", targetAgent)
				return nil
			}

			var targetPath string
			if profileFlag != "" {
				if pm == nil {
					return fmt.Errorf("profile manager not initialized")
				}
				pDir := pm.ProfileDir(profileFlag)
				targetDir := filepath.Join(pDir, ".claude")
				if err := os.MkdirAll(targetDir, 0700); err != nil {
					return fmt.Errorf("failed to create profile directory: %w", err)
				}
				targetPath = filepath.Join(targetDir, ".claude.json")
			} else {
				homeDir := config.RealHomeDir()
				targetPath = filepath.Join(homeDir, ".claude.json")
			}

			if err := installClaudeMCPEntry(targetPath); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Installed aim MCP server into %s\n", targetPath)
			return nil
		},
	}

	installCmd.Flags().StringVar(&profileFlag, "profile", "", "Target specific profile instead of global host config")

	return installCmd
}

func installClaudeMCPEntry(path string) error {
	type rawEntry struct {
		Command string   `json:"command"`
		Args    []string `json:"args"`
	}

	var fullDoc map[string]any
	if data, err := os.ReadFile(path); err == nil && len(data) > 0 {
		_ = json.Unmarshal(data, &fullDoc)
	}
	if fullDoc == nil {
		fullDoc = make(map[string]any)
	}

	serversMap := make(map[string]any)
	if existingServers, ok := fullDoc["mcpServers"].(map[string]any); ok {
		for k, v := range existingServers {
			serversMap[k] = v
		}
	}

	serversMap["aim"] = rawEntry{
		Command: "aim",
		Args:    []string{"mcp", "serve"},
	}
	fullDoc["mcpServers"] = serversMap

	out, err := json.MarshalIndent(fullDoc, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal claude config: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	tmpFile := filepath.Join(dir, fmt.Sprintf(".claude.json.tmp.%d", os.Getpid()))
	if err := os.WriteFile(tmpFile, out, 0600); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	if err := os.Rename(tmpFile, path); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}
