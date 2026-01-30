package steps

import (
	"github.com/bimalpaudels/finks/internal/checker"
	"github.com/charmbracelet/lipgloss"
)

var (
	doneSuccess = lipgloss.NewStyle().Foreground(lipgloss.Color("78")).Bold(true)
	doneDim     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	doneError   = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
)

// DoneView renders the "You're good to go" stage.
// If verifyOK is false, shows error message and install hint.
func DoneView(verifyOK bool, verifyErr error, installErr error, checkResult checker.CheckResultMsg) string {
	if verifyOK {
		msg := doneSuccess.Render("You're good to go!")
		hint := doneDim.Render("Finks is ready. Run `finks --help` to get started.")
		return "\n  " + msg + "\n\n  " + hint + "\n\n"
	}
	// Show failure: Docker missing or verify failed
	var errMsg string
	if verifyErr != nil {
		errMsg = verifyErr.Error()
	} else if installErr != nil {
		errMsg = installErr.Error()
	} else if !checkResult.DockerOK && checkResult.Err != nil {
		errMsg = checkResult.Err.Error()
	} else {
		errMsg = "Docker is required. Install from https://docs.docker.com/get-docker/"
	}
	line := doneError.Render("Docker is not ready: " + errMsg)
	hint := doneDim.Render("After installing Docker, run this wizard again.")
	return "\n  " + line + "\n\n  " + hint + "\n\n"
}

// DoneViewSuccess is a simple success-only view (no checker types).
func DoneViewSuccess() string {
	msg := doneSuccess.Render("You're good to go!")
	hint := doneDim.Render("Finks is ready. Run `finks --help` to get started.")
	return "\n  " + msg + "\n\n  " + hint + "\n\n"
}

// DoneViewError renders the done stage when verification failed.
func DoneViewError(err error) string {
	line := doneError.Render("Docker is not ready: " + err.Error())
	hint := doneDim.Render("After installing Docker, run this wizard again.")
	return "\n  " + line + "\n\n  " + hint + "\n\n"
}
