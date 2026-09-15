package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines the keybindings for the TUI application.
type KeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Run      key.Binding
	Sessions key.Binding
	Shell    key.Binding
	Login    key.Binding
	Tab      key.Binding
	Doctor   key.Binding
	Rename   key.Binding
	Move     key.Binding
	Delete   key.Binding
	Refresh  key.Binding
	Filter   key.Binding
	Help     key.Binding
	Quit     key.Binding
}

// DefaultKeyMap returns the default set of keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Run: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "run"),
		),
		Sessions: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "sessions"),
		),
		Shell: key.NewBinding(
			key.WithKeys("S", "$"),
			key.WithHelp("S", "shell"),
		),
		Login: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "login"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch agent"),
		),
		Doctor: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "doctor"),
		),
		Rename: key.NewBinding(
			key.WithKeys("m", "R"),
			key.WithHelp("m/R", "rename"),
		),
		Move: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "move"),
		),
		Delete: key.NewBinding(
			key.WithKeys("x", "delete"),
			key.WithHelp("x", "delete"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc", "ctrl+c"),
			key.WithHelp("q/esc", "quit"),
		),
	}
}

// ShortHelp returns keybindings to be shown in the mini help view.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Run, k.Sessions, k.Shell, k.Login, k.Tab, k.Doctor, k.Rename, k.Move, k.Delete, k.Quit}
}

// FullHelp returns keybindings for the expanded help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Run, k.Sessions},
		{k.Shell, k.Login, k.Tab, k.Doctor},
		{k.Rename, k.Move, k.Delete, k.Refresh, k.Filter, k.Help, k.Quit},
	}
}
