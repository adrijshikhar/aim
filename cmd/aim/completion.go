package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newCompletionCmd(rootCmd *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:   "completion <bash|zsh|fish>",
		Short: "Generate shell completion script (zsh, bash, fish)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 || args[0] == "" {
				fmt.Println("Usage: aim completion <zsh|bash|fish>")
				return &ExitError{Code: 1}
			}
			shell := args[0]
			switch shell {
			case "bash":
				return rootCmd.GenBashCompletionV2(cmd.OutOrStdout(), true)
			case "zsh":
				return rootCmd.GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				return rootCmd.GenFishCompletion(cmd.OutOrStdout(), true)
			default:
				return fmt.Errorf("unsupported shell %q (supported: bash, zsh, fish)", shell)
			}
		},
		ValidArgs: []string{"bash", "zsh", "fish"},
	}
}
