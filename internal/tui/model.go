// Package tui provides the per-session Bubble Tea model for SSH Battle Arena.
// It renders world snapshots and forwards keystrokes — it never mutates game state.
package tui

import tea "github.com/charmbracelet/bubbletea"

// Model is the per-session Bubble Tea model.
type Model struct{}

// NewModel returns a fresh TUI model for a new SSH session.
func NewModel() Model {
	return Model{}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	return "Welcome to SSH Battle Arena!\n"
}
