package main

import (
	"fmt"
	"os"

	"strings"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/logger"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/tui"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

func newUICmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	return &cobra.Command{
		Use:   "ui",
		Short: "Open interactive TUI dashboard (default)",
		RunE: func(cmd *cobra.Command, args []string) error {
			code := runTUI(reg, pm)
			if code != 0 {
				return &ExitError{Code: code}
			}
			return nil
		},
	}
}

type profileInputModel struct {
	textInput textinput.Model
	cancelled bool
}

func initialProfileInputModel(agent, defaultProfile string) profileInputModel {
	ti := textinput.New()
	ti.Placeholder = "profile-name"
	if defaultProfile != "" {
		ti.SetValue(defaultProfile)
	}
	ti.Focus()
	ti.CharLimit = 64
	ti.Width = 32
	if defaultProfile != "" {
		ti.Prompt = fmt.Sprintf("Enter profile name for %s (default: %s): ", agent, defaultProfile)
	} else {
		ti.Prompt = fmt.Sprintf("Enter new profile name for %s: ", agent)
	}
	ti.PromptStyle = lipgloss.NewStyle().Bold(true).Foreground(tui.AccentBlue)

	return profileInputModel{
		textInput: ti,
	}
}

func (m profileInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m profileInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			return m, tea.Quit
		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancelled = true
			return m, tea.Quit
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m profileInputModel) View() string {
	return fmt.Sprintf("\n%s\n\n%s\n",
		m.textInput.View(),
		lipgloss.NewStyle().Foreground(tui.TextMuted).Render("(enter to confirm, esc to cancel)"),
	)
}

func promptProfileName(agent, defaultProfile string) string {
	p := tea.NewProgram(initialProfileInputModel(agent, defaultProfile))
	finalModel, err := p.Run()
	if err != nil {
		if defaultProfile != "" {
			fmt.Printf("Enter profile name for %s (default: %s): ", agent, defaultProfile)
		} else {
			fmt.Printf("Enter new profile name for %s: ", agent)
		}
		var name string
		fmt.Scanln(&name)
		name = strings.TrimSpace(name)
		if name == "" && defaultProfile != "" {
			return defaultProfile
		}
		return name
	}
	res, ok := finalModel.(profileInputModel)
	if !ok || res.cancelled {
		return ""
	}
	val := strings.TrimSpace(res.textInput.Value())
	if val == "" && defaultProfile != "" {
		return defaultProfile
	}
	return val
}

var tuiRunner = func(reg *agents.Registry, pm *profile.ProfileManager) int {
	cfg, _ := config.LoadConfig()
	m := tui.NewModel(reg, pm, cfg).WithVersion(Version)
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
		return 1
	}

	res, ok := finalModel.(tui.Model)
	if !ok {
		return 0
	}
	agent := res.SelectedAgent()
	switch res.Outcome() {
	case tui.ActionRun:
		return executeRun(reg, pm, agent, res.SelectedProfile(), nil)
	case tui.ActionShell:
		return executeShell(reg, pm, agent, res.SelectedProfile())
	case tui.ActionLogin:
		name := promptProfileName(agent, res.SelectedProfile())
		if name != "" {
			return executeLogin(reg, pm, agent, name)
		}
	}
	return 0
}

func runTUI(reg *agents.Registry, pm *profile.ProfileManager) int {
	logger.SetConsoleOutput(false)
	defer logger.SetConsoleOutput(true)
	logger.Debug("[tui] Launching interactive TUI dashboard")
	code := tuiRunner(reg, pm)
	logger.Debug("[tui] TUI exited with code %d", code)
	return code
}
