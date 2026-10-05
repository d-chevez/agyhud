package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Run launches the interactive configuration TUI in full alternate screen mode.
func Run(cfgPath string) error {
	m, err := InitialModel(cfgPath)
	if err != nil {
		return err
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}
