package installer

import (
	"github.com/bimalpaudels/finks/internal/checker"
	tea "github.com/charmbracelet/bubbletea"
)

// Stage constants for the wizard flow.
const (
	StageWelcome   = iota
	StageChecking
	StageInstalling
	StageVerifying
	StageDone
)

// welcomeDoneMsg is sent after the welcome delay to advance to checking.
type welcomeDoneMsg struct{}

// doneQuitMsg is sent after the "You're good to go" delay to quit.
type doneQuitMsg struct{}

// WizardState holds shared state for the wizard (stage, results, errors).
type WizardState struct {
	Stage       int
	Quitting    bool
	CheckResult checker.CheckResultMsg
	VerifyOK    bool
	VerifyErr   error
	InstallErr  error
}

// NewWizardState returns initial wizard state.
func NewWizardState() WizardState {
	return WizardState{Stage: StageWelcome}
}

// Advance advances to the next stage and returns the Cmd to run for that stage, if any.
func (s *WizardState) Advance(msg tea.Msg) tea.Cmd {
	switch s.Stage {
	case StageWelcome:
		if _, ok := msg.(welcomeDoneMsg); ok {
			s.Stage = StageChecking
			return checker.CheckDocker()
		}
	case StageChecking:
		if res, ok := msg.(checker.CheckResultMsg); ok {
			s.CheckResult = res
			if res.DockerOK {
				s.Stage = StageVerifying
				return checker.VerifyDocker()
			}
			s.Stage = StageInstalling
			return checker.RunInstallStep(res.DockerOK)
		}
	case StageInstalling:
		if res, ok := msg.(checker.InstallDoneMsg); ok {
			s.InstallErr = res.Err
			s.Stage = StageVerifying
			return checker.VerifyDocker()
		}
	case StageVerifying:
		if res, ok := msg.(checker.VerifyResultMsg); ok {
			s.VerifyOK = res.DockerOK
			s.VerifyErr = res.Err
			s.Stage = StageDone
			return nil
		}
	case StageDone:
		// No advance; done step handles quit tick
	}
	return nil
}
