package main

import (
	"context"
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
)

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

func handleTUIOutcome(res tui.Model, reg *agents.Registry, pm *profile.ProfileManager) int {
	agent := res.SelectedAgent()
	switch res.Outcome() {
	case tui.ActionRun:
		return executeRun(reg, pm, agent, res.SelectedProfile(), nil)
	case tui.ActionLogin:
		name := promptProfileName(agent, res.SelectedProfile())
		if name != "" {
			return executeLogin(reg, pm, agent, name)
		}
	case tui.ActionResumeExact:
		sess := res.SelectedSession()
		if sess != nil {
			targetProfile := res.SelectedProfile()
			if targetProfile == "" || targetProfile == "<host>" {
				targetProfile = sess.Profile
			}
			if targetProfile == "" || targetProfile == "<host>" {
				targetProfile = "default"
			}
			mgr := getSessionManager(context.Background())
			pDir, err := pm.EnsureProfile(targetProfile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error ensuring profile: %v\n", err)
				return 1
			}
			resumeCmd := newResumeCmd(reg, pm)
			extraArgs := res.SelectedArgs()
			if err := executeExactResume(resumeCmd, reg, pm, mgr, sess.Agent, targetProfile, pDir, sess, res.IsForkResume(), extraArgs); err != nil {
				fmt.Fprintf(os.Stderr, "Resume error: %v\n", err)
				return 1
			}
			return 0
		}
	case tui.ActionResumeCatalyst:
		sess := res.SelectedSession()
		if sess != nil {
			targetProfile := res.SelectedProfile()
			if targetProfile == "" || targetProfile == "<host>" {
				targetProfile = sess.Profile
			}
			if targetProfile == "" || targetProfile == "<host>" {
				targetProfile = "default"
			}
			resumeCmd := newResumeCmd(reg, pm)
			extraArgs := res.SelectedArgs()
			if err := executeCatalystResume(resumeCmd, reg, pm, sess.Agent, targetProfile, sess, extraArgs); err != nil {
				fmt.Fprintf(os.Stderr, "Catalyst resume error: %v\n", err)
				return 1
			}
			return 0
		}
	}
	return 0
}

var tuiRunner = func(reg *agents.Registry, pm *profile.ProfileManager) int {
	cfg, _ := config.LoadConfig()
	m := tui.NewModel(reg, pm, cfg).WithVersion(Version)
	p := tea.NewProgram(m)
	// Console logging would draw over the TUI; the session it launches below
	// runs after the TUI has closed and must show its warnings.
	logger.SetConsoleOutput(false)
	finalModel, err := p.Run()
	logger.SetConsoleOutput(true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
		return 1
	}

	res, ok := finalModel.(tui.Model)
	if !ok {
		return 0
	}
	return handleTUIOutcome(res, reg, pm)
}

var tuiSessionsRunner = func(reg *agents.Registry, pm *profile.ProfileManager, initialAgent, profileFilter string, activeOnly bool) int {
	cfg, _ := config.LoadConfig()
	m := tui.NewModel(reg, pm, cfg).
		WithVersion(Version).
		WithSessionsDrawerConfig(initialAgent, profileFilter, activeOnly, true)
	p := tea.NewProgram(m)
	logger.SetConsoleOutput(false)
	finalModel, err := p.Run()
	logger.SetConsoleOutput(true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
		return 1
	}

	res, ok := finalModel.(tui.Model)
	if !ok {
		return 0
	}
	return handleTUIOutcome(res, reg, pm)
}

func runTUIWithContext(ctx context.Context, reg *agents.Registry, pm *profile.ProfileManager) int {
	logger.Debug("[tui] Launching interactive TUI dashboard")
	if ctx != nil {
		if fn := TUIRunnerFromContext(ctx); fn != nil {
			return fn(reg, pm)
		}
	}
	code := tuiRunner(reg, pm)
	logger.Debug("[tui] TUI exited with code %d", code)
	return code
}

func runTUI(reg *agents.Registry, pm *profile.ProfileManager) int {
	return runTUIWithContext(context.Background(), reg, pm)
}

func runTUISessionsWithContext(ctx context.Context, reg *agents.Registry, pm *profile.ProfileManager, initialAgent, profileFilter string, activeOnly bool) int {
	logger.Debug("[tui] Launching interactive sessions drawer")
	if pm == nil {
		pm = profile.NewProfileManager(config.BaseDir())
	}
	if ctx != nil {
		if fn := TUISessionsRunnerFromContext(ctx); fn != nil {
			return fn(reg, pm, initialAgent, profileFilter, activeOnly)
		}
	}
	code := tuiSessionsRunner(reg, pm, initialAgent, profileFilter, activeOnly)
	logger.Debug("[tui] Sessions drawer exited with code %d", code)
	return code
}

func runTUISessions(reg *agents.Registry, pm *profile.ProfileManager, initialAgent, profileFilter string, activeOnly bool) int {
	return runTUISessionsWithContext(context.Background(), reg, pm, initialAgent, profileFilter, activeOnly)
}
