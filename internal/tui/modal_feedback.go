package tui

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/diagnostics"
	"github.com/aim-cli/aim/internal/feedback"
	"github.com/aim-cli/aim/internal/telemetry"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type feedbackModalState struct {
	active        bool
	categoryIndex int // 0: general, 1: bug, 2: feature
	input         textinput.Model
	includeDoctor bool
	err           string
}

type feedbackResultMsg struct {
	err         error
	fallbackURL string
}

var feedbackCategories = []struct {
	category feedback.Category
	label    string
	icon     string
}{
	{feedback.CategoryGeneral, "General Feedback", "💬"},
	{feedback.CategoryBug, "Bug Report", "🐛"},
	{feedback.CategoryFeature, "Feature Request", "💡"},
}

func (m Model) IsFeedbackModalActive() bool {
	return m.feedbackModal.active
}

func (m Model) FeedbackModalInputValue() string {
	return m.feedbackModal.input.Value()
}

func (m Model) openFeedbackModal() (Model, tea.Cmd) {
	ti := textinput.New()
	ti.Placeholder = "Share suggestions, friction, or report an issue..."
	ti.CharLimit = 500
	ti.Width = 60
	cmd := ti.Focus()

	m.feedbackModal = feedbackModalState{
		active:        true,
		categoryIndex: 0,
		input:         ti,
		includeDoctor: false,
	}
	return m, cmd
}

func (m Model) updateFeedbackModal(msg tea.Msg) (Model, tea.Cmd) {
	km := m.keys.FeedbackModal

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, km.Quit):
			m.cancelStream()
			return m, tea.Quit

		case key.Matches(msg, km.Cancel):
			m.feedbackModal = feedbackModalState{}
			return m, nil

		case key.Matches(msg, km.Category):
			m.feedbackModal.categoryIndex = (m.feedbackModal.categoryIndex + 1) % len(feedbackCategories)
			return m, nil

		case key.Matches(msg, km.ToggleDoctor):
			m.feedbackModal.includeDoctor = !m.feedbackModal.includeDoctor
			return m, nil

		case key.Matches(msg, km.Submit):
			val := strings.TrimSpace(m.feedbackModal.input.Value())
			if val == "" {
				m.feedbackModal.err = "Feedback message cannot be empty"
				return m, nil
			}

			cat := feedbackCategories[m.feedbackModal.categoryIndex].category
			includeDoctor := m.feedbackModal.includeDoctor

			sub := feedback.Submission{
				Category:      cat,
				Message:       val,
				IncludeDoctor: includeDoctor,
				AIMVersion:    m.version,
				OS:            runtime.GOOS,
				Arch:          runtime.GOARCH,
				Timestamp:     time.Now().UTC(),
			}

			if includeDoctor {
				sub.DoctorReport = diagnostics.GenerateReport(m.reg, m.pm, m.cfg, m.agent, m.version, "")
			}

			endpoint := os.Getenv("AIM_FEEDBACK_ENDPOINT")
			d := feedback.NewDispatcher(endpoint)

			m.feedbackModal = feedbackModalState{}
			m.statusMessage = "Submitting feedback..."

			return m, submitFeedbackCmd(d, sub)

		default:
			var cmd tea.Cmd
			m.feedbackModal.input, cmd = m.feedbackModal.input.Update(msg)
			return m, cmd
		}

	default:
		var cmd tea.Cmd
		m.feedbackModal.input, cmd = m.feedbackModal.input.Update(msg)
		return m, cmd
	}
}

func submitFeedbackCmd(d *feedback.HTTPDispatcher, sub feedback.Submission) tea.Cmd {
	return func() tea.Msg {
		if d == nil || d.Endpoint == "" {
			telClient := telemetry.NewClient(config.BaseDir(), config.CacheDir(), sub.AIMVersion, nil)
			props := map[string]any{
				"category":       string(sub.Category),
				"message":        sub.Message,
				"include_doctor": sub.IncludeDoctor,
			}
			if sub.IncludeDoctor && sub.DoctorReport != "" {
				props["doctor_report"] = sub.DoctorReport
			}
			telClient.Track(telemetry.EventFeedbackSubmitted, props)

			flushCtx, flushCancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer flushCancel()
			_ = telClient.Flush(flushCtx)
			_ = telClient.Close()

			url := ""
			if d != nil {
				url = d.FallbackURL(sub)
			}
			return feedbackResultMsg{
				err:         nil,
				fallbackURL: url,
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := d.Submit(ctx, sub)
		fallbackURL := ""
		if err != nil {
			fallbackURL = d.FallbackURL(sub)
		}
		return feedbackResultMsg{err: err, fallbackURL: fallbackURL}
	}
}

func (m Model) renderFeedbackModal() string {
	var b strings.Builder

	title := FeedbackModalTitleStyle.Render("💬 Direct Feedback")
	b.WriteString(title + "\n\n")

	// Categories row
	b.WriteString(lipgloss.NewStyle().Foreground(TextMuted).Render("Category (Tab to cycle):") + "\n ")
	for i, c := range feedbackCategories {
		badge := fmt.Sprintf(" %s %s ", c.icon, c.label)
		if i == m.feedbackModal.categoryIndex {
			b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(tuiColorWhite).Background(AccentBlue).Render(badge))
		} else {
			b.WriteString(lipgloss.NewStyle().Foreground(TextSecondary).Render(badge))
		}
		b.WriteString("  ")
	}
	b.WriteString("\n\n")

	b.WriteString(lipgloss.NewStyle().Foreground(TextSecondary).Render("Your Message:") + "\n")
	b.WriteString("  " + m.feedbackModal.input.View() + "\n\n")

	// Diagnostic toggle
	doctorCheck := "[ ]"
	if m.feedbackModal.includeDoctor {
		doctorCheck = "[✓]"
	}
	b.WriteString(lipgloss.NewStyle().Foreground(TextMuted).Render(
		fmt.Sprintf("  %s Attach sanitized diagnostic report (Ctrl+D to toggle)", doctorCheck),
	) + "\n\n")

	if m.feedbackModal.err != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(StatusRed).Bold(true).Render(
			"  ✕ "+m.feedbackModal.err,
		) + "\n\n")
	}

	km := m.keys.FeedbackModal
	b.WriteString("  " + m.help.ShortHelpView(km.ShortHelp()))

	box := FeedbackModalBoxStyle.Render(b.String())
	return "\n" + box + "\n"
}

var tuiColorWhite = lipgloss.Color("#FFFFFF")
