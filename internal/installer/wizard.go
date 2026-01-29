package installer

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Run starts the installation wizard.
func Run() error {
	p := tea.NewProgram(newModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run wizard: %w", err)
	}
	return nil
}

// model represents the wizard's state.
type model struct {
	quitting bool
}

// newModel creates a new wizard model.
func newModel() model {
	return model{
		quitting: false,
	}
}

// Init initializes the model.
func (m model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			m.quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the UI.
func (m model) View() string {
	if m.quitting {
		return ""
	}

	return "\n" +
		"┌─────────────────────────────────────┐\n" +
		"│      Welcome to Finks 🦜           │\n" +
		"└─────────────────────────────────────┘\n" +
		"\nPress Enter or 'q' to exit...\n"
}
