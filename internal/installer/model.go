package installer

import (
	"github.com/bimalpaudels/finks/internal/installer/requirements"
	tea "github.com/charmbracelet/bubbletea"
)

// Stage constants for the wizard flow.
const (
	StageWelcome = iota
	StageChecking
	StagePromptInstall
	StageInstalling
	StageVerifying
	StageDone
)

// welcomeDoneMsg is sent after the welcome delay to advance to checking.
type welcomeDoneMsg struct{}

// doneQuitMsg is sent after the "You're good to go" delay to quit.
type doneQuitMsg struct{}

// enterPressedMsg signals the user pressed Enter in the prompt stage.
type enterPressedMsg struct{}

// WizardState holds shared state for the wizard (stage, results, errors).
type WizardState struct {
	Stage int

	// Requirements tracking
	Requirements []requirements.Requirement
	ReqIndex     int // current requirement index

	// Current requirement state (for view rendering)
	CurrentReqName       string
	CurrentInstallCmd    string
	CurrentCanAutoInstall bool

	// Results
	CheckResult requirements.Result
	VerifyOK    bool
	VerifyErr   error
	InstallErr  error

	Quitting bool
}

// NewWizardState returns initial wizard state with all requirements loaded.
func NewWizardState() WizardState {
	reqs := requirements.AllRequirements()
	return WizardState{
		Stage:        StageWelcome,
		Requirements: reqs,
		ReqIndex:     0,
	}
}

// CurrentRequirement returns the current requirement being processed, or nil if done.
func (s *WizardState) CurrentRequirement() requirements.Requirement {
	if s.ReqIndex >= len(s.Requirements) {
		return nil
	}
	return s.Requirements[s.ReqIndex]
}

// Advance advances to the next stage and returns the Cmd to run for that stage, if any.
func (s *WizardState) Advance(msg tea.Msg) tea.Cmd {
	switch s.Stage {
	case StageWelcome:
		if _, ok := msg.(welcomeDoneMsg); ok {
			return s.startCheckingCurrentRequirement()
		}

	case StageChecking:
		if res, ok := msg.(requirements.CheckRequirementMsg); ok {
			s.CheckResult = res.Result
			if res.Result.OK {
				// Requirement OK, move to next or done
				return s.advanceToNextRequirementOrDone()
			}
			// Requirement not OK, prompt for installation
			s.Stage = StagePromptInstall
			req := s.CurrentRequirement()
			if req != nil {
				s.CurrentReqName = req.Name()
				s.CurrentInstallCmd = req.InstallCommand()
				s.CurrentCanAutoInstall = req.CanAutoInstall()
			}
			return nil
		}

	case StagePromptInstall:
		if _, ok := msg.(enterPressedMsg); ok {
			req := s.CurrentRequirement()
			if req == nil {
				s.Stage = StageDone
				return nil
			}
			if req.CanAutoInstall() {
				// Run install
				s.Stage = StageInstalling
				return func() tea.Msg {
					return requirements.InstallRequirementCmd(req, s.ReqIndex)().(tea.Msg)
				}
			}
			// Can't auto-install, go straight to verify (user should have installed manually)
			s.Stage = StageVerifying
			return func() tea.Msg {
				return requirements.VerifyRequirementCmd(req, s.ReqIndex)().(tea.Msg)
			}
		}

	case StageInstalling:
		if res, ok := msg.(requirements.InstallRequirementMsg); ok {
			s.InstallErr = res.Err
			// After install, always verify
			s.Stage = StageVerifying
			req := s.CurrentRequirement()
			if req == nil {
				s.Stage = StageDone
				return nil
			}
			return func() tea.Msg {
				return requirements.VerifyRequirementCmd(req, s.ReqIndex)().(tea.Msg)
			}
		}

	case StageVerifying:
		if res, ok := msg.(requirements.VerifyRequirementMsg); ok {
			s.VerifyOK = res.OK
			s.VerifyErr = res.Err
			if res.OK {
				// Move to next requirement or done
				return s.advanceToNextRequirementOrDone()
			}
			// Verify failed, go to done with error
			s.Stage = StageDone
			return nil
		}

	case StageDone:
		// No advance; done step handles quit tick
	}
	return nil
}

// startCheckingCurrentRequirement sets stage to Checking and returns the check command.
func (s *WizardState) startCheckingCurrentRequirement() tea.Cmd {
	req := s.CurrentRequirement()
	if req == nil {
		s.Stage = StageDone
		s.VerifyOK = true // All done, no requirements failed
		return nil
	}
	s.Stage = StageChecking
	return func() tea.Msg {
		return requirements.CheckRequirementCmd(req, s.ReqIndex)().(tea.Msg)
	}
}

// advanceToNextRequirementOrDone moves to the next requirement or to Done if all are processed.
func (s *WizardState) advanceToNextRequirementOrDone() tea.Cmd {
	s.ReqIndex++
	if s.ReqIndex >= len(s.Requirements) {
		s.Stage = StageDone
		s.VerifyOK = true
		return nil
	}
	// More requirements to process
	return s.startCheckingCurrentRequirement()
}

// CloseRequirements releases resources held by all requirements.
func (s *WizardState) CloseRequirements() {
	for _, req := range s.Requirements {
		if req != nil {
			req.Close()
		}
	}
}
