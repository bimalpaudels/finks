package installer

import (
	"fmt"
	"time"

	"github.com/bimalpaudels/finks/internal/installer/requirements"
	"github.com/bimalpaudels/finks/internal/installer/steps"
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

// model represents the wizard's state and Bubble Tea model.
type model struct {
	state WizardState
}

// newModel creates a new wizard model.
func newModel() model {
	return model{
		state: NewWizardState(),
	}
}

// Init initializes the model: start welcome delay, then send welcomeDoneMsg.
func (m model) Init() tea.Cmd {
	return tea.Tick(1500*time.Millisecond, func(t time.Time) tea.Msg {
		return welcomeDoneMsg{}
	})
}

// Update handles messages and advances stages.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		// Global quit keys
		if key == "q" || key == "esc" || key == "ctrl+c" {
			m.state.CloseRequirements()
			m.state.Quitting = true
			return m, tea.Quit
		}

		// Handle Enter in prompt stage
		if m.state.Stage == StagePromptInstall && key == "enter" {
			cmd := m.state.Advance(enterPressedMsg{})
			return m, cmd
		}
	}

	cmd := m.state.Advance(msg)
	if cmd != nil {
		return m, cmd
	}

	// After transitioning to Done from verify, start quit delay
	if m.state.Stage == StageDone {
		if _, ok := msg.(requirements.VerifyRequirementMsg); ok {
			return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
				return doneQuitMsg{}
			})
		}
		// Also handle when we skip directly to done (all requirements already OK)
		if _, ok := msg.(requirements.CheckRequirementMsg); ok {
			return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
				return doneQuitMsg{}
			})
		}
	}

	switch msg.(type) {
	case doneQuitMsg:
		m.state.CloseRequirements()
		m.state.Quitting = true
		return m, tea.Quit
	}

	return m, nil
}

// View renders the UI for the current stage.
func (m model) View() string {
	if m.state.Quitting {
		return ""
	}

	switch m.state.Stage {
	case StageWelcome:
		return steps.WelcomeView()
	case StageChecking:
		return steps.CheckView()
	case StagePromptInstall:
		return steps.PromptInstallView(
			m.state.CurrentReqName,
			m.state.CurrentInstallCmd,
			m.state.CurrentCanAutoInstall,
		)
	case StageInstalling:
		return steps.InstallView(false) // false = not already present, we're installing
	case StageVerifying:
		return steps.VerifyView()
	case StageDone:
		return steps.DoneViewSimple(m.state.VerifyOK, m.state.VerifyErr, m.state.InstallErr)
	default:
		return steps.WelcomeView()
	}
}
