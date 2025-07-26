package ui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines keybindings for the application. It satisfies help.KeyMap.
type KeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Enter    key.Binding
	Back     key.Binding
	Left     key.Binding
	Right    key.Binding
	Quit     key.Binding
	QR       key.Binding
	Refresh  key.Binding
	DateTime key.Binding
	Now      key.Binding
	AddVia   key.Binding
	DelVia   key.Binding
	Modify   key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Quit, k.Back, k.Left, k.Right, k.Up, k.Down, k.Enter, k.Refresh, k.DateTime, k.Now, k.AddVia, k.DelVia, k.Modify, k.QR}
}

// FullHelp returns keybindings for the expanded help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Quit, k.Back, k.Refresh, k.DateTime, k.Now, k.AddVia, k.DelVia, k.Modify, k.QR},
		{k.Left, k.Right, k.Up, k.Down, k.Enter},
	}
}

// DefaultKeyMap provides a default set of keybindings.
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
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "prev"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "next"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		QR: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "show QR"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		DateTime: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "date/time"),
		),
		Now: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "now"),
		),
		AddVia: key.NewBinding(
			key.WithKeys("+"),
			key.WithHelp("+", "add via"),
		),
		DelVia: key.NewBinding(
			key.WithKeys("-"),
			key.WithHelp("-", "remove via"),
		),
		Modify: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "modify"),
		),
	}
}
