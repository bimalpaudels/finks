package steps

import (
	"github.com/charmbracelet/lipgloss"
)

var verifyDim = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

// VerifyView renders the "Verifying..." stage.
func VerifyView() string {
	msg := verifyDim.Render("Verifying dependencies are ready...")
	return "\n  " + msg + "\n\n"
}
