package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/presenter"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/tui"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
)

func newConfigCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	var jsonFlag bool

	cmd := &cobra.Command{
		Use:   "config [command]",
		Short: "Inspect and manage AIM global configuration",
		Long: `Inspect and modify settings in AIM's central configuration file (~/.aim/config.json).

Subcommands:
  path                  Print absolute path to active config.json
  list                  Display all configuration keys and values
  get <key>             Print the value of a configuration key
  set <key> <value>     Update the value of a configuration key`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigList(cmd.OutOrStdout(), jsonFlag)
		},
	}

	cmd.Flags().BoolVarP(&jsonFlag, "json", "j", false, "Output configuration as JSON")

	pathCmd := &cobra.Command{
		Use:   "path",
		Short: "Print absolute path to active config.json",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), config.ConfigFilePath())
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all configuration keys and their current values",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigList(cmd.OutOrStdout(), jsonFlag)
		},
	}
	listCmd.Flags().BoolVarP(&jsonFlag, "json", "j", false, "Output configuration as JSON")

	configKeys := []string{"default_agent", "default_profile", "auto_update", "telemetry", "debug"}

	getCmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get the value of a configuration key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigGet(cmd.OutOrStdout(), args[0], jsonFlag)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				var matches []string
				for _, k := range configKeys {
					if strings.HasPrefix(k, toComplete) {
						matches = append(matches, k)
					}
				}
				return matches, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}
	getCmd.Flags().BoolVarP(&jsonFlag, "json", "j", false, "Output result as JSON")

	setCmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set the value of a configuration key",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigSet(cmd.OutOrStdout(), reg, pm, args[0], args[1])
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				var matches []string
				for _, k := range configKeys {
					if strings.HasPrefix(k, toComplete) {
						matches = append(matches, k)
					}
				}
				return matches, cobra.ShellCompDirectiveNoFileComp
			}
			if len(args) == 1 {
				key := strings.ToLower(strings.TrimSpace(args[0]))
				switch key {
				case "auto_update":
					return []string{config.AutoUpdatePrompt, config.AutoUpdateNotify, config.AutoUpdateAuto, config.AutoUpdateOff}, cobra.ShellCompDirectiveNoFileComp
				case "telemetry":
					return []string{"on", "off"}, cobra.ShellCompDirectiveNoFileComp
				case "debug":
					return []string{"true", "false"}, cobra.ShellCompDirectiveNoFileComp
				case "default_agent":
					return completeAgents(reg, toComplete), cobra.ShellCompDirectiveNoFileComp
				case "default_profile":
					if pm != nil {
						profs, _ := pm.ListProfiles()
						return profs, cobra.ShellCompDirectiveNoFileComp
					}
				}
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.AddCommand(pathCmd, listCmd, getCmd, setCmd)
	return cmd
}

func runConfigList(out io.Writer, asJSON bool) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if asJSON {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(cfg)
	}

	banner := lipgloss.NewStyle().Bold(true).Foreground(tui.AccentBlue).Render("=== AIM Configuration ===")
	pathNote := lipgloss.NewStyle().Foreground(tui.TextMuted).Render(fmt.Sprintf("Path: %s", config.ConfigFilePath()))
	fmt.Fprintf(out, "%s\n%s\n\n", banner, pathNote)

	t := presenter.NewTable(out).
		Border(lipgloss.HiddenBorder()).
		Headers("KEY", "VALUE", "DESCRIPTION")

	telStatus := "on"
	if cfg.TelemetryDisabled {
		telStatus = "off"
	}

	debugStatus := "false"
	if cfg.Debug {
		debugStatus = "true"
	}

	t.Row("default_agent", cfg.DefaultAgent, "Default agent to launch when unspecified")
	t.Row("default_profile", cfg.DefaultProfile, "Default profile activated when none specified")
	t.Row("auto_update", cfg.GetAutoUpdate(), "Update policy (prompt, notify, auto, off)")
	t.Row("telemetry", telStatus, "Anonymous usage telemetry (on, off)")
	t.Row("debug", debugStatus, "Verbose debug logging to stderr and aim-debug.log")
	t.Row("custom_bridged_paths", fmt.Sprintf("%d path(s)", len(cfg.CustomBridgedPaths)), "Host paths bridged into profile sandboxes")
	t.Row("profiles", fmt.Sprintf("%d configured", len(cfg.Profiles)), "Profiles registered in config")

	t.StyleFunc(func(row, col int) lipgloss.Style {
		if row == table.HeaderRow {
			return lipgloss.NewStyle().Bold(true).Foreground(tui.AccentBlue)
		}
		switch col {
		case 0:
			return lipgloss.NewStyle().Bold(true).Foreground(tui.AccentCyan)
		case 1:
			return lipgloss.NewStyle().Foreground(tui.TextBright)
		default:
			return lipgloss.NewStyle().Foreground(tui.TextDim)
		}
	})

	fmt.Fprintln(out, t.Render())
	return nil
}

func runConfigGet(out io.Writer, key string, asJSON bool) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	val, err := cfg.Get(key)
	if err != nil {
		return err
	}

	if asJSON {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(map[string]string{
			"key":   key,
			"value": val,
		})
	}

	fmt.Fprintln(out, val)
	return nil
}

func runConfigSet(out io.Writer, reg *agents.Registry, pm *profile.ProfileManager, key, val string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	normKey := strings.ToLower(strings.TrimSpace(key))

	if normKey == "default_agent" && reg != nil {
		canonical := val
		if ad, err := reg.Get(val); err == nil {
			canonical = ad.Name()
		} else {
			return fmt.Errorf("unknown agent %q", val)
		}
		val = canonical
	}

	if err := cfg.Set(key, val); err != nil {
		return err
	}

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	displayVal, _ := cfg.Get(normKey)
	if displayVal == "" {
		displayVal = val
	}

	successBadge := lipgloss.NewStyle().Foreground(tui.StatusGreen).Bold(true).Render("✔")
	keyStyle := lipgloss.NewStyle().Foreground(tui.AccentCyan).Render(normKey)
	valStyle := lipgloss.NewStyle().Foreground(tui.TextBright).Bold(true).Render(displayVal)

	fmt.Fprintf(out, "%s Set %s = %s\n", successBadge, keyStyle, valStyle)
	if normKey == "default_profile" && pm != nil && val != "" && !pm.ProfileExists(val) {
		noteBadge := lipgloss.NewStyle().Foreground(tui.StatusYellow).Render("Note:")
		fmt.Fprintf(out, "  %s Profile %q does not exist yet. Run 'aim login <agent> %s' to create it.\n", noteBadge, val, val)
	}
	return nil
}
