package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aim-cli/aim/internal/profile"
	"github.com/spf13/cobra"
)

func newBindCmd(pm *profile.ProfileManager) *cobra.Command {
	var targetDir string

	cmd := &cobra.Command{
		Use:   "bind <profile>",
		Short: "Bind current directory to an AIM profile",
		Long: `Binds the current directory or specified repository to an AIM profile
by writing an .aim-profile file. Commands like 'aim run <agent>' will automatically
use this profile when launched inside this directory.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]
			if err := profile.ValidateProfileName(profileName); err != nil {
				return err
			}
			dir := targetDir
			if dir == "" {
				var err error
				dir, err = os.Getwd()
				if err != nil {
					return err
				}
			}

			// Warn if profile is not configured
			if pm != nil && !pm.ProfileExists(profileName) {
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: profile '%s' does not currently exist. Run 'aim login <agent> %s' to configure it.\n", profileName, profileName)
			}

			filePath, err := profile.BindDirectory(dir, profileName)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "✔ Bound directory to profile '%s' (%s)\n", profileName, filePath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&targetDir, "dir", "d", "", "Directory to bind (defaults to current directory)")
	return cmd
}

func newUnbindCmd() *cobra.Command {
	var targetDir string

	cmd := &cobra.Command{
		Use:   "unbind",
		Short: "Remove profile binding from current directory",
		Long:  `Removes the .aim-profile file from the current directory or specified directory.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := targetDir
			if dir == "" {
				var err error
				dir, err = os.Getwd()
				if err != nil {
					return err
				}
			}

			removedPath, err := profile.UnbindDirectory(dir)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "✔ Removed directory profile binding (%s)\n", removedPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&targetDir, "dir", "d", "", "Directory to unbind (defaults to current directory)")
	return cmd
}

func newResolveCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "resolve [dir]",
		Short: "Resolve the AIM profile bound to a directory",
		Long: `Ascends from the target directory upwards looking for an .aim-profile file,
stopping at a .git repository boundary or filesystem root.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := ""
			if len(args) > 0 {
				dir = args[0]
			} else {
				var err error
				dir, err = os.Getwd()
				if err != nil {
					return err
				}
			}

			absDir, err := filepath.Abs(dir)
			if err != nil {
				return err
			}

			res, err := profile.ResolveDirProfile(absDir)
			if err != nil {
				return err
			}

			if jsonOutput {
				type jsonResult struct {
					Profile    *string `json:"profile"`
					SourceFile *string `json:"source_file,omitempty"`
					Directory  *string `json:"directory,omitempty"`
				}
				var out jsonResult
				if res != nil {
					out.Profile = &res.Profile
					out.SourceFile = &res.SourceFile
					out.Directory = &res.Directory
				}
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				return encoder.Encode(out)
			}

			if res == nil {
				fmt.Fprintf(cmd.OutOrStdout(), "No profile binding found for %s\n", absDir)
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Profile:     %s\n", res.Profile)
			fmt.Fprintf(cmd.OutOrStdout(), "Directory:   %s\n", res.Directory)
			fmt.Fprintf(cmd.OutOrStdout(), "Source File: %s\n", res.SourceFile)
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output resolved binding as JSON")
	return cmd
}
