package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/tui"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
)

func newMCPCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:     "mcp [agent] [profile]",
		Short:   "Inspect configured MCP servers for an agent and profile",
		Long:    "Inspect configured MCP servers for an agent and profile with a readable formatted table or JSON output.",
		Aliases: []string{"mcps"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMCPList(cmd.OutOrStdout(), reg, pm, args, jsonOutput)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completeAgentAndProfile(reg, pm, args, toComplete)
		},
	}

	listCmd := &cobra.Command{
		Use:   "list [agent] [profile]",
		Short: "List configured MCP servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMCPList(cmd.OutOrStdout(), reg, pm, args, jsonOutput)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completeAgentAndProfile(reg, pm, args, toComplete)
		},
	}

	cmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output MCP servers as JSON")
	cmd.AddCommand(listCmd)

	return cmd
}

func resolveAgentAndProfile(reg *agents.Registry, pm *profile.ProfileManager, args []string) (string, string, error) {
	cfg, _ := config.LoadConfig()

	agentName := ""
	profileName := ""

	if len(args) == 0 {
		if cfg != nil && cfg.DefaultProfile != "" {
			profileName = cfg.DefaultProfile
		} else if pm != nil {
			profiles, _ := pm.ListProfiles()
			if len(profiles) > 0 {
				profileName = profiles[0]
			}
		}
		if profileName == "" {
			profileName = "default"
		}
		if cfg != nil {
			if profCfg, ok := cfg.Profiles[profileName]; ok && len(profCfg.Agents) > 0 {
				agentName = profCfg.Agents[0]
			}
		}
		if agentName == "" {
			agentName = "codex"
		}
		return agentName, profileName, nil
	}

	if len(args) == 1 {
		arg := args[0]
		if reg != nil {
			if _, err := reg.Get(arg); err == nil {
				agentName = arg
				if cfg != nil && cfg.DefaultProfile != "" {
					profileName = cfg.DefaultProfile
				} else {
					profileName = "default"
				}
				return agentName, profileName, nil
			}
		}
		profileName = arg
		if cfg != nil {
			if profCfg, ok := cfg.Profiles[profileName]; ok && len(profCfg.Agents) > 0 {
				agentName = profCfg.Agents[0]
			}
		}
		if agentName == "" {
			agentName = "codex"
		}
		return agentName, profileName, nil
	}

	arg0, arg1 := args[0], args[1]
	if reg != nil {
		if _, err := reg.Get(arg0); err == nil {
			agentName = arg0
			profileName = arg1
			return agentName, profileName, nil
		}
		if _, err := reg.Get(arg1); err == nil {
			agentName = arg1
			profileName = arg0
			return agentName, profileName, nil
		}
	}

	agentName = arg0
	profileName = arg1
	return agentName, profileName, nil
}

func runMCPList(w io.Writer, reg *agents.Registry, pm *profile.ProfileManager, args []string, jsonOutput bool) error {
	agentName, profileName, err := resolveAgentAndProfile(reg, pm, args)
	if err != nil {
		return err
	}

	adapter, err := reg.Get(agentName)
	if err != nil {
		return fmt.Errorf("unknown agent %q: %w", agentName, err)
	}

	listProv, ok := adapter.(agents.MCPListProvider)
	if !ok {
		return fmt.Errorf("agent %q does not support listing MCP servers", agentName)
	}

	pDir, err := pm.EnsureProfile(profileName)
	if err != nil {
		return fmt.Errorf("failed to prepare profile %q: %w", profileName, err)
	}

	cfg, _ := config.LoadConfig()

	var servers []agents.MCPServerInfo
	var fetchErr error

	if _, isMCPProv := adapter.(agents.MCPProvider); isMCPProv {
		withSessionMerge(adapter, pm.MergeStateStore(), profileName, pDir, cfg, []string{"mcp", "list"}, func() int {
			servers, fetchErr = listProv.ListMCPServers(context.Background(), profileName, pDir)
			return 0
		})
	} else {
		servers, fetchErr = listProv.ListMCPServers(context.Background(), profileName, pDir)
	}

	if fetchErr != nil {
		return fmt.Errorf("failed to list MCP servers: %w", fetchErr)
	}

	if jsonOutput {
		if servers == nil {
			servers = []agents.MCPServerInfo{}
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(servers)
	}

	renderMCPListTable(w, agentName, profileName, servers)
	return nil
}

func renderMCPListTable(w io.Writer, agentName, profileName string, servers []agents.MCPServerInfo) {
	banner := fmt.Sprintf("=== Configured MCP Servers (%s: %s) ===", agentName, profileName)
	fmt.Fprintf(w, "\n%s\n\n", lipgloss.NewStyle().Bold(true).Foreground(tui.AccentBlue).Render(banner))

	if len(servers) == 0 {
		fmt.Fprintf(w, "  %s\n\n", lipgloss.NewStyle().Foreground(tui.TextMuted).Render("(no MCP servers configured)"))
		return
	}

	t := table.New().
		Border(lipgloss.HiddenBorder()).
		Headers("#", "NAME", "STATUS", "AUTH", "TYPE", "TARGET / COMMAND")

	enabledCount := 0
	disabledCount := 0

	for i, s := range servers {
		numStr := fmt.Sprintf("%d", i+1)

		statusStyled := lipgloss.NewStyle().Foreground(tui.StatusGreen).Render("enabled")
		if s.Status == "disabled" {
			statusStyled = lipgloss.NewStyle().Foreground(tui.StatusRed).Render("disabled")
			disabledCount++
		} else {
			enabledCount++
		}

		authStyled := lipgloss.NewStyle().Foreground(tui.TextMuted).Render(s.Auth)
		if s.Auth == "OAuth" || s.Auth == "connected" {
			authStyled = lipgloss.NewStyle().Foreground(tui.StatusGreen).Render(s.Auth)
		} else if s.Auth == "auth required" {
			authStyled = lipgloss.NewStyle().Foreground(tui.StatusYellow).Render(s.Auth)
		}

		typeStyled := lipgloss.NewStyle().Foreground(tui.AccentCyan).Render(s.Type)

		t.Row(
			numStr,
			s.Name,
			statusStyled,
			authStyled,
			typeStyled,
			s.Target,
		)
	}

	t.StyleFunc(func(row, col int) lipgloss.Style {
		if row == table.HeaderRow {
			return lipgloss.NewStyle().Bold(true).Foreground(tui.AccentBlue)
		}
		switch col {
		case 0:
			return lipgloss.NewStyle().Foreground(tui.TextDim)
		case 1:
			return lipgloss.NewStyle().Bold(true).Foreground(tui.TextBright)
		case 5:
			return lipgloss.NewStyle().Foreground(tui.TextPrimary)
		default:
			return lipgloss.NewStyle()
		}
	})

	fmt.Fprintln(w, t.Render())
	summary := fmt.Sprintf("Total: %d server(s) configured (%d enabled, %d disabled)", len(servers), enabledCount, disabledCount)
	fmt.Fprintf(w, "\n%s\n\n", lipgloss.NewStyle().Foreground(tui.TextMuted).Render(summary))
}
