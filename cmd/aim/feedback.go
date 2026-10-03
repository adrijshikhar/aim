package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/feedback"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/telemetry"
	"github.com/aim-cli/aim/internal/tui"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

func newFeedbackCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	var categoryStr string
	var message string
	var includeDoctor bool
	var endpoint string

	cmd := &cobra.Command{
		Use:   "feedback",
		Short: "Submit feedback, feature suggestions, or bug reports",
		RunE: func(cmd *cobra.Command, args []string) error {
			cat := parseCategory(categoryStr)

			if strings.TrimSpace(message) == "" && len(args) > 0 {
				message = strings.Join(args, " ")
			}

			if strings.TrimSpace(message) == "" {
				if isTerminal() {
					var err error
					cat, message, includeDoctor, err = runInteractiveFeedbackPrompt()
					if err != nil {
						return err
					}
				} else {
					// Read piped input
					data, err := io.ReadAll(os.Stdin)
					if err == nil {
						message = strings.TrimSpace(string(data))
					}
				}
			}

			if strings.TrimSpace(message) == "" {
				return fmt.Errorf("feedback message cannot be empty (use --message or provide input)")
			}

			sub := feedback.Submission{
				Category:      cat,
				Message:       message,
				IncludeDoctor: includeDoctor,
				AIMVersion:    Version,
				OS:            runtime.GOOS,
				Arch:          runtime.GOARCH,
				Timestamp:     time.Now().UTC(),
			}

			if includeDoctor {
				sub.DoctorReport = generateDiagnosticReport(reg, pm, "")
			}

			if endpoint == "" {
				endpoint = os.Getenv("AIM_FEEDBACK_ENDPOINT")
			}

			d := feedback.NewDispatcher(endpoint)
			if d.Endpoint == "" {
				fallbackURL := d.FallbackURL(sub)
				fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(tui.AccentBlue).Render("Direct In-Tool Feedback"))
				fmt.Println("No automated ingest endpoint configured.")
				fmt.Println("You can submit this directly as a pre-filled GitHub issue:")
				fmt.Printf("\n  %s\n\n", lipgloss.NewStyle().Foreground(tui.AccentCyan).Underline(true).Render(fallbackURL))
				return nil
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			fmt.Print("Submitting feedback...")
			if err := d.Submit(ctx, sub); err != nil {
				fmt.Println()
				fmt.Printf("⚠️  Failed to submit feedback directly: %v\n", err)
				fallbackURL := d.FallbackURL(sub)
				fmt.Println("You can submit this via GitHub instead:")
				fmt.Printf("  %s\n", fallbackURL)
				return nil
			}

			fmt.Println("\r" + tui.GaugeGreenStyle.Render("✓ Thank you! Your feedback has been submitted."))
			telClient := telemetry.NewClient(config.BaseDir(), config.CacheDir(), Version, nil)
			telClient.Track(telemetry.EventFeedbackSubmitted, map[string]any{"category": string(cat)})
			_ = telClient.Close()
			return nil
		},
	}

	cmd.Flags().StringVarP(&categoryStr, "type", "t", "general", "Feedback category (bug, feature, general)")
	cmd.Flags().StringVarP(&message, "message", "m", "", "Feedback message text")
	cmd.Flags().BoolVar(&includeDoctor, "include-doctor", false, "Attach sanitized diagnostic report (aim doctor --report)")
	cmd.Flags().StringVar(&endpoint, "endpoint", "", "Custom feedback ingest endpoint")

	return cmd
}

func parseCategory(s string) feedback.Category {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "bug", "crash", "error", "1":
		return feedback.CategoryBug
	case "feature", "enhancement", "idea", "2":
		return feedback.CategoryFeature
	default:
		return feedback.CategoryGeneral
	}
}

func runInteractiveFeedbackPrompt() (feedback.Category, string, bool, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(tui.AccentBlue).Render("=== AIM Feedback ==="))
	fmt.Println("Select category:")
	fmt.Println("  [1] 🐛 Bug report")
	fmt.Println("  [2] 💡 Feature request")
	fmt.Println("  [3] 💬 General feedback / UX thoughts")
	fmt.Print("Enter choice [1-3] (default 3): ")

	catInput, _ := reader.ReadString('\n')
	cat := parseCategory(catInput)

	fmt.Print("\nYour message / details: ")
	msgInput, err := reader.ReadString('\n')
	if err != nil {
		return cat, "", false, err
	}
	message := strings.TrimSpace(msgInput)

	fmt.Print("\nInclude sanitized diagnostic report (aim doctor --report)? [y/N]: ")
	docInput, _ := reader.ReadString('\n')
	includeDoc := strings.HasPrefix(strings.ToLower(strings.TrimSpace(docInput)), "y")

	return cat, message, includeDoc, nil
}
