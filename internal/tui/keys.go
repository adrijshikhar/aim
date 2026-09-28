package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines the keybindings for the root TUI dashboard as well as sub-component keymaps.
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

	SessionsDrawer SessionsDrawerKeyMap
	DoctorDrawer   DoctorDrawerKeyMap
	ResumeModal    ResumeModalKeyMap
	DeleteModal    DeleteModalKeyMap
	MoveModal      MoveModalKeyMap
	RenameModal    RenameModalKeyMap
}

// DefaultKeyMap returns the default set of keybindings for the root dashboard and all child components.
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

		SessionsDrawer: DefaultSessionsDrawerKeyMap(),
		DoctorDrawer:   DefaultDoctorDrawerKeyMap(),
		ResumeModal:    DefaultResumeModalKeyMap(),
		DeleteModal:    DefaultDeleteModalKeyMap(),
		MoveModal:      DefaultMoveModalKeyMap(),
		RenameModal:    DefaultRenameModalKeyMap(),
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

// SessionsDrawerKeyMap manages keys for the sessions explorer.
type SessionsDrawerKeyMap struct {
	Up          key.Binding
	Down        key.Binding
	Enter       key.Binding
	Flags       key.Binding
	Catalyst    key.Binding
	Fork        key.Binding
	Filter      key.Binding
	TabFocus    key.Binding
	AgentCycle  key.Binding
	AgentAll    key.Binding
	AgentAgy    key.Binding
	AgentCodex  key.Binding
	ClearFilter key.Binding
	Close       key.Binding
	Quit        key.Binding
}

// DefaultSessionsDrawerKeyMap returns default bindings for list navigation mode.
func DefaultSessionsDrawerKeyMap() SessionsDrawerKeyMap {
	return SessionsDrawerKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k", "ctrl+p"),
			key.WithHelp("[↑/↓]", "Navigate"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j", "ctrl+n"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("[Enter]", "Resume"),
		),
		Flags: key.NewBinding(
			key.WithKeys("f", "ctrl+f"),
			key.WithHelp("[f]", "Flags"),
		),
		Catalyst: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("[c]", "Catalyst"),
		),
		Fork: key.NewBinding(
			key.WithKeys("b"),
			key.WithHelp("[b]", "Fork"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("[/]", "Filter"),
		),
		TabFocus: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("[/ / Tab]", "Filter"),
		),
		AgentCycle: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "cycle agent"),
		),
		AgentAll: key.NewBinding(
			key.WithKeys("0"),
			key.WithHelp("0", "all"),
		),
		AgentAgy: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "antigravity"),
		),
		AgentCodex: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "codex"),
		),
		ClearFilter: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("[Esc]", "Clear"),
		),
		Close: key.NewBinding(
			key.WithKeys("esc", "q", "s"),
			key.WithHelp("[Esc]", "Close"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
	}
}

// ForFilterMode returns bindings configured specifically for active filter text entry.
// In this mode, letter keys (k, j, f, c, b) are NOT intercepted as navigation, allowing them
// to be typed into the search box, while arrow keys, ctrl sequences, Tab, and Esc remain active.
func (k SessionsDrawerKeyMap) ForFilterMode() SessionsDrawerKeyMap {
	m := k
	m.Up = key.NewBinding(key.WithKeys("up", "ctrl+p", "ctrl+k"), key.WithHelp("[↑/↓]", "Navigate"))
	m.Down = key.NewBinding(key.WithKeys("down", "ctrl+n", "ctrl+j"))
	m.Enter = key.NewBinding(key.WithKeys("enter"), key.WithHelp("[Enter]", "Resume"))
	m.Flags = key.NewBinding(key.WithKeys("ctrl+f"), key.WithHelp("[Ctrl+F]", "Flags"))
	m.TabFocus = key.NewBinding(key.WithKeys("tab", "esc"), key.WithHelp("[Tab/Esc]", "Exit Filter"))
	m.ClearFilter = key.NewBinding(key.WithKeys("esc"), key.WithHelp("[Esc]", "Exit Filter"))
	m.Catalyst.SetEnabled(false)
	m.Fork.SetEnabled(false)
	m.Filter.SetEnabled(false)
	m.AgentAll.SetEnabled(false)
	m.AgentAgy.SetEnabled(false)
	m.AgentCodex.SetEnabled(false)
	return m
}

// ShortHelp returns default navigation bindings for sessions drawer footer.
func (k SessionsDrawerKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Enter, k.Flags, k.Catalyst, k.Fork, k.Filter, k.Close}
}

// ShortHelpFilter returns bindings when text filter input is actively focused.
func (k SessionsDrawerKeyMap) ShortHelpFilter() []key.Binding {
	fm := k.ForFilterMode()
	return []key.Binding{fm.Up, fm.Enter, fm.Flags, fm.TabFocus}
}

// ShortHelpQuery returns bindings when drawer has a filter query but list is focused.
func (k SessionsDrawerKeyMap) ShortHelpQuery() []key.Binding {
	return []key.Binding{k.Up, k.Enter, k.Flags, k.Catalyst, k.Fork, k.Filter, k.ClearFilter}
}

// FullHelp returns grouped bindings for extended help views.
func (k SessionsDrawerKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter},
		{k.Flags, k.Catalyst, k.Fork},
		{k.Filter, k.TabFocus, k.ClearFilter},
		{k.AgentAll, k.AgentAgy, k.AgentCodex, k.Close},
	}
}

// DoctorDrawerKeyMap manages keys for the doctor diagnostics drawer.
type DoctorDrawerKeyMap struct {
	Up        key.Binding
	Down      key.Binding
	NextAgent key.Binding
	PrevAgent key.Binding
	Agent1    key.Binding
	Agent2    key.Binding
	Agent3    key.Binding
	Close     key.Binding
	Quit      key.Binding
}

// DefaultDoctorDrawerKeyMap returns default bindings for doctor diagnostics drawer.
func DefaultDoctorDrawerKeyMap() DoctorDrawerKeyMap {
	return DoctorDrawerKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("[↑/↓]", "Inspect Profile"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
		),
		NextAgent: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("[Tab]", "Switch Agent"),
		),
		PrevAgent: key.NewBinding(
			key.WithKeys("shift+tab"),
		),
		Agent1: key.NewBinding(
			key.WithKeys("1"),
		),
		Agent2: key.NewBinding(
			key.WithKeys("2"),
		),
		Agent3: key.NewBinding(
			key.WithKeys("3"),
		),
		Close: key.NewBinding(
			key.WithKeys("d", "esc", "q"),
			key.WithHelp("[d/Esc/q]", "Close Drawer"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
	}
}

func (k DoctorDrawerKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.NextAgent, k.Close}
}

func (k DoctorDrawerKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},
		{k.NextAgent, k.PrevAgent},
		{k.Close},
	}
}

// ResumeModalKeyMap manages keys for the session resumption modal.
type ResumeModalKeyMap struct {
	Up          key.Binding
	Down        key.Binding
	Enter       key.Binding
	ToggleFlags key.Binding
	DoneFlags   key.Binding
	Cancel      key.Binding
	Quit        key.Binding
}

// DefaultResumeModalKeyMap returns default bindings for session resumption modal.
func DefaultResumeModalKeyMap() ResumeModalKeyMap {
	return ResumeModalKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("[↑/↓]", "Select Profile"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("[Enter]", "Confirm & Resume"),
		),
		ToggleFlags: key.NewBinding(
			key.WithKeys("tab", "f", "e"),
			key.WithHelp("[f/Tab]", "Flags"),
		),
		DoneFlags: key.NewBinding(
			key.WithKeys("tab", "esc"),
			key.WithHelp("[Tab/Esc]", "Done Editing Flags"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc", "q"),
			key.WithHelp("[Esc]", "Back"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
	}
}

func (k ResumeModalKeyMap) ShortHelp() []key.Binding {
	return k.ShortHelpProfile()
}

func (k ResumeModalKeyMap) ShortHelpProfile() []key.Binding {
	return []key.Binding{k.Up, k.ToggleFlags, k.Enter, k.Cancel}
}

func (k ResumeModalKeyMap) ShortHelpFlags() []key.Binding {
	return []key.Binding{k.Enter, k.DoneFlags}
}

func (k ResumeModalKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter},
		{k.ToggleFlags, k.DoneFlags, k.Cancel},
	}
}

// DeleteModalKeyMap manages keys for the profile deletion modal.
type DeleteModalKeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Left    key.Binding
	Right   key.Binding
	Confirm key.Binding
	QuickY  key.Binding
	Select1 key.Binding
	Select2 key.Binding
	Cancel  key.Binding
	Quit    key.Binding
}

// DefaultDeleteModalKeyMap returns default bindings for delete confirmation modal.
func DefaultDeleteModalKeyMap() DeleteModalKeyMap {
	return DeleteModalKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h", "shift+tab"),
			key.WithHelp("[←/→/Tab]", "Select"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l", "tab"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("[Enter]", "Confirm"),
		),
		QuickY: key.NewBinding(
			key.WithKeys("y", "Y"),
			key.WithHelp("[Enter/y]", "Confirm"),
		),
		Select1: key.NewBinding(
			key.WithKeys("1"),
		),
		Select2: key.NewBinding(
			key.WithKeys("2"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc", "q", "n"),
			key.WithHelp("[Esc]", "Cancel"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
	}
}

func (k DeleteModalKeyMap) ShortHelp() []key.Binding {
	return k.ShortHelpSingle()
}

func (k DeleteModalKeyMap) ShortHelpShared() []key.Binding {
	return []key.Binding{k.Left, k.Confirm, k.Cancel}
}

func (k DeleteModalKeyMap) ShortHelpSingle() []key.Binding {
	return []key.Binding{k.Left, k.QuickY, k.Cancel}
}

func (k DeleteModalKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Left, k.Right, k.Confirm},
		{k.QuickY, k.Cancel},
	}
}

// MoveModalKeyMap manages keys for the profile move/link modal.
type MoveModalKeyMap struct {
	Submit key.Binding
	Cancel key.Binding
	Quit   key.Binding
}

// DefaultMoveModalKeyMap returns default bindings for move profile modal.
func DefaultMoveModalKeyMap() MoveModalKeyMap {
	return MoveModalKeyMap{
		Submit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("[Enter]", "Confirm"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("[Esc]", "Cancel"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
	}
}

func (k MoveModalKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Submit, k.Cancel}
}

func (k MoveModalKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Submit, k.Cancel},
	}
}

// RenameModalKeyMap manages keys for the profile rename modal.
type RenameModalKeyMap struct {
	Submit key.Binding
	Cancel key.Binding
	Quit   key.Binding
}

// DefaultRenameModalKeyMap returns default bindings for rename profile modal.
func DefaultRenameModalKeyMap() RenameModalKeyMap {
	return RenameModalKeyMap{
		Submit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("[Enter]", "Confirm"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("[Esc]", "Cancel"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
	}
}

func (k RenameModalKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Submit, k.Cancel}
}

func (k RenameModalKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Submit, k.Cancel},
	}
}
