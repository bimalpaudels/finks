package steps

import (
	"github.com/charmbracelet/lipgloss"
)

var checkDim = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

// CheckView renders the "Checking requirements..." stage.
func CheckView() string {
	msg := checkDim.Render("Checking requirements...")
	return "\n  " + msg + "\n\n"
}
