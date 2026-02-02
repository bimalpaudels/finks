package steps

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	doneSuccess = lipgloss.NewStyle().Foreground(lipgloss.Color("78")).Bold(true)
	doneDim     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	doneError   = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
)

// DoneViewSimple renders the done stage with just success/error state.
// If verifyOK is true, shows success message.
// If verifyOK is false, shows error message from verifyErr or installErr.
func DoneViewSimple(verifyOK bool, verifyErr error, installErr error) string {
	if verifyOK {
		msg := doneSuccess.Render("Everything is set!")
		hint := doneDim.Render("Finks is ready. Run `finks --help` to get started.")
		return "\n  " + msg + "\n\n  " + hint + "\n\n"
	}
	// Show failure
	var errMsg string
	if verifyErr != nil {
		errMsg = verifyErr.Error()
	} else if installErr != nil {
		errMsg = installErr.Error()
	} else {
		errMsg = "A required dependency is not ready. Install it and run this wizard again."
	}
	line := doneError.Render("Setup failed: " + errMsg)
	hint := doneDim.Render("After fixing the issue, run this wizard again.")
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
