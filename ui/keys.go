package ui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines keybindings for the application
type KeyMap struct {
	// Navigation
	Up       key.Binding
	Down     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Home     key.Binding
	End      key.Binding

	// Selection
	Select    key.Binding
	SelectAll key.Binding
	Deselect  key.Binding

	// Actions
	Apply         key.Binding
	Commit        key.Binding
	ModifyHead    key.Binding
	Search        key.Binding
	TogglePreview key.Binding
	ToggleHelp    key.Binding
	Quit          key.Binding
}

// DefaultKeyMap returns the default keybindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "b"),
			key.WithHelp("pgup/b", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdn", "f"),
			key.WithHelp("pgdn/f", "page down"),
		),
		Home: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("home/g", "go to top"),
		),
		End: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("end/G", "go to bottom"),
		),
		Select: key.NewBinding(
			key.WithKeys("space", "tab"),
			key.WithHelp("space/tab", "select file"),
		),
		SelectAll: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "select all"),
		),
		Deselect: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "deselect all"),
		),
		Apply: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "stage/unstage"),
		),
		Commit: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "commit"),
		),
		ModifyHead: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "modify HEAD"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		TogglePreview: key.NewBinding(
			key.WithKeys("p", "P"),
			key.WithHelp("p", "toggle preview"),
		),
		ToggleHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

// ShortHelp returns bindings to show in the short help
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Select, k.Apply, k.TogglePreview, k.ToggleHelp, k.Quit}
}

// FullHelp returns all bindings grouped for the full help view
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.PageUp, k.PageDown, k.Home, k.End},
		{k.Select, k.SelectAll, k.Deselect},
		{k.Apply, k.Commit, k.ModifyHead},
		{k.Search, k.TogglePreview, k.ToggleHelp, k.Quit},
	}
}
